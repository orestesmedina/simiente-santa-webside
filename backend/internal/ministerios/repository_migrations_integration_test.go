//go:build integration

package ministerios

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de la capa de datos de F6 contra PostgreSQL real
// (T611): la verificación ejecutable de `data-model.md` §«Validación del
// modelo» en lo que toca al esquema.
//
//   1. las migraciones `000010` y `000011` aplican up → down → up sin errores
//      (incluida la restauración del CHECK de `admin_actions`);
//   2. el catálogo de `permissions` sigue en 10 filas con `ministerios` entre
//      ellas (F6 no lo toca; los `down` no lo rozan);
//   3. los `CHECK` de la tabla se disparan en la BD y su nombre coincide con el
//      que traduce el repository al detalle por campo (T607);
//   4. «Oculto por defecto»: un `INSERT` que omite `leader_visible` guarda
//      `FALSE` y su lectura pública sale sin ningún dato del encargado.
//
// Aislamiento: cada prueba trabaja dentro de una TRANSACCIÓN que se revierte al
// terminar (el DDL de PostgreSQL es transaccional), así que los `down` —que son
// destructivos a propósito— no tocan el estado que ven las pruebas de
// integración de los otros dominios, que corren en paralelo. Esas transacciones
// además toman el lock de testutil.SerializeMigrations (ver migrationTx), porque
// en CI todas las pruebas comparten una única base y dos `down` destructivos
// solapados se matan entre sí con 40P01. La atomicidad de
// mutación + auditoría y el registro cerrado de códigos se verifican en
// repository_audit_integration_test.go.

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

// testTx abre una transacción de prueba que se revierte al terminar.
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

// migrationTx abre la transacción de una prueba de MIGRACIÓN: además de
// revertirse al terminar, toma el lock de serialización de testutil como primera
// sentencia. Las pruebas de auditoría y las de CHECK siguen usando testTx: solo
// los `down` destructivos (DELETE + ALTER TABLE de `admin_actions`) pueden
// matarse entre sí con 40P01 cuando los paquetes comparten base (ver
// testutil.SerializeMigrations).
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

// insertRawMinistry inserta en `ministries` las columnas indicadas (nombres
// literales del esquema) con valores parametrizados y devuelve el error de
// PostgreSQL sin abortar la prueba: las columnas que no aparecen en el mapa
// toman su DEFAULT (o NULL), que es justo lo que algunas pruebas necesitan.
func insertRawMinistry(ctx context.Context, tx pgx.Tx, values map[string]any) error {
	columns := make([]string, 0, len(values))
	for name := range values {
		columns = append(columns, name)
	}
	sort.Strings(columns)

	placeholders := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	for i, name := range columns {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, values[name])
	}

	query := fmt.Sprintf("INSERT INTO ministries (%s) VALUES (%s)",
		strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	_, err := tx.Exec(ctx, query, args...)
	return err
}

// insertInSavepoint ejecuta el INSERT dentro de un savepoint y lo revierte
// siempre: en PostgreSQL un error aborta la transacción entera, así que sin el
// savepoint el primer caso inválido dejaría la transacción de prueba inservible
// para los siguientes; y revirtiendo también los aciertos, ningún caso deja
// filas que interfieran con el siguiente.
func insertInSavepoint(ctx context.Context, tx pgx.Tx, values map[string]any) error {
	sp, err := tx.Begin(ctx)
	if err != nil {
		return err
	}
	insertErr := insertRawMinistry(ctx, sp, values)
	if rollbackErr := sp.Rollback(ctx); rollbackErr != nil && insertErr == nil {
		return rollbackErr
	}
	return insertErr
}

// constraintName devuelve el nombre de la restricción que rompió un INSERT.
func constraintName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// TestIntegrationMigration000010UpDownUp aplica `000010_create_ministries` up →
// down → up y comprueba que la tabla queda como la describe `data-model.md`:
// columnas obligatorias, `leader_visible` con `DEFAULT FALSE` («Oculto por
// defecto», P6-3) y los dos índices del orden y del filtro público.
func TestIntegrationMigration000010UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const (
		up   = "000010_create_ministries.up.sql"
		down = "000010_create_ministries.down.sql"
	)

	// Punto de partida normalizado: sin la tabla (la migración ya está aplicada
	// en la base de pruebas).
	execMigrationInTx(t, ctx, tx, down)
	if got := countTx(t, ctx, tx,
		`SELECT count(*) FROM pg_class WHERE relname = 'ministries' AND relkind = 'r'`); got != 0 {
		t.Fatalf("la tabla sigue existiendo tras el down: %d", got)
	}

	execMigrationInTx(t, ctx, tx, up) // 1) up
	assertMinistriesSchema(t, ctx, tx)

	execMigrationInTx(t, ctx, tx, down) // 2) down
	if got := countTx(t, ctx, tx,
		`SELECT count(*) FROM pg_class WHERE relname = 'ministries' AND relkind = 'r'`); got != 0 {
		t.Fatalf("la tabla sigue existiendo tras el segundo down: %d", got)
	}

	execMigrationInTx(t, ctx, tx, up) // 3) up
	assertMinistriesSchema(t, ctx, tx)
}

// assertMinistriesSchema comprueba el estado esperado de la tabla `ministries`.
func assertMinistriesSchema(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	// El interruptor de la ficha del encargado nace apagado y lo garantiza la
	// propia columna (un INSERT que la omita guarda FALSE).
	var defaultValue *string
	if err := tx.QueryRow(ctx, `SELECT column_default FROM information_schema.columns
		WHERE table_name = 'ministries' AND column_name = 'leader_visible'`).Scan(&defaultValue); err != nil {
		t.Fatalf("leer el DEFAULT de leader_visible: %v", err)
	}
	if defaultValue == nil || !strings.HasPrefix(*defaultValue, "false") {
		t.Fatalf("leader_visible DEFAULT = %v, se esperaba false («Oculto por defecto»)", defaultValue)
	}

	// Obligatorios en español (el nombre y la descripción son NOT NULL).
	for _, column := range []string{"name_es", "description_es"} {
		var nullable string
		if err := tx.QueryRow(ctx, `SELECT is_nullable FROM information_schema.columns
			WHERE table_name = 'ministries' AND column_name = $1`, column).Scan(&nullable); err != nil {
			t.Fatalf("leer is_nullable de %s: %v", column, err)
		}
		if nullable != "NO" {
			t.Errorf("%s es anulable, se esperaba NOT NULL", column)
		}
	}

	// El estado de publicación por elemento y el orden manual, con sus defaults.
	var stateDefault, sortDefault *string
	if err := tx.QueryRow(ctx, `SELECT
			(SELECT column_default FROM information_schema.columns
			  WHERE table_name = 'ministries' AND column_name = 'publication_state'),
			(SELECT column_default FROM information_schema.columns
			  WHERE table_name = 'ministries' AND column_name = 'sort_order')`).Scan(&stateDefault, &sortDefault); err != nil {
		t.Fatalf("leer los defaults de estado y orden: %v", err)
	}
	if stateDefault == nil || !strings.Contains(*stateDefault, "draft") {
		t.Errorf("publication_state DEFAULT = %v, se esperaba 'draft'", stateDefault)
	}
	if sortDefault == nil || !strings.HasPrefix(*sortDefault, "0") {
		t.Errorf("sort_order DEFAULT = %v, se esperaba 0", sortDefault)
	}

	// Los dos índices del módulo: el orden público/panel y el filtro por estado.
	indexes := countTx(t, ctx, tx, `SELECT count(*) FROM pg_indexes
		WHERE tablename = 'ministries'
		  AND indexname IN ('ministries_sort_order_id_idx', 'ministries_publication_state_idx')`)
	if indexes != 2 {
		t.Errorf("índices de ministries = %d, se esperaban 2 (orden y estado)", indexes)
	}
}

// TestIntegrationMigration000011UpDownUp aplica
// `000011_extend_admin_actions_for_ministries` up → down → up y comprueba que el
// registro cerrado de códigos gana los 8 `ministry.*` y que el `down` restaura
// el de F4 sin tocar los códigos de F2/F3/F4.
func TestIntegrationMigration000011UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const (
		up   = "000011_extend_admin_actions_for_ministries.up.sql"
		down = "000011_extend_admin_actions_for_ministries.down.sql"
	)

	// Punto de partida normalizado: el CHECK previo a F6.
	execMigrationInTx(t, ctx, tx, down)
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 1) up
	assertActionCheck(t, ctx, tx, true)

	execMigrationInTx(t, ctx, tx, down) // 2) down
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 3) up
	assertActionCheck(t, ctx, tx, true)

	// El tipo de objetivo NO cambia: los ministerios son `content` como los
	// elementos de F3/F4 (R6-12).
	kindDef := constraintDefTx(t, ctx, tx, "admin_actions_target_kind_check", "admin_actions")
	for _, kind := range []string{"'user'", "'role'", "'content'"} {
		if !strings.Contains(kindDef, kind) {
			t.Fatalf("target_kind no admite %s: %s", kind, kindDef)
		}
	}
}

// assertActionCheck comprueba que el CHECK de códigos incluye (o no) los 8
// códigos de F6 y que los de F2/F3/F4 siguen intactos (regresión).
func assertActionCheck(t *testing.T, ctx context.Context, tx pgx.Tx, wantF6 bool) {
	t.Helper()
	def := constraintDefTx(t, ctx, tx, "admin_actions_action_check", "admin_actions")

	ministryCodes := ministryActionCodes()
	if len(ministryCodes) != 8 {
		t.Fatalf("platform/audit declara %d códigos ministry.*, se esperaban 8", len(ministryCodes))
	}
	for _, code := range ministryCodes {
		if got := strings.Contains(def, "'"+code+"'"); got != wantF6 {
			t.Fatalf("el CHECK incluye %q = %v, se esperaba %v", code, got, wantF6)
		}
	}
	if inCheck := strings.Count(def, "'ministry."); inCheck != 8 && wantF6 {
		t.Fatalf("el CHECK tiene %d códigos ministry.*, se esperaban 8", inCheck)
	}
	if inCheck := strings.Count(def, "'ministry."); inCheck != 0 && !wantF6 {
		t.Fatalf("el CHECK de F4 tiene %d códigos ministry.*, se esperaban 0", inCheck)
	}

	// Los códigos de las funcionalidades anteriores no se pierden (uno
	// representativo por módulo, verificado también en las pruebas de F3/F4).
	for _, code := range []string{"user.create", "home.publish", "event.create", "activity.publish", "category.delete"} {
		if !strings.Contains(def, "'"+code+"'") {
			t.Fatalf("el CHECK perdió %q: %s", code, def)
		}
	}
}

// TestIntegrationPermissionsCatalogUnchanged fija que F6 **no** toca el catálogo
// de permisos de F2 (FR-011): sigue con 10 permisos, `ministerios` entre ellos, y
// ninguno se llama `ministry.*`; los `down` de F6 tampoco lo rozan.
func TestIntegrationPermissionsCatalogUnchanged(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const catalog = `SELECT count(*) FROM permissions`

	if got := countTx(t, ctx, tx, catalog); got != 10 {
		t.Fatalf("permisos = %d, se esperaban 10", got)
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM permissions WHERE code = 'ministerios'`); got != 1 {
		t.Fatalf("el permiso `ministerios` no está en el catálogo: %d", got)
	}
	var label string
	if err := tx.QueryRow(ctx, `SELECT label FROM permissions WHERE code = 'ministerios'`).Scan(&label); err != nil {
		t.Fatalf("leer la etiqueta de `ministerios`: %v", err)
	}
	if label != "Ministerios" {
		t.Errorf("etiqueta de `ministerios` = %q, se esperaba «Ministerios»", label)
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM permissions WHERE code LIKE 'ministry%'`); got != 0 {
		t.Errorf("F6 creó %d permisos nuevos, se esperaban 0", got)
	}

	// Los `down` de F6 no tocan el catálogo.
	execMigrationInTx(t, ctx, tx, "000011_extend_admin_actions_for_ministries.down.sql")
	execMigrationInTx(t, ctx, tx, "000010_create_ministries.down.sql")
	if got := countTx(t, ctx, tx, catalog); got != 10 {
		t.Fatalf("permisos tras los down de F6 = %d, se esperaban 10", got)
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM permissions WHERE code = 'ministerios'`); got != 1 {
		t.Fatalf("los down de F6 retiraron el permiso `ministerios`: %d", got)
	}
}

// TestIntegrationMinistryChecksFireInTheDatabase comprueba que cada invariante
// del modelo está protegido por la BD y que el nombre de la restricción que
// PostgreSQL reporta coincide con el que el repository traduce al detalle por
// campo (T607/FR-014/FR-015/FR-020). Los casos dicen también qué columnas
// quedan fuera por construcción: `name_es` y `description_es` son `NOT NULL`, así
// que los CHECK «*_en exige *_es» de esos dos pares no pueden alcanzarse (el
// `NOT NULL` es la garantía efectiva).
func TestIntegrationMinistryChecksFireInTheDatabase(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	const (
		name = "Alabanza"
		desc = "Grupo de alabanza"
	)

	tests := []struct {
		name       string
		values     map[string]any
		constraint string
		field      string
	}{
		{
			name: "celular y correo juntos",
			values: map[string]any{
				"name_es": name, "description_es": desc,
				"leader_name_es": "Ana", "leader_phone": "+584120000000", "leader_email": "ana@example.com",
			},
			constraint: "ministries_check",
			field:      "leaderEmail",
		},
		{
			name: "contacto sin nombre del encargado",
			values: map[string]any{
				"name_es": name, "description_es": desc, "leader_phone": "+584120000000",
			},
			constraint: "ministries_check1",
			field:      "leaderNameEs",
		},
		{
			name: "encargado en inglés sin español",
			values: map[string]any{
				"name_es": name, "description_es": desc, "leader_name_en": "Ana Perez",
			},
			constraint: "ministries_check4",
			field:      "leaderNameEs",
		},
		{
			name: "horario en inglés sin español",
			values: map[string]any{
				"name_es": name, "description_es": desc, "schedule_en": "Fridays",
			},
			constraint: "ministries_check5",
			field:      "scheduleEs",
		},
		{
			name: "lugar en inglés sin español",
			values: map[string]any{
				"name_es": name, "description_es": desc, "place_en": "Temple",
			},
			constraint: "ministries_check6",
			field:      "placeEs",
		},
		{
			name: "texto alternativo en inglés sin español",
			values: map[string]any{
				"name_es": name, "description_es": desc,
				"image_file": "img_11111111-1111-1111-1111-111111111111.jpg", "image_alt_en": "Alt",
			},
			constraint: "ministries_check7",
			field:      "imageAltEs",
		},
		{
			name: "imagen sin texto alternativo",
			values: map[string]any{
				"name_es": name, "description_es": desc,
				"image_file": "img_11111111-1111-1111-1111-111111111111.jpg",
			},
			constraint: "ministries_check8",
			field:      "imageAltEs",
		},
		{
			name: "enlace http",
			values: map[string]any{
				"name_es": name, "description_es": desc, "whatsapp_url": "http://chat.whatsapp.com/x",
			},
			constraint: "ministries_whatsapp_url_check1",
			field:      "whatsappUrl",
		},
		{
			name: "enlace javascript",
			values: map[string]any{
				"name_es": name, "description_es": desc, "whatsapp_url": "javascript:alert(1)",
			},
			constraint: "ministries_whatsapp_url_check1",
			field:      "whatsappUrl",
		},
		{
			name: "enlace con espacio",
			values: map[string]any{
				"name_es": name, "description_es": desc, "whatsapp_url": "https://chat.whatsapp.com/a b",
			},
			constraint: "ministries_whatsapp_url_check1",
			field:      "whatsappUrl",
		},
		{
			name: "celular con espacios sobrantes",
			values: map[string]any{
				"name_es": name, "description_es": desc,
				"leader_name_es": "Ana", "leader_phone": " +584120000000 ",
			},
			constraint: "ministries_leader_phone_check1",
			field:      "leaderPhone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := insertInSavepoint(ctx, tx, tt.values)
			if err == nil {
				t.Fatal("la BD aceptó una fila que viola un invariante")
			}
			if got := constraintName(err); got != tt.constraint {
				t.Fatalf("restricción = %q, se esperaba %q (%v)", got, tt.constraint, err)
			}
			if details := checkDetails(constraintName(err)); details == nil || details[tt.field] == nil {
				t.Fatalf("checkDetails(%q) = %v, se esperaba el campo %q", tt.constraint, details, tt.field)
			}
		})
	}

	// Los enlaces `https` válidos de cualquier dominio sí entran (Q3).
	for _, url := range []string{"https://chat.whatsapp.com/AbCdEf", "https://ejemplo.com/grupo"} {
		if err := insertInSavepoint(ctx, tx, map[string]any{
			"name_es": name + " " + url, "description_es": desc, "whatsapp_url": url,
		}); err != nil {
			t.Fatalf("la BD rechazó un enlace https válido (%s): %v", url, err)
		}
	}
}

// TestIntegrationHiddenByDefault fija «Oculto por defecto» (P6-3/RG6-10) en la
// capa de datos: un INSERT que OMITE `leader_visible` guarda `FALSE` y, con
// nombre y celular del encargado guardados, la lectura pública sale sin ningún
// dato del encargado (ni el nombre ni el contacto), mientras el panel los ve.
func TestIntegrationHiddenByDefault(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)
	repo := &repository{q: gendb.New(tx)}

	const (
		name = "Alabanza"
		desc = "Grupo de alabanza"
	)

	if err := insertRawMinistry(ctx, tx, map[string]any{
		"name_es": name + " " + t.Name(), "description_es": desc,
		"leader_name_es": "Ana Pérez", "leader_phone": "+584120000000",
		"publication_state": "published",
	}); err != nil {
		t.Fatalf("INSERT que omite leader_visible: %v", err)
	}

	var visible bool
	if err := tx.QueryRow(ctx, `SELECT leader_visible FROM ministries WHERE name_es = $1`,
		name+" "+t.Name()).Scan(&visible); err != nil {
		t.Fatalf("leer leader_visible: %v", err)
	}
	if visible {
		t.Error("leader_visible = true tras un INSERT que omite la columna, se esperaba false")
	}

	items, err := repo.ListPublishedMinistries(ctx, 200)
	if err != nil {
		t.Fatalf("ListPublishedMinistries: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("la lectura pública devolvió %d ministerios, se esperaba 1", len(items))
	}
	assertLeaderHidden(t, "alta que omite leader_visible", items[0])
	if items[0].LeaderVisible {
		t.Error("la lectura pública dice que la ficha está publicada")
	}
}
