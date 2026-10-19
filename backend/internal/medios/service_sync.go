package medios

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// service_sync.go implementa «Buscar episodios nuevos» (FR-011…FR-018): la única
// automatización de F9, **a demanda** (sin cron, sin worker y sin cola, Out of
// Scope) y en la misma petición del panel.
//
// El camino completo de una búsqueda es:
//
//  1. **Leer los dos feeds en paralelo** (T920): un feed que no responde queda
//     `failed` en el resumen y **no** se interpreta como que desaparecieron
//     episodios (FR-016).
//  2. **Importar lo que falte** (T920), por `(provider, external_id)`: la
//     búsqueda solo **agrega** (FR-012) y la instantánea se toma una sola vez
//     (R9-6: nunca refresca título, fecha, carátula ni enlaces).
//  3. **Unir automáticamente** el mismo sermón cuando la correspondencia es
//     segura —título normalizado igual **y** el mismo día en America/Costa_Rica,
//     sin margen— (FR-013) o dejar el par **marcado** «posible duplicado» cuando
//     no lo es (FR-014).
//  4. **Comprobar la disponibilidad** de lo que salió de la ventana de su feed
//     (FR-017), **copiar las carátulas** de lo importado (P9-7) y cerrar con el
//     **resumen** de FR-018 más su registro (`media_sync_runs` + una fila
//     `media.sync`).
//
// La búsqueda entera cabe en un presupuesto de `presupuestoBusqueda` (90 s,
// P9-3/R9-9): al agotarse se cierra con lo hecho —cada elemento ya quedó guardado
// en su transacción— y el resumen lleva conteos parciales.
//
// Todo el contenido que llega de los feeds se trata como **dato, nunca código**
// (constitución §IV): las descripciones se guardan como texto plano y los
// mensajes del resumen no llevan nunca información técnica de red.

// busquedaLabel es la etiqueta base de una «Buscar episodios nuevos» en la
// auditoría (P9-14); el extracto del resumen se le añade al cerrar.
const busquedaLabel = "Medios · Búsqueda"

// lecturaFeed es el resultado de leer un origen (FR-010/FR-016).
type lecturaFeed struct {
	proveedor Proveedor
	// estado es `ok` si el feed respondió (aunque no traiga nada) y `failed` si
	// no se pudo leer.
	estado EstadoFeed
	// entradas son los elementos aprovechables del feed (vacío si falló).
	entradas []EntradaFeed
}

// busqueda es el estado de UNA ejecución de «Buscar episodios nuevos». Se
// construye al empezar y muere al terminar: no hay estado compartido entre
// búsquedas (FR-011: dos búsquedas seguidas no se pisan).
type busqueda struct {
	servicio *service
	ctx      context.Context
	resumen  ResumenBusqueda

	// fuentes son las fuentes ya importadas de cada proveedor que respondió:
	// son la llave de idempotencia (R9-11) y la ventana del feed (FR-017).
	fuentes map[Proveedor][]FuenteEpisodio
	// importados marca `(proveedor, id externo)` ya importados.
	importados map[Proveedor]map[string]bool
	// conFuente marca las fichas que ya tienen fuente de cada proveedor (una
	// ficha = 1 video y/o 1 audio).
	conFuente map[Proveedor]map[uuid.UUID]bool
	// fichas es la caché de fichas leídas durante la búsqueda.
	fichas map[uuid.UUID]Episodio
	// reclamadas marca las fichas que ya recibieron una fuente en esta búsqueda:
	// dos elementos del mismo feed no se unen a la misma ficha.
	reclamadas map[uuid.UUID]bool
	// elementos cuenta los elementos importados en esta búsqueda (tope duro de
	// P9-3: 30).
	elementos int
	// caratulas son las carátulas de los elementos importados en esta búsqueda
	// que hay que copiar al sitio (P9-7).
	caratulas []caratulaPendiente
	// parcial indica que la búsqueda se cerró antes de tiempo (presupuesto
	// agotado): el resumen lo dice y los conteos son parciales (R9-9).
	parcial bool
}

// BuscarNuevos implementa «Buscar episodios nuevos» (FR-011): lee los dos feeds,
// agrega lo que falte, une lo que es el mismo sermón, marca lo dudoso, aplica
// FR-017 sobre lo que salió de su feed y devuelve el resumen de FR-018.
//
// El error solo aparece ante un fallo de la base de datos: un feed que no
// responde **no** es un error (el resumen lo informa). Cada elemento se guarda en
// su propia transacción, así que un fallo a mitad de camino nunca deja el
// catálogo a medias ni pierde lo ya importado.
func (s *service) BuscarNuevos(ctx context.Context, actorID uuid.UUID) (ResumenBusqueda, error) {
	iniciado := time.Now()
	// Presupuesto de la búsqueda (P9-3): el corte vive en el contexto, de modo
	// que también lo respetan las lecturas y las comprobaciones salientes.
	ctxBusqueda, cancelar := context.WithTimeout(ctx, presupuestoBusqueda)
	defer cancelar()

	youtube, podcast := s.leerFeeds(ctxBusqueda)

	b := &busqueda{
		servicio:   s,
		ctx:        ctxBusqueda,
		resumen:    ResumenBusqueda{IniciadoEn: iniciado, EstadoYouTube: string(youtube.estado), EstadoPodcast: string(podcast.estado)},
		fuentes:    map[Proveedor][]FuenteEpisodio{},
		importados: map[Proveedor]map[string]bool{},
		conFuente:  map[Proveedor]map[uuid.UUID]bool{},
		fichas:     map[uuid.UUID]Episodio{},
		reclamadas: map[uuid.UUID]bool{},
	}

	// Importación de lo que falte, proveedor por proveedor (YouTube primero: el
	// orden es determinista y decide cuál es «la primera» de un par dudoso).
	for _, lectura := range []lecturaFeed{youtube, podcast} {
		if lectura.estado != FeedOK {
			continue
		}
		if err := b.importar(lectura); err != nil {
			b.resumen.TerminadoEn = time.Now()
			return b.resumen, err
		}
	}

	// FR-017: la ventana del feed, la disponibilidad a prueba de fallos y las
	// carátulas de lo importado (T921).
	if err := b.revisarDisponibilidad(youtube, podcast); err != nil {
		b.resumen.TerminadoEn = time.Now()
		return b.resumen, err
	}
	if err := b.copiarCaratulas(); err != nil {
		b.resumen.TerminadoEn = time.Now()
		return b.resumen, err
	}

	b.resumen.TerminadoEn = time.Now()
	if err := b.registrar(actorID); err != nil {
		return b.resumen, err
	}
	return b.resumen, nil
}

// leerFeeds lee los dos orígenes **en paralelo** (FR-010/P9-3): cada uno tiene su
// propio estado, de modo que el fallo de uno no arrastra al otro.
func (s *service) leerFeeds(ctx context.Context) (lecturaFeed, lecturaFeed) {
	if s.feeds == nil {
		// Sin fuente de feeds configurada la búsqueda no inventa nada: los dos
		// orígenes quedan `failed` y no se toca el catálogo.
		return lecturaFeed{proveedor: ProveedorYouTube, estado: FeedFallido},
			lecturaFeed{proveedor: ProveedorPodcast, estado: FeedFallido}
	}
	lecturas := [2]lecturaFeed{{proveedor: ProveedorYouTube}, {proveedor: ProveedorPodcast}}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		lecturas[0] = s.leerFeed(ctx, ProveedorYouTube)
	}()
	go func() {
		defer wg.Done()
		lecturas[1] = s.leerFeed(ctx, ProveedorPodcast)
	}()
	wg.Wait()
	return lecturas[0], lecturas[1]
}

// leerFeed lee un origen y traduce su fallo a `failed` (FR-016): un feed que no
// responde **nunca** se interpreta como que hayan desaparecido episodios, así que
// su rama no oculta ni marca nada.
func (s *service) leerFeed(ctx context.Context, proveedor Proveedor) lecturaFeed {
	var (
		resultado ResultadoFeed
		err       error
	)
	if proveedor == ProveedorYouTube {
		resultado, err = s.feeds.FetchYouTube(ctx)
	} else {
		resultado, err = s.feeds.FetchPodcast(ctx)
	}
	if err != nil {
		s.logError(ctx, "no se pudo leer el feed de "+nombreProveedor(proveedor), err)
		return lecturaFeed{proveedor: proveedor, estado: FeedFallido}
	}
	return lecturaFeed{proveedor: proveedor, estado: FeedOK, entradas: resultado.Entradas}
}

// ---------------------------------------------------------------------------
// Ventana del feed y disponibilidad (FR-017/P9-6/P9-19)
// ---------------------------------------------------------------------------

// comprobacion es una ficha de la cola de comprobación con las fuentes que
// salieron de su ventana.
type comprobacion struct {
	ficha   Episodio
	fuentes []FuenteEpisodio
}

// caratulaPendiente es una carátula que hay que copiar al sitio: la del elemento
// que se acaba de importar, en su ficha. `proveedor` es el del elemento que la
// trajo: cuando el mismo sermón la trae de los dos feeds manda la del video
// (default #7).
type caratulaPendiente struct {
	episodioID  uuid.UUID
	proveedor   Proveedor
	caratulaURL string
}

// revisarDisponibilidad aplica la regla del «estante de 15» (FR-017/SC-014): para
// cada proveedor **que respondió** compara los elementos importados con los que
// trajo el feed; lo que ya no aparece **solo se comprueba**, nunca se oculta por
// esa razón.
//
// La **cola** (corrección de la fase 5) es: fuentes fuera de la ventana que
// pertenecen a fichas `visibility='visible'` no absorbidas. Una ficha oculta por
// el equipo no se re-comprueba (nada de lo que decidió el equipo se reescribe) y
// «Mostrar de nuevo» la devuelve a la cola (D-28: si el enlace sigue muerto, la
// próxima búsqueda la vuelve a ocultar). El tope es `maxComprobacionesPorBusqueda`.
//
// La comprobación es **por fuente** y el resultado de la ficha es **el peor** de
// las comprobadas (regla M2): un enlace confirmado muerto oculta la ficha entera
// aunque el otro responda 200 (precisión 2).
func (b *busqueda) revisarDisponibilidad(youtube, podcast lecturaFeed) error {
	// Ventana de cada feed que respondió: los identificadores que trajo.
	ventanas := make(map[Proveedor]map[string]bool, 2)
	for _, lectura := range []lecturaFeed{youtube, podcast} {
		if lectura.estado != FeedOK {
			continue
		}
		enVentana := make(map[string]bool, len(lectura.entradas))
		for _, entrada := range lectura.entradas {
			enVentana[entrada.IDExterno] = true
		}
		ventanas[lectura.proveedor] = enVentana
	}

	cola, err := b.colaDeComprobacion(ventanas)
	if err != nil {
		return err
	}
	if len(cola) == 0 {
		return nil
	}

	peores := b.comprobar(cola)
	ahora := time.Now()
	for _, item := range cola {
		resultado, comprobada := peores[item.ficha.ID]
		if !comprobada {
			// Se quedó sin presupuesto o fuera del tope: la ficha no se toca.
			continue
		}
		fuentes, err := b.servicio.repository.ListMediaItemSourcesByItem(b.ctx, item.ficha.ID)
		if err != nil {
			if b.presupuestoAgotado() {
				return nil
			}
			return fmt.Errorf("leer las fuentes del episodio: %w", err)
		}
		if err := b.aplicarDisponibilidad(item.ficha, resultado, todasFueraDeVentana(fuentes, ventanas), ahora); err != nil {
			if b.presupuestoAgotado() {
				return nil
			}
			return err
		}
	}
	return nil
}

// colaDeComprobacion arma la lista de fichas a comprobar (una por ficha, con sus
// fuentes fuera de ventana) en un orden determinista.
func (b *busqueda) colaDeComprobacion(ventanas map[Proveedor]map[string]bool) ([]comprobacion, error) {
	porFicha := make(map[uuid.UUID][]FuenteEpisodio)
	orden := make([]uuid.UUID, 0)
	for proveedor, enVentana := range ventanas {
		for _, fuente := range b.fuentes[proveedor] {
			if enVentana[fuente.IDExterno] {
				continue
			}
			if _, visto := porFicha[fuente.EpisodioID]; !visto {
				orden = append(orden, fuente.EpisodioID)
			}
			porFicha[fuente.EpisodioID] = append(porFicha[fuente.EpisodioID], fuente)
		}
	}
	sort.Slice(orden, func(i, j int) bool { return orden[i].String() < orden[j].String() })

	cola := make([]comprobacion, 0, len(orden))
	for _, id := range orden {
		ficha, encontrada, err := b.fichaActual(id)
		if err != nil {
			return nil, err
		}
		// La cola solo mira fichas visibles no absorbidas (FR-017/D-28).
		if !encontrada || ficha.Visibilidad != VisibilidadVisible || ficha.UnidoAID != nil {
			continue
		}
		cola = append(cola, comprobacion{ficha: ficha, fuentes: porFicha[id]})
	}
	return cola, nil
}

// comprobar ejecuta las comprobaciones con `maxEnVuelo` en vuelo y el tope duro de
// P9-3, y devuelve el **peor** resultado de cada ficha (M2). El error del probador
// es solo informativo (la regla a prueba de fallos manda): se queda en el log y
// nunca cambia el desenlace.
func (b *busqueda) comprobar(cola []comprobacion) map[uuid.UUID]ResultadoDisponibilidad {
	tareas := make([]comprobacion, 0, len(cola))
	for _, item := range cola {
		for _, fuente := range item.fuentes {
			if len(tareas) >= maxComprobacionesPorBusqueda {
				break
			}
			tareas = append(tareas, comprobacion{ficha: item.ficha, fuentes: []FuenteEpisodio{fuente}})
		}
		if len(tareas) >= maxComprobacionesPorBusqueda {
			break
		}
	}

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		enVuelo = make(chan struct{}, maxEnVuelo)
		peores  = make(map[uuid.UUID]ResultadoDisponibilidad, len(cola))
	)
	for _, tarea := range tareas {
		if b.presupuestoAgotado() {
			break
		}
		wg.Add(1)
		enVuelo <- struct{}{}
		go func(tarea comprobacion) {
			defer wg.Done()
			defer func() { <-enVuelo }()
			resultado, err := b.comprobarFuente(tarea.fuentes[0])
			mu.Lock()
			defer mu.Unlock()
			b.resumen.Comprobados++
			if err != nil {
				b.servicio.logError(b.ctx, "no se pudo comprobar la disponibilidad del enlace", err)
			}
			if resultado == DisponibilidadInconclusa {
				b.resumen.Inconclusos++
			}
			if actual, ok := peores[tarea.ficha.ID]; !ok || peorQue(resultado, actual) {
				peores[tarea.ficha.ID] = resultado
			}
		}(tarea)
	}
	wg.Wait()
	return peores
}

// comprobarFuente comprueba el enlace de una fuente con el probador del dominio
// (oEmbed de YouTube para los videos y el propio enlace para el podcast). Sin
// probador configurado el resultado es inconcluso: nunca se oculta nada.
func (b *busqueda) comprobarFuente(fuente FuenteEpisodio) (ResultadoDisponibilidad, error) {
	if b.servicio.probador == nil {
		return DisponibilidadInconclusa, nil
	}
	if fuente.Proveedor == ProveedorYouTube {
		return b.servicio.probador.ProbeYouTube(b.ctx, fuente.IDExterno)
	}
	return b.servicio.probador.ProbePodcast(b.ctx, fuente.Enlace)
}

// aplicarDisponibilidad escribe el resultado agregado de la ficha y decide su
// marca (P9-19):
//
//   - **404/410 confirmado** → `availability_result='unavailable'`, la ficha se
//     **oculta** y queda marcada «no disponible» con la marca **nueva abierta**:
//     la indisponibilidad confirmada **manda siempre**, aunque el equipo hubiera
//     resuelto antes «mantener visible» (A3, precisión de la fase 5) o aunque la
//     marca anterior fuera de otra clase (cambiar de clase resetea la resolución).
//   - **todas sus fuentes fuera de ventana** y sin confirmación de muerte → marca
//     «ya no aparece en el feed» y la ficha **sigue visible**. Si la misma clase
//     ya estaba **resuelta**, no se reabre (default #6).
//   - **alguna fuente sigue en su ventana** → no se marca nada (M2: decir «ya no
//     aparece en el feed» sería falso); solo se guarda el dato de la comprobación.
//
// El eje de duplicados **no se toca** (es independiente: C1).
func (b *busqueda) aplicarDisponibilidad(ficha Episodio, resultado ResultadoDisponibilidad, todasFuera bool, ahora time.Time) error {
	actualizada := ficha
	actualizada.DisponibilidadComprobadaEn = &ahora
	actualizada.ResultadoDisponibilidad = &resultado

	oculta := false
	switch {
	case resultado == DisponibilidadNoDisponible:
		marca := MarcaDisponibilidadNoDisponible
		actualizada.MarcaDisponibilidad = &marca
		actualizada.MarcaDisponibilidadResueltaEn = nil
		oculta = actualizada.Visibilidad != VisibilidadOculta
		actualizada.Visibilidad = VisibilidadOculta
	case todasFuera:
		yaResuelta := actualizada.MarcaDisponibilidad != nil &&
			*actualizada.MarcaDisponibilidad == MarcaDisponibilidadFueraDelFeed &&
			actualizada.MarcaDisponibilidadResueltaEn != nil
		if !yaResuelta {
			marca := MarcaDisponibilidadFueraDelFeed
			actualizada.MarcaDisponibilidad = &marca
			actualizada.MarcaDisponibilidadResueltaEn = nil
		}
	}

	if _, err := b.servicio.repository.UpdateMediaItem(b.ctx, actualizada); err != nil {
		return fmt.Errorf("marcar la disponibilidad del episodio: %w", err)
	}
	if oculta {
		b.resumen.Ocultos++
	}
	return nil
}

// todasFueraDeVentana indica si **todas** las fuentes de la ficha salieron de sus
// ventanas. La fuente de un proveedor que **no respondió** no cuenta como fuera de
// ventana (FR-016: un feed caído no se interpreta como que desaparecieron
// episodios).
func todasFueraDeVentana(fuentes []FuenteEpisodio, ventanas map[Proveedor]map[string]bool) bool {
	if len(fuentes) == 0 {
		return false
	}
	for _, fuente := range fuentes {
		enVentana, respondio := ventanas[fuente.Proveedor]
		if !respondio || enVentana[fuente.IDExterno] {
			return false
		}
	}
	return true
}

// peorQue indica si el candidato es peor que el resultado actual (M2:
// `unavailable` peor que `inconclusive`, peor que `available`).
func peorQue(candidato, actual ResultadoDisponibilidad) bool {
	return gravedadDisponibilidad(candidato) > gravedadDisponibilidad(actual)
}

// gravedadDisponibilidad ordena los resultados de disponibilidad.
func gravedadDisponibilidad(resultado ResultadoDisponibilidad) int {
	switch resultado {
	case DisponibilidadNoDisponible:
		return 2
	case DisponibilidadInconclusa:
		return 1
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// Carátulas (P9-7/SC-005)
// ---------------------------------------------------------------------------

// copiarCaratulas copia al sitio la carátula de los elementos importados en esta
// búsqueda (≤ `maxCaratulasPorBusqueda`, `maxEnVuelo` en vuelo y el timeout del
// descargador): la ficha se queda con el nombre que generó el servidor y el
// visitante no pide nada a YouTube ni a Spotify (SC-005). Un fallo **nunca**
// aborta la importación: la ficha entra sin carátula (marcador neutro) y la
// búsqueda sigue.
//
// La carátula es una **instantánea** (P9-9/default #10): una ficha que ya tiene
// carátula **no se toca** —ni se descarga la otra— y, cuando el mismo sermón trae
// carátula de los dos feeds, manda la del **video** (default #7). Por eso cada
// ficha tiene **una sola** descarga: el ganador no depende de qué goroutine
// termine antes.
func (b *busqueda) copiarCaratulas() error {
	if b.servicio.caratulas == nil || len(b.caratulas) == 0 {
		return nil
	}
	tareas := b.caratulasPorFicha()
	if len(tareas) > maxCaratulasPorBusqueda {
		tareas = tareas[:maxCaratulasPorBusqueda]
	}

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		enVuelo = make(chan struct{}, maxEnVuelo)
		fallo   error
	)
	for _, tarea := range tareas {
		if b.presupuestoAgotado() {
			break
		}
		wg.Add(1)
		enVuelo <- struct{}{}
		go func(tarea caratulaPendiente) {
			defer wg.Done()
			defer func() { <-enVuelo }()
			if err := b.copiarCaratula(tarea); err != nil {
				mu.Lock()
				if fallo == nil {
					fallo = err
				}
				mu.Unlock()
			}
		}(tarea)
	}
	wg.Wait()

	if fallo != nil && !b.presupuestoAgotado() {
		return fallo
	}
	return nil
}

// caratulasPorFicha devuelve **una** carátula por ficha, en el orden en que se
// encolaron: el mismo sermón puede traer carátula del video y del podcast y el
// ganador no puede depender del orden en que terminen las descargas (default #7:
// manda la del video).
func (b *busqueda) caratulasPorFicha() []caratulaPendiente {
	elegidas := make([]caratulaPendiente, 0, len(b.caratulas))
	posicion := make(map[uuid.UUID]int, len(b.caratulas))
	for _, pendiente := range b.caratulas {
		i, vista := posicion[pendiente.episodioID]
		if !vista {
			posicion[pendiente.episodioID] = len(elegidas)
			elegidas = append(elegidas, pendiente)
			continue
		}
		if pendiente.proveedor == ProveedorYouTube && elegidas[i].proveedor != ProveedorYouTube {
			elegidas[i] = pendiente
		}
	}
	return elegidas
}

// copiarCaratula copia la carátula de UNA ficha y la deja escrita si la ficha
// sigue sin carátula. Lee la fila viva (sin la caché de la búsqueda: este paso
// corre en paralelo) y la escritura condicional resuelve la carrera con otra
// búsqueda. La copia que no se usa se retira del almacén en mejor esfuerzo.
func (b *busqueda) copiarCaratula(tarea caratulaPendiente) error {
	ficha, encontrada, err := b.servicio.repository.GetMediaItemByID(b.ctx, tarea.episodioID)
	if err != nil {
		return fmt.Errorf("leer el episodio de la carátula: %w", err)
	}
	if !encontrada || ficha.Caratula != nil {
		// La ficha ya tiene carátula (la del video, la de otra búsqueda o la que
		// el equipo puso): la instantánea no se refresca (P9-9/default #10) y la
		// descarga ni se intenta.
		return nil
	}

	archivo, err := b.servicio.caratulas.FetchCover(b.ctx, tarea.caratulaURL)
	if err != nil {
		b.servicio.logError(b.ctx, "no se pudo copiar la carátula del episodio", err)
		return nil
	}

	if _, aplicada, err := b.servicio.repository.UpdateMediaItemCoverIfNull(b.ctx, tarea.episodioID, archivo); err != nil {
		return fmt.Errorf("guardar la carátula del episodio: %w", err)
	} else if !aplicada {
		// Otra búsqueda simultánea se adelantó: su copia es la que sirve el sitio.
		b.servicio.borrarCaratulasSinUso(b.ctx, nil, &archivo)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Cierre de la búsqueda (FR-018/FR-022)
// ---------------------------------------------------------------------------

// registrar persiste el resumen de la búsqueda en `media_sync_runs` y deja **una**
// sola fila de auditoría `media.sync` con el extracto del resumen en
// `targetLabel` (P9-14: el alta importada nunca emite `media.*` extra). Los dos
// registros van en la misma transacción (FR-022).
//
// El registro se escribe con un contexto **sin la cancelación de la búsqueda**: una
// búsqueda parcial (presupuesto agotado) también se registra, que es justamente
// cuando el equipo más necesita ver su resumen (R9-9).
func (b *busqueda) registrar(actorID uuid.UUID) error {
	run := SyncRun{
		IniciadoEn:    b.resumen.IniciadoEn,
		TerminadoEn:   b.resumen.TerminadoEn,
		Importados:    b.resumen.Importados,
		Unidos:        b.resumen.Unidos,
		Marcados:      b.resumen.Marcados,
		Ocultos:       b.resumen.Ocultos,
		Comprobados:   b.resumen.Comprobados,
		Inconclusos:   b.resumen.Inconclusos,
		EstadoYouTube: EstadoFeed(b.resumen.EstadoYouTube),
		EstadoPodcast: EstadoFeed(b.resumen.EstadoPodcast),
	}
	if actorID != uuid.Nil {
		run.ActorUserID = &actorID
	}

	action := audit.Action{
		ActorUserID: run.ActorUserID,
		Code:        audit.ActionMediaSync,
		TargetKind:  audit.TargetContent,
		TargetLabel: etiquetaBusqueda(b.resumen, b.parcial),
		Result:      audit.ResultSuccess,
	}
	if _, err := b.servicio.repository.InsertMediaSyncRun(context.WithoutCancel(b.ctx), run, action); err != nil {
		return fmt.Errorf("registrar la búsqueda: %w", err)
	}
	return nil
}

// etiquetaBusqueda arma `Medios · Búsqueda · <extracto>` (P9-14): el detalle
// numérico vive en `media_sync_runs` y este extracto —sin tecnicismos— es lo que
// la auditoría muestra. Una búsqueda parcial lo dice.
func etiquetaBusqueda(resumen ResumenBusqueda, parcial bool) string {
	partes := make([]string, 0, 6)
	if resumen.Importados > 0 {
		partes = append(partes, fmt.Sprintf("%d entraron", resumen.Importados))
	}
	if resumen.Unidos > 0 {
		partes = append(partes, fmt.Sprintf("%d unieron", resumen.Unidos))
	}
	if resumen.Marcados > 0 {
		partes = append(partes, fmt.Sprintf("%d marcaron", resumen.Marcados))
	}
	if resumen.Ocultos > 0 {
		partes = append(partes, fmt.Sprintf("%d se ocultaron", resumen.Ocultos))
	}
	if resumen.Comprobados > 0 {
		partes = append(partes, fmt.Sprintf("%d comprobaciones", resumen.Comprobados))
	}
	if len(partes) == 0 {
		partes = append(partes, "sin novedades")
	}
	if parcial {
		partes = append(partes, "sin terminar")
	}

	label := busquedaLabel + " · " + strings.Join(partes, ", ")
	if runes := []rune(label); len(runes) > maxTargetLabel {
		label = string(runes[:maxTargetLabel])
	}
	return label
}

// importar agrega los elementos que falten de un feed (FR-012): los ya importados
// se saltan sin tocarlos (idempotencia por `(provider, external_id)`, R9-11).
func (b *busqueda) importar(lectura lecturaFeed) error {
	if err := b.cargarFuentes(lectura.proveedor); err != nil {
		return err
	}
	for _, entrada := range lectura.entradas {
		if b.elementos >= maxElementosPorBusqueda {
			// Tope duro de P9-3: lo que sobra entra en la próxima búsqueda.
			b.parcial = true
			return nil
		}
		if b.presupuestoAgotado() {
			return nil
		}
		if b.importados[lectura.proveedor][entrada.IDExterno] {
			continue
		}
		if err := b.importarEntrada(lectura.proveedor, entrada); err != nil {
			return err
		}
	}
	return nil
}

// importarEntrada decide el destino de un elemento nuevo: unión automática con el
// mismo sermón del otro feed, par dudoso marcado o entrada independiente
// (FR-013/FR-014).
func (b *busqueda) importarEntrada(proveedor Proveedor, entrada EntradaFeed) error {
	candidato, seguro, dudoso, err := b.candidato(proveedor, entrada)
	if err != nil {
		return err
	}
	switch {
	case seguro:
		return b.unirEntrada(proveedor, entrada, candidato)
	case dudoso:
		return b.marcarDudoso(proveedor, entrada, candidato)
	default:
		return b.crearFicha(proveedor, entrada)
	}
}

// candidato busca la ficha del **otro** feed que corresponde a este elemento
// (FR-013): `seguro` cuando el título normalizado coincide **y** es el mismo día
// en la zona de la iglesia —sin margen—; `dudoso` cuando solo coincide una de las
// dos cosas (FR-014: entradas separadas y marcadas). Una ficha absorbida, una que
// ya tiene fuente de este proveedor, una ya reclamada en esta búsqueda y una que
// **ya tiene pareja** (par abierto o resuelto: la pareja recíproca es de dos, M1)
// no son candidatas.
func (b *busqueda) candidato(proveedor Proveedor, entrada EntradaFeed) (Episodio, bool, bool, error) {
	otro := proveedorContrario(proveedor)
	llave := normalizarTitulo(entrada.Titulo)
	dia := diaEnLugar(entrada.Publicado, b.servicio.zona())

	var dudoso Episodio
	for _, fuente := range b.fuentes[otro] {
		if b.conFuente[proveedor][fuente.EpisodioID] || b.reclamadas[fuente.EpisodioID] {
			continue
		}
		ficha, encontrada, err := b.ficha(fuente.EpisodioID)
		if err != nil {
			return Episodio{}, false, false, err
		}
		if !encontrada || ficha.UnidoAID != nil || ficha.DuplicadoDeID != nil {
			continue
		}
		mismoTitulo := normalizarTitulo(ficha.Titulo) == llave
		mismoDia := diaEnLugar(ficha.PublicadoEn, b.servicio.zona()).Equal(dia)
		switch {
		case mismoTitulo && mismoDia:
			return ficha, true, false, nil
		case (mismoTitulo || mismoDia) && dudoso.ID == uuid.Nil:
			dudoso = ficha
		}
	}
	if dudoso.ID != uuid.Nil {
		return dudoso, false, true, nil
	}
	return Episodio{}, false, false, nil
}

// unirEntrada une el elemento nuevo con la ficha del mismo sermón (FR-013/P9-10):
// la fuente entra en la ficha existente (`ON CONFLICT` = idempotencia), la ficha
// queda con los dos enlaces (`origin='both'`) y, si el elemento nuevo es el
// **video**, su nombre y su fecha pasan a mandar (precisión 5/default #7). La
// carátula la copia T921 si la consigue: mientras tanto se conserva la que la
// ficha tuviera.
func (b *busqueda) unirEntrada(proveedor Proveedor, entrada EntradaFeed, candidato Episodio) error {
	resultante := candidato
	if proveedor == ProveedorYouTube {
		// La candidata es siempre del otro feed, así que un elemento de YouTube
		// llega a una ficha del podcast: el video manda (precisión 5).
		resultante.Titulo = recortarTitulo(entrada.Titulo)
		resultante.PublicadoEn = entrada.Publicado
	}
	resultante.Origen = OrigenAmbos

	if _, ok, err := b.insertarFuente(proveedor, entrada, candidato.ID); err != nil {
		return err
	} else if !ok {
		// Otra búsqueda simultánea importó el elemento: la base ya decidió y
		// esta búsqueda no toca nada (R9-11/default #12).
		return nil
	}
	if _, err := b.servicio.repository.UpdateMediaItem(b.ctx, resultante); err != nil {
		return fmt.Errorf("actualizar el episodio unido: %w", err)
	}
	b.reclamadas[candidato.ID] = true
	b.elementos++
	b.resumen.Unidos++
	return nil
}

// marcarDudoso importa el elemento como **entrada separada** (FR-014/default #2):
// la adicional nace **oculta** y **las dos** fichas quedan marcadas «posible
// duplicado» con el par recíproco abierto, de modo que el visitante nunca ve el
// mismo sermón dos veces (SC-003) y el equipo las une con un clic.
//
// El alta de la adicional, su fuente y el puntero de vuelta del candidato van en
// **una sola transacción** (`MarkDuplicatePair`): si la fuente conflictúa (otra
// búsqueda simultánea importó el mismo elemento) no queda nada escrito —ni una
// ficha oculta sin enlaces ni una pareja a medias (H3)—.
func (b *busqueda) marcarDudoso(proveedor Proveedor, entrada EntradaFeed, candidato Episodio) error {
	nueva := fichaDe(entrada, proveedor, VisibilidadOculta)
	nueva.DuplicadoDeID = &candidato.ID

	_, fuente, ok, err := b.servicio.repository.MarkDuplicatePair(
		b.ctx, nueva, fuenteDe(proveedor, entrada, uuid.Nil), candidato.ID)
	if err != nil {
		return fmt.Errorf("crear el episodio adicional: %w", err)
	}
	if !ok {
		// Otra búsqueda simultánea importó el elemento: la base ya decidió y esta
		// búsqueda no toca nada (nada quedó escrito).
		return nil
	}

	b.registrarFuente(fuente, entrada)
	// El candidato queda «reclamado» y su copia en la caché queda al día: ni esta
	// búsqueda ni la siguiente lectura lo vuelven a ofrecer como candidato
	// (invariante de la pareja recíproca, M1).
	b.reclamadas[fuente.EpisodioID] = true
	b.reclamadas[candidato.ID] = true
	primera := candidato
	primera.DuplicadoDeID = &fuente.EpisodioID
	primera.MarcaDuplicadoResueltaEn = nil
	b.fichas[candidato.ID] = primera
	b.elementos++
	b.resumen.Importados++
	// `markedCount` cuenta las fichas marcadas «posible duplicado» nuevas de esta
	// búsqueda (FR-018, criterio B2): las dos del par.
	b.resumen.Marcados += 2
	return nil
}

// crearFicha importa el elemento como entrada independiente (US2 esc. 4): una
// ficha visible, sin marca y con su único enlace. La descripción corta se recorta
// al importar (P9-17) y el campo queda ausente si el feed no traía descripción.
//
// El alta (ficha + fuente) es **una sola transacción** (H3/SC-003): si la fuente
// conflictúa —otra búsqueda simultánea importó el mismo elemento— la ficha se
// revierte y **no queda ninguna ficha huérfana sin enlaces**; esta búsqueda lo
// cuenta como no importado.
func (b *busqueda) crearFicha(proveedor Proveedor, entrada EntradaFeed) error {
	nueva := fichaDe(entrada, proveedor, VisibilidadVisible)
	_, fuente, ok, err := b.servicio.repository.InsertMediaItemWithSource(
		b.ctx, nueva, fuenteDe(proveedor, entrada, uuid.Nil))
	if err != nil {
		return fmt.Errorf("crear el episodio: %w", err)
	}
	if !ok {
		// Otra búsqueda simultánea importó el elemento (R9-11): nada quedó escrito.
		return nil
	}
	b.registrarFuente(fuente, entrada)
	b.elementos++
	b.resumen.Importados++
	return nil
}

// registrarFuente anota en el estado de la búsqueda una fuente recién escrita: la
// llave de idempotencia del proveedor, la ficha que ya tiene esa fuente y la
// carátula pendiente (P9-7/SC-005).
func (b *busqueda) registrarFuente(fuente FuenteEpisodio, entrada EntradaFeed) {
	b.importados[fuente.Proveedor][fuente.IDExterno] = true
	b.conFuente[fuente.Proveedor][fuente.EpisodioID] = true
	// La lista de fuentes del proveedor se mantiene al día: el segundo feed de la
	// búsqueda encuentra así las fichas que acaba de crear el primero.
	b.fuentes[fuente.Proveedor] = append(b.fuentes[fuente.Proveedor], fuente)
	if entrada.CaratulaURL != "" {
		// La carátula se copia al sitio al importar (P9-7/SC-005): se encola con
		// su proveedor y la elección por ficha (el video manda) la hace
		// `caratulasPorFicha` (default #7).
		b.caratulas = append(b.caratulas, caratulaPendiente{
			episodioID:  fuente.EpisodioID,
			proveedor:   fuente.Proveedor,
			caratulaURL: entrada.CaratulaURL,
		})
	}
}

// fuenteDe arma la procedencia de un elemento de feed para su ficha (el id de la
// ficha lo fija el alta, dentro de su transacción).
func fuenteDe(proveedor Proveedor, entrada EntradaFeed, episodioID uuid.UUID) FuenteEpisodio {
	return FuenteEpisodio{
		EpisodioID: episodioID,
		Proveedor:  proveedor,
		IDExterno:  entrada.IDExterno,
		TituloFeed: recortarTitulo(entrada.Titulo),
		Enlace:     entrada.Enlace,
	}
}

// insertarFuente guarda la procedencia de un elemento que se **une** a una ficha
// existente con `ON CONFLICT (provider, external_id) DO NOTHING` (R9-11): el
// segundo valor es false cuando la base decidió que el elemento ya estaba
// importado (dos búsquedas simultáneas), y entonces esta búsqueda no toca nada
// (aquí no se escribió ninguna ficha que revertir). Un fallo de escritura sube
// como error: el elemento no queda a medias.
func (b *busqueda) insertarFuente(proveedor Proveedor, entrada EntradaFeed, episodioID uuid.UUID) (FuenteEpisodio, bool, error) {
	insertada, ok, err := b.servicio.repository.InsertMediaEpisodeSource(
		b.ctx, fuenteDe(proveedor, entrada, episodioID))
	if err != nil {
		return FuenteEpisodio{}, false, fmt.Errorf("importar la fuente del episodio: %w", err)
	}
	if ok {
		b.registrarFuente(insertada, entrada)
	}
	return insertada, ok, nil
}

// cargarFuentes lee las fuentes ya importadas de un proveedor y arma las llaves
// de idempotencia de la búsqueda (R9-11) y el mapa de fichas con fuente de ese
// proveedor.
func (b *busqueda) cargarFuentes(proveedor Proveedor) error {
	fuentes, err := b.servicio.repository.ListMediaEpisodeSourcesByProvider(b.ctx, proveedor)
	if err != nil {
		return fmt.Errorf("leer las fuentes importadas de %s: %w", nombreProveedor(proveedor), err)
	}
	b.fuentes[proveedor] = fuentes
	b.importados[proveedor] = make(map[string]bool, len(fuentes))
	b.conFuente[proveedor] = make(map[uuid.UUID]bool, len(fuentes))
	for _, fuente := range fuentes {
		b.importados[proveedor][fuente.IDExterno] = true
		b.conFuente[proveedor][fuente.EpisodioID] = true
	}
	return nil
}

// ficha devuelve una ficha del catálogo con su caché de la búsqueda: la unión
// automática y la ventana del feed miran muchas veces las mismas fichas.
func (b *busqueda) ficha(id uuid.UUID) (Episodio, bool, error) {
	if ficha, ok := b.fichas[id]; ok {
		return ficha, true, nil
	}
	ficha, encontrada, err := b.servicio.repository.GetMediaItemByID(b.ctx, id)
	if err != nil {
		return Episodio{}, false, fmt.Errorf("leer el episodio: %w", err)
	}
	if encontrada {
		b.fichas[id] = ficha
	}
	return ficha, encontrada, nil
}

// fichaActual lee la ficha del catálogo **sin caché**: las escrituras de esta
// búsqueda (la unión automática y el marcado del par dudoso) cambian filas que la
// caché ya vio, y escribir una copia vieja desharía el trabajo de la búsqueda.
func (b *busqueda) fichaActual(id uuid.UUID) (Episodio, bool, error) {
	ficha, encontrada, err := b.servicio.repository.GetMediaItemByID(b.ctx, id)
	if err != nil {
		return Episodio{}, false, fmt.Errorf("leer el episodio: %w", err)
	}
	if encontrada {
		b.fichas[id] = ficha
	}
	return ficha, encontrada, nil
}

// presupuestoAgotado indica si la búsqueda agotó su presupuesto (o si la petición
// se canceló): en ese caso se cierra con lo hecho y el resumen lo dice (R9-9).
func (b *busqueda) presupuestoAgotado() bool {
	if b.ctx.Err() == nil {
		return false
	}
	b.parcial = true
	return true
}

// fichaDe traduce un elemento de feed a la ficha que se importa: el nombre tal
// como lo trae el feed (FR-003, solo recortado a su límite), la fecha del feed,
// la descripción corta ya recortada (P9-17), el origen derivado del proveedor y
// la visibilidad que decide la importación.
func fichaDe(entrada EntradaFeed, proveedor Proveedor, visibilidad Visibilidad) Episodio {
	return Episodio{
		Titulo:        recortarTitulo(entrada.Titulo),
		PublicadoEn:   entrada.Publicado,
		DescripcionEs: descripcionCorta(entrada.Descripcion),
		Origen:        origenDelProveedor(proveedor),
		Visibilidad:   visibilidad,
	}
}

// descripcionCorta recorta la descripción del feed para la ficha (P9-17/FR-020):
// HTML → texto plano (constitución §IV: el contenido es dato, nunca código),
// primeras frases, **≤ 300 caracteres** y corte en frontera de frase o de palabra
// marcado con «…». Un feed sin descripción deja el campo **ausente** (precisión 3
// de la puerta 2): nunca se inventa texto ni se guarda `""`.
func descripcionCorta(descripcion string) *string {
	texto := textoPlanoDescripcion(descripcion)
	if texto == "" {
		return nil
	}
	runas := []rune(texto)
	if len(runas) <= maxDescripcion {
		return &texto
	}

	ventana := runas[:maxDescripcion]
	corte := fronteraDeFrase(ventana)
	if corte < len(ventana)/2 {
		// Sin una frase que cierre la mitad del texto se corta en palabra.
		corte = fronteraDePalabra(ventana)
	}
	recorte := strings.TrimRight(string(ventana[:corte]), " ")
	// El «…» cuenta dentro del límite: el recorte nunca supera los 300.
	if recortadas := []rune(recorte); len(recortadas) > maxDescripcion-1 {
		recorte = string(recortadas[:maxDescripcion-1])
	}
	recortada := recorte + "…"
	return &recortada
}

// fronteraDeFrase devuelve el índice justo después del último cierre de frase de
// la ventana (`.`/`!`/`?`/`…`), o 0 si no hay ninguno.
func fronteraDeFrase(ventana []rune) int {
	for i := len(ventana) - 1; i >= 0; i-- {
		if esCierreDeFrase(ventana[i]) {
			return i + 1
		}
	}
	return 0
}

// esCierreDeFrase indica si el carácter cierra una frase.
func esCierreDeFrase(c rune) bool {
	return c == '.' || c == '!' || c == '?' || c == '…'
}

// fronteraDePalabra devuelve el índice del último espacio de la ventana, o 0 si
// la ventana es una sola palabra.
func fronteraDePalabra(ventana []rune) int {
	for i := len(ventana) - 1; i >= 0; i-- {
		if ventana[i] == ' ' {
			return i
		}
	}
	return 0
}

// recortarTitulo normaliza el nombre que llega del feed (trim y espacios
// colapsados) y lo acota al límite de la columna (`maxTitulo`): los feeds reales
// van muy por debajo y la base es la última línea de defensa.
func recortarTitulo(titulo string) string {
	titulo = colapsarEspacios(titulo)
	if runas := []rune(titulo); len(runas) > maxTitulo {
		titulo = string(runas[:maxTitulo])
	}
	return titulo
}

// origenDelProveedor deriva los formatos de una ficha nueva (FR-003).
func origenDelProveedor(proveedor Proveedor) OrigenMedio {
	if proveedor == ProveedorYouTube {
		return OrigenVideo
	}
	return OrigenAudio
}

// proveedorContrario devuelve el otro origen.
func proveedorContrario(proveedor Proveedor) Proveedor {
	if proveedor == ProveedorYouTube {
		return ProveedorPodcast
	}
	return ProveedorYouTube
}

// nombreProveedor es el nombre con el que la iglesia llama a cada servicio (nunca
// «Atom», «RSS» ni «Anchor»: los mensajes son para personas).
func nombreProveedor(proveedor Proveedor) string {
	if proveedor == ProveedorYouTube {
		return "YouTube"
	}
	return "el podcast"
}
