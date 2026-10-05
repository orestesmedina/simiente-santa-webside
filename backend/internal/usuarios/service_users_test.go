package usuarios

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/password"
	"simiente-santa/backend/internal/platform/session"
)

// Tests unitarios de service_users.go (T231–T233) con fakes. No tocan
// PostgreSQL ni Redis: el fake implementa a la vez UserRepository y GuardTx y
// observa cada escritura; el Store de sesiones falso permite comprobar que
// desactivar o restablecer revoca las sesiones de la cuenta.

// --- Fake del repositorio ---

type fakeUserRepo struct {
	mu sync.Mutex

	users map[uuid.UUID]User
	roles map[uuid.UUID]Role

	getUserErr  error
	emailErr    error
	roleErr     error
	listErr     error
	countErr    error
	insertErr   error
	updateErr   error
	passwordErr error
	mustErr     error
	actionErr   error
	guardErr    error

	guardCalls int
	inserted   []NewUser
	updated    []UserUpdate
	actions    []audit.Action
	passwords  map[uuid.UUID]string
	mustFlags  map[uuid.UUID]bool
	listParams []paginate.Params
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:     map[uuid.UUID]User{},
		roles:     map[uuid.UUID]Role{},
		passwords: map[uuid.UUID]string{},
		mustFlags: map[uuid.UUID]bool{},
	}
}

func (f *fakeUserRepo) WithAdminGuard(_ context.Context, mutate func(tx GuardTx) error) error {
	f.mu.Lock()
	f.guardCalls++
	guardErr := f.guardErr
	f.mu.Unlock()

	if err := mutate(f); err != nil {
		return err
	}
	return guardErr
}

func (f *fakeUserRepo) GetUserByID(_ context.Context, id uuid.UUID) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getUserErr != nil {
		return User{}, f.getUserErr
	}
	user, ok := f.users[id]
	if !ok {
		return User{}, apperr.NotFound("La cuenta no existe")
	}
	return user, nil
}

func (f *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.emailErr != nil {
		return User{}, f.emailErr
	}
	for _, user := range f.users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, apperr.NotFound("La cuenta no existe")
}

func (f *fakeUserRepo) GetRoleByID(_ context.Context, id uuid.UUID) (Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.roleErr != nil {
		return Role{}, f.roleErr
	}
	role, ok := f.roles[id]
	if !ok {
		return Role{}, apperr.NotFound("El rol no existe")
	}
	return role, nil
}

func (f *fakeUserRepo) ListUsers(_ context.Context, params paginate.Params) ([]User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.listErr != nil {
		return nil, f.listErr
	}
	f.listParams = append(f.listParams, params)
	all := make([]User, 0, len(f.users))
	for _, user := range f.users {
		all = append(all, user)
	}
	if params.Offset > len(all) {
		return []User{}, nil
	}
	all = all[params.Offset:]
	if params.Limit < len(all) {
		all = all[:params.Limit]
	}
	return all, nil
}

func (f *fakeUserRepo) CountUsers(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.countErr != nil {
		return 0, f.countErr
	}
	return int64(len(f.users)), nil
}

func (f *fakeUserRepo) InsertUser(_ context.Context, in NewUser) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertErr != nil {
		return User{}, f.insertErr
	}
	for _, user := range f.users {
		if user.Email == in.Email {
			return User{}, apperr.Conflict("Ya existe una cuenta con ese correo electrónico")
		}
	}
	saved := in
	f.inserted = append(f.inserted, saved)
	user := User{
		ID:                 uuid.New(),
		Email:              in.Email,
		FirstName:          in.FirstName,
		LastName:           in.LastName,
		Phone:              in.Phone,
		MustChangePassword: in.MustChangePassword,
		IsActive:           in.IsActive,
		RoleID:             in.RoleID,
	}
	f.users[user.ID] = user
	return user, nil
}

func (f *fakeUserRepo) UpdateUser(_ context.Context, in UserUpdate) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.updateErr != nil {
		return User{}, f.updateErr
	}
	for _, user := range f.users {
		if user.Email == in.Email && user.ID != in.ID {
			return User{}, apperr.Conflict("Ya existe una cuenta con ese correo electrónico")
		}
	}
	current, ok := f.users[in.ID]
	if !ok {
		return User{}, apperr.NotFound("La cuenta no existe")
	}
	current.Email = in.Email
	current.FirstName = in.FirstName
	current.LastName = in.LastName
	current.Phone = in.Phone
	current.RoleID = in.RoleID
	current.IsActive = in.IsActive
	f.users[in.ID] = current
	f.updated = append(f.updated, in)
	return current, nil
}

func (f *fakeUserRepo) InsertAdminAction(_ context.Context, action audit.Action) (AdminAction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.actionErr != nil {
		return AdminAction{}, f.actionErr
	}
	f.actions = append(f.actions, action)
	return AdminAction{ID: uuid.New(), Action: action.Code}, nil
}

// InsertRole, ListPermissions e InsertRolePermission completan GuardTx; la
// gestión de cuentas no las usa (son de la inicialización y de los roles).
func (f *fakeUserRepo) InsertRole(context.Context, string) (Role, error) { return Role{}, nil }

func (f *fakeUserRepo) ListPermissions(context.Context) ([]Permission, error) { return nil, nil }

func (f *fakeUserRepo) InsertRolePermission(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (f *fakeUserRepo) UpdateUserPassword(_ context.Context, id uuid.UUID, hash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.passwordErr != nil {
		return f.passwordErr
	}
	f.passwords[id] = hash
	return nil
}

func (f *fakeUserRepo) SetUserMustChangePassword(_ context.Context, id uuid.UUID, must bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mustErr != nil {
		return f.mustErr
	}
	f.mustFlags[id] = must
	return nil
}

// UpdateRoleName, DeleteRolePermissions y DeleteRole completan GuardTx; la
// gestión de cuentas no las usa (son de la gestión de roles, T236).
func (f *fakeUserRepo) UpdateRoleName(context.Context, uuid.UUID, string) (Role, error) {
	return Role{}, nil
}

func (f *fakeUserRepo) DeleteRolePermissions(context.Context, uuid.UUID) error { return nil }

func (f *fakeUserRepo) DeleteRole(context.Context, uuid.UUID) (int64, error) { return 0, nil }

// --- Fakes de dependencias ---

type fakeActionRecorder struct {
	mu      sync.Mutex
	actions []audit.Action
}

func (f *fakeActionRecorder) RecordActionBestEffort(_ context.Context, action audit.Action) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.actions = append(f.actions, action)
}

func (f *fakeActionRecorder) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.actions)
}

type fakeUserSessionStore struct {
	mu            sync.Mutex
	revokedUsers  []uuid.UUID
	revokeUserErr error
}

func (f *fakeUserSessionStore) Create(context.Context, uuid.UUID) (string, error) {
	return "token", nil
}

func (f *fakeUserSessionStore) Resolve(context.Context, string) (session.Session, error) {
	return session.Session{}, session.ErrSessionNotFound
}

func (f *fakeUserSessionStore) Revoke(context.Context, string) error { return nil }

func (f *fakeUserSessionStore) RevokeUser(_ context.Context, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.revokeUserErr != nil {
		return f.revokeUserErr
	}
	f.revokedUsers = append(f.revokedUsers, id)
	return nil
}

func (f *fakeUserSessionStore) RevokeUserExcept(context.Context, uuid.UUID, string) error {
	return nil
}

func (f *fakeUserSessionStore) revoked() []uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]uuid.UUID(nil), f.revokedUsers...)
}

// --- Helpers ---

func newTestUserService(repo *fakeUserRepo, sessions session.Store, recorder ActionRecorder) *userService {
	return NewUserService(UserServiceDeps{
		Repository: repo,
		Sessions:   sessions,
		Audit:      recorder,
	})
}

func testRole(name string, permissions ...string) Role {
	return Role{ID: uuid.New(), Name: name, Permissions: permissions}
}

func testUser(role Role, email string, active bool) User {
	return User{
		ID:                 uuid.New(),
		Email:              email,
		FirstName:          "Ana",
		LastName:           "Pérez",
		Phone:              "612345678",
		MustChangePassword: false,
		IsActive:           active,
		RoleID:             role.ID,
		RoleName:           role.Name,
	}
}

func validCreateInput(roleID uuid.UUID) CreateUserInput {
	return CreateUserInput{
		FirstName: "Carlos",
		LastName:  "Ayudante",
		Email:     "carlos@ejemplo.com",
		Phone:     "612 345 678",
		RoleID:    roleID.String(),
		Password:  "Cambio.2026",
	}
}

func lastAction(t *testing.T, recorder *fakeActionRecorder) audit.Action {
	t.Helper()
	if recorder.count() == 0 {
		t.Fatal("no se registró ninguna acción")
	}
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return recorder.actions[len(recorder.actions)-1]
}

// lastRepoAction devuelve la última acción escrita DENTRO de la transacción del
// guard (éxito); las de fallo van al ActionRecorder best-effort.
func lastRepoAction(t *testing.T, repo *fakeUserRepo) audit.Action {
	t.Helper()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.actions) == 0 {
		t.Fatal("no se registró ninguna acción en la transacción")
	}
	return repo.actions[len(repo.actions)-1]
}

// --- T231: CreateUser ---

func TestCreateUserSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	recorder := &fakeActionRecorder{}
	sessions := &fakeUserSessionStore{}
	svc := newTestUserService(repo, sessions, recorder)
	actor := uuid.New()

	in := validCreateInput(role.ID)
	in.Email = "  Carlos@Ejemplo.COM "
	item, err := svc.CreateUser(context.Background(), actor, in)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if item.Email != "carlos@ejemplo.com" || item.FirstName != "Carlos" {
		t.Errorf("datos sin normalizar: %+v", item)
	}
	if !item.IsActive || !item.MustChangePassword {
		t.Errorf("la cuenta debe nacer activa y con cambio obligatorio: %+v", item)
	}
	if item.RoleName != "Contenido" {
		t.Errorf("rol = %q, se esperaba Contenido", item.RoleName)
	}

	if len(repo.inserted) != 1 {
		t.Fatalf("cuentas insertadas = %d", len(repo.inserted))
	}
	saved := repo.inserted[0]
	if !saved.IsActive || !saved.MustChangePassword {
		t.Errorf("estado guardado inesperado: %+v", saved)
	}
	if saved.PasswordHash == in.Password {
		t.Fatal("la contraseña se guardó en claro")
	}
	if err := password.Verify(saved.PasswordHash, "Cambio.2026"); err != nil {
		t.Fatalf("el hash no verifica la contraseña: %v", err)
	}
	if repo.guardCalls != 1 {
		t.Errorf("transacciones del guard = %d, se esperaba 1", repo.guardCalls)
	}

	action := lastRepoAction(t, repo)
	if action.Code != audit.ActionUserCreate || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba user.create success", action)
	}
	if action.ActorUserID == nil || *action.ActorUserID != actor {
		t.Errorf("actor = %v", action.ActorUserID)
	}
	if action.TargetUserID == nil || action.TargetUserID.String() != item.ID {
		t.Errorf("objetivo = %v, se esperaba %s", action.TargetUserID, item.ID)
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	existing := testUser(role, "carlos@ejemplo.com", true)
	repo.users[existing.ID] = existing
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	in := validCreateInput(role.ID)
	in.Email = "  Carlos@Ejemplo.com  "
	_, err := svc.CreateUser(context.Background(), uuid.New(), in)
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Message != messageEmailInUse {
		t.Errorf("mensaje = %q, se esperaba %q", domainErr.Message, messageEmailInUse)
	}
	if len(repo.inserted) != 0 {
		t.Fatal("un correo duplicado no debe crear nada")
	}
	if action := lastAction(t, recorder); action.Code != audit.ActionUserCreate || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba user.create failure", action)
	}
}

func TestCreateUserValidation(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*CreateUserInput)
		wantDetail string
	}{
		{name: "sin nombre", mutate: func(in *CreateUserInput) { in.FirstName = "" }, wantDetail: "firstName"},
		{name: "sin apellidos", mutate: func(in *CreateUserInput) { in.LastName = "" }, wantDetail: "lastName"},
		{name: "sin correo", mutate: func(in *CreateUserInput) { in.Email = "" }, wantDetail: "email"},
		{name: "sin teléfono", mutate: func(in *CreateUserInput) { in.Phone = "" }, wantDetail: "phone"},
		{name: "teléfono no telefónico", mutate: func(in *CreateUserInput) { in.Phone = "no es un teléfono" }, wantDetail: "phone"},
		{name: "sin rol", mutate: func(in *CreateUserInput) { in.RoleID = "" }, wantDetail: "roleId"},
		{name: "contraseña corta", mutate: func(in *CreateUserInput) { in.Password = "Ab1!" }, wantDetail: "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			role := testRole("Contenido", "eventos")
			repo.roles[role.ID] = role
			recorder := &fakeActionRecorder{}
			svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

			in := validCreateInput(role.ID)
			tt.mutate(&in)
			_, err := svc.CreateUser(context.Background(), uuid.New(), in)
			domainErr := requireKind(t, err, apperr.KindInvalid)
			if _, ok := domainErr.Details[tt.wantDetail]; !ok {
				t.Errorf("details = %v, se esperaba %q", domainErr.Details, tt.wantDetail)
			}
			if len(repo.inserted) != 0 || repo.guardCalls != 0 {
				t.Error("unos datos inválidos no deben escribir nada")
			}
			if action := lastAction(t, recorder); action.Result != audit.ResultFailure {
				t.Errorf("acción = %+v, se esperaba failure", action)
			}
		})
	}
}

func TestCreateUserRoleNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	in := validCreateInput(uuid.New())
	_, err := svc.CreateUser(context.Background(), uuid.New(), in)
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if _, ok := domainErr.Details["roleId"]; !ok {
		t.Errorf("details = %v, se esperaba roleId", domainErr.Details)
	}
	if len(repo.inserted) != 0 || repo.guardCalls != 0 {
		t.Error("un rol inexistente no debe escribir nada")
	}
}

func TestCreateUserPolicyViolation(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	in := validCreateInput(role.ID)
	in.Password = "sinmayusculas1!" // no cumple FR-010
	_, err := svc.CreateUser(context.Background(), uuid.New(), in)
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if _, ok := domainErr.Details["newPassword"]; !ok {
		t.Errorf("details = %v, se esperaba el requisito de la política", domainErr.Details)
	}
	if len(repo.inserted) != 0 {
		t.Fatal("una contraseña fuera de política no debe guardarse")
	}
}

func TestCreateUserResponseNeverContainsPassword(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	svc := newTestUserService(repo, &fakeUserSessionStore{}, &fakeActionRecorder{})

	item, err := svc.CreateUser(context.Background(), uuid.New(), validCreateInput(role.ID))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	payload, _ := json.Marshal(item)
	lower := strings.ToLower(string(payload))
	for _, marker := range []string{"passwordhash", "password_hash", "$2a$", "$2b$", "$2y$"} {
		if strings.Contains(lower, marker) {
			t.Fatalf("el DTO de salida expone credenciales (%q): %s", marker, payload)
		}
	}
	if strings.Contains(string(payload), "Cambio.2026") {
		t.Fatalf("el DTO de salida expone la contraseña: %s", payload)
	}
}

// --- T232: UpdateUser ---

func TestUpdateUserEditsFieldsAndReplacesRole(t *testing.T) {
	repo := newFakeUserRepo()
	oldRole := testRole("Contenido", "eventos")
	newRole := testRole("Ministerios", "ministerios")
	repo.roles[oldRole.ID] = oldRole
	repo.roles[newRole.ID] = newRole
	user := testUser(oldRole, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)
	actor := uuid.New()

	item, err := svc.UpdateUser(context.Background(), actor, user.ID, UpdateUserInput{
		FirstName: "  Carlos  ",
		Phone:     "699 111 222",
		RoleID:    newRole.ID.String(),
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if item.FirstName != "Carlos" || item.Phone != "699 111 222" {
		t.Errorf("datos editados = %+v", item)
	}
	if item.RoleID != newRole.ID.String() || item.RoleName != "Ministerios" {
		t.Errorf("el rol no se reemplazó: %+v", item)
	}
	// El rol se reemplaza, nunca se suma (un solo rol por cuenta, Q4).
	if stored := repo.users[user.ID]; stored.RoleID != newRole.ID {
		t.Errorf("rol almacenado = %v, se esperaba %v", stored.RoleID, newRole.ID)
	}
	if action := lastRepoAction(t, repo); action.Code != audit.ActionUserUpdate || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba user.update success", action)
	}
}

func TestUpdateUserDeactivateRevokesSessionsAndKeepsData(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	recorder := &fakeActionRecorder{}
	sessions := &fakeUserSessionStore{}
	svc := newTestUserService(repo, sessions, recorder)

	inactive := false
	item, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{IsActive: &inactive})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if item.IsActive {
		t.Error("la cuenta debería quedar desactivada")
	}
	if item.Email != user.Email {
		t.Errorf("los datos deben conservarse: %+v", item)
	}
	revoked := sessions.revoked()
	if len(revoked) != 1 || revoked[0] != user.ID {
		t.Fatalf("sesiones revocadas = %v, se esperaba la cuenta", revoked)
	}
	if action := lastRepoAction(t, repo); action.Code != audit.ActionUserDeactivate || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba user.deactivate success", action)
	}
}

func TestUpdateUserReactivateRestoresAccess(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", false)
	repo.users[user.ID] = user
	recorder := &fakeActionRecorder{}
	sessions := &fakeUserSessionStore{}
	svc := newTestUserService(repo, sessions, recorder)

	active := true
	item, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{IsActive: &active})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if !item.IsActive {
		t.Error("la cuenta debería quedar activa")
	}
	if len(sessions.revoked()) != 0 {
		t.Error("reactivar no debe revocar sesiones")
	}
	if action := lastRepoAction(t, repo); action.Code != audit.ActionUserActivate || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba user.activate success", action)
	}
}

// TestUpdateUserLastAdminGuard demuestra que el 409 del guard se propaga tal
// cual (con details.reason=admin_required) y queda registrado como fallo.
func TestUpdateUserLastAdminGuard(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Administrador", "admin_usuarios_roles")
	repo.roles[role.ID] = role
	user := testUser(role, "ana@ejemplo.com", true)
	repo.users[user.ID] = user
	repo.guardErr = apperr.Conflict(
		"No se puede dejar el panel sin administración",
		apperr.WithDetails(map[string]any{"reason": "admin_required"}),
	)
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	inactive := false
	_, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{IsActive: &inactive})
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Details["reason"] != "admin_required" {
		t.Errorf("details = %v, se esperaba admin_required", domainErr.Details)
	}
	if action := lastAction(t, recorder); action.Code != audit.ActionUserDeactivate || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba user.deactivate failure", action)
	}
}

func TestUpdateUserChangeRoleToNonAdminGuard(t *testing.T) {
	repo := newFakeUserRepo()
	adminRole := testRole("Administrador", "admin_usuarios_roles")
	other := testRole("Contenido", "eventos")
	repo.roles[adminRole.ID] = adminRole
	repo.roles[other.ID] = other
	user := testUser(adminRole, "ana@ejemplo.com", true)
	repo.users[user.ID] = user
	repo.guardErr = apperr.Conflict(
		"No se puede dejar el panel sin administración",
		apperr.WithDetails(map[string]any{"reason": "admin_required"}),
	)
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	_, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{RoleID: other.ID.String()})
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Details["reason"] != "admin_required" {
		t.Errorf("details = %v", domainErr.Details)
	}
	if action := lastAction(t, recorder); action.Code != audit.ActionUserUpdate || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba user.update failure", action)
	}
}

func TestUpdateUserNoChanges(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	svc := newTestUserService(repo, &fakeUserSessionStore{}, &fakeActionRecorder{})

	_, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{})
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if domainErr.Message != messageNoChanges {
		t.Errorf("mensaje = %q", domainErr.Message)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	_, err := svc.UpdateUser(context.Background(), uuid.New(), uuid.New(), UpdateUserInput{FirstName: "X"})
	requireKind(t, err, apperr.KindNotFound)
	if action := lastAction(t, recorder); action.Result != audit.ResultFailure {
		t.Errorf("una cuenta inexistente debe dejar fallo: %+v", action)
	}
}

func TestUpdateUserDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	other := testUser(role, "luis@ejemplo.com", true)
	repo.users[user.ID] = user
	repo.users[other.ID] = other
	svc := newTestUserService(repo, &fakeUserSessionStore{}, &fakeActionRecorder{})

	_, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{Email: "Luis@Ejemplo.com"})
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Message != messageEmailInUse {
		t.Errorf("mensaje = %q, se esperaba %q", domainErr.Message, messageEmailInUse)
	}
}

func TestUpdateUserRoleNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	svc := newTestUserService(repo, &fakeUserSessionStore{}, &fakeActionRecorder{})

	_, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{RoleID: uuid.New().String()})
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if _, ok := domainErr.Details["roleId"]; !ok {
		t.Errorf("details = %v, se esperaba roleId", domainErr.Details)
	}
}

// TestUpdateUserRevokeFailureIsBestEffort verifica que un fallo de Redis al
// revocar no revierte la baja (authn ya rechaza la cuenta inactiva).
func TestUpdateUserRevokeFailureIsBestEffort(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	sessions := &fakeUserSessionStore{revokeUserErr: errors.New("redis caído")}
	svc := newTestUserService(repo, sessions, &fakeActionRecorder{})

	inactive := false
	item, err := svc.UpdateUser(context.Background(), uuid.New(), user.ID, UpdateUserInput{IsActive: &inactive})
	if err != nil {
		t.Fatalf("un fallo best-effort no debe cambiar la respuesta: %v", err)
	}
	if item.IsActive {
		t.Error("la cuenta debe quedar desactivada igualmente")
	}
}

// --- T233: ResetUserPassword ---

func TestResetUserPasswordSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	recorder := &fakeActionRecorder{}
	sessions := &fakeUserSessionStore{}
	svc := newTestUserService(repo, sessions, recorder)
	actor := uuid.New()

	err := svc.ResetUserPassword(context.Background(), actor, user.ID, ResetPasswordInput{Password: "Nueva9#Aa"})
	if err != nil {
		t.Fatalf("ResetUserPassword: %v", err)
	}

	hash, ok := repo.passwords[user.ID]
	if !ok {
		t.Fatal("no se guardó el hash nuevo")
	}
	if err := password.Verify(hash, "Nueva9#Aa"); err != nil {
		t.Fatalf("el hash no corresponde a la contraseña nueva: %v", err)
	}
	if !repo.mustFlags[user.ID] {
		t.Errorf("mustChangePassword debe quedar en true: %v", repo.mustFlags)
	}
	revoked := sessions.revoked()
	if len(revoked) != 1 || revoked[0] != user.ID {
		t.Fatalf("sesiones revocadas = %v, se esperaba la cuenta (R17)", revoked)
	}

	action := lastRepoAction(t, repo)
	if action.Code != audit.ActionUserPasswordReset || action.Result != audit.ResultSuccess {
		t.Errorf("acción = %+v, se esperaba user.password_reset success", action)
	}
	if action.ActorUserID == nil || *action.ActorUserID != actor {
		t.Errorf("actor = %v", action.ActorUserID)
	}
	if action.TargetUserID == nil || *action.TargetUserID != user.ID {
		t.Errorf("objetivo = %v", action.TargetUserID)
	}
	// El registro no puede contener ningún valor de contraseña (FR-026).
	for _, recorded := range recorder.actions {
		if strings.Contains(recorded.TargetLabel, "Nueva9#Aa") {
			t.Fatalf("el registro contiene una contraseña: %+v", recorded)
		}
	}
}

func TestResetUserPasswordPolicyViolation(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	err := svc.ResetUserPassword(context.Background(), uuid.New(), user.ID, ResetPasswordInput{Password: "sinmayusculas1!"})
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if _, ok := domainErr.Details["newPassword"]; !ok {
		t.Errorf("details = %v, se esperaba el requisito", domainErr.Details)
	}
	if len(repo.passwords) != 0 {
		t.Fatal("una contraseña fuera de política no debe guardarse")
	}
	if action := lastAction(t, recorder); action.Code != audit.ActionUserPasswordReset || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba user.password_reset failure", action)
	}
}

func TestResetUserPasswordNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	recorder := &fakeActionRecorder{}
	svc := newTestUserService(repo, &fakeUserSessionStore{}, recorder)

	err := svc.ResetUserPassword(context.Background(), uuid.New(), uuid.New(), ResetPasswordInput{Password: "Nueva9#Aa"})
	requireKind(t, err, apperr.KindNotFound)
	if action := lastAction(t, recorder); action.Code != audit.ActionUserPasswordReset || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba user.password_reset failure", action)
	}
}

func TestResetUserPasswordRevokeFailureIsBestEffort(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	repo.users[user.ID] = user
	sessions := &fakeUserSessionStore{revokeUserErr: errors.New("redis caído")}
	svc := newTestUserService(repo, sessions, &fakeActionRecorder{})

	if err := svc.ResetUserPassword(context.Background(), uuid.New(), user.ID, ResetPasswordInput{Password: "Nueva9#Aa"}); err != nil {
		t.Fatalf("un fallo best-effort al revocar no debe cambiar la respuesta: %v", err)
	}
	if _, ok := repo.passwords[user.ID]; !ok {
		t.Fatal("la contraseña debe quedar guardada igualmente")
	}
}

// --- Lectura ---

func TestListUsersAndGetUser(t *testing.T) {
	repo := newFakeUserRepo()
	role := testRole("Contenido", "eventos")
	repo.roles[role.ID] = role
	user := testUser(role, "carlos@ejemplo.com", true)
	user.LastLoginAt = nil
	user.LastLoginIP = nil
	repo.users[user.ID] = user
	svc := newTestUserService(repo, &fakeUserSessionStore{}, &fakeActionRecorder{})

	list, err := svc.ListUsers(context.Background(), paginate.Params{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if list.Total != 1 || len(list.Items) != 1 || list.Limit != 20 {
		t.Fatalf("sobre del listado = %+v", list)
	}
	if list.Items[0].LastLoginAt != nil || list.Items[0].LastLoginIP != nil {
		t.Error("una cuenta que nunca entró debe llevar lastLoginAt/lastLoginIp en nil")
	}

	item, err := svc.GetUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if item.ID != user.ID.String() {
		t.Errorf("GetUser = %+v", item)
	}
}
