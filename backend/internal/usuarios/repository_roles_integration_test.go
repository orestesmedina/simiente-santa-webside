//go:build integration

package usuarios

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/paginate"
)

// Pruebas de integración de repository_roles.go contra PostgreSQL real,
// incluida la carrera del guard anti-bloqueo (FR-008/SC-004).

func TestIntegrationRoleNameLowerAndCatalog(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Misiones", "eventos")

	// Unicidad normalizada: mayúsculas distintas = el mismo rol (Q5).
	if _, err := repo.InsertRole(ctx, "misiones"); err == nil {
		t.Fatal("nombre de rol duplicado (mayúsculas) debería fallar")
	} else {
		assertKind(t, err, apperr.KindConflict)
	}

	// GetRoleByNameLower encuentra el rol aunque la entrada difiera en
	// mayúsculas y lleve espacios en los extremos.
	found, err := repo.GetRoleByNameLower(ctx, "  misiones  ")
	if err != nil {
		t.Fatalf("GetRoleByNameLower: %v", err)
	}
	if found.ID != role.ID {
		t.Fatalf("GetRoleByNameLower devolvió %v, se esperaba %v", found.ID, role.ID)
	}

	if _, err := repo.GetRoleByNameLower(ctx, "no-existe"); err == nil {
		t.Fatal("rol inexistente debería devolver error")
	} else {
		assertKind(t, err, apperr.KindNotFound)
	}

	permissions, err := repo.ListPermissions(ctx)
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(permissions) != 9 {
		t.Fatalf("catálogo de permisos = %d, se esperaban 9", len(permissions))
	}

	ids, err := repo.GetPermissionIDsByCodes(ctx, []string{"eventos", "no-existe"})
	if err != nil {
		t.Fatalf("GetPermissionIDsByCodes: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("GetPermissionIDsByCodes = %v, se esperaba solo 'eventos'", ids)
	}
	if _, ok := ids["no-existe"]; ok {
		t.Fatal("un código inexistente no debe traducirse")
	}
}

func TestIntegrationRoleDetailListAndCount(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos", "noticias")
	detail, err := repo.GetRoleByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRoleByID: %v", err)
	}
	if detail.UserCount != 0 || len(detail.Permissions) != 2 {
		t.Fatalf("GetRoleByID = %+v", detail)
	}

	if _, err := repo.GetRoleByID(ctx, uuid.New()); err == nil {
		t.Fatal("rol inexistente debería devolver error")
	} else {
		assertKind(t, err, apperr.KindNotFound)
	}

	roles, err := repo.ListRoles(ctx, paginate.Params{Limit: 10, Offset: 0})
	if err != nil || len(roles) != 1 {
		t.Fatalf("ListRoles = %d (%v)", len(roles), err)
	}
	total, err := repo.CountRoles(ctx)
	if err != nil || total != 1 {
		t.Fatalf("CountRoles = %d (%v)", total, err)
	}

	renamed, err := repo.UpdateRoleName(ctx, role.ID, "Editora")
	if err != nil || renamed.Name != "Editora" {
		t.Fatalf("UpdateRoleName = %+v (%v)", renamed, err)
	}
}

func TestIntegrationReplaceRolePermissionsAndDelete(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	ids, err := repo.GetPermissionIDsByCodes(ctx, []string{"noticias"})
	if err != nil {
		t.Fatalf("GetPermissionIDsByCodes: %v", err)
	}
	if err := repo.DeleteRolePermissions(ctx, role.ID); err != nil {
		t.Fatalf("DeleteRolePermissions: %v", err)
	}
	if err := repo.InsertRolePermission(ctx, role.ID, ids["noticias"]); err != nil {
		t.Fatalf("InsertRolePermission: %v", err)
	}

	detail, err := repo.GetRoleByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRoleByID: %v", err)
	}
	if len(detail.Permissions) != 1 || detail.Permissions[0] != "noticias" {
		t.Fatalf("permisos tras reemplazo = %v", detail.Permissions)
	}

	if affected, err := repo.DeleteRole(ctx, role.ID); err != nil || affected != 1 {
		t.Fatalf("DeleteRole = %d (%v)", affected, err)
	}
}

func TestIntegrationDeleteRoleWithUsersRESTRICT(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	if count, err := repo.CountRoleUsers(ctx, role.ID); err != nil || count != 1 {
		t.Fatalf("CountRoleUsers = %d (%v)", count, err)
	}

	// FK ON DELETE RESTRICT: el rol en uso no se elimina (FR-017) y el error es
	// un conflicto con el estado actual (409), no una entrada inválida.
	_, err := repo.DeleteRole(ctx, role.ID)
	if err == nil {
		t.Fatal("eliminar un rol con cuentas debería fallar")
	}
	assertKind(t, err, apperr.KindConflict)

	if _, err := repo.GetRoleByID(ctx, role.ID); err != nil {
		t.Fatalf("el rol debe conservarse: %v", err)
	}
}

// TestIntegrationAdminGuardRace demuestra SC-004: dos transacciones
// concurrentes que intentarían dejar el panel sin administración → exactamente
// una completa y queda ≥1 administrador activo.
func TestIntegrationAdminGuardRace(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Administrador", "admin_usuarios_roles")
	first := mustInsertUser(t, repo, "uno@ejemplo.com", role.ID, true)
	second := mustInsertUser(t, repo, "dos@ejemplo.com", role.ID, true)

	total, err := repo.CountActiveAdmins(ctx)
	if err != nil || total != 2 {
		t.Fatalf("CountActiveAdmins = %d (%v), se esperaban 2", total, err)
	}

	users := []User{first, second}
	results := make([]error, len(users))
	start := make(chan struct{})
	var wg sync.WaitGroup

	for i, user := range users {
		wg.Add(1)
		go func(i int, user User) {
			defer wg.Done()
			<-start
			results[i] = repo.WithAdminGuard(context.Background(), func(tx GuardTx) error {
				_, updateErr := tx.UpdateUser(ctx, UserUpdate{
					ID:        user.ID,
					Email:     user.Email,
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Phone:     user.Phone,
					RoleID:    user.RoleID,
					IsActive:  false,
				})
				return updateErr
			})
		}(i, user)
	}
	close(start)
	wg.Wait()

	successes, conflicts := 0, 0
	for _, err := range results {
		if err == nil {
			successes++
			continue
		}
		assertKind(t, err, apperr.KindConflict)
		conflicts++
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("resultados = %v (éxitos=%d, conflictos=%d), se esperaba 1 y 1", results, successes, conflicts)
	}

	remaining, err := repo.CountActiveAdmins(context.Background())
	if err != nil {
		t.Fatalf("CountActiveAdmins tras la carrera: %v", err)
	}
	if remaining < 1 {
		t.Fatalf("el panel quedó sin administración activa: %d", remaining)
	}
}
