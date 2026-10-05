package session

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHasPermission(t *testing.T) {
	identity := Identity{
		UserID:      uuid.New(),
		Email:       "ana@ejemplo.com",
		FirstName:   "Ana",
		LastName:    "García",
		RoleName:    "Administrador",
		Permissions: []string{"portada", "admin_usuarios_roles"},
	}

	if !identity.HasPermission("portada") {
		t.Error("HasPermission(portada) = false, se esperaba true")
	}
	if identity.HasPermission("donaciones") {
		t.Error("HasPermission(donaciones) = true, se esperaba false")
	}
	if identity.HasPermission("") {
		t.Error("HasPermission(\"\") = true, se esperaba false")
	}
	if got := identity.FullName(); got != "Ana García" {
		t.Errorf("FullName() = %q, se esperaba Ana García", got)
	}
}

func TestIdentityFromContext(t *testing.T) {
	want := Identity{UserID: uuid.New(), Email: "ana@ejemplo.com"}

	ctx := ContextWithIdentity(context.Background(), want)
	got, ok := IdentityFromContext(ctx)
	if !ok {
		t.Fatal("IdentityFromContext() = false, se esperaba true")
	}
	if got.UserID != want.UserID || got.Email != want.Email {
		t.Errorf("IdentityFromContext() = %+v, se esperaba %+v", got, want)
	}

	if _, ok := IdentityFromContext(context.Background()); ok {
		t.Error("IdentityFromContext() sin identidad = true, se esperaba false")
	}
}
