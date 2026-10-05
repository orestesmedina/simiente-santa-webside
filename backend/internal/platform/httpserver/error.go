package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"simiente-santa/backend/internal/platform/apperr"
)

// headerRetryAfter es la cabecera del contrato que acompaña a un 429 y lleva
// los segundos hasta poder reintentar.
const headerRetryAfter = "Retry-After"

// Claves del contexto de petición. Las escribe middleware/request-id y las
// consumen WriteError y middleware/logging; viven aquí porque WriteError es el
// único punto de traducción de errores a HTTP y no puede depender de middleware
// (sería un ciclo).
type contextKey int

const (
	requestIDKey contextKey = iota
	requestLoggerKey
)

// ContextWithRequestID guarda el identificador de la petición en el contexto.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext devuelve el identificador de la petición, o "" si no lo
// hay.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// ContextWithRequestLogger guarda el logger por petición (hij con request_id,
// método y ruta) en el contexto.
func ContextWithRequestLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, requestLoggerKey, l)
}

// RequestLoggerFromContext devuelve el logger por petición, o nil si no lo hay.
func RequestLoggerFromContext(ctx context.Context) *slog.Logger {
	l, _ := ctx.Value(requestLoggerKey).(*slog.Logger)
	return l
}

// errorEnvelope es el sobre de error estándar: {"error":{code,message,details?}}.
type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// WriteJSON escribe el sobre de éxito: el DTO directo de la operación (sin
// wrapper {"data":…}), con Content-Type JSON. No hay un segundo formato de
// éxito (D13).
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		// No debería ocurrir con los DTOs de la API; si ocurre no se filtra el
		// error interno, se responde el genérico.
		writeErrorBody(w, http.StatusInternalServerError, "internal", apperr.MessageInternal, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// WriteError es el ÚNICO punto donde un error de dominio (apperr) se traduce a
// HTTP (arq. §5.11). Cualquier error inesperado se responde como 500 `internal`
// con mensaje genérico; el detalle interno solo va al log estructurado con su
// request_id (FR-013, SC-009). Si el error lleva Retry-After (RateLimited), la
// cabecera homónima del contrato se fija aquí con su valor en segundos.
func WriteError(ctx context.Context, w http.ResponseWriter, logger *slog.Logger, err error) {
	domainErr := toDomainError(err)

	message := domainErr.Message
	var details map[string]any
	if domainErr.Kind != apperr.KindInternal {
		details = domainErr.Details
	}

	if retryAfter := domainErr.RetryAfter(); retryAfter > 0 {
		w.Header().Set(headerRetryAfter, strconv.Itoa(retryAfter))
	}

	logError(ctx, logger, domainErr)
	writeErrorBody(w, domainErr.HTTPStatus(), domainErr.Code(), message, details)
}

// writeErrorBody serializa el sobre de error ya traducido.
func writeErrorBody(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	body, err := json.Marshal(errorEnvelope{Error: errorBody{
		Code:    code,
		Message: message,
		Details: details,
	}})
	if err != nil {
		http.Error(w, apperr.MessageInternal, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// toDomainError normaliza cualquier error a un *apperr.Error: los que ya lo son
// se devuelven tal cual; el resto (o nil) caen en el fallback 500.
func toDomainError(err error) *apperr.Error {
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		return domainErr
	}
	if err == nil {
		return apperr.Internal(errors.New("error desconocido"))
	}
	return apperr.Internal(err)
}

// logError registra el error. Los 500 llevan el detalle interno a nivel Error;
// el resto a nivel Warn. El request_id siempre acompaña si está en el contexto
// (el logger por petición ya lo trae consigo y no se duplica).
func logError(ctx context.Context, logger *slog.Logger, domainErr *apperr.Error) {
	reqLogger := RequestLoggerFromContext(ctx)
	if reqLogger != nil {
		logger = reqLogger
	} else if logger != nil {
		if id := RequestIDFromContext(ctx); id != "" {
			logger = logger.With(slog.String("request_id", id))
		}
	}
	if logger == nil {
		return
	}

	attrs := []any{slog.String("code", domainErr.Code())}
	if domainErr.Kind == apperr.KindInternal {
		attrs = append(attrs, slog.String("error", domainErr.Error()))
		logger.Error("error interno", attrs...)
		return
	}
	// El 503 database_unavailable es previsible (no es un 500), pero su causa
	// —DNS, timeout, credenciales— es clave para diagnosticar; va SOLO al log,
	// nunca al cuerpo de la respuesta (FR-013).
	if domainErr.Kind == apperr.KindDatabaseUnavailable {
		attrs = append(attrs, slog.String("error", domainErr.Error()))
	}
	logger.Warn("error de dominio", attrs...)
}
