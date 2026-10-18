package medios

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/paginate"
)

// service_public_test.go cubre la cara pública del dominio (T916/T917): la
// sección paginada `GET /api/v1/medios` y el extracto de la portada
// `GET /api/v1/portada/medios`. Son las pruebas más importantes del módulo: fijan
// SC-002 (0 fichas no publicables expuestas), SC-011 (0 huecos y datos idénticos
// entre idiomas) y SC-015 (el bloque muestra exactamente los 5 primeros).

// TestGetMediosSoloPublicables fija la única vía pública (FR-009/SC-002): una
// ficha oculta, absorbida o con un «posible duplicado» ABIERTO no sale jamás; una
// marcada «ya no aparece en el feed» **sí** sale (FR-017) y un par **resuelto**
// también.
func TestGetMediosSoloPublicables(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(
		Episodio{Titulo: "visible"},
		Episodio{Titulo: "oculta por el equipo", Visibilidad: VisibilidadOculta},
		Episodio{Titulo: "absorbida", UnidoAID: ptrUUID(uuid.New())},
		Episodio{Titulo: "par abierto", DuplicadoDeID: ptrUUID(uuid.New())},
		Episodio{
			Titulo:                   "par resuelto",
			DuplicadoDeID:            ptrUUID(uuid.New()),
			MarcaDuplicadoResueltaEn: ptrTime(time.Now()),
		},
		Episodio{
			Titulo:              "ya no aparece en el feed",
			MarcaDisponibilidad: ptrMarca(MarcaDisponibilidadFueraDelFeed),
		},
		Episodio{
			Titulo:                        "marca de disponibilidad resuelta",
			MarcaDisponibilidad:           ptrMarca(MarcaDisponibilidadFueraDelFeed),
			MarcaDisponibilidadResueltaEn: ptrTime(time.Now()),
		},
	)
	service := serviceDePrueba(repo)

	out, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	quier := []string{
		"visible",
		"par resuelto",
		"ya no aparece en el feed",
		"marca de disponibilidad resuelta",
	}
	if len(out.Items) != len(quier) {
		t.Fatalf("salieron %d fichas, se esperaban %d: %v", len(out.Items), len(quier), titulosPublicos(out.Items))
	}
	for i, titulo := range quier {
		if out.Items[i].Titulo != titulo {
			t.Errorf("ficha %d = %q, se esperaba %q (%v)", i, out.Items[i].Titulo, titulo, titulosPublicos(out.Items))
		}
	}
	if out.Total != int64(len(quier)) {
		t.Errorf("total = %d, se esperaba %d", out.Total, len(quier))
	}
	if out.Lang != LangES {
		t.Errorf("lang = %q, se esperaba %q", out.Lang, LangES)
	}
	if out.Limit != pageDefault || out.Offset != 0 {
		t.Errorf("tanda = {%d, %d}, se esperaba {%d, 0}", out.Limit, out.Offset, pageDefault)
	}
}

// TestGetMediosOrdenDeLaSeccion fija el orden público (FR-006/default #11):
// `published_at DESC` y, ante la misma fecha, el orden de importación
// (`created_at ASC`, `id`).
func TestGetMediosOrdenDeLaSeccion(t *testing.T) {
	mismaFecha := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)
	antes := time.Date(2026, 7, 1, 15, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.seed(
		Episodio{Titulo: "segunda importada", PublicadoEn: mismaFecha, CreadoEn: mismaFecha.Add(time.Hour)},
		Episodio{Titulo: "más antigua", PublicadoEn: antes, CreadoEn: antes},
		Episodio{Titulo: "primera importada", PublicadoEn: mismaFecha, CreadoEn: mismaFecha},
	)
	service := serviceDePrueba(repo)

	out, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	quier := []string{"primera importada", "segunda importada", "más antigua"}
	for i, titulo := range quier {
		if out.Items[i].Titulo != titulo {
			t.Fatalf("orden = %v, se esperaba %v", titulosPublicos(out.Items), quier)
		}
	}
}

// TestGetMediosPaginacion fija la tanda (default 12, tope 50) y el sobre
// `{lang, items, total, limit, offset}` de «Cargar más» (R9-10/§8.1.3).
func TestGetMediosPaginacion(t *testing.T) {
	repo := newFakeRepository()
	for i := 0; i < 13; i++ {
		repo.seed(Episodio{
			Titulo:      "episodio " + string(rune('a'+i)),
			PublicadoEn: time.Date(2026, 7, 1, 15, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour),
		})
	}
	service := serviceDePrueba(repo)

	primera, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	if len(primera.Items) != pageDefault {
		t.Errorf("primera tanda = %d fichas, se esperaban %d", len(primera.Items), pageDefault)
	}
	if primera.Total != 13 {
		t.Errorf("total = %d, se esperaba 13", primera.Total)
	}

	segunda, err := service.GetMedios(context.Background(), LangES, 0, pageDefault)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	if len(segunda.Items) != 1 {
		t.Errorf("segunda tanda = %d fichas, se esperaba 1", len(segunda.Items))
	}
	if segunda.Offset != pageDefault {
		t.Errorf("offset = %d, se esperaba %d", segunda.Offset, pageDefault)
	}
	if segunda.Items[0].Titulo != "episodio a" {
		t.Errorf("la segunda tanda no continúa el orden: %q", segunda.Items[0].Titulo)
	}
	if params := paramsPublicos(repo); params.Limit != pageDefault || params.Offset != pageDefault {
		t.Errorf("parámetros de la última tanda = %+v, se esperaban {12, 12}", params)
	}

	tope, err := service.GetMedios(context.Background(), LangES, maxPage+10, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	if tope.Limit != maxPage {
		t.Errorf("limit = %d, se esperaba el tope %d", tope.Limit, maxPage)
	}
	if _, err := service.GetMedios(context.Background(), LangES, -1, 0); err == nil {
		t.Error("un limit negativo debía ser invalid")
	}
	if _, err := service.GetMedios(context.Background(), LangES, 0, -1); err == nil {
		t.Error("un offset negativo debía ser invalid")
	}
	if _, err := service.GetMedios(context.Background(), "fr", 0, 0); err == nil {
		t.Error("un idioma fuera de es|en debía ser invalid")
	}
}

// TestGetMediosBilingue fija el fallback `en → es` **por campo** resuelto en el
// servidor (FR-025/SC-011): la descripción en inglés solo si existe, la española
// si no, y la clave omitida cuando el feed no traía descripción (precisión 3).
// El nombre, la fecha, la carátula y los enlaces son idénticos en ambos idiomas
// (FR-026).
func TestGetMediosBilingue(t *testing.T) {
	publicado := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.seed(
		Episodio{
			Titulo:        "Traducida",
			PublicadoEn:   publicado,
			DescripcionEs: ptr("Una prédica"),
			DescripcionEn: ptr("A sermon"),
			Caratula:      ptr("img_11111111-1111-1111-1111-111111111111.jpg"),
			VideoURL:      "https://www.youtube.com/watch?v=SimienteAA1",
			AudioURL:      "https://podcasters.spotify.com/pod/show/simiente/episodes/abc",
		},
		Episodio{
			Titulo:        "Solo español",
			PublicadoEn:   publicado.Add(-time.Hour),
			DescripcionEs: ptr("Sin traducción"),
		},
		Episodio{
			Titulo:      "Sin descripción",
			PublicadoEn: publicado.Add(-2 * time.Hour),
		},
	)
	service := serviceDePrueba(repo)

	espanol, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios(es) devolvió %v", err)
	}
	ingles, err := service.GetMedios(context.Background(), LangEN, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios(en) devolvió %v", err)
	}

	if espanol.Items[0].Descripcion != "Una prédica" {
		t.Errorf("español = %q, se esperaba el español", espanol.Items[0].Descripcion)
	}
	if ingles.Items[0].Descripcion != "A sermon" {
		t.Errorf("inglés = %q, se esperaba la traducción", ingles.Items[0].Descripcion)
	}
	if ingles.Items[1].Descripcion != "Sin traducción" {
		t.Errorf("respaldo = %q, se esperaba el español (nunca un campo vacío)", ingles.Items[1].Descripcion)
	}
	if ingles.Items[2].Descripcion != "" {
		t.Errorf("sin descripción = %q, se esperaba la clave omitida", ingles.Items[2].Descripcion)
	}

	// Los datos no traducibles son un único valor (FR-026).
	for i := range espanol.Items {
		if espanol.Items[i].Titulo != ingles.Items[i].Titulo ||
			!espanol.Items[i].PublicadoEn.Equal(ingles.Items[i].PublicadoEn) ||
			espanol.Items[i].CaratulaURL != ingles.Items[i].CaratulaURL ||
			espanol.Items[i].VideoURL != ingles.Items[i].VideoURL ||
			espanol.Items[i].AudioURL != ingles.Items[i].AudioURL {
			t.Errorf("la ficha %d cambia con el idioma: %+v vs %+v", i, espanol.Items[i], ingles.Items[i])
		}
	}
}

// TestMedioPublicoOmiteClavesVaciais fija SC-010 en la ficha: la carátula, los
// enlaces y la descripción se **omiten** cuando no aplican (nunca `null` ni una
// línea vacía), y la carátula viaja como ruta relativa del propio sitio
// (SC-005: 0 recursos de terceros).
func TestMedioPublicoOmiteClavesVaciais(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(Episodio{
		Titulo:      "Solo video",
		PublicadoEn: time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
		Caratula:    ptr("img_22222222-2222-2222-2222-222222222222.png"),
		VideoURL:    "https://www.youtube.com/watch?v=SimienteAA1",
	})
	service := serviceDePrueba(repo)

	out, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	ficha := out.Items[0]
	if ficha.CaratulaURL != "/api/v1/media/img_22222222-2222-2222-2222-222222222222.png" {
		t.Errorf("coverUrl = %q, se esperaba la ruta relativa del sitio", ficha.CaratulaURL)
	}

	datos, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	cuerpo := string(datos)
	if strings.Contains(cuerpo, `"audioUrl"`) {
		t.Errorf("la ficha sin audio no debe traer audioUrl: %s", cuerpo)
	}
	if strings.Contains(cuerpo, `"description"`) {
		t.Errorf("la ficha sin descripción no debe traer description: %s", cuerpo)
	}
	if strings.Contains(cuerpo, "null") {
		t.Errorf("las claves opcionales nunca viajan como null: %s", cuerpo)
	}

	// Ninguna clave interna del panel puede salir por la vía pública.
	for _, clave := range []string{
		"visibility", "duplicateOfId", "duplicateMarkResolvedAt", "availabilityMark",
		"availabilityResult", "availabilityMarkResolvedAt", "availabilityCheckedAt",
		"mergedIntoId", "sources", "descriptionEdited", "feedTitle",
	} {
		if strings.Contains(cuerpo, clave) {
			t.Errorf("el DTO público no debe exponer %q: %s", clave, cuerpo)
		}
	}
}

// TestGetMediosSinPublicables fija el caso «nada visible» (FR-008/SC-016): la
// respuesta sigue siendo `200` con `items` **vacío** (nunca `null`) y `total: 0`;
// la redirección de `/medios` es del frontend.
func TestGetMediosSinPublicables(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(Episodio{Titulo: "oculta", Visibilidad: VisibilidadOculta})
	service := serviceDePrueba(repo)

	out, err := service.GetMedios(context.Background(), LangES, 0, 0)
	if err != nil {
		t.Fatalf("GetMedios devolvió %v", err)
	}
	if out.Items == nil {
		t.Error("items quedó nil: debe ser un arreglo vacío (nunca null)")
	}
	if len(out.Items) != 0 || out.Total != 0 {
		t.Errorf("salieron %d fichas (total %d), se esperaba 0", len(out.Items), out.Total)
	}
	datos, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(datos), `"items":[]`) {
		t.Errorf("items debe viajar como arreglo vacío: %s", string(datos))
	}
}

// TestGetMediosPropagaErroresDelRepositorio comprueba que un fallo de lectura no
// se convierte en una respuesta vacía (una caída transitoria no puede parecer
// «no hay nada»: el frontend no redirige ante error, P9-16).
func TestGetMediosPropagaErroresDelRepositorio(t *testing.T) {
	repo := newFakeRepository()
	repo.listErr = errDePrueba
	service := serviceDePrueba(repo)

	if _, err := service.GetMedios(context.Background(), LangES, 0, 0); err == nil {
		t.Fatal("se esperaba el error del repositorio")
	}
}

// ---------------------------------------------------------------------------
// Helpers de prueba de la cara pública
// ---------------------------------------------------------------------------

// serviceDePrueba arma el service con el repositorio falso y los dobles de la red
// saliente (nada de red en las pruebas).
func serviceDePrueba(repo Repository) *service {
	return NewService(ServiceDeps{
		Repository:         repo,
		FeedSource:         &fakeFeedSource{},
		AvailabilityProber: &fakeProbador{},
		CoverFetcher:       &fakeCaratulas{},
		Store:              &fakeAlmacen{},
	})
}

// titulosPublicos imprime los títulos de una lista de fichas públicas.
func titulosPublicos(items []MedioPublico) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Titulo)
	}
	return out
}

// errDePrueba es el fallo simulado de una lectura del repositorio.
var errDePrueba = errors.New("fallo de lectura simulado")

// paramsPublicos recuerda los parámetros de paginación que el service pasó al
// repositorio en la última llamada.
func paramsPublicos(repo *fakeRepository) paginate.Params {
	if len(repo.publicParams) == 0 {
		return paginate.Params{}
	}
	return repo.publicParams[len(repo.publicParams)-1]
}
