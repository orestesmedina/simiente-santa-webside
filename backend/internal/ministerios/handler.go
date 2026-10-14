package ministerios

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

// --- Helpers HTTP compartidos (patrón de F2/F3/F4) ---

// maxRequestBytes acota el cuerpo JSON de una petición (1 MiB): evita leer
// entradas ilimitadas antes de decodificarlas (CWE-400).
const maxRequestBytes = 1 << 20

// decodeAndValidate decodifica el cuerpo JSON de la petición en dto y lo valida
// con platform/validate (CWE-20). Devuelve true si el DTO es utilizable; si no,
// ya escribió el 400 invalid (con `details` por campo cuando la validación lo
// produce) y devuelve false. Rechaza campos desconocidos
// (additionalProperties: false del contrato) y más de un valor JSON.
//
// Todo rechazo pasa por recordRejected: deja la traza y el registro best-effort
// en admin_actions con `result='failure'` (R6-12).
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
// petición (con request_id) cuando está en el contexto y no transporta el cuerpo
// de la petición (FR-026 de F2).
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

// errAlreadyHandled señala que el helper ya escribió la respuesta (p. ej. el 400
// de un DTO inválido): el envoltorio de las operaciones no debe volver a
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

// ministerioIDFromRequest lee el id de la ruta ({id}) y lo convierte a UUID. Un
// valor que no es UUID → 400 invalid con details.id.
func ministerioIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, apperr.Invalid(
			"El identificador del ministerio no es válido",
			apperr.WithDetails(map[string]any{"id": "Debe ser un identificador UUID válido"}),
		)
	}
	return id, nil
}

// listParams lee la paginación del panel (`limit`/`offset`, §8.1.3) y el filtro
// opcional por estado de publicación (FR-012). Un valor inválido → 400 invalid
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
var _ MinisteriosService = (*service)(nil)
