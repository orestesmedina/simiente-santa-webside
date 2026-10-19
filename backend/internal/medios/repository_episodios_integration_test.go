//go:build integration

package medios

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de las fichas de medios contra PostgreSQL real (T912):
// el predicado público completo —única vía del visitante (FR-009/SC-002/R9-7)—,
// el orden y las tandas, el tope de la portada, los filtros del panel, el
// detalle, las mutaciones de fila completa y la unión por puntero (R9-12).
//
// Requieren DATABASE_URL_TEST con las migraciones aplicadas (lo aporta el CI o
// el compose local; si no existe, se omiten solas). Las tres tablas de F9 son
// exclusivas del módulo: cada prueba parte de tablas vacías y las deja vacías al
// terminar.

// execQuerier es lo mínimo que necesitan los sembrados por SQL directo: sirven
// igual el pool de las pruebas y una transacción.
type execQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// newMediosRepo construye el repositorio contra la base de pruebas y vacía las
// tablas de F9 para partir de un estado conocido (y las deja vacías al
// terminar). Las mutaciones se confirman (la transacción del dominio es real),
// así que se limpia por `DELETE`.
func newMediosRepo(t *testing.T) *repository {
	t.Helper()
	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	wipeMedios(t, pool)
	t.Cleanup(func() { wipeMediosSilencioso(t, pool) })
	return NewRepository(pool)
}

// wipeMedios vacía las tres tablas de F9 en orden de dependencia. Las dos FK
// propias de `media_items` (`duplicate_of_id`/`merged_into_id`) se despejan
// antes: son RESTRICT y apuntan a la propia tabla.
func wipeMedios(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if err := wipe(context.Background(), pool); err != nil {
		t.Fatalf("vaciar las tablas de medios: %v", err)
	}
}

// wipeMediosSilencioso es la limpieza final: un fallo se informa pero no rompe
// la prueba (ya terminó).
func wipeMediosSilencioso(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if err := wipe(context.Background(), pool); err != nil {
		t.Logf("no se pudieron vaciar las tablas de medios al terminar la prueba: %v", err)
	}
}

func wipe(ctx context.Context, pool *pgxpool.Pool) error {
	for _, stmt := range []string{
		"DELETE FROM media_sync_runs",
		"DELETE FROM media_episode_sources",
		"UPDATE media_items SET duplicate_of_id = NULL, duplicate_mark_resolved_at = NULL, merged_into_id = NULL",
		"DELETE FROM media_items",
	} {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("%s: %w", stmt, err)
		}
	}
	return nil
}

// insertRaw inserta en la tabla indicada las columnas del mapa (nombres
// literales del esquema) con valores parametrizados. Las columnas que no
// aparecen toman su DEFAULT (o NULL).
func insertRaw(ctx context.Context, q execQuerier, table string, valores map[string]any) error {
	columnas := make([]string, 0, len(valores))
	for nombre := range valores {
		columnas = append(columnas, nombre)
	}
	sort.Strings(columnas)

	marcadores := make([]string, 0, len(columnas))
	args := make([]any, 0, len(columnas))
	for i, nombre := range columnas {
		marcadores = append(marcadores, fmt.Sprintf("$%d", i+1))
		args = append(args, valores[nombre])
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(columnas, ", "), strings.Join(marcadores, ", "))
	_, err := q.Exec(ctx, query, args...)
	return err
}

// seedFicha siembra una ficha con SQL directo (independiente del código bajo
// prueba) y devuelve su id. `id`, `title`, `published_at` y `origin` tienen
// valores por defecto; el resto son los de la tabla.
func seedFicha(t *testing.T, pool *pgxpool.Pool, valores map[string]any) uuid.UUID {
	t.Helper()
	if _, ok := valores["id"]; !ok {
		valores["id"] = uuid.NewString()
	}
	if _, ok := valores["title"]; !ok {
		valores["title"] = "Sermón de prueba"
	}
	if _, ok := valores["published_at"]; !ok {
		valores["published_at"] = time.Now()
	}
	if _, ok := valores["origin"]; !ok {
		valores["origin"] = "video"
	}
	if err := insertRaw(context.Background(), pool, "media_items", valores); err != nil {
		t.Fatalf("sembrar ficha (%v): %v", valores, err)
	}
	id, err := uuid.Parse(fmt.Sprint(valores["id"]))
	if err != nil {
		t.Fatalf("id sembrado inválido: %v", err)
	}
	return id
}

// seedParDuplicado siembra un par «posible duplicado» recíproco (cada miembro
// apunta al otro) y devuelve sus dos ids. Con `resuelto` los dos miembros llevan
// fecha de resolución («no son el mismo sermón»). La pareja se escribe en dos
// pasos porque la FK impide referenciar una fila que aún no existe.
func seedParDuplicado(t *testing.T, pool *pgxpool.Pool, resuelto bool) (uuid.UUID, uuid.UUID) {
	t.Helper()
	a := uuid.New()
	b := uuid.New()
	seedFicha(t, pool, map[string]any{"id": a.String(), "title": "Par A"})
	valores := map[string]any{
		"id": b.String(), "title": "Par B", "duplicate_of_id": a.String(),
	}
	if resuelto {
		valores["duplicate_mark_resolved_at"] = time.Now()
	}
	seedFicha(t, pool, valores)

	sql := "UPDATE media_items SET duplicate_of_id = $1"
	args := []any{b.String()}
	if resuelto {
		sql += ", duplicate_mark_resolved_at = now()"
	}
	sql += " WHERE id = $2"
	args = append(args, a.String())
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("cerrar la pareja del par: %v", err)
	}
	return a, b
}

// seedFuente siembra una fuente de importación con SQL directo (la consulta
// idempotente del repositorio llega en T913).
func seedFuente(t *testing.T, pool *pgxpool.Pool, itemID uuid.UUID, proveedor, externo, enlace string) {
	t.Helper()
	err := insertRaw(context.Background(), pool, "media_episode_sources", map[string]any{
		"item_id":     itemID.String(),
		"provider":    proveedor,
		"external_id": externo,
		"feed_title":  "Sermón de prueba",
		"url":         enlace,
	})
	if err != nil {
		t.Fatalf("sembrar fuente (%s/%s): %v", proveedor, externo, err)
	}
}

// idsDe devuelve el conjunto de ids del listado, en orden.
func idsDe(items []Episodio) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

// contiene indica si el id está en el listado.
func contiene(ids []uuid.UUID, id uuid.UUID) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}

// publicables pregunta a la ÚNICA vía pública con una tanda amplia.
func publicables(t *testing.T, repo *repository) []uuid.UUID {
	t.Helper()
	items, err := repo.ListPublishedMediaItems(context.Background(), paginate.Params{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("ListPublishedMediaItems: %v", err)
	}
	return idsDe(items)
}

// TestIntegrationListPublishedMediaItemsNeverLeaksUnpublishable fija el
// predicado público completo (FR-009/SC-002/R9-7 + precisión 1 de la puerta 2):
// **nunca** salen las ocultas, las absorbidas ni un par «posible duplicado»
// ABIERTO; **sí** salen las marcadas «ya no aparece en el feed» sin resolver
// (FR-017) y el par **resuelto**.
func TestIntegrationListPublishedMediaItemsNeverLeaksUnpublishable(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool

	visible := seedFicha(t, pool, map[string]any{"title": "Visible suelto"})
	oculta := seedFicha(t, pool, map[string]any{"title": "Oculto por el equipo", "visibility": "hidden"})
	resultante := seedFicha(t, pool, map[string]any{"title": "Resultante de una unión", "origin": "both"})
	absorbida := seedFicha(t, pool, map[string]any{
		"title": "Absorbida por una unión", "visibility": "hidden", "merged_into_id": resultante.String(),
	})

	// Par ABIERTO: cada miembro apunta al otro y ninguno tiene resolución.
	abiertoA, abiertoB := seedParDuplicado(t, pool, false)

	// Par RESUELTO («no son el mismo sermón»): sale al visitante.
	resueltoA, resueltoB := seedParDuplicado(t, pool, true)

	// Marca de disponibilidad sin resolver: la ficha SIGUE visible (FR-017).
	fueraDelFeed := seedFicha(t, pool, map[string]any{
		"title": "Ya no aparece en el feed", "availability_mark": "missing_from_feed",
		"availability_result": "available", "availability_checked_at": time.Now(),
	})

	// Indisponibilidad confirmada: el estado real es oculta (A3).
	noDisponible := seedFicha(t, pool, map[string]any{
		"title": "Enlace confirmado muerto", "visibility": "hidden",
		"availability_mark": "unavailable", "availability_result": "unavailable",
		"availability_checked_at": time.Now(),
	})

	ids := publicables(t, repo)

	for _, id := range []uuid.UUID{visible, resultante, resueltoA, resueltoB, fueraDelFeed} {
		if !contiene(ids, id) {
			t.Errorf("la ficha publicable %v no salió en la sección", id)
		}
	}
	for _, id := range []uuid.UUID{oculta, absorbida, abiertoA, abiertoB, noDisponible} {
		if contiene(ids, id) {
			t.Errorf("la ficha NO publicable %v salió en la sección", id)
		}
	}

	total, err := repo.CountPublishedMediaItems(context.Background())
	if err != nil {
		t.Fatalf("CountPublishedMediaItems: %v", err)
	}
	if int(total) != len(ids) {
		t.Errorf("total = %d y el listado trajo %d", total, len(ids))
	}
}

// TestIntegrationListPublishedMediaItemsOrder fija el orden de la sección
// (FR-006/default #11): `published_at DESC` y, ante la misma fecha, orden de
// importación FIFO (`created_at ASC, id`).
func TestIntegrationListPublishedMediaItemsOrder(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool

	base := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	mismoDia := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)

	antigua := seedFicha(t, pool, map[string]any{
		"title": "La más antigua", "published_at": base, "created_at": base,
	})
	// Tres fichas con la MISMA fecha: el desempate es created_at ASC.
	primera := seedFicha(t, pool, map[string]any{
		"title": "Misma fecha 1", "published_at": mismoDia, "created_at": mismoDia,
	})
	segunda := seedFicha(t, pool, map[string]any{
		"title": "Misma fecha 2", "published_at": mismoDia, "created_at": mismoDia.Add(time.Minute),
	})
	tercera := seedFicha(t, pool, map[string]any{
		"title": "Misma fecha 3", "published_at": mismoDia, "created_at": mismoDia.Add(2 * time.Minute),
	})
	reciente := seedFicha(t, pool, map[string]any{
		"title": "La más reciente", "published_at": base.Add(72 * time.Hour), "created_at": base.Add(72 * time.Hour),
	})

	ids := publicables(t, repo)
	want := []uuid.UUID{reciente, primera, segunda, tercera, antigua}
	if len(ids) != len(want) {
		t.Fatalf("el listado trajo %d fichas, se esperaban %d", len(ids), len(want))
	}
	for i, id := range want {
		if ids[i] != id {
			t.Errorf("posición %d = %v, se esperaba %v", i, ids[i], id)
		}
	}
}

// TestIntegrationListPublishedMediaItemsPagination comprueba las tandas de la
// sección (R9-10): 12 por defecto y el `LIMIT/OFFSET` de «Cargar más» con el
// total del sobre.
func TestIntegrationListPublishedMediaItemsPagination(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	base := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 13; i++ {
		seedFicha(t, pool, map[string]any{
			"title":        fmt.Sprintf("Sermón %02d", i),
			"published_at": base,
			"created_at":   base.Add(time.Duration(i) * time.Minute),
		})
	}

	primeraTanda, err := repo.ListPublishedMediaItems(ctx, paginate.Params{Limit: 12, Offset: 0})
	if err != nil {
		t.Fatalf("ListPublishedMediaItems (tanda 1): %v", err)
	}
	if len(primeraTanda) != 12 {
		t.Fatalf("la primera tanda trajo %d fichas, se esperaban 12", len(primeraTanda))
	}
	if primeraTanda[0].Titulo != "Sermón 00" {
		t.Errorf("la primera tanda empieza en %q, se esperaba «Sermón 00»", primeraTanda[0].Titulo)
	}

	total, err := repo.CountPublishedMediaItems(ctx)
	if err != nil {
		t.Fatalf("CountPublishedMediaItems: %v", err)
	}
	if total != 13 {
		t.Errorf("total = %d, se esperaba 13", total)
	}

	segundaTanda, err := repo.ListPublishedMediaItems(ctx, paginate.Params{Limit: 12, Offset: 12})
	if err != nil {
		t.Fatalf("ListPublishedMediaItems (tanda 2): %v", err)
	}
	if len(segundaTanda) != 1 || segundaTanda[0].Titulo != "Sermón 12" {
		t.Errorf("la segunda tanda = %+v, se esperaba solo «Sermón 12»", idsDe(segundaTanda))
	}
}

// TestIntegrationListPortadaMediaItemsTope fija el tope del bloque de la portada
// (FR-027/SC-015): exactamente los 5 primeros del orden de la sección con 8
// publicables, y exactos cuando hay menos.
func TestIntegrationListPortadaMediaItemsTope(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	base := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	sembradas := make([]uuid.UUID, 0, 8)
	for i := 0; i < 8; i++ {
		sembradas = append(sembradas, seedFicha(t, pool, map[string]any{
			"title":        fmt.Sprintf("Sermón %02d", i),
			"published_at": base,
			"created_at":   base.Add(time.Duration(i) * time.Minute),
		}))
	}
	// Una oculta no cuenta ni ocupa lugar.
	seedFicha(t, pool, map[string]any{"title": "Oculto", "visibility": "hidden"})

	bloque, err := repo.ListPortadaMediaItems(ctx, 5)
	if err != nil {
		t.Fatalf("ListPortadaMediaItems: %v", err)
	}
	if len(bloque) != 5 {
		t.Fatalf("el bloque trajo %d fichas, se esperaban 5", len(bloque))
	}
	for i := 0; i < 5; i++ {
		if bloque[i].ID != sembradas[i] {
			t.Errorf("posición %d del bloque = %v, se esperaba %v (el orden de la sección)", i, bloque[i].ID, sembradas[i])
		}
	}
}

// TestIntegrationListPortadaMediaItemsExactos comprueba que con menos
// publicables el bloque trae exactamente los que hay (sin huecos) y que sin
// ninguno viene vacío.
func TestIntegrationListPortadaMediaItemsExactos(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	bloque, err := repo.ListPortadaMediaItems(ctx, 5)
	if err != nil {
		t.Fatalf("ListPortadaMediaItems (vacío): %v", err)
	}
	if len(bloque) != 0 {
		t.Fatalf("el bloque trajo %d fichas sin publicables, se esperaba vacío", len(bloque))
	}

	for i := 0; i < 3; i++ {
		seedFicha(t, repo.pool, map[string]any{"title": fmt.Sprintf("Sermón %02d", i)})
	}
	bloque, err = repo.ListPortadaMediaItems(ctx, 5)
	if err != nil {
		t.Fatalf("ListPortadaMediaItems: %v", err)
	}
	if len(bloque) != 3 {
		t.Errorf("el bloque trajo %d fichas, se esperaban 3", len(bloque))
	}
}

// TestIntegrationPanelListFilters fija los filtros del listado del panel
// (FR-019): por defecto **sin** absorbidas, con la visibilidad y la marca
// **pendiente** de esa clase como filtros opcionales, y el mismo orden que el
// visitante.
func TestIntegrationPanelListFilters(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	visible := seedFicha(t, pool, map[string]any{"title": "Visible"})
	oculta := seedFicha(t, pool, map[string]any{"title": "Oculta", "visibility": "hidden"})
	resultante := seedFicha(t, pool, map[string]any{"title": "Resultante"})
	absorbida := seedFicha(t, pool, map[string]any{
		"title": "Absorbida", "visibility": "hidden", "merged_into_id": resultante.String(),
	})

	abiertoA, abiertoB := seedParDuplicado(t, pool, false)

	// Par resuelto: no está «pendiente».
	resueltoA, resueltoB := seedParDuplicado(t, pool, true)

	fueraDelFeed := seedFicha(t, pool, map[string]any{
		"title": "Ya no aparece en el feed", "availability_mark": "missing_from_feed",
	})
	resueltaDisponibilidad := seedFicha(t, pool, map[string]any{
		"title": "Marca de disponibilidad resuelta", "availability_mark": "missing_from_feed",
		"availability_mark_resolved_at": time.Now(),
	})

	params := paginate.Params{Limit: 50, Offset: 0}
	listar := func(t *testing.T, filtro FiltroPanel) []uuid.UUID {
		t.Helper()
		items, err := repo.ListMediaItems(ctx, filtro, params)
		if err != nil {
			t.Fatalf("ListMediaItems: %v", err)
		}
		total, err := repo.CountMediaItems(ctx, filtro)
		if err != nil {
			t.Fatalf("CountMediaItems: %v", err)
		}
		if int(total) != len(items) {
			t.Errorf("total = %d y el listado trajo %d", total, len(items))
		}
		return idsDe(items)
	}

	// Por defecto: todo menos las absorbidas.
	porDefecto := listar(t, FiltroPanel{})
	if contiene(porDefecto, absorbida) {
		t.Error("las absorbidas no deben salir por defecto")
	}
	for _, id := range []uuid.UUID{
		visible, oculta, resultante, abiertoA, abiertoB,
		resueltoA, resueltoB, fueraDelFeed, resueltaDisponibilidad,
	} {
		if !contiene(porDefecto, id) {
			t.Errorf("la ficha %v falta en el listado por defecto", id)
		}
	}

	// Revisión avanzada: con las absorbidas.
	conAbsorbidas := listar(t, FiltroPanel{IncluirAbsorbidas: true})
	if !contiene(conAbsorbidas, absorbida) {
		t.Error("con includeMerged la absorbida debe salir")
	}

	// Filtro por visibilidad.
	ocultaFiltro := VisibilidadOculta
	soloOcultas := listar(t, FiltroPanel{Visibilidad: &ocultaFiltro})
	if len(soloOcultas) != 1 || soloOcultas[0] != oculta {
		t.Errorf("el filtro de ocultas = %v, se esperaba solo %v", soloOcultas, oculta)
	}
	visibleFiltro := VisibilidadVisible
	for _, id := range listar(t, FiltroPanel{Visibilidad: &visibleFiltro}) {
		if id == oculta || id == absorbida {
			t.Errorf("el filtro de visibles trajo una ficha no visible: %v", id)
		}
	}

	// Filtro por marca PENDIENTE de cada eje.
	duplicado := MarcaPosibleDuplicado
	pendientesDuplicado := listar(t, FiltroPanel{MarcaRevision: &duplicado})
	if len(pendientesDuplicado) != 2 || !contiene(pendientesDuplicado, abiertoA) || !contiene(pendientesDuplicado, abiertoB) {
		t.Errorf("el filtro de «posible duplicado» = %v, se esperaba el par abierto", pendientesDuplicado)
	}

	fuera := MarcaFueraDelFeed
	pendientesDisponibilidad := listar(t, FiltroPanel{MarcaRevision: &fuera})
	if len(pendientesDisponibilidad) != 1 || pendientesDisponibilidad[0] != fueraDelFeed {
		t.Errorf("el filtro de «ya no aparece en el feed» = %v, se esperaba solo %v",
			pendientesDisponibilidad, fueraDelFeed)
	}

	noDisponible := MarcaNoDisponible
	if got := listar(t, FiltroPanel{MarcaRevision: &noDisponible}); len(got) != 0 {
		t.Errorf("el filtro de «no disponible» = %v, se esperaba vacío", got)
	}
}

// TestIntegrationPanelOrderMatchesPublic fija que el panel ve el MISMO orden que
// el visitante (FR-006/default #11).
func TestIntegrationPanelOrderMatchesPublic(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	base := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		seedFicha(t, pool, map[string]any{
			"title":        fmt.Sprintf("Sermón %02d", i),
			"published_at": base.Add(time.Duration(i) * time.Hour),
		})
	}
	// Una oculta intercalada: no sale en el público pero sí en el panel.
	seedFicha(t, pool, map[string]any{
		"title": "Oculta intercalada", "visibility": "hidden", "published_at": base.Add(90 * time.Minute),
	})

	publico := publicables(t, repo)
	panel, err := repo.ListMediaItems(ctx, FiltroPanel{}, paginate.Params{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("ListMediaItems: %v", err)
	}
	panelIDs := idsDe(panel)
	if len(panelIDs) != 5 {
		t.Fatalf("el panel trajo %d fichas, se esperaban 5", len(panelIDs))
	}

	// Al quitar la oculta, el orden del panel coincide con el público.
	filtrados := make([]uuid.UUID, 0, len(panelIDs))
	for _, id := range panelIDs {
		if contiene(publico, id) {
			filtrados = append(filtrados, id)
		}
	}
	if len(filtrados) != len(publico) {
		t.Fatalf("el panel y el público no coinciden: %d vs %d", len(filtrados), len(publico))
	}
	for i := range publico {
		if filtrados[i] != publico[i] {
			t.Errorf("posición %d: panel %v, público %v", i, filtrados[i], publico[i])
		}
	}
}

// TestIntegrationGetMediaItemByID cubre el detalle del panel: la fila completa
// con las marcas de los dos ejes y los dos enlaces de sus fuentes; un id
// inexistente no es un error (segundo valor false).
func TestIntegrationGetMediaItemByID(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	id := seedFicha(t, pool, map[string]any{
		"title":               "Potencial sin Resultados",
		"origin":              "both",
		"description_es":      "Una descripción corta",
		"description_edited":  true,
		"cover_file":          "img_abc.jpg",
		"availability_mark":   "missing_from_feed",
		"availability_result": "available",
	})
	seedFuente(t, pool, id, "youtube", "abcdefghijk", "https://www.youtube.com/watch?v=abcdefghijk")
	seedFuente(t, pool, id, "podcast", "episodio-1", "https://podcasters.spotify.com/pod/show/x/episodes/y")

	got, found, err := repo.GetMediaItemByID(ctx, id)
	if err != nil {
		t.Fatalf("GetMediaItemByID: %v", err)
	}
	if !found {
		t.Fatal("la ficha existente no se encontró")
	}
	if got.Titulo != "Potencial sin Resultados" || got.Origen != OrigenAmbos {
		t.Errorf("ficha mal traducida: %+v", got)
	}
	if got.DescripcionEs == nil || *got.DescripcionEs != "Una descripción corta" || !got.DescripcionEditada {
		t.Errorf("descripción mal traducida: %+v", got)
	}
	if got.Caratula == nil || *got.Caratula != "img_abc.jpg" {
		t.Errorf("carátula = %v", got.Caratula)
	}
	if got.MarcaDisponibilidad == nil || *got.MarcaDisponibilidad != MarcaDisponibilidadFueraDelFeed {
		t.Errorf("marca de disponibilidad = %v", got.MarcaDisponibilidad)
	}
	if got.ResultadoDisponibilidad == nil || *got.ResultadoDisponibilidad != DisponibilidadDisponible {
		t.Errorf("resultado = %v", got.ResultadoDisponibilidad)
	}
	if got.VideoURL != "https://www.youtube.com/watch?v=abcdefghijk" {
		t.Errorf("VideoURL = %q", got.VideoURL)
	}
	if got.AudioURL != "https://podcasters.spotify.com/pod/show/x/episodes/y" {
		t.Errorf("AudioURL = %q", got.AudioURL)
	}

	if _, found, err := repo.GetMediaItemByID(ctx, uuid.New()); err != nil || found {
		t.Errorf("un id inexistente = (found %v, err %v), se esperaba (false, nil)", found, err)
	}
}

// TestIntegrationInsertMediaItemPersistsEveryColumn comprueba que el alta
// escribe la fila y deja los ejes de marcas y la unión en NULL, con
// `description_edited` en false (la descripción es del feed hasta que el equipo
// la ajuste, FR-015).
func TestIntegrationInsertMediaItemPersistsEveryColumn(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	publicado := time.Date(2026, 7, 2, 6, 0, 0, 0, time.UTC)
	saved, err := repo.InsertMediaItem(ctx, Episodio{
		Titulo:        "Potencial sin Resultados: La Lección de la Higuera",
		PublicadoEn:   publicado,
		DescripcionEs: textOf("Una descripción corta"),
		Caratula:      textOf("img_abc.jpg"),
		Origen:        OrigenVideo,
		Visibilidad:   VisibilidadVisible,
	})
	if err != nil {
		t.Fatalf("InsertMediaItem: %v", err)
	}
	if saved.ID == uuid.Nil {
		t.Fatal("el alta no devolvió id")
	}
	if saved.CreadoEn.IsZero() || saved.ActualizadoEn.IsZero() {
		t.Errorf("el alta no devolvió created_at/updated_at: %v / %v", saved.CreadoEn, saved.ActualizadoEn)
	}
	if saved.Titulo != "Potencial sin Resultados: La Lección de la Higuera" || !saved.PublicadoEn.Equal(publicado) {
		t.Errorf("datos del feed mal guardados: %+v", saved)
	}
	if saved.DescripcionEs == nil || *saved.DescripcionEs != "Una descripción corta" {
		t.Errorf("DescripcionEs = %v", saved.DescripcionEs)
	}
	if saved.DescripcionEn != nil || saved.DescripcionEditada {
		t.Errorf("la traducción opcional debe nacer NULL y la descripción sin editar: %+v", saved)
	}
	if saved.Origen != OrigenVideo || saved.Visibilidad != VisibilidadVisible {
		t.Errorf("origen/visibilidad = %q/%q", saved.Origen, saved.Visibilidad)
	}
	if saved.DuplicadoDeID != nil || saved.MarcaDuplicadoResueltaEn != nil ||
		saved.MarcaDisponibilidad != nil || saved.UnidoAID != nil {
		t.Errorf("los ejes de marcas y la unión deben nacer vacíos: %+v", saved)
	}
}

// TestIntegrationInsertMediaItemRejectsInvalidData comprueba que la BD es la
// última línea de defensa del alta: un título con espacios sobrantes y una
// visibilidad fuera del catálogo se traducen a `invalid` con el campo en el
// detalle.
func TestIntegrationInsertMediaItemRejectsInvalidData(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	base := Episodio{
		Titulo:      "Sermón",
		PublicadoEn: time.Now(),
		Origen:      OrigenVideo,
		Visibilidad: VisibilidadVisible,
	}

	conEspacios := base
	conEspacios.Titulo = " Sermón "
	if _, err := repo.InsertMediaItem(ctx, conEspacios); err == nil {
		t.Error("la BD aceptó un título con espacios sobrantes")
	} else if details := errorDetails(err); details == nil || details["title"] == nil {
		t.Errorf("el detalle por campo se perdió: %v (%v)", details, err)
	}

	fueraDeCatalogo := base
	fueraDeCatalogo.Visibilidad = Visibilidad("draft")
	if _, err := repo.InsertMediaItem(ctx, fueraDeCatalogo); err == nil {
		t.Error("la BD aceptó una visibilidad fuera del catálogo")
	} else if details := errorDetails(err); details == nil || details["visibility"] == nil {
		t.Errorf("el detalle por campo se perdió: %v (%v)", details, err)
	}
}

// TestIntegrationUpdateMediaItemWritesTheWholeRow comprueba que la escritura
// escribe exactamente la fila completa que recibe del service —descripción,
// visibilidad, marcas de los dos ejes, disponibilidad y unión— y que un id
// inexistente es `not_found`.
func TestIntegrationUpdateMediaItemWritesTheWholeRow(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	otra := seedFicha(t, repo.pool, map[string]any{"title": "Otra ficha"})
	publicado := time.Date(2026, 7, 2, 6, 0, 0, 0, time.UTC)
	saved, err := repo.InsertMediaItem(ctx, Episodio{
		Titulo:      "Sermón a editar",
		PublicadoEn: publicado,
		Origen:      OrigenVideo,
		Visibilidad: VisibilidadVisible,
	})
	if err != nil {
		t.Fatalf("InsertMediaItem: %v", err)
	}

	comprobado := time.Now().UTC().Truncate(time.Millisecond)
	resuelto := time.Now().UTC().Truncate(time.Millisecond)
	fusionada := saved
	fusionada.DescripcionEs = textOf("Descripción del equipo")
	fusionada.DescripcionEn = textOf("Team description")
	fusionada.DescripcionEditada = true
	fusionada.Visibilidad = VisibilidadOculta
	fusionada.DuplicadoDeID = &otra
	fusionada.MarcaDuplicadoResueltaEn = &resuelto
	fusionada.DisponibilidadComprobadaEn = &comprobado
	resultado := DisponibilidadInconclusa
	fusionada.ResultadoDisponibilidad = &resultado
	marca := MarcaDisponibilidadNoDisponible
	fusionada.MarcaDisponibilidad = &marca

	got, err := repo.UpdateMediaItem(ctx, fusionada)
	if err != nil {
		t.Fatalf("UpdateMediaItem: %v", err)
	}
	if got.DescripcionEs == nil || *got.DescripcionEs != "Descripción del equipo" || !got.DescripcionEditada {
		t.Errorf("descripción mal escrita: %+v", got)
	}
	if got.DescripcionEn == nil || *got.DescripcionEn != "Team description" {
		t.Errorf("DescripcionEn = %v", got.DescripcionEn)
	}
	if got.Visibilidad != VisibilidadOculta {
		t.Errorf("Visibilidad = %q, se esperaba hidden", got.Visibilidad)
	}
	if got.DuplicadoDeID == nil || *got.DuplicadoDeID != otra {
		t.Errorf("DuplicadoDeID = %v", got.DuplicadoDeID)
	}
	if got.MarcaDuplicadoResueltaEn == nil || !got.MarcaDuplicadoResueltaEn.Equal(resuelto) {
		t.Errorf("MarcaDuplicadoResueltaEn = %v", got.MarcaDuplicadoResueltaEn)
	}
	if got.MarcaDisponibilidad == nil || *got.MarcaDisponibilidad != MarcaDisponibilidadNoDisponible {
		t.Errorf("MarcaDisponibilidad = %v", got.MarcaDisponibilidad)
	}
	if got.ResultadoDisponibilidad == nil || *got.ResultadoDisponibilidad != DisponibilidadInconclusa {
		t.Errorf("ResultadoDisponibilidad = %v", got.ResultadoDisponibilidad)
	}
	// La instantánea no se toca: título y fecha se reescriben tal cual.
	if got.Titulo != "Sermón a editar" || !got.PublicadoEn.Equal(publicado) {
		t.Errorf("la instantánea cambió: %+v", got)
	}

	// Despejar los ejes (NULL) también es una escritura completa.
	limpia := got
	limpia.DuplicadoDeID = nil
	limpia.MarcaDuplicadoResueltaEn = nil
	limpia.MarcaDisponibilidad = nil
	limpia.ResultadoDisponibilidad = nil
	limpia.DisponibilidadComprobadaEn = nil
	limpia.Visibilidad = VisibilidadVisible
	despejada, err := repo.UpdateMediaItem(ctx, limpia)
	if err != nil {
		t.Fatalf("UpdateMediaItem (despejar): %v", err)
	}
	if despejada.DuplicadoDeID != nil || despejada.MarcaDisponibilidad != nil || despejada.MarcaDuplicadoResueltaEn != nil {
		t.Errorf("los ejes no se despejaron: %+v", despejada)
	}

	inexistente := limpia
	inexistente.ID = uuid.New()
	if _, err := repo.UpdateMediaItem(ctx, inexistente); err == nil {
		t.Error("editar una ficha inexistente debería fallar")
	} else if kind := errorKind(err); kind != apperr.KindNotFound {
		t.Errorf("kind = %v, se esperaba not_found", kind)
	}
}

// TestIntegrationRepointMediaItemSourcesMovesSources cubre la unión por puntero
// (R9-12/P9-10): las fuentes de la absorbida pasan a la resultante (que queda
// con los dos enlaces), la absorbida queda con `merged_into_id` y oculta, y
// **ninguna** fuente se borra (cero DELETE).
func TestIntegrationRepointMediaItemSourcesMovesSources(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	video := "https://www.youtube.com/watch?v=abcdefghijk"
	audio := "https://podcasters.spotify.com/pod/show/x/episodes/y"

	resultante := seedFicha(t, pool, map[string]any{"title": "Sermón (video)", "origin": "video"})
	absorbida := seedFicha(t, pool, map[string]any{"title": "Sermón (audio)", "origin": "audio"})
	seedFuente(t, pool, resultante, "youtube", "abcdefghijk", video)
	seedFuente(t, pool, absorbida, "podcast", "episodio-1", audio)

	// El service ya fusionó las dos fichas: la resultante adopta el enlace que
	// faltaba (origin='both') y la absorbida apunta a la resultante.
	fusionada := Episodio{
		ID:          resultante,
		Titulo:      "Sermón (video)",
		PublicadoEn: time.Now(),
		Origen:      OrigenAmbos,
		Visibilidad: VisibilidadVisible,
	}
	absorbidaRow := Episodio{
		ID:          absorbida,
		Titulo:      "Sermón (audio)",
		PublicadoEn: time.Now(),
		Origen:      OrigenAudio,
		Visibilidad: VisibilidadOculta,
		UnidoAID:    &resultante,
	}

	movidas, err := repo.RepointMediaItemSources(ctx, fusionada, absorbidaRow)
	if err != nil {
		t.Fatalf("RepointMediaItemSources: %v", err)
	}
	if len(movidas) != 1 || movidas[0].EpisodioID != resultante || movidas[0].Proveedor != ProveedorPodcast {
		t.Fatalf("fuentes movidas = %+v, se esperaba la del podcast en la resultante", movidas)
	}

	got, found, err := repo.GetMediaItemByID(ctx, resultante)
	if err != nil || !found {
		t.Fatalf("GetMediaItemByID(resultante) = (%v, %v)", found, err)
	}
	if got.VideoURL != video || got.AudioURL != audio {
		t.Errorf("la resultante debe quedar con los dos enlaces: %q / %q", got.VideoURL, got.AudioURL)
	}
	if got.Origen != OrigenAmbos {
		t.Errorf("Origen = %q, se esperaba both", got.Origen)
	}

	absorbidaGot, found, err := repo.GetMediaItemByID(ctx, absorbida)
	if err != nil || !found {
		t.Fatalf("GetMediaItemByID(absorbida) = (%v, %v)", found, err)
	}
	if absorbidaGot.UnidoAID == nil || *absorbidaGot.UnidoAID != resultante {
		t.Errorf("la absorbida no apunta a la resultante: %+v", absorbidaGot.UnidoAID)
	}
	if absorbidaGot.Visibilidad != VisibilidadOculta {
		t.Errorf("Visibilidad de la absorbida = %q, se esperaba hidden", absorbidaGot.Visibilidad)
	}

	// Ninguna fuente se borró y las dos cuelgan de la resultante.
	var enResultante, total int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FILTER (WHERE item_id = $1), count(*) FROM media_episode_sources`,
		resultante.String()).Scan(&enResultante, &total); err != nil {
		t.Fatalf("contar fuentes: %v", err)
	}
	if total != 2 || enResultante != 2 {
		t.Errorf("fuentes = %d (en la resultante %d), se esperaban 2 y 2", total, enResultante)
	}
}

// TestIntegrationInsertMediaItemWithSourceEsAtomico fija H3/SC-003 contra
// PostgreSQL real: el alta de un elemento (ficha + fuente) es UNA transacción, así
// que dos búsquedas simultáneas del mismo elemento dejan **una** ficha con su
// fuente —la que gana la carrera— y **ninguna ficha huérfana** sin enlaces (la
// perdedora se revierte entera).
func TestIntegrationInsertMediaItemWithSourceEsAtomico(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	enlace := "https://www.youtube.com/watch?v=abcdefghijk"
	ficha := Episodio{
		Titulo:      "Sermón simultáneo",
		PublicadoEn: time.Now().UTC().Truncate(time.Millisecond),
		Origen:      OrigenVideo,
		Visibilidad: VisibilidadVisible,
	}
	fuente := FuenteEpisodio{
		Proveedor:  ProveedorYouTube,
		IDExterno:  "abcdefghijk",
		TituloFeed: "Sermón simultáneo",
		Enlace:     enlace,
	}

	var (
		wg         sync.WaitGroup
		resultados [2]bool
		errores    [2]error
	)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, ok, err := repo.InsertMediaItemWithSource(ctx, ficha, fuente)
			resultados[i], errores[i] = ok, err
		}(i)
	}
	wg.Wait()

	for i, err := range errores {
		if err != nil {
			t.Fatalf("alta %d: %v", i, err)
		}
	}
	if resultados[0] == resultados[1] {
		t.Fatalf("exactamente una alta debía ganar la carrera: %v", resultados)
	}

	// Una sola ficha (ninguna huérfana) y con su fuente.
	var fichas, fuentes int
	if err := pool.QueryRow(ctx,
		`SELECT (SELECT count(*) FROM media_items), (SELECT count(*) FROM media_episode_sources)`).
		Scan(&fichas, &fuentes); err != nil {
		t.Fatalf("contar fichas y fuentes: %v", err)
	}
	if fichas != 1 || fuentes != 1 {
		t.Fatalf("fichas = %d, fuentes = %d; se esperaba 1 y 1 (sin ficha huérfana)", fichas, fuentes)
	}

	// Y la ficha que quedó tiene su enlace (no quedó «visible sin enlaces»).
	listado, err := repo.ListMediaItems(ctx, FiltroPanel{}, paginate.Params{Limit: 10})
	if err != nil {
		t.Fatalf("ListMediaItems: %v", err)
	}
	if len(listado) != 1 || listado[0].VideoURL != enlace {
		t.Fatalf("listado = %+v, se esperaba la ficha con su enlace", listado)
	}
}

// TestIntegrationMarkDuplicatePairEsAtomico fija la pareja recíproca de un par
// dudoso (FR-014/M1) en UNA transacción: la adicional nace oculta apuntando al
// candidato y el candidato apunta a la nueva —con su visibilidad y su descripción
// intactas (la marca se escribe por columnas)—. Si la fuente conflictúa no queda
// nada: ni ficha oculta sin enlaces ni pareja a medias.
func TestIntegrationMarkDuplicatePairEsAtomico(t *testing.T) {
	repo := newMediosRepo(t)
	pool := repo.pool
	ctx := context.Background()

	candidato := seedFicha(t, pool, map[string]any{
		"title":              "Candidato",
		"description_es":     "Descripción del equipo",
		"description_edited": true,
		"origin":             "audio",
	})
	seedFuente(t, pool, candidato, "podcast", "guid-01", "https://podcasters.spotify.com/pod/show/x/episodes/1")

	adicional := Episodio{
		Titulo:        "Adicional",
		PublicadoEn:   time.Now().UTC().Truncate(time.Millisecond),
		Origen:        OrigenVideo,
		Visibilidad:   VisibilidadOculta,
		DuplicadoDeID: &candidato,
	}
	creada, fuente, ok, err := repo.MarkDuplicatePair(ctx, adicional, FuenteEpisodio{
		Proveedor:  ProveedorYouTube,
		IDExterno:  "abcdefghijk",
		TituloFeed: "Adicional",
		Enlace:     "https://www.youtube.com/watch?v=abcdefghijk",
	}, candidato)
	if err != nil || !ok {
		t.Fatalf("MarkDuplicatePair = (ok %v, err %v)", ok, err)
	}
	if fuente.EpisodioID != creada.ID {
		t.Errorf("la fuente debe colgar de la ficha nueva: %+v", fuente)
	}

	nueva, found, err := repo.GetMediaItemByID(ctx, creada.ID)
	if err != nil || !found {
		t.Fatalf("GetMediaItemByID(adicional) = (%v, %v)", found, err)
	}
	if nueva.DuplicadoDeID == nil || *nueva.DuplicadoDeID != candidato {
		t.Errorf("la adicional debe apuntar al candidato: %+v", nueva.DuplicadoDeID)
	}
	if nueva.Visibilidad != VisibilidadOculta {
		t.Errorf("la adicional nace oculta: %q", nueva.Visibilidad)
	}
	primera, found, err := repo.GetMediaItemByID(ctx, candidato)
	if err != nil || !found {
		t.Fatalf("GetMediaItemByID(candidato) = (%v, %v)", found, err)
	}
	if primera.DuplicadoDeID == nil || *primera.DuplicadoDeID != creada.ID {
		t.Errorf("el candidato debe apuntar de vuelta (pareja recíproca): %+v", primera.DuplicadoDeID)
	}
	if primera.MarcaDuplicadoResueltaEn != nil {
		t.Errorf("la marca del par nace abierta: %+v", primera.MarcaDuplicadoResueltaEn)
	}
	// La marca se escribe por columnas: ni la visibilidad ni la descripción del
	// equipo se tocan (SC-007).
	if primera.Visibilidad != VisibilidadVisible {
		t.Errorf("la visibilidad del candidato no se toca: %q", primera.Visibilidad)
	}
	if primera.DescripcionEs == nil || *primera.DescripcionEs != "Descripción del equipo" || !primera.DescripcionEditada {
		t.Errorf("la descripción del equipo sigue intacta: %+v", primera)
	}

	// Fuente conflictuada: la transacción se revierte y no queda nada.
	otra := Episodio{
		Titulo: "Otra adicional", PublicadoEn: time.Now().UTC().Truncate(time.Millisecond),
		Origen: OrigenVideo, Visibilidad: VisibilidadOculta, DuplicadoDeID: &candidato,
	}
	_, _, ok, err = repo.MarkDuplicatePair(ctx, otra, FuenteEpisodio{
		Proveedor:  ProveedorYouTube,
		IDExterno:  "abcdefghijk",
		TituloFeed: "Otra adicional",
		Enlace:     "https://www.youtube.com/watch?v=abcdefghijk",
	}, candidato)
	if err != nil {
		t.Fatalf("MarkDuplicatePair (conflicto) = %v", err)
	}
	if ok {
		t.Error("una fuente ya importada no puede dar de alta la ficha")
	}

	// Candidato ya tomado (M1): un elemento DISTINTO que intenta marcar al mismo
	// candidato —otra búsqueda simultánea— tampoco escribe nada: la pareja
	// recíproca no se rompe (la marca del par escribe con `duplicate_of_id IS
	// NULL`).
	tercera := Episodio{
		Titulo: "Tercera adicional", PublicadoEn: time.Now().UTC().Truncate(time.Millisecond),
		Origen: OrigenVideo, Visibilidad: VisibilidadOculta, DuplicadoDeID: &candidato,
	}
	_, _, ok, err = repo.MarkDuplicatePair(ctx, tercera, FuenteEpisodio{
		Proveedor:  ProveedorYouTube,
		IDExterno:  "lmnopqrstuv",
		TituloFeed: "Tercera adicional",
		Enlace:     "https://www.youtube.com/watch?v=lmnopqrstuv",
	}, candidato)
	if err != nil {
		t.Fatalf("MarkDuplicatePair (candidato tomado) = %v", err)
	}
	if ok {
		t.Error("un candidato que ya tiene pareja no puede entrar en otra")
	}
	var fichas int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM media_items`).Scan(&fichas); err != nil {
		t.Fatalf("contar fichas: %v", err)
	}
	if fichas != 2 {
		t.Errorf("fichas = %d, se esperaban 2 (el candidato y su única adicional)", fichas)
	}
	// La pareja sigue recíproca y apuntando a la primera adicional.
	primeraFinal, found, err := repo.GetMediaItemByID(ctx, candidato)
	if err != nil || !found {
		t.Fatalf("GetMediaItemByID(candidato) = (%v, %v)", found, err)
	}
	if primeraFinal.DuplicadoDeID == nil || *primeraFinal.DuplicadoDeID != creada.ID {
		t.Errorf("la pareja del candidato cambió: %+v", primeraFinal.DuplicadoDeID)
	}
}
