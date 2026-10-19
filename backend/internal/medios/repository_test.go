package medios

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/database"
)

// text devuelve un pgtype.Text válido (columna con valor).
func text(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }

// textOf es un atajo para construir un *string.
func textOf(s string) *string { return &s }

// timeOf es un atajo para construir un *time.Time.
func timeOf(t time.Time) *time.Time { return &t }

// stamp devuelve un pgtype.Timestamptz válido.
func stamp(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }

// errorKind clasifica el error de dominio (mismo helper que usan las pruebas de
// los otros dominios: la traducción a HTTP vive solo en httpserver).
func errorKind(err error) apperr.Kind {
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		return domainErr.Kind
	}
	return apperr.KindInternal
}

// errorDetails devuelve el detalle seguro del error de dominio (nil si no es un
// error de dominio o no lleva detalle).
func errorDetails(err error) map[string]any {
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		return domainErr.Details
	}
	return nil
}

// TestUUIDHelpers verifica la traducción de UUID anulables (§8.1.6).
func TestUUIDHelpers(t *testing.T) {
	id := uuid.New()
	if got := uuidValue(pgtype.UUID{Bytes: [16]byte(id), Valid: true}); got != id {
		t.Errorf("uuidValue = %v, se esperaba %v", got, id)
	}
	if got := uuidPtr(pgtype.UUID{}); got != nil {
		t.Errorf("uuidPtr(NULL) = %v, se esperaba nil", got)
	}
	if got := uuidPtr(pgtype.UUID{Bytes: [16]byte(id), Valid: true}); got == nil || *got != id {
		t.Errorf("uuidPtr(válido) = %v", got)
	}
	if got := pgUUID(id); !got.Valid || uuid.UUID(got.Bytes) != id {
		t.Errorf("pgUUID = %+v", got)
	}
	if got := pgUUIDPtr(nil); got.Valid {
		t.Errorf("pgUUIDPtr(nil) = %+v, se esperaba NULL", got)
	}
	if got := pgUUIDPtr(&id); !got.Valid || uuid.UUID(got.Bytes) != id {
		t.Errorf("pgUUIDPtr(válido) = %+v", got)
	}
}

// TestTextHelpers verifica la traducción de textos opcionales: "" (y solo
// espacios) se guarda como NULL (analyze I6), NULL se lee como nil/"" y un valor
// no vacío viaja como TEXT.
func TestTextHelpers(t *testing.T) {
	if got := textPtr(pgtype.Text{}); got != nil {
		t.Errorf("textPtr(NULL) = %v, se esperaba nil", got)
	}
	if got := textPtr(text("La lección de la higuera")); got == nil || *got != "La lección de la higuera" {
		t.Errorf("textPtr(válido) = %v", got)
	}
	if got := textValue(pgtype.Text{}); got != "" {
		t.Errorf("textValue(NULL) = %q, se esperaba \"\"", got)
	}
	if got := textValue(text("https://youtube.com/watch?v=abc")); got != "https://youtube.com/watch?v=abc" {
		t.Errorf("textValue(válido) = %q", got)
	}
	if got := pgTextPtr(nil); got.Valid {
		t.Errorf("pgTextPtr(nil) = %+v, se esperaba NULL", got)
	}
	if got := pgTextPtr(textOf("x")); !got.Valid || got.String != "x" {
		t.Errorf("pgTextPtr = %+v", got)
	}
	if got := pgText(""); got.Valid {
		t.Errorf("pgText(\"\") = %+v, se esperaba NULL", got)
	}
	if got := pgText("   "); got.Valid {
		t.Errorf("pgText(espacios) = %+v, se esperaba NULL", got)
	}
	if got := pgText("Una descripción"); !got.Valid || got.String != "Una descripción" {
		t.Errorf("pgText(válido) = %+v", got)
	}
	if got := pgLabel(""); got.Valid {
		t.Errorf("pgLabel(\"\") = %+v, se esperaba NULL", got)
	}
	if got := pgLabel("Medios · Búsqueda"); !got.Valid {
		t.Errorf("pgLabel(válido) = %+v", got)
	}
}

// TestTimeHelpers verifica la traducción de fechas: una columna anulable pasa a
// *time.Time (nil si NULL) y un *time.Time del dominio a NULL o a TIMESTAMPTZ.
func TestTimeHelpers(t *testing.T) {
	moment := time.Date(2026, 7, 3, 1, 22, 48, 0, time.UTC)

	if got := timePtr(pgtype.Timestamptz{}); got != nil {
		t.Errorf("timePtr(NULL) = %v, se esperaba nil", got)
	}
	if got := timePtr(stamp(moment)); got == nil || !got.Equal(moment) {
		t.Errorf("timePtr(válido) = %v", got)
	}
	if got := pgTime(moment); !got.Valid || !got.Time.Equal(moment) {
		t.Errorf("pgTime = %+v", got)
	}
	if got := pgTimePtr(nil); got.Valid {
		t.Errorf("pgTimePtr(nil) = %+v, se esperaba NULL", got)
	}
	if got := pgTimePtr(timeOf(moment)); !got.Valid || !got.Time.Equal(moment) {
		t.Errorf("pgTimePtr(válido) = %+v", got)
	}
}

// TestEnumHelpers verifica la traducción de los enums del dominio: una columna
// TEXT NOT NULL pasa al enum, una anulable a *enum (nil si NULL, **nunca** el
// valor cero) y un *enum del dominio a NULL o al texto de la columna.
func TestEnumHelpers(t *testing.T) {
	if got := enumValue[Visibilidad]("hidden"); got != VisibilidadOculta {
		t.Errorf("enumValue = %q, se esperaba hidden", got)
	}
	if got := enumPtr[Visibilidad](pgtype.Text{}); got != nil {
		t.Errorf("enumPtr(NULL) = %v, se esperaba nil", got)
	}
	if got := enumPtr[Visibilidad](text("visible")); got == nil || *got != VisibilidadVisible {
		t.Errorf("enumPtr(válido) = %v", got)
	}

	marca := MarcaDisponibilidadFueraDelFeed
	if got := pgEnumPtr(&marca); !got.Valid || got.String != "missing_from_feed" {
		t.Errorf("pgEnumPtr(válido) = %+v", got)
	}
	if got := pgEnumPtr[MarcaDisponibilidad](nil); got.Valid {
		t.Errorf("pgEnumPtr(nil) = %+v, se esperaba NULL", got)
	}
}

// TestRegistriesValid fija los registros cerrados del contrato: visibilidad,
// marcas de los dos ejes, proveedores, orígenes y estados de feed.
func TestRegistriesValid(t *testing.T) {
	visibilidades := []struct {
		value Visibilidad
		want  bool
	}{
		{VisibilidadVisible, true}, {VisibilidadOculta, true},
		{Visibilidad("draft"), false}, {Visibilidad(""), false},
	}
	for _, tt := range visibilidades {
		if got := tt.value.Valid(); got != tt.want {
			t.Errorf("Visibilidad(%q).Valid() = %v, se esperaba %v", tt.value, got, tt.want)
		}
	}

	marcas := []struct {
		value MarcaRevision
		want  bool
	}{
		{MarcaPosibleDuplicado, true}, {MarcaNoDisponible, true}, {MarcaFueraDelFeed, true},
		{MarcaRevision("review"), false}, {MarcaRevision(""), false},
	}
	for _, tt := range marcas {
		if got := tt.value.Valid(); got != tt.want {
			t.Errorf("MarcaRevision(%q).Valid() = %v, se esperaba %v", tt.value, got, tt.want)
		}
	}

	disponibilidad := []struct {
		value MarcaDisponibilidad
		want  bool
	}{
		{MarcaDisponibilidadNoDisponible, true}, {MarcaDisponibilidadFueraDelFeed, true},
		{MarcaDisponibilidad("possible_duplicate"), false}, {MarcaDisponibilidad(""), false},
	}
	for _, tt := range disponibilidad {
		if got := tt.value.Valid(); got != tt.want {
			t.Errorf("MarcaDisponibilidad(%q).Valid() = %v, se esperaba %v", tt.value, got, tt.want)
		}
	}

	proveedores := []struct {
		value Proveedor
		want  bool
	}{
		{ProveedorYouTube, true}, {ProveedorPodcast, true},
		{Proveedor("vimeo"), false}, {Proveedor(""), false},
	}
	for _, tt := range proveedores {
		if got := tt.value.Valid(); got != tt.want {
			t.Errorf("Proveedor(%q).Valid() = %v, se esperaba %v", tt.value, got, tt.want)
		}
	}

	origenes := []struct {
		value OrigenMedio
		want  bool
	}{
		{OrigenVideo, true}, {OrigenAudio, true}, {OrigenAmbos, true},
		{OrigenMedio("stream"), false}, {OrigenMedio(""), false},
	}
	for _, tt := range origenes {
		if got := tt.value.Valid(); got != tt.want {
			t.Errorf("OrigenMedio(%q).Valid() = %v, se esperaba %v", tt.value, got, tt.want)
		}
	}

	estados := []struct {
		value EstadoFeed
		want  bool
	}{
		{FeedOK, true}, {FeedFallido, true},
		{EstadoFeed("partial"), false}, {EstadoFeed(""), false},
	}
	for _, tt := range estados {
		if got := tt.value.Valid(); got != tt.want {
			t.Errorf("EstadoFeed(%q).Valid() = %v, se esperaba %v", tt.value, got, tt.want)
		}
	}
}

// TestMarcaRevisionAxis fija el reparto de las clases entre los dos ejes
// (P9-19): `possible_duplicate` es del eje de duplicados y las otras dos del de
// disponibilidad; la traducción nunca inventa una marca de disponibilidad.
func TestMarcaRevisionAxis(t *testing.T) {
	tests := []struct {
		mark       MarcaRevision
		enEje      bool
		disponible MarcaDisponibilidad
		ok         bool
	}{
		{MarcaPosibleDuplicado, false, "", false},
		{MarcaNoDisponible, true, MarcaDisponibilidadNoDisponible, true},
		{MarcaFueraDelFeed, true, MarcaDisponibilidadFueraDelFeed, true},
	}
	for _, tt := range tests {
		if got := tt.mark.EnEjeDeDisponibilidad(); got != tt.enEje {
			t.Errorf("%q.EnEjeDeDisponibilidad() = %v, se esperaba %v", tt.mark, got, tt.enEje)
		}
		got, ok := tt.mark.MarcaDisponibilidad()
		if ok != tt.ok || got != tt.disponible {
			t.Errorf("%q.MarcaDisponibilidad() = (%q, %v), se esperaba (%q, %v)", tt.mark, got, ok, tt.disponible, tt.ok)
		}
	}
}

// TestMapEpisodioMapsNullableColumns traduce la fila de una escritura: las
// columnas anulables pasan a punteros, los enums a su tipo y las fechas a
// time.Time (§8.1.6).
func TestMapEpisodioMapsNullableColumns(t *testing.T) {
	id := uuid.New()
	par := uuid.New()
	publicado := time.Date(2026, 7, 2, 6, 0, 0, 0, time.UTC)
	creado := time.Date(2026, 10, 14, 9, 30, 0, 0, time.UTC)
	resuelto := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

	got := mapEpisodio(gendb.MediaItem{
		ID:                         pgUUID(id),
		Title:                      "Potencial sin Resultados",
		PublishedAt:                stamp(publicado),
		DescriptionEs:              text("Una descripción corta"),
		DescriptionEn:              pgtype.Text{},
		DescriptionEdited:          true,
		CoverFile:                  text("img_abc.jpg"),
		Origin:                     "both",
		Visibility:                 "hidden",
		DuplicateOfID:              pgUUID(par),
		DuplicateMarkResolvedAt:    stamp(resuelto),
		AvailabilityCheckedAt:      pgtype.Timestamptz{},
		AvailabilityResult:         text("inconclusive"),
		AvailabilityMark:           text("missing_from_feed"),
		AvailabilityMarkResolvedAt: pgtype.Timestamptz{},
		MergedIntoID:               pgtype.UUID{},
		CreatedAt:                  stamp(creado),
		UpdatedAt:                  stamp(creado),
	})

	if got.ID != id || got.Titulo != "Potencial sin Resultados" || !got.PublicadoEn.Equal(publicado) {
		t.Errorf("campos no traducidos: %+v", got)
	}
	if got.DescripcionEs == nil || *got.DescripcionEs != "Una descripción corta" {
		t.Errorf("DescripcionEs = %v", got.DescripcionEs)
	}
	if got.DescripcionEn != nil {
		t.Errorf("DescripcionEn = %v, se esperaba nil (NULL)", got.DescripcionEn)
	}
	if !got.DescripcionEditada {
		t.Error("DescripcionEditada = false, se esperaba true")
	}
	if got.Caratula == nil || *got.Caratula != "img_abc.jpg" {
		t.Errorf("Caratula = %v", got.Caratula)
	}
	if got.Origen != OrigenAmbos || got.Visibilidad != VisibilidadOculta {
		t.Errorf("origen/visibilidad = %q/%q", got.Origen, got.Visibilidad)
	}
	if got.DuplicadoDeID == nil || *got.DuplicadoDeID != par {
		t.Errorf("DuplicadoDeID = %v", got.DuplicadoDeID)
	}
	if got.MarcaDuplicadoResueltaEn == nil || !got.MarcaDuplicadoResueltaEn.Equal(resuelto) {
		t.Errorf("MarcaDuplicadoResueltaEn = %v", got.MarcaDuplicadoResueltaEn)
	}
	if got.DisponibilidadComprobadaEn != nil {
		t.Errorf("DisponibilidadComprobadaEn = %v, se esperaba nil", got.DisponibilidadComprobadaEn)
	}
	if got.ResultadoDisponibilidad == nil || *got.ResultadoDisponibilidad != DisponibilidadInconclusa {
		t.Errorf("ResultadoDisponibilidad = %v", got.ResultadoDisponibilidad)
	}
	if got.MarcaDisponibilidad == nil || *got.MarcaDisponibilidad != MarcaDisponibilidadFueraDelFeed {
		t.Errorf("MarcaDisponibilidad = %v", got.MarcaDisponibilidad)
	}
	if got.MarcaDisponibilidadResueltaEn != nil {
		t.Errorf("MarcaDisponibilidadResueltaEn = %v, se esperaba nil", got.MarcaDisponibilidadResueltaEn)
	}
	if got.UnidoAID != nil {
		t.Errorf("UnidoAID = %v, se esperaba nil", got.UnidoAID)
	}
	if got.VideoURL != "" || got.AudioURL != "" {
		t.Errorf("enlaces = %q/%q, se esperaban vacíos (la escritura no los trae)", got.VideoURL, got.AudioURL)
	}
	if !got.CreadoEn.Equal(creado) || !got.ActualizadoEn.Equal(creado) {
		t.Errorf("fechas = %v / %v", got.CreadoEn, got.ActualizadoEn)
	}
}

// TestMapEpisodioListadoConEnlaces traduce la fila del listado del panel: los
// dos enlaces del `LEFT JOIN` por proveedor y el puntero de unión.
func TestMapEpisodioListadoConEnlaces(t *testing.T) {
	id := uuid.New()
	absorbente := uuid.New()
	creado := time.Now()

	got := mapEpisodioListado(gendb.ListMediaItemsRow{
		ID:            pgUUID(id),
		Title:         "La lección de la higuera",
		PublishedAt:   stamp(creado),
		DescriptionEs: pgtype.Text{},
		DescriptionEn: pgtype.Text{},
		CoverFile:     pgtype.Text{},
		Origin:        "both",
		Visibility:    "visible",
		MergedIntoID:  pgUUID(absorbente),
		VideoUrl:      text("https://www.youtube.com/watch?v=abcdefghijk"),
		AudioUrl:      text("https://podcasters.spotify.com/pod/show/x/episodes/y"),
		CreatedAt:     stamp(creado),
		UpdatedAt:     stamp(creado),
	})

	if got.VideoURL != "https://www.youtube.com/watch?v=abcdefghijk" {
		t.Errorf("VideoURL = %q", got.VideoURL)
	}
	if got.AudioURL != "https://podcasters.spotify.com/pod/show/x/episodes/y" {
		t.Errorf("AudioURL = %q", got.AudioURL)
	}
	if got.UnidoAID == nil || *got.UnidoAID != absorbente {
		t.Errorf("UnidoAID = %v", got.UnidoAID)
	}
	if got.Caratula != nil {
		t.Errorf("Caratula = %v, se esperaba nil (NULL → marcador neutro)", got.Caratula)
	}
	if got.Visibilidad != VisibilidadVisible || got.Origen != OrigenAmbos {
		t.Errorf("visibilidad/origen = %q/%q", got.Visibilidad, got.Origen)
	}
}

// TestMapEpisodioDetalle traduce la fila del detalle del panel: mismo contenido
// que el listado (sqlc genera un tipo de fila por consulta).
func TestMapEpisodioDetalle(t *testing.T) {
	id := uuid.New()
	creado := time.Now()

	got := mapEpisodioDetalle(gendb.GetMediaItemByIDRow{
		ID:                         pgUUID(id),
		Title:                      "Sermón",
		PublishedAt:                stamp(creado),
		DescriptionEs:              text("Descripción"),
		DescriptionEn:              pgtype.Text{},
		CoverFile:                  text("img_x.webp"),
		Origin:                     "video",
		Visibility:                 "visible",
		AvailabilityCheckedAt:      stamp(creado),
		AvailabilityResult:         text("available"),
		AvailabilityMark:           pgtype.Text{},
		AvailabilityMarkResolvedAt: pgtype.Timestamptz{},
		VideoUrl:                   text("https://www.youtube.com/watch?v=abcdefghijk"),
		AudioUrl:                   pgtype.Text{},
		CreatedAt:                  stamp(creado),
		UpdatedAt:                  stamp(creado),
	})

	if got.ID != id || got.Origen != OrigenVideo {
		t.Errorf("fila del detalle mal traducida: %+v", got)
	}
	if got.ResultadoDisponibilidad == nil || *got.ResultadoDisponibilidad != DisponibilidadDisponible {
		t.Errorf("ResultadoDisponibilidad = %v", got.ResultadoDisponibilidad)
	}
	if got.MarcaDisponibilidad != nil {
		t.Errorf("MarcaDisponibilidad = %v, se esperaba nil", got.MarcaDisponibilidad)
	}
	if got.VideoURL == "" || got.AudioURL != "" {
		t.Errorf("enlaces = %q/%q, se esperaba solo video", got.VideoURL, got.AudioURL)
	}
}

// TestMapEpisodioPublicoFixesVisibility fija que la fila pública —que no
// selecciona `visibility`— se completa con lo que la consulta garantiza: visible,
// sin absorción y sin par abierto.
func TestMapEpisodioPublicoFixesVisibility(t *testing.T) {
	creado := time.Now()

	got := mapEpisodioPublico(gendb.ListPublishedMediaItemsRow{
		ID:          pgUUID(uuid.New()),
		Title:       "Potencial sin Resultados",
		PublishedAt: stamp(creado),
		Origin:      "audio",
		AudioUrl:    text("https://podcasters.spotify.com/pod/show/x/episodes/y"),
		CreatedAt:   stamp(creado),
		UpdatedAt:   stamp(creado),
	})

	if got.Visibilidad != VisibilidadVisible {
		t.Errorf("Visibilidad = %q, se esperaba visible", got.Visibilidad)
	}
	if got.Origen != OrigenAudio || got.VideoURL != "" {
		t.Errorf("origen/enlaces = %q/%q/%q", got.Origen, got.VideoURL, got.AudioURL)
	}
	if got.UnidoAID != nil || got.DuplicadoDeID != nil || got.MarcaDisponibilidad != nil {
		t.Errorf("la ficha pública no debe traer marcas ni unión: %+v", got)
	}
}

// TestMapEpisodioPortada traduce la fila del bloque de la portada (mismo
// predicado y mismas columnas que la sección).
func TestMapEpisodioPortada(t *testing.T) {
	creado := time.Now()

	got := mapEpisodioPortada(gendb.ListPortadaMediaItemsRow{
		ID:          pgUUID(uuid.New()),
		Title:       "Sermón del domingo",
		PublishedAt: stamp(creado),
		CoverFile:   text("img_p.png"),
		Origin:      "both",
		VideoUrl:    text("https://www.youtube.com/watch?v=abcdefghijk"),
		AudioUrl:    text("https://podcasters.spotify.com/pod/show/x/episodes/y"),
		CreatedAt:   stamp(creado),
		UpdatedAt:   stamp(creado),
	})

	if got.Visibilidad != VisibilidadVisible || got.Origen != OrigenAmbos {
		t.Errorf("fila de portada mal traducida: %+v", got)
	}
	if got.Caratula == nil || *got.Caratula != "img_p.png" {
		t.Errorf("Caratula = %v", got.Caratula)
	}
}

// TestMapFuenteEpisodio traduce la fila de una fuente de importación.
func TestMapFuenteEpisodio(t *testing.T) {
	id := uuid.New()
	episodio := uuid.New()
	creado := time.Now()

	got := mapFuenteEpisodio(gendb.MediaEpisodeSource{
		ID:         pgUUID(id),
		ItemID:     pgUUID(episodio),
		Provider:   "youtube",
		ExternalID: "abcdefghijk",
		FeedTitle:  "Potencial sin Resultados: La Lección de la Higuera",
		Url:        "https://www.youtube.com/watch?v=abcdefghijk",
		CreatedAt:  stamp(creado),
		UpdatedAt:  stamp(creado),
	})

	if got.ID != id || got.EpisodioID != episodio {
		t.Errorf("identificadores = %v / %v", got.ID, got.EpisodioID)
	}
	if got.Proveedor != ProveedorYouTube || got.IDExterno != "abcdefghijk" {
		t.Errorf("proveedor/identificador = %q/%q", got.Proveedor, got.IDExterno)
	}
	if got.TituloFeed == "" || got.Enlace != "https://www.youtube.com/watch?v=abcdefghijk" {
		t.Errorf("procedencia mal traducida: %+v", got)
	}
}

// TestMapSyncRun traduce la fila del registro de una búsqueda.
func TestMapSyncRun(t *testing.T) {
	id := uuid.New()
	actor := uuid.New()
	inicio := time.Now().Add(-time.Minute)
	fin := time.Now()

	got := mapSyncRun(gendb.MediaSyncRun{
		ID:                            pgUUID(id),
		ActorUserID:                   pgUUID(actor),
		StartedAt:                     stamp(inicio),
		FinishedAt:                    stamp(fin),
		ImportedCount:                 3,
		JoinedCount:                   1,
		MarkedCount:                   2,
		HiddenCount:                   1,
		AvailabilityCheckedCount:      4,
		AvailabilityInconclusiveCount: 1,
		YoutubeStatus:                 "ok",
		PodcastStatus:                 "failed",
		CreatedAt:                     stamp(fin),
		UpdatedAt:                     stamp(fin),
	})

	if got.ID != id || got.ActorUserID == nil || *got.ActorUserID != actor {
		t.Errorf("identificadores mal traducidos: %+v", got)
	}
	if got.Importados != 3 || got.Unidos != 1 || got.Marcados != 2 || got.Ocultos != 1 {
		t.Errorf("conteos = %+v", got)
	}
	if got.Comprobados != 4 || got.Inconclusos != 1 {
		t.Errorf("conteos de disponibilidad = %+v", got)
	}
	if got.EstadoYouTube != FeedOK || got.EstadoPodcast != FeedFallido {
		t.Errorf("estados = %q/%q", got.EstadoYouTube, got.EstadoPodcast)
	}
	if !got.IniciadoEn.Equal(inicio) || !got.TerminadoEn.Equal(fin) {
		t.Errorf("fechas = %v / %v", got.IniciadoEn, got.TerminadoEn)
	}
}

// TestMapSyncRunWithoutActor traduce el actor nulo (NULL): la búsqueda se
// registra sin cuenta cuando no se pudo resolver.
func TestMapSyncRunWithoutActor(t *testing.T) {
	got := mapSyncRun(gendb.MediaSyncRun{
		ID:            pgUUID(uuid.New()),
		ActorUserID:   pgtype.UUID{},
		StartedAt:     stamp(time.Now()),
		FinishedAt:    stamp(time.Now()),
		YoutubeStatus: "ok",
		PodcastStatus: "ok",
	})
	if got.ActorUserID != nil {
		t.Errorf("ActorUserID = %v, se esperaba nil", got.ActorUserID)
	}
}

// TestAuditParamsMapsMediaActions verifica el mapeo audit.Action →
// InsertAdminActionParams (P9-14): un objetivo `content` no lleva FK y conserva
// la etiqueta; los 6 códigos `media.*` son válidos y no hay create/delete.
func TestAuditParamsMapsMediaActions(t *testing.T) {
	actor := uuid.New()
	tests := []struct {
		name string
		code string
	}{
		{name: "búsqueda", code: audit.ActionMediaSync},
		{name: "descripción", code: audit.ActionMediaDescriptionUpdate},
		{name: "ocultar", code: audit.ActionMediaHide},
		{name: "mostrar", code: audit.ActionMediaShow},
		{name: "unión", code: audit.ActionMediaMerge},
		{name: "resolución de marca", code: audit.ActionMediaMarkResolve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := audit.Action{
				ActorUserID: &actor,
				Code:        tt.code,
				TargetKind:  audit.TargetContent,
				TargetLabel: "Medios · Potencial sin Resultados",
				Result:      audit.ResultSuccess,
			}
			if err := action.Validate(); err != nil {
				t.Fatalf("la acción %q no valida: %v", tt.code, err)
			}

			params := auditParams(action)
			if !params.ActorUserID.Valid || uuid.UUID(params.ActorUserID.Bytes) != actor {
				t.Errorf("ActorUserID = %+v", params.ActorUserID)
			}
			if params.Action != tt.code {
				t.Errorf("Action = %q, se esperaba %q", params.Action, tt.code)
			}
			if params.TargetKind != "content" {
				t.Errorf("TargetKind = %q, se esperaba content", params.TargetKind)
			}
			if params.TargetUserID.Valid || params.TargetRoleID.Valid {
				t.Errorf("las FK de contenido deben ir NULL: %+v / %+v", params.TargetUserID, params.TargetRoleID)
			}
			if !params.TargetLabel.Valid || params.TargetLabel.String != "Medios · Potencial sin Resultados" {
				t.Errorf("TargetLabel = %+v", params.TargetLabel)
			}
			if params.Result != "success" {
				t.Errorf("Result = %q", params.Result)
			}
		})
	}

	empty := auditParams(audit.Action{TargetKind: audit.TargetContent})
	if empty.TargetLabel.Valid {
		t.Errorf("TargetLabel vacío = %+v, se esperaba NULL", empty.TargetLabel)
	}
}

// TestMediaActionCodesAreTheSixOfTheContract fija que el registro cerrado de F9
// son exactamente los 6 códigos del contrato (sin create ni delete: F9 no tiene
// alta manual ni borrado, FR-019).
func TestMediaActionCodesAreTheSixOfTheContract(t *testing.T) {
	codes := make([]string, 0, 6)
	for _, code := range audit.ActionCodes {
		if strings.HasPrefix(code, "media.") {
			codes = append(codes, code)
		}
	}
	want := []string{
		audit.ActionMediaSync,
		audit.ActionMediaDescriptionUpdate,
		audit.ActionMediaHide,
		audit.ActionMediaShow,
		audit.ActionMediaMerge,
		audit.ActionMediaMarkResolve,
	}
	if len(codes) != len(want) {
		t.Fatalf("el registro declara %d códigos media.*, se esperaban %d (%v)", len(codes), len(want), codes)
	}
	for i, code := range want {
		if codes[i] != code {
			t.Errorf("código %d = %q, se esperaba %q", i, codes[i], code)
		}
	}
}

// TestInsertActionsValidatesBeforeTouchingTheDatabase fija el fail-closed del
// registro de auditoría (FR-022): un código fuera del registro cerrado se rechaza
// ANTES de ejecutar SQL —el `tx` es nil a propósito— y sin acciones no hay nada
// que hacer.
func TestInsertActionsValidatesBeforeTouchingTheDatabase(t *testing.T) {
	err := insertActions(context.Background(), nil, []audit.Action{{
		Code:        "media.archive", // fuera del registro cerrado (P9-14)
		TargetKind:  audit.TargetContent,
		TargetLabel: "Medios · X",
		Result:      audit.ResultSuccess,
	}})
	if !errors.Is(err, audit.ErrUnknownActionCode) {
		t.Fatalf("insertActions(código desconocido) = %v, se esperaba %v", err, audit.ErrUnknownActionCode)
	}

	if err := insertActions(context.Background(), nil, nil); err != nil {
		t.Errorf("insertActions(sin acciones) = %v, se esperaba nil", err)
	}
}

// TestWithTxFailsWhenTransactionCannotStart comprueba el contrato de la base de
// escritura: si la transacción no se puede abrir, `withTx` devuelve el error y NO
// ejecuta la mutación. Se usa un pool perezoso apuntando a un puerto cerrado
// (ninguna conexión posible) para no depender de una base real.
func TestWithTxFailsWhenTransactionCannotStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, "postgres://app:app@127.0.0.1:1/app?sslmode=disable")
	if err != nil {
		t.Fatalf("crear pool: %v", err)
	}
	defer pool.Close()

	called := false
	err = NewRepository(pool).withTx(ctx, func(*repository) error {
		called = true
		return nil
	})
	if err == nil {
		t.Fatal("withTx debería fallar si no puede abrir la transacción")
	}
	if called {
		t.Error("la mutación no debe ejecutarse sin transacción")
	}
}

// TestClassifyTranslatesPostgresErrors fija la traducción de errores de
// PostgreSQL a errores de dominio: sin fila → 404; duplicado (23505) → 409;
// CHECK (23514) → 400 con el detalle por campo; y cualquier otro error se
// devuelve tal cual (nunca se disfraza).
func TestClassifyTranslatesPostgresErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want apperr.Kind
	}{
		{name: "sin fila", err: pgx.ErrNoRows, want: apperr.KindNotFound},
		{name: "duplicado", err: &pgconn.PgError{Code: codeUniqueViolation}, want: apperr.KindConflict},
		{
			name: "check de visibilidad",
			err:  &pgconn.PgError{Code: codeCheckViolation, ConstraintName: "media_items_visibility_check"},
			want: apperr.KindInvalid,
		},
		{name: "otro error de PostgreSQL", err: &pgconn.PgError{Code: "57014"}, want: apperr.KindInternal},
		{name: "error cualquiera", err: errors.New("boom"), want: apperr.KindInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.err, messageEpisodioNoExiste, messageFuenteDuplicada)
			if kind := errorKind(got); kind != tt.want {
				t.Fatalf("classify = %v (%v), se esperaba %v", got, kind, tt.want)
			}
			if tt.want == apperr.KindInternal && !errors.Is(got, tt.err) {
				t.Fatalf("classify perdió el error original: %v", got)
			}
		})
	}

	if err := classify(nil, "", ""); err != nil {
		t.Errorf("classify(nil) = %v, se esperaba nil", err)
	}
}

// TestCheckDetailsMapsInvariants fija el detalle por campo de las violaciones de
// CHECK de las tablas de F9: la BD es la última línea de defensa y su 400 debe
// decir qué campo lo rompió. Los nombres son los que PostgreSQL asigna a las
// restricciones sin nombre de la migración `000014` (verificados en T914).
func TestCheckDetailsMapsInvariants(t *testing.T) {
	tests := []struct {
		constraint string
		field      string
	}{
		{constraint: "media_items_check", field: "descriptionEs"},
		{constraint: "media_items_check1", field: "duplicateOfId"},
		{constraint: "media_items_check2", field: "availabilityMark"},
		{constraint: "media_items_origin_check", field: "origin"},
		{constraint: "media_items_visibility_check", field: "visibility"},
		{constraint: "media_items_availability_mark_check", field: "availabilityMark"},
		{constraint: "media_items_availability_result_check", field: "availabilityResult"},
		{constraint: "media_items_title_check", field: "title"},
		{constraint: "media_items_title_check1", field: "title"},
		{constraint: "media_items_description_es_check", field: "descriptionEs"},
		{constraint: "media_items_description_es_check1", field: "descriptionEs"},
		{constraint: "media_items_description_en_check", field: "descriptionEn"},
		{constraint: "media_items_description_en_check1", field: "descriptionEn"},
		{constraint: "media_items_cover_file_check", field: "coverFile"},
		{constraint: "media_episode_sources_check", field: "url"},
		{constraint: "media_episode_sources_provider_check", field: "provider"},
		{constraint: "media_episode_sources_external_id_check", field: "externalId"},
		{constraint: "media_episode_sources_external_id_check1", field: "externalId"},
		{constraint: "media_episode_sources_feed_title_check", field: "feedTitle"},
		{constraint: "media_episode_sources_feed_title_check1", field: "feedTitle"},
		{constraint: "media_episode_sources_url_check", field: "url"},
		{constraint: "media_episode_sources_url_check1", field: "url"},
		{constraint: "media_sync_runs_check", field: "finishedAt"},
		{constraint: "media_sync_runs_imported_count_check", field: "importedCount"},
		{constraint: "media_sync_runs_joined_count_check", field: "joinedCount"},
		{constraint: "media_sync_runs_marked_count_check", field: "markedCount"},
		{constraint: "media_sync_runs_hidden_count_check", field: "hiddenCount"},
		{constraint: "media_sync_runs_availability_checked_count_check", field: "availabilityCheckedCount"},
		{constraint: "media_sync_runs_availability_inconclusive_count_check", field: "availabilityInconclusiveCount"},
		{constraint: "media_sync_runs_youtube_status_check", field: "youtube"},
		{constraint: "media_sync_runs_podcast_status_check", field: "podcast"},
	}

	for _, tt := range tests {
		details := checkDetails(tt.constraint)
		if details == nil {
			t.Errorf("checkDetails(%q) = nil, se esperaba el detalle del campo", tt.constraint)
			continue
		}
		if _, ok := details[tt.field]; !ok {
			t.Errorf("checkDetails(%q) = %v, se esperaba la clave %q", tt.constraint, details, tt.field)
		}
	}

	if got := checkDetails("donation_methods_check"); got != nil {
		t.Errorf("checkDetails(desconocido) = %v, se esperaba nil", got)
	}
}

// TestWrapKeepsOperationContext comprueba que wrap envuelve el error traducido
// con el nombre de la operación sin romper errors.Is/As.
func TestWrapKeepsOperationContext(t *testing.T) {
	err := wrap(&pgconn.PgError{Code: codeUniqueViolation}, "insert media episode source", "", messageFuenteDuplicada)
	if err == nil {
		t.Fatal("wrap devolvió nil")
	}
	if got := err.Error(); !strings.HasPrefix(got, "insert media episode source: ") {
		t.Errorf("wrap = %q, se esperaba el contexto de la operación", got)
	}
	if kind := errorKind(err); kind != apperr.KindConflict {
		t.Errorf("kind = %v, se esperaba conflict", kind)
	}
	if err := wrap(nil, "insert media item", "", ""); err != nil {
		t.Errorf("wrap(nil) = %v, se esperaba nil", err)
	}

	wrapped := wrap(
		&pgconn.PgError{Code: codeCheckViolation, ConstraintName: "media_episode_sources_check"},
		"insert media episode source", "", "",
	)
	if details := errorDetails(wrapped); details == nil || details["url"] == nil {
		t.Errorf("el detalle por campo se perdió: %v", details)
	}
}
