package agenda

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
)

// service.go reúne la base del service del dominio agenda (T417): el puerto
// `Repository` que declara para sí (arq. R3), las constantes de negocio
// revisables (único punto de verdad de los límites), la normalización de textos
// (analyze I6), la validación de fechas y horas de calendario (FR-015) y la
// **regla de "próximo"** (R4-2) como una sola función de dominio, calculada con
// la zona horaria del lugar (`APP_TZ`, T409).
//
// No conoce net/http ni pgx: solo reglas de negocio. La persistencia y la
// auditoría atómica viven en el repository; aquí se decide QUÉ datos guardar y
// QUÉ filas de auditoría acompañan cada mutación (R4-10).
//
// Sin cron y sin cambios de estado: el retiro del contenido vencido es un
// **filtro de lectura** (FR-006/Q4); el `publication_state` de un elemento
// vencido sigue siendo `published` y el panel lo marca `isPast`.

// --- Constantes de negocio (único punto de verdad de los límites) ---

const (
	// MaxTitleLength es el límite del título por idioma (FR-018/FR-019).
	MaxTitleLength = 160
	// MaxDetailLength es el límite del detalle por idioma (FR-018/FR-019).
	MaxDetailLength = 2000
	// MaxPlaceLength es el límite del lugar por idioma.
	MaxPlaceLength = 200
	// MaxURLLength acota el enlace de mapa (coincide con la etiqueta `url` de
	// platform/validate y con el contrato).
	MaxURLLength = 500
	// MaxOrganizerLength es el límite del organizador por idioma.
	MaxOrganizerLength = 160
	// MaxFrequencyLength es el límite del texto de frecuencia por idioma (Q1).
	MaxFrequencyLength = 160
	// MaxImageAltLength es el límite del texto alternativo por idioma (FR-020).
	MaxImageAltLength = 300
	// MaxCategoryNameLength es el límite del nombre de categoría por idioma.
	MaxCategoryNameLength = 120
	// MaxEventDates acota la lista de fechas de un evento (FR-015).
	MaxEventDates = 50
	// MaxCategories acota el catálogo de categorías (Q6).
	MaxCategories = 100
	// MaxBlockItems acota CADA bloque del listado público (RG4-8): es un tope de
	// seguridad del servidor, lo aplica GetProximamente (T418) y lo documenta el
	// contrato.
	MaxBlockItems = 200
	// MediaPathPrefix es el prefijo público de las imágenes (contrato:
	// `/api/v1/media/<fileName>`).
	MediaPathPrefix = "/api/v1/media/"
)

// Idiomas admitidos por la API pública (R3-2/FR-008).
const (
	// LangES es el idioma base (obligatorio).
	LangES = "es"
	// LangEN es el idioma opcional con fallback a `es`.
	LangEN = "en"
)

// Module identifica el módulo de contenido: fija la etiqueta y los códigos de
// auditoría que exige R4-10 (un código por módulo y operación).
type Module string

const (
	// ModuleEvents es el módulo de eventos (`event.*`).
	ModuleEvents Module = "event"
	// ModuleActivities es el módulo de actividades (`activity.*`).
	ModuleActivities Module = "activity"
	// ModuleCategories es el módulo del catálogo de categorías (`category.*`).
	ModuleCategories Module = "category"
)

// Label es la etiqueta legible del módulo en `targetLabel` (R4-10).
func (m Module) Label() string {
	switch m {
	case ModuleEvents:
		return "Eventos"
	case ModuleActivities:
		return "Actividades"
	case ModuleCategories:
		return "Categorías"
	default:
		return string(m)
	}
}

// CreateCode, UpdateCode y DeleteCode devuelven los códigos del registro cerrado
// de `platform/audit` (R4-10). El alta usa `*.create` **aunque el elemento nazca
// publicado**: `*.publish`/`*.unpublish` registran SOLO cambios de estado.
func (m Module) CreateCode() string { return string(m) + ".create" }

// UpdateCode es el código de una edición de datos sin cambio de estado.
func (m Module) UpdateCode() string { return string(m) + ".update" }

// DeleteCode es el código del borrado.
func (m Module) DeleteCode() string { return string(m) + ".delete" }

// ImageCode es el código de la subida de imagen del módulo (analyze I8/R4-10).
func (m Module) ImageCode() string { return string(m) + ".image.upload" }

// PublishCode es el código de publicar (solo cambio de estado).
func (m Module) PublishCode() string { return string(m) + ".publish" }

// UnpublishCode es el código de retirar (solo cambio de estado).
func (m Module) UnpublishCode() string { return string(m) + ".unpublish" }

// StateCode traduce un cambio de estado a su código de auditoría (R4-10).
func (m Module) StateCode(state PublicationState) string {
	if state == StatePublished {
		return m.PublishCode()
	}
	return m.UnpublishCode()
}

// --- Puerto del repositorio (definido por quien lo consume, arq. R3) ---

// Repository es el puerto de datos que el service de agenda necesita. Los
// errores de PostgreSQL llegan ya traducidos a apperr. Las mutaciones reciben
// las filas de auditoría que deben insertarse en la MISMA transacción
// (variádicas): o el cambio y su registro, o nada (FR-017/R4-10).
type Repository interface {
	// Catálogo de categorías (Q6/R4-5).
	ListCategories(ctx context.Context) ([]Category, error)
	GetCategoryByID(ctx context.Context, id uuid.UUID) (Category, bool, error)
	InsertCategory(ctx context.Context, category Category, actions ...audit.Action) (Category, error)
	UpdateCategory(ctx context.Context, category Category, actions ...audit.Action) (Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	CountCategories(ctx context.Context) (int64, error)
	CountCategoryUsage(ctx context.Context, id uuid.UUID) (int64, error)

	// Eventos con su lista de fechas (FR-004/FR-010/FR-013).
	InsertEvent(ctx context.Context, event Event, dates []EventDate, actions ...audit.Action) (Event, error)
	UpdateEvent(ctx context.Context, event Event, dates []EventDate, actions ...audit.Action) (Event, error)
	DeleteEvent(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (Event, bool, error)
	ListEventDates(ctx context.Context, eventID uuid.UUID) ([]EventDate, error)
	ListEventDatesForEvents(ctx context.Context, eventIDs []uuid.UUID) ([]EventDate, error)
	ListUpcomingEventDatesForEvents(ctx context.Context, eventIDs []uuid.UUID, today, nowTime string) ([]EventDate, error)
	ExistsEventDuplicate(ctx context.Context, title, place string, dates []string, excludeID *uuid.UUID) (bool, error)
	ListEvents(ctx context.Context, state *PublicationState, params paginate.Params) ([]Event, error)
	CountEvents(ctx context.Context, state *PublicationState) (int64, error)
	ListUpcomingEvents(ctx context.Context, today, nowTime string, limit int) ([]Event, error)

	// Actividades de fecha única (FR-005/FR-011/FR-013).
	InsertActivity(ctx context.Context, activity Activity, actions ...audit.Action) (Activity, error)
	UpdateActivity(ctx context.Context, activity Activity, actions ...audit.Action) (Activity, error)
	DeleteActivity(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	GetActivityByID(ctx context.Context, id uuid.UUID) (Activity, bool, error)
	ListActivities(ctx context.Context, state *PublicationState, params paginate.Params) ([]Activity, error)
	CountActivities(ctx context.Context, state *PublicationState) (int64, error)
	ListUpcomingActivities(ctx context.Context, today, nowTime string, limit int) ([]Activity, error)

	// RecordAction persiste una acción de auditoría best-effort (denegaciones y
	// rechazos): fuera de transacción y sin cambiar la respuesta (R4-10).
	RecordAction(ctx context.Context, action audit.Action) error

	// ListNextUpcomingItems es el extracto público de la portada (US7/FR-025):
	// eventos y actividades publicados y próximos, mezclados por cercanía, con
	// la ventana `[today, windowEnd]` y su respaldo (R4-16).
	ListNextUpcomingItems(ctx context.Context, today, nowTime, windowEnd string, limit int) ([]NextItem, error)
}

// ServiceDeps agrupa las dependencias del service.
type ServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository Repository
	// Store es el almacén de imágenes (platform/storage). Puede ser nil en
	// pruebas que no ejerciten la subida.
	Store storage.Store
	// Location es la zona horaria del lugar (APP_TZ, T409) con la que se
	// calcula el corte "hoy"/"ahora" de la regla R4-2. Obligatoria: sin ella el
	// service falla con un error claro (fail-closed) en vez de cortar en UTC.
	Location *time.Location
	// Now es el reloj (inyectable en pruebas). nil usa time.Now.
	Now func() time.Time
	// Logger registra los fallos best-effort; nil usa el logger por defecto.
	Logger *slog.Logger
}

// service implementa «lo que viene», la gestión del panel y el registro de
// auditoría del dominio. Una sola instancia satisface las tres
// responsabilidades; los archivos service_public.go, service_admin.go y
// service_audit.go reparten los métodos.
type service struct {
	repository Repository
	store      storage.Store
	location   *time.Location
	now        func() time.Time
	logger     *slog.Logger
}

// NewService construye el service de agenda. Si Logger es nil se usa el logger
// por defecto; si Now es nil, time.Now.
func NewService(deps ServiceDeps) *service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	now := deps.Now
	if now == nil {
		now = time.Now
	}
	return &service{
		repository: deps.Repository,
		store:      deps.Store,
		location:   deps.Location,
		now:        now,
		logger:     logger,
	}
}

// --- Normalización ---

// horizontalSpace colapsa las secuencias de espacios y tabulaciones internas a
// un único espacio, sin tocar los saltos de línea (los párrafos del detalle).
var horizontalSpace = regexp.MustCompile(`[ \t]+`)

// normalizeText aplica trim y colapsa los espacios internos. Conserva los saltos
// de línea y todo el contenido (tildes, ñ, ¿¡ y emojis intactos).
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

// --- La regla de "próximo" (R4-2): una sola implementación ---

// cutoff devuelve el "hoy" (fecha de calendario) y el "ahora" (hora) en la zona
// del lugar (APP_TZ, T409). Son los parámetros que reciben las consultas
// públicas (`…Upcoming`): el corte es el día real de la iglesia, no el del
// contenedor (UTC). Sin zona configurada devuelve un error interno (fail-closed:
// mejor un error claro que un corte silencioso).
func (s *service) cutoff() (string, string, error) {
	if s.location == nil {
		return "", "", apperr.Internal(errors.New("agenda: zona horaria (APP_TZ) no configurada"))
	}
	in := s.now().In(s.location)
	return in.Format(DateFormat), in.Format(TimeFormat), nil
}

// isUpcoming aplica la regla de "próximo" (R4-2) a una fecha concreta:
//
//   - fecha posterior a hoy → próxima;
//   - hoy sin hora de fin → próxima todo el día (se retira al terminar el día);
//   - hoy con hora de fin ya terminada → NO próxima (se retira al terminar su
//     hora);
//   - fecha anterior a hoy → NO próxima.
//
// Las horas "HH:MM" se comparan como texto: el formato de 24 h con dos dígitos
// ordena igual que la hora real.
func isUpcoming(date, today, nowTime string, endTime *string) bool {
	switch {
	case date > today:
		return true
	case date < today:
		return false
	default:
		return endTime == nil || *endTime >= nowTime
	}
}

// --- Validación de fechas y horas (FR-015) ---

// parseDate valida una fecha de calendario real en formato "YYYY-MM-DD"
// (rechaza 2026-13-32) y la devuelve normalizada.
func parseDate(value string) (string, error) {
	date := strings.TrimSpace(value)
	parsed, err := time.Parse(DateFormat, date)
	if err != nil || parsed.Format(DateFormat) != date {
		return "", apperr.Invalid(
			"La fecha no es válida",
			apperr.WithDetails(map[string]any{"date": "usa una fecha real en formato AAAA-MM-DD"}),
		)
	}
	return date, nil
}

// parseClock valida una hora "HH:MM" de 24 h y la devuelve normalizada.
func parseClock(value string) (string, error) {
	clock := strings.TrimSpace(value)
	if !clockPattern.MatchString(clock) {
		return "", apperr.Invalid(
			"La hora no es válida",
			apperr.WithDetails(map[string]any{"startTime": "usa el formato HH:MM"}),
		)
	}
	return clock, nil
}

// normalizeRange valida el horario opcional de una fecha (FR-015/Q2): las horas
// son "HH:MM"; `endTime` **exige** `startTime` (decisión firme de la puerta 2) y
// es posterior a él.
func normalizeRange(startRaw, endRaw string) (*string, *string, error) {
	startText := strings.TrimSpace(startRaw)
	endText := strings.TrimSpace(endRaw)

	if startText == "" && endText == "" {
		return nil, nil, nil
	}
	if endText != "" && startText == "" {
		return nil, nil, apperr.Invalid(
			"La hora de fin no es válida",
			apperr.WithDetails(map[string]any{"endTime": "indica primero la hora de inicio"}),
		)
	}

	start, err := parseClock(startText)
	if err != nil {
		return nil, nil, err
	}
	if endText == "" {
		return &start, nil, nil
	}
	end, err := parseClock(endText)
	if err != nil {
		return nil, nil, apperr.Invalid(
			"La hora de fin no es válida",
			apperr.WithDetails(map[string]any{"endTime": "usa el formato HH:MM"}),
		)
	}
	if end <= start {
		return nil, nil, apperr.Invalid(
			"La hora de fin no es válida",
			apperr.WithDetails(map[string]any{"endTime": "debe ser posterior a la hora de inicio"}),
		)
	}
	return &start, &end, nil
}

// normalizeEventDates valida y normaliza la lista de fechas de un evento
// (FR-015/R4-11): entre 1 y MaxEventDates fechas, cada una un día real con
// horario coherente y sin repetir (una fecha por día; `UNIQUE (event_id,
// event_date)` lo refuerza en la base). Las fechas pasadas están permitidas: el
// elemento simplemente no será público sin ninguna próxima (edge case).
func normalizeEventDates(inputs []FechaInput) ([]EventDate, error) {
	if len(inputs) == 0 {
		return nil, apperr.Invalid(
			"El evento necesita al menos una fecha",
			apperr.WithDetails(map[string]any{"dates": "carga al menos una fecha"}),
		)
	}
	if len(inputs) > MaxEventDates {
		return nil, apperr.Invalid(
			fmt.Sprintf("No se pueden guardar más de %d fechas por evento", MaxEventDates),
			apperr.WithDetails(map[string]any{"dates": "demasiadas fechas"}),
		)
	}

	dates := make([]EventDate, 0, len(inputs))
	seen := make(map[string]bool, len(inputs))
	for _, input := range inputs {
		date, err := parseDate(input.Date)
		if err != nil {
			return nil, err
		}
		if seen[date] {
			return nil, apperr.Invalid(
				"Hay fechas repetidas",
				apperr.WithDetails(map[string]any{"dates": "cada fecha puede aparecer una sola vez"}),
			)
		}
		seen[date] = true

		start, end, err := normalizeRange(input.StartTime, input.EndTime)
		if err != nil {
			return nil, err
		}
		dates = append(dates, EventDate{Date: date, StartTime: start, EndTime: end})
	}
	return dates, nil
}

// --- Imágenes (FR-020) ---

// validateImage exige que el nombre tenga el patrón generado por el servidor y
// que, si hay imagen, el texto alternativo en español esté presente (FR-020);
// ambos CHECK viven también en la base.
func validateImage(imageFile, imageAltEs string) error {
	file := strings.TrimSpace(imageFile)
	if file == "" {
		return nil
	}
	if !storage.ValidName(file) {
		return apperr.Invalid(
			"El nombre de la imagen no es válido",
			apperr.WithDetails(map[string]any{"imageFile": "sube la imagen desde el formulario"}),
		)
	}
	if normalizeText(imageAltEs) == "" {
		return apperr.Invalid(
			"Falta el texto alternativo de la imagen",
			apperr.WithDetails(map[string]any{"imageAltEs": "describe la imagen para quien no puede verla"}),
		)
	}
	return nil
}

// --- Auditoría del dominio (R4-10) ---

// contentAction construye una acción de auditoría de contenido (`result`
// success), la unidad que el repository inserta dentro de la transacción de la
// mutación (R4-10).
func contentAction(actorID uuid.UUID, code, label string) audit.Action {
	return audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// contentLabel arma `Eventos · <elemento>` (R4-10); con `element` vacío queda la
// sección sola (`Eventos`), y para las subidas se añade `Imagen · <fileName>`.
func contentLabel(module Module, element string) string {
	label := module.Label()
	element = normalizeText(element)
	if element == "" {
		return label
	}
	return label + " · " + element
}

// imageLabel arma la etiqueta de una subida de imagen
// (`Eventos · Imagen · <fileName>`, R4-10).
func imageLabel(module Module, fileName string) string {
	return contentLabel(module, "Imagen · "+normalizeText(fileName))
}

// recordActionBestEffort persiste una acción que ya va a fallar (rechazo o
// denegación) sin cambiar la respuesta: su fallo se queda en el log (R4-10).
func (s *service) recordActionBestEffort(ctx context.Context, action audit.Action) {
	if err := s.repository.RecordAction(ctx, action); err != nil {
		s.logError(ctx, "no se pudo registrar la acción administrativa", err)
	}
}

// recordContentFailure registra best-effort el rechazo de una mutación del
// módulo (`result='failure'`), con el mismo objetivo que habría tenido la acción
// (R4-10).
func (s *service) recordContentFailure(ctx context.Context, actorID uuid.UUID, code, label string) {
	s.recordActionBestEffort(ctx, audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultFailure,
	})
}

// isConflict indica si el error es un conflicto de dominio (409). Se usa para
// registrar el rechazo best-effort cuando lo detecta la BASE (el duplicado de
// actividades lo garantiza `UNIQUE (title_es, activity_date, place_es)`, R4-9):
// los intentos fallidos también quedan en la auditoría (FR-017/R4-10).
func isConflict(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindConflict
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
// punto que escribe `admin_actions` (R4-10).
func (r *repository) RecordAction(ctx context.Context, action audit.Action) error {
	return r.insertAdminAction(ctx, action)
}

// compile-time check: *repository must satisfy the port the service declares.
var _ Repository = (*repository)(nil)
