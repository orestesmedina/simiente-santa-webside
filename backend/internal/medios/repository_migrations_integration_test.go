//go:build integration

package medios

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de la capa de datos de F9 contra PostgreSQL real
// (T914): la verificación ejecutable de `data-model.md` §«Validación del
// modelo» en lo que toca al esquema.
//
//   1. las migraciones `000014` y `000015` aplican up → down → up sin errores
//      (incluida la restauración del CHECK de `admin_actions`);
//   2. el catálogo de `permissions` sigue en 10 filas con `medios` entre ellas
//      (F9 no lo toca: el permiso ya lo sembró `000002`);
//   3. los `CHECK` del modelo se disparan en la BD y su nombre coincide con el
//      que traduce el repository al detalle por campo (T911);
//   3bis. la matriz de marcas «par abierto/resuelto × enlace vivo/muerto/
//      inconcluso» (correcciones C1/A3): los seis estados son legales, los dos
//      ejes conviven y el predicado público bloquea **solo** el par abierto;
//   5. `IsFilePublished` con la rama de F9: la carátula de una ficha publicable
//      se sirve y la de una oculta, absorbida o con par abierto no (sin borrar
//      el archivo).
//
// La atomicidad de mutación + auditoría y el registro cerrado de códigos se
// verifican en repository_audit_integration_test.go (bloque 4).
//
// Aislamiento: cada prueba de MIGRACIÓN trabaja dentro de una TRANSACCIÓN que se
// revierte al terminar (el DDL de PostgreSQL es transaccional) y que toma
// primero el lock de `testutil.SerializeMigrations` (decisión 25 del roadmap):
// en CI todas las pruebas comparten una única base y dos `down` destructivos
// solapados se matan entre sí con 40P01.

// execMigrationInTx ejecuta un archivo .sql completo de migración dentro de la
// transacción indicada.
func execMigrationInTx(t *testing.T, ctx context.Context, tx pgx.Tx, name string) {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", name)
	sql, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leer %s: %v", name, err)
	}
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		t.Fatalf("ejecutar %s: %v", name, err)
	}
}

// migrationTx abre la transacción de una prueba de MIGRACIÓN: además de
// revertirse al terminar, toma el lock de serialización de testutil como primera
// sentencia (decisión 25: los ciclos destructivos de migración se serializan).
func migrationTx(t *testing.T, ctx context.Context) (*pgxpool.Pool, pgx.Tx) {
	t.Helper()
	pool, tx := testTx(t, ctx)
	testutil.SerializeMigrations(t, ctx, tx)
	return pool, tx
}

// constraintDefTx devuelve la definición textual de una restricción.
func constraintDefTx(t *testing.T, ctx context.Context, tx pgx.Tx, name, table string) string {
	t.Helper()
	var def string
	if err := tx.QueryRow(ctx, `SELECT pg_get_constraintdef(oid) FROM pg_constraint
		WHERE conname = $1 AND conrelid = $2::regclass`, name, table).Scan(&def); err != nil {
		t.Fatalf("definición de la restricción %s: %v", name, err)
	}
	return def
}

// countTx cuenta filas con una consulta fija.
func countTx(t *testing.T, ctx context.Context, tx pgx.Tx, query string, args ...any) int {
	t.Helper()
	var total int
	if err := tx.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		t.Fatalf("contar (%s): %v", query, err)
	}
	return total
}

// tableExistsTx indica si una tabla existe en la transacción.
func tableExistsTx(t *testing.T, ctx context.Context, tx pgx.Tx, table string) bool {
	t.Helper()
	return countTx(t, ctx, tx,
		`SELECT count(*) FROM pg_class WHERE relname = $1 AND relkind = 'r'`, table) != 0
}

// columnDefaultTx devuelve el DEFAULT de una columna ("" si no tiene).
func columnDefaultTx(t *testing.T, ctx context.Context, tx pgx.Tx, table, column string) string {
	t.Helper()
	var def *string
	if err := tx.QueryRow(ctx, `SELECT column_default FROM information_schema.columns
		WHERE table_name = $1 AND column_name = $2`, table, column).Scan(&def); err != nil {
		t.Fatalf("leer el DEFAULT de %s.%s: %v", table, column, err)
	}
	if def == nil {
		return ""
	}
	return *def
}

// isNullableTx indica si una columna admite NULL.
func isNullableTx(t *testing.T, ctx context.Context, tx pgx.Tx, table, column string) bool {
	t.Helper()
	var nullable string
	if err := tx.QueryRow(ctx, `SELECT is_nullable FROM information_schema.columns
		WHERE table_name = $1 AND column_name = $2`, table, column).Scan(&nullable); err != nil {
		t.Fatalf("leer is_nullable de %s.%s: %v", table, column, err)
	}
	return nullable == "YES"
}

// TestIntegrationMigration000014UpDownUp aplica `000014_create_media_content`
// up → down → up y comprueba que las tres tablas quedan como las describe
// `data-model.md`: columnas obligatorias, defaults, índices y `ON DELETE`
// explícitos.
func TestIntegrationMigration000014UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const (
		up   = "000014_create_media_content.up.sql"
		down = "000014_create_media_content.down.sql"
	)

	tablas := []string{"media_items", "media_episode_sources", "media_sync_runs"}

	// Punto de partida normalizado: sin las tablas (la migración ya está
	// aplicada en la base de pruebas).
	execMigrationInTx(t, ctx, tx, down)
	for _, tabla := range tablas {
		if tableExistsTx(t, ctx, tx, tabla) {
			t.Fatalf("%s sigue existiendo tras el down", tabla)
		}
	}

	execMigrationInTx(t, ctx, tx, up) // 1) up
	assertMediaSchema(t, ctx, tx)

	execMigrationInTx(t, ctx, tx, down) // 2) down
	for _, tabla := range tablas {
		if tableExistsTx(t, ctx, tx, tabla) {
			t.Fatalf("%s sigue existiendo tras el segundo down", tabla)
		}
	}

	execMigrationInTx(t, ctx, tx, up) // 3) up
	assertMediaSchema(t, ctx, tx)
}

// assertMediaSchema comprueba el estado esperado de las tablas de F9.
func assertMediaSchema(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	for _, tabla := range []string{"media_items", "media_episode_sources", "media_sync_runs"} {
		if !tableExistsTx(t, ctx, tx, tabla) {
			t.Fatalf("falta la tabla %s", tabla)
		}
		// Toda tabla del proyecto tiene id, created_at y updated_at (§VI).
		for _, columna := range []string{"id", "created_at", "updated_at"} {
			if isNullableTx(t, ctx, tx, tabla, columna) {
				t.Errorf("%s.%s es anulable, se esperaba NOT NULL", tabla, columna)
			}
		}
	}

	// Obligatorios de una ficha (FR-001/FR-003): título, fecha y origen.
	for _, columna := range []string{"title", "published_at", "origin", "visibility", "description_edited"} {
		if isNullableTx(t, ctx, tx, "media_items", columna) {
			t.Errorf("media_items.%s es anulable, se esperaba NOT NULL", columna)
		}
	}
	// Los dos ejes de marcas y la carátula son opcionales.
	for _, columna := range []string{"description_es", "description_en", "cover_file", "duplicate_of_id",
		"duplicate_mark_resolved_at", "availability_checked_at", "availability_result",
		"availability_mark", "availability_mark_resolved_at", "merged_into_id"} {
		if !isNullableTx(t, ctx, tx, "media_items", columna) {
			t.Errorf("media_items.%s es obligatoria, se esperaba anulable", columna)
		}
	}
	// Una ficha entra visible y con la descripción del feed (FR-015).
	if got := columnDefaultTx(t, ctx, tx, "media_items", "visibility"); !strings.Contains(got, "visible") {
		t.Errorf("visibility DEFAULT = %q, se esperaba 'visible'", got)
	}
	if got := columnDefaultTx(t, ctx, tx, "media_items", "description_edited"); !strings.Contains(got, "false") {
		t.Errorf("description_edited DEFAULT = %q, se esperaba false", got)
	}
	// Sin `sort_order` (el orden es por fecha) ni `deleted_at` (fuera de alcance).
	for _, columna := range []string{"sort_order", "deleted_at"} {
		if countTx(t, ctx, tx, `SELECT count(*) FROM information_schema.columns
			WHERE table_name = 'media_items' AND column_name = $1`, columna) != 0 {
			t.Errorf("media_items.%s no debería existir", columna)
		}
	}

	// Los cinco índices de la ficha (orden, visibilidad y los dos ejes + unión).
	indices := countTx(t, ctx, tx, `SELECT count(*) FROM pg_indexes WHERE tablename = 'media_items'
		AND indexname IN ('media_items_order_idx', 'media_items_visibility_idx',
		                  'media_items_duplicate_of_idx', 'media_items_availability_mark_idx',
		                  'media_items_merged_into_idx')`)
	if indices != 5 {
		t.Errorf("índices de media_items = %d, se esperaban 5", indices)
	}

	// `media_episode_sources`: el `UNIQUE (provider, external_id)` de la
	// idempotencia (R9-11) y los dos índices.
	if uniques := countTx(t, ctx, tx, `SELECT count(*) FROM pg_constraint
		WHERE conrelid = 'media_episode_sources'::regclass AND contype = 'u'`); uniques != 1 {
		t.Errorf("UNIQUE de media_episode_sources = %d, se esperaba 1 (provider, external_id)", uniques)
	}
	indicesFuentes := countTx(t, ctx, tx, `SELECT count(*) FROM pg_indexes
		WHERE tablename = 'media_episode_sources'
		  AND indexname IN ('media_episode_sources_item_idx', 'media_episode_sources_provider_idx')`)
	if indicesFuentes != 2 {
		t.Errorf("índices de media_episode_sources = %d, se esperaban 2", indicesFuentes)
	}

	// `media_sync_runs`: los conteos nacen en 0 y existe el índice de «la última
	// búsqueda» (FR-018).
	for _, columna := range []string{"imported_count", "joined_count", "marked_count", "hidden_count",
		"availability_checked_count", "availability_inconclusive_count"} {
		if got := columnDefaultTx(t, ctx, tx, "media_sync_runs", columna); !strings.HasPrefix(got, "0") {
			t.Errorf("%s DEFAULT = %q, se esperaba 0", columna, got)
		}
	}
	if countTx(t, ctx, tx, `SELECT count(*) FROM pg_indexes WHERE tablename = 'media_sync_runs'
		AND indexname = 'media_sync_runs_started_at_idx'`) != 1 {
		t.Error("falta el índice media_sync_runs_started_at_idx")
	}

	// Las FK llevan `ON DELETE` explícito (convención del bloque anterior): las
	// fichas no se borran nunca (FR-019) y las cuentas de F2 tampoco.
	fks := map[string]string{
		"media_episode_sources_item_id_fkey": "media_episode_sources",
		"media_items_duplicate_of_id_fkey":   "media_items",
		"media_items_merged_into_id_fkey":    "media_items",
		"media_sync_runs_actor_user_id_fkey": "media_sync_runs",
	}
	for nombre, tabla := range fks {
		def := constraintDefTx(t, ctx, tx, nombre, tabla)
		if !strings.Contains(def, "ON DELETE RESTRICT") {
			t.Errorf("%s = %q, se esperaba ON DELETE RESTRICT", nombre, def)
		}
	}
}

// TestIntegrationMigration000015UpDownUp aplica
// `000015_extend_admin_actions_for_media` up → down → up y comprueba que el
// registro cerrado de códigos gana los 6 `media.*` y que el `down` restaura el
// de F7 sin tocar los códigos anteriores.
func TestIntegrationMigration000015UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const (
		up   = "000015_extend_admin_actions_for_media.up.sql"
		down = "000015_extend_admin_actions_for_media.down.sql"
	)

	// Punto de partida normalizado: el CHECK previo a F9.
	execMigrationInTx(t, ctx, tx, down)
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 1) up
	assertActionCheck(t, ctx, tx, true)

	execMigrationInTx(t, ctx, tx, down) // 2) down
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 3) up
	assertActionCheck(t, ctx, tx, true)

	// El tipo de objetivo NO cambia: las fichas son `content`.
	kindDef := constraintDefTx(t, ctx, tx, "admin_actions_target_kind_check", "admin_actions")
	for _, kind := range []string{"'user'", "'role'", "'content'"} {
		if !strings.Contains(kindDef, kind) {
			t.Fatalf("target_kind no admite %s: %s", kind, kindDef)
		}
	}
}

// assertActionCheck comprueba que el CHECK de códigos incluye (o no) los 6
// códigos de F9 y que los de F2/F3/F4/F6/F7 siguen intactos (regresión).
func assertActionCheck(t *testing.T, ctx context.Context, tx pgx.Tx, wantF9 bool) {
	t.Helper()
	def := constraintDefTx(t, ctx, tx, "admin_actions_action_check", "admin_actions")

	codes := mediaActionCodes()
	if len(codes) != 6 {
		t.Fatalf("platform/audit declara %d códigos media.*, se esperaban 6", len(codes))
	}
	for _, code := range codes {
		if got := strings.Contains(def, "'"+code+"'"); got != wantF9 {
			t.Fatalf("el CHECK incluye %q = %v, se esperaba %v", code, got, wantF9)
		}
	}
	if inCheck := strings.Count(def, "'media."); wantF9 && inCheck != 6 {
		t.Fatalf("el CHECK tiene %d códigos media.*, se esperaban 6", inCheck)
	}
	if inCheck := strings.Count(def, "'media."); !wantF9 && inCheck != 0 {
		t.Fatalf("el CHECK de F7 tiene %d códigos media.*, se esperaban 0", inCheck)
	}

	// Los códigos de las funcionalidades anteriores no se pierden.
	for _, code := range []string{"user.create", "home.publish", "event.create", "category.delete",
		"ministry.publish", "donation.text.update"} {
		if !strings.Contains(def, "'"+code+"'") {
			t.Fatalf("el CHECK perdió %q: %s", code, def)
		}
	}
}

// TestIntegrationPermissionsCatalogUnchanged fija que F9 **no** toca el catálogo
// de permisos de F2 (FR-021/P9-15): sigue con 10 permisos, `medios` entre ellos
// con su etiqueta, y los `down` de F9 tampoco lo rozan.
func TestIntegrationPermissionsCatalogUnchanged(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const catalog = `SELECT count(*) FROM permissions`

	if got := countTx(t, ctx, tx, catalog); got != 10 {
		t.Fatalf("permisos = %d, se esperaban 10", got)
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM permissions WHERE code = 'medios'`); got != 1 {
		t.Fatalf("el permiso `medios` no está en el catálogo: %d", got)
	}
	var label string
	if err := tx.QueryRow(ctx, `SELECT label FROM permissions WHERE code = 'medios'`).Scan(&label); err != nil {
		t.Fatalf("leer la etiqueta de `medios`: %v", err)
	}
	if label != "Medios (prédicas y podcasts)" {
		t.Errorf("etiqueta de `medios` = %q", label)
	}

	// Los `down` de F9 no tocan el catálogo.
	execMigrationInTx(t, ctx, tx, "000015_extend_admin_actions_for_media.down.sql")
	execMigrationInTx(t, ctx, tx, "000014_create_media_content.down.sql")
	if got := countTx(t, ctx, tx, catalog); got != 10 {
		t.Fatalf("permisos tras los down de F9 = %d, se esperaban 10", got)
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM permissions WHERE code = 'medios'`); got != 1 {
		t.Fatalf("los down de F9 retiraron el permiso `medios`: %d", got)
	}
}

// TestIntegrationMediaChecksFireInTheDatabase comprueba que cada invariante del
// modelo está protegido por la BD y que el nombre de la restricción que
// PostgreSQL reporta coincide con el que el repository traduce al detalle por
// campo (T911).
func TestIntegrationMediaChecksFireInTheDatabase(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	base := map[string]any{
		"title": "Potencial sin Resultados", "published_at": time.Now(), "origin": "video",
	}

	tests := []struct {
		name       string
		valores    map[string]any
		constraint string
		field      string
	}{
		{
			name:       "descripción en inglés sin español",
			valores:    conCambios(base, map[string]any{"description_en": "Short description"}),
			constraint: "media_items_check", field: "descriptionEs",
		},
		{
			name: "resolución del duplicado sin par",
			valores: conCambios(base, map[string]any{
				"duplicate_mark_resolved_at": time.Now(),
			}),
			constraint: "media_items_check1", field: "duplicateOfId",
		},
		{
			name: "resolución de disponibilidad sin marca",
			valores: conCambios(base, map[string]any{
				"availability_mark_resolved_at": time.Now(),
			}),
			constraint: "media_items_check2", field: "availabilityMark",
		},
		{
			name:       "origen fuera del catálogo",
			valores:    conCambios(base, map[string]any{"origin": "stream"}),
			constraint: "media_items_origin_check", field: "origin",
		},
		{
			name:       "visibilidad fuera del catálogo",
			valores:    conCambios(base, map[string]any{"visibility": "draft"}),
			constraint: "media_items_visibility_check", field: "visibility",
		},
		{
			name:       "marca de disponibilidad fuera del catálogo",
			valores:    conCambios(base, map[string]any{"availability_mark": "review"}),
			constraint: "media_items_availability_mark_check", field: "availabilityMark",
		},
		{
			name:       "resultado de disponibilidad fuera del catálogo",
			valores:    conCambios(base, map[string]any{"availability_result": "maybe"}),
			constraint: "media_items_availability_result_check", field: "availabilityResult",
		},
		{
			name:       "título con espacios sobrantes",
			valores:    conCambios(base, map[string]any{"title": " Potencial sin Resultados "}),
			constraint: "media_items_title_check1", field: "title",
		},
		{
			name: "descripción en español con espacios sobrantes",
			valores: conCambios(base, map[string]any{
				"description_es": " Una descripción ",
			}),
			constraint: "media_items_description_es_check1", field: "descriptionEs",
		},
		{
			name: "descripción en inglés con espacios sobrantes",
			valores: conCambios(base, map[string]any{
				"description_es": "Una descripción", "description_en": " A description ",
			}),
			constraint: "media_items_description_en_check1", field: "descriptionEn",
		},
		{
			name:       "título vacío",
			valores:    conCambios(base, map[string]any{"title": ""}),
			constraint: "media_items_title_check", field: "title",
		},
		{
			name:       "nombre de carátula de más de 120 caracteres",
			valores:    conCambios(base, map[string]any{"cover_file": strings.Repeat("i", 121)}),
			constraint: "media_items_cover_file_check", field: "coverFile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := insertInSavepoint(ctx, tx, "media_items", tt.valores)
			if err == nil {
				t.Fatal("la BD aceptó una ficha que viola un invariante")
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
	if err := insertRaw(ctx, tx, "media_items", base); err != nil {
		t.Errorf("la BD rechazó una ficha válida: %v", err)
	}
}

// TestIntegrationNoSymmetryCheckBetweenMarkAndPair fija la corrección C1: **no**
// hay ningún `CHECK` que ate la marca de disponibilidad al par de duplicados
// (era justo lo que impedía marcar «no disponible» sobre una ficha con par); la
// pareja recíproca del par es invariante de service, y la BD la admite escrita
// en los dos sentidos.
func TestIntegrationNoSymmetryCheckBetweenMarkAndPair(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	rows, err := tx.Query(ctx, `SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint
		WHERE conrelid = 'media_items'::regclass AND contype = 'c'`)
	if err != nil {
		t.Fatalf("leer los CHECK de media_items: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var nombre, def string
		if err := rows.Scan(&nombre, &def); err != nil {
			t.Fatalf("leer un CHECK: %v", err)
		}
		if strings.Contains(def, "availability_mark") && strings.Contains(def, "duplicate_of_id") {
			t.Fatalf("el CHECK %s ata la marca de disponibilidad al par: %s", nombre, def)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("recorrer los CHECK: %v", err)
	}

	// La pareja recíproca (A→B y B→A) se escribe y se conserva: el recíproco es
	// invariante del service (T920) con prueba.
	a := uuid.New()
	b := uuid.New()
	if err := insertRaw(ctx, tx, "media_items", map[string]any{
		"id": a.String(), "title": "Par A", "published_at": time.Now(), "origin": "video",
	}); err != nil {
		t.Fatalf("sembrar A: %v", err)
	}
	if err := insertRaw(ctx, tx, "media_items", map[string]any{
		"id": b.String(), "title": "Par B", "published_at": time.Now(), "origin": "audio",
	}); err != nil {
		t.Fatalf("sembrar B: %v", err)
	}
	cerrarParTx(t, ctx, tx, a, b, nil)

	var deA, deB string
	if err := tx.QueryRow(ctx, `SELECT duplicate_of_id::text FROM media_items WHERE id = $1`, a.String()).Scan(&deA); err != nil {
		t.Fatalf("leer el par de A: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT duplicate_of_id::text FROM media_items WHERE id = $1`, b.String()).Scan(&deB); err != nil {
		t.Fatalf("leer el par de B: %v", err)
	}
	if deA != b.String() || deB != a.String() {
		t.Errorf("la pareja recíproca no se conservó: A→%s, B→%s", deA, deB)
	}

	// Marcar la disponibilidad de A NO rompe su par (ni el de B): los dos ejes
	// conviven.
	if _, err := tx.Exec(ctx, `UPDATE media_items SET availability_mark = 'unavailable',
		availability_result = 'unavailable', availability_checked_at = now() WHERE id = $1`, a.String()); err != nil {
		t.Fatalf("marcar la disponibilidad de A: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT duplicate_of_id::text FROM media_items WHERE id = $1`, a.String()).Scan(&deA); err != nil {
		t.Fatalf("releer el par de A: %v", err)
	}
	if deA != b.String() {
		t.Errorf("marcar la disponibilidad rompió el par de A: %s", deA)
	}
}

// TestIntegrationMarksMatrix fija la matriz «par abierto/resuelto × enlace
// vivo/muerto/inconcluso» (correcciones C1/A3): los seis estados son **legales**
// —los dos ejes conviven— y el **único** bloqueo público por marca es el par
// ABIERTO; una marca de disponibilidad sin resolver **no** bloquea (FR-017), y
// una ficha oculta no sale aunque su par esté resuelto.
func TestIntegrationMarksMatrix(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	// Estados del eje de disponibilidad (FR-017/R9-2): el enlace vivo y el
	// inconcluso dejan la ficha visible; el muerto confirmado la oculta (A3).
	tipos := []struct {
		nombre      string
		resultado   string
		marca       any
		visibilidad string
		visible     bool
	}{
		{nombre: "enlace vivo", resultado: "available", marca: "missing_from_feed", visibilidad: "visible", visible: true},
		{nombre: "enlace muerto", resultado: "unavailable", marca: "unavailable", visibilidad: "hidden", visible: false},
		{nombre: "comprobación inconclusa", resultado: "inconclusive", marca: "missing_from_feed", visibilidad: "visible", visible: true},
	}

	for _, parResuelto := range []bool{false, true} {
		for _, tipo := range tipos {
			nombre := "par abierto"
			if parResuelto {
				nombre = "par resuelto"
			}
			t.Run(nombre+"/"+tipo.nombre, func(t *testing.T) {
				a, b := seedParDuplicado(t, repo.pool, parResuelto)
				// El estado de disponibilidad se escribe sobre el miembro A.
				if _, err := repo.pool.Exec(ctx, `UPDATE media_items
					SET availability_result = $1, visibility = $2, availability_mark = $3
					WHERE id = $4`,
					tipo.resultado, tipo.visibilidad, tipo.marca, a.String()); err != nil {
					t.Fatalf("escribir el estado de disponibilidad: %v", err)
				}

				// Los dos ejes conviven: el par sobrevive en los dos sentidos.
				var deA, deB string
				if err := repo.pool.QueryRow(ctx,
					`SELECT duplicate_of_id::text FROM media_items WHERE id = $1`, a.String()).Scan(&deA); err != nil {
					t.Fatalf("leer el par de A: %v", err)
				}
				if err := repo.pool.QueryRow(ctx,
					`SELECT duplicate_of_id::text FROM media_items WHERE id = $1`, b.String()).Scan(&deB); err != nil {
					t.Fatalf("leer el par de B: %v", err)
				}
				if deA != b.String() || deB != a.String() {
					t.Fatalf("el par se rompió: A→%s, B→%s", deA, deB)
				}

				// El único bloqueo público por marca es el par ABIERTO.
				publicables := publicables(t, repo)
				esperado := parResuelto && tipo.visible
				if got := contiene(publicables, a); got != esperado {
					t.Errorf("publicada = %v, se esperaba %v (par resuelto %v, visible %v)",
						got, esperado, parResuelto, tipo.visible)
				}
			})
		}
	}
}

// TestIntegrationIsFilePublishedFollowsMediaPredicate cubre la rama de F9 de la
// consulta compartida `IsFilePublished` (R9-8/T903): la carátula copiada de una
// ficha **publicable** se sirve; la de una oculta, absorbida o con «posible
// duplicado» abierto **no**, y el archivo **no se borra** (al volver a ser
// publicable, la misma URL vuelve a responder).
func TestIntegrationIsFilePublishedFollowsMediaPredicate(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	// El corte de la regla R4-2 (identidad/agenda) no afecta a la rama de F9,
	// pero la consulta lo pide: se pasa fijo.
	hoy := pgtype.Date{Time: time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC), Valid: true}
	ahora := pgtype.Time{Microseconds: 12 * 3600 * 1e6, Valid: true}

	publicada := func(nombre string) bool {
		t.Helper()
		ok, err := gendb.New(tx).IsFilePublished(ctx, gendb.IsFilePublishedParams{
			FileName: nombre, Today: hoy, NowTime: ahora,
		})
		if err != nil {
			t.Fatalf("IsFilePublished(%q): %v", nombre, err)
		}
		return ok
	}

	// Ficha publicable: su carátula se sirve.
	visible := uuid.New()
	insertFichaTx(t, ctx, tx, visible, "img_medios_visible.png", "visible", nil)

	// Ficha oculta: no se sirve (el archivo sigue en la fila).
	oculta := uuid.New()
	insertFichaTx(t, ctx, tx, oculta, "img_medios_oculta.png", "hidden", nil)

	// Ficha absorbida: no se sirve.
	resultante := uuid.New()
	insertFichaTx(t, ctx, tx, resultante, "img_medios_resultante.png", "visible", nil)
	absorbida := uuid.New()
	resultanteTexto := resultante.String()
	insertFichaTx(t, ctx, tx, absorbida, "img_medios_absorbida.png", "hidden", &resultanteTexto)

	// Par ABIERTO: ninguna de las dos sirve su carátula.
	abiertoA := uuid.New()
	abiertoB := uuid.New()
	insertFichaTx(t, ctx, tx, abiertoA, "img_medios_abierto_a.png", "visible", nil)
	insertFichaTx(t, ctx, tx, abiertoB, "img_medios_abierto_b.png", "visible", nil)
	cerrarParTx(t, ctx, tx, abiertoA, abiertoB, nil)

	// Par RESUELTO: sí se sirve.
	resueltoA := uuid.New()
	resueltoB := uuid.New()
	resuelto := time.Now()
	insertFichaTx(t, ctx, tx, resueltoA, "img_medios_resuelto_a.png", "visible", nil)
	insertFichaTx(t, ctx, tx, resueltoB, "img_medios_resuelto_b.png", "visible", nil)
	cerrarParTx(t, ctx, tx, resueltoA, resueltoB, &resuelto)

	// Marca de disponibilidad sin resolver: la ficha SIGUE visible y sirve su
	// carátula (FR-017).
	fueraDelFeed := uuid.New()
	insertFichaTx(t, ctx, tx, fueraDelFeed, "img_medios_fuera.png", "visible", nil)
	if _, err := tx.Exec(ctx, `UPDATE media_items SET availability_mark = 'missing_from_feed' WHERE id = $1`,
		fueraDelFeed.String()); err != nil {
		t.Fatalf("marcar la disponibilidad: %v", err)
	}

	casos := []struct {
		nombre   string
		archivo  string
		servirse bool
	}{
		{"publicable", "img_medios_visible.png", true},
		{"oculta", "img_medios_oculta.png", false},
		{"resultante de una unión", "img_medios_resultante.png", true},
		{"absorbida", "img_medios_absorbida.png", false},
		{"par abierto A", "img_medios_abierto_a.png", false},
		{"par abierto B", "img_medios_abierto_b.png", false},
		{"par resuelto A", "img_medios_resuelto_a.png", true},
		{"par resuelto B", "img_medios_resuelto_b.png", true},
		{"ya no aparece en el feed", "img_medios_fuera.png", true},
		{"huérfana", "img_medios_huerfana.png", false},
	}
	for _, tt := range casos {
		t.Run(tt.nombre, func(t *testing.T) {
			if got := publicada(tt.archivo); got != tt.servirse {
				t.Errorf("IsFilePublished(%q) = %v, se esperaba %v", tt.archivo, got, tt.servirse)
			}
		})
	}

	// El archivo no se borra al dejar de servirse: ocultar la ficha publicable
	// deja la misma URL lista para responder al volver a mostrarla.
	if _, err := tx.Exec(ctx, `UPDATE media_items SET visibility = 'hidden' WHERE id = $1`, visible.String()); err != nil {
		t.Fatalf("ocultar la ficha: %v", err)
	}
	if publicada("img_medios_visible.png") {
		t.Error("una ficha oculta no debe servir su carátula")
	}
	var guardado *string
	if err := tx.QueryRow(ctx, `SELECT cover_file FROM media_items WHERE id = $1`, visible.String()).Scan(&guardado); err != nil {
		t.Fatalf("leer la carátula: %v", err)
	}
	if guardado == nil || *guardado != "img_medios_visible.png" {
		t.Error("el archivo se borró de la fila: solo debía dejar de servirse")
	}
	if _, err := tx.Exec(ctx, `UPDATE media_items SET visibility = 'visible' WHERE id = $1`, visible.String()); err != nil {
		t.Fatalf("volver a mostrar la ficha: %v", err)
	}
	if !publicada("img_medios_visible.png") {
		t.Error("al volver a mostrar la ficha, la misma URL debe responder otra vez")
	}
}

// insertFichaTx siembra una ficha con su carátula dentro de la transacción de la
// prueba (SQL directo: el predicado se pregunta a la consulta compartida, no al
// repositorio). `unida` es la ficha que la absorbió, si la hay.
func insertFichaTx(t *testing.T, ctx context.Context, tx pgx.Tx, id uuid.UUID, archivo, visibilidad string, unida *string) {
	t.Helper()
	valores := map[string]any{
		"id": id.String(), "title": "Sermón " + archivo, "published_at": time.Now(),
		"origin": "video", "cover_file": archivo, "visibility": visibilidad,
	}
	if unida != nil {
		valores["merged_into_id"] = *unida
	}
	if err := insertRaw(ctx, tx, "media_items", valores); err != nil {
		t.Fatalf("sembrar la ficha %s: %v", archivo, err)
	}
}

// cerrarParTx escribe la pareja recíproca del par «posible duplicado» (cada
// miembro apunta al otro) y, si `resuelto` no es nil, su resolución. La pareja se
// escribe en dos pasos porque la FK impide referenciar una fila que aún no
// existe.
func cerrarParTx(t *testing.T, ctx context.Context, tx pgx.Tx, a, b uuid.UUID, resuelto *time.Time) {
	t.Helper()
	for _, pareja := range [][2]uuid.UUID{{a, b}, {b, a}} {
		query := `UPDATE media_items SET duplicate_of_id = $1 WHERE id = $2`
		args := []any{pareja[1].String(), pareja[0].String()}
		if resuelto != nil {
			query = `UPDATE media_items SET duplicate_of_id = $1, duplicate_mark_resolved_at = $3 WHERE id = $2`
			args = append(args, *resuelto)
		}
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			t.Fatalf("cerrar la pareja del par: %v", err)
		}
	}
}
