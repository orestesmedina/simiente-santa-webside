package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"log/slog"

	"simiente-santa/backend/internal/platform/apperr"
)

// memStore y memHandler capturan los registros de slog para inspeccionarlos en
// las pruebas (sin dependencias).
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

// discardLogger es un logger que no escribe en ninguna parte.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(discardWriter{}, nil))
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// testEnvelope replica el sobre de error para decodificarlo en las pruebas.
type testEnvelope struct {
	Error struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details"`
	} `json:"error"`
}

func decodeEnvelope(t *testing.T, body []byte) testEnvelope {
	t.Helper()
	var env testEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("el cuerpo no es un sobre de error válido: %v (cuerpo: %q)", err, body)
	}
	return env
}

func TestWriteJSONWritesDTOWithoutWrapper(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}{Status: "ok", Database: "connected"}

	WriteJSON(rec, http.StatusOK, payload)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, se esperaba 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, se esperaba application/json", ct)
	}
	const want = `{"status":"ok","database":"connected"}`
	if got := rec.Body.String(); got != want {
		t.Errorf("cuerpo = %q, se esperaba %q (sin wrapper)", got, want)
	}
}

func TestWriteErrorTranslatesEachKind(t *testing.T) {
	const secret = "sql: conexión a db-interna falló"
	tests := []struct {
		name            string
		err             error
		wantStatus      int
		wantCode        string
		wantMessage     string
		wantDetails     map[string]any
		forbiddenInBody []string
	}{
		{
			name:        "NotFound",
			err:         apperr.NotFound("Recurso no encontrado"),
			wantStatus:  http.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "Recurso no encontrado",
		},
		{
			name:        "MethodNotAllowed",
			err:         apperr.MethodNotAllowed("Método no permitido"),
			wantStatus:  http.StatusMethodNotAllowed,
			wantCode:    "method_not_allowed",
			wantMessage: "Método no permitido",
		},
		{
			name:        "DatabaseUnavailable",
			err:         apperr.DatabaseUnavailable("La base de datos no está conectada", apperr.WithDetails(map[string]any{"database": "disconnected"})),
			wantStatus:  http.StatusServiceUnavailable,
			wantCode:    "database_unavailable",
			wantMessage: "La base de datos no está conectada",
			wantDetails: map[string]any{"database": "disconnected"},
		},
		{
			name:            "Internal",
			err:             apperr.Internal(errors.New(secret)),
			wantStatus:      http.StatusInternalServerError,
			wantCode:        "internal",
			wantMessage:     apperr.MessageInternal,
			forbiddenInBody: []string{"sql", "db-interna"},
		},
		{
			name:            "error inesperado no apperr",
			err:             errors.New(secret),
			wantStatus:      http.StatusInternalServerError,
			wantCode:        "internal",
			wantMessage:     apperr.MessageInternal,
			forbiddenInBody: []string{"sql", "db-interna"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteError(context.Background(), rec, discardLogger(), tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, se esperaba %d", rec.Code, tt.wantStatus)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
				t.Errorf("Content-Type = %q, se esperaba application/json", ct)
			}

			env := decodeEnvelope(t, rec.Body.Bytes())
			if env.Error.Code != tt.wantCode {
				t.Errorf("code = %q, se esperaba %q", env.Error.Code, tt.wantCode)
			}
			if env.Error.Message != tt.wantMessage {
				t.Errorf("message = %q, se esperaba %q", env.Error.Message, tt.wantMessage)
			}
			for key, want := range tt.wantDetails {
				if env.Error.Details[key] != want {
					t.Errorf("details[%q] = %v, se esperaba %v", key, env.Error.Details[key], want)
				}
			}
			for _, forbidden := range tt.forbiddenInBody {
				if strings.Contains(rec.Body.String(), forbidden) {
					t.Errorf("el cuerpo filtra el detalle interno %q: %s", forbidden, rec.Body.String())
				}
			}
		})
	}
}

func TestWriteErrorInternalLogsDetailWithRequestID(t *testing.T) {
	store := &memStore{}
	logger := slog.New(newMemHandler(store))
	ctx := ContextWithRequestID(context.Background(), "req-123")
	rec := httptest.NewRecorder()
	cause := errors.New("sql: conexión a db-interna falló")

	WriteError(ctx, rec, logger, apperr.Internal(cause))

	record, ok := store.find(slog.LevelError)
	if !ok {
		t.Fatal("no se registró el error interno en el log")
	}
	if record.attrs["request_id"] != "req-123" {
		t.Errorf("request_id = %v, se esperaba req-123", record.attrs["request_id"])
	}
	detail, _ := record.attrs["error"].(string)
	if !strings.Contains(detail, "db-interna") {
		t.Errorf("el log no contiene el detalle interno: %v", record.attrs)
	}
	if strings.Contains(rec.Body.String(), "db-interna") {
		t.Errorf("la respuesta filtra el detalle interno: %s", rec.Body.String())
	}
}

func TestWriteErrorDatabaseUnavailableLogsCauseNotInBody(t *testing.T) {
	const cause = "dial tcp 127.0.0.1:5432: connection refused"
	store := &memStore{}
	logger := slog.New(newMemHandler(store))
	ctx := ContextWithRequestID(context.Background(), "req-db")
	rec := httptest.NewRecorder()

	WriteError(ctx, rec, logger, apperr.DatabaseUnavailable(
		"La base de datos no está conectada",
		apperr.WithCause(errors.New(cause)),
	))

	record, ok := store.find(slog.LevelWarn)
	if !ok {
		t.Fatal("no se registró el 503 en el log a nivel warn")
	}
	if record.attrs["code"] != "database_unavailable" {
		t.Errorf("code = %v, se esperaba database_unavailable", record.attrs["code"])
	}
	detail, _ := record.attrs["error"].(string)
	if !strings.Contains(detail, cause) {
		t.Errorf("el log no incluye la causa interna: %v", record.attrs)
	}
	if strings.Contains(rec.Body.String(), cause) || strings.Contains(rec.Body.String(), "connection refused") {
		t.Errorf("la respuesta filtra la causa interna: %s", rec.Body.String())
	}
}

func TestWriteErrorUsesRequestLoggerFromContext(t *testing.T) {
	store := &memStore{}
	perRequest := slog.New(newMemHandler(store)).With(slog.String("request_id", "req-ctx"))
	ctx := ContextWithRequestLogger(context.Background(), perRequest)
	rec := httptest.NewRecorder()

	WriteError(ctx, rec, discardLogger(), apperr.NotFound("Recurso no encontrado"))

	record, ok := store.find(slog.LevelWarn)
	if !ok {
		t.Fatal("no se registró el error de dominio en el logger de la petición")
	}
	if record.attrs["request_id"] != "req-ctx" {
		t.Errorf("request_id = %v, se esperaba req-ctx", record.attrs["request_id"])
	}
	if record.attrs["code"] != "not_found" {
		t.Errorf("code = %v, se esperaba not_found", record.attrs["code"])
	}
}
