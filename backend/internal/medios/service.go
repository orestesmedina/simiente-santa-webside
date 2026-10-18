// Package medios es el dominio de la sección «Medios» (F9): las prédicas y los
// podcasts de la iglesia.
//
// Este archivo reúne la base del service (T915): el puerto `Repository` que el
// service declara para sí (arq. R3), las dependencias que recibe por constructor
// —incluidas las tres interfaces de la red saliente, que **se declaran en su
// propio archivo** (`FeedSource` en feeds.go, `AvailabilityProber` en
// disponibilidad.go y `CoverFetcher` en caratula.go: T915 no las redeclara, M1)—,
// las **constantes revisables** del módulo (P9-18/R9-9: el único punto de verdad
// de los topes), la **normalización** (trim; `""`/espacios → NULL, patrón I6) y la
// **validación de dominio** (FR-023).
//
// No conoce net/http ni pgx: solo reglas de negocio. La persistencia y la
// auditoría atómica viven en el repository; aquí se decide QUÉ datos guardar y
// QUÉ filas de auditoría acompañan cada mutación. Los archivos reparten las
// responsabilidades: `service_public.go` (sección y bloque de la portada),
// `service_admin.go` (panel: listado, edición, unión y resolución de marcas),
// `service_sync.go` (la búsqueda a demanda) y `service_audit.go` (denegaciones y
// rechazos).
package medios

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/validate"
)

// --- Constantes del dominio (P9-18/R9-9: constantes revisables) ---
//
// Las dos URLs de los feeds (FR-010) y los timeouts por petición saliente
// (10/10/5 s) viven con su consumidor —`urlFeedYouTube`/`urlFeedPodcast` y
// `tiempoEsperaFeeds` en feeds.go, `tiempoEsperaProbador` en disponibilidad.go,
// `tiempoEsperaCaratula` y `maxCaratulasPorBusqueda` en caratula.go— porque son
// la política de la red saliente, junto a la fábrica única de cliente HTTP
// (T907A/P9-20). Aquí viven los topes de negocio que usan los servicios.
const (
	// pageDefault es el tamaño de la tanda de la sección pública (default
	// revisable #5/R9-10: 12 por tanda, «Cargar más»).
	pageDefault = 12
	// maxPage es el tope duro de la sección (R9-10: la respuesta no crece con
	// el catálogo ni con un parámetro del visitante).
	maxPage = 50
	// maxPortada es el tope FIRME del bloque de la portada (FR-027/SC-015): lo
	// aplica el servidor, no crece con el contenido y el contrato no expone
	// ningún parámetro de límite.
	maxPortada = 5

	// maxTitulo y maxDescripcion son los límites de las columnas de
	// `media_items` (migración 000014/FR-020): el nombre del feed y la
	// descripción corta.
	maxTitulo      = 300
	maxDescripcion = 300

	// maxElementosPorBusqueda es el tope duro de elementos importados por
	// búsqueda (P9-3: 30).
	maxElementosPorBusqueda = 30
	// maxComprobacionesPorBusqueda es el tope duro de comprobaciones de
	// disponibilidad por búsqueda (P9-3: 30).
	maxComprobacionesPorBusqueda = 30
	// maxEnVuelo es el máximo de descargas y comprobaciones simultáneas
	// (P9-3: 4).
	maxEnVuelo = 4
	// presupuestoBusqueda es el presupuesto total de una «Buscar episodios
	// nuevos» (P9-3/R9-9: 90 s, constante revisable). Al agotarse la búsqueda
	// cierra con lo hecho y el resumen lleva conteos parciales.
	presupuestoBusqueda = 90 * time.Second

	// maxTargetLabel es la longitud que admite `admin_actions.target_label`
	// (migración 000004).
	maxTargetLabel = 254
)

// Idiomas admitidos por la API pública (FR-025).
const (
	// LangES es el idioma base (obligatorio).
	LangES = "es"
	// LangEN es el idioma opcional con fallback a `es` (solo la descripción).
	LangEN = "en"
)

// --- Puerto del repositorio (definido por quien lo consume, arq. R3) ---

// Repository es el puerto de datos que el service de medios necesita: las
// consultas del dominio `medios` (T912/T913) con sus errores ya traducidos a
// apperr. Las mutaciones reciben las filas de auditoría que deben insertarse en
// la MISMA transacción (variádicas): o el cambio y su registro, o nada (FR-022).
type Repository interface {
	// Vía pública (FR-002/FR-009/FR-027): el predicado público completo vive en
	// la consulta, no en el service (SC-002).
	ListPublishedMediaItems(ctx context.Context, params paginate.Params) ([]Episodio, error)
	CountPublishedMediaItems(ctx context.Context) (int64, error)
	ListPortadaMediaItems(ctx context.Context, limit int) ([]Episodio, error)

	// Panel (FR-019): todo menos las absorbidas (salvo `IncluirAbsorbidas`),
	// con el mismo orden que ve el visitante.
	ListMediaItems(ctx context.Context, filtro FiltroPanel, params paginate.Params) ([]Episodio, error)
	CountMediaItems(ctx context.Context, filtro FiltroPanel) (int64, error)
	GetMediaItemByID(ctx context.Context, id uuid.UUID) (Episodio, bool, error)

	// Mutaciones de fichas (FR-012/FR-015/FR-019/FR-022).
	InsertMediaItem(ctx context.Context, episodio Episodio, actions ...audit.Action) (Episodio, error)
	UpdateMediaItem(ctx context.Context, episodio Episodio, actions ...audit.Action) (Episodio, error)
	// RepointMediaItemSources aplica la unión por puntero en UNA transacción:
	// repunta las fuentes de la absorbida, escribe las dos filas y registra la
	// auditoría (P9-10/FR-019).
	RepointMediaItemSources(ctx context.Context, resultante, absorbida Episodio, actions ...audit.Action) ([]FuenteEpisodio, error)

	// Fuentes de importación (R9-11): `ON CONFLICT (provider, external_id) DO
	// NOTHING` = idempotencia de la búsqueda, sin refrescar la instantánea.
	InsertMediaEpisodeSource(ctx context.Context, fuente FuenteEpisodio) (FuenteEpisodio, bool, error)
	ListMediaEpisodeSourcesByProvider(ctx context.Context, proveedor Proveedor) ([]FuenteEpisodio, error)
	GetMediaEpisodeSource(ctx context.Context, proveedor Proveedor, idExterno string) (FuenteEpisodio, bool, error)
	ListMediaItemSourcesByItem(ctx context.Context, itemID uuid.UUID) ([]FuenteEpisodio, error)

	// Registro de búsquedas (FR-018).
	InsertMediaSyncRun(ctx context.Context, run SyncRun, actions ...audit.Action) (SyncRun, error)
	GetLastMediaSyncRun(ctx context.Context) (SyncRun, bool, error)

	// RecordAction persiste una acción de auditoría best-effort (denegaciones y
	// rechazos): fuera de transacción y sin cambiar la respuesta (R7-10).
	RecordAction(ctx context.Context, action audit.Action) error
}

// ServiceDeps agrupa las dependencias del service.
type ServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository Repository
	// FeedSource lee los dos feeds del cliente (FR-010). Puede ser nil en los
	// procesos que no buscan episodios.
	FeedSource FeedSource
	// AvailabilityProber comprueba si un enlace sigue disponible (FR-017).
	AvailabilityProber AvailabilityProber
	// CoverFetcher copia al sitio la carátula de un elemento importado
	// (P9-7/SC-005).
	CoverFetcher CoverFetcher
	// Store es el almacén de archivos: lo usa la unión para borrar en mejor
	// esfuerzo la carátula que quedó sin uso (FR-019).
	Store storage.Store
	// Location es la zona horaria del lugar (APP_TZ, decisión 17 de F4): decide
	// el «mismo día» de la unión automática (FR-013) y el corte de la ventana del
	// feed (FR-017). nil degrada a UTC (la configuración siempre la aporta).
	Location *time.Location
	// Logger registra los fallos best-effort; nil usa el logger por defecto.
	Logger *slog.Logger
}

// service implementa la cara pública, la gestión del panel, la búsqueda a
// demanda y el registro de auditoría del dominio. Una sola instancia satisface
// las cuatro responsabilidades; los archivos service_*.go reparten los métodos.
type service struct {
	repository  Repository
	feeds       FeedSource
	probador    AvailabilityProber
	caratulas   CoverFetcher
	store       storage.Store
	zonaHoraria *time.Location
	logger      *slog.Logger
}

// NewService construye el service de medios. Si Logger es nil se usa el logger
// por defecto.
func NewService(deps ServiceDeps) *service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &service{
		repository:  deps.Repository,
		feeds:       deps.FeedSource,
		probador:    deps.AvailabilityProber,
		caratulas:   deps.CoverFetcher,
		store:       deps.Store,
		zonaHoraria: deps.Location,
		logger:      logger,
	}
}

// zona devuelve la zona horaria del lugar (`APP_TZ`, decisión 17 de F4): la llave
// del «mismo día» de la unión automática (FR-013). Sin configuración degrada a
// UTC —el corte de día deja de ser el de la iglesia—, así que `cmd/api` siempre
// pasa la zona que cargó `platform/config`.
func (s *service) zona() *time.Location {
	if s.zonaHoraria == nil {
		return time.UTC
	}
	return s.zonaHoraria
}

// --- Normalización (analyze I6) ---

// normalizarOpcional normaliza un campo opcional: `""` o solo espacios pasa a
// nil (analyze I6), nunca a la cadena vacía que los CHECK de longitud rechazan.
// El contenido (tildes, ñ, ¿¡, emojis) se conserva intacto.
func normalizarOpcional(valor string) *string {
	normalizado := colapsarEspacios(valor)
	if normalizado == "" {
		return nil
	}
	return &normalizado
}

// validLang indica si `lang` pertenece al registro (es|en).
func validLang(lang string) bool {
	return lang == LangES || lang == LangEN
}

// normalizeLang valida y normaliza el idioma pedido por el visitante: vacío es
// el idioma base y cualquier otro valor es `400 invalid` con `details.lang`
// (FR-025). Es la única implementación: la usan la sección y el bloque (T916/T917).
func normalizeLang(lang string) (string, error) {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return LangES, nil
	}
	if !validLang(lang) {
		return "", apperr.Invalid(
			"El idioma solicitado no es válido",
			apperr.WithDetails(map[string]any{"lang": "solo se admite es o en"}),
		)
	}
	return lang, nil
}

// normalizarPaginacionPublica normaliza la tanda de la sección (R9-10): sin
// indicar usa `pageDefault` (12), un valor por encima del tope se acota a
// `maxPage` (50) —el tope nunca es un error— y un valor negativo es
// `400 invalid` con su campo en `details` (§8.1.3).
func normalizarPaginacionPublica(limit, offset int) (paginate.Params, error) {
	switch {
	case limit < 0:
		return paginate.Params{}, apperr.Invalid(
			"Parámetros de paginación inválidos",
			apperr.WithDetails(map[string]any{"limit": "debe ser mayor o igual a 0"}),
		)
	case offset < 0:
		return paginate.Params{}, apperr.Invalid(
			"Parámetros de paginación inválidos",
			apperr.WithDetails(map[string]any{"offset": "debe ser mayor o igual a 0"}),
		)
	}
	if limit == 0 {
		limit = pageDefault
	}
	if limit > maxPage {
		limit = maxPage
	}
	return paginate.Params{Limit: limit, Offset: offset}, nil
}

// --- Validación de dominio (FR-019/FR-023) ---

// patchMedio es el `PATCH` ya normalizado: los campos que el cuerpo toca (los
// punteros del DTO distinguen «ausente» de «valor») con sus textos recortados.
// `tocaEs` con `descripcionEs == nil` es «sin descripción» (decisión del
// 2026-10-18: `""`/espacios vacían la descripción, igual que el inglés) y
// `tocaEn` con `descripcionEn == nil` es «borra la traducción opcional» (patrón
// I6: `""`/espacios → NULL).
type patchMedio struct {
	descripcionEs *string
	descripcionEn *string
	tocaEs        bool
	tocaEn        bool
	visibilidad   *Visibilidad
}

// patchMedioFromInput normaliza y valida el único cuerpo editable del módulo
// (FR-019/FR-020): `descriptionEs` opcional de hasta 300 caracteres —`""` o solo
// espacios la vacían: «sin descripción», con el mismo tratamiento que el inglés
// (patrón I6, decisión del 2026-10-18)—, `descriptionEn` opcional (1…300, `""`
// borra la traducción) y `visibility` dentro del catálogo. Devuelve un
// `400 invalid` con **un motivo por campo** en `details` (las mismas claves del
// contrato y de los `CHECK` de la migración 000014: la BD es la última línea de
// defensa).
//
// La coherencia «español por campo» (FR-023) se comprueba sobre la ficha
// FUSIONADA (`validarEpisodio`), porque un `PATCH` que solo trae el inglés debe
// fallar si la ficha no tiene español. No se usa `validate.Struct`: sus reglas
// solo miran campos de texto y enteros, y los del `PATCH` son punteros (aquí la
// validación de forma y la de negocio son la misma y viven juntas).
func patchMedioFromInput(patch MedioPatch) (patchMedio, error) {
	details := make(map[string]any)
	out := patchMedio{}

	if patch.DescriptionEs != nil {
		out.tocaEs = true
		// `""`/espacios significa «sin descripción» (se guarda como NULL): la
		// ficha omite la línea en la superficie pública (SC-010), igual que la
		// versión en inglés (patrón I6).
		out.descripcionEs = normalizarOpcional(*patch.DescriptionEs)
		if out.descripcionEs != nil && len([]rune(*out.descripcionEs)) > maxDescripcion {
			details["descriptionEs"] = tooLongMessage(maxDescripcion)
		}
	}

	if patch.DescriptionEn != nil {
		out.tocaEn = true
		// `""`/espacios significa «sin traducción» (se guarda como NULL): el
		// fallback `en → es` del servidor nunca publica un campo vacío.
		out.descripcionEn = normalizarOpcional(*patch.DescriptionEn)
		if out.descripcionEn != nil && len([]rune(*out.descripcionEn)) > maxDescripcion {
			details["descriptionEn"] = tooLongMessage(maxDescripcion)
		}
	}

	if patch.Visibility != nil {
		visibilidad := Visibilidad(strings.TrimSpace(*patch.Visibility))
		if !visibilidad.Valid() {
			details["visibility"] = "solo se admite visible u oculto"
		} else {
			out.visibilidad = &visibilidad
		}
	}

	if len(details) == 0 {
		return out, nil
	}
	return patchMedio{}, apperr.Invalid(validate.MessageInvalid, apperr.WithDetails(details))
}

// validarEpisodio aplica las invariantes del dominio sobre la ficha ya
// normalizada (creada o **fusionada**: un `PATCH` y una unión exigen revalidar
// el resultado). Devuelve un `400 invalid` con **un motivo por campo** en
// `details`, con las mismas claves que el contrato y que los `CHECK` de la
// migración 000014 (la BD es la última línea de defensa de estos invariantes).
//
// Reglas: nombre 1…300 (el del feed, no editable); descripción corta 1…300 por
// idioma con **español por campo** (FR-025: `descriptionEn` exige
// `descriptionEs`); `visibility` y `origin` dentro de sus catálogos. El vacío
// nunca llega aquí: la normalización del `PATCH` y de la importación lo
// convierten en «sin descripción» (NULL) antes de persistir, y un `NULL` es
// legal (decisión del 2026-10-18/SC-010).
func validarEpisodio(ep Episodio) error {
	details := make(map[string]any)

	if titulo := strings.TrimSpace(ep.Titulo); titulo == "" {
		details["title"] = "Este campo es obligatorio."
	} else if len([]rune(titulo)) > maxTitulo {
		details["title"] = tooLongMessage(maxTitulo)
	}

	if ep.DescripcionEs != nil {
		texto := *ep.DescripcionEs
		switch {
		case strings.TrimSpace(texto) == "":
			details["descriptionEs"] = "Este campo no puede quedar vacío."
		case len([]rune(texto)) > maxDescripcion:
			details["descriptionEs"] = tooLongMessage(maxDescripcion)
		}
	}
	if ep.DescripcionEn != nil {
		texto := *ep.DescripcionEn
		switch {
		case strings.TrimSpace(texto) == "":
			details["descriptionEn"] = "Este campo no puede quedar vacío."
		case len([]rune(texto)) > maxDescripcion:
			details["descriptionEn"] = tooLongMessage(maxDescripcion)
		}
	}
	if ep.DescripcionEn != nil && ep.DescripcionEs == nil {
		details["descriptionEs"] = "La versión en inglés exige el texto en español."
	}

	if !ep.Visibilidad.Valid() {
		details["visibility"] = "solo se admite visible u oculto"
	}
	if !ep.Origen.Valid() {
		details["origin"] = "el origen debe ser video, audio o both"
	}

	if len(details) == 0 {
		return nil
	}
	return apperr.Invalid(validate.MessageInvalid, apperr.WithDetails(details))
}

// validarUnion comprueba la regla del contrato «`otherId` distinto de `{id}`»
// (FR-019): unir una ficha consigo misma es `400 invalid` con `details.otherId`.
func validarUnion(id, otherID uuid.UUID) error {
	if id == otherID {
		return apperr.Invalid(
			validate.MessageInvalid,
			apperr.WithDetails(map[string]any{"otherId": "no puedes unir un episodio consigo mismo"}),
		)
	}
	return nil
}

// validarMarca comprueba que la marca que se resuelve pertenece al registro del
// contrato (FR-014/FR-017; P9-19: la marca indica el eje que se resuelve).
func validarMarca(marca MarcaRevision) error {
	if !marca.Valid() {
		return apperr.Invalid(
			validate.MessageInvalid,
			apperr.WithDetails(map[string]any{
				"mark": "solo se admite possible_duplicate, unavailable o missing_from_feed",
			}),
		)
	}
	return nil
}

// validarVisibilidad comprueba que la visibilidad elegida pertenece al registro
// del contrato (FR-015).
func validarVisibilidad(visibilidad Visibilidad) error {
	if !visibilidad.Valid() {
		return apperr.Invalid(
			validate.MessageInvalid,
			apperr.WithDetails(map[string]any{"visibility": "solo se admite visible u oculto"}),
		)
	}
	return nil
}

// tooLongMessage es la redacción de platform/validate para un texto que supera
// su límite (una sola forma de decirlo en toda la aplicación).
func tooLongMessage(max int) string {
	return fmt.Sprintf("No puede tener más de %d caracteres.", max)
}

// RecordAction persiste una acción de auditoría fuera de la transacción de una
// mutación (denegaciones y rechazos best-effort). El repository es el único
// punto que escribe `admin_actions` (R7-10).
func (r *repository) RecordAction(ctx context.Context, action audit.Action) error {
	return r.insertAdminAction(ctx, action)
}

// compile-time check: *repository must satisfy the port the service declares.
var _ Repository = (*repository)(nil)
