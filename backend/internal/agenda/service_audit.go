package agenda

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// service_audit.go implementa el registro de auditoría del dominio (T422):
// resuelve método+ruta → acción de las rutas de F4 (`event.*`, `activity.*` y
// `category.*`) e implementa `audit.Recorder` para que el middleware authz
// registre la denegación de un permiso (`result='denied'`), además del helper de
// rechazo por JSON o DTO inválido (`result='failure'`).
//
// Las denegaciones y los rechazos son best-effort (R4-10): nunca cambian la
// respuesta ni transportan el cuerpo de la petición (FR-026 de F2). El éxito de
// una mutación NO pasa por aquí: va en la misma transacción que la mutación
// (repository + service_admin.go).

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
// ruta (R4-10). Una petición sin acción asociada (las lecturas del panel y las
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
// fallo se queda en el log y no cambia la respuesta (R4-10).
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
// registro (R4-10) y su objetivo (`target_kind='content'`). Devuelve false
// cuando la petición no corresponde a ninguna acción (las lecturas del panel o
// rutas ajenas al módulo): esas no dejan fila. Es la única tabla que comparten
// las denegaciones y los rechazos.
//
// Rutas de F4 (R4-7): `/api/v1/admin/eventos`, `/api/v1/admin/actividades` y
// `/api/v1/admin/categorias`, con `{id}` para PATCH/DELETE y `/imagenes` para
// las subidas (solo eventos y actividades). El elemento del objetivo es el UUID
// de la ruta; en las subidas, el nombre del archivo no viaja en la ruta, así que
// la etiqueta es la de la sección de imágenes (`Eventos · Imagen`), igual que en
// F3.
func auditActionFromRoute(method, path string) (audit.Action, bool) {
	segments := splitPath(path)
	if len(segments) < 4 || segments[0] != "api" || segments[1] != "v1" || segments[2] != "admin" {
		return audit.Action{}, false
	}

	module, ok := moduleFromRoute(segments[3])
	if !ok {
		return audit.Action{}, false
	}

	hasResource := len(segments) >= 5
	resource := ""
	if hasResource {
		resource = segments[4]
	}

	// Subida de imágenes: se audita por sí misma (analyze I8/R4-10). El catálogo
	// de categorías no tiene imágenes: su código no existe en el registro.
	if hasResource && resource == "imagenes" && method == http.MethodPost {
		if module == ModuleCategories {
			return audit.Action{}, false
		}
		return contentRouteAction(module.ImageCode(), module, sectionImage), true
	}

	// Elemento del objetivo: el UUID de la ruta (sin nombre, el id).
	element := ""
	isElement := false
	if hasResource {
		if id, err := uuid.Parse(resource); err == nil {
			element = id.String()
			isElement = true
		}
	}

	switch {
	case !hasResource && method == http.MethodPost:
		return contentRouteAction(module.CreateCode(), module, ""), true
	case isElement && method == http.MethodPatch:
		return contentRouteAction(module.UpdateCode(), module, element), true
	case isElement && method == http.MethodDelete:
		return contentRouteAction(module.DeleteCode(), module, element), true
	}
	return audit.Action{}, false
}

// moduleFromRoute traduce el segmento de ruta del módulo a su Module.
func moduleFromRoute(segment string) (Module, bool) {
	switch segment {
	case "eventos":
		return ModuleEvents, true
	case "actividades":
		return ModuleActivities, true
	case "categorias":
		return ModuleCategories, true
	default:
		return "", false
	}
}

// sectionImage es la sección del objetivo de una subida (`… · Imagen`).
const sectionImage = "Imagen"

// contentRouteAction arma la acción de contenido de una ruta (sin actor ni
// resultado: los fija quien la registra).
func contentRouteAction(code string, module Module, element string) audit.Action {
	return audit.Action{
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: contentLabel(module, element),
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
