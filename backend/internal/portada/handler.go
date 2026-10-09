package portada

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/validate"
)

// handler.go reúne el constructor del Handler y los helpers HTTP comunes a
// todos los handlers del dominio (patrón de la capa de acceso de F2, arq.
// §5.4): decodificación del cuerpo con validación de forma y extracción de la
// identidad resuelta por authn. Los handlers concretos viven en
// handler_public.go, handler_admin.go, handler_images.go y handler_media.go.
//
// Es una capa fina: no conoce SQL ni las reglas de negocio (constitución §II).
// Decodifica, valida y delega en el puerto `Service`; cualquier error se deriva
// a httpserver.WriteError, el único punto de traducción a HTTP (arq. R7).

// messageUnauthenticated es el 401 genérico del dominio. Coincide a propósito
// con el de la cadena de sesión (middleware.authn).
const messageUnauthenticated = "Necesitas iniciar sesión para continuar"

// maxRequestBytes acota el cuerpo JSON de una petición (1 MiB): evita leer
// entradas ilimitadas antes de decodificarlas (CWE-400).
const maxRequestBytes = 1 << 20

// defaultMaxUploadBytes es el tope por subida si el cableado no lo indica
// (UPLOAD_MAX_BYTES, T312): 8 MB, el valor por defecto de la configuración.
const defaultMaxUploadBytes = 8388608

// multipartOverhead es el margen que se suma al tope de la imagen para el
// envoltorio multipart antes de aplicar http.MaxBytesReader: el tamaño real del
// archivo se comprueba después con el límite exacto.
const multipartOverhead = 1 << 20

// Service es el puerto que los handlers de la portada necesitan del servicio.
// La define quien la consume (arq. R3) y la implementa *service (service.go,
// service_public.go, service_admin.go y service_audit.go). Crece con cada
// responsabilidad: en T322 expone la lectura pública y el registro best-effort
// de los rechazos.
type PortadaService interface {
	// GetPortada devuelve la portada publicada resuelta al idioma pedido
	// (FR-001/FR-008/FR-009).
	GetPortada(ctx context.Context, lang string) (PublicPortada, error)
	// OpenMedia aplica la política de descarga pública de imágenes (analyze
	// C4/M6) y abre el archivo del almacén.
	OpenMedia(ctx context.Context, fileName string) (io.ReadCloser, error)
	// UploadImage guarda una imagen subida desde el panel y la audita con
	// `home.image.upload` (fail-closed, analyze I8).
	UploadImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error)
	// SaveIdentity, SaveAbout y SaveContact reemplazan los singletons del
	// módulo (FR-002/FR-003/FR-007) con auditoría transaccional (FR-017).
	SaveIdentity(ctx context.Context, actorID uuid.UUID, in IdentityInput) (IdentityAdmin, error)
	SaveAbout(ctx context.Context, actorID uuid.UUID, in AboutInput) (AboutAdmin, error)
	SaveContact(ctx context.Context, actorID uuid.UUID, in ContactInput) (ContactAdmin, error)
	// RecordRejectedBestEffort registra un rechazo por JSON o DTO inválido sin
	// cambiar la respuesta (R3-11); nunca transporta el cuerpo de la petición.
	RecordRejectedBestEffort(ctx context.Context, rejection Rejection)
}

// Handler expone las operaciones HTTP del dominio portada. Es solo HTTP:
// decodifica y valida, delega en el Service y responde con el sobre uniforme de
// éxito o deriva cualquier error a httpserver.WriteError.
type Handler struct {
	service        PortadaService
	maxUploadBytes int64
	logger         *slog.Logger
}

// HandlerDeps reúne las dependencias del Handler. Es un struct para que el
// cableado (cmd/api) sea explícito.
type HandlerDeps struct {
	// Service es el servicio del dominio. Obligatorio.
	Service PortadaService
	// MaxUploadBytes es el tope por subida (UPLOAD_MAX_BYTES, T312). Si es ≤ 0
	// se usa el valor por defecto de la configuración (8 MB).
	MaxUploadBytes int64
	// Logger registra errores; nil usa el logger por defecto.
	Logger *slog.Logger
}

// NewHandler construye el Handler del dominio portada. Si Logger es nil se usa
// el logger por defecto.
func NewHandler(deps HandlerDeps) *Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	maxUploadBytes := deps.MaxUploadBytes
	if maxUploadBytes <= 0 {
		maxUploadBytes = defaultMaxUploadBytes
	}
	return &Handler{service: deps.Service, maxUploadBytes: maxUploadBytes, logger: logger}
}

// decodeAndValidate decodifica el cuerpo JSON de la petición en dto y lo valida
// con platform/validate (CWE-20). Devuelve true si el DTO es utilizable; si no,
// ya escribió el 400 invalid (con `details` por campo cuando la validación lo
// produce) y devuelve false. Rechaza campos desconocidos
// (additionalProperties: false del contrato) y más de un valor JSON.
//
// Todo rechazo pasa por recordRejected: deja la traza y, desde T321, el
// registro best-effort en admin_actions con `result='failure'` (nunca cambia la
// respuesta — R3-11).
func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dto any) bool {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dto); err != nil {
		h.rejectInvalid(ctx, w, r, err)
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.rejectInvalid(ctx, w, r, errors.New("el cuerpo debe contener un único objeto JSON"))
		return false
	}

	if err := validate.Struct(dto); err != nil {
		h.recordRejected(ctx, r)
		httpserver.WriteError(ctx, w, h.logger, err)
		return false
	}
	return true
}

// rejectInvalid responde el 400 del contrato por un cuerpo que no se pudo
// decodificar o que traía basura tras el objeto JSON.
func (h *Handler) rejectInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, cause error) {
	h.recordRejected(ctx, r)
	httpserver.WriteError(ctx, w, h.logger, apperr.Invalid(validate.MessageInvalid, apperr.WithCause(cause)))
}

// recordRejected deja constancia de una petición rechazada por JSON o DTO
// inválido. Es best-effort y nunca cambia la respuesta; usa el logger por
// petición (con request_id) cuando está en el contexto.
func (h *Handler) recordRejected(ctx context.Context, r *http.Request) {
	logger := httpserver.RequestLoggerFromContext(ctx)
	if logger == nil {
		logger = h.logger
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Warn("petición rechazada: cuerpo o datos inválidos",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	if h.service == nil {
		return
	}
	h.service.RecordRejectedBestEffort(ctx, Rejection{
		ActorUserID: identityUserID(ctx),
		Method:      r.Method,
		Path:        r.URL.Path,
	})
}

// adminIdentity recupera la identidad resuelta por authn. Las rutas del panel
// van dentro de la cadena (authn → …), así que siempre debería estar; sin ella
// responde 401 como red de seguridad.
func (h *Handler) adminIdentity(ctx context.Context, w http.ResponseWriter) (session.Identity, bool) {
	identity, ok := session.IdentityFromContext(ctx)
	if !ok {
		httpserver.WriteError(ctx, w, h.logger, apperr.Unauthenticated(messageUnauthenticated))
		return session.Identity{}, false
	}
	return identity, true
}

// identityUserID devuelve el id de la identidad autenticada, o nil en las rutas
// públicas.
func identityUserID(ctx context.Context) *uuid.UUID {
	identity, ok := session.IdentityFromContext(ctx)
	if !ok {
		return nil
	}
	id := identity.UserID
	return &id
}

// serviceConfigured responde 500 si el servicio no está cableado: es un error
// de composición interno, no una entrada inválida.
func (h *Handler) serviceConfigured(ctx context.Context, w http.ResponseWriter) bool {
	if h.service != nil {
		return true
	}
	httpserver.WriteError(ctx, w, h.logger,
		apperr.Internal(errors.New("portada: servicio no configurado")))
	return false
}

// compile-time check: *service must satisfy the port the handlers declare.
var _ PortadaService = (*service)(nil)
