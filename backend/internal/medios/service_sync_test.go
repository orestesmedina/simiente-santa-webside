package medios

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// service_sync_test.go cubre la búsqueda a demanda (T920/T921): la importación
// idempotente desde los dos feeds, el recorte de la descripción corta, la unión
// automática del mismo sermón, el marcado de lo dudoso, la ventana del feed, la
// disponibilidad a prueba de fallos, las carátulas y el resumen de FR-018.
//
// Todas las pruebas usan **dobles** de `FeedSource`, `AvailabilityProber`,
// `CoverFetcher` y del repositorio: ninguna sale a la red.

// ---------------------------------------------------------------------------
// Datos reales del cliente (spec.md §Datos de origen)
// ---------------------------------------------------------------------------

// El set real de los dos feeds (verificado el 2026-10-17): 9 elementos en cada
// uno, **8 en común** (mismo sermón, mismo día en America/Costa_Rica) y uno
// exclusivo de cada feed. La primera búsqueda debe producir **10 fichas**
// (8 unidas + 2 con un solo enlace) y **0 marcas**.
var (
	sermonesReales = []struct {
		youtubeTitulo string
		podcastTitulo string
		youtubeFecha  time.Time
		podcastFecha  time.Time
	}{
		{
			youtubeTitulo: "La Fe y la confianza en el Abba",
			podcastTitulo: "La Fe y la confianza en el Abba",
			youtubeFecha:  time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 7, 3, 1, 22, 48, 0, time.UTC),
		},
		{
			youtubeTitulo: "La Fe y la herida. Deja de esconderte: atrévete a sanar tus heridas hoy",
			podcastTitulo: "La Fe y la herida. Deja de esconderte: atrévete a sanar tus heridas hoy",
			youtubeFecha:  time.Date(2026, 6, 27, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 6, 27, 18, 24, 49, 0, time.UTC),
		},
		{
			youtubeTitulo: "La Fe y la Actitud Parte II: Rompiendo la Fábrica de Excusas",
			podcastTitulo: "La Fe y la Actitud Parte II: Rompiendo la Fábrica de Excusas",
			youtubeFecha:  time.Date(2026, 6, 20, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 6, 20, 19, 23, 3, 0, time.UTC),
		},
		{
			youtubeTitulo: "La Fe y la Actitud: Por Qué Tu Mentalidad lo Cambia Todo",
			podcastTitulo: "La Fe y la Actitud: Por Qué Tu Mentalidad lo Cambia Todo",
			youtubeFecha:  time.Date(2026, 6, 10, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 6, 10, 19, 29, 55, 0, time.UTC),
		},
		{
			youtubeTitulo: "La Fe y La Impaciencia: El Peligro de No Saber Esperar",
			podcastTitulo: "La Fe y La Impaciencia: El Peligro de No Saber Esperar",
			youtubeFecha:  time.Date(2026, 6, 6, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 6, 6, 17, 44, 57, 0, time.UTC),
		},
		{
			// El par real cuyos títulos NO coinciden carácter por carácter: la
			// llave normalizada sí (FR-013).
			youtubeTitulo: "Potencial sin Resultados La Lección de la Higuera",
			podcastTitulo: "Potencial sin Resultados: La Lección de la Higuera",
			youtubeFecha:  time.Date(2026, 5, 25, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 5, 25, 17, 58, 50, 0, time.UTC),
		},
		{
			youtubeTitulo: "El Síndrome del Microondas: Por qué tu fe no está dando resultados",
			podcastTitulo: "El Síndrome del Microondas: Por qué tu fe no está dando resultados",
			youtubeFecha:  time.Date(2026, 5, 19, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 5, 20, 3, 42, 27, 0, time.UTC),
		},
		{
			youtubeTitulo: "La paternidad de Dios y la fe",
			podcastTitulo: "La paternidad de Dios y la fe",
			youtubeFecha:  time.Date(2026, 4, 19, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 4, 19, 22, 42, 59, 0, time.UTC),
		},
	}
	youtubeExclusivo = entradaDePrueba(ProveedorYouTube, "SimienteAA9", "A Mayor Fe, Mayor Presión",
		time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC), "Sermón sobre la presión de la fe.")
	podcastExclusivo = entradaDePrueba(ProveedorPodcast, "guid-09", "Fe como fruto del Espíritu",
		time.Date(2026, 4, 13, 16, 34, 8, 0, time.UTC), "Episodio sobre el fruto del Espíritu.")
)

// feedRealYouTube y feedRealPodcast arman los dos feeds del cliente.
func feedRealYouTube() ResultadoFeed {
	entradas := make([]EntradaFeed, 0, len(sermonesReales)+1)
	for i, sermon := range sermonesReales {
		entradas = append(entradas, entradaDePrueba(ProveedorYouTube,
			fmt.Sprintf("SimienteAA%d", i+1), sermon.youtubeTitulo, sermon.youtubeFecha,
			"Descripción del video "+sermon.youtubeTitulo))
	}
	return ResultadoFeed{Entradas: append(entradas, youtubeExclusivo)}
}

func feedRealPodcast() ResultadoFeed {
	entradas := make([]EntradaFeed, 0, len(sermonesReales)+1)
	for i, sermon := range sermonesReales {
		entradas = append(entradas, entradaDePrueba(ProveedorPodcast,
			fmt.Sprintf("guid-%02d", i+1), sermon.podcastTitulo, sermon.podcastFecha,
			"<p>Descripción del episodio "+sermon.podcastTitulo+"</p>"))
	}
	return ResultadoFeed{Entradas: append(entradas, podcastExclusivo)}
}

// entradaDePrueba arma un elemento de feed con el enlace del proveedor.
func entradaDePrueba(proveedor Proveedor, id, titulo string, publicado time.Time, descripcion string) EntradaFeed {
	enlace := "https://www.youtube.com/watch?v=" + id
	if proveedor == ProveedorPodcast {
		enlace = "https://podcasters.spotify.com/pod/show/simiente/episodes/" + id
	}
	return EntradaFeed{
		IDExterno:   id,
		Titulo:      titulo,
		Enlace:      enlace,
		Publicado:   publicado,
		Descripcion: descripcion,
	}
}

// ---------------------------------------------------------------------------
// Importación (T920)
// ---------------------------------------------------------------------------

// TestBuscarNuevosImportaElSetReal fija el resultado de referencia de la spec:
// con los feeds del cliente (9 + 9, 8 en común) la primera búsqueda produce
// **10 fichas** —8 unidas, con los dos enlaces, y 2 con un solo enlace—, **0
// marcas** y 0 fichas ocultas, sin que el equipo escriba ningún dato (SC-004).
func TestBuscarNuevosImportaElSetReal(t *testing.T) {
	repo := newFakeRepository()
	service := servicioConFeeds(t, repo, feedRealYouTube(), feedRealPodcast())

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if resumen.Importados != 10 || resumen.Unidos != 8 || resumen.Marcados != 0 || resumen.Ocultos != 0 {
		t.Errorf("resumen = %+v, se esperaban 10 importados, 8 unidos, 0 marcados y 0 ocultos", resumen)
	}
	if resumen.EstadoYouTube != string(FeedOK) || resumen.EstadoPodcast != string(FeedOK) {
		t.Errorf("estados = %q/%q, se esperaban ok/ok", resumen.EstadoYouTube, resumen.EstadoPodcast)
	}
	if len(repo.episodios) != 10 {
		t.Fatalf("se crearon %d fichas, se esperaban 10", len(repo.episodios))
	}
	if len(repo.fuentes) != 18 {
		t.Errorf("se importaron %d fuentes, se esperaban 18 (9 + 9)", len(repo.fuentes))
	}
	for _, ficha := range repo.episodios {
		if ficha.Visibilidad != VisibilidadVisible || ficha.DuplicadoDeID != nil {
			t.Errorf("la primera búsqueda no marca ni oculta nada: %+v", ficha)
		}
		switch ficha.Origen {
		case OrigenAmbos:
			if len(fuentesDe(repo, ficha.ID)) != 2 {
				t.Errorf("una ficha unida debe tener dos fuentes: %s", ficha.Titulo)
			}
		case OrigenVideo, OrigenAudio:
			if len(fuentesDe(repo, ficha.ID)) != 1 {
				t.Errorf("una ficha con un solo enlace debe tener una fuente: %s", ficha.Titulo)
			}
		default:
			t.Errorf("origen inesperado %q en %s", ficha.Origen, ficha.Titulo)
		}
	}

	// El visitante ve las 10, con los enlaces que la unión juntó (SC-004).
	publicos, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	if publicos.Total != 10 {
		t.Errorf("el público ve %d fichas, se esperaban 10", publicos.Total)
	}
	sinEnlace := 0
	for _, item := range publicos.Items {
		if item.VideoURL == "" && item.AudioURL == "" {
			t.Errorf("la ficha %q salió sin ningún enlace", item.Titulo)
		}
		if item.VideoURL == "" || item.AudioURL == "" {
			sinEnlace++
		}
	}
	if sinEnlace != 2 {
		t.Errorf("%d fichas con un solo enlace, se esperaban 2 (las exclusivas)", sinEnlace)
	}
}

// TestBuscarNuevosEsIdempotente fija FR-012/SC-003: repetir la búsqueda no
// importa ni duplica nada, no refresca la instantánea (R9-6: ni el título ni la
// fecha ni la carátula cambian) y la base es quien decide ante una carrera
// (`ON CONFLICT (provider, external_id) DO NOTHING`).
func TestBuscarNuevosEsIdempotente(t *testing.T) {
	repo := newFakeRepository()
	feeds := &fakeFeedSource{youtube: feedRealYouTube(), podcast: feedRealPodcast()}
	service := servicioConDobles(t, repo, feeds)

	if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
		t.Fatalf("la primera búsqueda devolvió %v", err)
	}
	fichasAntes := len(repo.episodios)
	fuentesAntes := len(repo.fuentes)

	// El feed cambia por completo (título, fecha y descripción): la instantánea
	// importada NO se refresca (R9-6).
	feeds.youtube = ResultadoFeed{Entradas: []EntradaFeed{{
		IDExterno: "SimienteAA1", Titulo: "Título cambiado en YouTube",
		Enlace:    "https://www.youtube.com/watch?v=SimienteAA1",
		Publicado: time.Date(2026, 12, 31, 15, 0, 0, 0, time.UTC),
	}}}
	feeds.podcast = ResultadoFeed{Entradas: []EntradaFeed{{
		IDExterno: "guid-01", Titulo: "Título cambiado en el podcast",
		Enlace:    "https://podcasters.spotify.com/pod/show/simiente/episodes/guid-01",
		Publicado: time.Date(2026, 12, 31, 15, 0, 0, 0, time.UTC),
	}}}

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("la segunda búsqueda devolvió %v", err)
	}
	if resumen.Importados != 0 || resumen.Unidos != 0 || resumen.Marcados != 0 {
		t.Errorf("repetir la búsqueda no cambia nada: %+v", resumen)
	}
	if len(repo.episodios) != fichasAntes || len(repo.fuentes) != fuentesAntes {
		t.Errorf("la búsqueda duplicó: %d fichas (antes %d) y %d fuentes (antes %d)",
			len(repo.episodios), fichasAntes, len(repo.fuentes), fuentesAntes)
	}
	for _, ficha := range repo.episodios {
		if strings.Contains(ficha.Titulo, "cambiado") {
			t.Errorf("la instantánea se refrescó: %q", ficha.Titulo)
		}
		if ficha.PublicadoEn.Year() == 2026 && ficha.PublicadoEn.Month() == time.December {
			t.Errorf("la fecha se refrescó: %v", ficha.PublicadoEn)
		}
	}

	// Carrera de dos búsquedas: el UNIQUE decide (una sola fuente por elemento).
	insertada, ok, err := repo.InsertMediaEpisodeSource(context.Background(), FuenteEpisodio{
		EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
		TituloFeed: "Otra búsqueda", Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
	})
	if err != nil {
		t.Fatalf("el ON CONFLICT no debe fallar: %v", err)
	}
	if ok || insertada.ID != uuid.Nil {
		t.Error("el elemento ya importado no debe volver a insertarse (R9-11)")
	}
	if len(repo.fuentes) != fuentesAntes {
		t.Errorf("la carrera duplicó la fuente: %d", len(repo.fuentes))
	}
}

// TestBuscarNuevosRecortaLaDescripcionCorta fija P9-17/FR-020: la descripción se
// recorta al importar (HTML → texto plano, ≤ 300 caracteres, corte en frontera de
// frase o palabra con «…»), un feed sin descripción deja el campo **ausente**
// (precisión 3) y una descripción ajustada por el equipo **nunca se pisa**
// (FR-015/SC-007).
func TestBuscarNuevosRecortaLaDescripcionCorta(t *testing.T) {
	larga := strings.Repeat("Frase completa de la prédica. ", 60) // ~1.800 caracteres
	repo := newFakeRepository()
	feeds := &fakeFeedSource{
		youtube: ResultadoFeed{Entradas: []EntradaFeed{
			entradaDePrueba(ProveedorYouTube, "SimienteAA1", "Con descripción larga",
				time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC), larga),
			entradaDePrueba(ProveedorYouTube, "SimienteAA2", "Sin descripción",
				time.Date(2026, 7, 1, 15, 0, 0, 0, time.UTC), "   "),
		}},
		podcast: ResultadoFeed{Entradas: []EntradaFeed{
			entradaDePrueba(ProveedorPodcast, "guid-01", "Con HTML del podcast",
				time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC),
				"<p>Primera frase &amp; algo más.</p><p><strong>Segunda</strong> frase.</p>"),
		}},
	}
	service := servicioConDobles(t, repo, feeds)

	if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}

	porTitulo := map[string]Episodio{}
	for _, ficha := range repo.episodios {
		porTitulo[ficha.Titulo] = ficha
	}

	recortada := porTitulo["Con descripción larga"].DescripcionEs
	if recortada == nil {
		t.Fatal("la descripción larga debía recortarse, no quedar ausente")
	}
	if len([]rune(*recortada)) > maxDescripcion {
		t.Errorf("el recorte mide %d caracteres", len([]rune(*recortada)))
	}
	if !strings.HasSuffix(*recortada, "…") {
		t.Errorf("el recorte debe marcar el corte con «…»: %q", *recortada)
	}
	if !strings.HasSuffix(*recortada, "prédica.…") && !strings.Contains(*recortada, ". …") {
		t.Logf("recorte = %q", *recortada)
	}

	if porTitulo["Sin descripción"].DescripcionEs != nil {
		t.Errorf("sin descripción en el feed el campo queda ausente: %q", *porTitulo["Sin descripción"].DescripcionEs)
	}

	conHTML := porTitulo["Con HTML del podcast"].DescripcionEs
	if conHTML == nil {
		t.Fatal("el episodio del podcast debía traer descripción")
	}
	if strings.ContainsAny(*conHTML, "<>") {
		t.Errorf("el HTML del feed no se publica: %q", *conHTML)
	}
	if !strings.Contains(*conHTML, "Primera frase & algo más.") || !strings.Contains(*conHTML, "Segunda frase.") {
		t.Errorf("el texto plano perdió contenido: %q", *conHTML)
	}

	// La descripción del equipo no se pisa: se marca como editada y la búsqueda
	// la respeta (el elemento ya está importado).
	repo.episodios[0].DescripcionEs = ptr("Del equipo")
	repo.episodios[0].DescripcionEditada = true
	if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
		t.Fatalf("la segunda búsqueda devolvió %v", err)
	}
	for _, ficha := range repo.episodios {
		if ficha.DescripcionEditada && ficha.DescripcionEs != nil && *ficha.DescripcionEs != "Del equipo" {
			t.Errorf("la búsqueda pisó la descripción del equipo: %q", *ficha.DescripcionEs)
		}
	}
}

// TestBuscarNuevosUnionAutomatica fija FR-013/precisión 5: el mismo sermón en los
// dos feeds (títulos que coinciden al normalizar y el mismo día en
// America/Costa_Rica) queda en **una sola ficha** con los dos enlaces, el nombre,
// la carátula y la fecha del **video**, y cuenta como una importación y una
// unión.
func TestBuscarNuevosUnionAutomatica(t *testing.T) {
	repo := newFakeRepository()
	feeds := &fakeFeedSource{
		youtube: ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(ProveedorYouTube,
			"SimienteAA1", "Potencial sin Resultados La Lección de la Higuera",
			time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC), "Del video")}},
		podcast: ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(ProveedorPodcast,
			"guid-01", "Potencial sin Resultados: La Lección de la Higuera",
			time.Date(2026, 7, 3, 1, 22, 48, 0, time.UTC), "Del podcast")}},
	}
	service := servicioConDobles(t, repo, feeds)

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if resumen.Importados != 1 || resumen.Unidos != 1 {
		t.Errorf("resumen = %+v, se esperaba 1 importado y 1 unido", resumen)
	}
	if len(repo.episodios) != 1 || len(repo.fuentes) != 2 {
		t.Fatalf("la unión dejó %d fichas y %d fuentes", len(repo.episodios), len(repo.fuentes))
	}
	ficha := repo.episodios[0]
	if ficha.Titulo != "Potencial sin Resultados La Lección de la Higuera" {
		t.Errorf("el nombre que manda es el del video: %q", ficha.Titulo)
	}
	if ficha.Origen != OrigenAmbos {
		t.Errorf("origin = %q, se esperaba both", ficha.Origen)
	}
	if !ficha.PublicadoEn.Equal(time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)) {
		t.Errorf("la fecha que manda es la del video: %v", ficha.PublicadoEn)
	}
	if ficha.Visibilidad != VisibilidadVisible || ficha.DuplicadoDeID != nil {
		t.Errorf("la unión automática no oculta ni marca: %+v", ficha)
	}

	publicos, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	if len(publicos.Items) != 1 {
		t.Fatalf("el público ve %d fichas, se esperaba 1", len(publicos.Items))
	}
	if publicos.Items[0].VideoURL == "" || publicos.Items[0].AudioURL == "" {
		t.Errorf("la ficha unida debe traer los dos enlaces: %+v", publicos.Items[0])
	}
}

// TestBuscarNuevosDuplicadoDudoso fija FR-014/default #2: cuando la
// correspondencia no es segura (mismo día con títulos distintos, o el mismo
// título con días distintos) quedan **entradas separadas**, la adicional
// **oculta** y **las dos** marcadas «posible duplicado» con el par recíproco
// abierto; el visitante nunca ve el mismo sermón dos veces (SC-003).
func TestBuscarNuevosDuplicadoDudoso(t *testing.T) {
	casos := []struct {
		name          string
		youtubeTitulo string
		podcastTitulo string
		youtubeFecha  time.Time
		podcastFecha  time.Time
	}{
		{
			name:          "mismo día con títulos distintos",
			youtubeTitulo: "Sermón del domingo",
			podcastTitulo: "Otra cosa distinta",
			youtubeFecha:  time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 7, 2, 20, 0, 0, 0, time.UTC),
		},
		{
			name:          "mismo título con días distintos",
			youtubeTitulo: "Sermón del domingo",
			podcastTitulo: "Sermón del domingo",
			youtubeFecha:  time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
			podcastFecha:  time.Date(2026, 7, 5, 15, 0, 0, 0, time.UTC),
		},
	}
	for _, caso := range casos {
		t.Run(caso.name, func(t *testing.T) {
			repo := newFakeRepository()
			feeds := &fakeFeedSource{
				youtube: ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(
					ProveedorYouTube, "SimienteAA1", caso.youtubeTitulo, caso.youtubeFecha, "Del video")}},
				podcast: ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(
					ProveedorPodcast, "guid-01", caso.podcastTitulo, caso.podcastFecha, "Del podcast")}},
			}
			service := servicioConDobles(t, repo, feeds)

			resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
			if err != nil {
				t.Fatalf("BuscarNuevos devolvió %v", err)
			}
			if resumen.Importados != 2 || resumen.Unidos != 0 || resumen.Marcados != 2 {
				t.Errorf("resumen = %+v, se esperaban 2 importados, 0 unidos y 2 marcados", resumen)
			}
			if len(repo.episodios) != 2 {
				t.Fatalf("quedaron %d fichas, se esperaban 2 separadas", len(repo.episodios))
			}
			primera, segunda := repo.episodios[0], repo.episodios[1]
			if primera.Visibilidad != VisibilidadVisible {
				t.Errorf("la primera conserva su visibilidad: %+v", primera)
			}
			if segunda.Visibilidad != VisibilidadOculta {
				t.Errorf("la adicional queda oculta: %+v", segunda)
			}
			// Pareja recíproca y abierta (invariante de service).
			if primera.DuplicadoDeID == nil || *primera.DuplicadoDeID != segunda.ID {
				t.Errorf("la primera no apunta a la segunda: %v", primera.DuplicadoDeID)
			}
			if segunda.DuplicadoDeID == nil || *segunda.DuplicadoDeID != primera.ID {
				t.Errorf("la segunda no apunta a la primera: %v", segunda.DuplicadoDeID)
			}
			if primera.MarcaDuplicadoResueltaEn != nil || segunda.MarcaDuplicadoResueltaEn != nil {
				t.Error("las marcas del par nacen abiertas")
			}
			// El visitante no ve el sermón dos veces (SC-003): con el par
			// ABIERTO ninguna de las dos fichas sale (R9-7: el par entero queda
			// fuera del público; la primera conserva su flag `visible` para
			// publicarse cuando el equipo resuelva «no son duplicados»).
			publicos, err := service.GetMedios(context.Background(), LangES, 0, 0)
			if err != nil {
				t.Fatalf("GetMedios devolvió %v", err)
			}
			if len(publicos.Items) != 0 {
				t.Errorf("el par abierto no puede salir al público: %v", titulosPublicos(publicos.Items))
			}
		})
	}
}

// TestBuscarNuevosSinCandidatura fija US2 esc. 4: un elemento que solo está en un
// feed entra como entrada con **un solo enlace**, visible y sin marca.
func TestBuscarNuevosSinCandidatura(t *testing.T) {
	repo := newFakeRepository()
	feeds := &fakeFeedSource{
		youtube: ResultadoFeed{Entradas: []EntradaFeed{youtubeExclusivo}},
		podcast: ResultadoFeed{Entradas: []EntradaFeed{podcastExclusivo}},
	}
	service := servicioConDobles(t, repo, feeds)

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if resumen.Importados != 2 || resumen.Unidos != 0 || resumen.Marcados != 0 {
		t.Errorf("resumen = %+v", resumen)
	}
	for _, ficha := range repo.episodios {
		if ficha.Visibilidad != VisibilidadVisible || ficha.DuplicadoDeID != nil {
			t.Errorf("sin candidatura no hay marca ni ocultación: %+v", ficha)
		}
	}
}

// TestBuscarNuevosFeedCaido fija FR-016: si uno de los feeds no responde se
// importa lo del otro, el resumen dice **cuál** falló y **no se oculta ni se
// marca nada** por esa causa (un feed caído no significa que desaparecieran
// episodios).
func TestBuscarNuevosFeedCaido(t *testing.T) {
	repo := newFakeRepository()
	feeds := &fakeFeedSource{
		youtube:    ResultadoFeed{Entradas: []EntradaFeed{youtubeExclusivo}},
		errPodcast: errFeedNoDisponible,
	}
	service := servicioConDobles(t, repo, feeds)

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("un feed caído no es un error de la búsqueda: %v", err)
	}
	if resumen.EstadoPodcast != string(FeedFallido) || resumen.EstadoYouTube != string(FeedOK) {
		t.Errorf("estados = %q/%q, se esperaba ok/failed", resumen.EstadoYouTube, resumen.EstadoPodcast)
	}
	if resumen.Importados != 1 || resumen.Ocultos != 0 || resumen.Marcados != 0 {
		t.Errorf("resumen = %+v, se esperaba solo lo del feed que respondió", resumen)
	}
	if len(repo.episodios) != 1 {
		t.Errorf("se importó %d fichas, se esperaba 1", len(repo.episodios))
	}

	// Los dos feeds caídos: nada se importa, nada se toca (US2 esc. 6).
	repo2 := newFakeRepository()
	repo2.seed(Episodio{Titulo: "Lo que ya había", PublicadoEn: time.Now()})
	feeds2 := &fakeFeedSource{errYoutube: errFeedNoDisponible, errPodcast: errFeedNoDisponible}
	service2 := servicioConDobles(t, repo2, feeds2)

	resumen2, err := service2.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if resumen2.EstadoYouTube != string(FeedFallido) || resumen2.EstadoPodcast != string(FeedFallido) {
		t.Errorf("estados = %q/%q, se esperaban failed/failed", resumen2.EstadoYouTube, resumen2.EstadoPodcast)
	}
	if resumen2.Importados != 0 || len(repo2.episodios) != 1 || repo2.episodios[0].Visibilidad != VisibilidadVisible {
		t.Errorf("con los dos feeds caídos nada cambia: %+v", resumen2)
	}
}

// TestBuscarNuevosNoPisaLoDelEquipo fija FR-015/SC-007: la búsqueda **nunca**
// cambia la visibilidad decidida por el equipo, ni la descripción ajustada, ni
// deshace una unión confirmada.
func TestBuscarNuevosNoPisaLoDelEquipo(t *testing.T) {
	repo := newFakeRepository()
	oculta := Episodio{Titulo: "Ocultada por el equipo", PublicadoEn: time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
		Visibilidad: VisibilidadOculta}
	editada := Episodio{Titulo: "Con descripción del equipo", PublicadoEn: time.Date(2026, 7, 1, 15, 0, 0, 0, time.UTC),
		DescripcionEs: ptr("Del equipo"), DescripcionEditada: true}
	repo.seed(oculta, editada)
	repo.seedFuentes(
		FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
			TituloFeed: oculta.Titulo, Enlace: "https://www.youtube.com/watch?v=SimienteAA1"},
		FuenteEpisodio{EpisodioID: repo.episodios[1].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA2",
			TituloFeed: editada.Titulo, Enlace: "https://www.youtube.com/watch?v=SimienteAA2"},
	)
	feeds := &fakeFeedSource{youtube: ResultadoFeed{Entradas: []EntradaFeed{
		entradaDePrueba(ProveedorYouTube, "SimienteAA1", oculta.Titulo, oculta.PublicadoEn, "Otra descripción"),
		entradaDePrueba(ProveedorYouTube, "SimienteAA2", editada.Titulo, editada.PublicadoEn, "Otra descripción"),
	}}}
	service := servicioConDobles(t, repo, feeds)

	if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if repo.episodios[0].Visibilidad != VisibilidadOculta {
		t.Error("un episodio oculto sigue oculto tras la búsqueda (FR-015)")
	}
	if repo.episodios[1].DescripcionEs == nil || *repo.episodios[1].DescripcionEs != "Del equipo" {
		t.Error("la descripción del equipo no se pisa (SC-007)")
	}
	if !repo.episodios[1].DescripcionEditada {
		t.Error("descriptionEdited debe seguir marcada")
	}

	// Una unión confirmada no se deshace: la absorbida conserva el puntero y la
	// resultante sus dos fuentes.
	repo2 := newFakeRepository()
	videoID, podcastID := sembrarVideoYAudio(repo2, time.Date(2026, 5, 25, 15, 0, 0, 0, time.UTC))
	service2 := servicioConDobles(t, repo2, &fakeFeedSource{
		youtube: ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(ProveedorYouTube,
			"SimienteAA1", "Potencial sin Resultados La Lección de la Higuera",
			time.Date(2026, 5, 25, 15, 0, 0, 0, time.UTC), "Del video")}},
		podcast: ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(ProveedorPodcast,
			"guid-podcast", "Potencial sin Resultados: La Lección de la Higuera",
			time.Date(2026, 5, 25, 17, 58, 50, 0, time.UTC), "Del podcast")}},
	})
	if _, err := service2.Merge(context.Background(), uuid.New(), podcastID, videoID); err != nil {
		t.Fatalf("Merge devolvió %v", err)
	}
	if _, err := service2.BuscarNuevos(context.Background(), uuid.New()); err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	resultante := repo2.episodios[1]
	if resultante.UnidoAID != nil || len(fuentesDe(repo2, resultante.ID)) != 2 {
		t.Errorf("la unión confirmada se deshizo: %+v", resultante)
	}
	absorbida := repo2.episodios[0]
	if absorbida.UnidoAID == nil || *absorbida.UnidoAID != resultante.ID {
		t.Errorf("la absorbida perdió su puntero: %+v", absorbida.UnidoAID)
	}
}

// TestBuscarNuevosTopeDuro fija el tope de P9-3 (30 elementos por búsqueda): el
// resto se recupera en la siguiente búsqueda, sin perder ni ocultar nada.
func TestBuscarNuevosTopeDuro(t *testing.T) {
	repo := newFakeRepository()
	entradas := make([]EntradaFeed, 0, maxElementosPorBusqueda+10)
	for i := 0; i < maxElementosPorBusqueda+10; i++ {
		entradas = append(entradas, entradaDePrueba(ProveedorYouTube,
			fmt.Sprintf("Simiente%03d", i), fmt.Sprintf("Sermón %d", i),
			time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC).Add(-time.Duration(i)*time.Hour), "Descripción"))
	}
	feeds := &fakeFeedSource{youtube: ResultadoFeed{Entradas: entradas}}
	service := servicioConDobles(t, repo, feeds)

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if resumen.Importados != maxElementosPorBusqueda {
		t.Errorf("importados = %d, se esperaba el tope %d", resumen.Importados, maxElementosPorBusqueda)
	}
	if len(repo.episodios) != maxElementosPorBusqueda {
		t.Errorf("fichas = %d, se esperaban %d", len(repo.episodios), maxElementosPorBusqueda)
	}

	// La búsqueda siguiente recupera lo que faltaba (FR-016).
	resumen2, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("la segunda búsqueda devolvió %v", err)
	}
	if resumen2.Importados != 10 {
		t.Errorf("la segunda búsqueda importó %d, se esperaban 10", resumen2.Importados)
	}
	if len(repo.episodios) != maxElementosPorBusqueda+10 {
		t.Errorf("fichas = %d, se esperaban %d", len(repo.episodios), maxElementosPorBusqueda+10)
	}
}

// ---------------------------------------------------------------------------
// Helpers de prueba de la búsqueda
// ---------------------------------------------------------------------------

// servicioConFeeds arma el service con los dos feeds indicados y la zona de la
// iglesia (APP_TZ).
func servicioConFeeds(t *testing.T, repo *fakeRepository, youtube, podcast ResultadoFeed) *service {
	t.Helper()
	return servicioConDobles(t, repo, &fakeFeedSource{youtube: youtube, podcast: podcast})
}

// servicioConDobles arma el service con los dobles de la red saliente y la zona
// del lugar (America/Costa_Rica: la llave del «mismo día» de FR-013).
func servicioConDobles(t *testing.T, repo *fakeRepository, feeds FeedSource) *service {
	t.Helper()
	return NewService(ServiceDeps{
		Repository:         repo,
		FeedSource:         feeds,
		AvailabilityProber: &fakeProbador{},
		CoverFetcher:       &fakeCaratulas{},
		Store:              &fakeAlmacen{},
		Location:           zonaDePrueba(t),
	})
}

// zonaDePrueba carga la zona de la iglesia (APP_TZ). Sin datos de zona horaria
// la prueba se omite: el «mismo día» de FR-013 no se puede comprobar.
func zonaDePrueba(t *testing.T) *time.Location {
	t.Helper()
	zona, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Skipf("sin datos de zona horaria (%v): se omite la prueba", err)
	}
	return zona
}

// fuentesDe devuelve las fuentes de una ficha (como el `LEFT JOIN` por proveedor).
func fuentesDe(repo *fakeRepository, id uuid.UUID) []FuenteEpisodio {
	fuentes := make([]FuenteEpisodio, 0, 2)
	for _, fuente := range repo.fuentes {
		if fuente.EpisodioID == id {
			fuentes = append(fuentes, fuente)
		}
	}
	return fuentes
}
