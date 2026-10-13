// Package agenda es el dominio de eventos y actividades (F4): «lo que viene» del
// sitio público y su gestión desde el panel. Cubre tres conceptos sobre cuatro
// tablas: eventos (con su lista explícita de fechas), actividades (de fecha
// única) y el catálogo de categorías compartido por ambos.
//
// Sigue las capas handler → service → repository (§II): los handlers no tocan
// SQL y el service no conoce net/http ni pgx. Este archivo reúne las entidades
// del dominio (con github.com/google/uuid) y los DTOs de entrada del contrato
// OpenAPI con sus etiquetas `validate` (R3-1); no tiene lógica de negocio: las
// reglas (normalización, límites, fechas, regla de "próximo") viven en el
// service (T417–T421).
package agenda

import (
	"time"

	"github.com/google/uuid"
)

// DateFormat es el formato de las fechas de calendario del lugar ("YYYY-MM-DD").
// No es un timestamp: no se convierte de zona horaria (FR-009/R4-3).
const DateFormat = "2006-01-02"

// TimeFormat es el formato de las horas del lugar ("HH:MM", 24 h). El formato
// a.m./p.m. es presentación del cliente.
const TimeFormat = "15:04"

// PublicationState es el estado de publicación de un elemento de contenido
// (FR-013): la publicación es por elemento (eventos y actividades; las
// categorías son catálogo y no tienen estado).
type PublicationState string

const (
	// StateDraft no es visible para el visitante.
	StateDraft PublicationState = "draft"
	// StatePublished es visible para el visitante mientras su fecha sea próxima.
	StatePublished PublicationState = "published"
)

// Valid indica si el estado pertenece al registro del contrato.
func (s PublicationState) Valid() bool {
	return s == StateDraft || s == StatePublished
}

// --- Entidades ---

// Category es una categoría del catálogo compartido (Q6/R4-5): nombre en español
// obligatorio, en inglés opcional y un orden de presentación.
type Category struct {
	ID        uuid.UUID
	NameEs    string
	NameEn    *string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Event es un evento (encuentro periódico, FR-018): sus datos y su estado. La
// periodicidad NO se calcula: es la lista de EventDate más el texto opcional de
// frecuencia (Q1).
type Event struct {
	ID               uuid.UUID
	TitleEs          string
	TitleEn          *string
	FrequencyEs      *string
	FrequencyEn      *string
	DetailEs         string
	DetailEn         *string
	PlaceEs          string
	PlaceEn          *string
	MapURL           *string
	OrganizerEs      *string
	OrganizerEn      *string
	CategoryID       *uuid.UUID
	ImageFile        *string
	ImageAltEs       *string
	ImageAltEn       *string
	PublicationState PublicationState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// EventDate es una fecha concreta de un evento (Q1/Q2): un día con hora de
// inicio y fin opcionales (rango dentro del mismo día). Date es una fecha de
// calendario "YYYY-MM-DD" (DateFormat) y las horas "HH:MM" (TimeFormat).
type EventDate struct {
	ID        uuid.UUID
	EventID   uuid.UUID
	Date      string
	StartTime *string
	EndTime   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Activity es una actividad (acción puntual, FR-019/Decisión 3): todos sus datos
// y UNA fecha concreta con su horario opcional, en la misma fila.
type Activity struct {
	ID               uuid.UUID
	TitleEs          string
	TitleEn          *string
	DetailEs         string
	DetailEn         *string
	PlaceEs          string
	PlaceEn          *string
	MapURL           *string
	OrganizerEs      *string
	OrganizerEn      *string
	CategoryID       *uuid.UUID
	ImageFile        *string
	ImageAltEs       *string
	ImageAltEn       *string
	Date             string
	StartTime        *string
	EndTime          *string
	PublicationState PublicationState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// --- Extracto de la portada (US7/FR-025, enmienda Q8) ---

// NextItemKind identifica el concepto de un ítem del extracto «Lo que viene» de
// la portada (Decisión 3): un evento o una actividad.
type NextItemKind string

const (
	// NextItemEvent es un evento (encuentro periódico).
	NextItemEvent NextItemKind = "event"
	// NextItemActivity es una actividad (acción puntual).
	NextItemActivity NextItemKind = "activity"
)

// NextItem es un ítem del extracto de la portada (US7/FR-025): solo lo que la
// tarjeta compacta muestra —concepto, título, la PRÓXIMA fecha con su horario
// opcional, lugar e imagen con su texto alternativo—, ya sin detalle, categoría
// ni lista de fechas (esos datos viven en la sección `/eventos`).
type NextItem struct {
	Kind       NextItemKind
	ID         uuid.UUID
	TitleEs    string
	TitleEn    *string
	Date       string
	StartTime  *string
	EndTime    *string
	PlaceEs    string
	PlaceEn    *string
	ImageFile  *string
	ImageAltEs *string
	ImageAltEn *string
}

// --- DTOs de entrada del contrato (etiquetas `validate`) ---
//
// Los campos `*En` son `string` con `omitempty`: `""`/espacios significa "sin
// traducción" y el service lo persiste como NULL (analyze I6). Los campos
// anulables del contrato (`null` explícito en un PATCH) usan Optional[T] para
// distinguir "ausente" de "vaciar".

// FechaInput es una fecha de un evento en el cuerpo de alta o edición (FR-004).
type FechaInput struct {
	Date      string `json:"date" validate:"required"`
	StartTime string `json:"startTime" validate:"omitempty"`
	EndTime   string `json:"endTime" validate:"omitempty"`
}

// EventoInput es el cuerpo de POST /api/v1/admin/eventos (FR-010/FR-015).
// `dates` exige al menos una fecha; el tope (≤50) y las reglas de calendario y
// de horario las valida el service (T419).
type EventoInput struct {
	TitleEs          string       `json:"titleEs" validate:"required,min=1,max=160"`
	TitleEn          string       `json:"titleEn" validate:"omitempty,max=160"`
	FrequencyEs      string       `json:"frequencyEs" validate:"omitempty,max=160"`
	FrequencyEn      string       `json:"frequencyEn" validate:"omitempty,max=160"`
	DetailEs         string       `json:"detailEs" validate:"required,min=1,max=2000"`
	DetailEn         string       `json:"detailEn" validate:"omitempty,max=2000"`
	PlaceEs          string       `json:"placeEs" validate:"required,min=1,max=200"`
	PlaceEn          string       `json:"placeEn" validate:"omitempty,max=200"`
	MapURL           string       `json:"mapUrl" validate:"omitempty,url,max=500"`
	OrganizerEs      string       `json:"organizerEs" validate:"omitempty,max=160"`
	OrganizerEn      string       `json:"organizerEn" validate:"omitempty,max=160"`
	CategoryID       string       `json:"categoryId" validate:"omitempty"`
	ImageFile        string       `json:"imageFile" validate:"omitempty,max=120"`
	ImageAltEs       string       `json:"imageAltEs" validate:"omitempty,max=300"`
	ImageAltEn       string       `json:"imageAltEn" validate:"omitempty,max=300"`
	PublicationState string       `json:"publicationState" validate:"omitempty,oneof=draft published"`
	Dates            []FechaInput `json:"dates" validate:"required"`
}

// ActividadInput es el cuerpo de POST /api/v1/admin/actividades (FR-011/FR-015):
// una única fecha con horario opcional (sin periodicidad).
type ActividadInput struct {
	TitleEs          string `json:"titleEs" validate:"required,min=1,max=160"`
	TitleEn          string `json:"titleEn" validate:"omitempty,max=160"`
	DetailEs         string `json:"detailEs" validate:"required,min=1,max=2000"`
	DetailEn         string `json:"detailEn" validate:"omitempty,max=2000"`
	PlaceEs          string `json:"placeEs" validate:"required,min=1,max=200"`
	PlaceEn          string `json:"placeEn" validate:"omitempty,max=200"`
	MapURL           string `json:"mapUrl" validate:"omitempty,url,max=500"`
	OrganizerEs      string `json:"organizerEs" validate:"omitempty,max=160"`
	OrganizerEn      string `json:"organizerEn" validate:"omitempty,max=160"`
	CategoryID       string `json:"categoryId" validate:"omitempty"`
	ImageFile        string `json:"imageFile" validate:"omitempty,max=120"`
	ImageAltEs       string `json:"imageAltEs" validate:"omitempty,max=300"`
	ImageAltEn       string `json:"imageAltEn" validate:"omitempty,max=300"`
	Date             string `json:"date" validate:"required"`
	StartTime        string `json:"startTime" validate:"omitempty"`
	EndTime          string `json:"endTime" validate:"omitempty"`
	PublicationState string `json:"publicationState" validate:"omitempty,oneof=draft published"`
}

// CategoriaInput es el cuerpo de POST /api/v1/admin/categorias (Q6): nombre en
// español obligatorio, en inglés opcional y orden opcional.
type CategoriaInput struct {
	NameEs    string `json:"nameEs" validate:"required,min=1,max=120"`
	NameEn    string `json:"nameEn" validate:"omitempty,max=120"`
	SortOrder int    `json:"sortOrder"`
}
