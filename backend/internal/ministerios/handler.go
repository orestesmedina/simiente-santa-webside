package ministerios

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
)

// handler.go reúne el constructor del Handler y el puerto que los handlers del
// dominio necesitan del service. Los helpers HTTP compartidos (decodificar y
// validar el cuerpo, resolver la identidad y los preludios del panel) viven
// también aquí, con el patrón de F2/F3/F4.
//
// Es una capa fina: no conoce SQL ni las reglas de negocio (constitución §II).
// Decodifica, valida y delega en el puerto `MinisteriosService`; cualquier error
// se deriva a httpserver.WriteError, el único punto de traducción a HTTP
// (arq. R7). La autorización la verifica la cadena de middleware en el servidor
// (authn → guard → authz(`ministerios`) → CSRF), montada en routes.go (T621): los
// handlers no comprueban permisos por su cuenta.

// messageUnauthenticated es el 401 genérico del dominio. Coincide a propósito
// con el de la cadena de sesión (middleware.authn).
const messageUnauthenticated = "Necesitas iniciar sesión para continuar"

// MinisteriosService es el puerto que los handlers del dominio necesitan del
// service. La define quien la consume (arq. R3) y la implementa *service
// (service_public.go, service_admin.go y service_audit.go).
type MinisteriosService interface {
	// GetMinisterios devuelve la sección pública resuelta a `lang` (FR-002…FR-008).
	GetMinisterios(ctx context.Context, lang string) (MinisteriosPublicos, error)
	// GetPortadaMinisterios devuelve el extracto del bloque de la portada
	// (US6/FR-027…FR-029), resuelto a `lang`.
	GetPortadaMinisterios(ctx context.Context, lang string) (PortadaMinisterios, error)

	// Gestión del panel (FR-010).
	CreateMinisterio(ctx context.Context, actorID uuid.UUID, in MinisterioInput) (MinisterioAdmin, error)
	UpdateMinisterio(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch MinisterioPatch) (MinisterioAdmin, error)
	DeleteMinisterio(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error
	GetMinisterio(ctx context.Context, id uuid.UUID) (MinisterioAdmin, error)
	ListMinisterios(ctx context.Context, state *PublicationState, params paginate.Params) (MinisterioAdminList, error)
	UploadMinisterioImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error)

	// RecordRejectedBestEffort registra un rechazo por JSON o DTO inválido sin
	// cambiar la respuesta (R6-12); nunca transporta el cuerpo de la petición.
	RecordRejectedBestEffort(ctx context.Context, rejection Rejection)
}

// Handler expone las operaciones HTTP del dominio ministerios. Es solo HTTP:
// decodifica y valida, delega en el MinisteriosService y responde con el sobre
// uniforme de éxito o deriva cualquier error a httpserver.WriteError.
type Handler struct {
	service        MinisteriosService
	maxUploadBytes int64
	logger         *slog.Logger
}

// HandlerDeps reúne las dependencias del Handler. Es un struct para que el
// cableado (cmd/api) sea explícito.
type HandlerDeps struct {
	// Service es el servicio del dominio. Obligatorio.
	Service MinisteriosService
	// MaxUploadBytes es el tope por subida (UPLOAD_MAX_BYTES). Si es ≤ 0 se usa
	// el valor por defecto de la configuración (8 MB).
	MaxUploadBytes int64
	// Logger registra errores; nil usa el logger por defecto.
	Logger *slog.Logger
}

// NewHandler construye el Handler del dominio ministerios. Si Logger es nil se
// usa el logger por defecto.
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

// defaultMaxUploadBytes es el tope por subida si el cableado no lo indica
// (UPLOAD_MAX_BYTES): 8 MB, el valor por defecto de la configuración.
const defaultMaxUploadBytes = 8388608

// serviceConfigured responde 500 si el servicio no está cableado: es un error de
// composición interno, no una entrada inválida.
func (h *Handler) serviceConfigured(ctx context.Context, w http.ResponseWriter) bool {
	if h.service != nil {
		return true
	}
	httpserver.WriteError(ctx, w, h.logger,
		apperr.Internal(errors.New("ministerios: servicio no configurado")))
	return false
}

// compile-time check: *service must satisfy the port the handlers declare.
var _ MinisteriosService = (*service)(nil)
