package agenda

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_admin.go publica la gestión de los módulos desde el panel (T424–T426):
// eventos, actividades y el catálogo de categorías. Es solo HTTP: decodifica y
// valida los DTOs del contrato, delega en el service y responde el DTO del
// panel. Nada de SQL (R4) ni de credenciales en las respuestas (FR-026 de F2).
//
// La cadena (authn → guard → authz(permiso del módulo) → CSRF) la monta
// routes.go (T428) con el permiso del módulo: `eventos` para este grupo,
// `actividades` y `categorias` para los otros dos.

// --- Eventos (FR-010, T424) ---

// ListEventos responde GET /api/v1/admin/eventos (FR-010): listado paginado del
// módulo —borradores y elementos con cualquier fecha incluidos— con filtro
// opcional por estado (FR-013). Cada fila trae sus fechas (con `isPast`),
// `nextDate` **siempre presente** (null cuando el evento no tiene ninguna fecha
// próxima) e `isPast` del elemento (FR-006).
// Códigos: 200 con EventoAdminList; `limit`/`offset` o `publicationState`
// inválidos → 400 con details; sin sesión → 401; sin permiso → 403 (cadena).
func (h *Handler) ListEventos(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.service.ListEventos(ctx, state, params)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, out)
}

// CreateEvento responde POST /api/v1/admin/eventos (FR-010/FR-018): alta de un
// evento con su lista de fechas (1..50) y su estado de publicación.
// Códigos: 201 con EventoAdmin; DTO inválido → 400 con details por campo;
// duplicado del módulo → 409; sin sesión → 401; sin permiso → 403 (cadena).
func (h *Handler) CreateEvento(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in EventoInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateEvento(ctx, actorID, in)
	})
}

// GetEvento responde GET /api/v1/admin/eventos/{id} (FR-010): ficha completa
// para el formulario de edición (ambos idiomas, todas sus fechas y su estado).
// Códigos: 200 con EventoAdmin; id no UUID → 400 con details.id; inexistente →
// 404.
func (h *Handler) GetEvento(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	out, err := h.service.GetEvento(ctx, id)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, out)
}

// UpdateEvento responde PATCH /api/v1/admin/eventos/{id} (FR-010): edición
// parcial (al menos un campo) y/o cambio de estado por elemento; si viene
// `dates`, reemplaza la lista completa (R4-11).
// Códigos: 200 con EventoAdmin; id no UUID, PATCH vacío o datos inválidos → 400;
// inexistente → 404; duplicado del módulo → 409.
func (h *Handler) UpdateEvento(w http.ResponseWriter, r *http.Request) {
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch EventoPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateEvento(ctx, actorID, id, patch)
	})
}

// DeleteEvento responde DELETE /api/v1/admin/eventos/{id} (FR-010): borrado
// físico con sus fechas (la auditoría conserva su etiqueta, R4-10).
// Códigos: 204 sin cuerpo; id no UUID → 400; inexistente → 404.
func (h *Handler) DeleteEvento(w http.ResponseWriter, r *http.Request) {
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteEvento(ctx, actorID, id)
	})
}

// --- Actividades (FR-011, T425) ---

// ListActividades responde GET /api/v1/admin/actividades (FR-011): listado
// paginado del módulo —borradoras y con cualquier fecha incluidas— con filtro
// opcional por estado (FR-013). Cada fila trae su fecha única y su `isPast`
// (FR-006).
// Códigos: 200 con ActividadAdminList; `limit`/`offset` o `publicationState`
// inválidos → 400 con details; sin sesión → 401; sin permiso → 403 (cadena).
func (h *Handler) ListActividades(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.service.ListActividades(ctx, state, params)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, out)
}

// CreateActividad responde POST /api/v1/admin/actividades (FR-011/FR-019): alta
// de una actividad puntual con su fecha única y su estado de publicación.
// Códigos: 201 con ActividadAdmin; DTO inválido → 400 con details por campo;
// duplicado exacto del módulo → 409; sin sesión → 401; sin permiso → 403.
func (h *Handler) CreateActividad(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in ActividadInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateActividad(ctx, actorID, in)
	})
}

// GetActividad responde GET /api/v1/admin/actividades/{id} (FR-011): ficha
// completa para el formulario de edición.
// Códigos: 200 con ActividadAdmin; id no UUID → 400 con details.id; inexistente
// → 404.
func (h *Handler) GetActividad(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	out, err := h.service.GetActividad(ctx, id)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, out)
}

// UpdateActividad responde PATCH /api/v1/admin/actividades/{id} (FR-011): edición
// parcial (al menos un campo) y/o cambio de estado por elemento.
// Códigos: 200 con ActividadAdmin; id no UUID, PATCH vacío o datos inválidos →
// 400; inexistente → 404; duplicado del módulo → 409.
func (h *Handler) UpdateActividad(w http.ResponseWriter, r *http.Request) {
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch ActividadPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateActividad(ctx, actorID, id, patch)
	})
}

// DeleteActividad responde DELETE /api/v1/admin/actividades/{id} (FR-011):
// borrado físico (la auditoría conserva su etiqueta, R4-10).
// Códigos: 204 sin cuerpo; id no UUID → 400; inexistente → 404.
func (h *Handler) DeleteActividad(w http.ResponseWriter, r *http.Request) {
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteActividad(ctx, actorID, id)
	})
}

// --- Catálogo de categorías (Q6/Q7, T426) ---

// ListCategorias responde GET /api/v1/admin/categorias (Q6): el catálogo
// completo —acotado en el servidor a ≤100 y ordenado por `sortOrder, id`— para
// los selectores de los formularios. Sin paginación: es un catálogo pequeño
// (§8.1.3, misma excepción documentada que las colecciones de F3). El catálogo
// no tiene estado de publicación (son etiquetas, no contenido).
// Códigos: 200 con CategoriaAdminList; sin sesión → 401; sin el permiso
// `categorias` → 403 (cadena).
func (h *Handler) ListCategorias(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	out, err := h.service.ListCategorias(ctx)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, out)
}

// CreateCategoria responde POST /api/v1/admin/categorias (Q6): alta en el
// catálogo compartido por eventos y actividades. Exige el permiso `categorias`
// (verificado en el servidor por la cadena; **no** basta `eventos` ni
// `actividades`, FR-012).
// Códigos: 201 con CategoriaAdmin; DTO inválido → 400 con details por campo;
// nombre duplicado → 409; sin permiso → 403.
func (h *Handler) CreateCategoria(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in CategoriaInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateCategoria(ctx, actorID, in)
	})
}

// UpdateCategoria responde PATCH /api/v1/admin/categorias/{id} (Q6): edición
// parcial; renombrar está permitido y el cambio se ve de inmediato en lo
// publicado. Exige el permiso `categorias`.
// Códigos: 200 con CategoriaAdmin; id no UUID, PATCH vacío o datos inválidos →
// 400; inexistente → 404; nombre duplicado → 409.
func (h *Handler) UpdateCategoria(w http.ResponseWriter, r *http.Request) {
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch CategoriaPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateCategoria(ctx, actorID, id, patch)
	})
}

// DeleteCategoria responde DELETE /api/v1/admin/categorias/{id} (Q6/Q7): borrado
// físico del catálogo. Si algún evento o actividad la usa responde `409
// conflict` con el recuento ("está en uso por N elementos", decisión firme de la
// puerta 2): hay que quitarla primero de esos elementos, no se desliga sola.
// Exige el permiso `categorias`.
// Códigos: 204 sin cuerpo; id no UUID → 400; inexistente → 404; en uso → 409.
func (h *Handler) DeleteCategoria(w http.ResponseWriter, r *http.Request) {
	id, err := agendaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteCategoria(ctx, actorID, id)
	})
}
