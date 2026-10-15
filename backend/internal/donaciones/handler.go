package donaciones

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
// dominio necesitan del service. Es una capa fina: no conoce SQL ni las reglas
// de negocio (constitución §II). Decodifica, valida y delega en el puerto
// `DonacionesService`; cualquier error se deriva a httpserver.WriteError, el
// único punto de traducción a HTTP (arq. R7). La autorización la verifica la
// cadena de middleware en el servidor (authn → guard → authz(`donaciones`) →
// CSRF), montada en routes.go (T721): los handlers no comprueban permisos por su
// cuenta.
//
// Los helpers HTTP compartidos del panel (decodificar y validar el cuerpo,
// resolver la identidad y los preludios del panel) viven junto a las operaciones
// del panel, en `handler_admin.go` (T719), con el patrón de F2/F3/F4/F6.

// DonacionesService es el puerto que los handlers del dominio necesitan del
// service. La define quien la consume (arq. R3) y la implementa *service
// (service_public.go, service_admin.go y service_audit.go).
type DonacionesService interface {
	// GetOfrendas devuelve la sección pública resuelta a `lang` (FR-002…FR-006).
	GetOfrendas(ctx context.Context, lang string) (OfrendasPublicas, error)
	// GetPortadaOfrendas devuelve el extracto del bloque de la portada
	// (US6/FR-022), resuelto a `lang`.
	GetPortadaOfrendas(ctx context.Context, lang string) (PortadaOfrendas, error)

	// Gestión del panel (FR-010).
	CreateDonacion(ctx context.Context, actorID uuid.UUID, in DonacionInput) (DonacionAdmin, error)
	UpdateDonacion(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch DonacionPatch) (DonacionAdmin, error)
	DeleteDonacion(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error
	GetDonacion(ctx context.Context, id uuid.UUID) (DonacionAdmin, error)
	ListDonaciones(ctx context.Context, state *PublicationState, params paginate.Params) (DonacionAdminList, error)

	// Textos del panel (FR-021): singleton por clave, sin borrado (P7-17).
	GetTextos(ctx context.Context) (DonacionTextosAdmin, error)
	PutTexto(ctx context.Context, actorID uuid.UUID, key TextKey, in DonacionTextoInput) (DonacionTextoAdmin, error)

	// RecordRejectedBestEffort registra un rechazo por JSON o DTO inválido sin
	// cambiar la respuesta (R7-10); nunca transporta el cuerpo de la petición.
	RecordRejectedBestEffort(ctx context.Context, rejection Rejection)
}

// Handler expone las operaciones HTTP del dominio donaciones. Es solo HTTP:
// decodifica y valida, delega en el DonacionesService y responde con el sobre
// uniforme de éxito o deriva cualquier error a httpserver.WriteError.
type Handler struct {
	service DonacionesService
	logger  *slog.Logger
}

// HandlerDeps reúne las dependencias del Handler. Es un struct para que el
// cableado (cmd/api) sea explícito.
type HandlerDeps struct {
	// Service es el servicio del dominio. Obligatorio.
	Service DonacionesService
	// Logger registra errores; nil usa el logger por defecto.
	Logger *slog.Logger
}

// NewHandler construye el Handler del dominio donaciones. Si Logger es nil se usa
// el logger por defecto.
func NewHandler(deps HandlerDeps) *Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{service: deps.Service, logger: logger}
}

// serviceConfigured responde 500 si el servicio no está cableado: es un error de
// composición interno, no una entrada inválida.
func (h *Handler) serviceConfigured(ctx context.Context, w http.ResponseWriter) bool {
	if h.service != nil {
		return true
	}
	httpserver.WriteError(ctx, w, h.logger,
		apperr.Internal(errors.New("donaciones: servicio no configurado")))
	return false
}

// compile-time check: *service must satisfy the port the handlers declare.
var _ DonacionesService = (*service)(nil)
