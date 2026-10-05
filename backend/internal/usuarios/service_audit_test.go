package usuarios

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	applogger "simiente-santa/backend/internal/platform/logger"
	"simiente-santa/backend/internal/platform/paginate"
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

	// Datos de consulta pre-sembrados y el total que devuelve Count*.
	accessEvents []AccessEvent
	adminEntries []AdminActionEntry
	accessTotal  int64
	adminTotal   int64
	listFilters  []AuditFilter
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

func (f *fakeAuditRepo) ListLoginEvents(_ context.Context, filter AuditFilter) ([]AccessEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listFilters = append(f.listFilters, filter)
	return f.accessEvents, nil
}

func (f *fakeAuditRepo) CountLoginEvents(_ context.Context, _ AuditFilter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.accessTotal, nil
}

func (f *fakeAuditRepo) ListAdminActions(_ context.Context, filter AuditFilter) ([]AdminActionEntry, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listFilters = append(f.listFilters, filter)
	return f.adminEntries, nil
}

func (f *fakeAuditRepo) CountAdminActions(_ context.Context, _ AuditFilter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.adminTotal, nil
}

func (f *fakeAuditRepo) filters() []AuditFilter {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]AuditFilter, len(f.listFilters))
	copy(out, f.listFilters)
	return out
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

// --- Consulta de audituría (T238, FR-024/FR-025) ---

func TestListAccessEventsWrapsPage(t *testing.T) {
	repo := &fakeAuditRepo{
		accessEvents: []AccessEvent{
			sampleAccessEvent(true),
			sampleAccessEvent(false),
		},
		accessTotal: 2,
	}
	service := NewAuditService(repo, nil)

	list, err := service.ListAccessEvents(context.Background(), AuditFilter{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListAccessEvents: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("items = %d, se esperaban 2", len(list.Items))
	}
	if list.Total != 2 || list.Limit != 1 || list.Offset != 1 {
		t.Fatalf("sobre = %+v", list)
	}
	if list.Items[0].UserEmail == nil || *list.Items[0].UserEmail != "ana@ejemplo.com" {
		t.Fatalf("userEmail derivado = %v", list.Items[0].UserEmail)
	}
	// El intento sin cuenta no lleva correo ni nombre (F-01/FR-026).
	anon := list.Items[1]
	if anon.UserID != nil || anon.UserEmail != nil || anon.UserName != nil {
		t.Fatalf("el intento sin cuenta filtró datos: %+v", anon)
	}
	filters := repo.filters()
	if len(filters) != 1 || filters[0].Limit != 1 || filters[0].Offset != 1 {
		t.Fatalf("filtro enviado al repo = %+v", filters)
	}
}

func TestListAdminActionsWrapsPage(t *testing.T) {
	actor := uuid.New()
	email := "admin@ejemplo.com"
	name := "Admin Uno"
	label := "Editor"
	repo := &fakeAuditRepo{
		adminEntries: []AdminActionEntry{{
			ID: uuid.New(), ActorUserID: &actor, ActorEmail: &email, ActorName: &name,
			Action: audit.ActionRoleDelete, TargetKind: audit.TargetRole, TargetLabel: &label,
			Result: audit.ResultDenied,
		}},
		adminTotal: 1,
	}
	service := NewAuditService(repo, nil)

	list, err := service.ListAdminActions(context.Background(), AuditFilter{UserID: &actor})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	if list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("sobre = %+v", list)
	}
	item := list.Items[0]
	if item.ActorID == nil || *item.ActorID != actor.String() {
		t.Fatalf("actorId = %v", item.ActorID)
	}
	if item.Action != audit.ActionRoleDelete || item.Result != string(audit.ResultDenied) {
		t.Fatalf("acción/resultado = %+v", item)
	}
	filters := repo.filters()
	if len(filters) != 1 || filters[0].UserID == nil || *filters[0].UserID != actor {
		t.Fatalf("filtro por cuenta no llegó: %+v", filters)
	}
}

func TestListAuditRejectsInvertedRange(t *testing.T) {
	from := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	to := from.Add(-time.Hour)
	repo := &fakeAuditRepo{}
	service := NewAuditService(repo, nil)

	if _, err := service.ListAccessEvents(context.Background(), AuditFilter{From: &from, To: &to}); !isInvalid(err) {
		t.Fatalf("ListAccessEvents(from>to) = %v, se esperaba 400 invalid", err)
	}
	if _, err := service.ListAdminActions(context.Background(), AuditFilter{From: &from, To: &to}); !isInvalid(err) {
		t.Fatalf("ListAdminActions(from>to) = %v, se esperaba 400 invalid", err)
	}
	if len(repo.filters()) != 0 {
		t.Fatal("un rango inválido no debe consultar el repositorio")
	}
}

func TestListAuditNormalizesPagination(t *testing.T) {
	repo := &fakeAuditRepo{}
	service := NewAuditService(repo, nil)
	ctx := context.Background()

	defaultList, err := service.ListAccessEvents(ctx, AuditFilter{})
	if err != nil {
		t.Fatalf("ListAccessEvents: %v", err)
	}
	if defaultList.Limit != paginate.DefaultLimit || defaultList.Offset != 0 {
		t.Fatalf("defecto = limit %d offset %d", defaultList.Limit, defaultList.Offset)
	}

	capped, err := service.ListAccessEvents(ctx, AuditFilter{Limit: paginate.MaxLimit + 50, Offset: -3})
	if err != nil {
		t.Fatalf("ListAccessEvents(tope): %v", err)
	}
	if capped.Limit != paginate.MaxLimit || capped.Offset != 0 {
		t.Fatalf("tope = limit %d offset %d", capped.Limit, capped.Offset)
	}
}

func TestListAuditEmptyIsEmptySlice(t *testing.T) {
	service := NewAuditService(&fakeAuditRepo{}, nil)
	ctx := context.Background()

	access, err := service.ListAccessEvents(ctx, AuditFilter{})
	if err != nil {
		t.Fatalf("ListAccessEvents: %v", err)
	}
	if access.Items == nil || len(access.Items) != 0 {
		t.Fatalf("items de accesos = %#v, se esperaba [] no nil", access.Items)
	}
	actions, err := service.ListAdminActions(ctx, AuditFilter{})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	if actions.Items == nil || len(actions.Items) != 0 {
		t.Fatalf("items de acciones = %#v, se esperaba [] no nil", actions.Items)
	}
}

func TestAuditQueriesNeverWrite(t *testing.T) {
	repo := &fakeAuditRepo{accessEvents: []AccessEvent{sampleAccessEvent(true)}, accessTotal: 1}
	service := NewAuditService(repo, nil)
	ctx := context.Background()

	if _, err := service.ListAccessEvents(ctx, AuditFilter{}); err != nil {
		t.Fatalf("ListAccessEvents: %v", err)
	}
	if _, err := service.ListAdminActions(ctx, AuditFilter{}); err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	if repo.eventCount() != 0 || repo.actionCount() != 0 {
		t.Fatalf("la consulta escribió: eventos=%d acciones=%d (FR-025)", repo.eventCount(), repo.actionCount())
	}
}

func TestRecordRejectedResolvesActionFromRoute(t *testing.T) {
	repo := &fakeAuditRepo{}
	logger, _ := testutil.NewLogger()
	service := NewAuditService(repo, logger)
	ctx := context.Background()
	actor := uuid.New()
	target := uuid.New()

	service.RecordRejectedBestEffort(ctx, Rejection{
		ActorUserID: &actor,
		Method:      http.MethodPatch,
		Path:        "/api/v1/admin/usuarios/" + target.String(),
	})
	if repo.actionCount() != 1 {
		t.Fatalf("rechazo no registrado")
	}
	last := repo.actions[0]
	if last.Code != audit.ActionUserUpdate || last.Result != audit.ResultFailure {
		t.Fatalf("acción de rechazo = %+v", last)
	}
	if last.ActorUserID == nil || *last.ActorUserID != actor {
		t.Fatalf("actor = %v", last.ActorUserID)
	}
	if !sameUUID(last.TargetUserID, &target) {
		t.Fatalf("objetivo = %v, se esperaba %v", last.TargetUserID, target)
	}

	// Un rechazo de una petición que no corresponde a ninguna acción sensible
	// (inicialización pública) no deja fila (R23).
	before := repo.actionCount()
	service.RecordRejectedBestEffort(ctx, Rejection{Method: http.MethodPost, Path: "/api/v1/setup/initialize"})
	if repo.actionCount() != before {
		t.Fatal("la inicialización no es una acción administrativa")
	}
}

// sampleAccessEvent construye un intento identificado (identified=true) o sin
// cuenta asociada.
func sampleAccessEvent(identified bool) AccessEvent {
	event := AccessEvent{
		ID:        uuid.New(),
		Result:    audit.ResultFailure,
		IP:        "10.0.0.1",
		CreatedAt: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
	}
	if identified {
		userID := uuid.New()
		email := "ana@ejemplo.com"
		name := "Ana Pérez"
		event.UserID = &userID
		event.UserEmail = &email
		event.UserName = &name
		event.Result = audit.ResultSuccess
	}
	return event
}

// isInvalid indica si el error es un 400 del contrato (apperr.Invalid).
func isInvalid(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindInvalid
}
