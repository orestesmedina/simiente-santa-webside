// Package apperr define los errores de dominio tipados y su traducción a HTTP.
//
// Un Error nunca expone su causa interna al cliente: solo Message (seguro, en
// español) y Details (opcional, también seguro) se serializan. La causa viaja
// envuelta con %w para el log estructurado y para errors.Is/As (FR-013).
//
// F1 implementa solo los kinds que emite: NotFound (404 del fallback del
// router), MethodNotAllowed (405 del fallback), DatabaseUnavailable (503 de
// /healthz) e Internal (500). El resto del registro (Invalid, Unauthenticated,
// Forbidden, Conflict, RateLimited) crece bajo demanda en F2+ con su productor.
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
)

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

	cause error
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

// HTTPStatus devuelve el código HTTP asociado al kind.
func (e *Error) HTTPStatus() int { return e.Kind.HTTPStatus() }

// Code devuelve el error.code del contrato asociado al kind.
func (e *Error) Code() string { return e.Kind.Code() }

// HTTPStatus traduce un Kind a su código HTTP (tabla de §5.11 de
// arquitectura.md). Los kinds desconocidos caen en 500, el fallback seguro.
func (k Kind) HTTPStatus() int {
	switch k {
	case KindNotFound:
		return http.StatusNotFound
	case KindMethodNotAllowed:
		return http.StatusMethodNotAllowed
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
	case KindNotFound:
		return "not_found"
	case KindMethodNotAllowed:
		return "method_not_allowed"
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

// Internal es el fallback 500: mensaje genérico y causa interna solo para el
// log.
func Internal(cause error) *Error {
	return New(KindInternal, MessageInternal, WithCause(cause))
}
