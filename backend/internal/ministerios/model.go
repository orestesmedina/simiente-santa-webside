// Package ministerios es el dominio de ministerios (F6): la sección pública
// «/ministerios», su bloque en la portada y la gestión desde el panel. Un
// ministerio es un área de servicio con nombre y descripción obligatorios en
// español, contenido bilingüe opcional por campo, un encargado **opcional** con
// su ficha (nombre + un solo medio de contacto) que se publica con su propio
// interruptor, un enlace opcional al grupo de WhatsApp, horario y lugar como
// textos libres, imagen opcional, orden manual y estado de publicación.
//
// Sigue las capas handler → service → repository (§II): los handlers no tocan
// SQL y el service no conoce net/http ni pgx. Este archivo reúne las entidades
// del dominio (con github.com/google/uuid), los DTOs de entrada del contrato
// OpenAPI con sus etiquetas `validate` (R3-1) y **los DTOs públicos**: la
// privacidad del encargado (FR-026/SC-014) es estructural —`MinisterioPublico`
// carece de campos para el nombre y el contacto fuera de su ficha `leader`— y su
// composición vive en un único lugar (`service_public.go`, P6-2/RG6-1).
//
// Aquí no hay lógica de negocio: las reglas (normalización, límites, contacto
// excluyente, enlace https, español por campo, visibilidad del encargado) viven
// en el service.
package ministerios

import (
	"time"

	"github.com/google/uuid"
)

// PublicationState es el estado de publicación de un ministerio (FR-012): la
// publicación es por elemento y «se publica al guardar» (FR-013).
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

// --- Entidades ---

// Ministry es un ministerio (FR-001): sus datos en ambos idiomas, el encargado
// como **campos** —nombre, un único medio de contacto opcional y el interruptor
// `LeaderVisible` que gobierna solo su publicación (Q9/FR-025)—, el enlace al
// grupo de WhatsApp, horario y lugar como textos, su imagen con texto
// alternativo, el orden manual y su estado.
//
// Los campos `*Idioma` son `*string`: nil es "sin traducción" (y el español es
// obligatorio en nombre y descripción, por `CHECK` de la tabla). Los datos del
// encargado se conservan aunque `LeaderVisible` esté apagado: el dato guardado
// no es el dato publicado (el panel los ve; el visitante no, FR-026).
type Ministry struct {
	ID               uuid.UUID
	NameEs           string
	NameEn           *string
	DescriptionEs    string
	DescriptionEn    *string
	LeaderNameEs     *string
	LeaderNameEn     *string
	LeaderPhone      *string
	LeaderEmail      *string
	LeaderVisible    bool
	WhatsappURL      *string
	ScheduleEs       *string
	ScheduleEn       *string
	PlaceEs          *string
	PlaceEn          *string
	ImageFile        *string
	ImageAltEs       *string
	ImageAltEn       *string
	PublicationState PublicationState
	SortOrder        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// MinistryListItem es una fila del panel: el ministerio completo más el aviso
// de enlace compartido, que se **recalcula en cada lectura** y nunca se persiste
// (R6-11): `true` cuando otro ministerio usa el mismo `whatsappUrl`. Es un aviso
// no bloqueante, no una regla: el guardado procede igual.
type MinistryListItem struct {
	Ministry
	WhatsappURLShared bool
}

// --- DTO de entrada del contrato (etiquetas `validate`) ---
//
// Los campos `*En` son `string` con `omitempty`: `""`/espacios significa "sin
// traducción" y el service lo persiste como NULL (analyze I6).

// MinisterioInput es el cuerpo de POST /api/v1/admin/ministerios (FR-014/FR-019):
// solo `nameEs` y `descriptionEs` son obligatorios. `LeaderVisible` es `bool`
// —no un puntero— porque el contrato lo declara con `default: false`: la clave
// ausente y la clave en `false` significan lo mismo, «Oculto por defecto»
// (P6-3). Las reglas de coherencia (español por campo, contacto excluyente,
// enlace https, alt con imagen) las valida el service sobre el DTO normalizado.
type MinisterioInput struct {
	NameEs           string `json:"nameEs" validate:"required,min=1,max=160"`
	NameEn           string `json:"nameEn" validate:"omitempty,max=160"`
	DescriptionEs    string `json:"descriptionEs" validate:"required,min=1,max=2000"`
	DescriptionEn    string `json:"descriptionEn" validate:"omitempty,max=2000"`
	LeaderNameEs     string `json:"leaderNameEs" validate:"omitempty,max=160"`
	LeaderNameEn     string `json:"leaderNameEn" validate:"omitempty,max=160"`
	LeaderPhone      string `json:"leaderPhone" validate:"omitempty,phone,max=32"`
	LeaderEmail      string `json:"leaderEmail" validate:"omitempty,email,max=254"`
	LeaderVisible    bool   `json:"leaderVisible"`
	WhatsappURL      string `json:"whatsappUrl" validate:"omitempty,url,max=500"`
	ScheduleEs       string `json:"scheduleEs" validate:"omitempty,max=200"`
	ScheduleEn       string `json:"scheduleEn" validate:"omitempty,max=200"`
	PlaceEs          string `json:"placeEs" validate:"omitempty,max=200"`
	PlaceEn          string `json:"placeEn" validate:"omitempty,max=200"`
	ImageFile        string `json:"imageFile" validate:"omitempty,max=120"`
	ImageAltEs       string `json:"imageAltEs" validate:"omitempty,max=300"`
	ImageAltEn       string `json:"imageAltEn" validate:"omitempty,max=300"`
	PublicationState string `json:"publicationState" validate:"omitempty,oneof=draft published"`
	SortOrder        int    `json:"sortOrder"`
}

// --- DTOs públicos (salida) ---
//
// La privacidad del encargado es **estructural** (FR-026/SC-014, P6-2/RG6-1):
// `MinisterioPublico` solo tiene el nombre y el contacto **dentro** de su ficha
// `leader`, y esa ficha solo se construye cuando el equipo publicó el
// interruptor y hay nombre (`service_public.go`). Con la ficha no publicada el
// tipo no tiene por dónde transportar el dato: no hay campos de encargado
// sueltos ni `LeaderVisible`, y `publicationState` tampoco existe (el DTO
// público nunca lo expone).

// MinisterioPublico es un ministerio **publicado** para el visitante (FR-002),
// ya resuelto a `lang`: nombre y descripción siempre; el resto de claves se
// omiten cuando el dato no existe (SC-010/SC-017: sin líneas vacías ni botones
// rotos).
type MinisterioPublico struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schedule    *string        `json:"schedule,omitempty"`
	Place       *string        `json:"place,omitempty"`
	Leader      *LiderPublico  `json:"leader,omitempty"`
	WhatsappURL *string        `json:"whatsappUrl,omitempty"`
	Image       *ImagenPublica `json:"image,omitempty"`
}

// LiderPublico es la ficha **publicada** del encargado (FR-025/Q8/Q9): nombre y,
// como mucho, **un** medio de contacto (celular **o** correo, nunca ambos).
// Solo se construye cuando `leaderVisible` está activo (FR-026); un encargado
// con nombre y sin contacto deja la ficha con solo `Name` (no es un error).
type LiderPublico struct {
	Name  string  `json:"name"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`
}

// ImagenPublica es la imagen de un elemento con su texto alternativo accesible
// (FR-020), con la URL relativa que sirve `GET /api/v1/media/{fileName}`.
type ImagenPublica struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}
