package medios

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
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

// TestBuscarNuevosParejaReciprocaUnica fija la invariante de la pareja recíproca
// (data-model.md/M1): una ficha que **ya tiene pareja** —abierta o resuelta— no
// vuelve a entrar en un par nuevo, ni dentro de la misma búsqueda (dos elementos
// del mismo feed no se marcan contra el mismo candidato) ni entre búsquedas
// simultáneas (la marca del par escribe con `duplicate_of_id IS NULL`).
//
// La exclusión vale **solo para la rama dudosa** —la que escribe el puntero
// recíproco—: la unión automática segura no toca `duplicate_of_id` y sigue
// disponible para una ficha emparejada (FR-013/N1), que es lo que evita que el
// mismo sermón quede visible dos veces (SC-003).
func TestBuscarNuevosParejaReciprocaUnica(t *testing.T) {
	fecha := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("dos elementos del mismo feed no comparten candidato", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Sermón del domingo", PublicadoEn: fecha, Origen: OrigenVideo})
		candidato := repo.episodios[0].ID
		repo.seedFuentes(FuenteEpisodio{
			EpisodioID: candidato, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
			TituloFeed: "Sermón del domingo", Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
		})
		// Los dos elementos coinciden en título con el candidato y difieren en el
		// día: cada uno sería «dudoso» contra él.
		entradas := []EntradaFeed{
			entradaDePrueba(ProveedorPodcast, "guid-01", "Sermón del domingo", fecha.AddDate(0, 0, 3), "Uno"),
			entradaDePrueba(ProveedorPodcast, "guid-02", "Sermón del domingo", fecha.AddDate(0, 0, 4), "Dos"),
		}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{podcast: ResultadoFeed{Entradas: entradas}}, &fakeProbador{}, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if resumen.Marcados != 2 {
			t.Errorf("marcados = %d, se esperaba un solo par (2 fichas)", resumen.Marcados)
		}
		if len(repo.episodios) != 3 {
			t.Fatalf("fichas = %d, se esperaban el candidato, su par y una entrada suelta", len(repo.episodios))
		}
		primera, adicional := repo.episodios[1], repo.episodios[2]
		if repo.episodios[0].DuplicadoDeID == nil || *repo.episodios[0].DuplicadoDeID != primera.ID {
			t.Errorf("el candidato debe apuntar a la PRIMERA adicional: %v", repo.episodios[0].DuplicadoDeID)
		}
		if primera.DuplicadoDeID == nil || *primera.DuplicadoDeID != candidato {
			t.Errorf("la primera adicional apunta al candidato: %v", primera.DuplicadoDeID)
		}
		if adicional.DuplicadoDeID != nil {
			t.Errorf("el segundo elemento no puede entrar en el par ya tomado: %+v", adicional)
		}
		if adicional.Visibilidad != VisibilidadVisible {
			t.Errorf("sin pareja el elemento entra visible: %q", adicional.Visibilidad)
		}
	})

	t.Run("un match seguro con una ficha emparejada se une (FR-013/N1)", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(
			Episodio{Titulo: "Sermón del domingo", PublicadoEn: fecha, Origen: OrigenVideo},
			Episodio{Titulo: "Otro sermón del mismo día", PublicadoEn: fecha, Origen: OrigenAudio},
		)
		// Un par ya existente **resuelto** («no son el mismo sermón»): cada
		// miembro conserva el puntero (default #6) y sigue siendo publicable.
		resuelto := fecha.AddDate(0, 0, 1)
		repo.episodios[0].DuplicadoDeID = ptrUUID(repo.episodios[1].ID)
		repo.episodios[0].MarcaDuplicadoResueltaEn = &resuelto
		repo.episodios[1].DuplicadoDeID = ptrUUID(repo.episodios[0].ID)
		repo.episodios[1].MarcaDuplicadoResueltaEn = &resuelto
		repo.seedFuentes(FuenteEpisodio{
			EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
			TituloFeed: "Sermón del domingo", Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
		})
		// El podcast publica (o re-publica) el audio del **mismo** sermón: título
		// normalizado y día coinciden, así que la correspondencia es segura.
		entrada := entradaDePrueba(ProveedorPodcast, "guid-01", "Sermón del domingo", fecha, "Del podcast")
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{podcast: ResultadoFeed{Entradas: []EntradaFeed{entrada}}}, &fakeProbador{}, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if resumen.Unidos != 1 || resumen.Marcados != 0 {
			t.Errorf("el match seguro debe unirse sin marcar: %+v", resumen)
		}
		if len(repo.episodios) != 2 {
			t.Fatalf("fichas = %d, la unión no puede crear una ficha aparte (SC-003)", len(repo.episodios))
		}
		unida := repo.episodios[0]
		if len(fuentesDe(repo, unida.ID)) != 2 {
			t.Errorf("la ficha unida debe tener los dos enlaces: %d fuentes", len(fuentesDe(repo, unida.ID)))
		}
		if unida.Origen != OrigenAmbos {
			t.Errorf("origin = %q, se esperaba both", unida.Origen)
		}
		// La unión es inocua para la pareja recíproca: no la rompe ni la reabre.
		if unida.DuplicadoDeID == nil || *unida.DuplicadoDeID != repo.episodios[1].ID {
			t.Errorf("la unión pisó el puntero de la pareja: %v", unida.DuplicadoDeID)
		}
		if repo.episodios[1].DuplicadoDeID == nil || *repo.episodios[1].DuplicadoDeID != unida.ID {
			t.Errorf("la pareja recíproca se rompió: %v", repo.episodios[1].DuplicadoDeID)
		}
		if unida.MarcaDuplicadoResueltaEn == nil || repo.episodios[1].MarcaDuplicadoResueltaEn == nil {
			t.Error("la unión reabrió la resolución del par")
		}
	})

	t.Run("un match solo por título no roba el puntero de una ficha emparejada", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(
			Episodio{Titulo: "Sermón del domingo", PublicadoEn: fecha, Origen: OrigenVideo},
			Episodio{Titulo: "Su pareja", PublicadoEn: fecha, Origen: OrigenAudio},
		)
		// Un par ya existente (resuelto): cada miembro apunta al otro.
		resuelto := fecha.AddDate(0, 0, 1)
		repo.episodios[0].DuplicadoDeID = ptrUUID(repo.episodios[1].ID)
		repo.episodios[0].MarcaDuplicadoResueltaEn = &resuelto
		repo.episodios[1].DuplicadoDeID = ptrUUID(repo.episodios[0].ID)
		repo.episodios[1].MarcaDuplicadoResueltaEn = &resuelto
		repo.seedFuentes(FuenteEpisodio{
			EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
			TituloFeed: "Sermón del domingo", Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
		})
		// Mismo título y **otro día**: la correspondencia no es segura y la rama
		// dudosa necesita candidato libre (M1): el elemento entra suelto.
		entrada := entradaDePrueba(ProveedorPodcast, "guid-01", "Sermón del domingo", fecha.AddDate(0, 0, 3), "Del podcast")
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{podcast: ResultadoFeed{Entradas: []EntradaFeed{entrada}}}, &fakeProbador{}, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if resumen.Unidos != 0 || resumen.Marcados != 0 {
			t.Errorf("una ficha con par no entra en un par nuevo: %+v", resumen)
		}
		if len(repo.episodios) != 3 {
			t.Fatalf("fichas = %d, se esperaba una entrada suelta nueva", len(repo.episodios))
		}
		nueva := repo.episodios[2]
		if nueva.DuplicadoDeID != nil || nueva.Visibilidad != VisibilidadVisible {
			t.Errorf("la entrada suelta nace visible y sin marca: %+v", nueva)
		}
		if repo.episodios[0].DuplicadoDeID == nil || *repo.episodios[0].DuplicadoDeID != repo.episodios[1].ID {
			t.Errorf("la ficha emparejada no puede perder su pareja: %v", repo.episodios[0].DuplicadoDeID)
		}
		if len(fuentesDe(repo, repo.episodios[0].ID)) != 1 {
			t.Error("la ficha con par no debe recibir la fuente del elemento nuevo")
		}
	})
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

// TestBuscarNuevosNoRevierteUnaResolucionDeMarca fija el residuo del eje de
// disponibilidad (N2/M3): la búsqueda escribe la marca **solo cuando decide una
// nueva**. Si el equipo resuelve la marca entre la lectura de la ficha y la
// escritura del sync —una ventana de milisegundos—, la resolución sobrevive; el
// sync sigue guardando el resultado de su comprobación.
func TestBuscarNuevosNoRevierteUnaResolucionDeMarca(t *testing.T) {
	base := time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	marca := MarcaDisponibilidadFueraDelFeed
	repo.seed(Episodio{
		Titulo: "Marcado y a medio salir", PublicadoEn: base, Origen: OrigenAmbos,
		MarcaDisponibilidad: &marca,
	})
	repo.seedFuentes(
		// El video salió de su ventana (su enlace se comprueba); el podcast
		// sigue en la suya, así que la ficha no se vuelve a marcar.
		FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
			TituloFeed: "Marcado y a medio salir", Enlace: "https://www.youtube.com/watch?v=SimienteAA1"},
		FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorPodcast, IDExterno: "guid-01",
			TituloFeed: "Marcado y a medio salir", Enlace: "https://podcasters.spotify.com/pod/show/simiente/episodes/guid-01"},
	)
	feedPodcast := ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(ProveedorPodcast, "guid-01",
		"Marcado y a medio salir", base, "Descripción")}}
	probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": DisponibilidadDisponible}}

	// El equipo resuelve la marca justo entre la lectura de la ficha y la
	// escritura del sync (el gancho corre dentro de esa escritura).
	resuelto := base.Add(time.Hour)
	id := repo.episodios[0].ID
	repo.antesDeEscribirDisponibilidad = func() {
		repo.mutarFicha(id, func(ficha *Episodio) { ficha.MarcaDisponibilidadResueltaEn = &resuelto })
	}

	service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{podcast: feedPodcast}, probador, &fakeCaratulas{})
	if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}

	ficha := repo.episodios[0]
	if ficha.MarcaDisponibilidadResueltaEn == nil || !ficha.MarcaDisponibilidadResueltaEn.Equal(resuelto) {
		t.Errorf("la resolución del equipo se revirtió: %v", ficha.MarcaDisponibilidadResueltaEn)
	}
	if !mismoMarca(ficha.MarcaDisponibilidad, &marca) {
		t.Errorf("la marca del equipo se pisó: %v", ficha.MarcaDisponibilidad)
	}
	if ficha.DisponibilidadComprobadaEn == nil || ficha.ResultadoDisponibilidad == nil ||
		*ficha.ResultadoDisponibilidad != DisponibilidadDisponible {
		t.Errorf("el sync debe guardar su comprobación: %+v", ficha)
	}
}

// probadorQueEdita es el doble del probador que, al comprobar el enlace, simula
// una **edición del equipo** (la descripción corta) hecha mientras la búsqueda
// corría: es la carrera de SC-007/M3 (la ventana de la búsqueda llega a los 90 s).
// La edición entra al almacén con el mutex del doble (`mutarFicha`): el probador
// corre en goroutines en vuelo (P9-3).
type probadorQueEdita struct {
	repo *fakeRepository
	id   uuid.UUID
	una  sync.Once
}

func (p *probadorQueEdita) ProbeYouTube(context.Context, string) (ResultadoDisponibilidad, error) {
	p.una.Do(func() {
		p.repo.mutarFicha(p.id, func(ficha *Episodio) {
			ficha.DescripcionEs = ptr("Descripción del equipo")
			ficha.DescripcionEditada = true
		})
	})
	return DisponibilidadInconclusa, nil
}

func (p *probadorQueEdita) ProbePodcast(context.Context, string) (ResultadoDisponibilidad, error) {
	return DisponibilidadInconclusa, nil
}

// TestBuscarNuevosNoRevierteUnaEdicionDuranteLaBusqueda fija M3/SC-007: la
// búsqueda escribe **solo las columnas que le competen**, así que una edición del
// equipo guardada mientras la búsqueda corría (la descripción, aquí) sobrevive a
// la escritura del eje de disponibilidad.
func TestBuscarNuevosNoRevierteUnaEdicionDuranteLaBusqueda(t *testing.T) {
	fecha := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.seed(Episodio{
		Titulo: "Sermón viejo", PublicadoEn: fecha, Origen: OrigenVideo,
		DescripcionEs: ptr("Del feed"),
	})
	repo.seedFuentes(FuenteEpisodio{
		EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
		TituloFeed: "Sermón viejo", Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
	})
	// El feed responde con otro elemento: la ficha sembrada queda fuera de la
	// ventana y entra en la cola de comprobación.
	entrada := entradaDePrueba(ProveedorYouTube, "SimienteAA9", "Otro sermón", fecha, "Descripción")
	probador := &probadorQueEdita{repo: repo, id: repo.episodios[0].ID}
	service := servicioConDoblesYProbador(t, repo,
		&fakeFeedSource{youtube: ResultadoFeed{Entradas: []EntradaFeed{entrada}}}, probador, &fakeCaratulas{})

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if resumen.Comprobados != 1 {
		t.Fatalf("comprobaciones = %d, se esperaba la de la ficha fuera de ventana", resumen.Comprobados)
	}
	ficha := repo.episodios[0]
	if ficha.DescripcionEs == nil || *ficha.DescripcionEs != "Descripción del equipo" || !ficha.DescripcionEditada {
		t.Errorf("la búsqueda no puede revertir la edición del equipo: %+v", ficha)
	}
	if ficha.DisponibilidadComprobadaEn == nil || ficha.ResultadoDisponibilidad == nil {
		t.Errorf("el resultado de la comprobación sí se escribe: %+v", ficha)
	}
	if ficha.MarcaDisponibilidad == nil || *ficha.MarcaDisponibilidad != MarcaDisponibilidadFueraDelFeed {
		t.Errorf("la marca de la ventana del feed sí se escribe: %+v", ficha.MarcaDisponibilidad)
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
// Ventana del feed, disponibilidad y carátulas (T921)
// ---------------------------------------------------------------------------

// TestBuscarNuevosVentanaDelFeed fija FR-017/SC-014 (la regla del «estante de
// 15»): salir de la ventana del feed **no oculta por sí solo**; se comprueba el
// enlace y solo un **404/410 confirmado** oculta y marca «no disponible»; un
// enlace vivo o una comprobación que no se pudo completar dejan la ficha
// **visible** con su marca de revisión.
func TestBuscarNuevosVentanaDelFeed(t *testing.T) {
	casos := []struct {
		name             string
		resultado        ResultadoDisponibilidad
		quierVisibilidad Visibilidad
		quierMarca       *MarcaDisponibilidad
		quierOcultos     int
		quierInconclusos int
	}{
		{
			name:             "enlace confirmado muerto (404) oculta y marca",
			resultado:        DisponibilidadNoDisponible,
			quierVisibilidad: VisibilidadOculta,
			quierMarca:       ptrMarca(MarcaDisponibilidadNoDisponible),
			quierOcultos:     1,
		},
		{
			name:             "enlace vivo sigue visible y se marca",
			resultado:        DisponibilidadDisponible,
			quierVisibilidad: VisibilidadVisible,
			quierMarca:       ptrMarca(MarcaDisponibilidadFueraDelFeed),
		},
		{
			name:             "comprobación incompleta: nunca oculta",
			resultado:        DisponibilidadInconclusa,
			quierVisibilidad: VisibilidadVisible,
			quierMarca:       ptrMarca(MarcaDisponibilidadFueraDelFeed),
			quierInconclusos: 1,
		},
	}
	for _, caso := range casos {
		t.Run(caso.name, func(t *testing.T) {
			repo := newFakeRepository()
			// La ficha ya estaba importada y su elemento salió de la ventana del
			// feed: el feed ya no lo trae.
			repo.seed(Episodio{Titulo: "Salió de la ventana", PublicadoEn: time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)})
			repo.seedFuentes(FuenteEpisodio{
				EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
				TituloFeed: "Salió de la ventana", Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
			})
			probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": caso.resultado}}
			service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{}, probador, &fakeCaratulas{})

			resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
			if err != nil {
				t.Fatalf("BuscarNuevos devolvió %v", err)
			}
			if len(probador.comprobadas()) != 1 || probador.comprobadas()[0] != "youtube/SimienteAA1" {
				t.Errorf("comprobaciones = %v, se esperaba una del enlace que salió", probador.comprobadas())
			}
			ficha := repo.episodios[0]
			if ficha.Visibilidad != caso.quierVisibilidad {
				t.Errorf("visibilidad = %q, se esperaba %q", ficha.Visibilidad, caso.quierVisibilidad)
			}
			if !mismoMarca(ficha.MarcaDisponibilidad, caso.quierMarca) {
				t.Errorf("marca = %v, se esperaba %v", ficha.MarcaDisponibilidad, caso.quierMarca)
			}
			if ficha.MarcaDisponibilidadResueltaEn != nil {
				t.Error("la marca nueva nace abierta")
			}
			if ficha.ResultadoDisponibilidad == nil || *ficha.ResultadoDisponibilidad != caso.resultado {
				t.Errorf("resultado = %v, se esperaba %q", ficha.ResultadoDisponibilidad, caso.resultado)
			}
			if ficha.DisponibilidadComprobadaEn == nil {
				t.Error("la comprobación debe dejar su fecha")
			}
			if resumen.Ocultos != caso.quierOcultos {
				t.Errorf("hiddenCount = %d, se esperaba %d", resumen.Ocultos, caso.quierOcultos)
			}
			if resumen.Comprobados != 1 {
				t.Errorf("availabilityCheckedCount = %d, se esperaba 1", resumen.Comprobados)
			}
			if resumen.Inconclusos != caso.quierInconclusos {
				t.Errorf("availabilityInconclusiveCount = %d, se esperaba %d", resumen.Inconclusos, caso.quierInconclusos)
			}
			if resumen.Marcados != 0 {
				t.Errorf("markedCount cuenta solo «posible duplicado» (B2): %d", resumen.Marcados)
			}
		})
	}
}

// TestBuscarNuevosAgregacionPorFicha fija la regla M2 (precisión 2): la
// comprobación es **por fuente** y el resultado de la ficha es **el peor** —un
// enlace confirmado muerto oculta la ficha entera aunque el otro responda 200—;
// «ya no aparece en el feed» solo se marca cuando **todas** sus fuentes salieron
// de su ventana.
func TestBuscarNuevosAgregacionPorFicha(t *testing.T) {
	t.Run("un enlace muerto oculta la ficha de dos enlaces", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Unida", PublicadoEn: time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC), Origen: OrigenAmbos})
		repo.seedFuentes(
			FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
				TituloFeed: "Unida", Enlace: "https://www.youtube.com/watch?v=SimienteAA1"},
			FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorPodcast, IDExterno: "guid-01",
				TituloFeed: "Unida", Enlace: "https://podcasters.spotify.com/pod/show/simiente/episodes/guid-01"},
		)
		probador := &fakeProbador{
			youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": DisponibilidadDisponible},
			podcast: map[string]ResultadoDisponibilidad{
				"https://podcasters.spotify.com/pod/show/simiente/episodes/guid-01": DisponibilidadNoDisponible,
			},
		}
		service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{}, probador, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		ficha := repo.episodios[0]
		if ficha.Visibilidad != VisibilidadOculta {
			t.Error("un enlace confirmado muerto oculta la ficha entera (precisión 2)")
		}
		if !mismoMarca(ficha.MarcaDisponibilidad, ptrMarca(MarcaDisponibilidadNoDisponible)) {
			t.Errorf("la marca debe ser «no disponible» (el peor gana): %v", ficha.MarcaDisponibilidad)
		}
		if ficha.ResultadoDisponibilidad == nil || *ficha.ResultadoDisponibilidad != DisponibilidadNoDisponible {
			t.Errorf("el resultado agregado es el peor: %v", ficha.ResultadoDisponibilidad)
		}
		if resumen.Ocultos != 1 || resumen.Comprobados != 2 {
			t.Errorf("resumen = %+v, se esperaban 1 oculto y 2 comprobaciones", resumen)
		}
		publicos, err := service.GetMedios(context.Background(), LangES, 0, 0)
		if err != nil {
			t.Fatalf("GetMedios devolvió %v", err)
		}
		if len(publicos.Items) != 0 {
			t.Errorf("la ficha oculta no puede salir: %v", titulosPublicos(publicos.Items))
		}
	})

	t.Run("con una sola fuente fuera de ventana no hay marca", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "A medias fuera", PublicadoEn: time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC), Origen: OrigenAmbos})
		repo.seedFuentes(
			// Solo el video salió de su ventana: el podcast sigue en la suya.
			FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
				TituloFeed: "A medias fuera", Enlace: "https://www.youtube.com/watch?v=SimienteAA1"},
			FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorPodcast, IDExterno: "guid-01",
				TituloFeed: "A medias fuera", Enlace: "https://podcasters.spotify.com/pod/show/simiente/episodes/guid-01"},
		)
		feedPodcast := ResultadoFeed{Entradas: []EntradaFeed{entradaDePrueba(ProveedorPodcast, "guid-01",
			"A medias fuera", time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC), "Descripción")}}
		probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": DisponibilidadDisponible}}
		service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{podcast: feedPodcast}, probador, &fakeCaratulas{})

		if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		ficha := repo.episodios[0]
		if ficha.MarcaDisponibilidad != nil {
			t.Errorf("decir «ya no aparece en el feed» sería falso: %v", ficha.MarcaDisponibilidad)
		}
		if ficha.Visibilidad != VisibilidadVisible {
			t.Error("la ficha sigue visible")
		}
	})
}

// TestBuscarNuevosMatrizDeMarcas fija la matriz «par abierto/resuelto × enlace
// vivo/muerto/inconcluso» (correcciones C1/A3): los dos ejes **conviven**, el par
// de duplicados **nunca se rompe** al marcar disponibilidad (sigue disponible
// «Revisar duplicado») y la indisponibilidad confirmada oculta aunque el equipo
// hubiera resuelto antes la marca.
func TestBuscarNuevosMatrizDeMarcas(t *testing.T) {
	base := time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)
	casos := []struct {
		name         string
		parResuelto  bool
		resultado    ResultadoDisponibilidad
		quierMarca   MarcaDisponibilidad
		quierVisible Visibilidad
	}{
		{name: "par abierto y enlace vivo", resultado: DisponibilidadDisponible, quierMarca: MarcaDisponibilidadFueraDelFeed, quierVisible: VisibilidadVisible},
		{name: "par abierto y enlace muerto", resultado: DisponibilidadNoDisponible, quierMarca: MarcaDisponibilidadNoDisponible, quierVisible: VisibilidadOculta},
		{name: "par abierto y comprobación incompleta", resultado: DisponibilidadInconclusa, quierMarca: MarcaDisponibilidadFueraDelFeed, quierVisible: VisibilidadVisible},
		{name: "par resuelto y enlace vivo", parResuelto: true, resultado: DisponibilidadDisponible, quierMarca: MarcaDisponibilidadFueraDelFeed, quierVisible: VisibilidadVisible},
		{name: "par resuelto y enlace muerto", parResuelto: true, resultado: DisponibilidadNoDisponible, quierMarca: MarcaDisponibilidadNoDisponible, quierVisible: VisibilidadOculta},
		{name: "par resuelto y comprobación incompleta", parResuelto: true, resultado: DisponibilidadInconclusa, quierMarca: MarcaDisponibilidadFueraDelFeed, quierVisible: VisibilidadVisible},
	}
	for _, caso := range casos {
		t.Run(caso.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.seed(
				Episodio{Titulo: "Con par", PublicadoEn: base},
				Episodio{Titulo: "Su par", PublicadoEn: base},
			)
			parID := repo.episodios[1].ID
			repo.episodios[0].DuplicadoDeID = ptrUUID(parID)
			repo.episodios[1].DuplicadoDeID = ptrUUID(repo.episodios[0].ID)
			if caso.parResuelto {
				repo.episodios[0].MarcaDuplicadoResueltaEn = ptrTime(base)
				repo.episodios[1].MarcaDuplicadoResueltaEn = ptrTime(base)
			}
			repo.seedFuentes(FuenteEpisodio{
				EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
				TituloFeed: "Con par", Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
			})
			probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": caso.resultado}}
			service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{}, probador, &fakeCaratulas{})

			if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
				t.Fatalf("BuscarNuevos devolvió %v", err)
			}
			ficha := repo.episodios[0]
			if !mismoMarca(ficha.MarcaDisponibilidad, ptrMarca(caso.quierMarca)) {
				t.Errorf("marca de disponibilidad = %v, se esperaba %q", ficha.MarcaDisponibilidad, caso.quierMarca)
			}
			if ficha.Visibilidad != caso.quierVisible {
				t.Errorf("visibilidad = %q, se esperaba %q", ficha.Visibilidad, caso.quierVisible)
			}
			// El eje de duplicados queda intacto: el par (y su resolución) siguen.
			if ficha.DuplicadoDeID == nil || *ficha.DuplicadoDeID != parID {
				t.Errorf("el par se rompió al marcar disponibilidad: %v", ficha.DuplicadoDeID)
			}
			if caso.parResuelto != (ficha.MarcaDuplicadoResueltaEn != nil) {
				t.Errorf("la resolución del par cambió: %v", ficha.MarcaDuplicadoResueltaEn)
			}
			// La marca de disponibilidad nueva nace abierta.
			if caso.resultado == DisponibilidadNoDisponible && ficha.MarcaDisponibilidadResueltaEn != nil {
				t.Error("la marca nueva de indisponibilidad nace abierta (A3)")
			}
		})
	}
}

// TestBuscarNuevosPrecedenciaA3 fija la precisión aprobada en la fase 5 (A3): la
// **indisponibilidad confirmada manda** — oculta la ficha aunque el equipo
// hubiera resuelto antes «mantener visible» — y, en el sentido contrario, una
// marca de la **misma clase** ya resuelta **no se reabre** sin confirmación nueva.
func TestBuscarNuevosPrecedenciaA3(t *testing.T) {
	base := time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)

	t.Run("marca resuelta + 404 confirmado → oculta y marca nueva abierta", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "El equipo la mantuvo visible", PublicadoEn: base,
			MarcaDisponibilidad:           ptrMarca(MarcaDisponibilidadFueraDelFeed),
			MarcaDisponibilidadResueltaEn: ptrTime(base),
		})
		repo.seedFuentes(FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube,
			IDExterno: "SimienteAA1", TituloFeed: "El equipo la mantuvo visible",
			Enlace: "https://www.youtube.com/watch?v=SimienteAA1"})
		probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": DisponibilidadNoDisponible}}
		service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{}, probador, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		ficha := repo.episodios[0]
		if ficha.Visibilidad != VisibilidadOculta {
			t.Error("la indisponibilidad confirmada manda sobre la resolución previa (A3)")
		}
		if !mismoMarca(ficha.MarcaDisponibilidad, ptrMarca(MarcaDisponibilidadNoDisponible)) {
			t.Errorf("la marca nueva es «no disponible»: %v", ficha.MarcaDisponibilidad)
		}
		if ficha.MarcaDisponibilidadResueltaEn != nil {
			t.Error("la marca nueva nace ABIERTA (cambiar de clase resetea la resolución)")
		}
		if resumen.Ocultos != 1 {
			t.Errorf("hiddenCount = %d, se esperaba 1", resumen.Ocultos)
		}
	})

	t.Run("misma clase ya resuelta + enlace vivo → no se reabre", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Sigue fuera del feed", PublicadoEn: base,
			MarcaDisponibilidad:           ptrMarca(MarcaDisponibilidadFueraDelFeed),
			MarcaDisponibilidadResueltaEn: ptrTime(base),
		})
		repo.seedFuentes(FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube,
			IDExterno: "SimienteAA1", TituloFeed: "Sigue fuera del feed",
			Enlace: "https://www.youtube.com/watch?v=SimienteAA1"})
		probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": DisponibilidadDisponible}}
		service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{}, probador, &fakeCaratulas{})

		if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		ficha := repo.episodios[0]
		if ficha.MarcaDisponibilidadResueltaEn == nil {
			t.Error("la misma causa no reabre una marca resuelta (default #6)")
		}
		if !ficha.MarcaDisponibilidadResueltaEn.Equal(base) {
			t.Errorf("la resolución se movió: %v", ficha.MarcaDisponibilidadResueltaEn)
		}
		if ficha.Visibilidad != VisibilidadVisible {
			t.Error("la ficha sigue visible")
		}
	})
}

// TestBuscarNuevosColaDeComprobacion fija la cola de FR-017: solo se comprueba lo
// que salió de la ventana y pertenece a fichas **visibles** no absorbidas —una
// ficha oculta por el equipo no se re-comprueba—, y un feed caído no aporta
// comprobaciones ni ocultaciones (FR-016).
func TestBuscarNuevosColaDeComprobacion(t *testing.T) {
	base := time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)

	t.Run("una ficha oculta por el equipo no se comprueba", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Oculta", PublicadoEn: base, Visibilidad: VisibilidadOculta})
		repo.seedFuentes(FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube,
			IDExterno: "SimienteAA1", TituloFeed: "Oculta", Enlace: "https://www.youtube.com/watch?v=SimienteAA1"})
		probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{"SimienteAA1": DisponibilidadNoDisponible}}
		service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{}, probador, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if len(probador.comprobadas()) != 0 {
			t.Errorf("no se comprueba lo que el equipo ocultó: %v", probador.comprobadas())
		}
		if resumen.Comprobados != 0 || resumen.Ocultos != 0 {
			t.Errorf("resumen = %+v, se esperaban 0 comprobaciones", resumen)
		}
		if repo.episodios[0].MarcaDisponibilidad != nil {
			t.Error("una ficha oculta por el equipo no se marca")
		}
	})

	t.Run("un feed caído no comprueba ni oculta nada", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Del podcast", PublicadoEn: base})
		repo.seedFuentes(FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorPodcast,
			IDExterno: "guid-01", TituloFeed: "Del podcast",
			Enlace: "https://podcasters.spotify.com/pod/show/simiente/episodes/guid-01"})
		probador := &fakeProbador{porDefecto: DisponibilidadNoDisponible}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{errPodcast: errFeedNoDisponible}, probador, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if resumen.EstadoPodcast != string(FeedFallido) {
			t.Errorf("el feed caído debe informarse: %q", resumen.EstadoPodcast)
		}
		if len(probador.comprobadas()) != 0 || resumen.Comprobados != 0 || resumen.Ocultos != 0 {
			t.Errorf("un feed caído no oculta ni comprueba nada: %+v", resumen)
		}
		if repo.episodios[0].Visibilidad != VisibilidadVisible || repo.episodios[0].MarcaDisponibilidad != nil {
			t.Error("su fuente no cuenta como «fuera de ventana» (FR-016)")
		}
	})

	t.Run("el tope de comprobaciones acota la cola", func(t *testing.T) {
		repo := newFakeRepository()
		entradas := make([]FuenteEpisodio, 0, maxComprobacionesPorBusqueda+5)
		for i := 0; i < maxComprobacionesPorBusqueda+5; i++ {
			repo.seed(Episodio{Titulo: fmt.Sprintf("Viejo %d", i), PublicadoEn: base})
			entradas = append(entradas, FuenteEpisodio{
				EpisodioID: repo.episodios[i].ID, Proveedor: ProveedorYouTube,
				IDExterno: fmt.Sprintf("Simiente%03d", i), TituloFeed: fmt.Sprintf("Viejo %d", i),
				Enlace: fmt.Sprintf("https://www.youtube.com/watch?v=Simiente%03d", i),
			})
		}
		repo.seedFuentes(entradas...)
		probador := &fakeProbador{youtube: map[string]ResultadoDisponibilidad{}}
		service := servicioConDoblesYProbador(t, repo, &fakeFeedSource{}, probador, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if resumen.Comprobados != maxComprobacionesPorBusqueda {
			t.Errorf("comprobaciones = %d, se esperaba el tope %d", resumen.Comprobados, maxComprobacionesPorBusqueda)
		}
	})
}

// TestBuscarNuevosCaratulas fija P9-7/SC-005: la carátula del elemento importado
// se **copia al sitio** (el nombre lo genera el servidor) y un fallo de descarga
// **nunca** aborta la importación: la ficha entra sin carátula (marcador neutro).
func TestBuscarNuevosCaratulas(t *testing.T) {
	fecha := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("copia la carátula del elemento importado", func(t *testing.T) {
		repo := newFakeRepository()
		entrada := entradaDePrueba(ProveedorYouTube, "SimienteAA1", "Con carátula", fecha, "Descripción")
		entrada.CaratulaURL = caratulaYouTube
		caratulas := &fakeCaratulas{archivo: "img_55555555-5555-5555-5555-555555555555.jpg"}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{youtube: ResultadoFeed{Entradas: []EntradaFeed{entrada}}}, &fakeProbador{}, caratulas)

		if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if len(caratulas.descargadas()) != 1 || caratulas.descargadas()[0] != caratulaYouTube {
			t.Errorf("carátulas pedidas = %v", caratulas.descargadas())
		}
		ficha := repo.episodios[0]
		if ficha.Caratula == nil || *ficha.Caratula != caratulas.archivo {
			t.Errorf("la ficha debe quedarse con la copia del sitio: %v", ficha.Caratula)
		}
	})

	t.Run("un fallo de descarga deja la ficha sin carátula y no aborta", func(t *testing.T) {
		repo := newFakeRepository()
		entrada := entradaDePrueba(ProveedorYouTube, "SimienteAA1", "Con carátula", fecha, "Descripción")
		entrada.CaratulaURL = caratulaYouTube
		caratulas := &fakeCaratulas{err: errCaratulaNoDisponible}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{youtube: ResultadoFeed{Entradas: []EntradaFeed{entrada}}}, &fakeProbador{}, caratulas)

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("una carátula caída no aborta la importación: %v", err)
		}
		if resumen.Importados != 1 || len(repo.episodios) != 1 {
			t.Errorf("la ficha debía importarse igual: %+v", resumen)
		}
		if repo.episodios[0].Caratula != nil {
			t.Errorf("la ficha entra sin carátula (marcador neutro): %v", repo.episodios[0].Caratula)
		}
	})

	t.Run("el tope de carátulas acota las descargas", func(t *testing.T) {
		repo := newFakeRepository()
		entradas := make([]EntradaFeed, 0, maxCaratulasPorBusqueda+5)
		for i := 0; i < maxCaratulasPorBusqueda+5; i++ {
			entrada := entradaDePrueba(ProveedorYouTube, fmt.Sprintf("Simiente%03d", i),
				fmt.Sprintf("Sermón %d", i), fecha.Add(-time.Duration(i)*time.Hour), "Descripción")
			entrada.CaratulaURL = caratulaYouTube
			entradas = append(entradas, entrada)
		}
		caratulas := &fakeCaratulas{}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{youtube: ResultadoFeed{Entradas: entradas}}, &fakeProbador{}, caratulas)

		if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if len(caratulas.descargadas()) != maxCaratulasPorBusqueda {
			t.Errorf("descargas = %d, se esperaba el tope %d", len(caratulas.descargadas()), maxCaratulasPorBusqueda)
		}
	})
}

// TestBuscarNuevosNoPisaLaCaratula fija P9-9/default #10 y el default #7: la
// carátula es una instantánea —la búsqueda **no** refresca la que ya tiene una
// ficha— y, cuando el mismo sermón la trae de los dos feeds, manda la del
// **video**. Ninguna de las dos reglas dependía del orden en que terminaran las
// descargas.
func TestBuscarNuevosNoPisaLaCaratula(t *testing.T) {
	fecha := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)
	tituloVideo := "Potencial sin Resultados La Lección de la Higuera"
	tituloPodcast := "Potencial sin Resultados: La Lección de la Higuera"

	t.Run("unir no pisa la carátula del video ni borra su archivo", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: tituloVideo, PublicadoEn: fecha, Origen: OrigenVideo,
			Caratula: ptr("img_11111111-1111-1111-1111-111111111111.jpg"),
		})
		repo.seedFuentes(FuenteEpisodio{
			EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
			TituloFeed: tituloVideo, Enlace: "https://www.youtube.com/watch?v=SimienteAA1",
		})
		entrada := entradaDePrueba(ProveedorPodcast, "guid-01", tituloPodcast, fecha, "Descripción")
		entrada.CaratulaURL = caratulaPodcast
		caratulas := &fakeCaratulas{porURL: map[string]string{caratulaPodcast: "img_22222222-2222-2222-2222-222222222222.jpg"}}
		almacen := &fakeAlmacen{}
		service := NewService(ServiceDeps{
			Repository:         repo,
			FeedSource:         &fakeFeedSource{podcast: ResultadoFeed{Entradas: []EntradaFeed{entrada}}},
			AvailabilityProber: &fakeProbador{},
			CoverFetcher:       caratulas,
			Store:              almacen,
			Location:           zonaDePrueba(t),
		})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if resumen.Unidos != 1 || len(repo.episodios) != 1 {
			t.Fatalf("el elemento del podcast debía unirse: %+v", resumen)
		}
		ficha := repo.episodios[0]
		if ficha.Caratula == nil || *ficha.Caratula != "img_11111111-1111-1111-1111-111111111111.jpg" {
			t.Errorf("la carátula del video debe seguir siendo la de la ficha: %v", ficha.Caratula)
		}
		if descargas := caratulas.descargadas(); len(descargas) != 0 {
			t.Errorf("no se descarga la carátula de una ficha que ya tiene: %v", descargas)
		}
		if retirados := almacen.retirados(); len(retirados) != 0 {
			t.Errorf("el archivo del video no se puede borrar: %v", retirados)
		}
	})

	t.Run("con las dos carátulas del mismo sermón gana la del video", func(t *testing.T) {
		repo := newFakeRepository()
		entradaVideo := entradaDePrueba(ProveedorYouTube, "SimienteAA1", tituloVideo, fecha, "Descripción")
		entradaVideo.CaratulaURL = caratulaYouTube
		entradaPodcast := entradaDePrueba(ProveedorPodcast, "guid-01", tituloPodcast, fecha, "Descripción")
		entradaPodcast.CaratulaURL = caratulaPodcast
		caratulas := &fakeCaratulas{porURL: map[string]string{
			caratulaYouTube: "img_33333333-3333-3333-3333-333333333333.jpg",
			caratulaPodcast: "img_44444444-4444-4444-4444-444444444444.jpg",
		}}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{
				youtube: ResultadoFeed{Entradas: []EntradaFeed{entradaVideo}},
				podcast: ResultadoFeed{Entradas: []EntradaFeed{entradaPodcast}},
			}, &fakeProbador{}, caratulas)

		if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if len(repo.episodios) != 1 {
			t.Fatalf("el sermón debe quedar en una sola ficha: %d", len(repo.episodios))
		}
		if descargas := caratulas.descargadas(); len(descargas) != 1 || descargas[0] != caratulaYouTube {
			t.Errorf("descargas = %v, se esperaba solo la del video", descargas)
		}
		if ficha := repo.episodios[0]; ficha.Caratula == nil || *ficha.Caratula != "img_33333333-3333-3333-3333-333333333333.jpg" {
			t.Errorf("la ficha debe quedarse con la carátula del video: %v", ficha.Caratula)
		}
	})
}

// TestBuscarNuevosNoDejaCaratulasHuerfanas fija N3: cuando la escritura
// condicional de la carátula falla, la copia recién descargada se retira del
// almacén **solo si la fila viva confirma que no la usa**. Un error de respuesta
// de una escritura que sí se aplicó no puede borrar la carátula que el sitio
// está sirviendo.
func TestBuscarNuevosNoDejaCaratulasHuerfanas(t *testing.T) {
	fecha := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)
	archivo := "img_77777777-7777-7777-7777-777777777777.jpg"

	preparar := func(repo *fakeRepository, almacen *fakeAlmacen) *service {
		entrada := entradaDePrueba(ProveedorYouTube, "SimienteAA1", "Con carátula", fecha, "Descripción")
		entrada.CaratulaURL = caratulaYouTube
		return NewService(ServiceDeps{
			Repository:         repo,
			FeedSource:         &fakeFeedSource{youtube: ResultadoFeed{Entradas: []EntradaFeed{entrada}}},
			AvailabilityProber: &fakeProbador{},
			CoverFetcher:       &fakeCaratulas{archivo: archivo},
			Store:              almacen,
			Location:           zonaDePrueba(t),
		})
	}

	t.Run("la escritura que no se aplicó retira su copia", func(t *testing.T) {
		repo := newFakeRepository()
		repo.coverErr = errors.New("no se pudo guardar la carátula")
		almacen := &fakeAlmacen{}
		service := preparar(repo, almacen)

		if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err == nil {
			t.Fatal("el fallo de la escritura de la carátula debe propagarse")
		}
		if retirados := almacen.retirados(); len(retirados) != 1 || retirados[0] != archivo {
			t.Errorf("la copia sin uso debe retirarse del almacén: %v", retirados)
		}
	})

	t.Run("un error de respuesta no borra la carátula que sí se guardó", func(t *testing.T) {
		repo := newFakeRepository()
		repo.coverErr = errors.New("respuesta perdida")
		// La escritura sí llegó a la base: la fila viva usa ese archivo.
		repo.antesDeCover = func(id uuid.UUID) {
			repo.mutarFicha(id, func(ficha *Episodio) { ficha.Caratula = &archivo })
		}
		almacen := &fakeAlmacen{}
		service := preparar(repo, almacen)

		if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err == nil {
			t.Fatal("el error de la escritura debe propagarse")
		}
		if retirados := almacen.retirados(); len(retirados) != 0 {
			t.Errorf("el archivo en uso no se puede borrar: %v", retirados)
		}
		if repo.episodios[0].Caratula == nil || *repo.episodios[0].Caratula != archivo {
			t.Errorf("la ficha conserva su carátula: %v", repo.episodios[0].Caratula)
		}
	})
}

// TestBuscarNuevosAltaAtomica fija H3/SC-003 y P9-11: el alta de un elemento
// (ficha + fuente, y la pareja del par dudoso) es **una sola transacción**, así
// que dos búsquedas simultáneas del mismo elemento no dejan ninguna ficha
// huérfana sin enlaces junto a la duplicada: la que pierde la carrera no escribe
// nada.
func TestBuscarNuevosAltaAtomica(t *testing.T) {
	fecha := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("el alta conflictuada no deja ficha huérfana", func(t *testing.T) {
		repo := newFakeRepository()
		entrada := entradaDePrueba(ProveedorYouTube, "SimienteAA1", "Sermón nuevo", fecha, "Descripción")
		// Otra búsqueda simultánea importa el mismo elemento justo antes del alta:
		// es la carrera real (la llave de idempotencia la resuelve la base).
		repo.antesDeInsertar = func() {
			repo.antesDeInsertar = nil
			repo.seedFuentes(FuenteEpisodio{
				EpisodioID: uuid.New(), Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
				TituloFeed: "Sermón nuevo", Enlace: entrada.Enlace,
			})
		}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{youtube: ResultadoFeed{Entradas: []EntradaFeed{entrada}}}, &fakeProbador{}, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if len(repo.episodios) != 0 {
			t.Errorf("quedó una ficha sin enlaces de la búsqueda que perdió: %+v", repo.episodios)
		}
		if resumen.Importados != 0 {
			t.Errorf("importados = %d, se esperaba 0 (nada se escribió)", resumen.Importados)
		}
	})

	t.Run("el par dudoso conflictuado no deja pareja a medias", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Otro título del mismo día", PublicadoEn: fecha, Origen: OrigenAudio})
		candidato := repo.episodios[0].ID
		repo.seedFuentes(FuenteEpisodio{
			EpisodioID: candidato, Proveedor: ProveedorPodcast, IDExterno: "guid-01",
			TituloFeed: "Otro título del mismo día", Enlace: "https://podcasters.spotify.com/pod/show/x/episodes/1",
		})
		entrada := entradaDePrueba(ProveedorYouTube, "SimienteAA1", "Un título distinto", fecha, "Descripción")
		repo.antesDeInsertar = func() {
			repo.antesDeInsertar = nil
			repo.seedFuentes(FuenteEpisodio{
				EpisodioID: uuid.New(), Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
				TituloFeed: "Un título distinto", Enlace: entrada.Enlace,
			})
		}
		service := servicioConDoblesYProbador(t, repo,
			&fakeFeedSource{youtube: ResultadoFeed{Entradas: []EntradaFeed{entrada}}}, &fakeProbador{}, &fakeCaratulas{})

		resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("BuscarNuevos devolvió %v", err)
		}
		if len(repo.episodios) != 1 {
			t.Errorf("fichas = %d, se esperaba solo el candidato (sin la adicional revertida)", len(repo.episodios))
		}
		if repo.episodios[0].DuplicadoDeID != nil {
			t.Errorf("el candidato no puede quedar marcado sin pareja: %+v", repo.episodios[0])
		}
		if resumen.Marcados != 0 {
			t.Errorf("marcados = %d, se esperaba 0", resumen.Marcados)
		}
	})
}

// TestBuscarNuevosResumenYRegistro fija FR-018/FR-022: cada búsqueda deja **una**
// fila en `media_sync_runs` con sus conteos y el estado de cada feed, y **una**
// sola fila `media.sync` (nunca una por elemento), con el extracto del resumen en
// `targetLabel`.
func TestBuscarNuevosResumenYRegistro(t *testing.T) {
	repo := newFakeRepository()
	service := servicioConFeeds(t, repo, feedRealYouTube(), feedRealPodcast())

	resumen, err := service.BuscarNuevos(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("BuscarNuevos devolvió %v", err)
	}
	if len(repo.syncRuns) != 1 {
		t.Fatalf("se registraron %d búsquedas, se esperaba 1", len(repo.syncRuns))
	}
	run := repo.syncRuns[0]
	if run.Importados != resumen.Importados || run.Unidos != resumen.Unidos || run.Marcados != resumen.Marcados ||
		run.Ocultos != resumen.Ocultos || run.Comprobados != resumen.Comprobados || run.Inconclusos != resumen.Inconclusos {
		t.Errorf("el registro no coincide con el resumen: %+v vs %+v", run, resumen)
	}
	if run.EstadoYouTube != FeedOK || run.EstadoPodcast != FeedOK {
		t.Errorf("el estado de los feeds debe persistirse: %+v", run)
	}
	if run.ActorUserID == nil {
		t.Error("la búsqueda registra quién la pidió (FR-022)")
	}
	if run.TerminadoEn.Before(run.IniciadoEn) {
		t.Error("la búsqueda no puede terminar antes de empezar")
	}

	filas := auditoriaDeBusqueda(repo)
	if len(filas) != 1 {
		t.Fatalf("se registraron %d filas media.sync, se esperaba exactamente 1", len(filas))
	}
	if filas[0].TargetKind != audit.TargetContent || filas[0].Result != audit.ResultSuccess {
		t.Errorf("fila media.sync = %+v", filas[0])
	}
	if !strings.HasPrefix(filas[0].TargetLabel, busquedaLabel) {
		t.Errorf("targetLabel = %q, se esperaba empezar por %q", filas[0].TargetLabel, busquedaLabel)
	}
	if !strings.Contains(filas[0].TargetLabel, "10 entraron") || !strings.Contains(filas[0].TargetLabel, "8 unieron") {
		t.Errorf("el extracto del resumen debe viajar en targetLabel: %q", filas[0].TargetLabel)
	}
	if len([]rune(filas[0].TargetLabel)) > maxTargetLabel {
		t.Errorf("targetLabel supera %d caracteres", maxTargetLabel)
	}

	// La segunda búsqueda no duplica nada y deja su propia fila.
	if _, err := service.BuscarNuevos(context.Background(), uuid.New()); err != nil {
		t.Fatalf("la segunda búsqueda devolvió %v", err)
	}
	if len(repo.syncRuns) != 2 || len(auditoriaDeBusqueda(repo)) != 2 {
		t.Errorf("cada búsqueda deja su registro: %d runs, %d filas",
			len(repo.syncRuns), len(auditoriaDeBusqueda(repo)))
	}
}

// TestBuscarNuevosPresupuestoAgotado fija P9-3/R9-9: al agotarse el presupuesto la
// búsqueda cierra con lo hecho —cada elemento ya quedó guardado en su
// transacción—, el resumen lleva conteos parciales y el registro se persiste
// igual: nada queda a medias ni se pierde.
func TestBuscarNuevosPresupuestoAgotado(t *testing.T) {
	repo := newFakeRepository()
	ctx, cancelar := context.WithCancel(context.Background())
	// El contexto se corta al empezar la segunda ficha: la segunda termina en su
	// transacción (nada a medias) y la tercera ya no arranca.
	insertadas := 0
	repo.antesDeInsertar = func() {
		insertadas++
		if insertadas == 2 {
			cancelar()
		}
	}
	entradas := make([]EntradaFeed, 0, 6)
	for i := 0; i < 6; i++ {
		entradas = append(entradas, entradaDePrueba(ProveedorYouTube, fmt.Sprintf("Simiente%03d", i),
			fmt.Sprintf("Sermón %d", i), time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC).Add(-time.Duration(i)*time.Hour),
			"Descripción"))
	}
	service := servicioConDoblesYProbador(t, repo,
		&fakeFeedSource{youtube: ResultadoFeed{Entradas: entradas}}, &fakeProbador{}, &fakeCaratulas{})

	resumen, err := service.BuscarNuevos(ctx, uuid.New())
	if err != nil {
		t.Fatalf("el presupuesto agotado no es un error: %v", err)
	}
	if resumen.Importados != 2 {
		t.Errorf("importados = %d, se esperaban los 2 ya guardados", resumen.Importados)
	}
	if len(repo.episodios) != 2 {
		t.Errorf("fichas = %d, se esperaban 2", len(repo.episodios))
	}
	for _, ficha := range repo.episodios {
		if len(fuentesDe(repo, ficha.ID)) != 1 {
			t.Errorf("una ficha quedó a medias (sin su fuente): %+v", ficha)
		}
	}
	if len(repo.syncRuns) != 1 {
		t.Fatalf("la búsqueda parcial también se registra: %d filas", len(repo.syncRuns))
	}
	if repo.syncRuns[0].Importados != 2 {
		t.Errorf("el registro lleva los conteos parciales: %+v", repo.syncRuns[0])
	}
	filas := auditoriaDeBusqueda(repo)
	if len(filas) != 1 || !strings.Contains(filas[0].TargetLabel, "sin terminar") {
		t.Errorf("el resumen parcial lo dice en su extracto: %+v", filas)
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

// servicioConDoblesYProbador arma el service con los dobles de la red saliente
// indicados (feeds, probador y carátulas) y la zona del lugar.
func servicioConDoblesYProbador(
	t *testing.T,
	repo *fakeRepository,
	feeds FeedSource,
	probador AvailabilityProber,
	caratulas CoverFetcher,
) *service {
	t.Helper()
	return NewService(ServiceDeps{
		Repository:         repo,
		FeedSource:         feeds,
		AvailabilityProber: probador,
		CoverFetcher:       caratulas,
		Store:              &fakeAlmacen{},
		Location:           zonaDePrueba(t),
	})
}

// mismoMarca compara dos marcas de disponibilidad opcionales.
func mismoMarca(a, b *MarcaDisponibilidad) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// auditoriaDeBusqueda devuelve las filas de auditoría de la búsqueda (una sola
// por ejecución: `media.sync`).
func auditoriaDeBusqueda(repo *fakeRepository) []audit.Action {
	out := make([]audit.Action, 0, 1)
	for _, action := range repo.committed {
		if action.Code == audit.ActionMediaSync {
			out = append(out, action)
		}
	}
	return out
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
