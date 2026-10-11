//go:build integration

package portada

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/testutil"
)

// T316: verificación ejecutable del modelo de auditoría de F3 contra PostgreSQL
// real (data-model.md §"Validación del modelo"): la migración 000006 es
// reversible (up → down → up) y el registro cerrado de admin_actions admite
// `target_kind='content'` con la coherencia de objetivo exigida.
//
// Estas pruebas NO dependen de la tabla `users`: el catálogo de roles/usuarios
// lo truncan en paralelo las pruebas de integración del dominio usuarios, así
// que se ejercitan las restricciones con el único camino sin actor que la BD
// admite (action='user.create') o con un actor fantasma (UUID inexistente).

// migrationsDir resuelve la carpeta de migraciones desde el directorio del
// paquete (go test ejecuta con CWD = backend/internal/portada).
func migrationsDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "migrations")
}

// execSQLFile ejecuta un archivo .sql completo (varias sentencias) con el
// protocolo simple de pgx. Se usa para reejecutar los .sql de la migración 000006.
func execSQLFile(t *testing.T, ctx context.Context, pool *pgxpool.Pool, path string) {
	t.Helper()
	sql, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leer %s: %v", path, err)
	}
	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		t.Fatalf("ejecutar %s: %v", filepath.Base(path), err)
	}
}

// hasConstraint indica si existe una restricción llamada name sobre admin_actions.
func hasConstraint(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name string) bool {
	t.Helper()
	var exists bool
	const query = `SELECT EXISTS (
		SELECT 1 FROM pg_constraint
		WHERE conname = $1 AND conrelid = 'admin_actions'::regclass)`
	if err := pool.QueryRow(ctx, query, name).Scan(&exists); err != nil {
		t.Fatalf("buscar restricción %s: %v", name, err)
	}
	return exists
}

// constraintDef devuelve la definición textual de una restricción de
// admin_actions (pg_get_constraintdef).
func constraintDef(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name string) string {
	t.Helper()
	var def string
	const query = `SELECT pg_get_constraintdef(oid) FROM pg_constraint
		WHERE conname = $1 AND conrelid = 'admin_actions'::regclass`
	if err := pool.QueryRow(ctx, query, name).Scan(&def); err != nil {
		t.Fatalf("definición de la restricción %s: %v", name, err)
	}
	return def
}

// assertCheckViolation comprueba que err es una violación de CHECK (SQLSTATE
// 23514) de la restricción concreta esperada.
func assertCheckViolation(t *testing.T, err error, constraint string) {
	t.Helper()
	if err == nil {
		t.Fatalf("se esperaba una violación de CHECK (%s)", constraint)
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != codeCheckViolation {
		t.Fatalf("error = %v, se esperaba SQLSTATE %s", err, codeCheckViolation)
	}
	if pgErr.ConstraintName != constraint {
		t.Fatalf("restricción violada = %q, se esperaba %q", pgErr.ConstraintName, constraint)
	}
}

// TestIntegrationMigration000006UpDownUp aplica la migración 000006 up → down
// → up sin errores y comprueba que el registro cerrado queda restaurado y
// ampliado (RG3-2/FR-017).
func TestIntegrationMigration000006UpDownUp(t *testing.T) {
	ctx := context.Background()
	pool := testutil.Pool(t, ctx)

	dir := migrationsDir(t)
	up := filepath.Join(dir, "000006_extend_admin_actions_for_home_content.up.sql")
	down := filepath.Join(dir, "000006_extend_admin_actions_for_home_content.down.sql")

	// Normaliza el punto de partida al estado previo a 000006: si la migración
	// ya está aplicada (BD del CI), se revierte primero para poder ejecutar el
	// ciclo up → down → up literal.
	if hasConstraint(t, ctx, pool, "admin_actions_target_coherence_check") {
		execSQLFile(t, ctx, pool, down)
	}
	if !hasConstraint(t, ctx, pool, "admin_actions_check1") {
		t.Fatal("estado previo a 000006 inesperado: falta admin_actions_check1")
	}

	execSQLFile(t, ctx, pool, up)   // 1) up
	execSQLFile(t, ctx, pool, down) // 2) down
	execSQLFile(t, ctx, pool, up)   // 3) up

	// El CHECK de acción incluye exactamente los 15 códigos home.* de
	// platform/audit (T311): el registro cerrado coincide con la BD.
	actionDef := constraintDef(t, ctx, pool, "admin_actions_action_check")
	homeCodes := 0
	for _, code := range audit.ActionCodes {
		if !strings.HasPrefix(code, "home.") {
			continue
		}
		homeCodes++
		if !strings.Contains(actionDef, "'"+code+"'") {
			t.Fatalf("el CHECK de admin_actions no incluye el código %q", code)
		}
	}
	if homeCodes != 15 {
		t.Fatalf("platform/audit declara %d códigos home.*, se esperaban 15", homeCodes)
	}
	if got := strings.Count(actionDef, "'home."); got != homeCodes {
		t.Fatalf("el CHECK de la BD tiene %d códigos home.* y platform/audit declara %d", got, homeCodes)
	}

	// target_kind admite user, role y content.
	kindDef := constraintDef(t, ctx, pool, "admin_actions_target_kind_check")
	for _, kind := range []string{"'user'", "'role'", "'content'"} {
		if !strings.Contains(kindDef, kind) {
			t.Fatalf("target_kind no admite %s: %s", kind, kindDef)
		}
	}

	// La coherencia de objetivo queda con el nombre nuevo tras el up y con el
	// de F2 tras el down.
	if !hasConstraint(t, ctx, pool, "admin_actions_target_coherence_check") {
		t.Fatal("falta admin_actions_target_coherence_check tras el up")
	}
	if hasConstraint(t, ctx, pool, "admin_actions_check1") {
		t.Fatal("admin_actions_check1 no debería existir tras el up")
	}
}

// TestIntegrationAdminActionContentCheck verifica que una fila de auditoría con
// objetivo `content` (ambas FK en NULL y target_label) se inserta y que el
// CHECK de coherencia rechaza una FK rellena o un target_label ausente (FR-017).
func TestIntegrationAdminActionContentCheck(t *testing.T) {
	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	ghost := uuid.New()

	// Fila válida: el único camino sin actor que admite la BD es
	// action='user.create' (CHECK del actor); se usa aquí para ejercitar el
	// CHECK de coherencia de `content` sin depender de la tabla `users`.
	if _, err := pool.Exec(ctx, `INSERT INTO admin_actions
		(actor_user_id, action, target_kind, target_user_id, target_role_id, target_label, result)
		VALUES (NULL, 'user.create', 'content', NULL, NULL, 'Portada · Identidad · X', 'success')`); err != nil {
		t.Fatalf("insertar acción content válida: %v", err)
	}

	t.Run("FK de usuario rellena", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO admin_actions
			(actor_user_id, action, target_kind, target_user_id, target_label, result)
			VALUES (NULL, 'user.create', 'content', $1, 'Portada · Quiénes somos', 'success')`, ghost)
		assertCheckViolation(t, err, "admin_actions_target_coherence_check")
	})

	t.Run("target_label ausente", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO admin_actions
			(actor_user_id, action, target_kind, result)
			VALUES (NULL, 'user.create', 'content', 'success')`)
		assertCheckViolation(t, err, "admin_actions_target_coherence_check")
	})

	t.Run("código de acción fuera del registro", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO admin_actions
			(actor_user_id, action, target_kind, target_label, result)
			VALUES ($1, 'home.inexistente', 'content', 'x', 'success')`, ghost)
		assertCheckViolation(t, err, "admin_actions_action_check")
	})
}
