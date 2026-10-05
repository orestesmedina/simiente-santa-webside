package usuarios

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
)

// handler_roles.go implementa la superficie de gestión de roles y del catálogo
// de permisos del contrato OpenAPI: listar, ver, crear, editar y eliminar roles,
// y leer el catálogo de permisos por módulo (FR-014…FR-018, FR-015). Es solo
// HTTP: decodifica y valida, delega en RolesService y responde con el sobre
// uniforme. Las rutas se publican en routes.go bajo la cadena del panel. Un rol
// solo se elimina si ninguna cuenta lo usa y toda mutación exige CSRF.

// RolesService es el puerto que los handlers de roles necesitan del servicio (lo
// define quien lo consume, arq. R3). Lo implementa *roleService.
type RolesService interface {
	// ListRoles devuelve una página de roles con permisos y recuento (FR-017).
	ListRoles(ctx context.Context, params paginate.Params) (RoleList, error)
	// GetRole devuelve la ficha de un rol (FR-017).
	GetRole(ctx context.Context, id uuid.UUID) (RoleItem, error)
	// CreateRole crea un rol con al menos un permiso (FR-014).
	CreateRole(ctx context.Context, actorID uuid.UUID, in RoleCreateInput) (RoleItem, error)
	// UpdateRole edita el nombre y/o los permisos (FR-017/FR-018).
	UpdateRole(ctx context.Context, actorID uuid.UUID, id uuid.UUID, in RoleUpdateInput) (RoleItem, error)
	// DeleteRole elimina un rol sin cuentas asignadas (FR-017).
	DeleteRole(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error
	// ListPermissions devuelve el catálogo fijo de permisos (FR-015).
	ListPermissions(ctx context.Context) (PermissionList, error)
}

// messageInvalidRoleID es el 400 de un identificador de ruta que no es UUID.
const messageInvalidRoleID = "El identificador del rol no es válido"

// ListRoles responde GET /api/v1/admin/roles (FR-017): normaliza limit/offset
// con platform/paginate (P14) y devuelve el sobre RoleList con sus permisos y
// userCount. Códigos: 200; parámetros fuera de rango → 400 invalid.
func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.rolesConfigured(ctx, w) {
		return
	}
	params, err := paginate.FromValues(r.URL.Query())
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	list, err := h.roles.ListRoles(ctx, params)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// CreateRole responde POST /api/v1/admin/roles (FR-014): valida el DTO, delega
// en el servicio con el actor de la sesión y responde 201 con el RoleItem.
// Códigos: 201; DTO inválido, sin permisos o permiso inexistente → 400; nombre
// duplicado normalizado → 409.
func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.rolesConfigured(ctx, w) {
		return
	}

	var in RoleCreateInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	item, err := h.roles.CreateRole(ctx, identity.UserID, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, item)
}

// GetRole responde GET /api/v1/admin/roles/{id} (FR-017): detalle de un rol con
// sus permisos para el formulario de edición.
// Códigos: 200; id no UUID → 400; inexistente → 404.
func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.rolesConfigured(ctx, w) {
		return
	}
	id, err := roleIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	item, err := h.roles.GetRole(ctx, id)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// UpdateRole responde PATCH /api/v1/admin/roles/{id} (FR-017/FR-018): cambia el
// nombre y/o los permisos; los cambios se reflejan de inmediato en las cuentas
// con ese rol. Sin campos → 400; dejar el rol sin permisos → 400; nombre
// duplicado o quitar el último permiso de administración → 409.
// Códigos: 200; id no UUID o DTO inválido → 400; inexistente → 404.
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.rolesConfigured(ctx, w) {
		return
	}
	id, err := roleIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	var in RoleUpdateInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	item, err := h.roles.UpdateRole(ctx, identity.UserID, id, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// DeleteRole responde DELETE /api/v1/admin/roles/{id} (FR-017/US6): elimina un
// rol solo si ninguna cuenta lo tiene asignado; si hay cuentas → 409 con
// details.userCount. Códigos: 200; id no UUID → 400; inexistente → 404; en uso o
// guard anti-bloqueo → 409.
func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.rolesConfigured(ctx, w) {
		return
	}
	id, err := roleIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	if err := h.roles.DeleteRole(ctx, identity.UserID, id); err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, RoleDeletedResponse{Deleted: true})
}

// ListPermissions responde GET /api/v1/admin/permisos (FR-015): el catálogo
// fijo de permisos por módulo con su etiqueta, para armar roles.
// Códigos: 200.
func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.rolesConfigured(ctx, w) {
		return
	}
	list, err := h.roles.ListPermissions(ctx)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// RegisterRoleRoutes publica las rutas de roles y el catálogo de permisos dentro
// de un Registrar ya protegido por la cadena del panel (authn → passwordguard →
// authz(admin_usuarios_roles) → CSRF, T227). No publica nada si el servicio no
// está cableado.
func RegisterRoleRoutes(root httpserver.Registrar, h *Handler) {
	if h == nil || h.roles == nil {
		return
	}
	root.Handle(http.MethodGet, "/roles", h.ListRoles)
	root.Handle(http.MethodPost, "/roles", h.CreateRole)
	root.Handle(http.MethodGet, "/roles/{id}", h.GetRole)
	root.Handle(http.MethodPatch, "/roles/{id}", h.UpdateRole)
	root.Handle(http.MethodDelete, "/roles/{id}", h.DeleteRole)
	root.Handle(http.MethodGet, "/permisos", h.ListPermissions)
}

// rolesConfigured responde 500 si el servicio de roles no está cableado: es un
// error de composición interno, no una entrada inválida.
func (h *Handler) rolesConfigured(ctx context.Context, w http.ResponseWriter) bool {
	if h.roles != nil {
		return true
	}
	httpserver.WriteError(ctx, w, h.logger,
		apperr.Internal(errors.New("gestión de roles: servicio no configurado")))
	return false
}

// roleIDFromRequest lee el id de la ruta ({id}) y lo convierte a UUID. Un valor
// que no es UUID → 400 invalid con details.id.
func roleIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, apperr.Invalid(
			messageInvalidRoleID,
			apperr.WithDetails(map[string]any{"id": "Debe ser un identificador UUID válido"}),
		)
	}
	return id, nil
}
