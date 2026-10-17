//go:build integration

package agenda

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/testutil"
)

// T416: verificación ejecutable de `data-model.md` §"Validación del modelo" en lo
// que toca a las migraciones de F4: `000008` (registro cerrado de auditoría) y
// `000009` (permiso `categorias` en el catálogo de F2) aplican up → down → up sin
// errores y dejan el esquema en el estado esperado.
//
// Aislamiento: cada ciclo se ejecuta dentro de una TRANSACCIÓN que se revierte al
// terminar (el DDL de PostgreSQL es transaccional), así que el `down` —que es
// destructivo a propósito— no toca el estado que ven las pruebas de integración
// de los otros dominios, que corren en paralelo. Esas transacciones además toman
// el lock de testutil.SerializeMigrations (ver migrationTx), porque en CI todas
// las pruebas comparten una única base y dos `down` destructivos solapados se
// matan entre sí con 40P01.
//
// La regla "próximo" (los tres casos de R4-2) y la atomicidad mutación +
// auditoría se verifican en las pruebas de eventos (T412), actividades (T413),
// media (T415) y la base de escritura (T411).

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
// sentencia. Las pruebas de auditoría siguen usando testTx: solo los `down`
// destructivos (DELETE + ALTER TABLE de `admin_actions`) pueden matarse entre sí
// con 40P01 cuando los paquetes comparten base (ver testutil.SerializeMigrations).
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
func countTx(t *testing.T, ctx context.Context, tx pgx.Tx, query string) int {
	t.Helper()
	var total int
	if err := tx.QueryRow(ctx, query).Scan(&total); err != nil {
		t.Fatalf("contar (%s): %v", query, err)
	}
	return total
}

// TestIntegrationMigration000008UpDownUp aplica la migración 000008 up → down →
// up y comprueba que el registro cerrado de `admin_actions` queda ampliado con
// los códigos de F4 y que el down restaura el de F3 (FR-017/R4-10).
func TestIntegrationMigration000008UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const (
		up   = "000008_extend_admin_actions_for_agenda.up.sql"
		down = "000008_extend_admin_actions_for_agenda.down.sql"
	)

	// Punto de partida normalizado: el estado previo a 000008 (la migración ya
	// está aplicada en la base de pruebas).
	execMigrationInTx(t, ctx, tx, down)
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 1) up
	assertActionCheck(t, ctx, tx, true)

	execMigrationInTx(t, ctx, tx, down) // 2) down
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 3) up
	assertActionCheck(t, ctx, tx, true)

	// El tipo de objetivo NO cambia: los elementos de F4 son `content` como los
	// de F3 (R4-10).
	kindDef := constraintDefTx(t, ctx, tx, "admin_actions_target_kind_check", "admin_actions")
	for _, kind := range []string{"'user'", "'role'", "'content'"} {
		if !strings.Contains(kindDef, kind) {
			t.Fatalf("target_kind no admite %s: %s", kind, kindDef)
		}
	}
}

// assertActionCheck comprueba que el CHECK de códigos de acción incluye (o no)
// los tres códigos representativos de F4, uno por módulo.
func assertActionCheck(t *testing.T, ctx context.Context, tx pgx.Tx, wantF4 bool) {
	t.Helper()
	def := constraintDefTx(t, ctx, tx, "admin_actions_action_check", "admin_actions")
	for _, code := range []string{"event.create", "activity.publish", "category.delete"} {
		if got := strings.Contains(def, "'"+code+"'"); got != wantF4 {
			t.Fatalf("el CHECK incluye %q = %v, se esperaba %v", code, got, wantF4)
		}
	}
}

// TestIntegrationMigration000009UpDownUp aplica la migración 000009 up → down →
// up y comprueba el catálogo de permisos de F2: 10 permisos con `categorias`
// (R4-4) y la reversa sin filas huérfanas en `role_permissions`.
func TestIntegrationMigration000009UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	const (
		up            = "000009_seed_permission_categorias.up.sql"
		down          = "000009_seed_permission_categorias.down.sql"
		catalogCount  = `SELECT count(*) FROM permissions`
		hasCategorias = `SELECT count(*) FROM permissions WHERE code = 'categorias'`
		orphans       = `SELECT count(*) FROM role_permissions rp
			LEFT JOIN permissions p ON p.id = rp.permission_id
			WHERE p.id IS NULL`
	)

	// Punto de partida normalizado: sin el permiso nuevo (el down retira la fila
	// y sus asignaciones; si no estuviera aplicada, no hace nada).
	execMigrationInTx(t, ctx, tx, down)
	if got := countTx(t, ctx, tx, hasCategorias); got != 0 {
		t.Fatalf("`categorias` sigue en el catálogo tras el down: %d", got)
	}
	if got := countTx(t, ctx, tx, catalogCount); got != 9 {
		t.Fatalf("permisos tras el down = %d, se esperaban 9", got)
	}

	execMigrationInTx(t, ctx, tx, up) // 1) up
	if got := countTx(t, ctx, tx, hasCategorias); got != 1 {
		t.Fatalf("`categorias` no está en el catálogo tras el up: %d", got)
	}
	if got := countTx(t, ctx, tx, catalogCount); got != 10 {
		t.Fatalf("permisos tras el up = %d, se esperaban 10", got)
	}

	execMigrationInTx(t, ctx, tx, down) // 2) down
	if got := countTx(t, ctx, tx, catalogCount); got != 9 {
		t.Fatalf("permisos tras el segundo down = %d, se esperaban 9", got)
	}
	if got := countTx(t, ctx, tx, orphans); got != 0 {
		t.Fatalf("el down dejó %d asignaciones huérfanas en role_permissions", got)
	}

	execMigrationInTx(t, ctx, tx, up) // 3) up
	if got := countTx(t, ctx, tx, catalogCount); got != 10 {
		t.Fatalf("permisos tras el segundo up = %d, se esperaban 10", got)
	}
	if got := countTx(t, ctx, tx, orphans); got != 0 {
		t.Fatalf("quedaron %d asignaciones huérfanas en role_permissions", got)
	}

	// El permiso nuevo tiene la etiqueta del contrato y las asignaciones que
	// tenía el rol que lo usaba se restauran solo si el rol las tenía (el down es
	// destructivo a propósito, documentado en data-model.md).
	var label string
	if err := tx.QueryRow(ctx, `SELECT label FROM permissions WHERE code = 'categorias'`).Scan(&label); err != nil {
		t.Fatalf("leer la etiqueta de `categorias`: %v", err)
	}
	if label != "Categorías de agenda" {
		t.Errorf("etiqueta = %q, se esperaba «Categorías de agenda»", label)
	}
}
