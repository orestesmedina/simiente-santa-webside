package usuarios

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de los handlers de solo lectura del registro de auditoría (T240) y del
// registro de rechazos por datos inválidos (T239). Usan httptest con un servicio
// falso (o el servicio real sobre un repositorio falso) sobre la cadena real del
// grupo de panel. No tocan PostgreSQL ni Redis.

// --- Fake del servicio de auditoría (consulta) ---

type fakeAuditService struct {
	mu sync.Mutex

	accessFn  func(ctx context.Context, filter AuditFilter) (AccessEventList, error)
	actionsFn func(ctx context.Context, filter AuditFilter) (AdminActionList, error)

	accessFilters  []AuditFilter
	actionFilters  []AuditFilter
	rejections     []Rejection
	denials        []audit.Denial
	recordDeniedFn func(ctx context.Context, denial audit.Denial) error
}

func (f *fakeAuditService) ListAccessEvents(ctx context.Context, filter AuditFilter) (AccessEventList, error) {
	f.mu.Lock()
	f.accessFilters = append(f.accessFilters, filter)
	f.mu.Unlock()
	if f.accessFn != nil {
		return f.accessFn(ctx, filter)
	}
	return AccessEventList{Items: []AccessEventItem{}, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (f *fakeAuditService) ListAdminActions(ctx context.Context, filter AuditFilter) (AdminActionList, error) {
	f.mu.Lock()
	f.actionFilters = append(f.actionFilters, filter)
	f.mu.Unlock()
	if f.actionsFn != nil {
		return f.actionsFn(ctx, filter)
	}
	return AdminActionList{Items: []AdminActionItem{}, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (f *fakeAuditService) RecordRejectedBestEffort(_ context.Context, rejection Rejection) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejections = append(f.rejections, rejection)
}

func (f *fakeAuditService) RecordDenied(ctx context.Context, denial audit.Denial) error {
	f.mu.Lock()
	f.denials = append(f.denials, denial)
	f.mu.Unlock()
	if f.recordDeniedFn != nil {
		return f.recordDeniedFn(ctx, denial)
	}
	return nil
}

// --- Servidor de panel con la cadena real ---

func newAdminAuditServer(t *testing.T, auditSvc AuditService, identity session.Identity) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	return newAdminAuditServerWithLogger(t, auditSvc, identity, logger)
}

func newAdminAuditServerWithLogger(
	t *testing.T,
	auditSvc AuditService,
	identity session.Identity,
	logger *slog.Logger,
) *httptest.Server {
	t.Helper()
	store := handlerFakeStore{Session: session.Session{
		UserID:            identity.UserID,
		AbsoluteExpiresAt: time.Now().Add(time.Hour),
	}}
	resolver := handlerFakeResolver{identity: identity}
	h := NewHandler(HandlerDeps{
		Access: &fakeAccessService{},
		Users:  &fakeUsersService{},
		Audit:  auditSvc,
		Logger: logger,
	})

	var recorder audit.Recorder
	if r, ok := auditSvc.(audit.Recorder); ok {
		recorder = r
	}

	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	RegisterAdmin(root, AdminDeps{
		Module: PermissionAdminUsersRoles,
		Deps: middleware.AdminDeps{
			Sessions:   store,
			Resolver:   resolver,
			Recorder:   recorder,
			CSRFSecret: adminUsersCSRFSecret,
			Logger:     logger,
		},
		Handler: h,
	})
	return testutil.NewServer(t, mux, logger)
}

func sampleAccessEventItem() AccessEventItem {
	userID := uuid.New().String()
	email := "ana@ejemplo.com"
	name := "Ana Pérez"
	return AccessEventItem{
		ID:        uuid.New().String(),
		UserID:    &userID,
		UserEmail: &email,
		UserName:  &name,
		Result:    string(audit.ResultSuccess),
		IP:        "10.0.0.1",
		CreatedAt: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
	}
}

// --- GET /api/v1/admin/auditoria/accesos ---

func TestListAccessEventsHandler(t *testing.T) {
	item := sampleAccessEventItem()
	svc := &fakeAuditService{
		accessFn: func(_ context.Context, filter AuditFilter) (AccessEventList, error) {
			return AccessEventList{Items: []AccessEventItem{item}, Total: 1, Limit: filter.Limit, Offset: filter.Offset}, nil
		},
	}
	srv := newAdminAuditServer(t, svc, authorizedIdentity())
	userID := uuid.New().String()

	resp := testutil.Do(t, srv, http.MethodGet,
		"/api/v1/admin/auditoria/accesos?limit=1000&offset=0&userId="+userID+
			"&from=2026-10-01T00:00:00Z&to=2026-10-05T00:00:00Z",
		"", adminUsersHeaders(t))

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	var got AccessEventList
	if err := json.Unmarshal(resp.Body, &got); err != nil {
		t.Fatalf("cuerpo no es AccessEventList: %v (%s)", err, resp.Body)
	}
	if got.Total != 1 || len(got.Items) != 1 || got.Limit != 100 || got.Offset != 0 {
		t.Fatalf("sobre = %+v", got)
	}
	if got.Items[0].UserEmail == nil || *got.Items[0].UserEmail != "ana@ejemplo.com" {
		t.Fatalf("userEmail = %v", got.Items[0].UserEmail)
	}

	filters := svc.accessFilters
	if len(filters) != 1 {
		t.Fatalf("el filtro no llegó al servicio: %+v", filters)
	}
	if filters[0].UserID == nil || filters[0].UserID.String() != userID {
		t.Fatalf("userId = %v", filters[0].UserID)
	}
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	if filters[0].From == nil || !filters[0].From.Equal(from) || filters[0].To == nil || !filters[0].To.Equal(to) {
		t.Fatalf("rango = %v..%v", filters[0].From, filters[0].To)
	}
}

// --- GET /api/v1/admin/auditoria/acciones ---

func TestListAdminActionsHandler(t *testing.T) {
	srv := newAdminAuditServer(t, &fakeAuditService{}, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/auditoria/acciones", "", adminUsersHeaders(t))
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	var got AdminActionList
	if err := json.Unmarshal(resp.Body, &got); err != nil {
		t.Fatalf("cuerpo no es AdminActionList: %v (%s)", err, resp.Body)
	}
	if got.Total != 0 || got.Limit != 20 || got.Offset != 0 {
		t.Fatalf("sobre = %+v", got)
	}
}

func TestAuditEmptyPageIsEmptyArray(t *testing.T) {
	// El servicio real sobre un repositorio vacío garantiza `items: []` (nunca
	// `null`) en el JSON (FR-024).
	real := NewAuditService(&fakeAuditRepo{}, nil)
	srv := newAdminAuditServer(t, real, authorizedIdentity())

	for _, path := range []string{
		"/api/v1/admin/auditoria/accesos",
		"/api/v1/admin/auditoria/acciones",
	} {
		resp := testutil.Do(t, srv, http.MethodGet, path, "", adminUsersHeaders(t))
		if resp.Status != http.StatusOK {
			t.Fatalf("%s: status = %d (%s)", path, resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"items":[]`) {
			t.Errorf("%s: items debe ser [] y no null: %s", path, resp.Body)
		}
	}
}

func TestAuditHandlersRejectInvalidFilters(t *testing.T) {
	svc := &fakeAuditService{}
	// Rango invertido: lo valida el servicio real (mismo camino que en
	// producción), así que este caso usa el servicio real.
	real := NewAuditService(&fakeAuditRepo{}, nil)
	realSrv := newAdminAuditServer(t, real, authorizedIdentity())
	inverted := testutil.Do(t, realSrv, http.MethodGet,
		"/api/v1/admin/auditoria/accesos?from=2026-10-05T00:00:00Z&to=2026-10-01T00:00:00Z", "", adminUsersHeaders(t))
	if inverted.Status != http.StatusBadRequest {
		t.Fatalf("from>to: status = %d, se esperaba 400 (%s)", inverted.Status, inverted.Body)
	}

	svcSrv := newAdminAuditServer(t, svc, authorizedIdentity())
	tests := []struct {
		name       string
		query      string
		wantDetail string
	}{
		{name: "userId no UUID", query: "?userId=no-es-uuid", wantDetail: "userId"},
		{name: "from mal formada", query: "?from=ayer", wantDetail: "from"},
		{name: "to mal formada", query: "?to=2026-13-40", wantDetail: "to"},
		{name: "limit no válido", query: "?limit=0", wantDetail: "limit"},
		{name: "offset no válido", query: "?offset=-1", wantDetail: "offset"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := testutil.Do(t, svcSrv, http.MethodGet, "/api/v1/admin/auditoria/accesos"+tc.query, "", adminUsersHeaders(t))
			if resp.Status != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
			}
			code, details := decodeErrorEnvelope(t, resp.Body)
			if code != "invalid" {
				t.Errorf("code = %q", code)
			}
			if _, ok := details[tc.wantDetail]; !ok {
				t.Errorf("details = %v, se esperaba %q", details, tc.wantDetail)
			}
			if len(svc.accessFilters) != 0 {
				t.Error("una entrada inválida no debe llegar al servicio")
			}
		})
	}
}

func TestAuditRoutesRequirePermission(t *testing.T) {
	svc := &fakeAuditService{}
	srv := newAdminAuditServer(t, svc, session.Identity{UserID: uuid.New()})

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/auditoria/accesos", "", adminUsersHeaders(t))
	if resp.Status != http.StatusForbidden {
		t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
	}
	code, _ := decodeErrorEnvelope(t, resp.Body)
	if code != "forbidden" {
		t.Errorf("code = %q", code)
	}
}

func TestAuditRoutesAreReadOnly(t *testing.T) {
	srv := newAdminAuditServer(t, &fakeAuditService{}, authorizedIdentity())
	headers := adminUsersHeaders(t)

	for _, path := range []string{
		"/api/v1/admin/auditoria/accesos",
		"/api/v1/admin/auditoria/acciones",
	} {
		// GET está publicado.
		if resp := testutil.Do(t, srv, http.MethodGet, path, "", headers); resp.Status != http.StatusOK {
			t.Fatalf("GET %s: status = %d", path, resp.Status)
		}
		// Ninguna operación de escritura existe (FR-025): el router responde 405.
		for _, method := range []string{http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete} {
			resp := testutil.Do(t, srv, method, path, "{}", headers)
			if resp.Status != http.StatusMethodNotAllowed {
				t.Errorf("%s %s: status = %d, se esperaba 405 (registro de solo lectura)", method, path, resp.Status)
			}
		}
	}
}

// --- T239: puntos de escritura del handler y de authz ---

func TestDeniedOperationLeavesDeniedRow(t *testing.T) {
	repo := &fakeAuditRepo{}
	real := NewAuditService(repo, nil)
	identity := session.Identity{UserID: uuid.New()} // sin permiso
	srv := newAdminAuditServer(t, real, identity)
	target := uuid.New()

	resp := testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/usuarios/"+target.String(), `{}`, adminUsersHeaders(t))
	if resp.Status != http.StatusForbidden {
		t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
	}

	if repo.actionCount() != 1 {
		t.Fatalf("acciones registradas = %d, se esperaba 1 (result=denied)", repo.actionCount())
	}
	last := repo.actions[0]
	if last.Result != audit.ResultDenied || last.Code != audit.ActionUserUpdate {
		t.Fatalf("acción = %+v", last)
	}
	if last.ActorUserID == nil || *last.ActorUserID != identity.UserID {
		t.Fatalf("actor = %v, se esperaba %v", last.ActorUserID, identity.UserID)
	}
	if !sameUUID(last.TargetUserID, &target) {
		t.Fatalf("objetivo = %v, se esperaba %v", last.TargetUserID, target)
	}
}

func TestInvalidJSONLeavesFailureRowWithoutCredentials(t *testing.T) {
	repo := &fakeAuditRepo{}
	real := NewAuditService(repo, nil)
	identity := authorizedIdentity()
	srv := newAdminAuditServer(t, real, identity)

	// El cuerpo inválido lleva una contraseña: no debe acabar en el registro
	// (FR-026), solo el intento con result='failure'.
	const secret = "S3creto.2026"
	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/usuarios",
		`{"firstName":"Carlos","password":"`+secret+`",`, adminUsersHeaders(t))

	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	if repo.actionCount() != 1 {
		t.Fatalf("acciones registradas = %d, se esperaba 1 (result=failure)", repo.actionCount())
	}
	last := repo.actions[0]
	if last.Result != audit.ResultFailure || last.Code != audit.ActionUserCreate {
		t.Fatalf("acción = %+v", last)
	}
	if last.ActorUserID == nil || *last.ActorUserID != identity.UserID {
		t.Fatalf("actor = %v", last.ActorUserID)
	}
	if strings.Contains(last.TargetLabel, secret) {
		t.Fatalf("el registro guardó una credencial: %q", last.TargetLabel)
	}
	assertNoCredentials(t, resp.Body, secret)
}

func TestAuditRecordFailureDoesNotChangeResponse(t *testing.T) {
	// Un repositorio caído hace fallar el registro best-effort; la respuesta ya
	// decidida (403 por denegación o 400 por JSON inválido) no cambia (R23).
	repo := &fakeAuditRepo{err: context.DeadlineExceeded}
	real := NewAuditService(repo, nil)
	identity := session.Identity{UserID: uuid.New()}
	srv := newAdminAuditServer(t, real, identity)

	denied := testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/usuarios/"+uuid.New().String(), `{}`, adminUsersHeaders(t))
	if denied.Status != http.StatusForbidden {
		t.Fatalf("denegación con registro caído: status = %d, se esperaba 403", denied.Status)
	}
	code, _ := decodeErrorEnvelope(t, denied.Body)
	if code != "forbidden" {
		t.Errorf("code = %q", code)
	}

	authorized := newAdminAuditServer(t, real, authorizedIdentity())
	invalid := testutil.Do(t, authorized, http.MethodPost, "/api/v1/admin/usuarios", `{`, adminUsersHeaders(t))
	if invalid.Status != http.StatusBadRequest {
		t.Fatalf("JSON inválido con registro caído: status = %d, se esperaba 400", invalid.Status)
	}
	code, _ = decodeErrorEnvelope(t, invalid.Body)
	if code != "invalid" {
		t.Errorf("code = %q", code)
	}
}
