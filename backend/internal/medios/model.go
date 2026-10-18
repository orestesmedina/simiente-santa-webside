// Package medios es el dominio de la sección «Medios» (F9): las prédicas y los
// podcasts de la iglesia (FR-001). Una ficha por sermón (`Episodio`) con una o
// dos fuentes de importación (`FuenteEpisodio`, una por elemento de feed) y el
// registro de cada «Buscar episodios nuevos» (`SyncRun`).
//
// Sigue las capas handler → service → repository (§II): los handlers no tocan
// SQL y el service no conoce net/http ni pgx. La red saliente del backend (los
// dos feeds, los probadores de disponibilidad y la copia de carátulas) vive en
// este paquete detrás de las interfaces falsables `FeedSource` (feeds.go),
// `AvailabilityProber` (disponibilidad.go) y `CoverFetcher` (caratula.go), y
// sale **solo** por la fábrica única de cliente HTTP (cliente_http.go, P9-20).
//
// Este archivo reúne el modelo del dominio: los enums del contrato (registros
// cerrados), las entidades y los DTOs con sus etiquetas `validate`. Los
// **DTOs públicos** (`MedioPublico`, `MediosPublicos`, `PortadaMedios`) se
// componen en un único lugar —`service_public.go`, T916/T917— y nunca exponen
// `visibility`, marcas ni `sources` (FR-009/SC-002).
//
// Aquí no hay lógica de negocio: las reglas (validación, normalización, topes,
// unión por título y día, marcas y su resolución) viven en el service.
package medios

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Enums del contrato (registros cerrados)
// ---------------------------------------------------------------------------

// Visibilidad es el estado de una ficha (Q5/FR-015). F9 **no** usa
// `publicationState`: el contenido entra solo y no se borra (Consecuencia 5).
type Visibilidad string

const (
	// VisibilidadVisible: la ficha puede salir al visitante (si cumple el
	// predicado público de R9-7).
	VisibilidadVisible Visibilidad = "visible"
	// VisibilidadOculta: la ficha no sale al visitante. La decide el equipo y
	// la sincronización solo la cambia por indisponibilidad **confirmada**
	// (FR-017, excepción A3 de la fase 5).
	VisibilidadOculta Visibilidad = "hidden"
)

// Valid indica si el estado pertenece al registro del contrato.
func (v Visibilidad) Valid() bool {
	return v == VisibilidadVisible || v == VisibilidadOculta
}

// MarcaRevision es la clase de marca de revisión (FR-014/FR-017): identifica
// **qué** marca se resuelve o se filtra. Las marcas viven en **dos ejes
// independientes** (P9-19, corrección C1): `possible_duplicate` pertenece al eje
// de duplicados y `unavailable`/`missing_from_feed` al de disponibilidad; los
// dos pueden estar activos a la vez en la misma ficha.
type MarcaRevision string

const (
	// MarcaPosibleDuplicado: par dudoso **abierto**. Es el único bloqueo
	// público por marca (síntesis R9-7/precisión 1 de la puerta 2).
	MarcaPosibleDuplicado MarcaRevision = "possible_duplicate"
	// MarcaNoDisponible: la comprobación confirmó que el enlace ya no está
	// (404/410); la ficha quedó oculta (FR-017/Q16).
	MarcaNoDisponible MarcaRevision = "unavailable"
	// MarcaFueraDelFeed: la ficha salió de la ventana de su feed pero su enlace
	// sigue disponible —o no pudo comprobarse—: **sigue visible** (FR-017).
	MarcaFueraDelFeed MarcaRevision = "missing_from_feed"
)

// Valid indica si la clase pertenece al registro del contrato.
func (m MarcaRevision) Valid() bool {
	switch m {
	case MarcaPosibleDuplicado, MarcaNoDisponible, MarcaFueraDelFeed:
		return true
	default:
		return false
	}
}

// EnEjeDeDisponibilidad indica si la clase pertenece al eje de disponibilidad
// (false para la clase del eje de duplicados). El service lo usa para resolver
// la marca indicada sin confundir los ejes (P9-19/T919).
func (m MarcaRevision) EnEjeDeDisponibilidad() bool {
	return m == MarcaNoDisponible || m == MarcaFueraDelFeed
}

// MarcaDisponibilidad traduce la clase al eje de disponibilidad; el segundo
// valor es false cuando la clase pertenece al eje de duplicados (nunca se
// inventa una marca de disponibilidad).
func (m MarcaRevision) MarcaDisponibilidad() (MarcaDisponibilidad, bool) {
	if !m.EnEjeDeDisponibilidad() {
		return "", false
	}
	return MarcaDisponibilidad(m), true
}

// MarcaDisponibilidad es la marca del **eje de disponibilidad** que tiene la
// ficha (P9-19/FR-017). Cambiar de clase resetea su resolución y la marca nueva
// nace abierta; con la misma clase, una marca resuelta no se reabre.
type MarcaDisponibilidad string

const (
	// MarcaDisponibilidadNoDisponible: enlace confirmado muerto (oculta).
	MarcaDisponibilidadNoDisponible MarcaDisponibilidad = "unavailable"
	// MarcaDisponibilidadFueraDelFeed: salió de la ventana de su feed sin
	// confirmación de muerte (sigue visible).
	MarcaDisponibilidadFueraDelFeed MarcaDisponibilidad = "missing_from_feed"
)

// Valid indica si la marca pertenece al registro del contrato.
func (m MarcaDisponibilidad) Valid() bool {
	return m == MarcaDisponibilidadNoDisponible || m == MarcaDisponibilidadFueraDelFeed
}

// Proveedor es el origen de importación de un elemento de feed (FR-010/R9-5):
// los dos únicos orígenes admitidos (Q15).
type Proveedor string

const (
	// ProveedorYouTube: feed Atom del canal de YouTube.
	ProveedorYouTube Proveedor = "youtube"
	// ProveedorPodcast: RSS del podcast en Spotify for Creators/Anchor.
	ProveedorPodcast Proveedor = "podcast"
)

// Valid indica si el proveedor pertenece al registro del contrato.
func (p Proveedor) Valid() bool {
	return p == ProveedorYouTube || p == ProveedorPodcast
}

// OrigenMedio son los formatos que ofrece la ficha, derivados de sus fuentes
// (FR-003/Q18). No es contenido traducible (FR-026).
type OrigenMedio string

const (
	// OrigenVideo: solo YouTube.
	OrigenVideo OrigenMedio = "video"
	// OrigenAudio: solo podcast.
	OrigenAudio OrigenMedio = "audio"
	// OrigenAmbos: el sermón está en los dos lados (una sola ficha con los dos
	// enlaces).
	OrigenAmbos OrigenMedio = "both"
)

// Valid indica si el origen pertenece al registro del contrato.
func (o OrigenMedio) Valid() bool {
	return o == OrigenVideo || o == OrigenAudio || o == OrigenAmbos
}

// EstadoFeed es el estado de cada feed en una búsqueda (FR-016). `failed`
// significa que el feed no respondió y que **nada** se ocultó ni se marcó por
// esa causa.
type EstadoFeed string

const (
	// FeedOK: el feed se leyó completo.
	FeedOK EstadoFeed = "ok"
	// FeedFallido: el feed no respondió (el resumen lo informa, no lo esconde).
	FeedFallido EstadoFeed = "failed"
)

// Valid indica si el estado pertenece al registro del contrato.
func (e EstadoFeed) Valid() bool {
	return e == FeedOK || e == FeedFallido
}

// ---------------------------------------------------------------------------
// Entidades
// ---------------------------------------------------------------------------

// Episodio es una ficha de `media_items`: **una entrada por sermón** (FR-001)
// con los datos que trajo el feed (título, fecha, carátula copiada y la
// descripción corta recortada), el estado que decide el equipo, las marcas de
// los dos ejes con su resolución y el puntero de unión.
//
// `DescripcionEs`/`DescripcionEn` son los **únicos** campos traducibles
// (FR-025/FR-026) y `DescripcionEditada` protege la versión del equipo
// (FR-015/SC-007: la sincronización jamás la pisa). `VideoURL`/`AudioURL` no son
// columnas de la tabla: los resuelve el `LEFT JOIN` por proveedor de cada
// consulta (una ficha = 1 video y/o 1 audio).
type Episodio struct {
	ID            uuid.UUID
	Titulo        string
	PublicadoEn   time.Time
	DescripcionEs *string
	DescripcionEn *string
	// DescripcionEditada: true cuando la descripción corta es del equipo.
	DescripcionEditada bool
	// Caratula es el nombre del archivo copiado al sitio (`img_<uuid>.<ext>`,
	// generado por el servidor); nil → la ficha entra sin carátula y el cliente
	// pinta el marcador neutro (default revisable #4).
	Caratula    *string
	Origen      OrigenMedio
	Visibilidad Visibilidad

	// Eje de duplicados (FR-014/P9-19): `DuplicadoDeID` presente = hay par
	// (cada miembro apunta al otro; la pareja recíproca la escribe el service) y
	// `MarcaDuplicadoResueltaEn` presente = «posible duplicado» resuelto.
	DuplicadoDeID            *uuid.UUID
	MarcaDuplicadoResueltaEn *time.Time

	// Eje de disponibilidad (FR-017/P9-19): la marca, su resolución y el dato
	// que la motivó (resultado agregado: el peor de sus fuentes comprobadas).
	DisponibilidadComprobadaEn    *time.Time
	ResultadoDisponibilidad       *ResultadoDisponibilidad
	MarcaDisponibilidad           *MarcaDisponibilidad
	MarcaDisponibilidadResueltaEn *time.Time

	// UnidoAID: la ficha que la absorbió en una unión (R9-12). Presente = la
	// ficha está absorbida y desaparece de todas las consultas (sin DELETE).
	UnidoAID *uuid.UUID

	// Enlaces resueltos por el LEFT JOIN por proveedor ("" si no hay).
	VideoURL string
	AudioURL string

	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// FuenteEpisodio es una fila de `media_episode_sources`: **un elemento de feed
// importado** (R9-11) con su procedencia. `Proveedor`+`IDExterno` es la llave de
// idempotencia de la búsqueda (`UNIQUE` en la BD) y `TituloFeed` conserva el
// título de **ese** feed (al unir, el del podcast se conserva como dato de
// origen: default revisable #7).
type FuenteEpisodio struct {
	ID            uuid.UUID
	EpisodioID    uuid.UUID
	Proveedor     Proveedor
	IDExterno     string
	TituloFeed    string
	Enlace        string
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// SyncRun es una fila de `media_sync_runs`: el registro de cada «Buscar
// episodios nuevos» (FR-018) con su resumen numérico y el estado de cada feed.
// `Marcados` cuenta **solo** las marcas «posible duplicado» nuevas de esa
// búsqueda (criterio B2 de la fase 5, literal de FR-018).
type SyncRun struct {
	ID          uuid.UUID
	ActorUserID *uuid.UUID
	IniciadoEn  time.Time
	TerminadoEn time.Time

	Importados  int
	Unidos      int
	Marcados    int
	Ocultos     int
	Comprobados int
	Inconclusos int

	EstadoYouTube EstadoFeed
	EstadoPodcast EstadoFeed

	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// ---------------------------------------------------------------------------
// DTOs de entrada del contrato (etiquetas `validate`)
// ---------------------------------------------------------------------------

// MedioPatch es el cuerpo de `PATCH /api/v1/admin/medios/{id}` (FR-019/FR-020):
// **solo** la descripción corta en sus dos idiomas y la visibilidad. Cualquier
// otra clave la rechaza el handler (contrato `additionalProperties: false`).
//
// Un campo ausente (nil) no toca nada; `descriptionEn` a `NULL` borra la
// traducción opcional y `descriptionEs` a `NULL` deja la ficha sin descripción
// (hueco documentado: la ficha omite la línea, SC-010). La coherencia «español
// por campo» (FR-023) la aplica el service sobre el resultado ya fusionado.
type MedioPatch struct {
	DescriptionEs *string `json:"descriptionEs" validate:"omitempty,min=1,max=300"`
	DescriptionEn *string `json:"descriptionEn" validate:"omitempty,max=300"`
	Visibility    *string `json:"visibility" validate:"omitempty,oneof=visible hidden"`
}

// UnirInput es el cuerpo de `POST /api/v1/admin/medios/{id}/unir`
// (FR-014/FR-019): `{id}` es la ficha que se conserva y `otherId` la absorbida.
// El formato UUID y la regla `otherId != id` los comprueba el service.
type UnirInput struct {
	OtherID string `json:"otherId" validate:"required"`
}

// ResolverMarcaInput es el cuerpo de
// `POST /api/v1/admin/medios/{id}/marcas/resolver` (FR-017/US3 esc. 6): `mark`
// indica **qué** marca se resuelve (los dos ejes pueden estar activos a la vez)
// y `visibility` decide si la ficha vuelve a estar visible.
type ResolverMarcaInput struct {
	Mark       string `json:"mark" validate:"required,oneof=possible_duplicate unavailable missing_from_feed"`
	Visibility string `json:"visibility" validate:"required,oneof=visible hidden"`
}

// ---------------------------------------------------------------------------
// DTOs de salida del panel (FR-019)
// ---------------------------------------------------------------------------

// FuenteAdmin es la procedencia de un enlace de la ficha (R9-11): una fila por
// elemento de feed importado.
type FuenteAdmin struct {
	Proveedor  string `json:"provider"`
	IDExterno  string `json:"externalId"`
	TituloFeed string `json:"feedTitle"`
	Enlace     string `json:"url"`
}

// MedioAdmin es la ficha completa del panel (FR-019): datos del feed **no
// editables**, la descripción corta en ambos idiomas con su marca de edición, la
// visibilidad, el origen (formato derivado de las fuentes más la fecha de
// importación), las **marcas por eje** —que solo viajan en este DTO—, la
// comprobación de disponibilidad y las `sources` que el equipo revisa antes de
// unir. Las claves opcionales se omiten cuando no aplican (sin `null`).
type MedioAdmin struct {
	ID                 string    `json:"id"`
	Titulo             string    `json:"title"`
	DescripcionEs      *string   `json:"descriptionEs,omitempty"`
	DescripcionEn      *string   `json:"descriptionEn,omitempty"`
	DescripcionEditada bool      `json:"descriptionEdited"`
	PublicadoEn        time.Time `json:"publishedAt"`
	CaratulaURL        string    `json:"coverUrl,omitempty"`
	VideoURL           string    `json:"videoUrl,omitempty"`
	AudioURL           string    `json:"audioUrl,omitempty"`
	Origen             string    `json:"origin"`
	Visibilidad        string    `json:"visibility"`

	DuplicadoDeID            string     `json:"duplicateOfId,omitempty"`
	MarcaDuplicadoResueltaEn *time.Time `json:"duplicateMarkResolvedAt,omitempty"`
	UnidoAID                 string     `json:"mergedIntoId,omitempty"`

	DisponibilidadComprobadaEn    *time.Time `json:"availabilityCheckedAt,omitempty"`
	ResultadoDisponibilidad       string     `json:"availabilityResult,omitempty"`
	MarcaDisponibilidad           string     `json:"availabilityMark,omitempty"`
	MarcaDisponibilidadResueltaEn *time.Time `json:"availabilityMarkResolvedAt,omitempty"`

	Fuentes       []FuenteAdmin `json:"sources"`
	CreadoEn      time.Time     `json:"createdAt"`
	ActualizadoEn time.Time     `json:"updatedAt"`
}

// MedioAdminList es el listado paginado del módulo (§8.1.2/§8.1.3). `lastSync`
// solo existe tras la primera «Buscar episodios nuevos» (FR-018).
type MedioAdminList struct {
	Items    []MedioAdmin  `json:"items"`
	Total    int64         `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
	LastSync *SyncRunAdmin `json:"lastSync,omitempty"`
}

// ResumenBusqueda es el resumen de una «Buscar episodios nuevos» (FR-018): sin
// información interna del sistema (nunca mensajes técnicos ni de red).
type ResumenBusqueda struct {
	IniciadoEn    time.Time `json:"startedAt"`
	TerminadoEn   time.Time `json:"finishedAt"`
	Importados    int       `json:"importedCount"`
	Unidos        int       `json:"joinedCount"`
	Marcados      int       `json:"markedCount"`
	Ocultos       int       `json:"hiddenCount"`
	Comprobados   int       `json:"availabilityCheckedCount"`
	Inconclusos   int       `json:"availabilityInconclusiveCount"`
	EstadoYouTube string    `json:"youtube"`
	EstadoPodcast string    `json:"podcast"`
}

// SyncRunAdmin es el registro persistido de una búsqueda (FR-018): el mismo
// resumen con su identificador y el actor. Se expone como `lastSync`.
type SyncRunAdmin struct {
	ID            string    `json:"id"`
	ActorUserID   string    `json:"actorUserId,omitempty"`
	IniciadoEn    time.Time `json:"startedAt"`
	TerminadoEn   time.Time `json:"finishedAt"`
	Importados    int       `json:"importedCount"`
	Unidos        int       `json:"joinedCount"`
	Marcados      int       `json:"markedCount"`
	Ocultos       int       `json:"hiddenCount"`
	Comprobados   int       `json:"availabilityCheckedCount"`
	Inconclusos   int       `json:"availabilityInconclusiveCount"`
	EstadoYouTube string    `json:"youtube"`
	EstadoPodcast string    `json:"podcast"`
}
