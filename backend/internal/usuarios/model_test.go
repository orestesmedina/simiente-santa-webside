package usuarios

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/validate"
)

// Tests unitarios de los DTOs del contrato (etiquetas `validate`, T213) y de
// las conversiones a los DTOs de salida.

func TestCreateUserInputValidation(t *testing.T) {
	valid := CreateUserInput{
		FirstName: "Ana",
		LastName:  "Pérez",
		Email:     "ana@ejemplo.com",
		Phone:     "+34 612 345 678",
		RoleID:    uuid.New().String(),
		Password:  "Segura123!",
	}
	if err := validate.Struct(valid); err != nil {
		t.Fatalf("DTO válido devolvió error: %v", err)
	}

	tests := []struct {
		name  string
		field string
		mutar func(*CreateUserInput)
	}{
		{"nombre ausente", "firstName", func(in *CreateUserInput) { in.FirstName = "  " }},
		{"correo mal formado", "email", func(in *CreateUserInput) { in.Email = "no-es-correo" }},
		{"teléfono no telefónico", "phone", func(in *CreateUserInput) { in.Phone = "abc" }},
		{"teléfono corto", "phone", func(in *CreateUserInput) { in.Phone = "12345" }},
		{"rol ausente", "roleId", func(in *CreateUserInput) { in.RoleID = "" }},
		{"contraseña corta", "password", func(in *CreateUserInput) { in.Password = "corta" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := valid
			tc.mutar(&in)
			err := validate.Struct(in)
			var appErr *apperr.Error
			if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalid {
				t.Fatalf("se esperaba apperr.Invalid, se obtuvo %v", err)
			}
			if _, ok := appErr.Details[tc.field]; !ok {
				t.Fatalf("details sin el campo %q: %v", tc.field, appErr.Details)
			}
		})
	}
}

func TestLoginInputValidation(t *testing.T) {
	if err := validate.Struct(LoginInput{Email: "ana@ejemplo.com", Password: "x"}); err != nil {
		t.Fatalf("login válido devolvió error: %v", err)
	}
	if err := validate.Struct(LoginInput{Email: "malo", Password: "x"}); err == nil {
		t.Fatal("correo mal formado debería fallar")
	}
	if err := validate.Struct(LoginInput{Email: "ana@ejemplo.com"}); err == nil {
		t.Fatal("contraseña ausente debería fallar")
	}
}

func TestInitializeAndRoleInputValidation(t *testing.T) {
	init := InitializeInput{
		FirstName: "Ana", LastName: "Pérez", Email: "ana@ejemplo.com",
		Phone: "612345678", Password: "Segura123!",
	}
	if err := validate.Struct(init); err != nil {
		t.Fatalf("InitializeInput válido devolvió error: %v", err)
	}
	if err := validate.Struct(RoleCreateInput{Name: "Editor"}); err != nil {
		t.Fatalf("el DTO sin permisos ya no falla en la forma: el mínimo de un permiso lo exige el service (FR-014): %v", err)
	}
	if err := validate.Struct(RoleCreateInput{Name: "Editor", Permissions: []string{"eventos"}}); err != nil {
		t.Fatalf("rol válido devolvió error: %v", err)
	}
	if err := validate.Struct(RoleCreateInput{Name: "  "}); err == nil {
		t.Fatal("un rol sin nombre debería fallar en la forma")
	}
}

func TestOutputConversions(t *testing.T) {
	id := uuid.New()
	roleID := uuid.New()
	when := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	ip := "10.0.0.1"

	user := User{
		ID: id, Email: "ana@ejemplo.com", FirstName: "Ana", LastName: "Pérez",
		Phone: "612345678", RoleID: roleID, RoleName: "Administrador", IsActive: true,
		LastLoginAt: &when, LastLoginIP: &ip, CreatedAt: when,
	}
	item := UserItemFrom(user)
	if item.ID != id.String() || item.RoleID != roleID.String() || item.LastLoginIP == nil || *item.LastLoginIP != ip {
		t.Fatalf("UserItemFrom = %+v", item)
	}
	session := SessionUserFrom(user, []string{"eventos"})
	if session.Permissions[0] != "eventos" || session.Phone != "612345678" {
		t.Fatalf("SessionUserFrom = %+v", session)
	}

	role := Role{ID: roleID, Name: "Editor", Permissions: []string{"eventos"}, UserCount: 3, CreatedAt: when}
	roleItem := RoleItemFrom(role)
	if roleItem.UserCount != 3 || len(roleItem.Permissions) != 1 || roleItem.ID != roleID.String() {
		t.Fatalf("RoleItemFrom = %+v", roleItem)
	}

	if got := PermissionItemFrom(Permission{Code: "eventos", Label: "Eventos"}); got.Code != "eventos" {
		t.Fatalf("PermissionItemFrom = %+v", got)
	}

	access := AccessEventItemFrom(AccessEvent{ID: id, UserID: &id, Result: audit.ResultSuccess, IP: ip, CreatedAt: when})
	if access.UserID == nil || *access.UserID != id.String() || access.Result != "success" {
		t.Fatalf("AccessEventItemFrom = %+v", access)
	}
	anon := AccessEventItemFrom(AccessEvent{ID: id, Result: audit.ResultFailure, IP: ip, CreatedAt: when})
	if anon.UserID != nil {
		t.Fatalf("intento sin cuenta debe dejar userId nil: %+v", anon)
	}

	action := AdminActionItemFrom(AdminActionEntry{
		ID: id, ActorUserID: &id, Action: audit.ActionUserCreate,
		TargetKind: audit.TargetUser, TargetID: &id, Result: audit.ResultSuccess, CreatedAt: when,
	})
	if action.ActorID == nil || *action.ActorID != id.String() || action.TargetKind != "user" {
		t.Fatalf("AdminActionItemFrom = %+v", action)
	}
}
