//go:build integration

package medios

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de las fuentes de importación contra PostgreSQL real
// (T913): la idempotencia de la búsqueda garantizada **por la base** (R9-11) con
// su instantánea (R9-6), los `CHECK` de forma de enlace por proveedor (R9-8), la
// ventana del feed por proveedor (FR-017) y la procedencia de una ficha.
//
// Requieren DATABASE_URL_TEST con las migraciones aplicadas (lo aporta el CI o
// el compose local; si no existe, se omiten solas).

// testTx abre una transacción de prueba que se revierte al terminar: sirve para
// ejercitar los `CHECK` (que abortan la transacción) sin dejar rastro.
func testTx(t *testing.T, ctx context.Context) (*pgxpool.Pool, pgx.Tx) {
	t.Helper()
	pool := testutil.Pool(t, ctx)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir transacción de prueba: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	return pool, tx
}

// insertInSavepoint ejecuta el INSERT dentro de un savepoint y lo revierte
// siempre: en PostgreSQL un error aborta la transacción entera, así que sin el
// savepoint el primer caso inválido dejaría la transacción inservible para los
// siguientes.
func insertInSavepoint(ctx context.Context, tx pgx.Tx, table string, valores map[string]any) error {
	sp, err := tx.Begin(ctx)
	if err != nil {
		return err
	}
	insertErr := insertRaw(ctx, sp, table, valores)
	if rollbackErr := sp.Rollback(ctx); rollbackErr != nil && insertErr == nil {
		return rollbackErr
	}
	return insertErr
}

// constraintName devuelve el nombre de la restricción que rompió una sentencia.
func constraintName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// sqlState devuelve el código SQLSTATE del error de PostgreSQL ("" si no lo es).
func sqlState(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// fuenteBase arma los valores de una fuente válida: es el punto de partida de
// las pruebas de `CHECK`, que cambian un solo aspecto cada vez.
func fuenteBase(itemID uuid.UUID) map[string]any {
	return map[string]any{
		"item_id":     itemID.String(),
		"provider":    "youtube",
		"external_id": "abcdefghijk",
		"feed_title":  "Potencial sin Resultados",
		"url":         "https://www.youtube.com/watch?v=abcdefghijk",
	}
}

// conCambios devuelve una copia de los valores base con los cambios indicados.
func conCambios(base, cambios map[string]any) map[string]any {
	valores := make(map[string]any, len(base)+len(cambios))
	for k, v := range base {
		valores[k] = v
	}
	for k, v := range cambios {
		valores[k] = v
	}
	return valores
}

// TestIntegrationInsertMediaEpisodeSourceIsIdempotent fija la idempotencia de la
// búsqueda (FR-012/R9-11) y la instantánea (P9-9/R9-6): el segundo INSERT del
// mismo `(provider, external_id)` no falla, no duplica y **no actualiza** nada
// de lo ya importado, aunque el feed haya cambiado el título o el enlace.
func TestIntegrationInsertMediaEpisodeSourceIsIdempotent(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	ficha := seedFicha(t, repo.pool, map[string]any{"title": "Sermón"})
	otraFicha := seedFicha(t, repo.pool, map[string]any{"title": "Otro sermón"})

	original := FuenteEpisodio{
		EpisodioID: ficha,
		Proveedor:  ProveedorYouTube,
		IDExterno:  "abcdefghijk",
		TituloFeed: "Potencial sin Resultados",
		Enlace:     "https://www.youtube.com/watch?v=abcdefghijk",
	}
	guardada, inserted, err := repo.InsertMediaEpisodeSource(ctx, original)
	if err != nil {
		t.Fatalf("InsertMediaEpisodeSource: %v", err)
	}
	if !inserted {
		t.Fatal("el primer INSERT debía importar el elemento")
	}
	if guardada.ID == uuid.Nil || guardada.EpisodioID != ficha {
		t.Errorf("la fuente guardada no es coherente: %+v", guardada)
	}
	if guardada.CreadoEn.IsZero() || guardada.ActualizadoEn.IsZero() {
		t.Errorf("la fuente no devolvió sus fechas: %+v", guardada)
	}

	// La búsqueda repetida: el feed trae otros datos (título nuevo, otro enlace)
	// e incluso otro destino. Nada de eso se aplica.
	repetida := FuenteEpisodio{
		EpisodioID: otraFicha,
		Proveedor:  ProveedorYouTube,
		IDExterno:  "abcdefghijk",
		TituloFeed: "Título nuevo del feed",
		Enlace:     "https://www.youtube.com/watch?v=ZZZZZZZZZZZ",
	}
	segunda, inserted, err := repo.InsertMediaEpisodeSource(ctx, repetida)
	if err != nil {
		t.Fatalf("segundo InsertMediaEpisodeSource: %v", err)
	}
	if inserted {
		t.Error("el segundo INSERT del mismo elemento no debía importar nada")
	}
	if segunda.ID != uuid.Nil {
		t.Errorf("el segundo INSERT devolvió una fila: %+v", segunda)
	}

	leida, found, err := repo.GetMediaEpisodeSource(ctx, ProveedorYouTube, "abcdefghijk")
	if err != nil || !found {
		t.Fatalf("GetMediaEpisodeSource = (%v, %v)", found, err)
	}
	if leida.EpisodioID != ficha || leida.TituloFeed != "Potencial sin Resultados" ||
		leida.Enlace != "https://www.youtube.com/watch?v=abcdefghijk" {
		t.Errorf("la instantánea cambió: %+v", leida)
	}

	var total int
	if err := repo.pool.QueryRow(ctx, `SELECT count(*) FROM media_episode_sources`).Scan(&total); err != nil {
		t.Fatalf("contar fuentes: %v", err)
	}
	if total != 1 {
		t.Errorf("fuentes = %d, se esperaba 1", total)
	}
}

// TestIntegrationInsertMediaEpisodeSourceConcurrent fija la carrera de dos
// búsquedas simultáneas (default revisable #12): exactamente una inserta el
// elemento y la otra no falla ni lo duplica.
func TestIntegrationInsertMediaEpisodeSourceConcurrent(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	ficha := seedFicha(t, repo.pool, map[string]any{"title": "Sermón"})
	fuente := FuenteEpisodio{
		EpisodioID: ficha,
		Proveedor:  ProveedorYouTube,
		IDExterno:  "abcdefghijk",
		TituloFeed: "Potencial sin Resultados",
		Enlace:     "https://www.youtube.com/watch?v=abcdefghijk",
	}

	const intentos = 4
	insertadas := make([]bool, intentos)
	errs := make([]error, intentos)
	var wg sync.WaitGroup
	inicio := make(chan struct{})
	for i := 0; i < intentos; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-inicio
			_, inserted, err := repo.InsertMediaEpisodeSource(ctx, fuente)
			insertadas[i] = inserted
			errs[i] = err
		}(i)
	}
	close(inicio)
	wg.Wait()

	cuantas := 0
	for i := 0; i < intentos; i++ {
		if errs[i] != nil {
			t.Errorf("intento %d falló: %v", i, errs[i])
		}
		if insertadas[i] {
			cuantas++
		}
	}
	if cuantas != 1 {
		t.Errorf("insertadas = %d, se esperaba exactamente 1", cuantas)
	}

	var total int
	if err := repo.pool.QueryRow(ctx, `SELECT count(*) FROM media_episode_sources`).Scan(&total); err != nil {
		t.Fatalf("contar fuentes: %v", err)
	}
	if total != 1 {
		t.Errorf("fuentes = %d, se esperaba 1 (nada se duplica)", total)
	}
}

// TestIntegrationEpisodeSourceChecksFireInTheDatabase comprueba que la BD exige
// la forma esperada del enlace por proveedor (defensa en profundidad de R9-8) y
// que el nombre de la restricción que PostgreSQL reporta se traduce al detalle
// por campo del contrato (T911).
//
// Nota: `media_episode_sources_url_check1` (solo https) y
// `media_episode_sources_provider_check` (catálogo de proveedores) son defensas
// redundantes —la forma por proveedor ya exige `^https://` y un proveedor del
// catálogo—, así que PostgreSQL reporta siempre el `CHECK` de forma
// (`media_episode_sources_check`), como en F6/F7.
func TestIntegrationEpisodeSourceChecksFireInTheDatabase(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	ficha := uuid.New()
	if err := insertRaw(ctx, tx, "media_items", map[string]any{
		"id": ficha.String(), "title": "Sermón de prueba",
		"published_at": time.Now(), "origin": "video",
	}); err != nil {
		t.Fatalf("sembrar la ficha: %v", err)
	}

	base := fuenteBase(ficha)
	tests := []struct {
		name       string
		valores    map[string]any
		constraint string
		field      string
	}{
		{
			name:       "url con javascript",
			valores:    conCambios(base, map[string]any{"url": "javascript:alert(1)"}),
			constraint: "media_episode_sources_check", field: "url",
		},
		{
			name:       "url http",
			valores:    conCambios(base, map[string]any{"url": "http://www.youtube.com/watch?v=abcdefghijk"}),
			constraint: "media_episode_sources_check", field: "url",
		},
		{
			name:       "video de youtube sin 11 caracteres",
			valores:    conCambios(base, map[string]any{"url": "https://www.youtube.com/watch?v=abc"}),
			constraint: "media_episode_sources_check", field: "url",
		},
		{
			name:       "host ajeno",
			valores:    conCambios(base, map[string]any{"url": "https://vimeo.com/123"}),
			constraint: "media_episode_sources_check", field: "url",
		},
		{
			name: "podcast fuera de spotify",
			valores: conCambios(base, map[string]any{
				"provider": "podcast", "url": "https://anchor.fm/s/1116351e8/podcast/rss",
			}),
			constraint: "media_episode_sources_check", field: "url",
		},
		{
			name:       "proveedor fuera del catálogo",
			valores:    conCambios(base, map[string]any{"provider": "vimeo"}),
			constraint: "media_episode_sources_check", field: "url",
		},
		{
			name:       "identificador con espacios sobrantes",
			valores:    conCambios(base, map[string]any{"external_id": " abcdefghijk "}),
			constraint: "media_episode_sources_external_id_check1", field: "externalId",
		},
		{
			name:       "título del feed con espacios sobrantes",
			valores:    conCambios(base, map[string]any{"feed_title": " Potencial sin Resultados "}),
			constraint: "media_episode_sources_feed_title_check1", field: "feedTitle",
		},
		{
			name:       "identificador vacío",
			valores:    conCambios(base, map[string]any{"external_id": ""}),
			constraint: "media_episode_sources_external_id_check", field: "externalId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := insertInSavepoint(ctx, tx, "media_episode_sources", tt.valores)
			if err == nil {
				t.Fatal("la BD aceptó una fuente que viola un invariante")
			}
			if got := sqlState(err); got != codeCheckViolation {
				t.Fatalf("SQLSTATE = %q, se esperaba %q (%v)", got, codeCheckViolation, err)
			}
			if got := constraintName(err); got != tt.constraint {
				t.Fatalf("restricción = %q, se esperaba %q (%v)", got, tt.constraint, err)
			}
			if details := checkDetails(constraintName(err)); details == nil || details[tt.field] == nil {
				t.Fatalf("checkDetails(%q) = %v, se esperaba el campo %q", tt.constraint, details, tt.field)
			}
		})
	}

	// Los valores válidos sí entran (defensa en profundidad, no bloqueo).
	if err := insertRaw(ctx, tx, "media_episode_sources", base); err != nil {
		t.Errorf("la BD rechazó una fuente válida de YouTube: %v", err)
	}
	podcast := conCambios(base, map[string]any{
		"provider": "podcast", "external_id": "episodio-1",
		"url": "https://podcasters.spotify.com/pod/show/x/episodes/y",
	})
	if err := insertRaw(ctx, tx, "media_episode_sources", podcast); err != nil {
		t.Errorf("la BD rechazó una fuente válida del podcast: %v", err)
	}
}

// TestIntegrationUniqueProviderExternalIdRejectsDuplicate fija la última línea
// de defensa de la idempotencia: el `UNIQUE (provider, external_id)` rechaza un
// INSERT directo (sin `ON CONFLICT`) con SQLSTATE 23505, que el repository
// traduce a `conflict`.
func TestIntegrationUniqueProviderExternalIdRejectsDuplicate(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	ficha := uuid.New()
	if err := insertRaw(ctx, tx, "media_items", map[string]any{
		"id": ficha.String(), "title": "Sermón de prueba",
		"published_at": time.Now(), "origin": "video",
	}); err != nil {
		t.Fatalf("sembrar la ficha: %v", err)
	}
	base := fuenteBase(ficha)
	if err := insertRaw(ctx, tx, "media_episode_sources", base); err != nil {
		t.Fatalf("sembrar la fuente: %v", err)
	}

	err := insertInSavepoint(ctx, tx, "media_episode_sources", conCambios(base, map[string]any{
		"feed_title": "Otro título",
	}))
	if err == nil {
		t.Fatal("la BD aceptó dos fuentes con el mismo proveedor e identificador")
	}
	if got := sqlState(err); got != codeUniqueViolation {
		t.Fatalf("SQLSTATE = %q, se esperaba %q (unique_violation) (%v)", got, codeUniqueViolation, err)
	}
	if got := constraintName(err); got != "media_episode_sources_provider_external_id_key" {
		t.Errorf("restricción = %q, se esperaba el UNIQUE por proveedor", got)
	}
	if kind := errorKind(classify(err, "", messageFuenteDuplicada)); kind != apperr.KindConflict {
		t.Errorf("classify = %v, se esperaba conflict", kind)
	}
}

// TestIntegrationSameProviderTwiceIsAServiceRule documenta el límite de la
// defensa de la BD: dos fuentes del MISMO proveedor caben en la misma ficha (la
// tabla solo garantiza la unicidad por `provider`+`external_id`). La regla «una
// ficha = 1 video y/o 1 audio» la mantiene el service en la misma transacción
// (T920) y esta prueba fija que la procedencia completa se puede leer para
// detectarla.
func TestIntegrationSameProviderTwiceIsAServiceRule(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	ficha := seedFicha(t, repo.pool, map[string]any{"title": "Sermón"})
	for _, externo := range []string{"abcdefghijk", "lmnopqrstuv"} {
		if _, _, err := repo.InsertMediaEpisodeSource(ctx, FuenteEpisodio{
			EpisodioID: ficha,
			Proveedor:  ProveedorYouTube,
			IDExterno:  externo,
			TituloFeed: "Sermón",
			Enlace:     "https://www.youtube.com/watch?v=" + externo,
		}); err != nil {
			t.Fatalf("insertar la fuente %q: %v", externo, err)
		}
	}

	fuentes, err := repo.ListMediaItemSourcesByItem(ctx, ficha)
	if err != nil {
		t.Fatalf("ListMediaItemSourcesByItem: %v", err)
	}
	if len(fuentes) != 2 {
		t.Fatalf("la BD guardó %d fuentes del mismo proveedor, se esperaban 2 (la regla es del service)", len(fuentes))
	}
	for _, fuente := range fuentes {
		if fuente.Proveedor != ProveedorYouTube {
			t.Errorf("proveedor = %q, se esperaba youtube", fuente.Proveedor)
		}
	}
}

// TestIntegrationListMediaEpisodeSourcesByProvider cubre la ventana del feed
// (FR-017): la lista por proveedor trae solo las suyas, en orden de importación.
func TestIntegrationListMediaEpisodeSourcesByProvider(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	ficha := seedFicha(t, repo.pool, map[string]any{"title": "Sermón"})
	seedFuente(t, repo.pool, ficha, "youtube", "abcdefghijk", "https://www.youtube.com/watch?v=abcdefghijk")
	seedFuente(t, repo.pool, ficha, "podcast", "episodio-1", "https://podcasters.spotify.com/pod/show/x/episodes/y")

	video, err := repo.ListMediaEpisodeSourcesByProvider(ctx, ProveedorYouTube)
	if err != nil {
		t.Fatalf("ListMediaEpisodeSourcesByProvider(youtube): %v", err)
	}
	if len(video) != 1 || video[0].IDExterno != "abcdefghijk" || video[0].Proveedor != ProveedorYouTube {
		t.Errorf("fuentes de YouTube = %+v", video)
	}

	audio, err := repo.ListMediaEpisodeSourcesByProvider(ctx, ProveedorPodcast)
	if err != nil {
		t.Fatalf("ListMediaEpisodeSourcesByProvider(podcast): %v", err)
	}
	if len(audio) != 1 || audio[0].IDExterno != "episodio-1" {
		t.Errorf("fuentes del podcast = %+v", audio)
	}
}

// TestIntegrationGetMediaEpisodeSource cubre la pregunta «¿ya está importado?»
// (FR-012) y el detalle para las comprobaciones: un elemento que no existe no es
// un error.
func TestIntegrationGetMediaEpisodeSource(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	ficha := seedFicha(t, repo.pool, map[string]any{"title": "Sermón"})
	seedFuente(t, repo.pool, ficha, "podcast", "episodio-1", "https://podcasters.spotify.com/pod/show/x/episodes/y")

	got, found, err := repo.GetMediaEpisodeSource(ctx, ProveedorPodcast, "episodio-1")
	if err != nil || !found {
		t.Fatalf("GetMediaEpisodeSource = (%v, %v)", found, err)
	}
	if got.EpisodioID != ficha || got.Enlace != "https://podcasters.spotify.com/pod/show/x/episodes/y" {
		t.Errorf("fuente mal traducida: %+v", got)
	}

	if _, found, err := repo.GetMediaEpisodeSource(ctx, ProveedorPodcast, "no-existe"); err != nil || found {
		t.Errorf("un elemento inexistente = (%v, %v), se esperaba (false, nil)", found, err)
	}
	// El identificador es único POR proveedor: el mismo valor en otro proveedor
	// no existe.
	if _, found, err := repo.GetMediaEpisodeSource(ctx, ProveedorYouTube, "episodio-1"); err != nil || found {
		t.Errorf("el identificador de otro proveedor = (%v, %v), se esperaba (false, nil)", found, err)
	}
}

// TestIntegrationListMediaItemSourcesByItem cubre la procedencia de una ficha
// (panel y unión): solo las suyas, en orden de importación.
func TestIntegrationListMediaItemSourcesByItem(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	primera := seedFicha(t, repo.pool, map[string]any{"title": "Sermón 1"})
	segunda := seedFicha(t, repo.pool, map[string]any{"title": "Sermón 2"})
	seedFuente(t, repo.pool, primera, "youtube", "abcdefghijk", "https://www.youtube.com/watch?v=abcdefghijk")
	seedFuente(t, repo.pool, primera, "podcast", "episodio-1", "https://podcasters.spotify.com/pod/show/x/episodes/y")
	seedFuente(t, repo.pool, segunda, "youtube", "lmnopqrstuv", "https://www.youtube.com/watch?v=lmnopqrstuv")

	fuentes, err := repo.ListMediaItemSourcesByItem(ctx, primera)
	if err != nil {
		t.Fatalf("ListMediaItemSourcesByItem: %v", err)
	}
	if len(fuentes) != 2 {
		t.Fatalf("fuentes de la primera ficha = %d, se esperaban 2", len(fuentes))
	}
	if fuentes[0].Proveedor != ProveedorYouTube || fuentes[1].Proveedor != ProveedorPodcast {
		t.Errorf("orden de importación roto: %+v", fuentes)
	}
	for _, fuente := range fuentes {
		if fuente.EpisodioID != primera {
			t.Errorf("fuente de otra ficha: %+v", fuente)
		}
	}

	solaUna, err := repo.ListMediaItemSourcesByItem(ctx, segunda)
	if err != nil {
		t.Fatalf("ListMediaItemSourcesByItem (segunda): %v", err)
	}
	if len(solaUna) != 1 || solaUna[0].IDExterno != "lmnopqrstuv" {
		t.Errorf("fuentes de la segunda ficha = %+v", solaUna)
	}
}
