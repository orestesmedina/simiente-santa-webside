// Package portada es el dominio de la portada e información general (F3):
// identidad, «quiénes somos», contacto, horario de servicios, canales de
// WhatsApp, redes sociales y las imágenes de la portada. Sigue las capas
// handler → service → repository (§II): los handlers no tocan SQL y el service
// no conoce net/http ni pgx.
//
// Este archivo reúne las entidades del dominio (con github.com/google/uuid) y
// los DTOs del contrato OpenAPI con sus etiquetas `validate` (R3-1). No tiene
// lógica de negocio: las reglas (normalización, límites, fallback de idioma)
// viven en el service (T317–T320).
package portada

import (
	"time"

	"github.com/google/uuid"
)

// PublicationState es el estado de publicación de un elemento (FR-013): la
// publicación es por elemento, también en los singletons.
type PublicationState string

const (
	// StateDraft no es visible para el visitante.
	StateDraft PublicationState = "draft"
	// StatePublished es visible para el visitante.
	StatePublished PublicationState = "published"
)

// Valid indica si el estado pertenece al registro del contrato.
func (s PublicationState) Valid() bool {
	return s == StateDraft || s == StatePublished
}

// Tipos de canal de WhatsApp (FR-005/R3-6).
const (
	// KindDirect es un canal de mensaje directo: el destino es un teléfono.
	KindDirect = "direct"
	// KindGroup es un enlace de grupo: el destino es una URL https de WhatsApp.
	KindGroup = "group"
)

// --- Entidades ---

// Identity es el singleton de identidad de la iglesia (FR-002). Los campos
// traducibles guardan ambos idiomas (`Es` obligatorio, `En` opcional → nil si
// no hay traducción); la resolución de idioma la hace el service público.
type Identity struct {
	ID               uuid.UUID
	NameEs           string
	NameEn           *string
	TaglineEs        *string
	TaglineEn        *string
	MissionEs        *string
	MissionEn        *string
	VisionEs         *string
	VisionEn         *string
	LogoFile         *string
	LogoAltEs        *string
	LogoAltEn        *string
	CoverImageFile   *string
	CoverImageAltEs  *string
	CoverImageAltEn  *string
	PublicationState PublicationState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// About es el singleton del texto «quiénes somos» (FR-003, ≤1.000 caracteres).
type About struct {
	ID               uuid.UUID
	TextEs           string
	TextEn           *string
	PublicationState PublicationState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Contact es el singleton de datos de contacto (FR-007).
type Contact struct {
	ID               uuid.UUID
	AddressEs        string
	AddressEn        *string
	Email            string
	Phone            string
	PublicationState PublicationState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Service es un elemento del horario (FR-004). DayOfWeek 0=domingo…6=sábado;
// StartTime/EndTime son "HH:MM" y EndTime es opcional (nil = sin fin).
type Service struct {
	ID               uuid.UUID
	DayOfWeek        int
	StartTime        string
	EndTime          *string
	NameEs           string
	NameEn           *string
	DescriptionEs    *string
	DescriptionEn    *string
	PlaceEs          string
	PlaceEn          *string
	PublicationState PublicationState
	SortOrder        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// WhatsappChannel es un canal de WhatsApp (FR-005). Kind ∈ {direct, group};
// Destination es el teléfono normalizado o la URL del grupo.
type WhatsappChannel struct {
	ID               uuid.UUID
	Kind             string
	Destination      string
	NameEs           string
	NameEn           *string
	PublicationState PublicationState
	SortOrder        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SocialLink es un enlace de red social del catálogo fijo (FR-006).
type SocialLink struct {
	ID               uuid.UUID
	Network          string
	URL              string
	PublicationState PublicationState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// --- DTOs de entrada del contrato (etiquetas `validate`) ---
//
// Los campos `*En` son `string` con `omitempty`: `""`/espacios significa "sin
// traducción" y el service lo persiste como NULL (analyze I6). Los campos
// anulables del contrato (`null` explícito) llegan también como `""`.

// IdentityInput es el cuerpo de PUT /api/v1/admin/portada/identidad.
type IdentityInput struct {
	NameEs           string `json:"nameEs" validate:"required,min=1,max=160"`
	NameEn           string `json:"nameEn" validate:"omitempty,max=160"`
	TaglineEs        string `json:"taglineEs" validate:"omitempty,max=300"`
	TaglineEn        string `json:"taglineEn" validate:"omitempty,max=300"`
	MissionEs        string `json:"missionEs" validate:"omitempty,max=600"`
	MissionEn        string `json:"missionEn" validate:"omitempty,max=600"`
	VisionEs         string `json:"visionEs" validate:"omitempty,max=600"`
	VisionEn         string `json:"visionEn" validate:"omitempty,max=600"`
	LogoFile         string `json:"logoFile" validate:"omitempty,max=120"`
	LogoAltEs        string `json:"logoAltEs" validate:"omitempty,max=300"`
	LogoAltEn        string `json:"logoAltEn" validate:"omitempty,max=300"`
	CoverImageFile   string `json:"coverImageFile" validate:"omitempty,max=120"`
	CoverImageAltEs  string `json:"coverImageAltEs" validate:"omitempty,max=300"`
	CoverImageAltEn  string `json:"coverImageAltEn" validate:"omitempty,max=300"`
	PublicationState string `json:"publicationState" validate:"required,oneof=draft published"`
}

// AboutInput es el cuerpo de PUT /api/v1/admin/portada/quienes-somos.
type AboutInput struct {
	TextEs           string `json:"textEs" validate:"required,min=1,max=1000"`
	TextEn           string `json:"textEn" validate:"omitempty,max=1000"`
	PublicationState string `json:"publicationState" validate:"required,oneof=draft published"`
}

// ContactInput es el cuerpo de PUT /api/v1/admin/portada/contacto.
type ContactInput struct {
	AddressEs        string `json:"addressEs" validate:"required,min=1,max=300"`
	AddressEn        string `json:"addressEn" validate:"omitempty,max=300"`
	Email            string `json:"email" validate:"required,email,max=254"`
	Phone            string `json:"phone" validate:"required,phone,min=7,max=32"`
	PublicationState string `json:"publicationState" validate:"required,oneof=draft published"`
}

// ScheduleItemInput es el cuerpo de POST /api/v1/admin/portada/horario.
type ScheduleItemInput struct {
	DayOfWeek        int    `json:"dayOfWeek" validate:"min=0,max=6"`
	StartTime        string `json:"startTime" validate:"required"`
	EndTime          string `json:"endTime" validate:"omitempty"`
	NameEs           string `json:"nameEs" validate:"required,min=1,max=160"`
	NameEn           string `json:"nameEn" validate:"omitempty,max=160"`
	DescriptionEs    string `json:"descriptionEs" validate:"omitempty,max=400"`
	DescriptionEn    string `json:"descriptionEn" validate:"omitempty,max=400"`
	PlaceEs          string `json:"placeEs" validate:"required,min=1,max=200"`
	PlaceEn          string `json:"placeEn" validate:"omitempty,max=200"`
	PublicationState string `json:"publicationState" validate:"omitempty,oneof=draft published"`
	SortOrder        int    `json:"sortOrder"`
}

// WhatsappChannelInput es el cuerpo de POST /api/v1/admin/portada/whatsapp.
type WhatsappChannelInput struct {
	NameEs           string `json:"nameEs" validate:"required,min=1,max=120"`
	NameEn           string `json:"nameEn" validate:"omitempty,max=120"`
	Kind             string `json:"kind" validate:"required,oneof=direct group"`
	Destination      string `json:"destination" validate:"required,min=1,max=320"`
	PublicationState string `json:"publicationState" validate:"omitempty,oneof=draft published"`
	SortOrder        int    `json:"sortOrder"`
}

// SocialLinkInput es el cuerpo de POST /api/v1/admin/portada/redes.
type SocialLinkInput struct {
	Network          string `json:"network" validate:"required"`
	URL              string `json:"url" validate:"required,url,max=500"`
	PublicationState string `json:"publicationState" validate:"omitempty,oneof=draft published"`
}
