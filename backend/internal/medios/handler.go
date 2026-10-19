package medios

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
// `MediosService`; cualquier error se deriva a httpserver.WriteError, el único
// punto de traducción a HTTP.
//
// La autorización la verifica la cadena de middleware en el servidor
// (authn → guard de cambio de contraseña → authz(`medios`) → CSRF), montada en
// routes.go (T928): los handlers **no** comprueban permisos por su cuenta. Las
// dos operaciones públicas del contrato (`GET /api/v1/medios` y
// `GET /api/v1/portada/medios`) van sin sesión y son de solo lectura.
//
// Los helpers HTTP compartidos del panel (decodificar y validar el cuerpo,
// resolver la identidad y los preludios) viven junto a las operaciones del panel,
// en `handler_admin.go` (T925), con el patrón de F2/F3/F4/F6/F7.

// MediosService es el puerto que los handlers del dominio necesitan del service.
// La define quien la consume (arq. R3) y la implementa *service (service.go,
// service_public.go, service_admin.go, service_sync.go y service_audit.go).
type MediosService interface {
	// Cara pública (FR-002…FR-009/FR-027): la sección paginada y el extracto de
	// la portada, ya resueltos al idioma pedido.
	GetMedios(ctx context.Context, lang string, limit, offset int) (MediosPublicos, error)
	GetPortadaMedios(ctx context.Context, lang string) (PortadaMedios, error)

	// Gestión del panel (FR-019): listado con `lastSync`, edición de la
	// descripción corta y la visibilidad, la unión por puntero y la resolución
	// de marcas de los dos ejes.
	List(ctx context.Context, filtro FiltroPanel, params paginate.Params) (MedioAdminList, error)
	Update(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch MedioPatch) (MedioAdmin, error)
	Merge(ctx context.Context, actorID, id, otherID uuid.UUID) (MedioAdmin, error)
	ResolveMark(ctx context.Context, actorID, id uuid.UUID, mark MarcaRevision, visibility Visibilidad) (MedioAdmin, error)

	// BuscarNuevos es la única automatización del módulo (FR-011): importa de
	// los dos feeds lo que falte y devuelve el resumen de FR-018.
	BuscarNuevos(ctx context.Context, actorID uuid.UUID) (ResumenBusqueda, error)

	// RecordRejectedBestEffort registra un rechazo por JSON o DTO inválido sin
	// cambiar la respuesta (R7-10); nunca transporta el cuerpo de la petición.
	RecordRejectedBestEffort(ctx context.Context, rejection Rejection)
}

// Handler expone las operaciones HTTP del dominio medios. Es solo HTTP:
// decodifica y valida, delega en el MediosService y responde con el sobre
// uniforme de éxito o deriva cualquier error a httpserver.WriteError.
type Handler struct {
	service MediosService
	logger  *slog.Logger
}

// HandlerDeps reúne las dependencias del Handler. Es un struct para que el
// cableado (cmd/api) sea explícito.
type HandlerDeps struct {
	// Service es el servicio del dominio. Obligatorio.
	Service MediosService
	// Logger registra errores; nil usa el logger por defecto.
	Logger *slog.Logger
}

// NewHandler construye el Handler del dominio medios. Si Logger es nil se usa el
// logger por defecto.
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
		apperr.Internal(errors.New("medios: servicio no configurado")))
	return false
}

// compile-time check: *service must satisfy the port the handlers declare.
var _ MediosService = (*service)(nil)
