// Package logger construye los logs estructurados de la aplicación (log/slog
// en JSON) y deriva el logger de cada petición con su request_id, método y
// ruta. Es el único punto donde se configura el formato y la severidad de los
// logs (D12, US4 esc. 3).
package logger

import (
	"io"
	"log/slog"
	"os"
)

// Claves de los campos estructurados comunes. Los nombres van en inglés como el
// resto del código (§V), salvo request_id, que se conserva tal cual porque es
// el identificador que viaja en el contrato y en los logs.
const (
	RequestIDKey = "request_id"
	MethodKey    = "method"
	PathKey      = "path"
)

// New devuelve un *slog.Logger con handler JSON que escribe en stdout al nivel
// indicado. platform/config valida y traduce LOG_LEVEL a slog.Level.
func New(level slog.Level) *slog.Logger {
	return NewWithWriter(os.Stdout, level)
}

// NewWithWriter es como New pero escribe en w. Existe para capturar la salida
// en las pruebas y para destinos alternativos.
func NewWithWriter(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}

// Request deriva el logger de una petición a partir del logger padre,
// añadiendo request_id, método y ruta sin perder los campos ni el handler del
// padre. Lo usa middleware/request-id.
func Request(parent *slog.Logger, requestID, method, path string) *slog.Logger {
	return parent.With(
		slog.String(RequestIDKey, requestID),
		slog.String(MethodKey, method),
		slog.String(PathKey, path),
	)
}
