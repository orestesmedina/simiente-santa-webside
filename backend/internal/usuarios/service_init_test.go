package usuarios

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/password"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests unitarios de service_init.go con un repositorio falso. No tocan
// PostgreSQL: el fake implementa a la vez el puerto InitRepository y la vista
// transaccional GuardTx, y registra todo lo que el service intenta escribir para
// comprobar que la comprobación de "sin cuentas" y las escrituras ocurren en la
// MISMA transacción del guard.

// fakeInitRepo implementa InitRepository (WithAdminGuard) y GuardTx.
type fakeInitRepo struct {
	mu sync.Mutex

	total      int64
	countErr   error
	catalog    []Permission
	catalogErr error

	guardCalls int

	insertRoleCalls []string
	insertRoleErr   error

	rolePermissions []uuid.UUID
	insertPermErr   error

	insertedUser    *NewUser
	insertUserErr   error
	insertedAction  *audit.Action
	insertActionErr error
}

func (f *fakeInitRepo) WithAdminGuard(_ context.Context, mutate func(tx GuardTx) error) error {
	f.mu.Lock()
	f.guardCalls++
	f.mu.Unlock()
	return mutate(f)
}

func (f *fakeInitRepo) CountUsers(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.total, f.countErr
}

func (f *fakeInitRepo) InsertRole(_ context.Context, name string) (Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertRoleErr != nil {
		return Role{}, f.insertRoleErr
	}
	f.insertRoleCalls = append(f.insertRoleCalls, name)
	return Role{ID: uuid.New(), Name: name}, nil
}

func (f *fakeInitRepo) ListPermissions(context.Context) ([]Permission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.catalog, f.catalogErr
}

func (f *fakeInitRepo) InsertRolePermission(_ context.Context, _, permissionID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertPermErr != nil {
		return f.insertPermErr
	}
	f.rolePermissions = append(f.rolePermissions, permissionID)
	return nil
}

func (f *fakeInitRepo) InsertUser(_ context.Context, user NewUser) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertUserErr != nil {
		return User{}, f.insertUserErr
	}
	saved := user
	f.insertedUser = &saved
	return User{
		ID:                 uuid.New(),
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Phone:              user.Phone,
		MustChangePassword: user.MustChangePassword,
		IsActive:           user.IsActive,
		RoleID:             user.RoleID,
		CreatedAt:          time.Now(),
	}, nil
}

func (f *fakeInitRepo) InsertAdminAction(_ context.Context, action audit.Action) (AdminAction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertActionErr != nil {
		return AdminAction{}, f.insertActionErr
	}
	saved := action
	f.insertedAction = &saved
	return AdminAction{ID: uuid.New(), Action: action.Code}, nil
}

// UpdateUser completa GuardTx; la inicialización no la usa.
func (f *fakeInitRepo) UpdateUser(context.Context, UserUpdate) (User, error) {
	return User{}, nil
}

// GetUserByID, UpdateUserPassword y SetUserMustChangePassword completan GuardTx;
// la inicialización no las usa (T232/T233).
func (f *fakeInitRepo) GetUserByID(context.Context, uuid.UUID) (User, error) {
	return User{}, nil
}

func (f *fakeInitRepo) UpdateUserPassword(context.Context, uuid.UUID, string) error { return nil }

func (f *fakeInitRepo) SetUserMustChangePassword(context.Context, uuid.UUID, bool) error { return nil }

// sampleCatalog devuelve los nueve permisos del catálogo fijo (FR-015).
func sampleCatalog() []Permission {
	codes := []string{
		"portada", "eventos", "actividades", "grupos", "ministerios",
		"donaciones", "noticias", "medios", "admin_usuarios_roles",
	}
	permissions := make([]Permission, 0, len(codes))
	for _, code := range codes {
		permissions = append(permissions, Permission{ID: uuid.New(), Code: code, Label: code})
	}
	return permissions
}

func validInitializeInput() InitializeInput {
	return InitializeInput{
		FirstName: "Ana",
		LastName:  "Responsable",
		Email:     "ana@ejemplo.com",
		Phone:     "+34 612 345 678",
		Password:  "Semilla.2026",
	}
}

func newTestInitService(t *testing.T, repo InitRepository) *initService {
	t.Helper()
	logger, _ := testutil.NewLogger()
	return NewInitService(InitServiceDeps{Repository: repo, Logger: logger})
}

func TestInitializeCreatesInitialAdmin(t *testing.T) {
	repo := &fakeInitRepo{catalog: sampleCatalog()}
	svc := newTestInitService(t, repo)

	in := validInitializeInput()
	in.FirstName = "  Ana  "
	in.Email = "  Ana@Ejemplo.com "
	item, err := svc.Initialize(context.Background(), in)
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	if item.ID == "" {
		t.Error("el DTO no trae id")
	}
	if item.Email != "ana@ejemplo.com" || item.FirstName != "Ana" {
		t.Errorf("datos sin normalizar: %+v", item)
	}
	if item.RoleName != adminRoleName || !item.IsActive || item.MustChangePassword {
		t.Errorf("el administrador inicial no es el esperado: %+v", item)
	}

	// Rol "Administrador" con TODO el catálogo (9 permisos).
	if len(repo.insertRoleCalls) != 1 || repo.insertRoleCalls[0] != adminRoleName {
		t.Fatalf("roles creados = %v, se esperaba solo %q", repo.insertRoleCalls, adminRoleName)
	}
	if len(repo.rolePermissions) != 9 {
		t.Fatalf("permisos asociados = %d, se esperaban 9", len(repo.rolePermissions))
	}

	// Cuenta activa, sin cambio obligatorio y con la contraseña hasheada.
	if repo.insertedUser == nil {
		t.Fatal("no se creó la cuenta")
	}
	if !repo.insertedUser.IsActive || repo.insertedUser.MustChangePassword {
		t.Errorf("estado de la cuenta inesperado: %+v", repo.insertedUser)
	}
	if repo.insertedUser.PasswordHash == in.Password {
		t.Fatal("la contraseña se guardó en claro")
	}
	if err := password.Verify(repo.insertedUser.PasswordHash, "Semilla.2026"); err != nil {
		t.Fatalf("el hash guardado no verifica la contraseña: %v", err)
	}

	// Registro `user.create` SIN actor (único caso permitido, FR-023).
	if repo.insertedAction == nil {
		t.Fatal("no quedó registro de la inicialización")
	}
	action := repo.insertedAction
	if action.Code != audit.ActionUserCreate || action.ActorUserID != nil {
		t.Errorf("acción = %+v, se esperaba user.create sin actor", action)
	}
	if action.TargetKind != audit.TargetUser || action.Result != audit.ResultSuccess {
		t.Errorf("objetivo/resultado = %+v", action)
	}
	if action.TargetUserID == nil || action.TargetUserID.String() != item.ID {
		t.Errorf("el objetivo no apunta a la cuenta creada: %+v", action.TargetUserID)
	}
	if repo.guardCalls != 1 {
		t.Errorf("se abrieron %d transacciones, se esperaba 1", repo.guardCalls)
	}
}

func TestInitializeRejectsWhenAccountsExist(t *testing.T) {
	repo := &fakeInitRepo{total: 1, catalog: sampleCatalog()}
	svc := newTestInitService(t, repo)

	_, err := svc.Initialize(context.Background(), validInitializeInput())
	assertKind(t, err, apperr.KindConflict)
	if !strings.Contains(err.Error(), messageAlreadyInitialized) {
		t.Errorf("mensaje = %q, se esperaba %q", err.Error(), messageAlreadyInitialized)
	}

	// Repetirla no crea NADA: ni rol, ni permisos, ni cuenta, ni registro.
	if len(repo.insertRoleCalls) != 0 || len(repo.rolePermissions) != 0 ||
		repo.insertedUser != nil || repo.insertedAction != nil {
		t.Fatalf("una inicialización repetida no debe escribir: %+v", repo)
	}
}

func TestInitializeRejectsInvalidDataWithoutWriting(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*InitializeInput)
		wantCode string
	}{
		{
			name:     "contraseña que no cumple la política",
			mutate:   func(in *InitializeInput) { in.Password = "corta" },
			wantCode: "newPassword",
		},
		{
			name:     "contraseña igual al correo",
			mutate:   func(in *InitializeInput) { in.Password = "ana@ejemplo.com" },
			wantCode: "newPassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeInitRepo{catalog: sampleCatalog()}
			svc := newTestInitService(t, repo)

			in := validInitializeInput()
			tt.mutate(&in)
			_, err := svc.Initialize(context.Background(), in)
			assertKind(t, err, apperr.KindInvalid)

			var domainErr *apperr.Error
			if !errors.As(err, &domainErr) {
				t.Fatalf("no es un *apperr.Error: %v", err)
			}
			if _, ok := domainErr.Details[tt.wantCode]; !ok {
				t.Errorf("details = %v, se esperaba %q", domainErr.Details, tt.wantCode)
			}
			if repo.guardCalls != 0 {
				t.Error("unos datos inválidos no deben abrir la transacción")
			}
			if len(repo.insertRoleCalls) != 0 || repo.insertedUser != nil {
				t.Error("unos datos inválidos no deben crear nada")
			}
		})
	}
}
