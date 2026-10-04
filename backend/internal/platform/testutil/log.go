package testutil

import (
	"context"
	"log/slog"
	"sync"
)

// Record es un registro de slog capturado en memoria, con sus atributos ya
// aplanados (incluidos los heredados con With/WithAttrs).
type Record struct {
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

// LogCapture acumula los registros que emite el logger que devuelve NewLogger.
// Es seguro para uso concurrente.
type LogCapture struct {
	mu      sync.Mutex
	records []Record
}

// NewLogger devuelve un *slog.Logger que escribe en una captura en memoria,
// pensado para las aserciones sobre logs (SC-009).
func NewLogger() (*slog.Logger, *LogCapture) {
	capture := &LogCapture{}
	return slog.New(&captureHandler{capture: capture}), capture
}

// Records devuelve una copia de los registros capturados.
func (c *LogCapture) Records() []Record {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Record, len(c.records))
	copy(out, c.records)
	return out
}

// Count devuelve cuántos registros se capturaron.
func (c *LogCapture) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.records)
}

// Find devuelve el primer registro con el mensaje indicado.
func (c *LogCapture) Find(message string) (Record, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, record := range c.records {
		if record.Message == message {
			return record, true
		}
	}
	return Record{}, false
}

// captureHandler implementa slog.Handler guardando cada registro en LogCapture.
type captureHandler struct {
	capture *LogCapture
	attrs   []slog.Attr
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	h.capture.mu.Lock()
	defer h.capture.mu.Unlock()
	h.capture.records = append(h.capture.records, Record{Level: r.Level, Message: r.Message, Attrs: attrs})
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &captureHandler{capture: h.capture, attrs: merged}
}

func (h *captureHandler) WithGroup(string) slog.Handler { return h }
