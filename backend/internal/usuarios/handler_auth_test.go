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
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// fakeAccessService guioniza el servicio de acceso para probar el handler
// aislado sin Redis ni PostgreSQL.
type fakeAccessService struct {
	loginFn  func(ctx context.Context, in LoginInput, ip string) (SessionUser, []*http.Cookie, error)
	logoutFn func(ctx context.Context, token string) ([]*http.Cookie, error)
	changeFn func(ctx context.Context, identity session.Identity, currentToken string, in ChangePasswordInput) error
}

func (f *fakeAccessService) Login(ctx context.Context, in LoginInput, ip string) (SessionUser, []*http.Cookie, error) {
	if f.loginFn != nil {
		return f.loginFn(ctx, in, ip)
	}
	return SessionUser{}, nil, nil
}

func (f *fakeAccessService) Logout(ctx context.Context, token string) ([]*http.Cookie, error) {
	if f.logoutFn != nil {
		return f.logoutFn(ctx, token)
	}
	return nil, nil
}

func (f *fakeAccessService) ChangeMyPassword(ctx context.Context, identity session.Identity, currentToken string, in ChangePasswordInput) error {
	if f.changeFn != nil {
		return f.changeFn(ctx, identity, currentToken, in)
	}
	return nil
}

// injectIdentity deja la identidad indicada en el contexto, como haría authn en
// la cadena real. Permite probar el handler sin arrastrar los middlewares (los
// cubre su propia prueba, T227).
func injectIdentity(identity session.Identity) httpserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(session.ContextWithIdentity(r.Context(), identity)))
		})
	}
}

// handlerFakeStore resuelve cualquier token no vacío con la sesión indicada.
type handlerFakeStore struct{ session.Session }

func (s handlerFakeStore) Create(context.Context, uuid.UUID) (string, error) { return "token", nil }
func (s handlerFakeStore) Resolve(_ context.Context, token string) (session.Session, error) {
	if token == "" {
		return session.Session{}, session.ErrSessionNotFound
	}
	return s.Session, nil
}
func (s handlerFakeStore) Revoke(context.Context, string) error                      { return nil }
func (s handlerFakeStore) RevokeUser(context.Context, uuid.UUID) error               { return nil }
func (s handlerFakeStore) RevokeUserExcept(context.Context, uuid.UUID, string) error { return nil }

// handlerFakeResolver devuelve siempre la identidad indicada.
type handlerFakeResolver struct{ identity session.Identity }

func (r handlerFakeResolver) Resolve(context.Context, uuid.UUID) (session.Identity, error) {
	return r.identity, nil
}

// newAccessServer publica la superficie de acceso con los middlewares de sesión
// indicados (vacío = sin authn, para las rutas públicas).
func newAccessServer(t *testing.T, svc AccessService, sessionMws ...httpserver.Middleware) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	RegisterPublic(root, NewHandler(HandlerDeps{Access: svc, Logger: logger}), PublicDeps{Session: sessionMws})
	return testutil.NewServer(t, mux, logger)
}

// decodeErrorEnvelope lee code y details del sobre de error.
func decodeErrorEnvelope(t *testing.T, body []byte) (code string, details map[string]any) {
	t.Helper()
	var env struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("cuerpo no es sobre de error: %v (%q)", err, body)
	}
	return env.Error.Code, env.Error.Details
}

// assertNoCredentials comprueba que la respuesta no filtra contraseñas (FR-003):
// ni un hash bcrypt, ni el valor en claro que se envió. Los nombres de campo del
// contrato (p. ej. details.password, que señala qué corregir) no son credenciales.
func assertNoCredentials(t *testing.T, body []byte, secrets ...string) {
	t.Helper()
	lower := strings.ToLower(string(body))
	for _, marker := range []string{"passwordhash", "password_hash", "$2a$", "$2b$", "$2y$"} {
		if strings.Contains(lower, marker) {
			t.Errorf("la respuesta filtra un hash de contraseña (%q): %s", marker, body)
		}
	}
	for _, secret := range secrets {
		if secret != "" && strings.Contains(string(body), secret) {
			t.Errorf("la respuesta filtra una contraseña en claro: %s", body)
		}
	}
}

func sampleSessionUser() SessionUser {
	return SessionUser{
		ID:                 uuid.New().String(),
		Email:              "ana@ejemplo.com",
		FirstName:          "Ana",
		LastName:           "Responsable",
		Phone:              "+34 600 000 000",
		RoleID:             uuid.New().String(),
		RoleName:           "Administrador",
		Permissions:        []string{PermissionAdminUsersRoles},
		MustChangePassword: false,
	}
}

func TestHandlerLoginSuccess(t *testing.T) {
	want := sampleSessionUser()
	cfg := session.NewCookieConfig(false, time.Hour)
	svc := &fakeAccessService{
		loginFn: func(_ context.Context, _ LoginInput, _ string) (SessionUser, []*http.Cookie, error) {
			return want, []*http.Cookie{session.NewSessionCookie(cfg, "tok"), session.NewCSRFCookie(cfg, "csrf")}, nil
		},
	}
	srv := newAccessServer(t, svc)

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/login",
		`{"email":"ana@ejemplo.com","password":"Semilla.2026"}`,
		map[string]string{"Content-Type": "application/json"})

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	body, _ := json.Marshal(want)
	if got := strings.TrimSpace(string(resp.Body)); got != string(body) {
		t.Errorf("cuerpo = %s, se esperaba el DTO directo %s", got, body)
	}
	cookies := strings.Join(resp.Header.Values("Set-Cookie"), "\n")
	if !strings.Contains(cookies, "ss_session=") || !strings.Contains(cookies, "csrf_token=") {
		t.Errorf("faltan las cookies de sesión/CSRF: %q", cookies)
	}
	assertNoCredentials(t, resp.Body, "Semilla.2026")
}

func TestLoginRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantDetail string
	}{
		{name: "json mal formado", body: `{`},
		{name: "sin correo", body: `{"password":"Semilla.2026"}`, wantDetail: "email"},
		{name: "sin contraseña", body: `{"email":"ana@ejemplo.com"}`, wantDetail: "password"},
		{name: "correo con formato inválido", body: `{"email":"malo","password":"Semilla.2026"}`, wantDetail: "email"},
		{name: "campo desconocido", body: `{"email":"ana@ejemplo.com","password":"x","extra":1}`},
		{name: "dos objetos json", body: `{"email":"ana@ejemplo.com","password":"x"} {}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newAccessServer(t, &fakeAccessService{})
			resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/login", tt.body,
				map[string]string{"Content-Type": "application/json"})

			if resp.Status != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
			}
			code, details := decodeErrorEnvelope(t, resp.Body)
			if code != "invalid" {
				t.Errorf("code = %q, se esperaba invalid", code)
			}
			if tt.wantDetail != "" {
				if _, ok := details[tt.wantDetail]; !ok {
					t.Errorf("details = %v, se esperaba la clave %q", details, tt.wantDetail)
				}
			}
			assertNoCredentials(t, resp.Body, "Semilla.2026")
		})
	}
}

func TestLoginServiceErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantCode       string
		wantRetryAfter string
	}{
		{
			name:       "credenciales incorrectas",
			err:        apperr.Unauthenticated("Correo o contraseña incorrectos"),
			wantStatus: http.StatusUnauthorized,
			wantCode:   "unauthenticated",
		},
		{
			name:       "cuenta desactivada",
			err:        apperr.Forbidden("Ese acceso está desactivado"),
			wantStatus: http.StatusForbidden,
			wantCode:   "forbidden",
		},
		{
			name:           "bloqueo por intentos",
			err:            apperr.RateLimited("Demasiados intentos fallidos", apperr.WithRetryAfter(900)),
			wantStatus:     http.StatusTooManyRequests,
			wantCode:       "rate_limited",
			wantRetryAfter: "900",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeAccessService{
				loginFn: func(context.Context, LoginInput, string) (SessionUser, []*http.Cookie, error) {
					return SessionUser{}, nil, tt.err
				},
			}
			srv := newAccessServer(t, svc)
			resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/login",
				`{"email":"ana@ejemplo.com","password":"malo"}`,
				map[string]string{"Content-Type": "application/json"})

			if resp.Status != tt.wantStatus {
				t.Fatalf("status = %d, se esperaba %d (%s)", resp.Status, tt.wantStatus, resp.Body)
			}
			code, _ := decodeErrorEnvelope(t, resp.Body)
			if code != tt.wantCode {
				t.Errorf("code = %q, se esperaba %q", code, tt.wantCode)
			}
			if tt.wantRetryAfter != "" {
				if got := resp.Header.Get("Retry-After"); got != tt.wantRetryAfter {
					t.Errorf("Retry-After = %q, se esperaba %q", got, tt.wantRetryAfter)
				}
			}
			assertNoCredentials(t, resp.Body, "malo")
		})
	}
}

func TestLogoutClearsCookies(t *testing.T) {
	cfg := session.NewCookieConfig(false, time.Hour)
	svc := &fakeAccessService{
		logoutFn: func(context.Context, string) ([]*http.Cookie, error) {
			return []*http.Cookie{session.ClearSessionCookie(cfg), session.ClearCSRFCookie(cfg)}, nil
		},
	}
	srv := newAccessServer(t, svc)

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/logout", "", nil)

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
	}
	if got := strings.TrimSpace(string(resp.Body)); got != `{"loggedOut":true}` {
		t.Errorf("cuerpo = %s, se esperaba {\"loggedOut\":true}", got)
	}
	cookies := strings.Join(resp.Header.Values("Set-Cookie"), "\n")
	if !strings.Contains(cookies, "ss_session=;") && !strings.Contains(cookies, "ss_session=") {
		t.Errorf("logout no borra la cookie de sesión: %q", cookies)
	}
	for _, name := range []string{"ss_session", "csrf_token"} {
		if !strings.Contains(cookies, name+"=") {
			t.Errorf("logout no borra la cookie %s: %q", name, cookies)
		}
	}
}

func TestGetSession(t *testing.T) {
	t.Run("sin identidad responde 401", func(t *testing.T) {
		srv := newAccessServer(t, &fakeAccessService{})
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/auth/session", "", nil)

		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
		code, _ := decodeErrorEnvelope(t, resp.Body)
		if code != "unauthenticated" {
			t.Errorf("code = %q, se esperaba unauthenticated", code)
		}
	})

	t.Run("con identidad devuelve el DTO de sesión", func(t *testing.T) {
		identity := session.Identity{
			UserID:      uuid.New(),
			Email:       "ana@ejemplo.com",
			FirstName:   "Ana",
			LastName:    "Responsable",
			Phone:       "+34 600 000 000",
			RoleID:      uuid.New(),
			RoleName:    "Administrador",
			Permissions: []string{PermissionAdminUsersRoles},
		}
		srv := newAccessServer(t, &fakeAccessService{}, injectIdentity(identity))

		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/auth/session", "", nil)

		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		want, _ := json.Marshal(sessionUserFromIdentity(identity))
		if got := strings.TrimSpace(string(resp.Body)); got != string(want) {
			t.Errorf("cuerpo = %s, se esperaba %s", got, want)
		}
		assertNoCredentials(t, resp.Body)
	})
}

func TestChangePassword(t *testing.T) {
	identity := session.Identity{UserID: uuid.New(), Email: "ana@ejemplo.com"}

	t.Run("sin identidad responde 401", func(t *testing.T) {
		srv := newAccessServer(t, &fakeAccessService{})
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/password",
			`{"currentPassword":"vieja","newPassword":"Nueva.2026"}`,
			map[string]string{"Content-Type": "application/json"})
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("contraseña nueva que no cumple la política", func(t *testing.T) {
		srv := newAccessServer(t, &fakeAccessService{}, injectIdentity(identity))
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/password",
			`{"currentPassword":"vieja","newPassword":"corta"}`,
			map[string]string{"Content-Type": "application/json"})
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
		_, details := decodeErrorEnvelope(t, resp.Body)
		if _, ok := details["newPassword"]; !ok {
			t.Errorf("details = %v, se esperaba newPassword", details)
		}
	})

	t.Run("contraseña actual incorrecta", func(t *testing.T) {
		svc := &fakeAccessService{
			changeFn: func(context.Context, session.Identity, string, ChangePasswordInput) error {
				return apperr.Invalid("Tu contraseña actual no coincide", apperr.WithDetails(map[string]any{
					"currentPassword": "La contraseña actual no es correcta",
				}))
			},
		}
		srv := newAccessServer(t, svc, injectIdentity(identity))
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/password",
			`{"currentPassword":"mala","newPassword":"Nueva.2026"}`,
			map[string]string{"Content-Type": "application/json"})
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
		_, details := decodeErrorEnvelope(t, resp.Body)
		if _, ok := details["currentPassword"]; !ok {
			t.Errorf("details = %v, se esperaba currentPassword", details)
		}
	})

	t.Run("cambio correcto", func(t *testing.T) {
		var gotToken string
		svc := &fakeAccessService{
			changeFn: func(_ context.Context, _ session.Identity, currentToken string, _ ChangePasswordInput) error {
				gotToken = currentToken
				return nil
			},
		}
		srv := newAccessServer(t, svc, injectIdentity(identity))
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/password",
			`{"currentPassword":"vieja","newPassword":"Nueva.2026"}`,
			map[string]string{"Content-Type": "application/json", "Cookie": "ss_session=tok"})
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if got := strings.TrimSpace(string(resp.Body)); got != `{"passwordChanged":true}` {
			t.Errorf("cuerpo = %s", got)
		}
		if gotToken != "tok" {
			t.Errorf("el service recibió el token %q, se esperaba tok", gotToken)
		}
	})
}

// TestRegisterAdminAppliesPasswordGuard comprueba que el guard de cambio de
// contraseña se monta SOLO en el grupo /api/v1/admin (entre authn y authz) y que
// las rutas blanqueadas /auth/session, /auth/logout y /auth/password responden a
// una cuenta con mustChangePassword.
func TestRegisterAdminAppliesPasswordGuard(t *testing.T) {
	logger, _ := testutil.NewLogger()

	newServer := func(t *testing.T, identity session.Identity) *httptest.Server {
		t.Helper()
		store := handlerFakeStore{Session: session.Session{
			UserID:            identity.UserID,
			AbsoluteExpiresAt: time.Now().Add(time.Hour),
		}}
		resolver := handlerFakeResolver{identity: identity}
		svc := &fakeAccessService{}

		mux := http.NewServeMux()
		root := httpserver.NewMuxRegistrar(mux)
		// Rutas de sesión sin el guard (authn → CSRF), como en producción.
		RegisterPublic(root, NewHandler(HandlerDeps{Access: svc, Logger: logger}), PublicDeps{
			Session: []httpserver.Middleware{middleware.Authn(store, resolver, logger)},
		})
		// Grupo de panel con la cadena aprobada (authn → guard → authz → CSRF).
		RegisterAdmin(root, AdminDeps{
			Module: PermissionAdminUsersRoles,
			Deps: middleware.AdminDeps{
				Sessions:   store,
				Resolver:   resolver,
				CSRFSecret: "secret",
				Logger:     logger,
			},
			Routes: func(r httpserver.Registrar) {
				r.Handle(http.MethodGet, "/probe", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
		})
		return testutil.NewServer(t, mux, logger)
	}

	authedHeaders := map[string]string{"Cookie": "ss_session=tok"}

	t.Run("sin sesión el panel responde 401", func(t *testing.T) {
		srv := newServer(t, session.Identity{UserID: uuid.New()})
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/probe", "", nil)
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("mustChangePassword bloquea el panel con su reason", func(t *testing.T) {
		srv := newServer(t, session.Identity{
			UserID:             uuid.New(),
			Permissions:        []string{PermissionAdminUsersRoles},
			MustChangePassword: true,
		})
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/probe", "", authedHeaders)
		if resp.Status != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		code, details := decodeErrorEnvelope(t, resp.Body)
		if code != "forbidden" {
			t.Errorf("code = %q, se esperaba forbidden", code)
		}
		if details["reason"] != "password_change_required" {
			t.Errorf("details.reason = %v, se esperaba password_change_required", details["reason"])
		}
	})

	t.Run("con permiso y sin cambio pendiente el panel responde", func(t *testing.T) {
		srv := newServer(t, session.Identity{
			UserID:      uuid.New(),
			Permissions: []string{PermissionAdminUsersRoles},
		})
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/probe", "", authedHeaders)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("las rutas blanqueadas responden con mustChangePassword", func(t *testing.T) {
		srv := newServer(t, session.Identity{
			UserID:             uuid.New(),
			Email:              "ana@ejemplo.com",
			Phone:              "+34 600 000 000",
			Permissions:        []string{PermissionAdminUsersRoles},
			MustChangePassword: true,
		})

		sessionResp := testutil.Do(t, srv, http.MethodGet, "/api/v1/auth/session", "", authedHeaders)
		if sessionResp.Status != http.StatusOK {
			t.Errorf("/auth/session status = %d, se esperaba 200 (%s)", sessionResp.Status, sessionResp.Body)
		}

		logoutResp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/logout", "", authedHeaders)
		if logoutResp.Status != http.StatusOK {
			t.Errorf("/auth/logout status = %d, se esperaba 200 (%s)", logoutResp.Status, logoutResp.Body)
		}

		passwordResp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/password",
			`{"currentPassword":"vieja","newPassword":"Nueva.2026"}`,
			map[string]string{"Cookie": "ss_session=tok", "Content-Type": "application/json"})
		if passwordResp.Status != http.StatusOK {
			t.Errorf("/auth/password status = %d, se esperaba 200 (%s)", passwordResp.Status, passwordResp.Body)
		}
	})
}

// --- Inicialización única (T229/T230) ---

// fakeSetupService guioniza el servicio de inicialización para probar el handler
// aislado, y cuenta cuántas veces se le llama (un token inválido o un DTO
// inválido NO deben llegar al servicio).
type fakeSetupService struct {
	mu           sync.Mutex
	calls        int
	initializeFn func(ctx context.Context, in InitializeInput) (UserItem, error)
}

func (f *fakeSetupService) Initialize(ctx context.Context, in InitializeInput) (UserItem, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	if f.initializeFn != nil {
		return f.initializeFn(ctx, in)
	}
	return UserItem{}, nil
}

func (f *fakeSetupService) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

const setupToken = "secreto-de-despliegue"

// setupBody es un InitializeInput válido (FR-009/FR-010).
const setupBody = `{"firstName":"Ana","lastName":"Responsable","email":"ana@ejemplo.com","phone":"+34 612 345 678","password":"Semilla.2026"}`

func sampleUserItem() UserItem {
	return UserItem{
		ID:                 uuid.New().String(),
		Email:              "ana@ejemplo.com",
		FirstName:          "Ana",
		LastName:           "Responsable",
		Phone:              "+34 612 345 678",
		RoleID:             uuid.New().String(),
		RoleName:           "Administrador",
		IsActive:           true,
		MustChangePassword: false,
		CreatedAt:          time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC),
	}
}

// newSetupServer publica la superficie pública con la inicialización cableada.
func newSetupServer(t *testing.T, setup SetupService, mws ...httpserver.Middleware) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	RegisterPublic(root, NewHandler(HandlerDeps{
		Access:     &fakeAccessService{},
		Setup:      setup,
		SetupToken: setupToken,
		Logger:     logger,
	}), PublicDeps{Setup: mws})
	return testutil.NewServer(t, mux, logger)
}

func setupHeaders() map[string]string {
	return map[string]string{
		"Content-Type":   "application/json",
		setupTokenHeader: setupToken,
	}
}

func TestInitializeHandlerSuccess(t *testing.T) {
	want := sampleUserItem()
	svc := &fakeSetupService{
		initializeFn: func(_ context.Context, in InitializeInput) (UserItem, error) {
			if in.Email != "ana@ejemplo.com" {
				t.Errorf("el servicio recibió el correo %q", in.Email)
			}
			return want, nil
		},
	}
	srv := newSetupServer(t, svc)

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", setupBody, setupHeaders())

	if resp.Status != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201 (%s)", resp.Status, resp.Body)
	}
	body, _ := json.Marshal(want)
	if got := strings.TrimSpace(string(resp.Body)); got != string(body) {
		t.Errorf("cuerpo = %s, se esperaba el UserItem %s", got, body)
	}
	assertNoCredentials(t, resp.Body, "Semilla.2026")
	if svc.callCount() != 1 {
		t.Errorf("el servicio se llamó %d veces, se esperaba 1", svc.callCount())
	}
}

func TestInitializeHandlerToken(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantCode   string
	}{
		{name: "sin cabecera", token: "", wantStatus: http.StatusUnauthorized, wantCode: "unauthenticated"},
		{name: "token erróneo", token: "otro-token", wantStatus: http.StatusForbidden, wantCode: "forbidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeSetupService{}
			srv := newSetupServer(t, svc)

			headers := map[string]string{"Content-Type": "application/json"}
			if tt.token != "" {
				headers[setupTokenHeader] = tt.token
			}
			resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", setupBody, headers)

			if resp.Status != tt.wantStatus {
				t.Fatalf("status = %d, se esperaba %d (%s)", resp.Status, tt.wantStatus, resp.Body)
			}
			code, _ := decodeErrorEnvelope(t, resp.Body)
			if code != tt.wantCode {
				t.Errorf("code = %q, se esperaba %q", code, tt.wantCode)
			}
			// El token esperado NUNCA viaja en la respuesta (RG13).
			if strings.Contains(string(resp.Body), setupToken) {
				t.Errorf("la respuesta revela el token esperado: %s", resp.Body)
			}
			if svc.callCount() != 0 {
				t.Error("un token inválido no debe llegar al servicio")
			}
		})
	}
}

func TestInitializeHandlerRepeated(t *testing.T) {
	svc := &fakeSetupService{
		initializeFn: func(context.Context, InitializeInput) (UserItem, error) {
			return UserItem{}, apperr.Conflict(messageAlreadyInitialized)
		},
	}
	srv := newSetupServer(t, svc)

	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", setupBody, setupHeaders())

	if resp.Status != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", resp.Status, resp.Body)
	}
	code, _ := decodeErrorEnvelope(t, resp.Body)
	if code != "conflict" {
		t.Errorf("code = %q, se esperaba conflict", code)
	}
	if !strings.Contains(string(resp.Body), "La inicialización ya se hizo y no puede repetirse") {
		t.Errorf("mensaje inesperado: %s", resp.Body)
	}
	assertNoCredentials(t, resp.Body, "Semilla.2026")
}

func TestInitializeHandlerInvalidDTO(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantDetail string
	}{
		{name: "sin teléfono", body: `{"firstName":"Ana","lastName":"R","email":"ana@ejemplo.com","password":"Semilla.2026"}`, wantDetail: "phone"},
		{name: "correo inválido", body: `{"firstName":"Ana","lastName":"R","email":"malo","phone":"612345678","password":"Semilla.2026"}`, wantDetail: "email"},
		{name: "sin contraseña", body: `{"firstName":"Ana","lastName":"R","email":"ana@ejemplo.com","phone":"612345678"}`, wantDetail: "password"},
		{name: "campo desconocido", body: `{"firstName":"Ana","lastName":"R","email":"ana@ejemplo.com","phone":"612345678","password":"Semilla.2026","extra":1}`},
		{name: "json mal formado", body: `{`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeSetupService{}
			srv := newSetupServer(t, svc)

			resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", tt.body, setupHeaders())

			if resp.Status != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
			}
			code, details := decodeErrorEnvelope(t, resp.Body)
			if code != "invalid" {
				t.Errorf("code = %q, se esperaba invalid", code)
			}
			if tt.wantDetail != "" {
				if _, ok := details[tt.wantDetail]; !ok {
					t.Errorf("details = %v, se esperaba la clave %q", details, tt.wantDetail)
				}
			}
			if svc.callCount() != 0 {
				t.Error("un DTO inválido no debe llegar al servicio")
			}
			assertNoCredentials(t, resp.Body, "Semilla.2026")
		})
	}
}

func TestInitializeHandlerRateLimited(t *testing.T) {
	logger, _ := testutil.NewLogger()
	svc := &fakeSetupService{}
	srv := newSetupServer(t, svc, middleware.RateLimit(middleware.RateLimitConfig{
		Limit: 2,
		Paths: middleware.DefaultRateLimitPaths(),
	}, logger))

	statuses := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", setupBody, setupHeaders())
		statuses = append(statuses, resp.Status)
		if resp.Status == http.StatusTooManyRequests {
			code, _ := decodeErrorEnvelope(t, resp.Body)
			if code != "rate_limited" {
				t.Errorf("code = %q, se esperaba rate_limited", code)
			}
			if resp.Header.Get("Retry-After") == "" {
				t.Error("el 429 no trae Retry-After")
			}
		}
	}

	want := []int{http.StatusCreated, http.StatusCreated, http.StatusTooManyRequests}
	for i, status := range statuses {
		if status != want[i] {
			t.Fatalf("statuses = %v, se esperaba %v", statuses, want)
		}
	}
}

// TestSetupRouteOnlyPublishedWithService comprueba que la ruta de inicialización
// no existe cuando el Handler no trae el servicio cableado.
func TestSetupRouteOnlyPublishedWithService(t *testing.T) {
	srv := newAccessServer(t, &fakeAccessService{})
	resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", setupBody, setupHeaders())
	if resp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 sin servicio de inicialización (%s)", resp.Status, resp.Body)
	}
}
