// Package apperr define los errores de dominio tipados y su traducción a HTTP.
//
// Un Error nunca expone su causa interna al cliente: solo Message (seguro, en
// español) y Details (opcional, también seguro) se serializan. La causa viaja
// envuelta con %w para el log estructurado y para errors.Is/As (FR-013).
//
// F1 implementa NotFound (404 del fallback del router), MethodNotAllowed (405
// del fallback), DatabaseUnavailable (503 de /healthz) e Internal (500). F2
// añade los kinds que emite su superficie: Invalid (400), Unauthenticated
// (401), Forbidden (403), Conflict (409) y RateLimited (429, con Retry-After).
// El registro sigue cerrado a la tabla de §5.11 de arquitectura.md: ningún
// error_code fuera de él.
package apperr

import (
	"fmt"
	"net/http"
)

// Kind clasifica el error de dominio y determina su traducción a HTTP.
// El valor cero es KindInternal: un Error sin kind cae en el fallback seguro.
type Kind int

const (
	KindInternal Kind = iota
	KindNotFound
	KindMethodNotAllowed
	KindDatabaseUnavailable

	// Kinds de F2 (P16, tabla de §5.11 de arquitectura.md). Se añaden al final
	// para no desplazar los valores ya usados por F1.
	KindInvalid
	KindUnauthenticated
	KindForbidden
	KindConflict
	KindRateLimited
)

// RetryAfterDetail es la clave de Details con los segundos hasta poder
// reintentar; acompaña siempre a RateLimited.
const RetryAfterDetail = "retryAfterSeconds"

// MessageInternal es el mensaje genérico que responde un error inesperado
// (FR-013): nunca incluye la causa interna.
const MessageInternal = "Error interno del servidor"

// Error es un error de dominio. Solo Message y Details son seguros para el
// cliente; la causa interna se conserva para el log y errors.Is/As, pero nunca
// se serializa.
type Error struct {
	Kind    Kind
	Message string
	Details map[string]any

	// retryAfter guarda los segundos hasta poder reintentar (RateLimited). No
	// se serializa por sí mismo: la cabecera Retry-After la fija WriteError a
	// partir de RetryAfter(), y Details["retryAfterSeconds"] viaja en el sobre.
	retryAfter int
	cause      error
}

// Error implementa error. Incluye la causa para que el log estructurado tenga
// el detalle interno; el texto seguro para el cliente es Message.
func (e *Error) Error() string {
	if e.cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.cause)
}

// Unwrap devuelve la causa interna envuelta con %w (nunca serializada).
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// RetryAfter devuelve los segundos hasta poder reintentar (0 si el error no es
// RateLimited o no se indicó el valor). WriteError lo usa para fijar la
// cabecera Retry-After del contrato.
func (e *Error) RetryAfter() int {
	if e == nil {
		return 0
	}
	return e.retryAfter
}

// HTTPStatus devuelve el código HTTP asociado al kind.
func (e *Error) HTTPStatus() int { return e.Kind.HTTPStatus() }

// Code devuelve el error.code del contrato asociado al kind.
func (e *Error) Code() string { return e.Kind.Code() }

// HTTPStatus traduce un Kind a su código HTTP (tabla de §5.11 de
// arquitectura.md). Los kinds desconocidos caen en 500, el fallback seguro.
func (k Kind) HTTPStatus() int {
	switch k {
	case KindInvalid:
		return http.StatusBadRequest
	case KindUnauthenticated:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case KindConflict:
		return http.StatusConflict
	case KindRateLimited:
		return http.StatusTooManyRequests
	case KindDatabaseUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Code traduce un Kind al error.code del contrato (snake_case). Los kinds
// desconocidos caen en "internal".
func (k Kind) Code() string {
	switch k {
	case KindInvalid:
		return "invalid"
	case KindUnauthenticated:
		return "unauthenticated"
	case KindForbidden:
		return "forbidden"
	case KindNotFound:
		return "not_found"
	case KindMethodNotAllowed:
		return "method_not_allowed"
	case KindConflict:
		return "conflict"
	case KindRateLimited:
		return "rate_limited"
	case KindDatabaseUnavailable:
		return "database_unavailable"
	default:
		return "internal"
	}
}

// String implementa fmt.Stringer con el code del kind (útil en logs).
func (k Kind) String() string { return k.Code() }

// Option ajusta un Error durante su construcción.
type Option func(*Error)

// WithDetails adjunta el detalle seguro que viaja en la respuesta.
func WithDetails(details map[string]any) Option {
	return func(e *Error) { e.Details = details }
}

// WithCause envuelve la causa interna (nunca se serializa).
func WithCause(cause error) Option {
	return func(e *Error) { e.cause = cause }
}

// WithRetryAfter adjunta los segundos hasta poder reintentar: viajan en
// Details["retryAfterSeconds"] y en RetryAfter() para que WriteError fije la
// cabecera Retry-After. Solo tiene sentido en RateLimited; los valores
// negativos se normalizan a 0.
func WithRetryAfter(seconds int) Option {
	return func(e *Error) {
		if seconds < 0 {
			seconds = 0
		}
		e.retryAfter = seconds
		if e.Details == nil {
			e.Details = make(map[string]any, 1)
		}
		e.Details[RetryAfterDetail] = seconds
	}
}

// New construye un Error del kind indicado con un mensaje seguro para el
// cliente.
func New(kind Kind, message string, opts ...Option) *Error {
	e := &Error{Kind: kind, Message: message}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// NotFound es el fallback 404 del router (ruta no documentada).
func NotFound(message string, opts ...Option) *Error {
	return New(KindNotFound, message, opts...)
}

// MethodNotAllowed es el fallback 405 del router (método no documentado sobre
// una ruta existente).
func MethodNotAllowed(message string, opts ...Option) *Error {
	return New(KindMethodNotAllowed, message, opts...)
}

// DatabaseUnavailable indica que el servicio está vivo pero la base de datos no
// responde (503 de /healthz). Suele llevar Details{"database":"disconnected"}.
func DatabaseUnavailable(message string, opts ...Option) *Error {
	return New(KindDatabaseUnavailable, message, opts...)
}

// Invalid indica datos de entrada inválidos o incompletos (400). El mensaje es
// seguro para el cliente; Details suele llevar un mensaje por campo.
func Invalid(message string, opts ...Option) *Error {
	return New(KindInvalid, message, opts...)
}

// Unauthenticated indica que falta una sesión válida o que las credenciales no
// son correctas (401). El mensaje de login debe ser genérico y no revelar si la
// cuenta existe (FR-003).
func Unauthenticated(message string, opts ...Option) *Error {
	return New(KindUnauthenticated, message, opts...)
}

// Forbidden indica que la cuenta está autenticada pero no puede realizar la
// operación (403): sin permiso, cuenta desactivada o cambio de contraseña
// obligatorio (Details{"reason":"password_change_required"}).
func Forbidden(message string, opts ...Option) *Error {
	return New(KindForbidden, message, opts...)
}

// Conflict indica un conflicto con el estado actual (409): correo o nombre de
// rol duplicado, rol con cuentas asignadas o una operación que dejaría el panel
// sin administración (Details{"reason":"admin_required"}).
func Conflict(message string, opts ...Option) *Error {
	return New(KindConflict, message, opts...)
}

// RateLimited indica demasiadas peticiones (429). Usa WithRetryAfter para
// adjuntar los segundos de espera: el valor viaja en
// Details["retryAfterSeconds"] y en RetryAfter(), de donde WriteError toma la
// cabecera Retry-After del contrato.
func RateLimited(message string, opts ...Option) *Error {
	return New(KindRateLimited, message, opts...)
}

// Internal es el fallback 500: mensaje genérico y causa interna solo para el
// log.
func Internal(cause error) *Error {
	return New(KindInternal, MessageInternal, WithCause(cause))
}
