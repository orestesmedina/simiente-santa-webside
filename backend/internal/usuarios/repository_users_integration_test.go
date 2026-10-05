//go:build integration

package usuarios

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/paginate"
)

// Pruebas de integración de repository_users.go contra PostgreSQL real.

func TestIntegrationInsertUserAndGetUserAuthByEmail(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Administrador", "eventos", "admin_usuarios_roles")
	user := mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	auth, err := repo.GetUserAuthByEmail(ctx, "ana@ejemplo.com")
	if err != nil {
		t.Fatalf("GetUserAuthByEmail: %v", err)
	}
	if auth.ID != user.ID {
		t.Fatalf("id = %v, se esperaba %v", auth.ID, user.ID)
	}
	if auth.PasswordHash != "hash-de-prueba" || !auth.IsActive || !auth.MustChangePassword {
		t.Fatalf("proyección de autenticación inesperada: %+v", auth)
	}
	if auth.RoleName != "Administrador" || auth.RoleID != role.ID {
		t.Fatalf("rol inesperado: %+v", auth)
	}
	if len(auth.Permissions) != 2 {
		t.Fatalf("permisos = %v, se esperaban 2", auth.Permissions)
	}
}

func TestIntegrationInsertUserDuplicateEmail(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	_, err := repo.InsertUser(ctx, NewUser{
		Email: "ana@ejemplo.com", FirstName: "Otra", LastName: "Persona",
		Phone: "699999999", PasswordHash: "hash", RoleID: role.ID,
	})
	if err == nil {
		t.Fatal("correo duplicado debería fallar")
	}
	assertKind(t, err, apperr.KindConflict)
}

func TestIntegrationUpdateUserLastLogin(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	user := mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	before, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if before.LastLoginAt != nil || before.LastLoginIP != nil {
		t.Fatalf("una cuenta nueva no debe tener último acceso: %+v", before)
	}

	at := time.Date(2026, 10, 4, 12, 30, 0, 0, time.UTC)
	if err := repo.UpdateUserLastLogin(ctx, user.ID, at, "10.0.0.5"); err != nil {
		t.Fatalf("UpdateUserLastLogin: %v", err)
	}

	after, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID tras login: %v", err)
	}
	if after.LastLoginAt == nil || !after.LastLoginAt.Equal(at) {
		t.Fatalf("lastLoginAt = %v, se esperaba %v", after.LastLoginAt, at)
	}
	if after.LastLoginIP == nil || *after.LastLoginIP != "10.0.0.5" {
		t.Fatalf("lastLoginIp = %v", after.LastLoginIP)
	}

	// Ningún DTO de entrada acepta el último acceso (FR-021).
	assertNoLastLoginFields(t, reflect.TypeOf(UpdateUserInput{}))
	assertNoLastLoginFields(t, reflect.TypeOf(CreateUserInput{}))
}

func TestIntegrationUpdateUserForeignKeyAndDuplicate(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	other := mustInsertRole(t, repo, "Otro", "eventos")
	user := mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)
	mustInsertUser(t, repo, "luis@ejemplo.com", other.ID, true)

	// Rol inexistente → violación de FK → apperr.Invalid CON details.roleId
	// (M6: se distingue del 400 genérico, como en los flujos de cuenta).
	_, err := repo.UpdateUser(ctx, UserUpdate{
		ID: user.ID, Email: "ana@ejemplo.com", FirstName: "Ana", LastName: "Pérez",
		Phone: "612345678", RoleID: uuid.New(), IsActive: true,
	})
	assertKind(t, err, apperr.KindInvalid)
	assertRoleDetail(t, err)

	// La misma traducción aplica a la creación (carrera con el rol).
	_, err = repo.InsertUser(ctx, NewUser{
		Email: "nuevo@ejemplo.com", FirstName: "Nuevo", LastName: "Usuario",
		Phone: "699999999", PasswordHash: "hash", RoleID: uuid.New(),
	})
	assertKind(t, err, apperr.KindInvalid)
	assertRoleDetail(t, err)

	// Correo ya en uso por otra cuenta → apperr.Conflict.
	_, err = repo.UpdateUser(ctx, UserUpdate{
		ID: user.ID, Email: "luis@ejemplo.com", FirstName: "Ana", LastName: "Pérez",
		Phone: "612345678", RoleID: role.ID, IsActive: true,
	})
	assertKind(t, err, apperr.KindConflict)
}

// assertRoleDetail comprueba que el error lleva details.roleId (M6).
func assertRoleDetail(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("se esperaba *apperr.Error, se obtuvo %v", err)
	}
	if _, ok := appErr.Details["roleId"]; !ok {
		t.Fatalf("details = %v, se esperaba roleId", appErr.Details)
	}
}

func TestIntegrationUpdateUserPasswordAndMustChange(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	user := mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	if err := repo.UpdateUserPassword(ctx, user.ID, "hash-nuevo"); err != nil {
		t.Fatalf("UpdateUserPassword: %v", err)
	}
	if err := repo.SetUserMustChangePassword(ctx, user.ID, false); err != nil {
		t.Fatalf("SetUserMustChangePassword: %v", err)
	}

	auth, err := repo.GetUserAuthByEmail(ctx, "ana@ejemplo.com")
	if err != nil {
		t.Fatalf("GetUserAuthByEmail: %v", err)
	}
	if auth.PasswordHash != "hash-nuevo" || auth.MustChangePassword {
		t.Fatalf("hash/flag tras la actualización: %+v", auth)
	}

	byEmail, err := repo.GetUserByEmail(ctx, "ana@ejemplo.com")
	if err != nil || byEmail.ID != user.ID {
		t.Fatalf("GetUserByEmail = %+v (%v)", byEmail, err)
	}
	if _, err := repo.GetUserByEmail(ctx, "nadie@ejemplo.com"); err == nil {
		t.Fatal("correo inexistente debería devolver error")
	} else {
		assertKind(t, err, apperr.KindNotFound)
	}
}

// TestIntegrationChangeUserPasswordIsAtomic fija M3: el cambio propio guarda el
// hash y resuelve must_change_password en una sola transacción.
func TestIntegrationChangeUserPasswordIsAtomic(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	user := mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	before, err := repo.GetUserAuthByEmail(ctx, "ana@ejemplo.com")
	if err != nil {
		t.Fatalf("GetUserAuthByEmail: %v", err)
	}
	if !before.MustChangePassword {
		t.Fatal("la cuenta de prueba debe nacer con cambio obligatorio")
	}

	if err := repo.ChangeUserPassword(ctx, user.ID, "hash-atomico"); err != nil {
		t.Fatalf("ChangeUserPassword: %v", err)
	}

	after, err := repo.GetUserAuthByEmail(ctx, "ana@ejemplo.com")
	if err != nil {
		t.Fatalf("GetUserAuthByEmail tras el cambio: %v", err)
	}
	if after.PasswordHash != "hash-atomico" || after.MustChangePassword {
		t.Fatalf("el cambio no fue atómico: %+v", after)
	}
}

func TestIntegrationGetUserByIDNotFound(t *testing.T) {
	repo, _ := newTestRepo(t)

	_, err := repo.GetUserByID(context.Background(), uuid.New())
	assertKind(t, err, apperr.KindNotFound)
}

func TestIntegrationListAndCountUsers(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	for _, email := range []string{"a@ejemplo.com", "b@ejemplo.com", "c@ejemplo.com"} {
		mustInsertUser(t, repo, email, role.ID, true)
	}

	total, err := repo.CountUsers(ctx)
	if err != nil || total != 3 {
		t.Fatalf("CountUsers = %d (%v), se esperaban 3", total, err)
	}

	page, err := repo.ListUsers(ctx, paginate.Params{Limit: 2, Offset: 1})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(page) != 2 {
		t.Fatalf("ListUsers(limit=2, offset=1) devolvió %d filas", len(page))
	}

	byRole, err := repo.CountUsersByRole(ctx, role.ID)
	if err != nil || byRole != 3 {
		t.Fatalf("CountUsersByRole = %d (%v), se esperaban 3", byRole, err)
	}
	if _, err := repo.CountUsersByRole(ctx, uuid.New()); err != nil {
		t.Fatalf("CountUsersByRole(rol sin cuentas): %v", err)
	}
}

// assertNoLastLoginFields comprueba que el DTO no expone el último acceso.
func assertNoLastLoginFields(t *testing.T, typ reflect.Type) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		if strings.Contains(strings.ToLower(tag), "lastlogin") {
			t.Fatalf("%s.%s acepta el último acceso como entrada", typ.Name(), typ.Field(i).Name)
		}
	}
}
