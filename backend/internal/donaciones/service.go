package donaciones

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/validate"
)

// service.go reúne la base del service del dominio donaciones (T712): el puerto
// `Repository` que declara para sí (arq. R3), las constantes de negocio
// revisables —único punto de verdad de los límites—, la normalización de textos
// (trim; `""`/espacios → NULL, analyze I6), la validación de dominio completa
// (FR-014/FR-015/FR-016/FR-019/FR-021) y los helpers de auditoría del módulo
// (R7-10).
//
// No conoce net/http ni pgx: solo reglas de negocio. La persistencia y la
// auditoría atómica viven en el repository; aquí se decide QUÉ datos guardar y
// QUÉ filas de auditoría acompañan cada mutación. Los archivos reparten las
// responsabilidades: `service_public.go` (sección y bloque de la portada),
// `service_admin.go` (panel) y `service_audit.go` (denegaciones y rechazos).

// --- Constantes de negocio (único punto de verdad de los límites, FR-014) ---

const (
	// MaxInstitutionLength es el límite de la entidad financiera (no traducible).
	MaxInstitutionLength = 160
	// MaxHolderLength es el límite del titular por idioma (HolderEs/HolderEn).
	MaxHolderLength = 160
	// MaxTextLength es el límite de los textos de la sección por idioma (Q5/Q11).
	MaxTextLength = 2000
	// MaxNumberInputLength acota el número tal como lo escribe el equipo (IBAN
	// con espacios o guiones; SINPE con o sin `+506`) — el tope del contrato.
	MaxNumberInputLength = 60
	// MaxNumberLength acota el número YA normalizado (IBAN ≤34; SINPE = 12).
	MaxNumberLength = 34
	// MinSortOrder y MaxSortOrder acotan el orden manual (Q1/FR-004) al rango de
	// la columna `sort_order INTEGER` (int32): un valor fuera de rango se rechaza
	// con `400` (`details.sortOrder`) en vez de truncarse en silencio al guardar.
	MinSortOrder = math.MinInt32
	// MaxSortOrder es el tope del orden manual.
	MaxSortOrder = math.MaxInt32
	// maxPublicItems es el tope de la sección pública (RG7-12): límite de
	// seguridad del servidor, lo aplica GetOfrendas (T713) y lo documenta el
	// contrato.
	maxPublicItems = 200
	// maxPortadaItems es el tope FIRME del bloque de la portada (FR-022): lo
	// aplica GetPortadaOfrendas (T714), no crece con el contenido y el contrato no
	// expone ningún parámetro de límite.
	maxPortadaItems = 5
	// maxTargetLabel es la longitud que admite `admin_actions.target_label`
	// (migración 000004).
	maxTargetLabel = 254
	// messageNotFound es el mensaje del 404 del módulo.
	messageNotFound = "El método de donación no existe"
)

// Idiomas admitidos por la API pública (FR-008).
const (
	// LangES es el idioma base (obligatorio).
	LangES = "es"
	// LangEN es el idioma opcional con fallback a `es`.
	LangEN = "en"
)

// moduleLabel es la etiqueta legible del módulo en `targetLabel` (R7-10):
// `Donaciones · <institution> · <number>` y `Donaciones · Texto introductorio` /
// `Donaciones · Texto de uso`.
const moduleLabel = "Donaciones"

// --- Puerto del repositorio (definido por quien lo consume, arq. R3) ---

// Repository es el puerto de datos que el service de donaciones necesita. Los
// errores de PostgreSQL llegan ya traducidos a apperr. Las mutaciones reciben las
// filas de auditoría que deben insertarse en la MISMA transacción (variádicas):
// o el cambio y su registro, o nada (FR-017/R7-10).
type Repository interface {
	// ListPublishedDonationMethods es la ÚNICA vía pública de métodos (sección y
	// bloque): solo `published`, orden `CASE kind …, sort_order, id` y tope.
	ListPublishedDonationMethods(ctx context.Context, limit int) ([]MetodoDonacion, error)
	// ListPublishedDonationTexts es la ÚNICA vía pública de textos: solo
	// `published`, orden de clave.
	ListPublishedDonationTexts(ctx context.Context) ([]TextoDonacion, error)

	// Gestión del panel (FR-010/FR-014/FR-015/FR-016/FR-019).
	InsertDonationMethod(ctx context.Context, method MetodoDonacion, actions ...audit.Action) (MetodoDonacion, error)
	UpdateDonationMethod(ctx context.Context, method MetodoDonacion, actions ...audit.Action) (MetodoDonacion, error)
	DeleteDonationMethod(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	GetDonationMethodByID(ctx context.Context, id uuid.UUID) (MetodoDonacion, bool, error)
	ListDonationMethods(ctx context.Context, state *PublicationState, params paginate.Params) ([]MetodoDonacion, error)
	CountDonationMethods(ctx context.Context, state *PublicationState) (int64, error)
	ExistsDonationMethodDuplicate(ctx context.Context, kind MetodoKind, number string, excludeID *uuid.UUID) (bool, error)

	// Textos del panel (FR-021): singleton por clave, sin borrado (P7-17).
	ListDonationTexts(ctx context.Context) ([]TextoDonacion, error)
	GetDonationTextByKey(ctx context.Context, key TextKey) (TextoDonacion, bool, error)
	UpsertDonationText(ctx context.Context, text TextoDonacion, actions ...audit.Action) (TextoDonacion, error)

	// RecordAction persiste una acción de auditoría best-effort (denegaciones y
	// rechazos): fuera de transacción y sin cambiar la respuesta (R7-10).
	RecordAction(ctx context.Context, action audit.Action) error
}

// ServiceDeps agrupa las dependencias del service.
type ServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository Repository
	// Logger registra los fallos best-effort; nil usa el logger por defecto.
	Logger *slog.Logger
}

// service implementa la cara pública, la gestión del panel y el registro de
// auditoría del dominio. Una sola instancia satisface las tres
// responsabilidades; los archivos service_public.go, service_admin.go y
// service_audit.go reparten los métodos.
type service struct {
	repository Repository
	logger     *slog.Logger
}

// NewService construye el service de donaciones. Si Logger es nil se usa el
// logger por defecto.
func NewService(deps ServiceDeps) *service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &service{repository: deps.Repository, logger: logger}
}

// --- Normalización (analyze I6) ---

// horizontalSpace colapsa las secuencias de espacios y tabulaciones internas a
// un único espacio, sin tocar los saltos de línea (los párrafos de los textos).
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

// normalizeLang valida y normaliza el idioma pedido por el visitante: vacío es el
// idioma base y cualquier otro valor es `400 invalid` con details.lang (FR-008).
// Es la única implementación: la usan la sección (T713) y el bloque (T714).
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

// --- Validación de dominio (FR-014/FR-015/FR-016/FR-019/FR-021) ---

// methodFromInput normaliza el DTO del alta y valida el método resultante: es el
// único punto donde se traduce el contrato a la entidad y se aplican las reglas
// de FR-014 (T715 lo usa como paso previo al alta). El número se normaliza con
// `numero.go` antes de persistir (FR-015/FR-016).
func methodFromInput(in DonacionInput) (MetodoDonacion, error) {
	method := MetodoDonacion{
		Kind:             MetodoKind(strings.TrimSpace(in.Kind)),
		Number:           in.Number,
		Institution:      normalizeText(in.Institution),
		HolderEs:         normalizeOptional(in.HolderEs),
		HolderEn:         normalizeOptional(in.HolderEn),
		PublicationState: publicationStateOr(in.PublicationState),
		SortOrder:        in.SortOrder,
	}
	if currency := strings.TrimSpace(in.Currency); currency != "" {
		c := Currency(currency)
		method.Currency = &c
	}
	if err := validateMethod(&method); err != nil {
		return MetodoDonacion{}, err
	}
	return method, nil
}

// validateMethod aplica las invariantes del dominio sobre la entidad ya
// normalizada (creada o **fusionada**: R7-15 exige revalidar el resultado del
// `PATCH`). Devuelve un `400 invalid` con **un motivo por campo** en `details`,
// con las mismas claves que el contrato y que los `CHECK` de la migración
// 000012 (la BD es la última línea de defensa de estos invariantes).
//
// Reglas: obligatorios **solo** `kind`, `number` e `institution`; número
// normalizado y validado por tipo con `numero.go` (se escribe de vuelta el valor
// normalizado); `currency` obligatoria en IBAN y rechazada en SINPE; `holderEn`
// exige `holderEs`; estado del registro y rango del orden manual.
func validateMethod(method *MetodoDonacion) error {
	details := make(map[string]any)

	// 1) Tipo.
	if !method.Kind.Valid() {
		details["kind"] = "este campo es obligatorio y solo admite iban o sinpe."
	}

	// 2) Entidad financiera (obligatoria, ≤160, no traducible).
	addRequired(details, "institution", method.Institution, MaxInstitutionLength)

	// 3) Número (obligatorio; normalizado y validado por tipo).
	raw := strings.TrimSpace(method.Number)
	switch {
	case raw == "":
		details["number"] = "Este campo es obligatorio."
	case !method.Kind.Valid():
		// Sin tipo válido no se puede validar el formato: el detalle de `kind`
		// ya lo explica; no se inventa el del número.
	case len([]rune(raw)) > MaxNumberInputLength:
		details["number"] = "El número no tiene el formato de su tipo"
	default:
		normalized, err := normalizarNumero(method.Kind, raw)
		if err != nil {
			details["number"] = "El número no tiene el formato de su tipo"
		} else {
			method.Number = normalized
		}
	}

	// 4) Moneda por tipo (Q2/FR-014): obligatoria en IBAN y prohibida en SINPE.
	switch {
	case method.Kind == KindIban && (method.Currency == nil || !method.Currency.Valid()):
		details["currency"] = "la moneda es obligatoria en las cuentas IBAN"
	case method.Kind == KindSinpe && method.Currency != nil:
		details["currency"] = "el SINPE Móvil no lleva moneda"
	case method.Currency != nil && !method.Currency.Valid():
		details["currency"] = "la moneda solo admite CRC o USD"
	}

	// 5) Titular (opcional y traducible): límites y español por campo (FR-007).
	addOptional(details, "holderEs", method.HolderEs, MaxHolderLength)
	addOptional(details, "holderEn", method.HolderEn, MaxHolderLength)
	addSpanishBase(details, "holderEs", method.HolderEs, method.HolderEn)

	// 6) Estado del registro.
	if !method.PublicationState.Valid() {
		details["publicationState"] = "solo se admite draft o published"
	}

	// 7) Orden manual (Q1/FR-004): la columna es `INTEGER` (int32).
	if method.SortOrder < MinSortOrder || method.SortOrder > MaxSortOrder {
		details["sortOrder"] = fmt.Sprintf("el orden admite valores entre %d y %d", MinSortOrder, MaxSortOrder)
	}

	if len(details) == 0 {
		return nil
	}
	return apperr.Invalid(validate.MessageInvalid, apperr.WithDetails(details))
}

// normalizarNumero delega en las funciones puras del dominio (`numero.go`): es
// la única vía de normalización de números del service.
func normalizarNumero(kind MetodoKind, raw string) (string, error) {
	switch kind {
	case KindIban:
		return NormalizarIban(raw)
	case KindSinpe:
		return NormalizarSinpe(raw)
	default:
		return "", ErrNumeroInvalido
	}
}

// textoFromInput normaliza el DTO del `PUT` de un texto y lo valida (FR-021):
// `textEs` obligatorio, `textEn` opcional y su propio estado (Q10).
func textoFromInput(key TextKey, in DonacionTextoInput) (TextoDonacion, error) {
	texto := TextoDonacion{
		TextKey:          key,
		BodyEs:           normalizeText(in.TextEs),
		BodyEn:           normalizeOptional(in.TextEn),
		PublicationState: publicationStateOr(in.PublicationState),
	}
	if err := validateTexto(&texto); err != nil {
		return TextoDonacion{}, err
	}
	return texto, nil
}

// validateTexto aplica las invariantes de un texto de la sección: clave del
// catálogo, `textEs` obligatorio (≤2.000), `textEn` opcional (≤2.000, exige
// español) y estado del registro.
func validateTexto(texto *TextoDonacion) error {
	details := make(map[string]any)

	if !texto.TextKey.Valid() {
		details["textKey"] = "la clave debe ser intro o usage"
	}
	addRequired(details, "textEs", texto.BodyEs, MaxTextLength)
	addOptional(details, "textEn", texto.BodyEn, MaxTextLength)
	if texto.BodyEn != nil && strings.TrimSpace(texto.BodyEs) == "" {
		details["textEs"] = "la versión en inglés exige el texto en español."
	}
	if !texto.PublicationState.Valid() {
		details["publicationState"] = "solo se admite draft o published"
	}

	if len(details) == 0 {
		return nil
	}
	return apperr.Invalid(validate.MessageInvalid, apperr.WithDetails(details))
}

// addRequired registra el motivo de un campo obligatorio: vacío (solo espacios
// cuenta como vacío) o por encima de su límite.
func addRequired(details map[string]any, field, value string, max int) {
	switch {
	case strings.TrimSpace(value) == "":
		details[field] = "Este campo es obligatorio."
	case len([]rune(value)) > max:
		details[field] = tooLongMessage(max)
	}
}

// addOptional registra el motivo de un campo opcional fuera de su límite (un
// campo ausente no obliga a nada).
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

// --- Auditoría del dominio (R7-10) ---

// stateCode traduce un cambio de estado de un método a su código (FR-017).
func stateCode(state PublicationState) string {
	if state == StatePublished {
		return audit.ActionDonationPublish
	}
	return audit.ActionDonationUnpublish
}

// textStateCode traduce un cambio de estado de un texto a su código (FR-017).
func textStateCode(state PublicationState) string {
	if state == StatePublished {
		return audit.ActionDonationTextPublish
	}
	return audit.ActionDonationTextUnpublish
}

// contentAction construye una acción de auditoría de contenido (`result`
// success), la unidad que el repository inserta dentro de la transacción de la
// mutación (R7-10).
func contentAction(actorID uuid.UUID, code, label string) audit.Action {
	return audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// contentLabel arma `Donaciones · <elemento>` (R7-10); con un elemento vacío
// queda la sección sola (`Donaciones`).
func contentLabel(element string) string {
	element = normalizeText(element)
	if element == "" {
		return moduleLabel
	}
	label := moduleLabel + " · " + element
	// Los límites del módulo (entidad ≤160, número ≤34) no alcanzan nunca el
	// tope de la tabla; el recorte es la última defensa para que ninguna etiqueta
	// pueda romper el `CHECK` de `admin_actions`.
	if runes := []rune(label); len(runes) > maxTargetLabel {
		label = string(runes[:maxTargetLabel])
	}
	return label
}

// methodLabel arma `Donaciones · <institution> · <number>` (R7-10): los dos
// valores ya normalizados del método.
func methodLabel(method MetodoDonacion) string {
	return contentLabel(method.Institution + " · " + method.Number)
}

// textLabel arma la etiqueta de un texto (`Donaciones · Texto introductorio` /
// `Donaciones · Texto de uso`, R7-10).
func textLabel(key TextKey) string {
	switch key {
	case TextKeyIntro:
		return moduleLabel + " · Texto introductorio"
	case TextKeyUsage:
		return moduleLabel + " · Texto de uso"
	default:
		return moduleLabel
	}
}

// recordActionBestEffort persiste una acción que ya va a fallar (rechazo o
// denegación) sin cambiar la respuesta: su fallo se queda en el log (R7-10).
func (s *service) recordActionBestEffort(ctx context.Context, action audit.Action) {
	if err := s.repository.RecordAction(ctx, action); err != nil {
		s.logError(ctx, "no se pudo registrar la acción administrativa", err)
	}
}

// recordContentFailure registra best-effort el rechazo de una mutación del
// módulo (`result='failure'`), con el mismo objetivo que habría tenido la acción
// (R7-10).
func (s *service) recordContentFailure(ctx context.Context, actorID uuid.UUID, code, label string) {
	s.recordActionBestEffort(ctx, audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultFailure,
	})
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
// punto que escribe `admin_actions` (R7-10).
func (r *repository) RecordAction(ctx context.Context, action audit.Action) error {
	return r.insertAdminAction(ctx, action)
}

// compile-time check: *repository must satisfy the port the service declares.
var _ Repository = (*repository)(nil)
