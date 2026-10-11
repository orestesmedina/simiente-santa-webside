package portada

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// service_audit.go implementa el registro de auditoría del dominio (T321):
// resuelve método+ruta → acción `home.*` de las rutas de F3 e implementa
// `audit.Recorder` para que el middleware authz registre la denegación de un
// permiso (`result='denied'`), además del helper de rechazo por DTO/JSON
// inválido (`result='failure'`). Las denegaciones y los rechazos son
// best-effort (R3-11): nunca cambian la respuesta ni transportan el cuerpo de la
// petición (FR-026 de F2).
//
// El éxito y el fallo de una mutación NO pasan por aquí: van en la misma
// transacción de la mutación (repository + service_admin.go).

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
// ruta (R3-11). Una petición sin acción asociada (p. ej. la lectura del
// agregado) no deja fila.
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
// fallo se queda en el log y no cambia la respuesta (R3-11).
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
// registro (FR-023) y su objetivo (`target_kind='content'`). Devuelve false
// cuando la petición no corresponde a ninguna acción (lecturas del panel o
// rutas ajenas): esas no dejan fila. Es la única tabla que comparten las
// denegaciones y los rechazos.
func auditActionFromRoute(method, path string) (audit.Action, bool) {
	segments := splitPath(path)
	if len(segments) < 5 ||
		segments[0] != "api" || segments[1] != "v1" ||
		segments[2] != "admin" || segments[3] != "portada" {
		return audit.Action{}, false
	}
	resource := segments[4]
	hasID := len(segments) >= 6

	// Elemento del objetivo: el UUID de la ruta (analyze M5: "sin nombre, el id").
	element := ""
	if hasID {
		if id, err := uuid.Parse(segments[5]); err == nil {
			element = id.String()
		}
	}

	switch {
	// Subida de imágenes: se audita por sí misma (analyze I8).
	case resource == "imagenes" && method == http.MethodPost && !hasID:
		return contentRouteAction(audit.ActionHomeImageUpload, sectionImage, ""), true

	// Singletons (PUT).
	case resource == "identidad" && method == http.MethodPut && !hasID:
		return contentRouteAction(audit.ActionHomeIdentityUpdate, sectionIdentity, ""), true
	case resource == "quienes-somos" && method == http.MethodPut && !hasID:
		return contentRouteAction(audit.ActionHomeAboutUpdate, sectionAbout, ""), true
	case resource == "contacto" && method == http.MethodPut && !hasID:
		return contentRouteAction(audit.ActionHomeContactUpdate, sectionContact, ""), true

	// Horario.
	case resource == "horario" && method == http.MethodPost && !hasID:
		return contentRouteAction(audit.ActionHomeScheduleCreate, sectionSchedule, ""), true
	case resource == "horario" && method == http.MethodPatch && hasID:
		return contentRouteAction(audit.ActionHomeScheduleUpdate, sectionSchedule, element), true
	case resource == "horario" && method == http.MethodDelete && hasID:
		return contentRouteAction(audit.ActionHomeScheduleDelete, sectionSchedule, element), true

	// WhatsApp.
	case resource == "whatsapp" && method == http.MethodPost && !hasID:
		return contentRouteAction(audit.ActionHomeWhatsappCreate, sectionWhatsapp, ""), true
	case resource == "whatsapp" && method == http.MethodPatch && hasID:
		return contentRouteAction(audit.ActionHomeWhatsappUpdate, sectionWhatsapp, element), true
	case resource == "whatsapp" && method == http.MethodDelete && hasID:
		return contentRouteAction(audit.ActionHomeWhatsappDelete, sectionWhatsapp, element), true

	// Redes.
	case resource == "redes" && method == http.MethodPost && !hasID:
		return contentRouteAction(audit.ActionHomeSocialCreate, sectionSocial, ""), true
	case resource == "redes" && method == http.MethodPatch && hasID:
		return contentRouteAction(audit.ActionHomeSocialUpdate, sectionSocial, element), true
	case resource == "redes" && method == http.MethodDelete && hasID:
		return contentRouteAction(audit.ActionHomeSocialDelete, sectionSocial, element), true
	}
	return audit.Action{}, false
}

// contentRouteAction arma la acción de contenido de una ruta (sin actor ni
// resultado: los fija quien la registra).
func contentRouteAction(code, section, element string) audit.Action {
	return audit.Action{
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: contentLabel(section, element),
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

// RecordAction persiste una acción de auditoría fuera de la transacción de una
// mutación (denegaciones y rechazos best-effort). El repository es el único
// punto que escribe `admin_actions` (R3-11).
func (r *repository) RecordAction(ctx context.Context, action audit.Action) error {
	return r.insertAdminAction(ctx, action)
}

// service implementa el puerto audit.Recorder que consume middleware/authz.
var _ audit.Recorder = (*service)(nil)
