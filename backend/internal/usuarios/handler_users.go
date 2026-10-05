package usuarios

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/session"
)

// handler_users.go implementa la superficie de gestión de cuentas del contrato
// OpenAPI: listar, ver, crear, editar (datos/rol/estado) y restablecer la
// contraseña de una cuenta (FR-009…FR-013, FR-019, FR-021). Es solo HTTP:
// decodifica y valida, delega en UsersService y responde con el sobre uniforme.
// Las rutas se publican en routes.go bajo la cadena del panel. NO existe ninguna
// operación de borrado de cuentas: retirar el acceso es desactivar (FR-013).

// UsersService es el puerto que los handlers de cuentas necesitan del servicio
// (lo define quien lo consume, arq. R3). Lo implementa *userService.
type UsersService interface {
	// ListUsers devuelve una página del listado de cuentas (FR-019/FR-021).
	ListUsers(ctx context.Context, params paginate.Params) (UserList, error)
	// GetUser devuelve la ficha de una cuenta (FR-019/FR-021).
	GetUser(ctx context.Context, id uuid.UUID) (UserItem, error)
	// CreateUser crea una cuenta activa con rol y contraseña inicial (FR-009).
	CreateUser(ctx context.Context, actorID uuid.UUID, in CreateUserInput) (UserItem, error)
	// UpdateUser edita datos, rol y/o estado (FR-011).
	UpdateUser(ctx context.Context, actorID uuid.UUID, id uuid.UUID, in UpdateUserInput) (UserItem, error)
	// ResetUserPassword define una contraseña nueva y revoca las sesiones
	// (FR-010/R17).
	ResetUserPassword(ctx context.Context, actorID uuid.UUID, id uuid.UUID, in ResetPasswordInput) error
}

// messageInvalidUserID es el 400 de un identificador de ruta que no es UUID.
const messageInvalidUserID = "El identificador de la cuenta no es válido"

// ListUsers responde GET /api/v1/admin/usuarios (FR-019/FR-021): normaliza
// limit/offset con platform/paginate (P14) y devuelve el sobre UserList.
// Códigos: 200; parámetros fuera de rango → 400 invalid.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.usersConfigured(ctx, w) {
		return
	}
	params, err := paginate.FromValues(r.URL.Query())
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	list, err := h.users.ListUsers(ctx, params)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// CreateUser responde POST /api/v1/admin/usuarios (FR-009): valida el DTO,
// delega en el servicio con el actor de la sesión y responde 201 con el
// UserItem. Nunca devuelve la contraseña ni su hash (FR-003).
// Códigos: 201; DTO inválido → 400; correo duplicado o guard → 409.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.usersConfigured(ctx, w) {
		return
	}

	var in CreateUserInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	item, err := h.users.CreateUser(ctx, identity.UserID, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, item)
}

// GetUser responde GET /api/v1/admin/usuarios/{id} (FR-019/FR-021): la ficha
// incluye el último acceso; una cuenta que nunca entró responde
// `lastLoginAt: null` y `lastLoginIp: null` (US8 esc. 6).
// Códigos: 200; id no UUID → 400; inexistente → 404.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.usersConfigured(ctx, w) {
		return
	}
	id, err := userIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	item, err := h.users.GetUser(ctx, id)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// UpdateUser responde PATCH /api/v1/admin/usuarios/{id} (FR-011): edita los
// datos, el rol (el nuevo reemplaza al anterior) y/o el estado. Desactivar
// corta el acceso y revoca las sesiones (FR-012). Ninguna combinación puede
// dejar el panel sin administración → 409 (FR-008).
// Códigos: 200; id no UUID o DTO inválido → 400; inexistente → 404; anti-bloqueo
// o correo duplicado → 409.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.usersConfigured(ctx, w) {
		return
	}
	id, err := userIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	var in UpdateUserInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	item, err := h.users.UpdateUser(ctx, identity.UserID, id, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// ResetUserPassword responde POST /api/v1/admin/usuarios/{id}/password
// (FR-010/US7 esc. 5): define la contraseña nueva, marca su cambio obligatorio y
// revoca las sesiones. Nunca se devuelve ni se registra la contraseña (FR-026).
// Códigos: 200; id no UUID o política incumplida → 400; inexistente → 404.
func (h *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.usersConfigured(ctx, w) {
		return
	}
	id, err := userIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	var in ResetPasswordInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	if err := h.users.ResetUserPassword(ctx, identity.UserID, id, in); err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, PasswordResetResponse{PasswordReset: true})
}

// RegisterUserRoutes publica las rutas de gestión de cuentas dentro de un
// Registrar ya protegido por la cadena del panel (authn → passwordguard →
// authz(admin_usuarios_roles) → CSRF, T227). No publica ninguna ruta de borrado:
// las cuentas nunca se eliminan (FR-013).
func RegisterUserRoutes(root httpserver.Registrar, h *Handler) {
	if h == nil || h.users == nil {
		return
	}
	root.Handle(http.MethodGet, "/usuarios", h.ListUsers)
	root.Handle(http.MethodPost, "/usuarios", h.CreateUser)
	root.Handle(http.MethodGet, "/usuarios/{id}", h.GetUser)
	root.Handle(http.MethodPatch, "/usuarios/{id}", h.UpdateUser)
	root.Handle(http.MethodPost, "/usuarios/{id}/password", h.ResetUserPassword)
}

// usersConfigured responde 500 si el servicio de cuentas no está cableado: es un
// error de composición interno, no una entrada inválida.
func (h *Handler) usersConfigured(ctx context.Context, w http.ResponseWriter) bool {
	if h.users != nil {
		return true
	}
	httpserver.WriteError(ctx, w, h.logger,
		apperr.Internal(errors.New("gestión de cuentas: servicio no configurado")))
	return false
}

// adminIdentity recupera la identidad resuelta por authn. Las rutas van dentro
// de la cadena del panel, así que siempre debería estar; sin ella responde 401
// como red de seguridad.
func (h *Handler) adminIdentity(ctx context.Context, w http.ResponseWriter) (session.Identity, bool) {
	identity, ok := session.IdentityFromContext(ctx)
	if !ok {
		httpserver.WriteError(ctx, w, h.logger, apperr.Unauthenticated(messageUnauthenticated))
		return session.Identity{}, false
	}
	return identity, true
}

// userIDFromRequest lee el id de la ruta ({id}) y lo convierte a UUID. Un valor
// que no es UUID → 400 invalid con details.id.
func userIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, apperr.Invalid(
			messageInvalidUserID,
			apperr.WithDetails(map[string]any{"id": "Debe ser un identificador UUID válido"}),
		)
	}
	return id, nil
}
