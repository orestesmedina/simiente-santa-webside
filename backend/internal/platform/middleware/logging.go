package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"simiente-santa/backend/internal/platform/httpserver"
	applogger "simiente-santa/backend/internal/platform/logger"
)

// Claves de los campos del registro de acceso (además de request_id, method y
// path, que los aporta el logger por petición de platform/logger).
const (
	statusKey   = "status"
	durationKey = "duration_ms"
)

// Logging registra al terminar la petición: método, ruta, status, duración y
// request_id (los cinco campos de D12).
func Logging(logger *slog.Logger) httpserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			base := logger
			if base == nil {
				base = slog.Default()
			}
			requestLog := applogger.Request(base, httpserver.RequestIDFromContext(r.Context()), r.Method, r.URL.Path)
			requestLog.Info("petición atendida",
				slog.Int(statusKey, rec.status),
				slog.Int64(durationKey, time.Since(start).Milliseconds()),
			)
		})
	}
}

// statusRecorder recuerda el status escrito para poder registrarlo al final.
type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wrote {
		r.status = status
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.wrote = true // un Write sin WriteHeader implica 200
	}
	return r.ResponseWriter.Write(b)
}
