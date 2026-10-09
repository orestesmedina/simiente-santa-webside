//go:build integration

package usuarios

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/password"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de service_users.go (T231–T233) contra PostgreSQL
// real: el guard anti-bloqueo con el recuento post-mutación, el registro en
// `admin_actions`, la unicidad del correo normalizado y la verificación de la
// contraseña restablecida. Las sesiones se sustituyen por un Store falso (Redis
// no es necesario para comprobar la revocación por cuenta).

func integrationUserService(t *testing.T, repo *repository, sessions session.Store) *userService {
	t.Helper()
	logger, _ := testutil.NewLogger()
	return NewUserService(UserServiceDeps{
		Repository: repo,
		Sessions:   sessions,
		Audit:      NewAuditService(repo, logger),
		Logger:     logger,
	})
}

// integrationAdminRole crea el rol "Administrador" con el permiso de gestión de
// usuarios y roles. El nombre es único, así que se crea UNA vez por prueba.
func integrationAdminRole(t *testing.T, repo *repository) Role {
	t.Helper()
	return mustInsertRole(t, repo, "Administrador", PermissionAdminUsersRoles)
}

// hasAdminAction indica si en el historial hay una acción con el código y
// resultado indicados.
func hasAdminAction(t *testing.T, repo *repository, code string, result audit.Result) bool {
	t.Helper()
	actions, err := repo.ListAdminActions(context.Background(), AuditFilter{Limit: 100})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	for _, action := range actions {
		if action.Action == code && action.Result == result {
			return true
		}
	}
	return false
}

func TestIntegrationCreateUserWithAudit(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	adminRole := integrationAdminRole(t, repo)
	admin := mustInsertUser(t, repo, "ana@ejemplo.com", adminRole.ID, true)
	content := mustInsertRole(t, repo, "Contenido", "eventos")
	svc := integrationUserService(t, repo, &fakeUserSessionStore{})

	in := CreateUserInput{
		FirstName: "Carlos",
		LastName:  "Ayudante",
		Email:     "  Carlos@Ejemplo.COM ",
		Phone:     "612 345 678",
		RoleID:    content.ID.String(),
		Password:  "Cambio.2026",
	}
	item, err := svc.CreateUser(ctx, admin.ID, in)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if item.Email != "carlos@ejemplo.com" || !item.IsActive || !item.MustChangePassword {
		t.Fatalf("cuenta creada inesperada: %+v", item)
	}
	if item.RoleName != "Contenido" {
		t.Fatalf("rol = %q", item.RoleName)
	}

	auth, err := repo.GetUserAuthByEmail(ctx, "carlos@ejemplo.com")
	if err != nil {
		t.Fatalf("GetUserAuthByEmail: %v", err)
	}
	if !auth.IsActive {
		t.Fatal("la cuenta debe nacer activa")
	}
	if err := password.Verify(auth.PasswordHash, "Cambio.2026"); err != nil {
		t.Fatalf("el hash guardado no verifica la contraseña: %v", err)
	}
	if !hasAdminAction(t, repo, audit.ActionUserCreate, audit.ResultSuccess) {
		t.Fatal("no quedó el registro user.create de éxito")
	}

	// Correo duplicado aunque se escriba con otras mayúsculas → 409 y fallo
	// registrado (Q5/SC-011, P20).
	in.Email = "  CARLOS@ejemplo.com "
	_, err = svc.CreateUser(ctx, admin.ID, in)
	assertKind(t, err, apperr.KindConflict)
	if !hasAdminAction(t, repo, audit.ActionUserCreate, audit.ResultFailure) {
		t.Fatal("la creación fallida debe dejar su fila con result='failure'")
	}
	if total, err := repo.CountUsers(ctx); err != nil || total != 2 {
		t.Fatalf("cuentas = %d (%v), se esperaban 2 (admin + Carlos)", total, err)
	}
}

func TestIntegrationUpdateUserGuardAntiLockout(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	adminRole := integrationAdminRole(t, repo)
	admin := mustInsertUser(t, repo, "ana@ejemplo.com", adminRole.ID, true)
	sessions := &fakeUserSessionStore{}
	svc := integrationUserService(t, repo, sessions)

	// Único administrador: desactivarlo → 409 con reason admin_required y nada
	// se escribe (la transacción se revierte).
	inactive := false
	_, err := svc.UpdateUser(ctx, admin.ID, admin.ID, UpdateUserInput{IsActive: &inactive})
	domainErr := requireKind(t, err, apperr.KindConflict)
	if domainErr.Details["reason"] != "admin_required" {
		t.Fatalf("details = %v, se esperaba admin_required", domainErr.Details)
	}
	stored, err := repo.GetUserByID(ctx, admin.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if !stored.IsActive {
		t.Fatal("el administrador único no debe quedar desactivado")
	}
	if !hasAdminAction(t, repo, audit.ActionUserDeactivate, audit.ResultFailure) {
		t.Fatal("el intento bloqueado debe registrarse como fallo")
	}

	// Cambiar su rol a uno sin admin → 409 por la misma regla.
	other := mustInsertRole(t, repo, "Contenido", "eventos")
	_, err = svc.UpdateUser(ctx, admin.ID, admin.ID, UpdateUserInput{RoleID: other.ID.String()})
	domainErr = requireKind(t, err, apperr.KindConflict)
	if domainErr.Details["reason"] != "admin_required" {
		t.Fatalf("details = %v", domainErr.Details)
	}

	// Con DOS administradores sí se puede desactivar uno (US2 esc. 5): se revocan
	// sus sesiones y queda el registro de éxito.
	second := mustInsertUser(t, repo, "bea@ejemplo.com", adminRole.ID, true)
	item, err := svc.UpdateUser(ctx, second.ID, admin.ID, UpdateUserInput{IsActive: &inactive})
	if err != nil {
		t.Fatalf("desactivar con dos administradores: %v", err)
	}
	if item.IsActive {
		t.Fatal("la cuenta debería quedar desactivada")
	}
	revoked := sessions.revoked()
	if len(revoked) != 1 || revoked[0] != admin.ID {
		t.Fatalf("sesiones revocadas = %v, se esperaba la cuenta desactivada", revoked)
	}
	if !hasAdminAction(t, repo, audit.ActionUserDeactivate, audit.ResultSuccess) {
		t.Fatal("no quedó el registro user.deactivate de éxito")
	}
}

func TestIntegrationResetUserPassword(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	adminRole := integrationAdminRole(t, repo)
	admin := mustInsertUser(t, repo, "ana@ejemplo.com", adminRole.ID, true)
	content := mustInsertRole(t, repo, "Contenido", "eventos")
	user := mustInsertUser(t, repo, "carlos@ejemplo.com", content.ID, true)
	sessions := &fakeUserSessionStore{}
	svc := integrationUserService(t, repo, sessions)

	if err := svc.ResetUserPassword(ctx, admin.ID, user.ID, ResetPasswordInput{Password: "Nueva9#Aa"}); err != nil {
		t.Fatalf("ResetUserPassword: %v", err)
	}

	auth, err := repo.GetUserAuthByEmail(ctx, "carlos@ejemplo.com")
	if err != nil {
		t.Fatalf("GetUserAuthByEmail: %v", err)
	}
	if err := password.Verify(auth.PasswordHash, "Nueva9#Aa"); err != nil {
		t.Fatalf("el hash no corresponde a la contraseña nueva: %v", err)
	}
	if !auth.MustChangePassword {
		t.Fatal("mustChangePassword debe quedar en true")
	}
	if revoked := sessions.revoked(); len(revoked) != 1 || revoked[0] != user.ID {
		t.Fatalf("sesiones revocadas = %v (R17)", revoked)
	}
	if !hasAdminAction(t, repo, audit.ActionUserPasswordReset, audit.ResultSuccess) {
		t.Fatal("no quedó el registro user.password_reset")
	}

	// El registro nunca guarda una contraseña (FR-026).
	actions, err := repo.ListAdminActions(ctx, AuditFilter{Limit: 100})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	for _, action := range actions {
		if action.TargetLabel != nil && strings.Contains(*action.TargetLabel, "Nueva9#Aa") {
			t.Fatalf("el registro contiene una contraseña: %+v", action)
		}
	}

	// Cuenta inexistente → 404 y fallo registrado.
	err = svc.ResetUserPassword(ctx, admin.ID, uuid.New(), ResetPasswordInput{Password: "Nueva9#Aa"})
	assertKind(t, err, apperr.KindNotFound)
	if !hasAdminAction(t, repo, audit.ActionUserPasswordReset, audit.ResultFailure) {
		t.Fatal("el restablecimiento fallido debe quedar registrado")
	}
}

func TestIntegrationListUsers(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	adminRole := integrationAdminRole(t, repo)
	mustInsertUser(t, repo, "ana@ejemplo.com", adminRole.ID, true)
	svc := integrationUserService(t, repo, &fakeUserSessionStore{})

	list, err := svc.ListUsers(ctx, paginate.Params{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("sobre = %+v", list)
	}
	if list.Items[0].LastLoginAt != nil || list.Items[0].LastLoginIP != nil {
		t.Fatal("una cuenta que nunca entró debe llevar el último acceso en nil")
	}
}
