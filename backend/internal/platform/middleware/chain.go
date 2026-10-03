package middleware

import (
	"log/slog"

	"simiente-santa/backend/internal/platform/httpserver"
)

// Chain arma la cadena transversal de F1 en el orden aprobado (el primero, el
// más externo): request-id → recover → logging → CORS. Es el único constructor
// de la composición: lo usan cmd/api en producción y testutil en las pruebas,
// para que ambas no puedan divergir (FR-011). platform no conoce dominios (R1).
func Chain(logger *slog.Logger, allowedOrigins []string) []httpserver.Middleware {
	return []httpserver.Middleware{
		RequestID,
		Recover(logger),
		Logging(logger),
		CORS(allowedOrigins),
	}
}
