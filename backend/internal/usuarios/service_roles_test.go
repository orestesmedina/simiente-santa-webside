package usuarios

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// Tests unitarios de service_roles.go (T235/T236) con fakes. No tocan
// PostgreSQL: el fake implementa a la vez RoleRepository y GuardTx y observa
// cada escritura y cada transacción (WithTx para crear, WithAdminGuard para
// editar/eliminar).

// --- Fake del repositorio de roles ---

type fakeRoleRepo struct {
	mu sync.Mutex

	roles       map[uuid.UUID]Role
	permissions []Permission
	permIDs     map[string]uuid.UUID

	getErr        error
	nameErr       error
	listErr       error
	countErr      error
	catalogErr    error
	idsErr        error
	countUsersErr error
	guardErr      error

	insertRoleErr  error
	updateNameErr  error
	deletePermsErr error
	insertPermErr  error
	deleteRoleErr  error
	actionErr      error

	txCalls    int
	guardCalls int

	insertedRoles []string
	renamed       []string
	deletedPerms  []uuid.UUID
	insertedPerms []uuid.UUID
	deletedRoles  []uuid.UUID
	actions       []audit.Action
}

func newFakeRoleRepo() *fakeRoleRepo {
	f := &fakeRoleRepo{
		roles:       map[uuid.UUID]Role{},
		permIDs:     map[string]uuid.UUID{},
		permissions: sampleCatalog(),
	}
	for _, permission := range f.permissions {
		f.permIDs[permission.Code] = permission.ID
	}
	return f
}

// seedRole registra un rol ya existente. userCount fija cuántas cuentas lo usan.
func (f *fakeRoleRepo) seedRole(name string, userCount int64, codes ...string) Role {
	f.mu.Lock()
	defer f.mu.Unlock()
	role := Role{ID: uuid.New(), Name: name, Permissions: codes, UserCount: userCount, CreatedAt: time.Now()}
	f.roles[role.ID] = role
	return role
}

func (f *fakeRoleRepo) WithTx(_ context.Context, mutate func(tx GuardTx) error) error {
	f.mu.Lock()
	f.txCalls++
	f.mu.Unlock()
	return mutate(f)
}

func (f *fakeRoleRepo) WithAdminGuard(_ context.Context, mutate func(tx GuardTx) error) error {
	f.mu.Lock()
	f.guardCalls++
	guardErr := f.guardErr
	f.mu.Unlock()

	if err := mutate(f); err != nil {
		return err
	}
	return guardErr
}

func (f *fakeRoleRepo) GetRoleByID(_ context.Context, id uuid.UUID) (Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return Role{}, f.getErr
	}
	role, ok := f.roles[id]
	if !ok {
		return Role{}, apperr.NotFound("El rol no existe")
	}
	return role, nil
}

func (f *fakeRoleRepo) GetRoleByNameLower(_ context.Context, name string) (Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.nameErr != nil {
		return Role{}, f.nameErr
	}
	for _, role := range f.roles {
		if strings.EqualFold(role.Name, name) {
			return role, nil
		}
	}
	return Role{}, apperr.NotFound("El rol no existe")
}

func (f *fakeRoleRepo) ListRoles(_ context.Context, _ paginate.Params) ([]Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.listErr != nil {
		return nil, f.listErr
	}
	roles := make([]Role, 0, len(f.roles))
	for _, role := range f.roles {
		roles = append(roles, role)
	}
	return roles, nil
}

func (f *fakeRoleRepo) CountRoles(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.countErr != nil {
		return 0, f.countErr
	}
	return int64(len(f.roles)), nil
}

func (f *fakeRoleRepo) ListPermissions(context.Context) ([]Permission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.catalogErr != nil {
		return nil, f.catalogErr
	}
	return f.permissions, nil
}

func (f *fakeRoleRepo) GetPermissionIDsByCodes(_ context.Context, codes []string) (map[string]uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.idsErr != nil {
		return nil, f.idsErr
	}
	ids := make(map[string]uuid.UUID, len(codes))
	for _, code := range codes {
		if id, ok := f.permIDs[code]; ok {
			ids[code] = id
		}
	}
	return ids, nil
}

func (f *fakeRoleRepo) CountRoleUsers(_ context.Context, roleID uuid.UUID) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.countUsersErr != nil {
		return 0, f.countUsersErr
	}
	role, ok := f.roles[roleID]
	if !ok {
		return 0, apperr.NotFound("El rol no existe")
	}
	return role.UserCount, nil
}

// --- GuardTx ---

func (f *fakeRoleRepo) CountUsers(context.Context) (int64, error) { return 0, nil }

func (f *fakeRoleRepo) InsertRole(_ context.Context, name string) (Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertRoleErr != nil {
		return Role{}, f.insertRoleErr
	}
	for _, role := range f.roles {
		if strings.EqualFold(role.Name, name) {
			return Role{}, apperr.Conflict(messageRoleNameInUse)
		}
	}
	role := Role{ID: uuid.New(), Name: name, CreatedAt: time.Now()}
	f.roles[role.ID] = role
	f.insertedRoles = append(f.insertedRoles, name)
	return role, nil
}

func (f *fakeRoleRepo) UpdateRoleName(_ context.Context, id uuid.UUID, name string) (Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.updateNameErr != nil {
		return Role{}, f.updateNameErr
	}
	for otherID, role := range f.roles {
		if otherID != id && strings.EqualFold(role.Name, name) {
			return Role{}, apperr.Conflict(messageRoleNameInUse)
		}
	}
	role, ok := f.roles[id]
	if !ok {
		return Role{}, apperr.NotFound("El rol no existe")
	}
	role.Name = name
	f.roles[id] = role
	f.renamed = append(f.renamed, name)
	return role, nil
}

func (f *fakeRoleRepo) DeleteRolePermissions(_ context.Context, roleID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.deletePermsErr != nil {
		return f.deletePermsErr
	}
	f.deletedPerms = append(f.deletedPerms, roleID)
	role := f.roles[roleID]
	role.Permissions = nil
	f.roles[roleID] = role
	return nil
}

func (f *fakeRoleRepo) InsertRolePermission(_ context.Context, roleID, permissionID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertPermErr != nil {
		return f.insertPermErr
	}
	f.insertedPerms = append(f.insertedPerms, permissionID)
	role := f.roles[roleID]
	for _, permission := range f.permissions {
		if permission.ID == permissionID {
			role.Permissions = append(role.Permissions, permission.Code)
		}
	}
	f.roles[roleID] = role
	return nil
}

func (f *fakeRoleRepo) DeleteRole(_ context.Context, id uuid.UUID) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.deleteRoleErr != nil {
		return 0, f.deleteRoleErr
	}
	if _, ok := f.roles[id]; !ok {
		return 0, nil
	}
	delete(f.roles, id)
	f.deletedRoles = append(f.deletedRoles, id)
	return 1, nil
}

func (f *fakeRoleRepo) InsertAdminAction(_ context.Context, action audit.Action) (AdminAction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.actionErr != nil {
		return AdminAction{}, f.actionErr
	}
	f.actions = append(f.actions, action)
	return AdminAction{ID: uuid.New(), Action: action.Code}, nil
}

// Stubs de la parte de cuentas de GuardTx: la gestión de roles no las usa.
func (f *fakeRoleRepo) InsertUser(context.Context, NewUser) (User, error) { return User{}, nil }
func (f *fakeRoleRepo) UpdateUser(context.Context, UserUpdate) (User, error) {
	return User{}, nil
}
func (f *fakeRoleRepo) GetUserByID(context.Context, uuid.UUID) (User, error) { return User{}, nil }
func (f *fakeRoleRepo) UpdateUserPassword(context.Context, uuid.UUID, string) error {
	return nil
}
func (f *fakeRoleRepo) SetUserMustChangePassword(context.Context, uuid.UUID, bool) error {
	return nil
}

// --- Helpers ---

func newTestRoleService(repo RoleRepository, recorder ActionRecorder) *roleService {
	return NewRoleService(RoleServiceDeps{Repository: repo, Audit: recorder})
}

func lastRoleAction(t *testing.T, repo *fakeRoleRepo) audit.Action {
	t.Helper()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.actions) == 0 {
		t.Fatal("no se registró ninguna acción en la transacción")
	}
	return repo.actions[len(repo.actions)-1]
}

func validRoleCreateInput(codes ...string) RoleCreateInput {
	return RoleCreateInput{Name: "Contenido", Permissions: codes}
}

// --- T235: CreateRole + catálogo ---

func TestCreateRoleCombinesPermissionsFreely(t *testing.T) {
	repo := newFakeRoleRepo()
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)
	actor := uuid.New()

	item, err := svc.CreateRole(context.Background(), actor, validRoleCreateInput("eventos", "actividades"))
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if item.Name != "Contenido" || item.UserCount != 0 {
		t.Errorf("RoleItem = %+v", item)
	}
	if len(item.Permissions) != 2 {
		t.Fatalf("permisos = %v, se esperaban dos combinados libremente", item.Permissions)
	}
	if repo.txCalls != 1 || repo.guardCalls != 0 {
		t.Errorf("transacciones: tx=%d guard=%d, se esperaba crear sin guard", repo.txCalls, repo.guardCalls)
	}
	if len(repo.insertedRoles) != 1 || repo.insertedRoles[0] != "Contenido" {
		t.Errorf("roles insertados = %v", repo.insertedRoles)
	}
	if len(repo.insertedPerms) != 2 {
		t.Errorf("permisos asociados = %d, se esperaban 2", len(repo.insertedPerms))
	}

	action := lastRoleAction(t, repo)
	if action.Code != audit.ActionRoleCreate || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba role.create success", action)
	}
	if action.ActorUserID == nil || *action.ActorUserID != actor {
		t.Errorf("actor = %v", action.ActorUserID)
	}
	if action.TargetKind != audit.TargetRole || action.TargetRoleID == nil || action.TargetRoleID.String() != item.ID {
		t.Errorf("objetivo = %+v, se esperaba el rol creado", action)
	}
	if action.TargetLabel != "Contenido" {
		t.Errorf("etiqueta = %q", action.TargetLabel)
	}
}

func TestCreateRoleNormalizesName(t *testing.T) {
	repo := newFakeRoleRepo()
	svc := newTestRoleService(repo, &fakeActionRecorder{})

	in := validRoleCreateInput("eventos")
	in.Name = "  Contenido   Musical  "
	item, err := svc.CreateRole(context.Background(), uuid.New(), in)
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if item.Name != "Contenido Musical" {
		t.Errorf("nombre = %q, se esperaba el colapso de espacios", item.Name)
	}
}

func TestCreateRoleDuplicateNormalized(t *testing.T) {
	repo := newFakeRoleRepo()
	repo.seedRole("Contenido", 0, "eventos")
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)

	in := validRoleCreateInput("eventos")
	in.Name = "  contenido  "
	_, err := svc.CreateRole(context.Background(), uuid.New(), in)
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Message != messageRoleNameInUse {
		t.Errorf("mensaje = %q, se esperaba %q", domainErr.Message, messageRoleNameInUse)
	}
	if len(repo.insertedRoles) != 0 || repo.txCalls != 0 {
		t.Fatal("un nombre duplicado no debe crear nada (SC-011)")
	}
	if action := lastAction(t, recorder); action.Code != audit.ActionRoleCreate || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba role.create failure", action)
	}
}

func TestCreateRoleRequiresAtLeastOnePermission(t *testing.T) {
	for _, codes := range [][]string{nil, {}, {"  "}} {
		repo := newFakeRoleRepo()
		recorder := &fakeActionRecorder{}
		svc := newTestRoleService(repo, recorder)

		_, err := svc.CreateRole(context.Background(), uuid.New(), validRoleCreateInput(codes...))
		domainErr := requireKind(t, err, apperr.KindInvalid)
		if domainErr.Message != messageRoleNeedsPermission {
			t.Errorf("mensaje = %q, se esperaba %q", domainErr.Message, messageRoleNeedsPermission)
		}
		if repo.txCalls != 0 || len(repo.insertedRoles) != 0 {
			t.Error("un rol sin permisos no debe escribir nada")
		}
		if action := lastAction(t, recorder); action.Result != audit.ResultFailure {
			t.Errorf("acción = %+v, se esperaba failure", action)
		}
	}
}

func TestCreateRoleUnknownPermission(t *testing.T) {
	repo := newFakeRoleRepo()
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)

	_, err := svc.CreateRole(context.Background(), uuid.New(), validRoleCreateInput("eventos", "no-existe"))
	domainErr := requireKind(t, err, apperr.KindInvalid)
	missing, ok := domainErr.Details["permissions"].([]string)
	if !ok || len(missing) != 1 || missing[0] != "no-existe" {
		t.Errorf("details = %v, se esperaba el permiso inexistente", domainErr.Details)
	}
	if repo.txCalls != 0 {
		t.Error("un permiso inexistente no debe abrir transacción")
	}
}

func TestCreateRoleValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*RoleCreateInput)
		field  string
	}{
		{"sin nombre", func(in *RoleCreateInput) { in.Name = "" }, "name"},
		{"nombre en blanco", func(in *RoleCreateInput) { in.Name = "   " }, "name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRoleRepo()
			svc := newTestRoleService(repo, &fakeActionRecorder{})
			in := validRoleCreateInput("eventos")
			tt.mutate(&in)

			_, err := svc.CreateRole(context.Background(), uuid.New(), in)
			domainErr := requireKind(t, err, apperr.KindInvalid)
			if _, ok := domainErr.Details[tt.field]; !ok {
				t.Errorf("details = %v, se esperaba %q", domainErr.Details, tt.field)
			}
			if repo.txCalls != 0 {
				t.Error("unos datos inválidos no deben escribir nada")
			}
		})
	}
}

func TestListPermissionsReturnsCatalog(t *testing.T) {
	repo := newFakeRoleRepo()
	svc := newTestRoleService(repo, &fakeActionRecorder{})

	list, err := svc.ListPermissions(context.Background())
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(list.Items) != 9 {
		t.Fatalf("catálogo = %d permisos, se esperaban 9 (FR-015)", len(list.Items))
	}
	if list.Items[0].Code == "" || list.Items[0].Label == "" {
		t.Errorf("una entrada del catálogo va sin código o etiqueta: %+v", list.Items[0])
	}
}

func TestListRolesAndGetRole(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Contenido", 3, "eventos", "actividades")
	svc := newTestRoleService(repo, &fakeActionRecorder{})

	list, err := svc.ListRoles(context.Background(), paginate.Params{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if list.Total != 1 || len(list.Items) != 1 || list.Limit != 20 {
		t.Fatalf("sobre = %+v", list)
	}
	if list.Items[0].UserCount != 3 || len(list.Items[0].Permissions) != 2 {
		t.Errorf("RoleItem = %+v", list.Items[0])
	}

	item, err := svc.GetRole(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("GetRole: %v", err)
	}
	if item.ID != role.ID.String() {
		t.Errorf("GetRole = %+v", item)
	}

	_, err = svc.GetRole(context.Background(), uuid.New())
	assertKind(t, err, apperr.KindNotFound)
}

// --- T236: UpdateRole + DeleteRole ---

func TestUpdateRoleReplacesPermissions(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Contenido", 2, "eventos")
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)
	actor := uuid.New()

	item, err := svc.UpdateRole(context.Background(), actor, role.ID, RoleUpdateInput{
		Permissions: []string{"actividades", "medios"},
	})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if len(item.Permissions) != 2 || item.UserCount != 2 {
		t.Errorf("RoleItem = %+v", item)
	}
	if len(repo.renamed) != 0 {
		t.Error("cambiar permisos no debe tocar el nombre")
	}
	if len(repo.deletedPerms) != 1 || len(repo.insertedPerms) != 2 {
		t.Errorf("reemplazo de permisos: borrados=%d insertados=%d", len(repo.deletedPerms), len(repo.insertedPerms))
	}
	if repo.guardCalls != 1 {
		t.Errorf("transacciones del guard = %d, se esperaba 1", repo.guardCalls)
	}
	if action := lastRoleAction(t, repo); action.Code != audit.ActionRoleUpdate || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba role.update success", action)
	}
}

func TestUpdateRoleRenamesKeepingPermissions(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Contenido", 1, "eventos")
	svc := newTestRoleService(repo, &fakeActionRecorder{})

	item, err := svc.UpdateRole(context.Background(), uuid.New(), role.ID, RoleUpdateInput{
		Name: "  Contenido   General ",
	})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if item.Name != "Contenido General" {
		t.Errorf("nombre = %q", item.Name)
	}
	if len(item.Permissions) != 1 || item.Permissions[0] != "eventos" {
		t.Errorf("renombrar debe conservar los permisos: %v", item.Permissions)
	}
	if len(repo.renamed) != 1 || repo.renamed[0] != "Contenido General" {
		t.Errorf("renombrados = %v", repo.renamed)
	}
	if len(repo.deletedPerms) != 0 {
		t.Error("renombrar no debe tocar los permisos")
	}
}

func TestUpdateRoleRejectsEmptyPermissions(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Contenido", 0, "eventos")
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)

	_, err := svc.UpdateRole(context.Background(), uuid.New(), role.ID, RoleUpdateInput{Permissions: []string{}})
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if domainErr.Message != messageRoleNeedsPermission {
		t.Errorf("mensaje = %q", domainErr.Message)
	}
	if repo.guardCalls != 0 {
		t.Error("dejar el rol sin permisos no debe abrir la transacción")
	}
}

func TestUpdateRoleNoChanges(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Contenido", 0, "eventos")
	svc := newTestRoleService(repo, &fakeActionRecorder{})

	_, err := svc.UpdateRole(context.Background(), uuid.New(), role.ID, RoleUpdateInput{})
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if domainErr.Message != messageNoChanges {
		t.Errorf("mensaje = %q", domainErr.Message)
	}
}

func TestUpdateRoleDuplicateName(t *testing.T) {
	repo := newFakeRoleRepo()
	repo.seedRole("Contenido", 0, "eventos")
	other := repo.seedRole("Editor", 0, "noticias")
	svc := newTestRoleService(repo, &fakeActionRecorder{})

	_, err := svc.UpdateRole(context.Background(), uuid.New(), other.ID, RoleUpdateInput{Name: "contenido"})
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Message != messageRoleNameInUse {
		t.Errorf("mensaje = %q", domainErr.Message)
	}
	if repo.guardCalls != 0 {
		t.Error("un nombre duplicado no debe abrir la transacción")
	}
}

// TestUpdateRoleAdminGuard simula el 409 del guard anti-bloqueo: quitar
// `admin_usuarios_roles` al único rol que lo tiene no se completa y el mensaje
// con details.reason=admin_required se propaga tal cual (FR-008).
func TestUpdateRoleAdminGuard(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Administrador", 1, "admin_usuarios_roles", "eventos")
	repo.guardErr = apperr.Conflict(
		"No se puede dejar el panel sin administración",
		apperr.WithDetails(map[string]any{"reason": "admin_required"}),
	)
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)

	_, err := svc.UpdateRole(context.Background(), uuid.New(), role.ID, RoleUpdateInput{
		Permissions: []string{"eventos"},
	})
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Details["reason"] != "admin_required" {
		t.Errorf("details = %v, se esperaba admin_required", domainErr.Details)
	}
	if action := lastAction(t, recorder); action.Code != audit.ActionRoleUpdate || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba role.update failure", action)
	}
}

func TestUpdateRoleNotFound(t *testing.T) {
	repo := newFakeRoleRepo()
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)

	_, err := svc.UpdateRole(context.Background(), uuid.New(), uuid.New(), RoleUpdateInput{Name: "X"})
	assertKind(t, err, apperr.KindNotFound)
	if action := lastAction(t, recorder); action.Result != audit.ResultFailure {
		t.Errorf("un rol inexistente debe dejar fallo: %+v", action)
	}
}

func TestDeleteRoleUnused(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Temporal", 0, "eventos")
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)
	actor := uuid.New()

	if err := svc.DeleteRole(context.Background(), actor, role.ID); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
	if repo.guardCalls != 1 {
		t.Errorf("transacciones del guard = %d, se esperaba 1", repo.guardCalls)
	}
	if len(repo.deletedRoles) != 1 || repo.deletedRoles[0] != role.ID {
		t.Errorf("roles eliminados = %v", repo.deletedRoles)
	}

	action := lastRoleAction(t, repo)
	if action.Code != audit.ActionRoleDelete || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba role.delete success", action)
	}
	if action.TargetLabel != "Temporal" {
		t.Errorf("etiqueta = %q, debe conservar el nombre (FR-025)", action.TargetLabel)
	}
	if action.TargetRoleID == nil || *action.TargetRoleID != role.ID {
		t.Errorf("target_role_id = %v, se esperaba el rol eliminado", action.TargetRoleID)
	}
}

func TestDeleteRoleInUse(t *testing.T) {
	repo := newFakeRoleRepo()
	role := repo.seedRole("Contenido", 2, "eventos")
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)

	err := svc.DeleteRole(context.Background(), uuid.New(), role.ID)
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Message != messageRoleInUse {
		t.Errorf("mensaje = %q", domainErr.Message)
	}
	if count, ok := domainErr.Details["userCount"]; !ok || count != int64(2) {
		t.Errorf("details = %v, se esperaba userCount=2", domainErr.Details)
	}
	if repo.guardCalls != 0 || len(repo.deletedRoles) != 0 {
		t.Fatal("un rol en uso no debe eliminarse (FR-017)")
	}
	if action := lastAction(t, recorder); action.Code != audit.ActionRoleDelete || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba role.delete failure", action)
	}
}

func TestDeleteRoleNotFound(t *testing.T) {
	repo := newFakeRoleRepo()
	recorder := &fakeActionRecorder{}
	svc := newTestRoleService(repo, recorder)

	err := svc.DeleteRole(context.Background(), uuid.New(), uuid.New())
	assertKind(t, err, apperr.KindNotFound)
	if action := lastAction(t, recorder); action.Result != audit.ResultFailure {
		t.Errorf("un rol inexistente debe dejar fallo: %+v", action)
	}
}
