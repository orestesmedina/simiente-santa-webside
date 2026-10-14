package ministerios

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// service_audit.go implementa el registro de auditoría del dominio (T616):
// resuelve método+ruta → acción `ministry.*` de las rutas de F6 e implementa
// `audit.Recorder` para que el middleware authz registre la denegación de un
// permiso (`result='denied'`), además del helper de rechazo por JSON o DTO
// inválido (`result='failure'`).
//
// Las denegaciones y los rechazos son best-effort (R6-12): nunca cambian la
// respuesta ni transportan el cuerpo de la petición (FR-026 de F2). El éxito de
// una mutación NO pasa por aquí: va en la misma transacción que la mutación
// (repository + service_admin.go), con la regla exacta de códigos de A6-10.
//
// Rutas del módulo (R6-12): `/api/v1/admin/ministerios`, `…/{id}` y
// `…/imagenes`. La etiqueta del objetivo es `Ministerios · <id>` (el UUID de la
// ruta; en el alta, sin elemento, queda `Ministerios`) y, en las subidas,
// `Ministerios · Imagen` —el nombre del archivo no viaja en la ruta; la subida
// correcta deja `Ministerios · Imagen · <fileName>` desde el service—.

// Rejection describe una petición rechazada por JSON o DTO inválido antes de
// llegar al service. El dominio resuelve la acción desde method+path; por diseño
// no lleva el cuerpo de la petición (FR-026).
type Rejection struct {
	ActorUserID *uuid.UUID
	Method      string
	Path        string
}

// RecordDenied implementa audit.Recorder: registra la denegación de un permiso
// (`result='denied'`) resolviendo la acción y el objetivo desde el método y la
// ruta (R6-12). Una petición sin acción asociada (las lecturas del panel y las
// rutas ajenas al módulo) no deja fila.
func (s *service) RecordDenied(ctx context.Context, denial audit.Denial) error {
	action, ok := auditActionFromRoute(denial.Method, denial.Path)
	if !ok {
		return nil
	}
	action.ActorUserID = denial.ActorUserID
	action.Result = audit.ResultDenied
	if err := s.repository.RecordAction(ctx, action); err != nil {
		return fmt.Errorf("registrar la denegación: %w", err)
	}
	return nil
}

// RecordRejectedBestEffort registra un intento rechazado antes de llegar al
// service (JSON o DTO inválido), con `result='failure'`. Es best-effort: su
// fallo se queda en el log y no cambia la respuesta (R6-12).
func (s *service) RecordRejectedBestEffort(ctx context.Context, rejection Rejection) {
	action, ok := auditActionFromRoute(rejection.Method, rejection.Path)
	if !ok {
		return
	}
	action.ActorUserID = rejection.ActorUserID
	action.Result = audit.ResultFailure
	s.recordActionBestEffort(ctx, action)
}

// auditActionFromRoute traduce method+path a la acción administrativa del
// registro (R6-12) y su objetivo (`target_kind='content'`, sin FK). Devuelve
// false cuando la petición no corresponde a ninguna acción (las lecturas del
// panel o rutas ajenas al módulo): esas no dejan fila. Es la única tabla que
// comparten las denegaciones y los rechazos.
func auditActionFromRoute(method, path string) (audit.Action, bool) {
	segments := splitPath(path)
	if len(segments) < 4 || segments[0] != "api" || segments[1] != "v1" || segments[2] != "admin" {
		return audit.Action{}, false
	}
	if segments[3] != "ministerios" {
		return audit.Action{}, false
	}

	// Subida de imágenes: se audita por sí misma (R6-7).
	if len(segments) == 5 && segments[4] == "imagenes" && method == http.MethodPost {
		return contentRouteAction(audit.ActionMinistryImageUpload, sectionImage), true
	}

	// Alta: sin recurso.
	if len(segments) == 4 && method == http.MethodPost {
		return contentRouteAction(audit.ActionMinistryCreate, ""), true
	}

	// Elemento del objetivo: el UUID de la ruta (sin nombre, el id).
	if len(segments) != 5 {
		return audit.Action{}, false
	}
	id, err := uuid.Parse(segments[4])
	if err != nil {
		return audit.Action{}, false
	}
	switch method {
	case http.MethodPatch:
		return contentRouteAction(audit.ActionMinistryUpdate, id.String()), true
	case http.MethodDelete:
		return contentRouteAction(audit.ActionMinistryDelete, id.String()), true
	default:
		return audit.Action{}, false
	}
}

// sectionImage es la sección del objetivo de una subida (`… · Imagen`).
const sectionImage = "Imagen"

// contentRouteAction arma la acción de contenido de una ruta (sin actor ni
// resultado: los fija quien la registra).
func contentRouteAction(code, element string) audit.Action {
	return audit.Action{
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: contentLabel(element),
	}
}

// splitPath parte la ruta en segmentos no vacíos.
func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

// service implementa el puerto audit.Recorder que consume middleware/authz.
var _ audit.Recorder = (*service)(nil)
