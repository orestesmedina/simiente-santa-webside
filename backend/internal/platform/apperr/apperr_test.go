package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestConstructorsMapKindToStatusAndCode(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		wantStatus int
		wantCode   string
	}{
		{name: "NotFound", err: NotFound("Recurso no encontrado"), wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "MethodNotAllowed", err: MethodNotAllowed("Método no permitido"), wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed"},
		{name: "DatabaseUnavailable", err: DatabaseUnavailable("La base de datos no está conectada"), wantStatus: http.StatusServiceUnavailable, wantCode: "database_unavailable"},
		{name: "Invalid", err: Invalid("Datos inválidos"), wantStatus: http.StatusBadRequest, wantCode: "invalid"},
		{name: "Unauthenticated", err: Unauthenticated("Correo o contraseña incorrectos"), wantStatus: http.StatusUnauthorized, wantCode: "unauthenticated"},
		{name: "Forbidden", err: Forbidden("No tienes permiso para realizar esta operación"), wantStatus: http.StatusForbidden, wantCode: "forbidden"},
		{name: "Conflict", err: Conflict("Ese correo ya está en uso"), wantStatus: http.StatusConflict, wantCode: "conflict"},
		{name: "RateLimited", err: RateLimited("Demasiados intentos", WithRetryAfter(900)), wantStatus: http.StatusTooManyRequests, wantCode: "rate_limited"},
		{name: "Internal", err: Internal(errors.New("boom")), wantStatus: http.StatusInternalServerError, wantCode: "internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.HTTPStatus(); got != tt.wantStatus {
				t.Errorf("HTTPStatus() = %d, se esperaba %d", got, tt.wantStatus)
			}
			if got := tt.err.Code(); got != tt.wantCode {
				t.Errorf("Code() = %q, se esperaba %q", got, tt.wantCode)
			}
			if got := tt.err.Kind.Code(); got != tt.wantCode {
				t.Errorf("Kind.Code() = %q, se esperaba %q", got, tt.wantCode)
			}
		})
	}
}

func TestUnknownKindFallsBackToInternal(t *testing.T) {
	unknown := Kind(99)

	if got := unknown.HTTPStatus(); got != http.StatusInternalServerError {
		t.Errorf("HTTPStatus() = %d, se esperaba %d", got, http.StatusInternalServerError)
	}
	if got := unknown.Code(); got != "internal" {
		t.Errorf("Code() = %q, se esperaba %q", got, "internal")
	}
	if got := unknown.String(); got != "internal" {
		t.Errorf("String() = %q, se esperaba %q", got, "internal")
	}
}

func TestWrapsCauseForErrorsIsAndAs(t *testing.T) {
	sentinel := errors.New("sql: conexión a db-interna falló")
	domainErr := DatabaseUnavailable(
		"La base de datos no está conectada",
		WithDetails(map[string]any{"database": "disconnected"}),
		WithCause(sentinel),
	)

	if !errors.Is(domainErr, sentinel) {
		t.Error("errors.Is no reconoce la causa envuelta con %w")
	}

	wrapped := fmt.Errorf("ping database: %w", domainErr)
	var target *Error
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As no encuentra el *Error envuelto")
	}
	if target.Code() != "database_unavailable" {
		t.Errorf("Code() = %q, se esperaba database_unavailable", target.Code())
	}
}

func TestUnwrapWithoutCauseIsNil(t *testing.T) {
	if got := NotFound("Recurso no encontrado").Unwrap(); got != nil {
		t.Errorf("Unwrap() = %v, se esperaba nil", got)
	}
	var nilErr *Error
	if got := nilErr.Unwrap(); got != nil {
		t.Errorf("Unwrap() sobre *Error nil = %v, se esperaba nil", got)
	}
}

func TestMessageNeverContainsInternalDetail(t *testing.T) {
	const secret = "sql: conexión a db-interna falló"
	tests := []struct {
		name string
		err  *Error
	}{
		{name: "Internal genérico", err: Internal(errors.New(secret))},
		{
			name: "DatabaseUnavailable con causa",
			err: DatabaseUnavailable(
				"La base de datos no está conectada",
				WithDetails(map[string]any{"database": "disconnected"}),
				WithCause(errors.New(secret)),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.err.Message, "sql") || strings.Contains(tt.err.Message, "db-interna") {
				t.Errorf("Message filtra información interna: %q", tt.err.Message)
			}
			for _, value := range tt.err.Details {
				if strings.Contains(fmt.Sprint(value), "sql") {
					t.Errorf("Details filtra información interna: %v", tt.err.Details)
				}
			}
		})
	}
}

func TestInternalUsesGenericMessage(t *testing.T) {
	err := Internal(errors.New("panic: nil pointer"))

	if err.Message != MessageInternal {
		t.Errorf("Message = %q, se esperaba %q", err.Message, MessageInternal)
	}
	if err.Code() != "internal" {
		t.Errorf("Code() = %q, se esperaba internal", err.Code())
	}
	if !strings.Contains(err.Error(), "nil pointer") {
		t.Errorf("Error() = %q, debería incluir la causa para el log", err.Error())
	}
}

func TestDetailsPreserved(t *testing.T) {
	details := map[string]any{"database": "disconnected"}
	err := DatabaseUnavailable("La base de datos no está conectada", WithDetails(details))

	if err.Details["database"] != "disconnected" {
		t.Errorf("Details = %v, se esperaba database=disconnected", err.Details)
	}
}
