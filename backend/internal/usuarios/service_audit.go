package usuarios

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
)

// auditService escribe el registro de auditoría que consumen el resto de
// servicios y el middleware (FR-022/FR-023, P20). No decide reglas de negocio:
// recibe un intento (audit.Event) o una acción (audit.Action) ya construidos y
// los persiste; de un intento no identificado nunca guarda el correo probado
// (FR-026).
//
// Hay dos formas de registrar:
//
//   - En la MISMA transacción que la mutación: los servicios de gestión usan
//     los métodos transaccionales del repositorio (withTx/WithAdminGuard) e
//     insertan la acción dentro de ella, de modo que el éxito y su registro son
//     atómicos (o ambos o ninguno, R23). Este servicio no participa ahí.
//   - Best-effort: un intento que YA va a fallar (login con mala contraseña,
//     acción denegada) se registra con las variantes *BestEffort, que registran
//     el error con su request_id y NUNCA cambian la respuesta que ve la persona.
type auditService struct {
	repo AuditRepository
	log  *slog.Logger
}

// AuditRepository es el puerto que el servicio de auditoría necesita del
// repositorio (lo define quien lo consume, arq. R3): escritura de cada intento y
// cada acción, y consulta de solo lectura de ambos historiales (FR-025).
type AuditRepository interface {
	InsertLoginEvent(ctx context.Context, event audit.Event) (LoginEvent, error)
	InsertAdminAction(ctx context.Context, action audit.Action) (AdminAction, error)
	// ListLoginEvents y CountLoginEvents consultan el historial de accesos con
	// los filtros de cuenta y semirango `[from, to)` (FR-024).
	ListLoginEvents(ctx context.Context, filter AuditFilter) ([]AccessEvent, error)
	CountLoginEvents(ctx context.Context, filter AuditFilter) (int64, error)
	// ListAdminActions y CountAdminActions consultan el historial de acciones;
	// el filtro por cuenta incluye actor y objetivo (FR-024).
	ListAdminActions(ctx context.Context, filter AuditFilter) ([]AdminActionEntry, error)
	CountAdminActions(ctx context.Context, filter AuditFilter) (int64, error)
}

// NewAuditService construye el servicio de registro de auditoría. Si logger es
// nil se usa el logger por defecto.
func NewAuditService(repo AuditRepository, logger *slog.Logger) *auditService {
	if logger == nil {
		logger = slog.Default()
	}
	return &auditService{repo: repo, log: logger}
}

// RecordLoginEvent registra un intento de acceso (éxito, fallo, cuenta inactiva
// o intento durante un bloqueo: FR-022). Devuelve el error de escritura para
// que el login correcto pueda ser fail-closed (sin registro, sin acceso).
func (s *auditService) RecordLoginEvent(ctx context.Context, event audit.Event) error {
	if err := event.Validate(); err != nil {
		return err
	}
	if _, err := s.repo.InsertLoginEvent(ctx, event); err != nil {
		return err
	}
	return nil
}

// RecordAction registra una acción administrativa sensible (FR-023), con
// cualquier resultado (success/failure/denied). Nunca recibe credenciales: un
// restablecimiento solo lleva actor, objetivo y fecha (FR-026).
func (s *auditService) RecordAction(ctx context.Context, action audit.Action) error {
	if err := action.Validate(); err != nil {
		return err
	}
	if _, err := s.repo.InsertAdminAction(ctx, action); err != nil {
		return err
	}
	return nil
}

// RecordLoginEventBestEffort registra un intento que ya va a fallar: si la
// escritura falla, se queda en el log con su request_id y la respuesta no
// cambia (R23).
func (s *auditService) RecordLoginEventBestEffort(ctx context.Context, event audit.Event) {
	if err := s.RecordLoginEvent(ctx, event); err != nil {
		s.logError(ctx, "no se pudo registrar el intento de acceso", err)
	}
}

// RecordActionBestEffort registra una acción que ya va a fallar o se deniega:
// un fallo se queda en el log con su request_id y no cambia la respuesta (R23).
func (s *auditService) RecordActionBestEffort(ctx context.Context, action audit.Action) {
	if err := s.RecordAction(ctx, action); err != nil {
		s.logError(ctx, "no se pudo registrar la acción administrativa", err)
	}
}

// RecordDenied implementa audit.Recorder (T217): registra la denegación de un
// permiso (`result='denied'`) resolviendo la acción y el objetivo desde el
// método y la ruta (R23). Una petición de lectura denegada no corresponde a
// ninguna acción sensible y no deja fila.
func (s *auditService) RecordDenied(ctx context.Context, denial audit.Denial) error {
	action, ok := actionFromRoute(denial.Method, denial.Path)
	if !ok {
		return nil
	}
	action.ActorUserID = denial.ActorUserID
	action.Result = audit.ResultDenied
	return s.RecordAction(ctx, action)
}

// RecordRejectedBestEffort registra un intento rechazado ANTES de llegar al
// servicio —JSON inválido o DTO no válido en el helper de decodificación del
// handler (P20/R23)—. Resuelve la acción y el objetivo desde method+path con la
// misma tabla del dominio que las denegaciones y lo escribe con
// `result='failure'`. Nunca transporta el cuerpo de la petición, así que no
// puede guardar credenciales (FR-026). Es best-effort: su fallo se queda en el
// log con el request_id y no cambia la respuesta.
func (s *auditService) RecordRejectedBestEffort(ctx context.Context, rejection Rejection) {
	action, ok := actionFromRoute(rejection.Method, rejection.Path)
	if !ok {
		return
	}
	action.ActorUserID = rejection.ActorUserID
	action.Result = audit.ResultFailure
	s.RecordActionBestEffort(ctx, action)
}

// Rejection describe una petición rechazada por JSON o DTO inválido (P20). El
// dominio resuelve la acción y el objetivo desde method+path; por diseño no
// lleva el cuerpo de la petición (FR-026).
type Rejection struct {
	ActorUserID *uuid.UUID
	Method      string
	Path        string
}

// ListAccessEvents devuelve una página del historial de intentos de acceso
// (FR-024/P22), con filtros de cuenta y semirango de fechas `[from, to)`, orden
// `createdAt DESC` y paginación de platform/paginate. Es SOLO lectura: no
// escribe, modifica ni borra el registro (FR-025). Un `from` posterior a `to`
// es un rango imposible → 400 invalid.
func (s *auditService) ListAccessEvents(ctx context.Context, filter AuditFilter) (AccessEventList, error) {
	filter, err := normalizeAuditFilter(filter)
	if err != nil {
		return AccessEventList{}, err
	}
	events, err := s.repo.ListLoginEvents(ctx, filter)
	if err != nil {
		return AccessEventList{}, fmt.Errorf("listar accesos: %w", err)
	}
	total, err := s.repo.CountLoginEvents(ctx, filter)
	if err != nil {
		return AccessEventList{}, fmt.Errorf("contar accesos: %w", err)
	}
	items := make([]AccessEventItem, 0, len(events))
	for _, event := range events {
		items = append(items, AccessEventItemFrom(event))
	}
	return AccessEventList{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset}, nil
}

// ListAdminActions devuelve una página del historial de acciones
// administrativas (FR-024/P22), con las mismas reglas de filtro, orden y
// paginación que ListAccessEvents. El filtro por cuenta incluye a quien hizo la
// acción y a la cuenta objetivo (lo resuelve la consulta; FR-024). Es SOLO
// lectura (FR-025).
func (s *auditService) ListAdminActions(ctx context.Context, filter AuditFilter) (AdminActionList, error) {
	filter, err := normalizeAuditFilter(filter)
	if err != nil {
		return AdminActionList{}, err
	}
	entries, err := s.repo.ListAdminActions(ctx, filter)
	if err != nil {
		return AdminActionList{}, fmt.Errorf("listar acciones: %w", err)
	}
	total, err := s.repo.CountAdminActions(ctx, filter)
	if err != nil {
		return AdminActionList{}, fmt.Errorf("contar acciones: %w", err)
	}
	items := make([]AdminActionItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, AdminActionItemFrom(entry))
	}
	return AdminActionList{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset}, nil
}

// normalizeAuditFilter valida el rango y normaliza la paginación de un listado
// de auditoría (P14/P22): límite por defecto si no llega (o no es positivo),
// tope duro, offset no negativo y semirango `[from, to)` con `from <= to`.
func normalizeAuditFilter(filter AuditFilter) (AuditFilter, error) {
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return AuditFilter{}, apperr.Invalid(
			"El rango de fechas no es válido",
			apperr.WithDetails(map[string]any{"from": "no puede ser posterior a to"}),
		)
	}
	if filter.Limit <= 0 {
		filter.Limit = paginate.DefaultLimit
	}
	if filter.Limit > paginate.MaxLimit {
		filter.Limit = paginate.MaxLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return filter, nil
}

// logError registra el fallo con el logger de la petición (que lleva el
// request_id) o, si no hay, con el logger base del servicio.
func (s *auditService) logError(ctx context.Context, message string, err error) {
	logger := s.log
	if requestLogger := httpserver.RequestLoggerFromContext(ctx); requestLogger != nil {
		logger = requestLogger
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Error(message, slog.Any("error", err))
}

// actionFromRoute traduce method+path a la acción administrativa sensible del
// registro (FR-023) y su objetivo. Devuelve false cuando la petición no
// corresponde a ninguna acción (p. ej. un GET del panel o las rutas de
// auditoría): esas peticiones no dejan fila (R23). La tabla es la única fuente
// que comparten las denegaciones (result='denied') y los rechazos por datos
// inválidos (result='failure', P20).
func actionFromRoute(method, path string) (audit.Action, bool) {
	segments := splitPath(path)
	if len(segments) < 4 || segments[0] != "api" || segments[1] != "v1" || segments[2] != "admin" {
		return audit.Action{}, false
	}
	resource := segments[3]
	hasID := len(segments) >= 5
	id := parsePathUUID(segments, 4)

	action := audit.Action{}

	switch {
	case resource == "usuarios" && method == http.MethodPost && !hasID:
		action.Code = audit.ActionUserCreate
		action.TargetKind = audit.TargetUser
	case resource == "usuarios" && method == http.MethodPost && hasID &&
		len(segments) == 6 && segments[5] == "password":
		action.Code = audit.ActionUserPasswordReset
		action.TargetKind = audit.TargetUser
		action.TargetUserID = id
	case resource == "usuarios" && method == http.MethodPatch && hasID:
		action.Code = audit.ActionUserUpdate
		action.TargetKind = audit.TargetUser
		action.TargetUserID = id
	case resource == "roles" && method == http.MethodPost && !hasID:
		action.Code = audit.ActionRoleCreate
		action.TargetKind = audit.TargetRole
	case resource == "roles" && method == http.MethodPatch && hasID:
		action.Code = audit.ActionRoleUpdate
		action.TargetKind = audit.TargetRole
		action.TargetRoleID = id
	case resource == "roles" && method == http.MethodDelete && hasID:
		action.Code = audit.ActionRoleDelete
		action.TargetKind = audit.TargetRole
		action.TargetRoleID = id
	default:
		return audit.Action{}, false
	}
	return action, true
}

// splitPath parte la ruta en segmentos no vacíos.
func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

// parsePathUUID lee el segmento index de la ruta y lo convierte a UUID; devuelve
// nil si no está o no es un UUID válido.
func parsePathUUID(segments []string, index int) *uuid.UUID {
	if index >= len(segments) {
		return nil
	}
	id, err := uuid.Parse(segments[index])
	if err != nil {
		return nil
	}
	return &id
}

// auditService implementa el puerto que consume middleware/authz.
var _ audit.Recorder = (*auditService)(nil)
