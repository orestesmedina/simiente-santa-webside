package usuarios

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
)

// handler_audit.go implementa la superficie de SOLO LECTURA del registro de
// auditoría (FR-024/FR-025, US8, P22):
//
//   - GET /api/v1/admin/auditoria/accesos — historial de intentos de acceso.
//   - GET /api/v1/admin/auditoria/acciones — historial de acciones sensibles.
//
// Ambas van bajo el permiso `admin_usuarios_roles` (la misma decisión que la
// gestión): sin él, authz responde 403. Este archivo NO publica ninguna
// operación de escritura: el registro no se edita ni se borra por ninguna vía
// (FR-025). Los campos de cuenta/actor se derivan por JOIN en la consulta y
// llegan ya resueltos en los DTOs (F-01); los intentos sin cuenta asociada van
// con esos campos en `null` y sin guardar ni mostrar correo alguno (FR-026).

// AuditService es el puerto que los handlers de auditoría (consulta) y el
// helper de decodificación del handler (registro de rechazos, T239) necesitan
// del servicio. Lo define quien lo consume (arq. R3) y lo implementa
// *auditService.
type AuditService interface {
	// ListAccessEvents devuelve una página del historial de accesos (FR-024).
	ListAccessEvents(ctx context.Context, filter AuditFilter) (AccessEventList, error)
	// ListAdminActions devuelve una página del historial de acciones (FR-024).
	ListAdminActions(ctx context.Context, filter AuditFilter) (AdminActionList, error)
	// RecordRejectedBestEffort registra un intento rechazado por JSON o DTO
	// inválido; es best-effort y nunca cambia la respuesta (R23).
	RecordRejectedBestEffort(ctx context.Context, rejection Rejection)
}

// Mensajes seguros de la entrada inválida de los listados de auditoría.
const (
	messageInvalidAuditUser = "El identificador de la cuenta no es válido"
	messageInvalidAuditDate = "El rango de fechas no es válido"
)

// ListAccessEvents responde GET /api/v1/admin/auditoria/accesos (FR-024/US8):
// normaliza `userId`/`from`/`to` y la paginación con platform/paginate (P14) y
// devuelve el sobre AccessEventList. Un `from` posterior a `to` (o una fecha
// mal formada) → 400 invalid. Códigos: 200; parámetros inválidos → 400; sin
// permiso → 403.
func (h *Handler) ListAccessEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.auditConfigured(ctx, w) {
		return
	}
	filter, err := auditFilterFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	list, err := h.audit.ListAccessEvents(ctx, filter)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// ListAdminActions responde GET /api/v1/admin/auditoria/acciones (FR-024/US8):
// con los mismos parámetros que el historial de accesos; `userId` filtra por la
// cuenta involucrada (actor o cuenta objetivo). Códigos: 200; parámetros
// inválidos → 400; sin permiso → 403.
func (h *Handler) ListAdminActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.auditConfigured(ctx, w) {
		return
	}
	filter, err := auditFilterFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	list, err := h.audit.ListAdminActions(ctx, filter)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// RegisterAuditRoutes publica las rutas de solo lectura del registro dentro de
// un Registrar ya protegido por la cadena del panel (authn → passwordguard →
// authz(admin_usuarios_roles) → CSRF, T227). Solo GET: no existe ninguna
// operación de escritura sobre /auditoria (FR-025). No publica nada si el
// servicio no está cableado.
func RegisterAuditRoutes(root httpserver.Registrar, h *Handler) {
	if h == nil || h.audit == nil {
		return
	}
	root.Handle(http.MethodGet, "/auditoria/accesos", h.ListAccessEvents)
	root.Handle(http.MethodGet, "/auditoria/acciones", h.ListAdminActions)
}

// auditConfigured responde 500 si el servicio de auditoría no está cableado: es
// un error de composición interno, no una entrada inválida.
func (h *Handler) auditConfigured(ctx context.Context, w http.ResponseWriter) bool {
	if h.audit != nil {
		return true
	}
	httpserver.WriteError(ctx, w, h.logger,
		apperr.Internal(errors.New("auditoría: servicio no configurado")))
	return false
}

// auditFilterFromRequest traduce los query params del listado a AuditFilter
// (P22): `userId` (UUID), `from`/`to` (ISO-8601/RFC3339) y `limit`/`offset`
// normalizados con platform/paginate. Devuelve apperr.Invalid (400) con el
// campo en Details cuando la entrada no sirve. El orden `from <= to` lo valida
// el servicio, que es el dueño de la regla.
func auditFilterFromRequest(r *http.Request) (AuditFilter, error) {
	values := r.URL.Query()
	params, err := paginate.FromValues(values)
	if err != nil {
		return AuditFilter{}, err
	}
	filter := AuditFilter{Limit: params.Limit, Offset: params.Offset}

	if raw := strings.TrimSpace(values.Get("userId")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return AuditFilter{}, apperr.Invalid(
				messageInvalidAuditUser,
				apperr.WithDetails(map[string]any{"userId": "Debe ser un identificador UUID válido"}),
			)
		}
		filter.UserID = &id
	}

	from, err := parseAuditTime(values.Get("from"), "from")
	if err != nil {
		return AuditFilter{}, err
	}
	filter.From = from

	to, err := parseAuditTime(values.Get("to"), "to")
	if err != nil {
		return AuditFilter{}, err
	}
	filter.To = to

	return filter, nil
}

// parseAuditTime lee una fecha de query param en ISO-8601 (RFC3339). Vacío es
// nil (sin límite); un valor mal formado → 400 invalid con el campo en Details.
func parseAuditTime(raw, field string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, apperr.Invalid(
			messageInvalidAuditDate,
			apperr.WithDetails(map[string]any{field: "Debe ser una fecha ISO-8601 en UTC"}),
		)
	}
	return &value, nil
}

// auditService implementa el puerto que consumen los handlers de auditoría y el
// helper de decodificación del handler.
var _ AuditService = (*auditService)(nil)
