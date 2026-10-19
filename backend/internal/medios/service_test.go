package medios

import (
	"context"
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
)

// service_test.go es la base de las pruebas del service del dominio medios
// (T915): fija las **constantes** revisables (P9-18/R9-9: los topes son el único
// punto de verdad), la **normalización** (trim; `""`/espacios → NULL, patrón I6),
// la **validación de dominio** (FR-023) y el **repositorio falso** que usan las
// pruebas de la cara pública (T916/T917), del panel (T918/T919) y de la
// búsqueda (T920/T921).
//
// Aquí no hay red ni base de datos: el fake imita el comportamiento observable
// del repository real (predicado público, orden, filtros, idempotencia de las
// fuentes, repunte de la unión) y recuerda lo que se le pidió.

// ---------------------------------------------------------------------------
// Constantes del dominio (P9-18/R9-9)
// ---------------------------------------------------------------------------

// TestConstantesDelDominio fija por prueba los límites y los orígenes del
// módulo: cambiarlos es cambiar la regla (P9-18: constantes revisables, el único
// punto de verdad de los topes).
func TestConstantesDelDominio(t *testing.T) {
	// Topes de la sección y del bloque (R9-10/FR-027).
	if pageDefault != 12 {
		t.Errorf("pageDefault = %d, se esperaba 12", pageDefault)
	}
	if maxPage != 50 {
		t.Errorf("maxPage = %d, se esperaba 50", maxPage)
	}
	if maxPortada != 5 {
		t.Errorf("maxPortada = %d, se esperaba 5 (tope firme de FR-027)", maxPortada)
	}

	// Límites de los textos (000014/FR-020).
	if maxTitulo != 300 {
		t.Errorf("maxTitulo = %d, se esperaba 300", maxTitulo)
	}
	if maxDescripcion != 300 {
		t.Errorf("maxDescripcion = %d, se esperaba 300", maxDescripcion)
	}

	// Topes duros de una búsqueda (P9-3: 30/30/30, 4 en vuelo, 90 s).
	if maxElementosPorBusqueda != 30 {
		t.Errorf("maxElementosPorBusqueda = %d, se esperaba 30", maxElementosPorBusqueda)
	}
	if maxComprobacionesPorBusqueda != 30 {
		t.Errorf("maxComprobacionesPorBusqueda = %d, se esperaba 30", maxComprobacionesPorBusqueda)
	}
	if maxCaratulasPorBusqueda != 30 {
		t.Errorf("maxCaratulasPorBusqueda = %d, se esperaba 30", maxCaratulasPorBusqueda)
	}
	if maxEnVuelo != 4 {
		t.Errorf("maxEnVuelo = %d, se esperaba 4", maxEnVuelo)
	}
	if presupuestoBusqueda != 90*time.Second {
		t.Errorf("presupuestoBusqueda = %s, se esperaba 90s", presupuestoBusqueda)
	}

	// Timeouts por petición saliente (P9-3: 10/10/5 s) y topes de cuerpo.
	if tiempoEsperaFeeds != 10*time.Second {
		t.Errorf("tiempoEsperaFeeds = %s, se esperaba 10s", tiempoEsperaFeeds)
	}
	if tiempoEsperaCaratula != 10*time.Second {
		t.Errorf("tiempoEsperaCaratula = %s, se esperaba 10s", tiempoEsperaCaratula)
	}
	if tiempoEsperaProbador != 5*time.Second {
		t.Errorf("tiempoEsperaProbador = %s, se esperaba 5s", tiempoEsperaProbador)
	}

	// Los dos orígenes de FR-010 (P9-18: públicos, sin claves ni variables de
	// entorno nuevas).
	if urlFeedYouTube != "https://www.youtube.com/feeds/videos.xml?channel_id=UCXXJbgowcGnjI7BDsPnKOTw" {
		t.Errorf("urlFeedYouTube = %q", urlFeedYouTube)
	}
	if urlFeedPodcast != "https://anchor.fm/s/1116351e8/podcast/rss" {
		t.Errorf("urlFeedPodcast = %q", urlFeedPodcast)
	}

	// Etiqueta de auditoría (000004/000015).
	if maxTargetLabel != 254 {
		t.Errorf("maxTargetLabel = %d, se esperaba 254", maxTargetLabel)
	}
}

// ---------------------------------------------------------------------------
// Idioma y paginación
// ---------------------------------------------------------------------------

// TestNormalizeLang fija el idioma de la superficie pública: vacío es el idioma
// base y cualquier otro valor es `400 invalid` con `details.lang` (FR-025).
func TestNormalizeLang(t *testing.T) {
	tests := []struct {
		name  string
		lang  string
		quier string
		campo string
	}{
		{name: "vacío es el idioma base", lang: "", quier: LangES},
		{name: "español", lang: "es", quier: LangES},
		{name: "inglés", lang: "en", quier: LangEN},
		{name: "mayúsculas y espacios", lang: " EN ", quier: LangEN},
		{name: "idioma fuera del registro", lang: "fr", campo: "lang"},
		{name: "código largo", lang: "es-CR", campo: "lang"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lang, err := normalizeLang(test.lang)
			if test.campo != "" {
				requireInvalid(t, err, test.campo)
				return
			}
			if err != nil {
				t.Fatalf("normalizeLang(%q) devolvió %v", test.lang, err)
			}
			if lang != test.quier {
				t.Errorf("normalizeLang(%q) = %q, se esperaba %q", test.lang, lang, test.quier)
			}
		})
	}
}

// TestNormalizarPaginacionPublica fija la tanda de la sección (12 por defecto,
// tope 50) y el rechazo de los valores imposibles (R9-10/§8.1.3).
func TestNormalizarPaginacionPublica(t *testing.T) {
	tests := []struct {
		name          string
		limit, offset int
		quierLimit    int
		quierOffset   int
		campoInvalido string
	}{
		{name: "sin indicar usa la tanda por defecto", limit: 0, offset: 0, quierLimit: pageDefault},
		{name: "tanda explícita", limit: 24, offset: 12, quierLimit: 24, quierOffset: 12},
		{name: "la última tanda puede ser menor", limit: 3, quierLimit: 3},
		{name: "el tope se aplica sin error", limit: 51, quierLimit: maxPage},
		{name: "el tope exacto pasa", limit: maxPage, quierLimit: maxPage},
		{name: "límite negativo", limit: -1, campoInvalido: "limit"},
		{name: "desplazamiento negativo", offset: -1, campoInvalido: "offset"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params, err := normalizarPaginacionPublica(test.limit, test.offset)
			if test.campoInvalido != "" {
				requireInvalid(t, err, test.campoInvalido)
				return
			}
			if err != nil {
				t.Fatalf("normalizarPaginacionPublica(%d, %d) devolvió %v", test.limit, test.offset, err)
			}
			if params.Limit != test.quierLimit || params.Offset != test.quierOffset {
				t.Errorf("params = {%d, %d}, se esperaba {%d, %d}",
					params.Limit, params.Offset, test.quierLimit, test.quierOffset)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Normalización de textos (analyze I6)
// ---------------------------------------------------------------------------

// TestNormalizarOpcional fija el patrón I6: `""` o solo espacios pasa a NULL
// (nunca la cadena vacía que los CHECK de longitud rechazan) y el contenido
// (tildes, ñ, ¿¡, emojis) se conserva intacto.
func TestNormalizarOpcional(t *testing.T) {
	tests := []struct {
		name  string
		valor string
		quier *string
	}{
		{name: "vacío es ausencia", valor: ""},
		{name: "solo espacios es ausencia", valor: "   "},
		{name: "recorta los bordes", valor: "  Hola  ", quier: ptr("Hola")},
		{name: "colapsa los espacios internos", valor: "Hola   mundo", quier: ptr("Hola mundo")},
		{name: "conserva tildes, ñ, ¿¡ y emojis", valor: "¿Lección de la niñez? 🙌", quier: ptr("¿Lección de la niñez? 🙌")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := normalizarOpcional(test.valor)
			if !mismoOpcional(got, test.quier) {
				t.Errorf("normalizarOpcional(%q) = %v, se esperaba %v", test.valor, valorDe(got), valorDe(test.quier))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Validación del PATCH (FR-019/FR-023)
// ---------------------------------------------------------------------------

// TestPatchMedioFromInput fija la validación del único cuerpo editable del
// módulo (FR-019/FR-020): descripción de hasta 300 caracteres por idioma,
// español por campo, `visibility` dentro del catálogo y `""` —en los dos
// idiomas— como «sin texto» (patrón I6: en español deja la ficha sin
// descripción, en inglés borra la traducción opcional; decisión del
// 2026-10-18).
func TestPatchMedioFromInput(t *testing.T) {
	tests := []struct {
		name        string
		patch       MedioPatch
		campo       string // campo esperado en `details` ("" = válido)
		quierEs     *string
		quierEn     *string
		tocaEs      bool
		tocaEn      bool
		visibilidad *Visibilidad
	}{
		{name: "sin cambios no toca nada", patch: MedioPatch{}},
		{
			name: "descripción en español", patch: MedioPatch{DescriptionEs: ptr("  Una prédica  ")},
			quierEs: ptr("Una prédica"), tocaEs: true,
		},
		{
			name: "descripción en español al límite", patch: MedioPatch{DescriptionEs: ptr(strings.Repeat("á", maxDescripcion))},
			quierEs: ptr(strings.Repeat("á", maxDescripcion)), tocaEs: true,
		},
		{
			name:  "descripción en español vacía deja la ficha sin descripción",
			patch: MedioPatch{DescriptionEs: ptr("")}, tocaEs: true,
		},
		{
			name:  "descripción en español de solo espacios deja la ficha sin descripción",
			patch: MedioPatch{DescriptionEs: ptr("   ")}, tocaEs: true,
		},
		{
			name: "descripción en español de 301", patch: MedioPatch{DescriptionEs: ptr(strings.Repeat("a", maxDescripcion+1))},
			campo: "descriptionEs",
		},
		{
			name: "descripción en inglés", patch: MedioPatch{DescriptionEn: ptr("A sermon")},
			quierEn: ptr("A sermon"), tocaEn: true,
		},
		{
			name: "descripción en inglés vacía borra la traducción", patch: MedioPatch{DescriptionEn: ptr("")},
			tocaEn: true,
		},
		{
			name: "descripción en inglés de solo espacios borra la traducción", patch: MedioPatch{DescriptionEn: ptr("  ")},
			tocaEn: true,
		},
		{
			name: "descripción en inglés de 301", patch: MedioPatch{DescriptionEn: ptr(strings.Repeat("a", maxDescripcion+1))},
			campo: "descriptionEn",
		},
		{
			name: "visibilidad visible", patch: MedioPatch{Visibility: ptr("visible")},
			visibilidad: ptrVisibilidad(VisibilidadVisible),
		},
		{
			name: "visibilidad oculta", patch: MedioPatch{Visibility: ptr("hidden")},
			visibilidad: ptrVisibilidad(VisibilidadOculta),
		},
		{name: "visibilidad fuera del catálogo", patch: MedioPatch{Visibility: ptr("draft")}, campo: "visibility"},
		{
			name:  "tildes, ñ, ¿¡ y emojis intactos",
			patch: MedioPatch{DescriptionEs: ptr("¿Qué dice la Lección de la niñez? 🙌")},
			// El contenido se publica tal cual: la normalización solo recorta
			// los bordes y colapsa espacios.
			quierEs: ptr("¿Qué dice la Lección de la niñez? 🙌"), tocaEs: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := patchMedioFromInput(test.patch)
			if test.campo != "" {
				requireInvalid(t, err, test.campo)
				return
			}
			if err != nil {
				t.Fatalf("patchMedioFromInput devolvió %v", err)
			}
			if got.tocaEs != test.tocaEs || !mismoOpcional(got.descripcionEs, test.quierEs) {
				t.Errorf("español = (%v, toca=%v), se esperaba (%v, toca=%v)",
					valorDe(got.descripcionEs), got.tocaEs, valorDe(test.quierEs), test.tocaEs)
			}
			if got.tocaEn != test.tocaEn || !mismoOpcional(got.descripcionEn, test.quierEn) {
				t.Errorf("inglés = (%v, toca=%v), se esperaba (%v, toca=%v)",
					valorDe(got.descripcionEn), got.tocaEn, valorDe(test.quierEn), test.tocaEn)
			}
			if test.visibilidad == nil {
				if got.visibilidad != nil {
					t.Errorf("visibilidad = %v, se esperaba sin cambio", *got.visibilidad)
				}
			} else if got.visibilidad == nil || *got.visibilidad != *test.visibilidad {
				t.Errorf("visibilidad = %v, se esperaba %v", got.visibilidad, *test.visibilidad)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Invariantes de la ficha fusionada (FR-023)
// ---------------------------------------------------------------------------

// TestValidarEpisodio fija las invariantes que la base de datos también
// defiende (000014) y que el service comprueba antes de escribir: textos dentro
// de sus límites, español por campo y catálogos del contrato.
func TestValidarEpisodio(t *testing.T) {
	base := Episodio{
		Titulo:      "Potencial sin Resultados La Lección de la Higuera",
		PublicadoEn: time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
		Origen:      OrigenVideo,
		Visibilidad: VisibilidadVisible,
	}

	tests := []struct {
		name  string
		mutar func(*Episodio)
		campo string
	}{
		{name: "ficha válida sin descripción"},
		{name: "ficha válida con las dos descripciones", mutar: func(ep *Episodio) {
			ep.DescripcionEs = ptr("Una prédica")
			ep.DescripcionEn = ptr("A sermon")
		}},
		{name: "título vacío", mutar: func(ep *Episodio) { ep.Titulo = "  " }, campo: "title"},
		{name: "título de 301", mutar: func(ep *Episodio) { ep.Titulo = strings.Repeat("a", maxTitulo+1) }, campo: "title"},
		{name: "descripción en español vacía", mutar: func(ep *Episodio) { ep.DescripcionEs = ptr("") }, campo: "descriptionEs"},
		{name: "descripción en español de 301", mutar: func(ep *Episodio) {
			ep.DescripcionEs = ptr(strings.Repeat("a", maxDescripcion+1))
		}, campo: "descriptionEs"},
		{name: "descripción en inglés vacía", mutar: func(ep *Episodio) {
			ep.DescripcionEs = ptr("Una prédica")
			ep.DescripcionEn = ptr(" ")
		}, campo: "descriptionEn"},
		{name: "inglés sin español", mutar: func(ep *Episodio) { ep.DescripcionEn = ptr("A sermon") }, campo: "descriptionEs"},
		{name: "visibilidad fuera del catálogo", mutar: func(ep *Episodio) { ep.Visibilidad = "draft" }, campo: "visibility"},
		{name: "origen fuera del catálogo", mutar: func(ep *Episodio) { ep.Origen = "hologram" }, campo: "origin"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep := base
			if test.mutar != nil {
				test.mutar(&ep)
			}
			err := validarEpisodio(ep)
			if test.campo == "" {
				if err != nil {
					t.Fatalf("validarEpisodio devolvió %v", err)
				}
				return
			}
			requireInvalid(t, err, test.campo)
		})
	}
}

// TestValidarUnion fija la regla «no puedes unir una ficha contigo misma»
// (FR-019/contrato: `otherId` distinto de `{id}` → `400 invalid`).
func TestValidarUnion(t *testing.T) {
	id := uuid.New()
	if err := validarUnion(id, uuid.New()); err != nil {
		t.Fatalf("validarUnion con ids distintos devolvió %v", err)
	}
	requireInvalid(t, validarUnion(id, id), "otherId")
}

// TestValidarMarcaYVisibilidad fija los registros cerrados de la resolución de
// marcas (P9-19: la marca indica el eje que se resuelve).
func TestValidarMarcaYVisibilidad(t *testing.T) {
	for _, marca := range []MarcaRevision{MarcaPosibleDuplicado, MarcaNoDisponible, MarcaFueraDelFeed} {
		if err := validarMarca(marca); err != nil {
			t.Errorf("validarMarca(%q) devolvió %v", marca, err)
		}
	}
	requireInvalid(t, validarMarca("otra"), "mark")
	for _, visibilidad := range []Visibilidad{VisibilidadVisible, VisibilidadOculta} {
		if err := validarVisibilidad(visibilidad); err != nil {
			t.Errorf("validarVisibilidad(%q) devolvió %v", visibilidad, err)
		}
	}
	requireInvalid(t, validarVisibilidad("draft"), "visibility")
}

// ---------------------------------------------------------------------------
// Helpers de prueba
// ---------------------------------------------------------------------------

// requireInvalid comprueba que el error es un `400 invalid` con el campo
// indicado en `details` (el sobre de error del contrato).
func requireInvalid(t *testing.T, err error, campo string) {
	t.Helper()
	if err == nil {
		t.Fatalf("se esperaba un error inválido con details.%s y no hubo error", campo)
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba un *apperr.Error y llegó %T: %v", err, err)
	}
	if domainErr.Kind != apperr.KindInvalid {
		t.Fatalf("se esperaba un error invalid y llegó %v (kind %d)", err, domainErr.Kind)
	}
	if _, ok := domainErr.Details[campo]; !ok {
		t.Fatalf("details no trae %q: %v", campo, domainErr.Details)
	}
}

// requireKind comprueba que el error es un *apperr.Error del kind indicado (404,
// 409…) sin exigir detalles.
func requireKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("se esperaba un error de kind %d y no hubo error", kind)
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba un *apperr.Error y llegó %T: %v", err, err)
	}
	if domainErr.Kind != kind {
		t.Fatalf("se esperaba el kind %d y llegó %d (%v)", kind, domainErr.Kind, err)
	}
}

// ptr devuelve el puntero de un texto (atajo de las tablas de casos).
func ptr(valor string) *string { return &valor }

// ptrVisibilidad devuelve el puntero de una visibilidad.
func ptrVisibilidad(valor Visibilidad) *Visibilidad { return &valor }

// valorDe imprime un opcional para los mensajes de error.
func valorDe(valor *string) string {
	if valor == nil {
		return "nil"
	}
	return *valor
}

// ---------------------------------------------------------------------------
// Cableado del service y contrato del doble (T915)
// ---------------------------------------------------------------------------

// TestNewService fija el cableado del service: los puertos del dominio quedan
// disponibles para sus métodos y un `Logger` ausente usa el logger por defecto
// (patrón de F4/F6/F7).
func TestNewService(t *testing.T) {
	feeds := &fakeFeedSource{}
	probador := &fakeProbador{}
	caratulas := &fakeCaratulas{}
	store := &fakeAlmacen{}
	service := NewService(ServiceDeps{
		Repository:         newFakeRepository(),
		FeedSource:         feeds,
		AvailabilityProber: probador,
		CoverFetcher:       caratulas,
		Store:              store,
	})

	if service.repository == nil {
		t.Error("el service quedó sin repositorio")
	}
	if service.feeds != feeds {
		t.Error("el service quedó sin la fuente de feeds")
	}
	if service.probador != probador {
		t.Error("el service quedó sin el probador de disponibilidad")
	}
	if service.caratulas != caratulas {
		t.Error("el service quedó sin el descargador de carátulas")
	}
	if service.store != store {
		t.Error("el service quedó sin el almacén de archivos")
	}
	if service.logger == nil {
		t.Error("el service quedó sin logger (debía usar el por defecto)")
	}
}

// TestRepositorioFalsoRespetaElPredicadoPublico documenta el contrato del doble:
// lo que el service ve en las pruebas es el MISMO predicado público de R9-7 que
// aplica la consulta real —visible, no absorbida y sin par de duplicados
// abierto— y una marca de disponibilidad sin resolver **sí** sale (FR-017). La
// consulta real se verifica contra PostgreSQL en las pruebas de integración
// (T912/T914).
func TestRepositorioFalsoRespetaElPredicadoPublico(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(
		Episodio{Titulo: "visible"},
		Episodio{Titulo: "oculta", Visibilidad: VisibilidadOculta},
		Episodio{Titulo: "absorbida", UnidoAID: ptrUUID(uuid.New())},
		Episodio{Titulo: "par abierto", DuplicadoDeID: ptrUUID(uuid.New())},
		Episodio{
			Titulo:                   "par resuelto",
			DuplicadoDeID:            ptrUUID(uuid.New()),
			MarcaDuplicadoResueltaEn: ptrTime(time.Now()),
		},
		Episodio{Titulo: "fuera del feed", MarcaDisponibilidad: ptrMarca(MarcaDisponibilidadFueraDelFeed)},
	)

	items, err := repo.ListPublishedMediaItems(context.Background(), paginate.Params{Limit: pageDefault})
	if err != nil {
		t.Fatalf("ListPublishedMediaItems devolvió %v", err)
	}
	quier := []string{"visible", "par resuelto", "fuera del feed"}
	if len(items) != len(quier) {
		t.Fatalf("salieron %d fichas, se esperaban %d: %v", len(items), len(quier), titulos(items))
	}
	for i, titulo := range quier {
		if items[i].Titulo != titulo {
			t.Errorf("ficha %d = %q, se esperaba %q (orden %v)", i, items[i].Titulo, titulo, titulos(items))
		}
	}

	if total, err := repo.CountPublishedMediaItems(context.Background()); err != nil || total != int64(len(quier)) {
		t.Errorf("CountPublishedMediaItems = (%d, %v), se esperaba (%d, nil)", total, err, len(quier))
	}
	if len(repo.publicParams) != 1 || repo.publicParams[0].Limit != pageDefault {
		t.Errorf("parámetros públicos registrados = %v", repo.publicParams)
	}

	portada, err := repo.ListPortadaMediaItems(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListPortadaMediaItems devolvió %v", err)
	}
	if len(portada) != 2 || len(repo.portadaLimits) != 1 || repo.portadaLimits[0] != 2 {
		t.Errorf("portada = %v (límites %v), se esperaban los 2 primeros", titulos(portada), repo.portadaLimits)
	}
}

// titulos imprime los títulos de una lista de fichas para los mensajes de error.
func titulos(episodios []Episodio) []string {
	out := make([]string, 0, len(episodios))
	for _, ep := range episodios {
		out = append(out, ep.Titulo)
	}
	return out
}

// ptrUUID devuelve el puntero de un identificador.
func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }

// ptrTime devuelve el puntero de una fecha.
func ptrTime(fecha time.Time) *time.Time { return &fecha }

// ptrMarca devuelve el puntero de una marca de disponibilidad.
func ptrMarca(marca MarcaDisponibilidad) *MarcaDisponibilidad { return &marca }

// ---------------------------------------------------------------------------
// Repositorio falso (lo usan T916–T921)
// ---------------------------------------------------------------------------

// fakeRepository es el repositorio falso del service: un almacén en memoria que
// imita el comportamiento observable del repository real (el predicado público
// completo, el orden de la sección, los filtros del panel, la idempotencia de
// las fuentes por `(provider, external_id)`, el repunte de la unión y el
// registro de búsquedas) y recuerda lo que se le pidió, para comprobar los
// topes, los parámetros y las filas de auditoría.
type fakeRepository struct {
	episodios []Episodio
	fuentes   []FuenteEpisodio
	syncRuns  []SyncRun

	// Auditoría: `recorded` son las acciones best-effort (denegaciones y
	// rechazos) y `committed` las que acompañan a una mutación en su transacción.
	recorded  []audit.Action
	committed []audit.Action

	// Guionización de errores.
	listErr     error
	getErr      error
	mutationErr error
	sourceErr   error
	syncRunErr  error
	recordErr   error
	repointErr  error
	// auditFalla fuerza el fallo del registro transaccional: la mutación NO se
	// aplica (FR-022, fail-closed).
	auditFalla bool
	// antesDeInsertar es un gancho de las pruebas (p. ej. para cortar el
	// presupuesto de la búsqueda a mitad de la importación).
	antesDeInsertar func()

	// Observaciones.
	publicParams     []paginate.Params
	portadaLimits    []int
	panelFiltros     []FiltroPanel
	panelParams      []paginate.Params
	sourceLookups    []string // "provider/external_id" consultados por idempotencia
	sourcesInserted  []FuenteEpisodio
	repoints         [][2]uuid.UUID
	updates          []uuid.UUID
	lastSyncCalls    int
	publicCalls      int
	portadaCalls     int
	countPublicCalls int
}

var _ Repository = (*fakeRepository)(nil)

func newFakeRepository() *fakeRepository { return &fakeRepository{} }

// seed agrega fichas al almacén sin pasar por el service.
func (f *fakeRepository) seed(episodios ...Episodio) {
	for _, ep := range episodios {
		if ep.ID == uuid.Nil {
			ep.ID = uuid.New()
		}
		if ep.Origen == "" {
			ep.Origen = OrigenVideo
		}
		if ep.Visibilidad == "" {
			ep.Visibilidad = VisibilidadVisible
		}
		if ep.CreadoEn.IsZero() {
			ep.CreadoEn = time.Now()
		}
		ep.ActualizadoEn = ep.CreadoEn
		f.episodios = append(f.episodios, ep)
	}
}

// seedFuentes agrega fuentes de importación al almacén.
func (f *fakeRepository) seedFuentes(fuentes ...FuenteEpisodio) {
	for _, fuente := range fuentes {
		if fuente.ID == uuid.Nil {
			fuente.ID = uuid.New()
		}
		f.fuentes = append(f.fuentes, fuente)
	}
}

// buscar devuelve la ficha del almacén por id.
func (f *fakeRepository) buscar(id uuid.UUID) (int, bool) {
	for i, ep := range f.episodios {
		if ep.ID == id {
			return i, true
		}
	}
	return 0, false
}

// publicable aplica el predicado público COMPLETO de R9-7 (el mismo de la
// consulta real): visible, no absorbida y sin «posible duplicado» abierto. Una
// marca de disponibilidad sin resolver **sí** sale (FR-017).
func publicable(ep Episodio) bool {
	if ep.Visibilidad != VisibilidadVisible || ep.UnidoAID != nil {
		return false
	}
	return ep.DuplicadoDeID == nil || ep.MarcaDuplicadoResueltaEn != nil
}

// ordenSeccion ordena como el `ORDER BY published_at DESC, created_at ASC, id`.
func ordenSeccion(episodios []Episodio) {
	sort.SliceStable(episodios, func(i, j int) bool {
		if !episodios[i].PublicadoEn.Equal(episodios[j].PublicadoEn) {
			return episodios[i].PublicadoEn.After(episodios[j].PublicadoEn)
		}
		if !episodios[i].CreadoEn.Equal(episodios[j].CreadoEn) {
			return episodios[i].CreadoEn.Before(episodios[j].CreadoEn)
		}
		return episodios[i].ID.String() < episodios[j].ID.String()
	})
}

// aplicarAuditoria imita la transacción: o la mutación y sus filas de auditoría,
// o nada (FR-022, fail-closed).
func (f *fakeRepository) aplicarAuditoria(actions []audit.Action) error {
	if f.auditFalla {
		return apperr.Internal(errors.New("no se pudo registrar la acción administrativa"))
	}
	f.committed = append(f.committed, actions...)
	return nil
}

func (f *fakeRepository) ListPublishedMediaItems(_ context.Context, params paginate.Params) ([]Episodio, error) {
	f.publicCalls++
	f.publicParams = append(f.publicParams, params)
	if f.listErr != nil {
		return nil, f.listErr
	}
	items := make([]Episodio, 0, len(f.episodios))
	for _, ep := range f.episodios {
		if publicable(ep) {
			items = append(items, f.conEnlaces(ep))
		}
	}
	ordenSeccion(items)
	return paginar(items, params), nil
}

func (f *fakeRepository) CountPublishedMediaItems(context.Context) (int64, error) {
	f.countPublicCalls++
	if f.listErr != nil {
		return 0, f.listErr
	}
	var total int64
	for _, ep := range f.episodios {
		if publicable(ep) {
			total++
		}
	}
	return total, nil
}

func (f *fakeRepository) ListPortadaMediaItems(_ context.Context, limit int) ([]Episodio, error) {
	f.portadaCalls++
	f.portadaLimits = append(f.portadaLimits, limit)
	if f.listErr != nil {
		return nil, f.listErr
	}
	items := make([]Episodio, 0, len(f.episodios))
	for _, ep := range f.episodios {
		if publicable(ep) {
			items = append(items, f.conEnlaces(ep))
		}
	}
	ordenSeccion(items)
	if limit >= 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

// coincideFiltro imita el `WHERE` de `ListMediaItems` (incluida la semántica de
// `review_mark`: la marca de esa clase **pendiente**).
func coincideFiltro(ep Episodio, filtro FiltroPanel) bool {
	if filtro.Visibilidad != nil && ep.Visibilidad != *filtro.Visibilidad {
		return false
	}
	if !filtro.IncluirAbsorbidas && ep.UnidoAID != nil {
		return false
	}
	if filtro.MarcaRevision != nil {
		switch *filtro.MarcaRevision {
		case MarcaPosibleDuplicado:
			if ep.DuplicadoDeID == nil || ep.MarcaDuplicadoResueltaEn != nil {
				return false
			}
		case MarcaNoDisponible, MarcaFueraDelFeed:
			clase, ok := filtro.MarcaRevision.MarcaDisponibilidad()
			if !ok || ep.MarcaDisponibilidad == nil || *ep.MarcaDisponibilidad != clase ||
				ep.MarcaDisponibilidadResueltaEn != nil {
				return false
			}
		}
	}
	return true
}

func (f *fakeRepository) ListMediaItems(_ context.Context, filtro FiltroPanel, params paginate.Params) ([]Episodio, error) {
	f.panelFiltros = append(f.panelFiltros, filtro)
	f.panelParams = append(f.panelParams, params)
	if f.listErr != nil {
		return nil, f.listErr
	}
	items := make([]Episodio, 0, len(f.episodios))
	for _, ep := range f.episodios {
		if coincideFiltro(ep, filtro) {
			items = append(items, f.conEnlaces(ep))
		}
	}
	ordenSeccion(items)
	return paginar(items, params), nil
}

func (f *fakeRepository) CountMediaItems(_ context.Context, filtro FiltroPanel) (int64, error) {
	if f.listErr != nil {
		return 0, f.listErr
	}
	var total int64
	for _, ep := range f.episodios {
		if coincideFiltro(ep, filtro) {
			total++
		}
	}
	return total, nil
}

func (f *fakeRepository) GetMediaItemByID(_ context.Context, id uuid.UUID) (Episodio, bool, error) {
	if f.getErr != nil {
		return Episodio{}, false, f.getErr
	}
	if i, ok := f.buscar(id); ok {
		return f.conEnlaces(f.episodios[i]), true, nil
	}
	return Episodio{}, false, nil
}

// conEnlaces resuelve `VideoURL`/`AudioURL` desde las fuentes de la ficha, igual
// que el `LEFT JOIN` por proveedor de las consultas reales (una ficha = 1 video
// y/o 1 audio). Una ficha sin fuentes conserva lo que se le haya sembrado.
func (f *fakeRepository) conEnlaces(episodio Episodio) Episodio {
	for _, fuente := range f.fuentes {
		if fuente.EpisodioID != episodio.ID {
			continue
		}
		switch fuente.Proveedor {
		case ProveedorYouTube:
			episodio.VideoURL = fuente.Enlace
		case ProveedorPodcast:
			episodio.AudioURL = fuente.Enlace
		}
	}
	return episodio
}

func (f *fakeRepository) InsertMediaItem(_ context.Context, episodio Episodio, actions ...audit.Action) (Episodio, error) {
	if f.antesDeInsertar != nil {
		f.antesDeInsertar()
	}
	if f.mutationErr != nil {
		return Episodio{}, f.mutationErr
	}
	if err := f.aplicarAuditoria(actions); err != nil {
		return Episodio{}, err
	}
	episodio.ID = uuid.New()
	episodio.CreadoEn, episodio.ActualizadoEn = time.Now(), time.Now()
	f.episodios = append(f.episodios, episodio)
	return episodio, nil
}

func (f *fakeRepository) UpdateMediaItem(_ context.Context, episodio Episodio, actions ...audit.Action) (Episodio, error) {
	if f.mutationErr != nil {
		return Episodio{}, f.mutationErr
	}
	i, ok := f.buscar(episodio.ID)
	if !ok {
		return Episodio{}, apperr.NotFound(messageEpisodioNoExiste)
	}
	if err := f.aplicarAuditoria(actions); err != nil {
		return Episodio{}, err
	}
	episodio.CreadoEn = f.episodios[i].CreadoEn
	episodio.ActualizadoEn = time.Now()
	f.episodios[i] = episodio
	f.updates = append(f.updates, episodio.ID)
	return episodio, nil
}

// UpdateMediaItemCoverIfNull imita la escritura condicional del repository real:
// escribe la carátula **solo** si la ficha todavía no tiene (y solo ese campo: lo
// demás del equipo queda intacto) y devuelve false cuando ya había carátula.
func (f *fakeRepository) UpdateMediaItemCoverIfNull(_ context.Context, id uuid.UUID, archivo string) (Episodio, bool, error) {
	if f.mutationErr != nil {
		return Episodio{}, false, f.mutationErr
	}
	i, ok := f.buscar(id)
	if !ok || f.episodios[i].Caratula != nil {
		return Episodio{}, false, nil
	}
	f.episodios[i].Caratula = &archivo
	f.episodios[i].ActualizadoEn = time.Now()
	f.updates = append(f.updates, id)
	return f.episodios[i], true, nil
}

func (f *fakeRepository) RepointMediaItemSources(_ context.Context, resultante, absorbida Episodio, actions ...audit.Action) ([]FuenteEpisodio, error) {
	if f.repointErr != nil {
		return nil, f.repointErr
	}
	if f.mutationErr != nil {
		return nil, f.mutationErr
	}
	ri, ok := f.buscar(resultante.ID)
	if !ok {
		return nil, apperr.NotFound(messageEpisodioNoExiste)
	}
	ai, ok := f.buscar(absorbida.ID)
	if !ok {
		return nil, apperr.NotFound(messageEpisodioNoExiste)
	}
	if err := f.aplicarAuditoria(actions); err != nil {
		return nil, err
	}
	movidas := make([]FuenteEpisodio, 0, 2)
	for i, fuente := range f.fuentes {
		if fuente.EpisodioID == absorbida.ID {
			f.fuentes[i].EpisodioID = resultante.ID
			movidas = append(movidas, f.fuentes[i])
		}
	}
	resultante.CreadoEn, resultante.ActualizadoEn = f.episodios[ri].CreadoEn, time.Now()
	absorbida.CreadoEn, absorbida.ActualizadoEn = f.episodios[ai].CreadoEn, time.Now()
	f.episodios[ri] = resultante
	f.episodios[ai] = absorbida
	f.repoints = append(f.repoints, [2]uuid.UUID{resultante.ID, absorbida.ID})
	return movidas, nil
}

func (f *fakeRepository) InsertMediaEpisodeSource(_ context.Context, fuente FuenteEpisodio) (FuenteEpisodio, bool, error) {
	if f.sourceErr != nil {
		return FuenteEpisodio{}, false, f.sourceErr
	}
	f.sourceLookups = append(f.sourceLookups, string(fuente.Proveedor)+"/"+fuente.IDExterno)
	for _, guardada := range f.fuentes {
		if guardada.Proveedor == fuente.Proveedor && guardada.IDExterno == fuente.IDExterno {
			// `ON CONFLICT (provider, external_id) DO NOTHING`: no duplica ni
			// actualiza (instantánea R9-6) y no es un error.
			return FuenteEpisodio{}, false, nil
		}
	}
	fuente.ID = uuid.New()
	f.fuentes = append(f.fuentes, fuente)
	f.sourcesInserted = append(f.sourcesInserted, fuente)
	return fuente, true, nil
}

func (f *fakeRepository) ListMediaEpisodeSourcesByProvider(_ context.Context, proveedor Proveedor) ([]FuenteEpisodio, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	fuentes := make([]FuenteEpisodio, 0, len(f.fuentes))
	for _, fuente := range f.fuentes {
		if fuente.Proveedor == proveedor {
			fuentes = append(fuentes, fuente)
		}
	}
	return fuentes, nil
}

func (f *fakeRepository) GetMediaEpisodeSource(_ context.Context, proveedor Proveedor, idExterno string) (FuenteEpisodio, bool, error) {
	if f.getErr != nil {
		return FuenteEpisodio{}, false, f.getErr
	}
	for _, fuente := range f.fuentes {
		if fuente.Proveedor == proveedor && fuente.IDExterno == idExterno {
			return fuente, true, nil
		}
	}
	return FuenteEpisodio{}, false, nil
}

func (f *fakeRepository) ListMediaItemSourcesByItem(_ context.Context, itemID uuid.UUID) ([]FuenteEpisodio, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	fuentes := make([]FuenteEpisodio, 0, 2)
	for _, fuente := range f.fuentes {
		if fuente.EpisodioID == itemID {
			fuentes = append(fuentes, fuente)
		}
	}
	return fuentes, nil
}

func (f *fakeRepository) InsertMediaSyncRun(_ context.Context, run SyncRun, actions ...audit.Action) (SyncRun, error) {
	if f.syncRunErr != nil {
		return SyncRun{}, f.syncRunErr
	}
	if err := f.aplicarAuditoria(actions); err != nil {
		return SyncRun{}, err
	}
	run.ID = uuid.New()
	f.syncRuns = append(f.syncRuns, run)
	return run, nil
}

func (f *fakeRepository) GetLastMediaSyncRun(context.Context) (SyncRun, bool, error) {
	f.lastSyncCalls++
	if f.syncRunErr != nil {
		return SyncRun{}, false, f.syncRunErr
	}
	if len(f.syncRuns) == 0 {
		return SyncRun{}, false, nil
	}
	last := f.syncRuns[0]
	for _, run := range f.syncRuns[1:] {
		if run.IniciadoEn.After(last.IniciadoEn) {
			last = run
		}
	}
	return last, true, nil
}

func (f *fakeRepository) RecordAction(_ context.Context, action audit.Action) error {
	if f.recordErr != nil {
		return f.recordErr
	}
	f.recorded = append(f.recorded, action)
	return nil
}

// paginar aplica limit/offset a una lista ya ordenada (§8.1.3).
func paginar[T any](items []T, params paginate.Params) []T {
	start := params.Offset
	if start > len(items) {
		return []T{}
	}
	end := start + params.Limit
	if params.Limit <= 0 || end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

// ---------------------------------------------------------------------------
// Dobles de la red saliente y del almacén (T920/T921)
// ---------------------------------------------------------------------------

// fakeProbador es el doble programable de `AvailabilityProber`: devuelve el
// resultado por enlace (o por proveedor) sin salir a la red y recuerda qué se le
// pidió (para fijar la cola de comprobación de FR-017). La búsqueda comprueba con
// varias comprobaciones en vuelo (P9-3): el registro va con mutex.
type fakeProbador struct {
	youtube    map[string]ResultadoDisponibilidad
	podcast    map[string]ResultadoDisponibilidad
	porDefecto ResultadoDisponibilidad
	err        error

	mu          sync.Mutex
	comprobados []string
}

var _ AvailabilityProber = (*fakeProbador)(nil)

func (p *fakeProbador) ProbeYouTube(_ context.Context, videoID string) (ResultadoDisponibilidad, error) {
	p.anotar("youtube/" + videoID)
	if p.err != nil {
		return DisponibilidadInconclusa, p.err
	}
	if r, ok := p.youtube[videoID]; ok {
		return r, nil
	}
	return p.porDefecto, nil
}

func (p *fakeProbador) ProbePodcast(_ context.Context, enlace string) (ResultadoDisponibilidad, error) {
	p.anotar("podcast/" + enlace)
	if p.err != nil {
		return DisponibilidadInconclusa, p.err
	}
	if r, ok := p.podcast[enlace]; ok {
		return r, nil
	}
	return p.porDefecto, nil
}

// anotar registra una comprobación (seguro con varias comprobaciones en vuelo).
func (p *fakeProbador) anotar(clave string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.comprobados = append(p.comprobados, clave)
}

// comprobadas devuelve una copia del registro de comprobaciones.
func (p *fakeProbador) comprobadas() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.comprobados...)
}

// fakeCaratulas es el doble programable de `CoverFetcher`: guarda en memoria lo
// que se le pidió y devuelve un nombre generado (o error, sin abortar la
// importación). Las descargas van con varias en vuelo (P9-3): el registro va con
// mutex.
type fakeCaratulas struct {
	archivo string
	// porURL devuelve un nombre distinto por carátula pedida: permite comprobar
	// cuál de las dos carátulas de un mismo sermón se copió.
	porURL map[string]string
	err    error

	mu      sync.Mutex
	pedidas []string
}

var _ CoverFetcher = (*fakeCaratulas)(nil)

func (c *fakeCaratulas) FetchCover(_ context.Context, caratulaURL string) (string, error) {
	c.mu.Lock()
	c.pedidas = append(c.pedidas, caratulaURL)
	c.mu.Unlock()
	if c.err != nil {
		return "", c.err
	}
	if nombre, ok := c.porURL[caratulaURL]; ok {
		return nombre, nil
	}
	if c.archivo == "" {
		return "img_" + uuid.NewString() + ".jpg", nil
	}
	return c.archivo, nil
}

// descargadas devuelve una copia del registro de descargas.
func (c *fakeCaratulas) descargadas() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.pedidas...)
}

// fakeAlmacen es el doble del almacén de archivos para la unión: recuerda los
// borrados best-effort de la carátula que quedó sin uso (el camino feliz del
// guardado lo cubre `fakeStore` de caratula_test.go, T910).
type fakeAlmacen struct {
	borrados []string
	err      error
}

var _ storage.Store = (*fakeAlmacen)(nil)

func (a *fakeAlmacen) Save(context.Context, []byte) (storage.File, error) {
	return storage.File{Name: "img_" + uuid.NewString() + ".jpg"}, a.err
}

func (a *fakeAlmacen) Open(context.Context, string) (io.ReadCloser, error) { return nil, a.err }

func (a *fakeAlmacen) Delete(_ context.Context, name string) error {
	a.borrados = append(a.borrados, name)
	return a.err
}
