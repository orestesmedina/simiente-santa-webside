package usuarios

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// Tests de los handlers de gestión de cuentas (T234) con httptest y un servicio
// falso, sobre la cadena real del grupo de panel (authn → passwordguard →
// authz → CSRF). No tocan PostgreSQL ni Redis.

// --- Fake del servicio ---

type fakeUsersService struct {
	mu sync.Mutex

	listFn   func(ctx context.Context, params paginate.Params) (UserList, error)
	getFn    func(ctx context.Context, id uuid.UUID) (UserItem, error)
	createFn func(ctx context.Context, actorID uuid.UUID, in CreateUserInput) (UserItem, error)
	updateFn func(ctx context.Context, actorID, id uuid.UUID, in UpdateUserInput) (UserItem, error)
	resetFn  func(ctx context.Context, actorID, id uuid.UUID, in ResetPasswordInput) error

	listParams []paginate.Params
	actors     []uuid.UUID
	creates    int
}

func (f *fakeUsersService) ListUsers(ctx context.Context, params paginate.Params) (UserList, error) {
	f.mu.Lock()
	f.listParams = append(f.listParams, params)
	f.mu.Unlock()
	if f.listFn != nil {
		return f.listFn(ctx, params)
	}
	return UserList{}, nil
}

func (f *fakeUsersService) GetUser(ctx context.Context, id uuid.UUID) (UserItem, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}
	return UserItem{}, nil
}

func (f *fakeUsersService) CreateUser(ctx context.Context, actorID uuid.UUID, in CreateUserInput) (UserItem, error) {
	f.mu.Lock()
	f.creates++
	f.actors = append(f.actors, actorID)
	f.mu.Unlock()
	if f.createFn != nil {
		return f.createFn(ctx, actorID, in)
	}
	return UserItem{}, nil
}

func (f *fakeUsersService) UpdateUser(ctx context.Context, actorID, id uuid.UUID, in UpdateUserInput) (UserItem, error) {
	f.mu.Lock()
	f.actors = append(f.actors, actorID)
	f.mu.Unlock()
	if f.updateFn != nil {
		return f.updateFn(ctx, actorID, id, in)
	}
	return UserItem{}, nil
}

func (f *fakeUsersService) ResetUserPassword(ctx context.Context, actorID, id uuid.UUID, in ResetPasswordInput) error {
	f.mu.Lock()
	f.actors = append(f.actors, actorID)
	f.mu.Unlock()
	if f.resetFn != nil {
		return f.resetFn(ctx, actorID, id, in)
	}
	return nil
}

func (f *fakeUsersService) lastActor() uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.actors) == 0 {
		return uuid.Nil
	}
	return f.actors[len(f.actors)-1]
}

func (f *fakeUsersService) listCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.listParams)
}

// --- Servidor de panel con la cadena real ---

const adminUsersCSRFSecret = "secreto-csrf-de-prueba"

func authorizedIdentity() session.Identity {
	return session.Identity{
		UserID:      uuid.New(),
		Email:       "ana@ejemplo.com",
		Permissions: []string{PermissionAdminUsersRoles},
	}
}

func adminUsersHeaders(t *testing.T) map[string]string {
	t.Helper()
	csrf, err := session.NewCSRFToken(adminUsersCSRFSecret)
	if err != nil {
		t.Fatalf("generar token CSRF: %v", err)
	}
	return map[string]string{
		"Cookie":       session.CookieSession + "=tok; " + session.CookieCSRF + "=" + csrf,
		"X-CSRF-Token": csrf,
		"Content-Type": "application/json",
	}
}

// newAdminUsersServer publica las rutas de cuentas bajo la cadena del panel con
// la identidad indicada (authn resuelve cualquier token no vacío).
func newAdminUsersServer(t *testing.T, users UsersService, identity session.Identity) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	store := handlerFakeStore{Session: session.Session{
		UserID:            identity.UserID,
		AbsoluteExpiresAt: time.Now().Add(time.Hour),
	}}
	resolver := handlerFakeResolver{identity: identity}
	h := NewHandler(HandlerDeps{Access: &fakeAccessService{}, Users: users, Logger: logger})

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

func sampleManagedUser() UserItem {
	return UserItem{
		ID:          uuid.New().String(),
		Email:       "carlos@ejemplo.com",
		FirstName:   "Carlos",
		LastName:    "Ayudante",
		Phone:       "612 345 678",
		RoleID:      uuid.New().String(),
		RoleName:    "Contenido",
		IsActive:    true,
		CreatedAt:   time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC),
		LastLoginAt: nil,
		LastLoginIP: nil,
	}
}

// --- GET /api/v1/admin/usuarios ---

func TestListUsersHandler(t *testing.T) {
	want := sampleManagedUser()
	svc := &fakeUsersService{
		listFn: func(_ context.Context, params paginate.Params) (UserList, error) {
			return UserList{Items: []UserItem{want}, Total: 1, Limit: params.Limit, Offset: params.Offset}, nil
		},
	}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/usuarios?limit=1000", "", adminUsersHeaders(t))
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if len(svc.listParams) != 1 || svc.listParams[0].Limit != paginate.MaxLimit {
		t.Fatalf("params = %+v, el tope debe acotarse a %d", svc.listParams, paginate.MaxLimit)
	}
	var got UserList
	if err := json.Unmarshal(resp.Body, &got); err != nil {
		t.Fatalf("cuerpo no es UserList: %v (%s)", err, resp.Body)
	}
	if got.Total != 1 || got.Limit != paginate.MaxLimit || len(got.Items) != 1 {
		t.Fatalf("sobre = %+v", got)
	}
	if !strings.Contains(string(resp.Body), `"lastLoginAt":null`) {
		t.Errorf("la cuenta sin acceso debe serializar lastLoginAt null: %s", resp.Body)
	}
}

func TestListUsersRejectsInvalidPagination(t *testing.T) {
	svc := &fakeUsersService{}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/usuarios?limit=0", "", adminUsersHeaders(t))
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
	if svc.listCalls() != 0 {
		t.Error("unos parámetros inválidos no deben llegar al servicio")
	}
}

// --- Autenticación y permiso (cadena del panel) ---

func TestUsersRoutesRequireSessionAndPermission(t *testing.T) {
	svc := &fakeUsersService{}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	t.Run("sin sesión responde 401", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/usuarios", "", nil)
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("sin permiso responde 403", func(t *testing.T) {
		sinPermiso := session.Identity{UserID: uuid.New()}
		deniedSrv := newAdminUsersServer(t, svc, sinPermiso)
		resp := testutil.Do(t, deniedSrv, http.MethodGet, "/api/v1/admin/usuarios", "", adminUsersHeaders(t))
		if resp.Status != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		code, _ := decodeErrorEnvelope(t, resp.Body)
		if code != "forbidden" {
			t.Errorf("code = %q", code)
		}
	})

	t.Run("POST sin CSRF responde 403", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/usuarios", `{}`,
			map[string]string{"Cookie": session.CookieSession + "=tok", "Content-Type": "application/json"})
		if resp.Status != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
	})
}

// --- POST /api/v1/admin/usuarios ---

func TestCreateUserHandler(t *testing.T) {
	want := sampleManagedUser()
	want.MustChangePassword = true
	identity := authorizedIdentity()
	svc := &fakeUsersService{
		createFn: func(_ context.Context, _ uuid.UUID, in CreateUserInput) (UserItem, error) {
			if in.Email != "carlos@ejemplo.com" {
				t.Errorf("correo recibido = %q", in.Email)
			}
			return want, nil
		},
	}
	srv := newAdminUsersServer(t, svc, identity)

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/usuarios",
		`{"firstName":"Carlos","lastName":"Ayudante","email":"carlos@ejemplo.com","phone":"612 345 678","roleId":"`+want.RoleID+`","password":"Cambio.2026"}`,
		adminUsersHeaders(t))

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
	assertNoCredentials(t, resp.Body, "Cambio.2026")
}

func TestCreateUserHandlerInvalidDTO(t *testing.T) {
	svc := &fakeUsersService{}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/usuarios",
		`{"firstName":"Carlos","lastName":"Ayudante","email":"carlos@ejemplo.com","roleId":"x","password":"Cambio.2026"}`,
		adminUsersHeaders(t))

	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if _, ok := details["phone"]; !ok {
		t.Errorf("details = %v, se esperaba phone", details)
	}
	if svc.creates != 0 {
		t.Error("un DTO inválido no debe llegar al servicio")
	}
	assertNoCredentials(t, resp.Body, "Cambio.2026")
}

func TestCreateUserHandlerConflict(t *testing.T) {
	svc := &fakeUsersService{
		createFn: func(context.Context, uuid.UUID, CreateUserInput) (UserItem, error) {
			return UserItem{}, apperr.Conflict(messageEmailInUse)
		},
	}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/usuarios",
		`{"firstName":"Carlos","lastName":"Ayudante","email":"carlos@ejemplo.com","phone":"612 345 678","roleId":"`+uuid.New().String()+`","password":"Cambio.2026"}`,
		adminUsersHeaders(t))

	if resp.Status != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", resp.Status, resp.Body)
	}
	if !strings.Contains(string(resp.Body), messageEmailInUse) {
		t.Errorf("mensaje inesperado: %s", resp.Body)
	}
}

// --- GET /api/v1/admin/usuarios/{id} ---

func TestGetUserHandler(t *testing.T) {
	want := sampleManagedUser()
	svc := &fakeUsersService{
		getFn: func(context.Context, uuid.UUID) (UserItem, error) { return want, nil },
	}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/usuarios/"+want.ID, "", adminUsersHeaders(t))
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if !strings.Contains(string(resp.Body), `"lastLoginAt":null`) ||
		!strings.Contains(string(resp.Body), `"lastLoginIp":null`) {
		t.Errorf("la ficha debe mostrar el último acceso como null: %s", resp.Body)
	}
	assertNoCredentials(t, resp.Body)
}

func TestGetUserHandlerNotFound(t *testing.T) {
	svc := &fakeUsersService{
		getFn: func(context.Context, uuid.UUID) (UserItem, error) {
			return UserItem{}, apperr.NotFound("La cuenta no existe")
		},
	}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/usuarios/"+uuid.New().String(), "", adminUsersHeaders(t))
	if resp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
	}
}

func TestGetUserHandlerInvalidID(t *testing.T) {
	svc := &fakeUsersService{}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/usuarios/no-es-un-uuid", "", adminUsersHeaders(t))
	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if _, ok := details["id"]; !ok {
		t.Errorf("details = %v, se esperaba id", details)
	}
}

// --- PATCH /api/v1/admin/usuarios/{id} ---

func TestUpdateUserHandler(t *testing.T) {
	want := sampleManagedUser()
	want.IsActive = false
	identity := authorizedIdentity()
	svc := &fakeUsersService{
		updateFn: func(_ context.Context, _ uuid.UUID, id uuid.UUID, in UpdateUserInput) (UserItem, error) {
			if in.IsActive == nil || *in.IsActive {
				t.Errorf("el servicio no recibió isActive=false: %+v", in)
			}
			if id.String() != want.ID {
				t.Errorf("id = %v", id)
			}
			return want, nil
		},
	}
	srv := newAdminUsersServer(t, svc, identity)

	resp := testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/usuarios/"+want.ID,
		`{"isActive":false}`, adminUsersHeaders(t))

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if svc.lastActor() != identity.UserID {
		t.Errorf("actor = %v, se esperaba %v", svc.lastActor(), identity.UserID)
	}
	if strings.Contains(strings.ToLower(string(resp.Body)), "$2a$") {
		t.Errorf("la respuesta filtra un hash: %s", resp.Body)
	}
}

func TestUpdateUserHandlerGuardConflict(t *testing.T) {
	svc := &fakeUsersService{
		updateFn: func(context.Context, uuid.UUID, uuid.UUID, UpdateUserInput) (UserItem, error) {
			return UserItem{}, apperr.Conflict(
				"No se puede dejar el panel sin administración",
				apperr.WithDetails(map[string]any{"reason": "admin_required"}),
			)
		},
	}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/usuarios/"+uuid.New().String(),
		`{"isActive":false}`, adminUsersHeaders(t))

	if resp.Status != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if details["reason"] != "admin_required" {
		t.Errorf("details = %v, se esperaba admin_required", details)
	}
}

// --- POST /api/v1/admin/usuarios/{id}/password ---

func TestResetUserPasswordHandler(t *testing.T) {
	identity := authorizedIdentity()
	svc := &fakeUsersService{
		resetFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, in ResetPasswordInput) error {
			if in.Password != "Nueva9#Aa" {
				t.Errorf("contraseña recibida = %q", in.Password)
			}
			return nil
		},
	}
	srv := newAdminUsersServer(t, svc, identity)

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/usuarios/"+uuid.New().String()+"/password",
		`{"password":"Nueva9#Aa"}`, adminUsersHeaders(t))

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if got := strings.TrimSpace(string(resp.Body)); got != `{"passwordReset":true}` {
		t.Errorf("cuerpo = %s", got)
	}
	if svc.lastActor() != identity.UserID {
		t.Errorf("actor = %v", svc.lastActor())
	}
	assertNoCredentials(t, resp.Body, "Nueva9#Aa")
}

func TestResetUserPasswordHandlerPolicyViolation(t *testing.T) {
	svc := &fakeUsersService{
		resetFn: func(context.Context, uuid.UUID, uuid.UUID, ResetPasswordInput) error {
			return apperr.Invalid("La contraseña no cumple la política de seguridad",
				apperr.WithDetails(map[string]any{"newPassword": "La contraseña debe incluir al menos una letra mayúscula"}))
		},
	}
	srv := newAdminUsersServer(t, svc, authorizedIdentity())

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/usuarios/"+uuid.New().String()+"/password",
		`{"password":"sinmayusculas1!"}`, adminUsersHeaders(t))

	if resp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
	}
	_, details := decodeErrorEnvelope(t, resp.Body)
	if _, ok := details["newPassword"]; !ok {
		t.Errorf("details = %v, se esperaba newPassword", details)
	}
	assertNoCredentials(t, resp.Body, "sinmayusculas1!")
}

// --- FR-013: no existe borrado de cuentas ---

func TestNoDeleteUserRoute(t *testing.T) {
	// Comprobación estática: el handler de cuentas no registra ningún DELETE.
	source, err := os.ReadFile("handler_users.go")
	if err != nil {
		t.Fatalf("leer handler_users.go: %v", err)
	}
	if strings.Contains(string(source), "MethodDelete") {
		t.Fatal("handler_users.go no debe registrar ninguna ruta de borrado (FR-013)")
	}

	// Comprobación en tiempo de ejecución: DELETE no está publicado.
	srv := newAdminUsersServer(t, &fakeUsersService{}, authorizedIdentity())
	headers := adminUsersHeaders(t)
	for _, path := range []string{"/api/v1/admin/usuarios", "/api/v1/admin/usuarios/" + uuid.New().String()} {
		resp := testutil.Do(t, srv, http.MethodDelete, path, "", headers)
		if resp.Status == http.StatusOK || resp.Status == http.StatusNoContent {
			t.Fatalf("DELETE %s está publicado: %d (FR-013)", path, resp.Status)
		}
		if resp.Status != http.StatusMethodNotAllowed {
			t.Fatalf("DELETE %s respondió %d, se esperaba 405 (%s)", path, resp.Status, resp.Body)
		}
	}
}
