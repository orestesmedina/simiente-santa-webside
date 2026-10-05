package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
)

// --- captura de logs en memoria (sin dependencias) ---

type memStore struct {
	mu      sync.Mutex
	records []memRecord
}

type memRecord struct {
	level slog.Level
	msg   string
	attrs map[string]any
}

type memHandler struct {
	store *memStore
	attrs []slog.Attr
}

func newMemHandler(store *memStore) *memHandler { return &memHandler{store: store} }

func (h *memHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *memHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	h.store.records = append(h.store.records, memRecord{level: r.Level, msg: r.Message, attrs: attrs})
	return nil
}

func (h *memHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &memHandler{store: h.store, attrs: merged}
}

func (h *memHandler) WithGroup(string) slog.Handler { return h }

func (s *memStore) find(level slog.Level) (memRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range s.records {
		if rec.level == level {
			return rec, true
		}
	}
	return memRecord{}, false
}

func captureLogger() (*slog.Logger, *memStore) {
	store := &memStore{}
	return slog.New(newMemHandler(store)), store
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(discardWriter{}, nil))
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// --- helpers ---

func muxWithHandler(h http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	httpserver.NewMuxRegistrar(mux).Handle(http.MethodGet, "/healthz", h)
	return mux
}

func decodeEnvelope(t *testing.T, body []byte) (code, message string, details map[string]any) {
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
	return env.Error.Code, env.Error.Message, env.Error.Details
}

// --- pruebas ---

func TestChainOrderIsApproved(t *testing.T) {
	var order []string
	record := func(name string) httpserver.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	mux := muxWithHandler(func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})
	handler := httpserver.NewHandler(mux, discardLogger(),
		record("request-id"), record("recover"), record("logging"), record("CORS"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	want := []string{"request-id", "recover", "logging", "CORS", "handler"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("orden de la cadena = %v, se esperaba %v", order, want)
	}
}

func TestRequestIDRespectedAndGenerated(t *testing.T) {
	mux := muxWithHandler(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(httpserver.RequestIDFromContext(r.Context())))
	})
	handler := httpserver.NewHandler(mux, discardLogger(), RequestID)

	t.Run("respeta el X-Request-ID del cliente", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set(HeaderRequestID, "cliente-123")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get(HeaderRequestID); got != "cliente-123" {
			t.Errorf("cabecera X-Request-ID = %q, se esperaba cliente-123", got)
		}
		if rec.Body.String() != "cliente-123" {
			t.Errorf("contexto = %q, se esperaba cliente-123", rec.Body.String())
		}
	})

	t.Run("genera uno si el cliente no lo envía", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

		id := rec.Header().Get(HeaderRequestID)
		if id == "" {
			t.Fatal("no se generó X-Request-ID")
		}
		if rec.Body.String() != id {
			t.Errorf("el contexto no tiene el id generado: cuerpo=%q cabecera=%q", rec.Body.String(), id)
		}
	})

	t.Run("descarta un X-Request-ID con caracteres raros", func(t *testing.T) {
		const weird = "abc<script>\n\"';"
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set(HeaderRequestID, weird)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		got := rec.Header().Get(HeaderRequestID)
		if got == weird {
			t.Errorf("se aceptó el X-Request-ID inválido del cliente: %q", got)
		}
		if got == "" || rec.Body.String() != got {
			t.Errorf("no se generó un id propio: cabecera=%q cuerpo=%q", got, rec.Body.String())
		}
	})

	t.Run("descarta un X-Request-ID demasiado largo", func(t *testing.T) {
		tooLong := strings.Repeat("a", 129)
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set(HeaderRequestID, tooLong)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		got := rec.Header().Get(HeaderRequestID)
		if got == tooLong {
			t.Error("se aceptó un X-Request-ID de 129 caracteres")
		}
		if len(got) > 128 || got == "" {
			t.Errorf("el id propio no es válido: %q (len=%d)", got, len(got))
		}
	})
}

func TestGeneratedRequestIDReachesLog(t *testing.T) {
	logger, store := captureLogger()
	mux := muxWithHandler(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := httpserver.NewHandler(mux, logger, RequestID, Logging(logger))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	record, ok := store.find(slog.LevelInfo)
	if !ok {
		t.Fatal("logging no emitió el registro de acceso")
	}
	if record.attrs["request_id"] != rec.Header().Get(HeaderRequestID) {
		t.Errorf("request_id del log = %v, se esperaba %q", record.attrs["request_id"], rec.Header().Get(HeaderRequestID))
	}
}

func TestRecoverTurnsPanicInto500(t *testing.T) {
	logger, store := captureLogger()
	mux := muxWithHandler(func(http.ResponseWriter, *http.Request) { panic("boom interno") })
	handler := httpserver.NewHandler(mux, logger, Recover(logger))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500", rec.Code)
	}
	code, message, _ := decodeEnvelope(t, rec.Body.Bytes())
	if code != "internal" {
		t.Errorf("code = %q, se esperaba internal", code)
	}
	if message != "Error interno del servidor" {
		t.Errorf("message = %q, se esperaba el genérico", message)
	}
	if strings.Contains(rec.Body.String(), "boom interno") {
		t.Error("la respuesta filtra el panic")
	}

	record, ok := store.find(slog.LevelError)
	if !ok {
		t.Fatal("el panic no quedó registrado en el log")
	}
	detail, _ := record.attrs["error"].(string)
	if !strings.Contains(detail, "boom interno") {
		t.Errorf("el log no contiene el detalle del panic: %v", record.attrs)
	}
}

func TestWriteErrorUsesPerRequestLoggerWithContext(t *testing.T) {
	logger, store := captureLogger()
	mux := muxWithHandler(func(w http.ResponseWriter, r *http.Request) {
		httpserver.WriteError(r.Context(), w, logger, apperr.NotFound("Recurso no encontrado"))
	})
	handler := httpserver.NewHandler(mux, logger, RequestID, Logging(logger))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(HeaderRequestID, "req-ctx-456")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	record, ok := store.find(slog.LevelWarn)
	if !ok {
		t.Fatal("WriteError no emitió el registro del error de dominio")
	}
	want := map[string]any{
		"request_id": "req-ctx-456",
		"method":     "GET",
		"path":       "/healthz",
		"code":       "not_found",
	}
	for key, value := range want {
		if record.attrs[key] != value {
			t.Errorf("campo %q = %v, se esperaba %v", key, record.attrs[key], value)
		}
	}
}

func TestChainWiresTransversalMiddlewares(t *testing.T) {
	const allowedOrigin = "http://localhost:5173"
	logger, store := captureLogger()
	mux := muxWithHandler(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := httpserver.NewHandler(mux, logger, Chain(logger, []string{allowedOrigin})...)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(headerOrigin, allowedOrigin)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(HeaderRequestID); got == "" {
		t.Error("Chain no montó request-id (falta X-Request-ID)")
	}
	if got := rec.Header().Get(headerAllowOrigin); got != allowedOrigin {
		t.Errorf("Chain no montó CORS: Allow-Origin = %q", got)
	}
	if _, ok := store.find(slog.LevelInfo); !ok {
		t.Error("Chain no montó logging (sin registro de acceso)")
	}
}

func TestLoggingEmitsFiveFields(t *testing.T) {
	logger, store := captureLogger()
	mux := muxWithHandler(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := httpserver.NewHandler(mux, logger, RequestID, Logging(logger))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(HeaderRequestID, "req-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	record, ok := store.find(slog.LevelInfo)
	if !ok {
		t.Fatal("logging no emitió el registro de acceso")
	}
	want := map[string]any{
		"request_id": "req-1",
		"method":     "GET",
		"path":       "/healthz",
		"status":     int64(http.StatusOK),
	}
	for key, value := range want {
		if record.attrs[key] != value {
			t.Errorf("campo %q = %v, se esperaba %v", key, record.attrs[key], value)
		}
	}
	if _, ok := record.attrs["duration_ms"]; !ok {
		t.Errorf("falta el campo duration_ms: %v", record.attrs)
	}
}

func TestCORSPreflight(t *testing.T) {
	const allowedOrigin = "http://localhost:5173"
	logger := discardLogger()
	mux := muxWithHandler(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := httpserver.NewHandler(mux, logger, CORS([]string{allowedOrigin}))

	t.Run("origen permitido recibe sus cabeceras", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
		req.Header.Set(headerOrigin, allowedOrigin)
		req.Header.Set(headerRequestMethod, "POST")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, se esperaba 204", rec.Code)
		}
		if got := rec.Header().Get(headerAllowOrigin); got != allowedOrigin {
			t.Errorf("Allow-Origin = %q, se esperaba %q", got, allowedOrigin)
		}
		if got := rec.Header().Get(headerAllowOrigin); got == "*" {
			t.Error("Allow-Origin nunca puede ser * con credenciales")
		}
		if got := rec.Header().Get(headerAllowCredentials); got != "true" {
			t.Errorf("Allow-Credentials = %q, se esperaba true (las cookies exigen credenciales)", got)
		}
		if got := rec.Header().Get(headerAllowMethods); got != allowMethodsValue {
			t.Errorf("Allow-Methods = %q, se esperaba %q", got, allowMethodsValue)
		}
		for _, verb := range []string{"POST", "PATCH", "DELETE"} {
			if !strings.Contains(rec.Header().Get(headerAllowMethods), verb) {
				t.Errorf("Allow-Methods no concede %s, que la superficie de F2 usa: %q", verb, rec.Header().Get(headerAllowMethods))
			}
		}
		for _, header := range []string{"Content-Type", "X-CSRF-Token", "X-Request-ID"} {
			if !strings.Contains(rec.Header().Get(headerAllowHeaders), header) {
				t.Errorf("Allow-Headers no incluye %s: %q", header, rec.Header().Get(headerAllowHeaders))
			}
		}
		if !strings.Contains(rec.Header().Get(headerVary), headerOrigin) {
			t.Errorf("falta Vary: Origin: %q", rec.Header().Get(headerVary))
		}
	})

	t.Run("origen no permitido no recibe cabeceras", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
		req.Header.Set(headerOrigin, "http://malicioso.example")
		req.Header.Set(headerRequestMethod, "GET")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get(headerAllowOrigin); got != "" {
			t.Errorf("Allow-Origin = %q, se esperaba vacío", got)
		}
		if rec.Header().Get(headerAllowMethods) != "" {
			t.Error("se enviaron métodos CORS a un origen no permitido")
		}
		if got := rec.Header().Get(headerAllowCredentials); got != "" {
			t.Errorf("Allow-Credentials = %q, se esperaba vacío en un origen no permitido", got)
		}
	})

	t.Run("GET simple permitido llega al handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set(headerOrigin, allowedOrigin)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, se esperaba 200", rec.Code)
		}
		if got := rec.Header().Get(headerAllowOrigin); got != allowedOrigin {
			t.Errorf("Allow-Origin = %q, se esperaba %q", got, allowedOrigin)
		}
		if got := rec.Header().Get(headerAllowCredentials); got != "true" {
			t.Errorf("Allow-Credentials = %q, se esperaba true", got)
		}
		if got := rec.Header().Get(headerExposeHeaders); got != exposeHeadersValue {
			t.Errorf("Expose-Headers = %q, se esperaba %q", got, exposeHeadersValue)
		}
	})
}

// --- dobles de las dependencias de la cadena de sesión ---

// stubSessionStore implementa session.Store para las pruebas de authn.
type stubSessionStore struct {
	sess  session.Session
	err   error
	token string
}

func (s *stubSessionStore) Create(context.Context, uuid.UUID) (string, error) { return "", nil }

func (s *stubSessionStore) Resolve(_ context.Context, token string) (session.Session, error) {
	s.token = token
	return s.sess, s.err
}

func (s *stubSessionStore) Revoke(context.Context, string) error { return nil }

func (s *stubSessionStore) RevokeUser(context.Context, uuid.UUID) error { return nil }

func (s *stubSessionStore) RevokeUserExcept(context.Context, uuid.UUID, string) error { return nil }

// stubResolver implementa session.Resolver para las pruebas de authn.
type stubResolver struct {
	identity session.Identity
	err      error
	userID   uuid.UUID
}

func (r *stubResolver) Resolve(_ context.Context, userID uuid.UUID) (session.Identity, error) {
	r.userID = userID
	return r.identity, r.err
}

// stubRecorder implementa audit.Recorder para las pruebas de authz.
type stubRecorder struct {
	denials []audit.Denial
	err     error
}

func (r *stubRecorder) RecordDenied(_ context.Context, denial audit.Denial) error {
	r.denials = append(r.denials, denial)
	return r.err
}

// serve ejecuta un middleware sobre okHandler y devuelve la grabación.
func serve(mw httpserver.Middleware, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, req)
	return rec
}

// sessionRequest construye una petición con (o sin) la cookie de sesión.
func sessionRequest(method, path, token string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: session.CookieSession, Value: token})
	}
	return req
}

// csrfRequest construye una petición con la cookie y la cabecera CSRF.
func csrfRequest(method, path, cookieToken, headerToken string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if cookieToken != "" {
		req.AddCookie(&http.Cookie{Name: session.CookieCSRF, Value: cookieToken})
	}
	if headerToken != "" {
		req.Header.Set(session.HeaderCSRF, headerToken)
	}
	return req
}

// remoteRequest construye una petición con la IP de origen indicada.
func remoteRequest(method, path, remoteAddr string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = remoteAddr
	return req
}

// --- authn ---

func TestAuthnRejectsInvalidSessions(t *testing.T) {
	userID := uuid.New()
	live := session.Session{UserID: userID, AbsoluteExpiresAt: time.Now().Add(time.Hour)}

	tests := []struct {
		name        string
		token       string
		stored      session.Session
		storeErr    error
		resolverErr error
	}{
		{name: "sin cookie", token: "", stored: live},
		{name: "cookie inválida", token: "desconocido", storeErr: session.ErrSessionNotFound},
		{name: "sesión expirada", token: "caducada", stored: session.Session{UserID: userID, AbsoluteExpiresAt: time.Now().Add(-time.Minute)}},
		{name: "cuenta inactiva", token: "viva", stored: live, resolverErr: errors.New("cuenta inactiva")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &stubSessionStore{sess: tc.stored, err: tc.storeErr}
			resolver := &stubResolver{identity: session.Identity{UserID: userID}, err: tc.resolverErr}
			rec := serve(Authn(store, resolver, discardLogger()),
				sessionRequest(http.MethodGet, "/api/v1/admin/usuarios", tc.token))

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, se esperaba 401", rec.Code)
			}
			code, _, _ := decodeEnvelope(t, rec.Body.Bytes())
			if code != "unauthenticated" {
				t.Errorf("code = %q, se esperaba unauthenticated", code)
			}
		})
	}
}

func TestAuthnLeavesResolvedIdentityInContext(t *testing.T) {
	userID := uuid.New()
	identity := session.Identity{
		UserID:      userID,
		Email:       "ana@ejemplo.com",
		Permissions: []string{"admin_usuarios_roles"},
	}
	store := &stubSessionStore{sess: session.Session{UserID: userID, AbsoluteExpiresAt: time.Now().Add(time.Hour)}}
	resolver := &stubResolver{identity: identity}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := session.IdentityFromContext(r.Context())
		if !ok {
			t.Error("authn no dejó la identidad en el contexto")
			return
		}
		_, _ = w.Write([]byte(got.UserID.String()))
	})

	rec := httptest.NewRecorder()
	Authn(store, resolver, discardLogger())(handler).
		ServeHTTP(rec, sessionRequest(http.MethodGet, "/api/v1/admin/usuarios", "viva"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", rec.Code)
	}
	if rec.Body.String() != userID.String() {
		t.Errorf("identidad = %q, se esperaba %q", rec.Body.String(), userID.String())
	}
	if resolver.userID != userID {
		t.Errorf("resolver recibió %v, se esperaba %v", resolver.userID, userID)
	}
}

// --- authz ---

func TestAuthzByModuleAllowsWithPermission(t *testing.T) {
	identity := session.Identity{UserID: uuid.New(), Permissions: []string{"admin_usuarios_roles"}}
	mw := AuthzByModule("admin_usuarios_roles", &stubRecorder{}, discardLogger())

	rec := httptest.NewRecorder()
	withIdentity(identity, mw(okHandler())).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/usuarios", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", rec.Code)
	}
}

func TestAuthzByModuleDeniesAndRecordsWithoutPermission(t *testing.T) {
	userID := uuid.New()
	identity := session.Identity{UserID: userID, Permissions: []string{"otro_modulo"}}
	recorder := &stubRecorder{}
	mw := AuthzByModule("admin_usuarios_roles", recorder, discardLogger())

	rec := httptest.NewRecorder()
	withIdentity(identity, mw(okHandler())).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/admin/usuarios", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, se esperaba 403", rec.Code)
	}
	code, message, _ := decodeEnvelope(t, rec.Body.Bytes())
	if code != "forbidden" {
		t.Errorf("code = %q, se esperaba forbidden", code)
	}
	if message == "" {
		t.Error("el 403 no lleva mensaje claro")
	}

	if len(recorder.denials) != 1 {
		t.Fatalf("denegaciones registradas = %d, se esperaba 1", len(recorder.denials))
	}
	denial := recorder.denials[0]
	if denial.ActorUserID == nil || *denial.ActorUserID != userID {
		t.Errorf("actor = %v, se esperaba %v", denial.ActorUserID, userID)
	}
	if denial.Method != http.MethodPost || denial.Path != "/api/v1/admin/usuarios" {
		t.Errorf("método/ruta registrados = %s %s, se esperaba POST /api/v1/admin/usuarios", denial.Method, denial.Path)
	}
}

func TestAuthzByModuleWithoutIdentityIsUnauthenticated(t *testing.T) {
	mw := AuthzByModule("admin_usuarios_roles", &stubRecorder{}, discardLogger())
	rec := serve(mw, httptest.NewRequest(http.MethodGet, "/api/v1/admin/usuarios", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 sin identidad", rec.Code)
	}
}

// --- CSRF ---

func TestCSRF(t *testing.T) {
	const secret = "secreto-de-prueba"
	valid, err := session.NewCSRFToken(secret)
	if err != nil {
		t.Fatalf("generar token CSRF: %v", err)
	}
	mw := CSRF(secret, discardLogger())

	t.Run("método seguro pasa sin token", func(t *testing.T) {
		rec := serve(mw, httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200", rec.Code)
		}
	})

	t.Run("método inseguro sin token", func(t *testing.T) {
		rec := serve(mw, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
		assertCSRFForbidden(t, rec)
	})

	t.Run("método inseguro con cookie y cabecera desalineadas", func(t *testing.T) {
		other, err := session.NewCSRFToken(secret)
		if err != nil {
			t.Fatalf("generar segundo token CSRF: %v", err)
		}
		rec := serve(mw, csrfRequest(http.MethodPost, "/api/v1/auth/logout", valid, other))
		assertCSRFForbidden(t, rec)
	})

	t.Run("método inseguro con firma inválida", func(t *testing.T) {
		const forged = "nonce.firmaFalsa"
		rec := serve(mw, csrfRequest(http.MethodPost, "/api/v1/auth/logout", forged, forged))
		assertCSRFForbidden(t, rec)
	})

	t.Run("método inseguro con par válido pasa", func(t *testing.T) {
		rec := serve(mw, csrfRequest(http.MethodPost, "/api/v1/auth/logout", valid, valid))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200", rec.Code)
		}
	})
}

func assertCSRFForbidden(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, se esperaba 403", rec.Code)
	}
	code, _, _ := decodeEnvelope(t, rec.Body.Bytes())
	if code != "forbidden" {
		t.Errorf("code = %q, se esperaba forbidden", code)
	}
}

// --- rate-limit ---

func TestRateLimitBlocksAfterThreshold(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	mw := RateLimit(RateLimitConfig{
		Limit:  3,
		Window: time.Minute,
		Paths:  DefaultRateLimitPaths(),
		Now:    func() time.Time { return now },
	}, discardLogger())

	for i := 1; i <= 3; i++ {
		rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "203.0.113.7:5555"))
		if rec.Code != http.StatusOK {
			t.Fatalf("petición %d: status = %d, se esperaba 200", i, rec.Code)
		}
	}

	rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "203.0.113.7:5555"))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, se esperaba 429", rec.Code)
	}
	code, _, details := decodeEnvelope(t, rec.Body.Bytes())
	if code != "rate_limited" {
		t.Errorf("code = %q, se esperaba rate_limited", code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("falta la cabecera Retry-After")
	}
	if details["retryAfterSeconds"] == nil {
		t.Errorf("falta details.retryAfterSeconds: %v", details)
	}
}

func TestRateLimitWindowSlides(t *testing.T) {
	base := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	now := base
	mw := RateLimit(RateLimitConfig{
		Limit:  2,
		Window: time.Minute,
		Now:    func() time.Time { return now },
	}, discardLogger())

	for i := 1; i <= 2; i++ {
		if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "203.0.113.7:1")); rec.Code != http.StatusOK {
			t.Fatalf("petición %d: status = %d, se esperaba 200", i, rec.Code)
		}
	}
	if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "203.0.113.7:1")); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, se esperaba 429 dentro de la ventana", rec.Code)
	}

	now = base.Add(2 * time.Minute)
	if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "203.0.113.7:1")); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 al deslizar la ventana", rec.Code)
	}
}

func TestRateLimitIsPerIP(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	mw := RateLimit(RateLimitConfig{
		Limit:  1,
		Window: time.Minute,
		Now:    func() time.Time { return now },
	}, discardLogger())

	if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "203.0.113.7:1")); rec.Code != http.StatusOK {
		t.Fatalf("primera IP: status = %d, se esperaba 200", rec.Code)
	}
	if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "198.51.100.9:2")); rec.Code != http.StatusOK {
		t.Fatalf("segunda IP: status = %d, se esperaba 200", rec.Code)
	}
	if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", "203.0.113.7:3")); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("primera IP reincidente: status = %d, se esperaba 429", rec.Code)
	}
}

func TestRateLimitOnlyScopesListedPaths(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	mw := RateLimit(RateLimitConfig{
		Limit:  1,
		Window: time.Minute,
		Paths:  DefaultRateLimitPaths(),
		Now:    func() time.Time { return now },
	}, discardLogger())

	const ip = "203.0.113.7:1"
	if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", ip)); rec.Code != http.StatusOK {
		t.Fatalf("login: status = %d, se esperaba 200", rec.Code)
	}
	if rec := serve(mw, remoteRequest(http.MethodPost, "/api/v1/auth/login", ip)); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("login repetido: status = %d, se esperaba 429", rec.Code)
	}
	if rec := serve(mw, remoteRequest(http.MethodGet, "/api/v1/auth/session", ip)); rec.Code != http.StatusOK {
		t.Fatalf("ruta no listada: status = %d, se esperaba 200 (sin límite)", rec.Code)
	}
}

// --- orden de la cadena de panel ---

func TestAdminChainOrderIsApproved(t *testing.T) {
	const (
		secret = "secreto-de-prueba"
		module = "admin_usuarios_roles"
	)
	userID := uuid.New()
	live := session.Session{UserID: userID, AbsoluteExpiresAt: time.Now().Add(time.Hour)}

	build := func(identity session.Identity) http.Handler {
		chain := AdminChain(module, AdminDeps{
			Sessions:   &stubSessionStore{sess: live},
			Resolver:   &stubResolver{identity: identity},
			Recorder:   &stubRecorder{},
			CSRFSecret: secret,
			Logger:     discardLogger(),
		})
		if len(chain) != 4 {
			t.Fatalf("AdminChain = %d middlewares, se esperaban 4", len(chain))
		}
		handler := http.Handler(okHandler())
		for i := len(chain) - 1; i >= 0; i-- {
			handler = chain[i](handler)
		}
		return handler
	}

	valid, err := session.NewCSRFToken(secret)
	if err != nil {
		t.Fatalf("generar token CSRF: %v", err)
	}

	t.Run("authn va primero: sin cookie → 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		build(session.Identity{UserID: userID}).ServeHTTP(rec,
			csrfRequest(http.MethodPost, "/api/v1/admin/usuarios", valid, valid))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401", rec.Code)
		}
	})

	t.Run("el guard va antes de authz", func(t *testing.T) {
		rec := httptest.NewRecorder()
		build(session.Identity{UserID: userID, MustChangePassword: true}).ServeHTTP(rec,
			sessionRequest(http.MethodGet, "/api/v1/admin/usuarios", "viva"))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), passwordChangeRequiredReason) {
			t.Errorf("se esperaba el guard de cambio obligatorio: %s", rec.Body.String())
		}
	})

	t.Run("authz va antes de CSRF", func(t *testing.T) {
		rec := httptest.NewRecorder()
		build(session.Identity{UserID: userID}).ServeHTTP(rec,
			sessionRequest(http.MethodPost, "/api/v1/admin/usuarios", "viva"))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), messageForbidden) {
			t.Errorf("se esperaba la denegación por permiso: %s", rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), messageCSRF) {
			t.Error("CSRF se evaluó antes que authz")
		}
	})

	t.Run("CSRF va al final", func(t *testing.T) {
		rec := httptest.NewRecorder()
		build(session.Identity{UserID: userID, Permissions: []string{module}}).ServeHTTP(rec,
			sessionRequest(http.MethodPost, "/api/v1/admin/usuarios", "viva"))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, se esperaba 403", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), messageCSRF) {
			t.Errorf("se esperaba el fallo de CSRF: %s", rec.Body.String())
		}
	})

	t.Run("cadena completa llega al handler", func(t *testing.T) {
		req := csrfRequest(http.MethodPost, "/api/v1/admin/usuarios", valid, valid)
		req.AddCookie(&http.Cookie{Name: session.CookieSession, Value: "viva"})
		rec := httptest.NewRecorder()
		build(session.Identity{UserID: userID, Permissions: []string{module}}).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200", rec.Code)
		}
	})
}
