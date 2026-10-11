// Package audit define el plumbing del registro de auditoría: los tipos que
// comparten el middleware authz y el dominio usuarios, y la interfaz Recorder
// que permite registrar una denegación de permiso sin que platform conozca
// ningún dominio (P20; plan §Complexity Tracking 4).
//
// No hace SQL, ni Redis, ni HTTP: solo tipos y validación. Las tablas
// (login_events/admin_actions) y su escritura viven en el dominio usuarios.
package audit

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Result es el desenlace de un intento de acceso o de una acción
// administrativa (FR-022/FR-023).
type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
	ResultDenied  Result = "denied"
)

// Valid indica si el resultado pertenece al registro del contrato.
func (r Result) Valid() bool {
	switch r {
	case ResultSuccess, ResultFailure, ResultDenied:
		return true
	default:
		return false
	}
}

// TargetKind identifica el tipo de objeto sobre el que actúa una acción
// administrativa (FR-023).
type TargetKind string

const (
	TargetUser TargetKind = "user"
	TargetRole TargetKind = "role"
	// TargetContent es un elemento de la portada (F3): identidad, «quiénes
	// somos», contacto, un servicio del horario, un canal de WhatsApp o un
	// enlace de red. No tiene FK (son seis tablas): ambas FK van en NULL y
	// TargetLabel identifica el elemento (R3-11).
	TargetContent TargetKind = "content"
)

// Códigos de acción administrativa (FR-023). Son la lista cerrada de la tabla
// admin_actions: los 8 de F2 y los 15 de F3 (`home.*`).
const (
	ActionUserCreate        = "user.create"
	ActionUserUpdate        = "user.update"
	ActionUserActivate      = "user.activate"
	ActionUserDeactivate    = "user.deactivate"
	ActionUserPasswordReset = "user.password_reset"
	ActionRoleCreate        = "role.create"
	ActionRoleUpdate        = "role.update"
	ActionRoleDelete        = "role.delete"

	// Códigos de F3 (R3-11.1): quién editó qué elemento de la portada.
	ActionHomeIdentityUpdate = "home.identity.update"
	ActionHomeAboutUpdate    = "home.about.update"
	ActionHomeContactUpdate  = "home.contact.update"
	ActionHomeScheduleCreate = "home.schedule.create"
	ActionHomeScheduleUpdate = "home.schedule.update"
	ActionHomeScheduleDelete = "home.schedule.delete"
	ActionHomeWhatsappCreate = "home.whatsapp.create"
	ActionHomeWhatsappUpdate = "home.whatsapp.update"
	ActionHomeWhatsappDelete = "home.whatsapp.delete"
	ActionHomeSocialCreate   = "home.social.create"
	ActionHomeSocialUpdate   = "home.social.update"
	ActionHomeSocialDelete   = "home.social.delete"
	ActionHomeImageUpload    = "home.image.upload"
	ActionHomePublish        = "home.publish"
	ActionHomeUnpublish      = "home.unpublish"
)

// ActionCodes es la lista cerrada de códigos válidos de FR-023, en el orden de
// la tabla de admin_actions. No se debe modificar en runtime.
var ActionCodes = []string{
	ActionUserCreate,
	ActionUserUpdate,
	ActionUserActivate,
	ActionUserDeactivate,
	ActionUserPasswordReset,
	ActionRoleCreate,
	ActionRoleUpdate,
	ActionRoleDelete,
	ActionHomeIdentityUpdate,
	ActionHomeAboutUpdate,
	ActionHomeContactUpdate,
	ActionHomeScheduleCreate,
	ActionHomeScheduleUpdate,
	ActionHomeScheduleDelete,
	ActionHomeWhatsappCreate,
	ActionHomeWhatsappUpdate,
	ActionHomeWhatsappDelete,
	ActionHomeSocialCreate,
	ActionHomeSocialUpdate,
	ActionHomeSocialDelete,
	ActionHomeImageUpload,
	ActionHomePublish,
	ActionHomeUnpublish,
}

// Errores de validación de los tipos de plumbing.
var (
	ErrUnknownActionCode  = errors.New("audit: código de acción desconocido")
	ErrUnknownTargetKind  = errors.New("audit: tipo de objetivo desconocido")
	ErrInvalidResult      = errors.New("audit: resultado inválido")
	ErrTargetMismatch     = errors.New("audit: objetivo incoherente con su tipo")
	ErrMissingTargetLabel = errors.New("audit: falta la etiqueta del objetivo de contenido")
	ErrMissingIP          = errors.New("audit: falta la IP del intento de acceso")
)

// Event es un intento de inicio de sesión (FR-022). UserID es nil cuando el
// correo no corresponde a ninguna cuenta: el intento se registra sin asociarse
// a nada (FR-003/FR-026, sin "cuentas fantasma").
type Event struct {
	UserID *uuid.UUID
	Result Result
	IP     string
}

// Validate comprueba que el evento es un intento de acceso registrable: solo
// success/failure (denied es propio de admin_actions) y con IP de origen.
func (e Event) Validate() error {
	switch e.Result {
	case ResultSuccess, ResultFailure:
	default:
		return fmt.Errorf("%w: %q", ErrInvalidResult, e.Result)
	}
	if e.IP == "" {
		return fmt.Errorf("%w", ErrMissingIP)
	}
	return nil
}

// Action es una acción administrativa sensible (FR-023): quién, qué, sobre qué
// y con qué resultado. Los campos de objetivo son anulables porque una creación
// rechazada no llega a tener id y un rol eliminado deja su FK en nil.
type Action struct {
	ActorUserID  *uuid.UUID
	Code         string
	TargetKind   TargetKind
	TargetUserID *uuid.UUID
	TargetRoleID *uuid.UUID
	TargetLabel  string
	Result       Result
}

// Validate comprueba la coherencia de la acción con el esquema de
// admin_actions: código del registro cerrado de FR-023, resultado válido y solo
// la FK del tipo de objetivo declarado.
func (a Action) Validate() error {
	if !ValidActionCode(a.Code) {
		return fmt.Errorf("%w: %q", ErrUnknownActionCode, a.Code)
	}
	if !a.Result.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidResult, a.Result)
	}
	switch a.TargetKind {
	case TargetUser:
		if a.TargetRoleID != nil {
			return fmt.Errorf("%w: objetivo usuario con target_role_id", ErrTargetMismatch)
		}
	case TargetRole:
		if a.TargetUserID != nil {
			return fmt.Errorf("%w: objetivo rol con target_user_id", ErrTargetMismatch)
		}
	case TargetContent:
		// Un objetivo de contenido no tiene FK (son seis tablas): ambas van en
		// nil y la etiqueta identifica el elemento (R3-11).
		if a.TargetUserID != nil || a.TargetRoleID != nil {
			return fmt.Errorf("%w: objetivo de contenido con una FK de usuario o rol", ErrTargetMismatch)
		}
		if a.TargetLabel == "" {
			return fmt.Errorf("%w", ErrMissingTargetLabel)
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnknownTargetKind, a.TargetKind)
	}
	return nil
}

// ValidActionCode indica si code pertenece al registro cerrado de FR-023.
func ValidActionCode(code string) bool {
	for _, candidate := range ActionCodes {
		if candidate == code {
			return true
		}
	}
	return false
}

// Denial describe la denegación de un permiso para que el dominio resuelva la
// acción y el objetivo a partir del método y la ruta (P20/R23), sin que authz
// necesite conocer los códigos de acción.
type Denial struct {
	ActorUserID *uuid.UUID
	Method      string
	Path        string
}

// Recorder es el puerto que consume middleware.authz para registrar en
// admin_actions la denegación de un permiso. Lo implementa el dominio usuarios,
// que es quien resuelve el código de acción y el objetivo desde la petición.
type Recorder interface {
	RecordDenied(ctx context.Context, denial Denial) error
}
