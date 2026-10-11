package portada

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/validate"
)

// service.go reúne la base del service de la portada (T317): el puerto
// `Repository` que el service declara para sí (arq. R3), las constantes de
// negocio revisables, la normalización de textos y teléfonos, las invariantes
// compartidas por el resto de los archivos (`service_public.go`,
// `service_admin.go`, `service_audit.go`) y los helpers de auditoría del dominio
// (`contentAction`/`contentLabel`).
//
// No conoce net/http ni pgx: solo reglas de negocio. La persistencia y la
// auditoría atómica viven en el repository; aquí se deciden QUÉ datos guardar y
// QUÉ filas de auditoría acompañan cada mutación (R3-11).

// --- Constantes de negocio (único punto de verdad de los límites, R3-14) ---

const (
	// MaxAboutLength es el límite de «quiénes somos» por idioma (FR-003).
	MaxAboutLength = 1000
	// MaxScheduleItems acota la colección del horario (R3-14).
	MaxScheduleItems = 50
	// MaxWhatsappChannels acota los canales de WhatsApp (R3-14).
	MaxWhatsappChannels = 20
	// MediaPathPrefix es el prefijo público de las imágenes (contrato:
	// `/api/v1/media/<fileName>`).
	MediaPathPrefix = "/api/v1/media/"
	// maxMediaFileName es el límite del nombre de archivo en el contrato.
	maxMediaFileName = 120
	// maxURLLength acota cualquier enlace del dominio (coincide con la etiqueta
	// `url` de platform/validate y con el contrato).
	maxURLLength = 500
)

// Idiomas admitidos por la API pública (R3-2/FR-008).
const (
	// LangES es el idioma base (obligatorio).
	LangES = "es"
	// LangEN es el idioma opcional con fallback a `es`.
	LangEN = "en"
)

// Secciones del módulo: etiqueta legible que forma parte de `targetLabel`
// (`Portada · <Sección> · <elemento>`, R3-11).
const (
	sectionIdentity = "Identidad"
	sectionAbout    = "Quiénes somos"
	sectionContact  = "Contacto"
	sectionSchedule = "Horario"
	sectionWhatsapp = "WhatsApp"
	sectionSocial   = "Redes"
	sectionImage    = "Imagen"
)

// Catálogo fijo de redes sociales (FR-006/Q5, R3-7). El `CHECK` de la BD es la
// red de seguridad; esta tabla es la fuente del service para validar y para
// resolver los hosts oficiales de cada red.
var socialNetworkHosts = map[string][]string{
	"facebook":  {"facebook.com"},
	"instagram": {"instagram.com"},
	"youtube":   {"youtube.com", "youtu.be"},
	"tiktok":    {"tiktok.com"},
	"spotify":   {"spotify.com"},
}

// Dominios oficiales de los enlaces de grupo de WhatsApp (R3-6).
var whatsappGroupHosts = []string{"chat.whatsapp.com", "wa.me"}

// timePattern es el patrón "HH:MM" en 24 h que exige el contrato y la BD.
var timePattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// horizontalSpace colapsa las secuencias de espacios y tabulaciones, sin tocar
// los saltos de línea (para no destrozar los párrafos de «quiénes somos»).
var horizontalSpace = regexp.MustCompile(`[ \t]+`)

// --- Puerto del repositorio (definido por quien lo consume, arq. R3) ---

// Repository es el puerto de datos que el service de la portada necesita. Los
// errores de PostgreSQL llegan ya traducidos a apperr. Las mutaciones reciben
// las filas de auditoría que deben insertarse en la MISMA transacción
// (variádicas): o contenido + registro, o nada (FR-017).
type Repository interface {
	// Singletons (lecturas del panel incluyen borradores; …Published solo lo
	// publicado, FR-013).
	GetHomeIdentity(ctx context.Context) (Identity, bool, error)
	GetHomeIdentityPublished(ctx context.Context) (Identity, bool, error)
	UpsertHomeIdentity(ctx context.Context, identity Identity, actions ...audit.Action) (Identity, error)
	GetHomeAbout(ctx context.Context) (About, bool, error)
	GetHomeAboutPublished(ctx context.Context) (About, bool, error)
	UpsertHomeAbout(ctx context.Context, about About, actions ...audit.Action) (About, error)
	GetHomeContact(ctx context.Context) (Contact, bool, error)
	GetHomeContactPublished(ctx context.Context) (Contact, bool, error)
	UpsertHomeContact(ctx context.Context, contact Contact, actions ...audit.Action) (Contact, error)

	// Archivos: TRUE solo si el nombre lo referencia contenido publicado
	// (analyze C4).
	IsHomeFilePublished(ctx context.Context, fileName string) (bool, error)

	// Horario.
	InsertHomeService(ctx context.Context, service Service, actions ...audit.Action) (Service, error)
	UpdateHomeService(ctx context.Context, service Service, actions ...audit.Action) (Service, error)
	DeleteHomeService(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	GetHomeServiceByID(ctx context.Context, id uuid.UUID) (Service, bool, error)
	ListHomeServices(ctx context.Context) ([]Service, error)
	ListHomeServicesPublished(ctx context.Context) ([]Service, error)

	// WhatsApp.
	InsertHomeWhatsappChannel(ctx context.Context, channel WhatsappChannel, actions ...audit.Action) (WhatsappChannel, error)
	UpdateHomeWhatsappChannel(ctx context.Context, channel WhatsappChannel, actions ...audit.Action) (WhatsappChannel, error)
	DeleteHomeWhatsappChannel(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	GetHomeWhatsappChannelByID(ctx context.Context, id uuid.UUID) (WhatsappChannel, bool, error)
	ListHomeWhatsappChannels(ctx context.Context) ([]WhatsappChannel, error)
	ListHomeWhatsappChannelsPublished(ctx context.Context) ([]WhatsappChannel, error)

	// Redes.
	InsertHomeSocialLink(ctx context.Context, link SocialLink, actions ...audit.Action) (SocialLink, error)
	UpdateHomeSocialLink(ctx context.Context, link SocialLink, actions ...audit.Action) (SocialLink, error)
	DeleteHomeSocialLink(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error)
	GetHomeSocialLinkByID(ctx context.Context, id uuid.UUID) (SocialLink, bool, error)
	ListHomeSocialLinks(ctx context.Context) ([]SocialLink, error)
	ListHomeSocialLinksPublished(ctx context.Context) ([]SocialLink, error)

	// RecordAction persiste una acción de auditoría best-effort (denegaciones y
	// rechazos): fuera de transacción y sin cambiar la respuesta (R3-11).
	RecordAction(ctx context.Context, action audit.Action) error
}

// ServiceDeps agrupa las dependencias del service para no encadenar una lista
// larga de parámetros en el constructor.
type ServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository Repository
	// Store es el almacén de imágenes (platform/storage). Puede ser nil en
	// pruebas que no ejerciten la descarga de archivos.
	Store storage.Store
	// Logger registra los fallos best-effort; nil usa el logger por defecto.
	Logger *slog.Logger
}

// service implementa la portada pública, la gestión del panel y el registro de
// auditoría del dominio. Una sola instancia satisface las tres
// responsabilidades; los archivos service_public.go, service_admin.go y
// service_audit.go reparten los métodos.
type service struct {
	repository Repository
	store      storage.Store
	logger     *slog.Logger
}

// NewService construye el service de la portada. Si logger es nil se usa el
// logger por defecto.
func NewService(deps ServiceDeps) *service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &service{repository: deps.Repository, store: deps.Store, logger: logger}
}

// --- Normalización ---

// normalizeText aplica trim y colapsa las secuencias de espacios/tabulaciones
// internas a un único espacio. Conserva los saltos de línea y todo el contenido
// (tildes, ñ, ¿¡ y emojis intactos).
func normalizeText(value string) string {
	value = strings.TrimSpace(value)
	return horizontalSpace.ReplaceAllString(value, " ")
}

// normalizeOptional normaliza un campo opcional: `""` o solo espacios pasa a
// nil (analyze I6), nunca a la cadena vacía que los CHECK de longitud rechazan.
func normalizeOptional(value string) *string {
	normalized := normalizeText(value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

// normalizePhone reduce un teléfono a sus dígitos, conservando un `+` inicial
// si lo traía (clave para armar el enlace wa.me sin el `+`).
func normalizePhone(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	if strings.HasPrefix(value, "+") {
		b.WriteByte('+')
	}
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// phoneDigits devuelve solo los dígitos de un teléfono normalizado (destino de
// `https://wa.me/`).
func phoneDigits(normalized string) string {
	return strings.TrimPrefix(normalized, "+")
}

// --- Invariantes compartidas ---

// publicationStateOr traduce el estado de entrada; un valor vacío es `draft`
// (el default del contrato y de la BD).
func publicationStateOr(value string) PublicationState {
	if value == "" {
		return StateDraft
	}
	return PublicationState(value)
}

// stateActionCode es el código de auditoría de un cambio de estado (R3-11.1):
// `home.publish` al publicar y `home.unpublish` al retirar.
func stateActionCode(state PublicationState) string {
	if state == StatePublished {
		return audit.ActionHomePublish
	}
	return audit.ActionHomeUnpublish
}

// validLang indica si `lang` pertenece al registro (es|en).
func validLang(lang string) bool {
	return lang == LangES || lang == LangEN
}

// --- Enlaces ---

// validURLShape comprueba la forma de un enlace con la etiqueta `url` de
// platform/validate (https + host, ≤500). El host concreto lo valida cada
// regla de negocio.
func validURLShape(link string) bool {
	link = strings.TrimSpace(link)
	if link == "" || len(link) > maxURLLength {
		return false
	}
	return validate.Struct(urlRule{URL: link}) == nil
}

// hostAllowed indica si `rawURL` es https y su host es uno de los dominios
// indicados o un subdominio suyo (p. ej. `www.facebook.com`).
func hostAllowed(rawURL string, domains []string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	for _, domain := range domains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

// validSocialNetwork indica si la red pertenece al catálogo fijo (R3-7).
func validSocialNetwork(network string) bool {
	_, ok := socialNetworkHosts[network]
	return ok
}

// urlRule envuelve la etiqueta `url` de platform/validate para reutilizarla
// desde el service (la forma, no el dominio concreto).
type urlRule struct {
	URL string `validate:"required,url,max=500"`
}

// phoneRule envuelve la etiqueta `phone` de F2 (mismo criterio para el destino
// directo de WhatsApp y para el teléfono de contacto, FR-015).
type phoneRule struct {
	Phone string `validate:"required,phone,max=32"`
}

// validPhoneDestination indica si un teléfono cumple el criterio de F2.
func validPhoneDestination(phone string) bool {
	return validate.Struct(phoneRule{Phone: strings.TrimSpace(phone)}) == nil
}

// validWhatsappGroupURL valida la URL de un grupo: forma `url` + host oficial de
// WhatsApp (R3-6).
func validWhatsappGroupURL(rawURL string) bool {
	return validURLShape(rawURL) && hostAllowed(rawURL, whatsappGroupHosts)
}

// validSocialURL valida el enlace de una red: forma `url` + host oficial de esa
// red (R3-7).
func validSocialURL(network, rawURL string) bool {
	domains, ok := socialNetworkHosts[network]
	if !ok {
		return false
	}
	return validURLShape(rawURL) && hostAllowed(rawURL, domains)
}

// --- Auditoría del dominio ---

// contentAction construye una acción de auditoría de contenido (`result`
// success), la unidad que el repository inserta dentro de la transacción de la
// mutación (R3-11).
func contentAction(actorID uuid.UUID, code, label string) audit.Action {
	return audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// contentLabel arma `Portada · <Sección> · <elemento>` (R3-11); sin elemento
// (una escritura de colección o un singleton) queda `Portada · <Sección>`.
func contentLabel(section, element string) string {
	element = normalizeText(element)
	if element == "" {
		return "Portada · " + section
	}
	return "Portada · " + section + " · " + element
}

// recordActionBestEffort persiste una acción que ya va a fallar (rechazo o
// denegación) sin cambiar la respuesta: su fallo se queda en el log (R3-11).
func (s *service) recordActionBestEffort(ctx context.Context, action audit.Action) {
	if err := s.repository.RecordAction(ctx, action); err != nil {
		s.logError(ctx, "no se pudo registrar la acción administrativa", err)
	}
}

// recordContentFailure registra best-effort el rechazo de una mutación del
// módulo (`result='failure'`), con el mismo objetivo que habría tenido la
// acción (R3-11).
func (s *service) recordContentFailure(ctx context.Context, actorID uuid.UUID, code, label string) {
	s.recordActionBestEffort(ctx, audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultFailure,
	})
}

// deleteFile borra best-effort una imagen reemplazada (R3-8): un fallo se queda
// en el log y no cambia el guardado ya realizado.
func (s *service) deleteFile(ctx context.Context, name string) {
	if s.store == nil || name == "" {
		return
	}
	if err := s.store.Delete(ctx, name); err != nil {
		s.logError(ctx, "no se pudo borrar la imagen reemplazada", err)
	}
}

// logError registra un fallo best-effort con el logger de la petición (que
// lleva el request_id) o, si no hay, con el logger base del service.
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

// compile-time check: *repository (repository.go/repository_home.go) must
// satisfy the port the service declares.
var _ Repository = (*repository)(nil)
