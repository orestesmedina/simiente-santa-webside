package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
)

// memoryHandler es un slog.Handler en memoria que conserva los registros y sus
// atributos (incluidos los heredados con With/WithAttrs) para inspeccionarlos
// en las pruebas.
type memoryHandler struct {
	state *memoryState
	attrs []slog.Attr
}

type memoryState struct {
	mu      sync.Mutex
	records []capturedRecord
}

type capturedRecord struct {
	level slog.Level
	msg   string
	attrs map[string]any
}

func newMemoryHandler() (*memoryHandler, *memoryState) {
	state := &memoryState{}
	return &memoryHandler{state: state}, state
}

func (h *memoryHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *memoryHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	h.state.mu.Lock()
	defer h.state.mu.Unlock()
	h.state.records = append(h.state.records, capturedRecord{level: r.Level, msg: r.Message, attrs: attrs})
	return nil
}

func (h *memoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	merged = append(merged, h.attrs...)
	merged = append(merged, attrs...)
	return &memoryHandler{state: h.state, attrs: merged}
}

func (h *memoryHandler) WithGroup(string) slog.Handler { return h }

func TestNewEmitsValidJSON(t *testing.T) {
	var buf bytes.Buffer
	log := NewWithWriter(&buf, slog.LevelInfo)

	log.Info("servidor iniciado", "service", "api")

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("la salida no es JSON válido: %v (salida: %q)", err, buf.String())
	}
	if got["msg"] != "servidor iniciado" {
		t.Errorf("msg = %v, se esperaba %q", got["msg"], "servidor iniciado")
	}
	if got["level"] != "INFO" {
		t.Errorf("level = %v, se esperaba INFO", got["level"])
	}
	if got["service"] != "api" {
		t.Errorf("service = %v, se esperaba api", got["service"])
	}
	if _, ok := got["time"]; !ok {
		t.Error("la salida JSON no incluye el campo time")
	}
}

func TestLevelApplied(t *testing.T) {
	tests := []struct {
		name       string
		level      slog.Level
		emit       func(*slog.Logger)
		wantLogged bool
	}{
		{name: "debug visible con nivel debug", level: slog.LevelDebug, emit: func(l *slog.Logger) { l.Debug("x") }, wantLogged: true},
		{name: "debug oculto con nivel info", level: slog.LevelInfo, emit: func(l *slog.Logger) { l.Debug("x") }, wantLogged: false},
		{name: "info oculto con nivel warn", level: slog.LevelWarn, emit: func(l *slog.Logger) { l.Info("x") }, wantLogged: false},
		{name: "warn visible con nivel warn", level: slog.LevelWarn, emit: func(l *slog.Logger) { l.Warn("x") }, wantLogged: true},
		{name: "error visible con nivel error", level: slog.LevelError, emit: func(l *slog.Logger) { l.Error("x") }, wantLogged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := NewWithWriter(&buf, tt.level)

			tt.emit(log)

			got := buf.Len() > 0
			if got != tt.wantLogged {
				t.Errorf("registro emitido = %v, se esperaba %v (salida: %q)", got, tt.wantLogged, buf.String())
			}
		})
	}
}

func TestRequestAddsFieldsWithoutLosingParent(t *testing.T) {
	handler, state := newMemoryHandler()
	parent := slog.New(handler).With("service", "api", "env", "test")

	reqLog := Request(parent, "req-123", "GET", "/healthz")
	reqLog.Info("petición atendida")

	if len(state.records) != 1 {
		t.Fatalf("registros capturados = %d, se esperaba 1", len(state.records))
	}
	attrs := state.records[0].attrs
	want := map[string]any{
		RequestIDKey: "req-123",
		MethodKey:    "GET",
		PathKey:      "/healthz",
		"service":    "api",
		"env":        "test",
	}
	for key, value := range want {
		if attrs[key] != value {
			t.Errorf("campo %q = %v, se esperaba %v", key, attrs[key], value)
		}
	}
}

func TestRequestDoesNotMutateParent(t *testing.T) {
	handler, state := newMemoryHandler()
	parent := slog.New(handler).With("service", "api")

	_ = Request(parent, "req-123", "GET", "/healthz")
	parent.Info("sin petición")

	if len(state.records) != 1 {
		t.Fatalf("registros capturados = %d, se esperaba 1", len(state.records))
	}
	attrs := state.records[0].attrs
	if _, ok := attrs[RequestIDKey]; ok {
		t.Error("el logger padre quedó contaminado con request_id")
	}
	if attrs["service"] != "api" {
		t.Errorf("el logger padre perdió sus campos: %v", attrs)
	}
}
