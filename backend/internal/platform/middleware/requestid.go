// Package middleware contiene la cadena transversal de F1 (D11/D12). El orden
// de montaje lo fija httpserver.New: el primero de la lista es el más externo.
//
//	request-id → recover → logging → CORS → handler
//
// F2+ añadirá rate-limit y, en grupos, authn/authz/CSRF (no en F1).
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"simiente-santa/backend/internal/platform/httpserver"
)

// HeaderRequestID es la cabecera que transporta el identificador de la
// petición, tanto de entrada (la puede traer el cliente) como de salida.
const HeaderRequestID = "X-Request-ID"

// RequestID toma el X-Request-ID del cliente o genera uno, lo publica en la
// respuesta y lo guarda en el contexto para que lo usen logging y WriteError.
// Va el primero de la cadena. Es una función —no una variable de paquete— para
// respetar R6 (sin estado global); se pasa como httpserver.Middleware por
// asignabilidad.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get(HeaderRequestID))
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(HeaderRequestID, id)
		next.ServeHTTP(w, r.WithContext(httpserver.ContextWithRequestID(r.Context(), id)))
	})
}

// newRequestID genera un identificador aleatorio de 128 bits en hexadecimal.
// Si la aleatoriedad falla (no debería) recurre al reloj para no romper la
// trazabilidad.
func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}
