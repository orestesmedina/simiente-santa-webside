//go:build integration

package usuarios

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de service_roles.go (T235/T236) contra PostgreSQL real:
// transacción de creación con su registro `role.create`, unicidad normalizada,
// permisos combinables y su existencia en el catálogo, reemplazo de permisos
// reflejado de inmediato (SC-009), guard anti-bloqueo al retirar
// `admin_usuarios_roles` y las reglas de eliminación (FR-017/FR-025).

func integrationRoleService(t *testing.T, repo *repository) *roleService {
	t.Helper()
	logger, _ := testutil.NewLogger()
	return NewRoleService(RoleServiceDeps{
		Repository: repo,
		Audit:      NewAuditService(repo, logger),
	})
}

func TestIntegrationCreateRoleAndCatalog(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	svc := integrationRoleService(t, repo)
	actor := mustInsertAdminUser(t, repo)

	item, err := svc.CreateRole(ctx, actor.ID, RoleCreateInput{
		Name:        "  Contenido   Musical ",
		Permissions: []string{"eventos", "actividades", "eventos"},
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if item.Name != "Contenido Musical" || item.UserCount != 0 {
		t.Fatalf("rol creado = %+v", item)
	}
	if len(item.Permissions) != 2 {
		t.Fatalf("permisos = %v, los repetidos se descartan", item.Permissions)
	}

	// El rol y sus permisos quedan persistidos (FR-014).
	stored, err := repo.GetRoleByID(ctx, uuid.MustParse(item.ID))
	if err != nil {
		t.Fatalf("GetRoleByID: %v", err)
	}
	if stored.Name != "Contenido Musical" || len(stored.Permissions) != 2 {
		t.Fatalf("rol persistido = %+v", stored)
	}
	if !hasAdminAction(t, repo, audit.ActionRoleCreate, audit.ResultSuccess) {
		t.Fatal("no quedó el registro role.create de éxito")
	}

	// Catálogo de FR-015: 9 permisos con etiqueta.
	catalog, err := svc.ListPermissions(ctx)
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(catalog.Items) != 9 {
		t.Fatalf("catálogo = %d, se esperaban 9", len(catalog.Items))
	}

	// Duplicado con otras mayúsculas/espacios → 409 y fallo registrado (Q5).
	_, err = svc.CreateRole(ctx, actor.ID, RoleCreateInput{
		Name: "  contenido musical ", Permissions: []string{"eventos"},
	})
	assertKind(t, err, apperr.KindConflict)
	if !hasAdminAction(t, repo, audit.ActionRoleCreate, audit.ResultFailure) {
		t.Fatal("la creación fallida debe dejar su fila result='failure'")
	}

	// Permiso inexistente → 400.
	_, err = svc.CreateRole(ctx, actor.ID, RoleCreateInput{
		Name: "Otro", Permissions: []string{"eventos", "no-existe"},
	})
	assertKind(t, err, apperr.KindInvalid)
}

func TestIntegrationUpdateRolePermissionsImmediate(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	admin := mustInsertAdminUser(t, repo)
	svc := integrationRoleService(t, repo)

	created, err := svc.CreateRole(ctx, admin.ID, RoleCreateInput{
		Name: "Contenido", Permissions: []string{"eventos"},
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	roleID := uuid.MustParse(created.ID)
	user := mustInsertUser(t, repo, "carlos@ejemplo.com", roleID, true)

	// Los cambios se resuelven por petición: sin tocar la sesión, la siguiente
	// lectura de la identidad de la cuenta ya lleva el conjunto nuevo (SC-009).
	updated, err := svc.UpdateRole(ctx, admin.ID, roleID, RoleUpdateInput{
		Permissions: []string{"medios", "noticias"},
	})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if len(updated.Permissions) != 2 || updated.UserCount != 1 {
		t.Fatalf("rol actualizado = %+v", updated)
	}
	auth, err := repo.GetUserAuthByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetUserAuthByEmail: %v", err)
	}
	if len(auth.Permissions) != 2 {
		t.Fatalf("permisos efectivos de la cuenta = %v, se esperaban los nuevos", auth.Permissions)
	}

	// Renombrar conserva cuentas y permisos.
	renamed, err := svc.UpdateRole(ctx, admin.ID, roleID, RoleUpdateInput{Name: "Contenido General"})
	if err != nil {
		t.Fatalf("UpdateRole (renombrar): %v", err)
	}
	if renamed.Name != "Contenido General" || renamed.UserCount != 1 || len(renamed.Permissions) != 2 {
		t.Fatalf("rol renombrado = %+v", renamed)
	}
	if !hasAdminAction(t, repo, audit.ActionRoleUpdate, audit.ResultSuccess) {
		t.Fatal("no quedó el registro role.update")
	}

	// Dejar el rol sin permisos → 400 (FR-014).
	_, err = svc.UpdateRole(ctx, admin.ID, roleID, RoleUpdateInput{Permissions: []string{}})
	assertKind(t, err, apperr.KindInvalid)
}

func TestIntegrationUpdateRoleAdminGuard(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	svc := integrationRoleService(t, repo)

	// Actor existente (el service no comprueba permisos; eso es del middleware,
	// y la FK de admin_actions solo exige que la cuenta exista) con un rol sin
	// permiso de administración.
	base := mustInsertRole(t, repo, "Sistema", "eventos")
	actor := mustInsertUser(t, repo, "actor@ejemplo.com", base.ID, true)

	adminRole, err := svc.CreateRole(ctx, actor.ID, RoleCreateInput{
		Name: "Administrador", Permissions: []string{"admin_usuarios_roles", "eventos"},
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	adminRoleID := uuid.MustParse(adminRole.ID)
	mustInsertUser(t, repo, "ana@ejemplo.com", adminRoleID, true)

	// Único rol con admin_usuarios_roles: retirárselo deja el panel sin
	// administración → 409 y rollback (FR-008).
	_, err = svc.UpdateRole(ctx, actor.ID, adminRoleID, RoleUpdateInput{
		Permissions: []string{"eventos"},
	})
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Details["reason"] != "admin_required" {
		t.Fatalf("details = %v, se esperaba admin_required", domainErr.Details)
	}
	stored, err := repo.GetRoleByID(ctx, adminRoleID)
	if err != nil {
		t.Fatalf("GetRoleByID: %v", err)
	}
	if len(stored.Permissions) != 2 {
		t.Fatalf("los permisos no deben cambiar tras el rollback: %v", stored.Permissions)
	}
	if !hasAdminAction(t, repo, audit.ActionRoleUpdate, audit.ResultFailure) {
		t.Fatal("el intento bloqueado debe registrarse como fallo")
	}
}

func TestIntegrationDeleteRoleRules(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	admin := mustInsertAdminUser(t, repo)
	svc := integrationRoleService(t, repo)

	// Rol en uso → 409 (FR-017).
	inUse, err := svc.CreateRole(ctx, admin.ID, RoleCreateInput{Name: "EnUso", Permissions: []string{"eventos"}})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	inUseID := uuid.MustParse(inUse.ID)
	mustInsertUser(t, repo, "carlos@ejemplo.com", inUseID, true)

	err = svc.DeleteRole(ctx, admin.ID, inUseID)
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Details["userCount"] != int64(1) {
		t.Fatalf("details = %v, se esperaba userCount=1", domainErr.Details)
	}
	if _, err := repo.GetRoleByID(ctx, inUseID); err != nil {
		t.Fatalf("el rol en uso debe conservarse: %v", err)
	}

	// Rol sin uso → 200; el registro conserva target_label y deja el id en NULL
	// (FR-025).
	free, err := svc.CreateRole(ctx, admin.ID, RoleCreateInput{Name: "Temporal", Permissions: []string{"noticias"}})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	freeID := uuid.MustParse(free.ID)
	if err := svc.DeleteRole(ctx, admin.ID, freeID); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
	if _, err := repo.GetRoleByID(ctx, freeID); err == nil {
		t.Fatal("el rol sin uso debe eliminarse")
	}

	entries, err := repo.ListAdminActions(ctx, AuditFilter{Limit: 100})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	var deletion *AdminActionEntry
	for i := range entries {
		if entries[i].Action == audit.ActionRoleDelete && entries[i].Result == audit.ResultSuccess {
			deletion = &entries[i]
		}
	}
	if deletion == nil {
		t.Fatal("no quedó el registro role.delete")
	}
	if deletion.TargetID != nil {
		t.Fatalf("target_role_id debe quedar NULL tras eliminar el rol: %v", deletion.TargetID)
	}
	if deletion.TargetLabel == nil || *deletion.TargetLabel != "Temporal" {
		t.Fatalf("target_label debe conservar el nombre: %v", deletion.TargetLabel)
	}

	// Rol inexistente → 404.
	assertKind(t, svc.DeleteRole(ctx, admin.ID, uuid.New()), apperr.KindNotFound)
}

func TestIntegrationListRolesPagination(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	admin := mustInsertAdminUser(t, repo)
	svc := integrationRoleService(t, repo)

	for _, name := range []string{"Editor", "Contenido", "Misiones"} {
		if _, err := svc.CreateRole(ctx, admin.ID, RoleCreateInput{Name: name, Permissions: []string{"eventos"}}); err != nil {
			t.Fatalf("CreateRole(%q): %v", name, err)
		}
	}
	list, err := svc.ListRoles(ctx, paginate.Params{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	// Cuatro roles: los tres creados aquí más el del actor (mustInsertAdminUser).
	if list.Total != 4 || len(list.Items) != 2 || list.Limit != 2 {
		t.Fatalf("sobre = %+v", list)
	}
}

// mustInsertAdminUser crea el rol "Administrador" con el permiso de gestión y su
// cuenta activa: varias pruebas necesitan un administrador para que el guard no
// bloquee la operación.
func mustInsertAdminUser(t *testing.T, repo *repository) User {
	t.Helper()
	role := mustInsertRole(t, repo, "Administrador", PermissionAdminUsersRoles)
	return mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)
}
