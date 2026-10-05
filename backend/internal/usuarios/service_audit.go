package usuarios

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
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

// AuditRepository es el puerto de escritura que el servicio de auditoría
// necesita del repositorio (lo define quien lo consume, arq. R3).
type AuditRepository interface {
	InsertLoginEvent(ctx context.Context, event audit.Event) (LoginEvent, error)
	InsertAdminAction(ctx context.Context, action audit.Action) (AdminAction, error)
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
	action, ok := denialAction(denial)
	if !ok {
		return nil
	}
	return s.RecordAction(ctx, action)
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

// denialAction traduce method+path a la acción administrativa denegada. Devuelve
// false cuando la petición no corresponde a una acción sensible (p. ej. un GET
// del panel): esas denegaciones no entran en `admin_actions` (R23).
func denialAction(denial audit.Denial) (audit.Action, bool) {
	segments := splitPath(denial.Path)
	if len(segments) < 4 || segments[0] != "api" || segments[1] != "v1" || segments[2] != "admin" {
		return audit.Action{}, false
	}
	resource := segments[3]
	hasID := len(segments) >= 5
	id := parsePathUUID(segments, 4)

	action := audit.Action{
		ActorUserID: denial.ActorUserID,
		Result:      audit.ResultDenied,
	}

	switch {
	case resource == "usuarios" && denial.Method == http.MethodPost && !hasID:
		action.Code = audit.ActionUserCreate
		action.TargetKind = audit.TargetUser
	case resource == "usuarios" && denial.Method == http.MethodPost && hasID &&
		len(segments) == 6 && segments[5] == "password":
		action.Code = audit.ActionUserPasswordReset
		action.TargetKind = audit.TargetUser
		action.TargetUserID = id
	case resource == "usuarios" && denial.Method == http.MethodPatch && hasID:
		action.Code = audit.ActionUserUpdate
		action.TargetKind = audit.TargetUser
		action.TargetUserID = id
	case resource == "roles" && denial.Method == http.MethodPost && !hasID:
		action.Code = audit.ActionRoleCreate
		action.TargetKind = audit.TargetRole
	case resource == "roles" && denial.Method == http.MethodPatch && hasID:
		action.Code = audit.ActionRoleUpdate
		action.TargetKind = audit.TargetRole
		action.TargetRoleID = id
	case resource == "roles" && denial.Method == http.MethodDelete && hasID:
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
