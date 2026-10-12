package agenda

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/validate"
)

// handler.go reúne el constructor del Handler y los helpers HTTP comunes a todos
// los handlers del dominio (patrón de la capa de acceso de F2, arq. §5.4):
// decodificación del cuerpo con validación de forma, lectura de la paginación y
// del filtro de estado, y extracción de la identidad resuelta por authn. Los
// handlers concretos viven en handler_public.go, handler_admin.go y
// handler_images.go.
//
// Es una capa fina: no conoce SQL ni las reglas de negocio (constitución §II).
// Decodifica, valida y delega en el puerto `AgendaService`; cualquier error se
// deriva a httpserver.WriteError, el único punto de traducción a HTTP (arq. R7).
// La autorización la verifica la cadena de middleware en el servidor
// (authn → guard → authz(permiso del módulo) → CSRF), montada en routes.go
// (T428): los handlers no comprueban permisos por su cuenta.

// messageUnauthenticated es el 401 genérico del dominio. Coincide a propósito
// con el de la cadena de sesión (middleware.authn).
const messageUnauthenticated = "Necesitas iniciar sesión para continuar"

// maxRequestBytes acota el cuerpo JSON de una petición (1 MiB): evita leer
// entradas ilimitadas antes de decodificarlas (CWE-400).
const maxRequestBytes = 1 << 20

// defaultMaxUploadBytes es el tope por subida si el cableado no lo indica
// (UPLOAD_MAX_BYTES): 8 MB, el valor por defecto de la configuración.
const defaultMaxUploadBytes = 8388608

// AgendaService es el puerto que los handlers del dominio necesitan del service.
// La define quien la consume (arq. R3) y la implementa *service (service.go,
// service_public.go, service_admin.go y service_audit.go).
type AgendaService interface {
	// GetProximamente devuelve «lo que viene» publicado y próximo, resuelto a
	// `lang` (FR-001…FR-008).
	GetProximamente(ctx context.Context, lang string) (PublicProximamente, error)

	// Eventos (FR-010): gestión del panel.
	CreateEvento(ctx context.Context, actorID uuid.UUID, in EventoInput) (EventoAdmin, error)
	UpdateEvento(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch EventoPatch) (EventoAdmin, error)
	DeleteEvento(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error
	GetEvento(ctx context.Context, id uuid.UUID) (EventoAdmin, error)
	ListEventos(ctx context.Context, state *PublicationState, params paginate.Params) (EventoAdminList, error)
	UploadEventoImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error)

	// Actividades (FR-011).
	CreateActividad(ctx context.Context, actorID uuid.UUID, in ActividadInput) (ActividadAdmin, error)
	UpdateActividad(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch ActividadPatch) (ActividadAdmin, error)
	DeleteActividad(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error
	GetActividad(ctx context.Context, id uuid.UUID) (ActividadAdmin, error)
	ListActividades(ctx context.Context, state *PublicationState, params paginate.Params) (ActividadAdminList, error)
	UploadActividadImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error)

	// Catálogo de categorías (Q6/Q7).
	CreateCategoria(ctx context.Context, actorID uuid.UUID, in CategoriaInput) (CategoriaAdmin, error)
	UpdateCategoria(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch CategoriaPatch) (CategoriaAdmin, error)
	DeleteCategoria(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error
	ListCategorias(ctx context.Context) (CategoriaAdminList, error)

	// RecordRejectedBestEffort registra un rechazo por JSON o DTO inválido sin
	// cambiar la respuesta (R4-10); nunca transporta el cuerpo de la petición.
	RecordRejectedBestEffort(ctx context.Context, rejection Rejection)
}

// Handler expone las operaciones HTTP del dominio agenda. Es solo HTTP:
// decodifica y valida, delega en el AgendaService y responde con el sobre
// uniforme de éxito o deriva cualquier error a httpserver.WriteError.
type Handler struct {
	service        AgendaService
	maxUploadBytes int64
	logger         *slog.Logger
}

// HandlerDeps reúne las dependencias del Handler. Es un struct para que el
// cableado (cmd/api) sea explícito.
type HandlerDeps struct {
	// Service es el servicio del dominio. Obligatorio.
	Service AgendaService
	// MaxUploadBytes es el tope por subida (UPLOAD_MAX_BYTES). Si es ≤ 0 se usa
	// el valor por defecto de la configuración (8 MB).
	MaxUploadBytes int64
	// Logger registra errores; nil usa el logger por defecto.
	Logger *slog.Logger
}

// NewHandler construye el Handler del dominio agenda. Si Logger es nil se usa el
// logger por defecto.
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
// Todo rechazo pasa por recordRejected: deja la traza y el registro best-effort
// en admin_actions con `result='failure'` (R4-10).
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

// serviceConfigured responde 500 si el servicio no está cableado: es un error de
// composición interno, no una entrada inválida.
func (h *Handler) serviceConfigured(ctx context.Context, w http.ResponseWriter) bool {
	if h.service != nil {
		return true
	}
	httpserver.WriteError(ctx, w, h.logger,
		apperr.Internal(errors.New("agenda: servicio no configurado")))
	return false
}

// --- Helpers de las colecciones del panel ---

// errAlreadyHandled señala que el helper ya escribió la respuesta (p. ej. el 400
// de un DTO inválido): el envoltorio de las colecciones no debe volver a
// responder.
var errAlreadyHandled = errors.New("respuesta ya escrita")

// createItem unifica el preludio de los POST del panel: exige sesión y servicio,
// delega en fn y responde 201 con el DTO del panel.
func (h *Handler) createItem(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actorID uuid.UUID) (any, error)) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	item, err := fn(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, errAlreadyHandled) {
			return
		}
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, item)
}

// updateItem unifica el preludio de los PATCH del panel: exige sesión y
// servicio, delega en fn y responde 200 con el DTO del panel.
func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actorID uuid.UUID) (any, error)) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	item, err := fn(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, errAlreadyHandled) {
			return
		}
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// deleteItem unifica el preludio de los DELETE del panel: exige sesión y
// servicio, delega en fn y responde 204 sin cuerpo.
func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actorID uuid.UUID) error) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	if err := fn(ctx, identity.UserID); err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// agendaItemIDFromRequest lee el id de la ruta ({id}) y lo convierte a UUID. Un
// valor que no es UUID → 400 invalid con details.id.
func agendaItemIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, apperr.Invalid(
			"El identificador del elemento no es válido",
			apperr.WithDetails(map[string]any{"id": "Debe ser un identificador UUID válido"}),
		)
	}
	return id, nil
}

// listParams lee la paginación del panel (`limit`/`offset`, §8.1.3) y el filtro
// opcional por estado de publicación (FR-013). Un valor inválido → 400 invalid
// con el campo en `details` (lo produce platform/paginate o esta validación).
func (h *Handler) listParams(r *http.Request) (paginate.Params, *PublicationState, error) {
	params, err := paginate.FromValues(r.URL.Query())
	if err != nil {
		return paginate.Params{}, nil, err
	}

	raw := strings.TrimSpace(r.URL.Query().Get("publicationState"))
	if raw == "" {
		return params, nil, nil
	}
	state := PublicationState(raw)
	if !state.Valid() {
		return paginate.Params{}, nil, apperr.Invalid(
			"El filtro de estado no es válido",
			apperr.WithDetails(map[string]any{"publicationState": "solo se admite draft o published"}),
		)
	}
	return params, &state, nil
}

// compile-time check: *service must satisfy the port the handlers declare.
var _ AgendaService = (*service)(nil)
