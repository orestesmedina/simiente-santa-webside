package usuarios

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de los handlers de roles y del catálogo de permisos (T237) con httptest
// y un servicio falso, sobre la cadena real del grupo de panel (authn →
// passwordguard → authz → CSRF). No tocan PostgreSQL ni Redis.

// --- Fake del servicio ---

type fakeRolesService struct {
	mu sync.Mutex

	listFn   func(ctx context.Context, params paginate.Params) (RoleList, error)
	getFn    func(ctx context.Context, id uuid.UUID) (RoleItem, error)
	createFn func(ctx context.Context, actorID uuid.UUID, in RoleCreateInput) (RoleItem, error)
	updateFn func(ctx context.Context, actorID, id uuid.UUID, in RoleUpdateInput) (RoleItem, error)
	deleteFn func(ctx context.Context, actorID, id uuid.UUID) error
	permFn   func(ctx context.Context) (PermissionList, error)

	listParams []paginate.Params
	actors     []uuid.UUID
	creates    int
	updates    int
	deletes    int
}

func (f *fakeRolesService) ListRoles(ctx context.Context, params paginate.Params) (RoleList, error) {
	f.mu.Lock()
	f.listParams = append(f.listParams, params)
	f.mu.Unlock()
	if f.listFn != nil {
		return f.listFn(ctx, params)
	}
	return RoleList{}, nil
}

func (f *fakeRolesService) GetRole(ctx context.Context, id uuid.UUID) (RoleItem, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}
	return RoleItem{}, nil
}

func (f *fakeRolesService) CreateRole(ctx context.Context, actorID uuid.UUID, in RoleCreateInput) (RoleItem, error) {
	f.mu.Lock()
	f.creates++
	f.actors = append(f.actors, actorID)
	f.mu.Unlock()
	if f.createFn != nil {
		return f.createFn(ctx, actorID, in)
	}
	return RoleItem{}, nil
}

func (f *fakeRolesService) UpdateRole(ctx context.Context, actorID, id uuid.UUID, in RoleUpdateInput) (RoleItem, error) {
	f.mu.Lock()
	f.updates++
	f.actors = append(f.actors, actorID)
	f.mu.Unlock()
	if f.updateFn != nil {
		return f.updateFn(ctx, actorID, id, in)
	}
	return RoleItem{}, nil
}

func (f *fakeRolesService) DeleteRole(ctx context.Context, actorID, id uuid.UUID) error {
	f.mu.Lock()
	f.deletes++
	f.actors = append(f.actors, actorID)
	f.mu.Unlock()
	if f.deleteFn != nil {
		return f.deleteFn(ctx, actorID, id)
	}
	return nil
}

func (f *fakeRolesService) ListPermissions(ctx context.Context) (PermissionList, error) {
	if f.permFn != nil {
		return f.permFn(ctx)
	}
	return PermissionList{}, nil
}

func (f *fakeRolesService) lastActor() uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.actors) == 0 {
		return uuid.Nil
	}
	return f.actors[len(f.actors)-1]
}

// --- Servidor de panel con la cadena real ---

func newAdminRolesServer(t *testing.T, roles RolesService, identity session.Identity) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	store := handlerFakeStore{Session: session.Session{
		UserID:            identity.UserID,
		AbsoluteExpiresAt: time.Now().Add(time.Hour),
	}}
	resolver := handlerFakeResolver{identity: identity}
	h := NewHandler(HandlerDeps{Access: &fakeAccessService{}, Roles: roles, Logger: logger})

	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	RegisterAdmin(root, AdminDeps{
		Module: PermissionAdminUsersRoles,
		Deps: middleware.AdminDeps{
			Sessions:   store,
			Resolver:   resolver,
			CSRFSecret: adminUsersCSRFSecret,
			Logger:     logger,
		},
		Handler: h,
	})
	return testutil.NewServer(t, mux, logger)
}

func sampleRoleItem() RoleItem {
	return RoleItem{
		ID:          uuid.New().String(),
		Name:        "Contenido",
		Permissions: []string{"actividades", "eventos"},
		UserCount:   2,
		CreatedAt:   time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC),
	}
}

// --- GET /api/v1/admin/roles ---

func TestListRolesHandler(t *testing.T) {
	want := sampleRoleItem()
	svc := &fakeRolesService{
		listFn: func(_ context.Context, params paginate.Params) (RoleList, error) {
			return RoleList{Items: []RoleItem{want}, Total: 1, Limit: params.Limit, Offset: params.Offset}, nil
		},
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/roles?limit=1000", "", adminUsersHeaders(t))
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if len(svc.listParams) != 1 || svc.listParams[0].Limit != paginate.MaxLimit {
		t.Fatalf("params = %+v, el tope debe acotarse a %d", svc.listParams, paginate.MaxLimit)
	}
	var got RoleList
	if err := json.Unmarshal(resp.Body, &got); err != nil {
		t.Fatalf("cuerpo no es RoleList: %v (%s)", err, resp.Body)
	}
	if got.Total != 1 || len(got.Items) != 1 || got.Items[0].UserCount != 2 || len(got.Items[0].Permissions) != 2 {
		t.Fatalf("sobre = %+v", got)
	}
	if !strings.Contains(string(resp.Body), `"userCount":2`) {
		t.Errorf("RoleItem debe incluir userCount: %s", resp.Body)
	}
}

func TestListRolesRejectsInvalidPagination(t *testing.T) {
	svc := &fakeRolesService{}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/roles?limit=0", "", adminUsersHeaders(t))
	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	code, details := decodeErrorEnvelope(t, resp.Body)
	if code != "invalid" {
		t.Errorf("code = %q", code)
	}
	if _, ok := details["limit"]; !ok {
		t.Errorf("details = %v, se esperaba limit", details)
	}
}

// --- POST /api/v1/admin/roles ---

func TestCreateRoleHandler(t *testing.T) {
	want := sampleRoleItem()
	identity := authorizedIdentity()
	svc := &fakeRolesService{
		createFn: func(_ context.Context, _ uuid.UUID, in RoleCreateInput) (RoleItem, error) {
			if in.Name != "Contenido" || len(in.Permissions) != 2 {
				t.Errorf("entrada recibida = %+v", in)
			}
			return want, nil
		},
	}
	srv := newAdminRolesServer(t, svc, identity)

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/roles",
		`{"name":"Contenido","permissions":["eventos","actividades"]}`, adminUsersHeaders(t))

	if resp.Status != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201 (%s)", resp.Status, resp.Body)
	}
	body, _ := json.Marshal(want)
	if got := strings.TrimSpace(string(resp.Body)); got != string(body) {
		t.Errorf("cuerpo = %s, se esperaba %s", got, body)
	}
	if svc.creates != 1 || svc.lastActor() != identity.UserID {
		t.Errorf("el servicio no recibió al actor de la sesión: %v", svc.lastActor())
	}
}

func TestCreateRoleHandlerInvalidDTO(t *testing.T) {
	svc := &fakeRolesService{}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/roles",
		`{"permissions":["eventos"]}`, adminUsersHeaders(t))

	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if _, ok := details["name"]; !ok {
		t.Errorf("details = %v, se esperaba name", details)
	}
	if svc.creates != 0 {
		t.Error("un DTO inválido no debe llegar al servicio")
	}
}

func TestCreateRoleHandlerNoPermissions(t *testing.T) {
	svc := &fakeRolesService{
		createFn: func(context.Context, uuid.UUID, RoleCreateInput) (RoleItem, error) {
			return RoleItem{}, apperr.Invalid(messageRoleNeedsPermission)
		},
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/roles",
		`{"name":"Contenido","permissions":[]}`, adminUsersHeaders(t))

	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	if !strings.Contains(string(resp.Body), messageRoleNeedsPermission) {
		t.Errorf("mensaje inesperado: %s", resp.Body)
	}
}

func TestCreateRoleHandlerConflict(t *testing.T) {
	svc := &fakeRolesService{
		createFn: func(context.Context, uuid.UUID, RoleCreateInput) (RoleItem, error) {
			return RoleItem{}, apperr.Conflict(messageRoleNameInUse)
		},
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/roles",
		`{"name":"Contenido","permissions":["eventos"]}`, adminUsersHeaders(t))

	if resp.Status != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", resp.Status, resp.Body)
	}
	if !strings.Contains(string(resp.Body), messageRoleNameInUse) {
		t.Errorf("mensaje inesperado: %s", resp.Body)
	}
}

// --- GET /api/v1/admin/roles/{id} ---

func TestGetRoleHandler(t *testing.T) {
	want := sampleRoleItem()
	svc := &fakeRolesService{
		getFn: func(context.Context, uuid.UUID) (RoleItem, error) { return want, nil },
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/roles/"+want.ID, "", adminUsersHeaders(t))
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if !strings.Contains(string(resp.Body), `"permissions"`) {
		t.Errorf("la ficha debe incluir los permisos: %s", resp.Body)
	}
}

func TestGetRoleHandlerNotFound(t *testing.T) {
	svc := &fakeRolesService{
		getFn: func(context.Context, uuid.UUID) (RoleItem, error) {
			return RoleItem{}, apperr.NotFound("El rol no existe")
		},
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/roles/"+uuid.New().String(), "", adminUsersHeaders(t))
	if resp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
	}
}

func TestGetRoleHandlerInvalidID(t *testing.T) {
	svc := &fakeRolesService{}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/roles/no-es-un-uuid", "", adminUsersHeaders(t))
	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if _, ok := details["id"]; !ok {
		t.Errorf("details = %v, se esperaba id", details)
	}
}

// --- PATCH /api/v1/admin/roles/{id} ---

func TestUpdateRoleHandler(t *testing.T) {
	want := sampleRoleItem()
	identity := authorizedIdentity()
	svc := &fakeRolesService{
		updateFn: func(_ context.Context, _ uuid.UUID, id uuid.UUID, in RoleUpdateInput) (RoleItem, error) {
			if id.String() != want.ID {
				t.Errorf("id = %v", id)
			}
			if in.Name != "Contenido General" {
				t.Errorf("entrada = %+v", in)
			}
			return want, nil
		},
	}
	srv := newAdminRolesServer(t, svc, identity)

	resp := testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/roles/"+want.ID,
		`{"name":"Contenido General"}`, adminUsersHeaders(t))

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if svc.lastActor() != identity.UserID {
		t.Errorf("actor = %v, se esperaba %v", svc.lastActor(), identity.UserID)
	}
}

func TestUpdateRoleHandlerGuardConflict(t *testing.T) {
	svc := &fakeRolesService{
		updateFn: func(context.Context, uuid.UUID, uuid.UUID, RoleUpdateInput) (RoleItem, error) {
			return RoleItem{}, apperr.Conflict(
				"No se puede dejar el panel sin administración",
				apperr.WithDetails(map[string]any{"reason": "admin_required"}),
			)
		},
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/roles/"+uuid.New().String(),
		`{"permissions":["eventos"]}`, adminUsersHeaders(t))

	if resp.Status != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if details["reason"] != "admin_required" {
		t.Errorf("details = %v, se esperaba admin_required", details)
	}
}

// --- DELETE /api/v1/admin/roles/{id} ---

func TestDeleteRoleHandler(t *testing.T) {
	identity := authorizedIdentity()
	svc := &fakeRolesService{}
	srv := newAdminRolesServer(t, svc, identity)

	resp := testutil.Do(t, srv, http.MethodDelete, "/api/v1/admin/roles/"+uuid.New().String(), "", adminUsersHeaders(t))

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if got := strings.TrimSpace(string(resp.Body)); got != `{"deleted":true}` {
		t.Errorf("cuerpo = %s", got)
	}
	if svc.deletes != 1 || svc.lastActor() != identity.UserID {
		t.Errorf("el servicio no recibió al actor: %v", svc.lastActor())
	}
}

func TestDeleteRoleHandlerInUse(t *testing.T) {
	svc := &fakeRolesService{
		deleteFn: func(context.Context, uuid.UUID, uuid.UUID) error {
			return apperr.Conflict(messageRoleInUse,
				apperr.WithDetails(map[string]any{"userCount": int64(3)}))
		},
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodDelete, "/api/v1/admin/roles/"+uuid.New().String(), "", adminUsersHeaders(t))

	if resp.Status != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if count, ok := details["userCount"]; !ok || count != float64(3) {
		t.Errorf("details = %v, se esperaba userCount=3", details)
	}
	if !strings.Contains(string(resp.Body), messageRoleInUse) {
		t.Errorf("mensaje inesperado: %s", resp.Body)
	}
}

// --- GET /api/v1/admin/permisos ---

func TestListPermissionsHandler(t *testing.T) {
	svc := &fakeRolesService{
		permFn: func(context.Context) (PermissionList, error) {
			return PermissionList{Items: []PermissionItem{
				{Code: "eventos", Label: "Eventos"},
				{Code: "admin_usuarios_roles", Label: "Administración de usuarios y roles"},
			}}, nil
		},
	}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/permisos", "", adminUsersHeaders(t))
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if !strings.Contains(string(resp.Body), `"code":"admin_usuarios_roles"`) {
		t.Errorf("catálogo inesperado: %s", resp.Body)
	}
}

// --- Autenticación, permiso y CSRF (cadena del panel) ---

func TestRolesRoutesRequireSessionPermissionAndCSRF(t *testing.T) {
	svc := &fakeRolesService{}
	srv := newAdminRolesServer(t, svc, authorizedIdentity())

	t.Run("sin sesión responde 401", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/roles", "", nil)
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("sin permiso responde 403", func(t *testing.T) {
		sinPermiso := session.Identity{UserID: uuid.New()}
		deniedSrv := newAdminRolesServer(t, svc, sinPermiso)
		resp := testutil.Do(t, deniedSrv, http.MethodGet, "/api/v1/admin/permisos", "", adminUsersHeaders(t))
		if resp.Status != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		code, _ := decodeErrorEnvelope(t, resp.Body)
		if code != "forbidden" {
			t.Errorf("code = %q", code)
		}
	})

	t.Run("POST sin CSRF responde 403", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/roles", `{"name":"X","permissions":["eventos"]}`,
			map[string]string{"Cookie": session.CookieSession + "=tok", "Content-Type": "application/json"})
		if resp.Status != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if svc.creates != 0 {
			t.Error("sin CSRF no debe llegar al servicio")
		}
	})

	t.Run("PATCH sin CSRF responde 403", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/roles/"+uuid.New().String(), `{"name":"X"}`,
			map[string]string{"Cookie": session.CookieSession + "=tok", "Content-Type": "application/json"})
		if resp.Status != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("DELETE sin CSRF responde 403", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodDelete, "/api/v1/admin/roles/"+uuid.New().String(), "",
			map[string]string{"Cookie": session.CookieSession + "=tok"})
		if resp.Status != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if svc.deletes != 0 {
			t.Error("sin CSRF no debe llegar al servicio")
		}
	})
}
