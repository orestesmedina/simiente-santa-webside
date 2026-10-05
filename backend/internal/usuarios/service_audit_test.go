package usuarios

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	applogger "simiente-santa/backend/internal/platform/logger"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests unitarios de service_audit.go con un repositorio falso. No tocan la
// base de datos.

// fakeAuditRepo implementa AuditRepository y registra lo que recibe.
type fakeAuditRepo struct {
	mu      sync.Mutex
	events  []audit.Event
	actions []audit.Action
	err     error
}

func (f *fakeAuditRepo) InsertLoginEvent(_ context.Context, event audit.Event) (LoginEvent, error) {
	if f.err != nil {
		return LoginEvent{}, f.err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
	return LoginEvent{Result: event.Result, IP: event.IP}, nil
}

func (f *fakeAuditRepo) InsertAdminAction(_ context.Context, action audit.Action) (AdminAction, error) {
	if f.err != nil {
		return AdminAction{}, f.err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.actions = append(f.actions, action)
	return AdminAction{Action: action.Code, Result: action.Result}, nil
}

func (f *fakeAuditRepo) eventCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

func (f *fakeAuditRepo) actionCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.actions)
}

func TestRecordLoginEventAllOutcomes(t *testing.T) {
	repo := &fakeAuditRepo{}
	service := NewAuditService(repo, nil)
	ctx := context.Background()
	userID := uuid.New()

	// Los cuatro desenlaces que registra el login (FR-022): éxito, credenciales
	// incorrectas, cuenta inactiva e intento durante el bloqueo. Todos salvo el
	// éxito se registran como failure.
	events := []audit.Event{
		{UserID: &userID, Result: audit.ResultSuccess, IP: "10.0.0.1"},
		{Result: audit.ResultFailure, IP: "10.0.0.2"},                  // correo inexistente
		{UserID: &userID, Result: audit.ResultFailure, IP: "10.0.0.3"}, // cuenta inactiva
		{Result: audit.ResultFailure, IP: "10.0.0.4"},                  // durante el bloqueo
	}
	for _, event := range events {
		if err := service.RecordLoginEvent(ctx, event); err != nil {
			t.Fatalf("RecordLoginEvent(%+v): %v", event, err)
		}
	}
	if repo.eventCount() != 4 {
		t.Fatalf("eventos registrados = %d, se esperaban 4", repo.eventCount())
	}
	if repo.events[0].Result != audit.ResultSuccess || repo.events[1].Result != audit.ResultFailure {
		t.Fatalf("resultados = %v", repo.events)
	}
	if repo.events[2].IP != "10.0.0.3" {
		t.Fatalf("IP no registrada: %+v", repo.events[2])
	}
}

func TestRecordActionAllResults(t *testing.T) {
	repo := &fakeAuditRepo{}
	service := NewAuditService(repo, nil)
	ctx := context.Background()
	actor := uuid.New()
	target := uuid.New()

	actions := []audit.Action{
		{ActorUserID: &actor, Code: audit.ActionUserCreate, TargetKind: audit.TargetUser, TargetUserID: &target, TargetLabel: "ana@ejemplo.com", Result: audit.ResultSuccess},
		{ActorUserID: &actor, Code: audit.ActionUserUpdate, TargetKind: audit.TargetUser, TargetUserID: &target, TargetLabel: "ana@ejemplo.com", Result: audit.ResultFailure},
		{ActorUserID: &actor, Code: audit.ActionRoleDelete, TargetKind: audit.TargetRole, TargetRoleID: &target, TargetLabel: "Editor", Result: audit.ResultDenied},
	}
	for _, action := range actions {
		if err := service.RecordAction(ctx, action); err != nil {
			t.Fatalf("RecordAction(%+v): %v", action, err)
		}
	}
	if repo.actionCount() != 3 {
		t.Fatalf("acciones registradas = %d, se esperaban 3", repo.actionCount())
	}
	if repo.actions[1].Result != audit.ResultFailure || repo.actions[2].Result != audit.ResultDenied {
		t.Fatalf("resultados = %v", repo.actions)
	}
	// De un restablecimiento solo queda quién, sobre qué y cuándo (FR-026): el
	// tipo no tiene ningún campo de contraseña.
	if err := service.RecordAction(ctx, audit.Action{
		ActorUserID: &actor, Code: audit.ActionUserPasswordReset, TargetKind: audit.TargetUser,
		TargetUserID: &target, TargetLabel: "ana@ejemplo.com", Result: audit.ResultSuccess,
	}); err != nil {
		t.Fatalf("RecordAction(password_reset): %v", err)
	}
}

func TestRecordActionRejectsInvalid(t *testing.T) {
	service := NewAuditService(&fakeAuditRepo{}, nil)
	ctx := context.Background()

	if err := service.RecordAction(ctx, audit.Action{Code: "user.borrar", TargetKind: audit.TargetUser, Result: audit.ResultSuccess}); err == nil {
		t.Fatal("código de acción desconocido debería fallar")
	}
	if err := service.RecordLoginEvent(ctx, audit.Event{Result: audit.ResultDenied, IP: "10.0.0.1"}); err == nil {
		t.Fatal("un intento de acceso no puede ser 'denied'")
	}
	if err := service.RecordLoginEvent(ctx, audit.Event{Result: audit.ResultFailure}); err == nil {
		t.Fatal("un intento sin IP debería fallar")
	}
}

func TestRecordDeniedResolvesActionFromRoute(t *testing.T) {
	repo := &fakeAuditRepo{}
	service := NewAuditService(repo, nil)
	ctx := context.Background()
	actor := uuid.New()
	target := uuid.New()

	denials := []struct {
		name       string
		method     string
		path       string
		wantCode   string
		wantKind   audit.TargetKind
		wantTarget *uuid.UUID
	}{
		{"crear cuenta", http.MethodPost, "/api/v1/admin/usuarios", audit.ActionUserCreate, audit.TargetUser, nil},
		{"editar cuenta", http.MethodPatch, "/api/v1/admin/usuarios/" + target.String(), audit.ActionUserUpdate, audit.TargetUser, &target},
		{"restablecer contraseña", http.MethodPost, "/api/v1/admin/usuarios/" + target.String() + "/password", audit.ActionUserPasswordReset, audit.TargetUser, &target},
		{"crear rol", http.MethodPost, "/api/v1/admin/roles", audit.ActionRoleCreate, audit.TargetRole, nil},
		{"editar rol", http.MethodPatch, "/api/v1/admin/roles/" + target.String(), audit.ActionRoleUpdate, audit.TargetRole, &target},
		{"eliminar rol", http.MethodDelete, "/api/v1/admin/roles/" + target.String(), audit.ActionRoleDelete, audit.TargetRole, &target},
	}
	for _, tc := range denials {
		t.Run(tc.name, func(t *testing.T) {
			before := repo.actionCount()
			if err := service.RecordDenied(ctx, audit.Denial{ActorUserID: &actor, Method: tc.method, Path: tc.path}); err != nil {
				t.Fatalf("RecordDenied: %v", err)
			}
			if repo.actionCount() != before+1 {
				t.Fatalf("no se registró la denegación")
			}
			last := repo.actions[repo.actionCount()-1]
			if last.Code != tc.wantCode || last.TargetKind != tc.wantKind || last.Result != audit.ResultDenied {
				t.Fatalf("acción resuelta = %+v", last)
			}
			if last.ActorUserID == nil || *last.ActorUserID != actor {
				t.Fatalf("actor = %v", last.ActorUserID)
			}
			switch tc.wantKind {
			case audit.TargetUser:
				if !sameUUID(last.TargetUserID, tc.wantTarget) {
					t.Fatalf("objetivo usuario = %v, se esperaba %v", last.TargetUserID, tc.wantTarget)
				}
			case audit.TargetRole:
				if !sameUUID(last.TargetRoleID, tc.wantTarget) {
					t.Fatalf("objetivo rol = %v, se esperaba %v", last.TargetRoleID, tc.wantTarget)
				}
			}
		})
	}

	// Una lectura denegada no es una acción administrativa: no deja fila (R23).
	before := repo.actionCount()
	if err := service.RecordDenied(ctx, audit.Denial{ActorUserID: &actor, Method: http.MethodGet, Path: "/api/v1/admin/auditoria/accesos"}); err != nil {
		t.Fatalf("RecordDenied(GET): %v", err)
	}
	// Una ruta ajena al panel y una ruta vacía tampoco.
	for _, path := range []string{"/api/v1/setup/initialize", ""} {
		if err := service.RecordDenied(ctx, audit.Denial{ActorUserID: &actor, Method: http.MethodPost, Path: path}); err != nil {
			t.Fatalf("RecordDenied(%q): %v", path, err)
		}
	}
	if repo.actionCount() != before {
		t.Fatalf("las rutas no sensibles no deben registrarse")
	}
}

func TestBestEffortNeverChangesResponse(t *testing.T) {
	repo := &fakeAuditRepo{err: errors.New("bd caída")}
	logger, capture := testutil.NewLogger()
	service := NewAuditService(repo, logger)

	ctx := httpserver.ContextWithRequestLogger(context.Background(),
		applogger.Request(logger, "req-123", http.MethodPost, "/api/v1/auth/login"))

	// Un fallo del registro no se propaga: queda en el log (R23).
	service.RecordLoginEventBestEffort(ctx, audit.Event{Result: audit.ResultFailure, IP: "10.0.0.1"})
	service.RecordActionBestEffort(ctx, audit.Action{
		Code: audit.ActionUserCreate, TargetKind: audit.TargetUser, Result: audit.ResultFailure,
	})

	if capture.Count() != 2 {
		t.Fatalf("registros de log = %d, se esperaban 2", capture.Count())
	}
	for _, record := range capture.Records() {
		if record.Attrs["request_id"] != "req-123" {
			t.Fatalf("el log no lleva el request_id: %+v", record.Attrs)
		}
	}

	// El camino fail-closed (éxito) sí propaga el error para que el login no
	// conceda acceso sin registro.
	if err := service.RecordLoginEvent(ctx, audit.Event{Result: audit.ResultSuccess, IP: "10.0.0.1"}); err == nil {
		t.Fatal("RecordLoginEvent debe propagar el fallo de escritura")
	}
}

func TestAuditTypesHaveNoCredentials(t *testing.T) {
	assertNoPasswordField(t, reflect.TypeOf(audit.Event{}))
	assertNoPasswordField(t, reflect.TypeOf(audit.Action{}))
	assertNoPasswordField(t, reflect.TypeOf(LoginEvent{}))
	assertNoPasswordField(t, reflect.TypeOf(AdminAction{}))
}

func assertNoPasswordField(t *testing.T, typ reflect.Type) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		if strings.Contains(strings.ToLower(typ.Field(i).Name), "password") {
			t.Fatalf("%s contiene el campo sensible %s (FR-026)", typ.Name(), typ.Field(i).Name)
		}
	}
}

func sameUUID(a, b *uuid.UUID) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}
