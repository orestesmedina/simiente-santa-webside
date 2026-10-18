package medios

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// service_audit.go implementa el registro de auditoría del dominio (T922/P9-14,
// patrón de F4/F6/F7): resuelve método+ruta → acción `media.*` de las rutas del
// módulo e implementa `audit.Recorder` para que el middleware authz registre la
// denegación de un permiso (`result='denied'`), además del helper de rechazo por
// JSON o DTO inválido (`result='failure'`).
//
// Las denegaciones y los rechazos son **best-effort** (R7-10): nunca cambian la
// respuesta ni transportan el cuerpo de la petición (regla de F2). El éxito de
// una mutación **no** pasa por aquí: va en la misma transacción que la mutación
// (repository + service_admin.go/service_sync.go), con la regla exacta de
// códigos de P9-14.
//
// Rutas del módulo (P9-14): `/api/v1/admin/medios/buscar-nuevos` (la única
// búsqueda), `…/{id}` (edición), `…/{id}/unir` y `…/{id}/marcas/resolver`. La
// etiqueta del objetivo es `Medios · <id>` (el UUID de la ruta; la búsqueda usa
// `Medios · Búsqueda`, sin elemento) y el `target_kind` es `content` (las fichas
// son contenido, igual que en F3–F7). F9 no tiene `create` ni `delete`: no hay
// alta manual ni borrado (FR-019).

// Rejection describe una petición rechazada por JSON o DTO inválido antes de
// llegar al service. El dominio resuelve la acción desde method+path; por diseño
// no lleva el cuerpo de la petición.
type Rejection struct {
	ActorUserID *uuid.UUID
	Method      string
	Path        string
}

// RecordDenied implementa audit.Recorder: registra la denegación de un permiso
// (`result='denied'`) resolviendo la acción y el objetivo desde el método y la
// ruta. Una petición sin acción asociada (las lecturas del panel y las rutas
// ajenas al módulo) no deja fila: el registro es de mutaciones.
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
// service (JSON o DTO inválido), con `result='failure'`. Es best-effort: su fallo
// se queda en el log y no cambia la respuesta (R7-10).
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
// registro y su objetivo (`target_kind='content'`, sin FK). Devuelve false
// cuando la petición no corresponde a ninguna acción (las lecturas del panel, la
// superficie pública o rutas ajenas al módulo): esas no dejan fila.
func auditActionFromRoute(method, path string) (audit.Action, bool) {
	segments := splitPath(path)
	if len(segments) < 5 || segments[0] != "api" || segments[1] != "v1" ||
		segments[2] != "admin" || segments[3] != "medios" {
		return audit.Action{}, false
	}

	// Búsqueda de episodios nuevos: `…/medios/buscar-nuevos` con POST (la
	// etiqueta base la comparte con el éxito de la búsqueda, P9-14).
	if len(segments) == 5 && segments[4] == "buscar-nuevos" {
		if method != http.MethodPost {
			return audit.Action{}, false
		}
		return contentRouteAction(audit.ActionMediaSync, busquedaLabel), true
	}

	// Elemento del objetivo: el UUID de la ruta (el título no viaja en la
	// petición, así que la etiqueta lleva el id, como en F2–F7).
	id, err := uuid.Parse(segments[4])
	if err != nil {
		return audit.Action{}, false
	}
	label := etiquetaContenido(id.String())

	switch {
	case len(segments) == 5 && method == http.MethodPatch:
		return contentRouteAction(audit.ActionMediaDescriptionUpdate, label), true
	case len(segments) == 6 && segments[5] == "unir" && method == http.MethodPost:
		return contentRouteAction(audit.ActionMediaMerge, label), true
	case len(segments) == 7 && segments[5] == "marcas" && segments[6] == "resolver" &&
		method == http.MethodPost:
		return contentRouteAction(audit.ActionMediaMarkResolve, label), true
	default:
		return audit.Action{}, false
	}
}

// contentRouteAction arma la acción de contenido de una ruta (sin actor ni
// resultado: los fija quien la registra).
func contentRouteAction(code, label string) audit.Action {
	return audit.Action{
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
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
