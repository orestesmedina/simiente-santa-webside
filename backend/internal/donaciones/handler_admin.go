package donaciones

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

// handler_admin.go publica la gestión del módulo desde el panel bajo
// /api/v1/admin/donaciones (T719/T720). Es solo HTTP: decodifica y valida los
// DTOs del contrato, delega en el service y responde el DTO del panel —que sí
// lleva los dos idiomas del titular, la moneda y el estado (dato guardado ≠ dato
// publicado)—. Nada de SQL ni de credenciales en las respuestas. La cadena
// (authn → guard → authz(`donaciones`) → CSRF) la monta routes.go (T721).
//
// Aquí viven también los helpers HTTP compartidos del panel (patrón de
// F2/F3/F4/F6): decodificar y validar el cuerpo, resolver la identidad y los
// preludios de las colecciones.

// messageUnauthenticated es el 401 genérico del dominio. Coincide a propósito con
// el de la cadena de sesión (middleware.authn).
const messageUnauthenticated = "Necesitas iniciar sesión para continuar"

// --- Gestión del panel: métodos (FR-010) ---

// ListDonaciones responde GET /api/v1/admin/donaciones (FR-010): listado paginado
// completo —borradores incluidos— en el orden del módulo (`CASE kind`,
// `sort_order`, `id`: el mismo que ve el visitante, P7-20/US2 esc. 10), con filtro
// opcional `publicationState`.
// Códigos: 200 con DonacionAdminList; parámetros inválidos → 400 con details.
func (h *Handler) ListDonaciones(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	params, state, err := h.listParams(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	list, err := h.service.ListDonaciones(ctx, state, params)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// CreateDonacion responde POST /api/v1/admin/donaciones (FR-010/FR-014): alta con
// validación completa en el servidor. El número se acepta tal como lo escribe el
// equipo y el service lo normaliza y valida (FR-015/FR-016).
// Códigos: 201 con DonacionAdmin; DTO inválido → 400 con details por campo;
// duplicado exacto → 409.
func (h *Handler) CreateDonacion(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in DonacionInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateDonacion(ctx, actorID, in)
	})
}

// GetDonacion responde GET /api/v1/admin/donaciones/{id} (FR-010): ficha completa
// para el formulario de edición (ambos idiomas del titular, la moneda y el estado,
// borradores incluidos).
// Códigos: 200 con DonacionAdmin; id no UUID → 400 con details.id;
// inexistente → 404.
func (h *Handler) GetDonacion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := donacionIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	item, err := h.service.GetDonacion(ctx, id)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// UpdateDonacion responde PATCH /api/v1/admin/donaciones/{id} (FR-010): el handler
// transmite el DTO **parcial** tal cual y el service fusiona los campos presentes
// sobre la fila actual (R7-15): los ausentes no cambian y un `null` explícito borra
// el opcional. `publicationState` cambia el estado por elemento (FR-012) y
// `sortOrder` reordena dentro de su grupo (FR-004).
// Códigos: 200 con DonacionAdmin; DTO inválido → 400 con details; id no UUID →
// 400; inexistente → 404; duplicado exacto con otro método → 409.
func (h *Handler) UpdateDonacion(w http.ResponseWriter, r *http.Request) {
	id, err := donacionIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch DonacionPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateDonacion(ctx, actorID, id, patch)
	})
}

// DeleteDonacion responde DELETE /api/v1/admin/donaciones/{id} (FR-010): borrado
// definitivo (la interfaz pide confirmación antes de llamar).
// Códigos: 204 sin cuerpo; id no UUID → 400; inexistente → 404.
func (h *Handler) DeleteDonacion(w http.ResponseWriter, r *http.Request) {
	id, err := donacionIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteDonacion(ctx, actorID, id)
	})
}

// --- Helpers HTTP compartidos del panel (patrón de F2/F3/F4/F6) ---

// maxRequestBytes acota el cuerpo JSON de una petición (1 MiB): evita leer
// entradas ilimitadas antes de decodificarlas (CWE-400).
const maxRequestBytes = 1 << 20

// decodeAndValidate decodifica el cuerpo JSON de la petición en dto y lo valida
// con platform/validate (CWE-20). Devuelve true si el DTO es utilizable; si no, ya
// escribió el 400 invalid (con `details` por campo cuando la validación lo
// produce) y devuelve false. Rechaza campos desconocidos
// (additionalProperties: false del contrato) y más de un valor JSON.
//
// Todo rechazo pasa por recordRejected: deja la traza y el registro best-effort en
// admin_actions con `result='failure'` (R7-10).
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
// inválido. Es best-effort y nunca cambia la respuesta; usa el logger por petición
// (con request_id) cuando está en el contexto y no transporta el cuerpo de la
// petición (FR-026 de F2).
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

// adminIdentity recupera la identidad resuelta por authn. Las rutas del panel van
// dentro de la cadena (authn → …), así que siempre debería estar; sin ella
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

// updateItem unifica el preludio de los PATCH del panel: exige sesión y servicio,
// delega en fn y responde 200 con el DTO del panel.
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

// donacionIDFromRequest lee el id de la ruta ({id}) y lo convierte a UUID. Un
// valor que no es UUID → 400 invalid con details.id.
func donacionIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, apperr.Invalid(
			"El identificador del método de donación no es válido",
			apperr.WithDetails(map[string]any{"id": "Debe ser un identificador UUID válido"}),
		)
	}
	return id, nil
}

// listParams lee la paginación del panel (`limit`/`offset`, §8.1.3) y el filtro
// opcional por estado de publicación (FR-012). Un valor inválido → 400 invalid con
// el campo en `details` (lo produce platform/paginate o esta validación).
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
