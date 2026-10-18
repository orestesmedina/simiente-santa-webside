package medios

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
//  4. Cerrar con el **resumen** de FR-018. La comprobación de disponibilidad de
//     lo que salió de la ventana del feed (FR-017), la copia de las **carátulas**
//     y el registro de la búsqueda (`media_sync_runs` + `media.sync`) llegan en
//     T921, que completa este archivo.
//
// La búsqueda entera cabe en un presupuesto de `presupuestoBusqueda` (90 s,
// P9-3/R9-9): al agotarse se cierra con lo hecho —cada elemento ya quedó guardado
// en su transacción— y el resumen lleva conteos parciales.
//
// Todo el contenido que llega de los feeds se trata como **dato, nunca código**
// (constitución §IV): las descripciones se guardan como texto plano y los
// mensajes del resumen no llevan nunca información técnica de red.

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

	b.resumen.TerminadoEn = time.Now()
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
// ya tiene fuente de este proveedor y una ya reclamada en esta búsqueda no son
// candidatas.
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
		if !encontrada || ficha.UnidoAID != nil {
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
func (b *busqueda) marcarDudoso(proveedor Proveedor, entrada EntradaFeed, candidato Episodio) error {
	nueva := fichaDe(entrada, proveedor, VisibilidadOculta)
	nueva.DuplicadoDeID = &candidato.ID

	creada, err := b.servicio.repository.InsertMediaItem(b.ctx, nueva)
	if err != nil {
		return fmt.Errorf("crear el episodio adicional: %w", err)
	}
	if _, ok, err := b.insertarFuente(proveedor, entrada, creada.ID); err != nil {
		return err
	} else if !ok {
		return nil
	}

	// La primera conserva su visibilidad y apunta de vuelta (pareja recíproca,
	// invariante de service: la BD no la puede comprobar porque mira otra fila).
	primera := candidato
	primera.DuplicadoDeID = &creada.ID
	primera.MarcaDuplicadoResueltaEn = nil
	if _, err := b.servicio.repository.UpdateMediaItem(b.ctx, primera); err != nil {
		return fmt.Errorf("marcar el episodio ya importado: %w", err)
	}

	b.reclamadas[creada.ID] = true
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
func (b *busqueda) crearFicha(proveedor Proveedor, entrada EntradaFeed) error {
	nueva := fichaDe(entrada, proveedor, VisibilidadVisible)
	creada, err := b.servicio.repository.InsertMediaItem(b.ctx, nueva)
	if err != nil {
		return fmt.Errorf("crear el episodio: %w", err)
	}
	if _, ok, err := b.insertarFuente(proveedor, entrada, creada.ID); err != nil {
		return err
	} else if !ok {
		return nil
	}
	b.elementos++
	b.resumen.Importados++
	return nil
}

// insertarFuente guarda la procedencia del elemento con `ON CONFLICT (provider,
// external_id) DO NOTHING` (R9-11): el segundo valor es false cuando la base
// decidió que el elemento ya estaba importado (dos búsquedas simultáneas). Si el
// guardado falla, la ficha se oculta en mejor esfuerzo: una ficha sin enlaces no
// se publica (US1 esc. 3).
func (b *busqueda) insertarFuente(proveedor Proveedor, entrada EntradaFeed, episodioID uuid.UUID) (FuenteEpisodio, bool, error) {
	fuente := FuenteEpisodio{
		EpisodioID: episodioID,
		Proveedor:  proveedor,
		IDExterno:  entrada.IDExterno,
		TituloFeed: recortarTitulo(entrada.Titulo),
		Enlace:     entrada.Enlace,
	}
	insertada, ok, err := b.servicio.repository.InsertMediaEpisodeSource(b.ctx, fuente)
	if err != nil {
		return FuenteEpisodio{}, false, fmt.Errorf("importar la fuente del episodio: %w", err)
	}
	if ok {
		b.importados[fuente.Proveedor][fuente.IDExterno] = true
		b.conFuente[fuente.Proveedor][episodioID] = true
		// La lista de fuentes del proveedor se mantiene al día: el segundo feed
		// de la búsqueda encuentra así las fichas que acaba de crear el primero.
		b.fuentes[fuente.Proveedor] = append(b.fuentes[fuente.Proveedor], insertada)
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
