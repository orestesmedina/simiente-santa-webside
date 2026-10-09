//go:build integration

package usuarios

import (
	"context"
	"sync"
	"testing"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de la inicialización única (FR-007/SC-003) contra
// PostgreSQL real: creación del administrador con los 9 permisos y su registro
// sin actor, y la carrera de dos `Initialize` simultáneos.

func integrationInitService(t *testing.T, repo *repository) *initService {
	t.Helper()
	logger, _ := testutil.NewLogger()
	return NewInitService(InitServiceDeps{Repository: repo, Logger: logger})
}

func integrationInitInput(email string) InitializeInput {
	return InitializeInput{
		FirstName: "Ana",
		LastName:  "Responsable",
		Email:     email,
		Phone:     "+34 612 345 678",
		Password:  "Semilla.2026",
	}
}

func TestIntegrationInitializeCreatesAdminOnce(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	svc := integrationInitService(t, repo)

	item, err := svc.Initialize(ctx, integrationInitInput("ana@ejemplo.com"))
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if item.RoleName != adminRoleName || !item.IsActive || item.MustChangePassword {
		t.Fatalf("administrador inicial inesperado: %+v", item)
	}

	// El rol nace con TODO el catálogo (9 permisos) e incluye administración.
	role, err := repo.GetRoleByNameLower(ctx, adminRoleName)
	if err != nil {
		t.Fatalf("GetRoleByNameLower: %v", err)
	}
	detail, err := repo.GetRoleByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRoleByID: %v", err)
	}
	if len(detail.Permissions) != 9 {
		t.Fatalf("permisos del rol inicial = %v, se esperaban 9", detail.Permissions)
	}
	hasAdmin := false
	for _, code := range detail.Permissions {
		if code == PermissionAdminUsersRoles {
			hasAdmin = true
		}
	}
	if !hasAdmin {
		t.Fatalf("el rol inicial no concede %q: %v", PermissionAdminUsersRoles, detail.Permissions)
	}
	if admins, err := repo.CountActiveAdmins(ctx); err != nil || admins != 1 {
		t.Fatalf("CountActiveAdmins = %d (%v), se esperaba 1", admins, err)
	}

	// El registro `user.create` queda SIN actor (único caso permitido, FR-023).
	actions, err := repo.ListAdminActions(ctx, AuditFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	initialized := false
	for _, action := range actions {
		if action.Action != audit.ActionUserCreate || action.ActorUserID != nil {
			continue
		}
		if action.Result != audit.ResultSuccess {
			t.Fatalf("resultado de la inicialización = %q", action.Result)
		}
		if action.TargetLabel == nil || *action.TargetLabel != "ana@ejemplo.com" {
			t.Fatalf("objetivo de la inicialización = %v", action.TargetLabel)
		}
		initialized = true
	}
	if !initialized {
		t.Fatal("no quedó el registro user.create sin actor")
	}

	// Repetirla → 409 y sin cuenta nueva (la acción es única, US2 esc. 2).
	_, err = svc.Initialize(ctx, integrationInitInput("otra@ejemplo.com"))
	assertKind(t, err, apperr.KindConflict)
	if total, err := repo.CountUsers(ctx); err != nil || total != 1 {
		t.Fatalf("CountUsers tras repetir = %d (%v), se esperaba 1", total, err)
	}
	if _, err := repo.GetUserByEmail(ctx, "otra@ejemplo.com"); err == nil {
		t.Fatal("la repetición no debe crear ninguna cuenta")
	}
}

// TestIntegrationInitializeRaceCreatesExactlyOne demuestra SC-003: dos
// `Initialize` simultáneos → exactamente uno crea (el advisory lock del guard
// serializa la comprobación de "sin cuentas" con las escrituras).
func TestIntegrationInitializeRaceCreatesExactlyOne(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	svc := integrationInitService(t, repo)

	inputs := []InitializeInput{
		integrationInitInput("ana@ejemplo.com"),
		integrationInitInput("bea@ejemplo.com"),
	}
	errs := make([]error, len(inputs))
	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := range inputs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[i] = svc.Initialize(context.Background(), inputs[i])
		}(i)
	}
	close(start)
	wg.Wait()

	successes, conflicts := 0, 0
	for _, err := range errs {
		if err == nil {
			successes++
			continue
		}
		assertKind(t, err, apperr.KindConflict)
		conflicts++
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("resultados = %v (éxitos=%d, conflictos=%d), se esperaba 1 y 1",
			errs, successes, conflicts)
	}

	if total, err := repo.CountUsers(ctx); err != nil || total != 1 {
		t.Fatalf("CountUsers tras la carrera = %d (%v), se esperaba 1", total, err)
	}
	if admins, err := repo.CountActiveAdmins(ctx); err != nil || admins != 1 {
		t.Fatalf("CountActiveAdmins tras la carrera = %d (%v), se esperaba 1", admins, err)
	}

	// Un único registro de inicialización sin actor.
	actions, err := repo.ListAdminActions(ctx, AuditFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	created := 0
	for _, action := range actions {
		if action.Action == audit.ActionUserCreate && action.ActorUserID == nil {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("registros user.create sin actor = %d, se esperaba 1", created)
	}
}
