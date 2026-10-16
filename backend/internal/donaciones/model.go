// Package donaciones es el dominio de la sección «Ofrendas y diezmos» (F7): los
// métodos de donación (cuentas IBAN y números de SINPE Móvil) y los dos textos
// libres de la sección («cómo donar» y «en qué se usan las donaciones»).
//
// Sigue las capas handler → service → repository (§II): los handlers no tocan
// SQL y el service no conoce net/http ni pgx. Este archivo reúne las entidades
// del dominio (con github.com/google/uuid), los DTOs de entrada del contrato
// OpenAPI con sus etiquetas `validate` (R3-1) y **los DTOs públicos**: el DTO
// público de un método (`MetodoDonacionPublico`) carece de `publicationState`
// (el visitante nunca ve borradores) y su composición —`currency` solo en IBAN y
// `holder` solo cuando existe— vive en un único lugar (`service_public.go`).
//
// Aquí no hay lógica de negocio: las reglas (validación y normalización de
// números, límites, moneda por tipo, español por campo, tope del bloque) viven
// en el service y en `numero.go`.
package donaciones

import (
	"time"

	"github.com/google/uuid"
)

// PublicationState es el estado de publicación de un método o de un texto
// (FR-012/Q10): la publicación es por elemento y «se publica al guardar»
// (FR-013).
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

// MetodoKind es el tipo de método de donación (Q1/FR-019): `iban` = cuenta
// bancaria, `sinpe` = número de SINPE Móvil. No es contenido traducible.
type MetodoKind string

const (
	// KindIban es una cuenta bancaria (IBAN).
	KindIban MetodoKind = "iban"
	// KindSinpe es un número de SINPE Móvil.
	KindSinpe MetodoKind = "sinpe"
)

// Valid indica si el tipo pertenece al registro del contrato.
func (k MetodoKind) Valid() bool {
	return k == KindIban || k == KindSinpe
}

// Currency es la moneda de una cuenta IBAN (Q2/FR-019). La etiqueta visible la
// localiza el frontend: la moneda no es contenido traducible.
type Currency string

const (
	// CurrencyCRC es colones costarricenses.
	CurrencyCRC Currency = "CRC"
	// CurrencyUSD es dólares estadounidenses.
	CurrencyUSD Currency = "USD"
)

// Valid indica si la moneda pertenece al registro del contrato.
func (c Currency) Valid() bool {
	return c == CurrencyCRC || c == CurrencyUSD
}

// TextKey es la clave de uno de los dos textos libres de la sección
// (FR-021/Q5/Q11): `intro` («cómo donar») y `usage` («en qué se usan las
// donaciones»). Es el `text_key UNIQUE` de `donation_texts`.
type TextKey string

const (
	// TextKeyIntro es el texto introductorio («cómo donar»).
	TextKeyIntro TextKey = "intro"
	// TextKeyUsage es el texto de uso de las donaciones.
	TextKeyUsage TextKey = "usage"
)

// Valid indica si la clave pertenece al registro del contrato.
func (k TextKey) Valid() bool {
	return k == TextKeyIntro || k == TextKeyUsage
}

// --- Entidades ---

// MetodoDonacion es un método de donación (FR-001/FR-019): su tipo, el número
// **ya normalizado** (IBAN a los caracteres del número; SINPE `+506XXXXXXXX`),
// la entidad financiera **no traducible**, el titular opcional **traducible**
// (`HolderEs`/`HolderEn`), la moneda (solo en IBAN), el orden manual dentro del
// grupo de tipo y su estado.
type MetodoDonacion struct {
	ID               uuid.UUID
	Kind             MetodoKind
	Number           string
	Institution      string
	HolderEs         *string
	HolderEn         *string
	Currency         *Currency
	PublicationState PublicationState
	SortOrder        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TextoDonacion es uno de los dos textos libres de la sección (FR-021), con su
// clave, el cuerpo en ambos idiomas (el español obligatorio, el inglés
// opcional) y su **propio** estado de publicación (Q10).
type TextoDonacion struct {
	ID               uuid.UUID
	TextKey          TextKey
	BodyEs           string
	BodyEn           *string
	PublicationState PublicationState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// --- DTOs de entrada del contrato (etiquetas `validate`) ---
//
// Los campos `*En` son `string`: `""`/espacios significa "sin traducción" y el
// service lo persiste como NULL (normalización I6). `institution` **no** tiene
// par de idiomas (R7-5): no es contenido traducible.

// DonacionInput es el cuerpo de POST /api/v1/admin/donaciones (FR-014/FR-019):
// obligatorios **solo** `kind`, `number` e `institution` (y `currency` si
// `kind` es `iban`, regla condicional del service). `number` se acepta tal como
// lo escribe el equipo (IBAN con espacios o guiones; SINPE con o sin `+506`) y
// el service lo normaliza y valida (FR-015/FR-016).
type DonacionInput struct {
	Kind             string `json:"kind" validate:"required,oneof=iban sinpe"`
	Number           string `json:"number" validate:"required,min=1,max=60"`
	Institution      string `json:"institution" validate:"required,min=1,max=160"`
	HolderEs         string `json:"holderEs" validate:"omitempty,max=160"`
	HolderEn         string `json:"holderEn" validate:"omitempty,max=160"`
	Currency         string `json:"currency" validate:"omitempty,oneof=CRC USD"`
	PublicationState string `json:"publicationState" validate:"omitempty,oneof=draft published"`
	SortOrder        int    `json:"sortOrder"`
}

// DonacionPatch es el cuerpo de PATCH /api/v1/admin/donaciones/{id}: solo los
// campos presentes se modifican y un `null` explícito borra su valor opcional
// (p. ej. `holderEs`, `holderEn`, `currency`). Las reglas de coherencia se
// aplican **sobre el resultado ya fusionado** (R7-15): por eso aquí no hay
// etiquetas `validate` (la validación de dominio del service revalida el
// conjunto completo).
type DonacionPatch struct {
	Kind             Optional[string] `json:"kind"`
	Number           Optional[string] `json:"number"`
	Institution      Optional[string] `json:"institution"`
	HolderEs         Optional[string] `json:"holderEs"`
	HolderEn         Optional[string] `json:"holderEn"`
	Currency         Optional[string] `json:"currency"`
	PublicationState Optional[string] `json:"publicationState"`
	SortOrder        Optional[int]    `json:"sortOrder"`
}

// DonacionTextoInput es el cuerpo de PUT
// /api/v1/admin/donaciones/textos/{textKey} (FR-021): guardado completo del
// texto (crea o reemplaza el singleton), `textEs` obligatorio y `textEn`
// opcional (exige `textEs`, regla del service).
type DonacionTextoInput struct {
	TextEs           string `json:"textEs" validate:"required,min=1,max=2000"`
	TextEn           string `json:"textEn" validate:"omitempty,max=2000"`
	PublicationState string `json:"publicationState" validate:"omitempty,oneof=draft published"`
}

// --- DTOs públicos (salida) ---
//
// El DTO público **nunca** expone `publicationState` (FR-012) y su composición
// vive en un único lugar (`service_public.go`): `currency` solo aparece en las
// cuentas IBAN y `holder` solo cuando el titular existe (US1 esc. 5, sin líneas
// vacías ni marcadores).

// MetodoDonacionPublico es un método **publicado** para el visitante (FR-003).
// Es el mismo esquema que usa el bloque de la portada (FR-022): `kind`,
// `number` e `institution` siempre; `currency` solo en IBAN y `holder` solo
// cuando existe.
type MetodoDonacionPublico struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Number      string `json:"number"`
	Institution string `json:"institution"`
	Currency    string `json:"currency,omitempty"`
	Holder      string `json:"holder,omitempty"`
}

// TextoPublico es un texto libre **publicado** y ya resuelto a `lang`
// (FR-021): la clave entera se omite cuando el texto no está publicado.
type TextoPublico struct {
	Text string `json:"text"`
}

// OfrendasPublicas es el documento de la sección (FR-002…FR-006): `items`,
// `intro` y `usage` están presentes **solo** si hay algo publicado (SC-010: sin
// huecos); nunca son `null`. `items` está acotada por el service
// (`maxPublicItems = 200`).
type OfrendasPublicas struct {
	Lang  string                  `json:"lang"`
	Items []MetodoDonacionPublico `json:"items,omitempty"`
	Intro *TextoPublico           `json:"intro,omitempty"`
	Usage *TextoPublico           `json:"usage,omitempty"`
}

// PortadaOfrendas es el bloque de ofrendas de la portada (US6/FR-022): hasta
// `maxPortadaItems = 5` métodos publicados en el orden de la sección. `items`
// está presente **solo** si hay métodos publicados (sin ninguno el bloque no se
// dibuja) y nunca es `null`. Este bloque solo mira los métodos.
type PortadaOfrendas struct {
	Lang  string                  `json:"lang"`
	Items []MetodoDonacionPublico `json:"items,omitempty"`
}
