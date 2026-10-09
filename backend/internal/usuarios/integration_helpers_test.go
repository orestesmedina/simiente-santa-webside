//go:build integration

package usuarios

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/testutil"
)

// Helpers compartidos de las pruebas de integración del dominio usuarios
// (T221–T223). Requieren una PostgreSQL real con las migraciones aplicadas:
// DATABASE_URL_TEST (lo aporta el CI o el compose local; si no existe, las
// pruebas se omiten solas).
//
// Estas pruebas truncan las tablas del dominio para partir de un estado
// conocido; por eso DATABASE_URL_TEST debe apuntar a una base de datos de
// pruebas dedicada (nunca a la de desarrollo).

// newTestRepo construye el repositorio contra la base de datos de pruebas y
// deja las tablas del dominio vacías.
func newTestRepo(t *testing.T) (*repository, *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	cleanDomainTables(t, pool)
	return NewRepository(pool), pool
}

// cleanDomainTables vacía las tablas de negocio respetando las FK. El catálogo
// `permissions` es fijo (FR-015) y no se toca.
func cleanDomainTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	const query = `TRUNCATE login_events, admin_actions, role_permissions, roles, users CASCADE`
	if _, err := pool.Exec(context.Background(), query); err != nil {
		t.Fatalf("vaciar tablas del dominio: %v", err)
	}
}

// mustInsertRole crea un rol con los códigos de permiso indicados.
func mustInsertRole(t *testing.T, repo *repository, name string, codes ...string) Role {
	t.Helper()

	ctx := context.Background()
	role, err := repo.InsertRole(ctx, name)
	if err != nil {
		t.Fatalf("InsertRole(%q): %v", name, err)
	}
	if len(codes) == 0 {
		return role
	}
	ids, err := repo.GetPermissionIDsByCodes(ctx, codes)
	if err != nil {
		t.Fatalf("GetPermissionIDsByCodes: %v", err)
	}
	for _, code := range codes {
		id, ok := ids[code]
		if !ok {
			t.Fatalf("el código de permiso %q no existe en el catálogo", code)
		}
		if err := repo.InsertRolePermission(ctx, role.ID, id); err != nil {
			t.Fatalf("InsertRolePermission(%q): %v", code, err)
		}
	}
	return role
}

// mustInsertUser crea una cuenta con la contraseña y rol indicados.
func mustInsertUser(t *testing.T, repo *repository, email string, roleID uuid.UUID, active bool) User {
	t.Helper()

	user, err := repo.InsertUser(context.Background(), NewUser{
		Email:              email,
		FirstName:          "Ana",
		LastName:           "Pérez",
		Phone:              "612345678",
		PasswordHash:       "hash-de-prueba",
		MustChangePassword: true,
		IsActive:           active,
		RoleID:             roleID,
	})
	if err != nil {
		t.Fatalf("InsertUser(%q): %v", email, err)
	}
	return user
}
