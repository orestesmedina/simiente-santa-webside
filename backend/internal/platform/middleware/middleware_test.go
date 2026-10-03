package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
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
		req.Header.Set(headerRequestMethod, "GET")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, se esperaba 204", rec.Code)
		}
		if got := rec.Header().Get(headerAllowOrigin); got != allowedOrigin {
			t.Errorf("Allow-Origin = %q, se esperaba %q", got, allowedOrigin)
		}
		if rec.Header().Get(headerAllowMethods) == "" {
			t.Error("falta Access-Control-Allow-Methods en el preflight permitido")
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
	})
}
