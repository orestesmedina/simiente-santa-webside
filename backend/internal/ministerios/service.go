package ministerios

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/validate"
)

// service.go reúne la base del service del dominio ministerios (T612): el
// puerto `Repository` que declara para sí (arq. R3), las constantes de negocio
// revisables —único punto de verdad de los límites—, la normalización de textos
// (trim; `""`/espacios → NULL, analyze I6), la validación de dominio completa
// (FR-014/FR-015/FR-020/FR-025) y los helpers de auditoría del módulo
// (`contentAction`/`contentLabel`, R6-12).
//
// No conoce net/http ni pgx: solo reglas de negocio. La persistencia y la
// auditoría atómica viven en el repository; aquí se decide QUÉ datos guardar y
// QUÉ filas de auditoría acompañan cada mutación. Los archivos reparten las
// responsabilidades: `service_public.go` (sección y bloque de la portada),
// `service_admin.go` (panel) y `service_audit.go` (denegaciones y rechazos).

// --- Constantes de negocio (único punto de verdad de los límites, FR-014) ---

const (
	// MaxNameLength es el límite del nombre del ministerio por idioma.
	MaxNameLength = 160
	// MaxDescriptionLength es el límite de la descripción por idioma (Q6).
	MaxDescriptionLength = 2000
	// MaxLeaderNameLength es el límite del nombre del encargado por idioma (Q1).
	MaxLeaderNameLength = 160
	// MinLeaderPhoneLength y MaxLeaderPhoneLength acotan el celular del encargado
	// (mismo `CHECK` que `home_contact`; R6-4).
	MinLeaderPhoneLength = 7
	// MaxLeaderPhoneLength es el tope del celular del encargado.
	MaxLeaderPhoneLength = 32
	// MinLeaderEmailLength y MaxLeaderEmailLength acotan el correo del encargado.
	MinLeaderEmailLength = 3
	// MaxLeaderEmailLength es el tope del correo del encargado.
	MaxLeaderEmailLength = 254
	// MaxScheduleLength es el límite del horario como texto libre por idioma
	// (Q6/FR-024: no es una fecha ni una recurrencia).
	MaxScheduleLength = 200
	// MaxPlaceLength es el límite del lugar como texto libre por idioma.
	MaxPlaceLength = 200
	// MaxURLLength acota el enlace al grupo (coincide con la etiqueta `url` de
	// platform/validate y con el contrato; FR-015).
	MaxURLLength = 500
	// MaxImageAltLength es el límite del texto alternativo por idioma (FR-020).
	MaxImageAltLength = 300
	// MediaPathPrefix es el prefijo público de las imágenes (contrato:
	// `/api/v1/media/<fileName>`).
	MediaPathPrefix = "/api/v1/media/"
	// maxPublicItems es el tope de la sección pública (RG6-12): es un límite de
	// seguridad del servidor, lo aplica GetMinisterios (T613) y lo documenta el
	// contrato. Si el volumen real lo superara, se sirven los primeros ya
	// ordenados.
	maxPublicItems = 200
	// maxPortadaItems es el tope FIRME del bloque de la portada (FR-027): lo
	// aplica GetPortadaMinisterios (T614), no crece con el contenido y el
	// contrato no expone ningún parámetro de límite.
	maxPortadaItems = 5
	// maxTargetLabel es la longitud que admite `admin_actions.target_label`
	// (migración 000004). Los límites del módulo (nombre ≤160 y archivo generado
	// por el servidor) garantizan que ninguna etiqueta lo alcanza: por eso no se
	// recorta nada.
	maxTargetLabel = 254
)

// Idiomas admitidos por la API pública (FR-008).
const (
	// LangES es el idioma base (obligatorio).
	LangES = "es"
	// LangEN es el idioma opcional con fallback a `es`.
	LangEN = "en"
)

// moduleLabel es la etiqueta legible del módulo en `targetLabel` (R6-12):
// `Ministerios · <nombre>` y `Ministerios · Imagen · <fileName>`.
const moduleLabel = "Ministerios"

// --- Puerto del repositorio (definido por quien lo consume, arq. R3) ---

// Repository es el puerto de datos que el service de ministerios necesita. Los
// errores de PostgreSQL llegan ya traducidos a apperr. Las mutaciones reciben
// las filas de auditoría que deben insertarse en la MISMA transacción
// (variádicas): o el cambio y su registro, o nada (FR-017/R6-12).
type Repository interface {
	// ListPublishedMinistries es la ÚNICA vía pública (sección y bloque):
	// ministerios publicados en orden `sort_order, id` con la ficha del encargado
	// enmascarada si no está publicada (FR-004/FR-012/FR-026).
	ListPublishedMinistries(ctx context.Context, limit int) ([]Ministry, error)

	// Gestión del panel (FR-010/FR-014/FR-016).
	InsertMinistry(ctx context.Context, ministry Ministry, actions ...audit.Action) (Ministry, error)
	UpdateMinistry(ctx context.Context, ministry Ministry, actions ...audit.Action) (Ministry, error)
	DeleteMinistry(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	GetMinistryByID(ctx context.Context, id uuid.UUID) (MinistryListItem, bool, error)
	ListMinistries(ctx context.Context, state *PublicationState, params paginate.Params) ([]MinistryListItem, error)
	CountMinistries(ctx context.Context, state *PublicationState) (int64, error)
	ExistsMinistryDuplicate(ctx context.Context, nameEs, descriptionEs string, leaderNameEs, whatsappURL *string, excludeID *uuid.UUID) (bool, error)

	// RecordAction persiste una acción de auditoría best-effort (denegaciones y
	// rechazos): fuera de transacción y sin cambiar la respuesta (R6-12).
	RecordAction(ctx context.Context, action audit.Action) error
}

// ServiceDeps agrupa las dependencias del service.
type ServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository Repository
	// Store es el almacén de imágenes (platform/storage). Puede ser nil en
	// pruebas que no ejerciten la subida ni la limpieza de archivos.
	Store storage.Store
	// Logger registra los fallos best-effort; nil usa el logger por defecto.
	Logger *slog.Logger
}

// service implementa la cara pública, la gestión del panel y el registro de
// auditoría del dominio. Una sola instancia satisface las tres
// responsabilidades; los archivos service_public.go, service_admin.go y
// service_audit.go reparten los métodos.
type service struct {
	repository Repository
	store      storage.Store
	logger     *slog.Logger
}

// NewService construye el service de ministerios. Si Logger es nil se usa el
// logger por defecto.
func NewService(deps ServiceDeps) *service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &service{repository: deps.Repository, store: deps.Store, logger: logger}
}

// --- Normalización (analyze I6) ---

// horizontalSpace colapsa las secuencias de espacios y tabulaciones internas a
// un único espacio, sin tocar los saltos de línea (los párrafos de la
// descripción).
var horizontalSpace = regexp.MustCompile(`[ \t]+`)

// normalizeText aplica trim y colapsa los espacios internos. Conserva los
// saltos de línea y todo el contenido (tildes, ñ, ¿¡ y emojis intactos).
func normalizeText(value string) string {
	value = strings.TrimSpace(value)
	return horizontalSpace.ReplaceAllString(value, " ")
}

// normalizeOptional normaliza un campo opcional: `""` o solo espacios pasa a nil
// (analyze I6), nunca a la cadena vacía que los CHECK de longitud rechazan.
func normalizeOptional(value string) *string {
	normalized := normalizeText(value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

// derefString devuelve el valor de un opcional, o "" si no existe.
func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// publicationStateOr traduce el estado de entrada; un valor vacío es `draft`
// (el default del contrato y de la BD).
func publicationStateOr(value string) PublicationState {
	if value == "" {
		return StateDraft
	}
	return PublicationState(value)
}

// validLang indica si `lang` pertenece al registro (es|en).
func validLang(lang string) bool {
	return lang == LangES || lang == LangEN
}

// normalizeLang valida y normaliza el idioma pedido por el visitante: vacío es
// el idioma base y cualquier otro valor es `400 invalid` con details.lang
// (FR-008). Es la única implementación: la usan la sección (T613) y el bloque de
// la portada (T614).
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

// --- Reglas de forma reutilizadas de platform/validate ---
//
// Son envoltorios de las etiquetas que ya existen (F2/F3): el service las aplica
// sobre los valores ya normalizados de la entidad fusionada, de modo que la
// redacción del 400 es UNA (la de platform/validate).

// urlRule envuelve la etiqueta `url`: solo `https` con host y ≤500. Rechaza por
// construcción `javascript:`, `data:`, `vbscript:`, `http:`, rutas relativas y
// cadenas sin host, sin mirar el dominio (FR-015/Q3).
type urlRule struct {
	URL string `validate:"required,url,max=500"`
}

// phoneRule envuelve la etiqueta `phone` de F2 (≥7 dígitos, separadores
// habituales y prefijo internacional opcional).
type phoneRule struct {
	Phone string `validate:"required,phone,max=32"`
}

// emailRule envuelve la etiqueta `email` de F2.
type emailRule struct {
	Email string `validate:"required,email,max=254"`
}

// isSpaceOrControl indica si el carácter es un espacio o un carácter de control
// (FR-015: el enlace no puede llevarlos).
func isSpaceOrControl(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}

// --- Validación de dominio (FR-014/FR-015/FR-020/FR-025) ---

// ministryFromInput normaliza el DTO del alta y valida el ministerio resultante:
// es el único punto donde se traduce el contrato a la entidad y se aplican las
// reglas de FR-014 (T615 lo usa también como paso previo al alta).
func ministryFromInput(in MinisterioInput) (Ministry, error) {
	ministry := Ministry{
		NameEs:           normalizeText(in.NameEs),
		NameEn:           normalizeOptional(in.NameEn),
		DescriptionEs:    normalizeText(in.DescriptionEs),
		DescriptionEn:    normalizeOptional(in.DescriptionEn),
		LeaderNameEs:     normalizeOptional(in.LeaderNameEs),
		LeaderNameEn:     normalizeOptional(in.LeaderNameEn),
		LeaderPhone:      normalizeOptional(in.LeaderPhone),
		LeaderEmail:      normalizeOptional(in.LeaderEmail),
		LeaderVisible:    in.LeaderVisible,
		WhatsappURL:      normalizeOptional(in.WhatsappURL),
		ScheduleEs:       normalizeOptional(in.ScheduleEs),
		ScheduleEn:       normalizeOptional(in.ScheduleEn),
		PlaceEs:          normalizeOptional(in.PlaceEs),
		PlaceEn:          normalizeOptional(in.PlaceEn),
		ImageFile:        normalizeOptional(in.ImageFile),
		ImageAltEs:       normalizeOptional(in.ImageAltEs),
		ImageAltEn:       normalizeOptional(in.ImageAltEn),
		PublicationState: publicationStateOr(in.PublicationState),
		SortOrder:        in.SortOrder,
	}
	if err := validateMinistry(ministry); err != nil {
		return Ministry{}, err
	}
	return ministry, nil
}

// validateMinistry aplica las invariantes del dominio sobre la entidad ya
// normalizada (creada o **fusionada**: P6-20/A6-3 exige revalidar el resultado
// del `PATCH`). Devuelve un `400 invalid` con **un motivo por campo** en
// `details`, con las mismas claves que el contrato y que los `CHECK` de la
// migración 000010 (la BD es la última línea de defensa de estos invariantes).
//
// Reglas: obligatorios en español, español por campo (`*En` exige su `*Es`; un
// campo vacío no obliga a nada, Q6), contacto excluyente (celular **o** correo,
// nunca ambos; el contacto exige nombre del encargado), enlace `https` seguro,
// imagen con texto alternativo y límites de longitud.
func validateMinistry(ministry Ministry) error {
	details := make(map[string]any)

	// 1) Obligatorios en español y límites de longitud por campo.
	addRequired(details, "nameEs", ministry.NameEs, MaxNameLength)
	addRequired(details, "descriptionEs", ministry.DescriptionEs, MaxDescriptionLength)
	addOptional(details, "nameEn", ministry.NameEn, MaxNameLength)
	addOptional(details, "descriptionEn", ministry.DescriptionEn, MaxDescriptionLength)
	addOptional(details, "leaderNameEs", ministry.LeaderNameEs, MaxLeaderNameLength)
	addOptional(details, "leaderNameEn", ministry.LeaderNameEn, MaxLeaderNameLength)
	addOptional(details, "scheduleEs", ministry.ScheduleEs, MaxScheduleLength)
	addOptional(details, "scheduleEn", ministry.ScheduleEn, MaxScheduleLength)
	addOptional(details, "placeEs", ministry.PlaceEs, MaxPlaceLength)
	addOptional(details, "placeEn", ministry.PlaceEn, MaxPlaceLength)
	addOptional(details, "imageAltEs", ministry.ImageAltEs, MaxImageAltLength)
	addOptional(details, "imageAltEn", ministry.ImageAltEn, MaxImageAltLength)

	// 2) Español por campo (FR-007): en los campos que pueden faltar en español,
	// la versión inglesa la exige (la BD lo refuerza con sus CHECK).
	addSpanishBase(details, "leaderNameEs", ministry.LeaderNameEs, ministry.LeaderNameEn)
	addSpanishBase(details, "scheduleEs", ministry.ScheduleEs, ministry.ScheduleEn)
	addSpanishBase(details, "placeEs", ministry.PlaceEs, ministry.PlaceEn)
	addSpanishBase(details, "imageAltEs", ministry.ImageAltEs, ministry.ImageAltEn)

	// 3) Encargado (FR-025/Q8): un solo medio de contacto y el contacto exige
	// nombre. Las claves y la redacción coinciden con los CHECK de la tabla y su
	// traducción en el repository, para que el 400 sea el mismo venga de donde
	// venga.
	if ministry.LeaderPhone != nil && ministry.LeaderEmail != nil {
		details["leaderEmail"] = "el celular y el correo del encargado no pueden ir juntos"
	}
	if (ministry.LeaderPhone != nil || ministry.LeaderEmail != nil) && ministry.LeaderNameEs == nil {
		details["leaderNameEs"] = "el contacto del encargado exige su nombre"
	}
	if ministry.LeaderPhone != nil {
		phone := *ministry.LeaderPhone
		if len([]rune(phone)) < MinLeaderPhoneLength || len([]rune(phone)) > MaxLeaderPhoneLength {
			details["leaderPhone"] = fmt.Sprintf("el celular admite entre %d y %d caracteres", MinLeaderPhoneLength, MaxLeaderPhoneLength)
		} else {
			mergeInvalid(details, "leaderPhone", validate.Struct(phoneRule{Phone: phone}))
		}
	}
	if ministry.LeaderEmail != nil {
		email := *ministry.LeaderEmail
		if len([]rune(email)) < MinLeaderEmailLength || len([]rune(email)) > MaxLeaderEmailLength {
			details["leaderEmail"] = fmt.Sprintf("el correo admite entre %d y %d caracteres", MinLeaderEmailLength, MaxLeaderEmailLength)
		} else {
			mergeInvalid(details, "leaderEmail", validate.Struct(emailRule{Email: email}))
		}
	}

	// 4) Enlace `https` válido y seguro (FR-015/Q3): forma `url` + sin espacios ni
	// caracteres de control. El valor NO se transforma: lo guardado se sirve tal
	// cual.
	if ministry.WhatsappURL != nil {
		if err := validate.Struct(urlRule{URL: *ministry.WhatsappURL}); err != nil {
			mergeInvalid(details, "whatsappUrl", err)
		} else if strings.ContainsFunc(*ministry.WhatsappURL, isSpaceOrControl) {
			details["whatsappUrl"] = "el enlace no puede llevar espacios ni caracteres de control"
		}
	}

	// 5) Imagen (FR-020): el nombre es el generado por el servidor y toda imagen
	// lleva su texto alternativo en español.
	if ministry.ImageFile != nil {
		if !storage.ValidName(*ministry.ImageFile) {
			details["imageFile"] = "sube la imagen desde el formulario"
		}
		if ministry.ImageAltEs == nil {
			details["imageAltEs"] = "toda imagen lleva su texto alternativo en español"
		}
	}

	if len(details) == 0 {
		return nil
	}
	return apperr.Invalid(validate.MessageInvalid, apperr.WithDetails(details))
}

// addRequired registra el motivo de un campo obligatorio en español: vacío
// (solo espacios cuenta como vacío) o por encima de su límite.
func addRequired(details map[string]any, field, value string, max int) {
	switch {
	case strings.TrimSpace(value) == "":
		details[field] = "Este campo es obligatorio."
	case len([]rune(value)) > max:
		details[field] = tooLongMessage(max)
	}
}

// addOptional registra el motivo de un campo opcional fuera de su límite (un
// campo ausente no obliga a nada, Q6).
func addOptional(details map[string]any, field string, value *string, max int) {
	if value != nil && len([]rune(*value)) > max {
		details[field] = tooLongMessage(max)
	}
}

// addSpanishBase registra el motivo de un campo inglés sin su español (FR-007).
func addSpanishBase(details map[string]any, fieldEs string, es, en *string) {
	if en != nil && es == nil {
		details[fieldEs] = "La versión en inglés exige el texto en español."
	}
}

// tooLongMessage es la redacción de platform/validate para un texto que supera
// su límite (una sola forma de decirlo en toda la aplicación).
func tooLongMessage(max int) string {
	return fmt.Sprintf("No puede tener más de %d caracteres.", max)
}

// mergeInvalid traspasa el motivo de un error de validación de
// platform/validate al campo indicado del mapa de detalles del ministerio: la
// redacción de cada regla de forma se escribe UNA vez (la del paquete de
// validación) y el campo se nombra como en el contrato.
func mergeInvalid(details map[string]any, field string, err error) {
	if err == nil {
		return
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		return
	}
	for _, reason := range domainErr.Details {
		details[field] = reason
	}
}

// --- Auditoría del dominio (R6-12) ---

// stateCode traduce un cambio de estado a su código (FR-017).
func stateCode(state PublicationState) string {
	if state == StatePublished {
		return audit.ActionMinistryPublish
	}
	return audit.ActionMinistryUnpublish
}

// leaderCode traduce un cambio de visibilidad de la ficha del encargado a su
// código (FR-017/A6-10).
func leaderCode(visible bool) string {
	if visible {
		return audit.ActionMinistryLeaderShow
	}
	return audit.ActionMinistryLeaderHide
}

// contentAction construye una acción de auditoría de contenido (`result`
// success), la unidad que el repository inserta dentro de la transacción de la
// mutación (R6-12).
func contentAction(actorID uuid.UUID, code, label string) audit.Action {
	return audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// contentLabel arma `Ministerios · <elemento>` (R6-12); con un elemento vacío
// queda la sección sola (`Ministerios`).
func contentLabel(element string) string {
	element = normalizeText(element)
	if element == "" {
		return moduleLabel
	}
	label := moduleLabel + " · " + element
	// Los límites del módulo (nombre ≤160, archivo generado por el servidor) no
	// alcanzan nunca el tope de la tabla; el recorte es la última defensa para
	// que ninguna etiqueta pueda romper el `CHECK` de `admin_actions`.
	if runes := []rune(label); len(runes) > maxTargetLabel {
		label = string(runes[:maxTargetLabel])
	}
	return label
}

// imageLabel arma la etiqueta de una subida (`Ministerios · Imagen ·
// <fileName>`, R6-12); el nombre lo genera el servidor, así que la etiqueta
// siempre cabe en los maxTargetLabel que admite la tabla.
func imageLabel(fileName string) string {
	return contentLabel("Imagen · " + normalizeText(fileName))
}

// recordActionBestEffort persiste una acción que ya va a fallar (rechazo o
// denegación) sin cambiar la respuesta: su fallo se queda en el log (R6-12).
func (s *service) recordActionBestEffort(ctx context.Context, action audit.Action) {
	if err := s.repository.RecordAction(ctx, action); err != nil {
		s.logError(ctx, "no se pudo registrar la acción administrativa", err)
	}
}

// recordContentFailure registra best-effort el rechazo de una mutación del
// módulo (`result='failure'`), con el mismo objetivo que habría tenido la acción
// (R6-12).
func (s *service) recordContentFailure(ctx context.Context, actorID uuid.UUID, code, label string) {
	s.recordActionBestEffort(ctx, audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultFailure,
	})
}

// deleteFile borra best-effort un archivo (A6-11): un fallo de limpieza se queda
// en el log y no cambia la respuesta ya escrita.
func (s *service) deleteFile(ctx context.Context, name string) {
	if s.store == nil || name == "" {
		return
	}
	if err := s.store.Delete(ctx, name); err != nil {
		s.logError(ctx, "no se pudo borrar la imagen reemplazada", err)
	}
}

// logError registra un fallo best-effort con el logger de la petición (que lleva
// el request_id) o, si no hay, con el logger base del service.
func (s *service) logError(ctx context.Context, message string, err error) {
	logger := s.logger
	if requestLogger := httpserver.RequestLoggerFromContext(ctx); requestLogger != nil {
		logger = requestLogger
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Error(message, slog.Any("error", err))
}

// RecordAction persiste una acción de auditoría fuera de la transacción de una
// mutación (denegaciones y rechazos best-effort). El repository es el único
// punto que escribe `admin_actions` (R6-12).
func (r *repository) RecordAction(ctx context.Context, action audit.Action) error {
	return r.insertAdminAction(ctx, action)
}

// compile-time check: *repository must satisfy the port the service declares.
var _ Repository = (*repository)(nil)
