package usuarios

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
)

// handler_auth.go implementa la superficie de acceso del contrato OpenAPI
// (T201): iniciar y cerrar sesión, consultar la sesión actual, cambiar la
// propia contraseña e inicializar el sistema. Es solo HTTP: delega en
// AccessService / SetupService. Las rutas se publican en routes.go.

// setupTokenHeader es la cabecera del contrato que transporta el token de
// despliegue de la inicialización única (P8).
const setupTokenHeader = "X-Setup-Token"

// Mensajes seguros del token de inicialización: nunca revelan el valor esperado
// ni si el sistema está inicializado (RG13).
const (
	messageSetupTokenMissing = "Falta el token de inicialización"
	messageSetupTokenInvalid = "El token de inicialización no es válido"
)

// Login responde POST /api/v1/auth/login (FR-002): decodifica y valida el DTO,
// delega en el service pasando la IP de origen y, en éxito, escribe las cookies
// de sesión y CSRF y el DTO SessionUser. Cualquier error del service
// (credenciales incorrectas, cuenta desactivada, bloqueo) sale por WriteError
// con su sobre del contrato.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	user, cookies, err := h.access.Login(r.Context(), in, clientIP(r))
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}

	writeCookies(w, cookies)
	httpserver.WriteJSON(w, http.StatusOK, user)
}

// Logout responde POST /api/v1/auth/logout (FR-004): cierra la sesión del token
// de la cookie y borra las cookies. Requiere sesión (authn) y CSRF, que monta
// delante el grupo de las rutas de sesión.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token, _ := session.SessionTokenFromRequest(r)

	cookies, err := h.access.Logout(r.Context(), token)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}

	writeCookies(w, cookies)
	httpserver.WriteJSON(w, http.StatusOK, LogoutResponse{LoggedOut: true})
}

// GetSession responde GET /api/v1/auth/session (FR-015): devuelve la identidad
// ya resuelta por authn (permisos vigentes del rol) como DTO SessionUser. Sin
// identidad en el contexto (petición sin la cadena de sesión) responde 401.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	identity, ok := session.IdentityFromContext(r.Context())
	if !ok {
		httpserver.WriteError(r.Context(), w, h.logger, apperr.Unauthenticated(messageUnauthenticated))
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, sessionUserFromIdentity(identity))
}

// ChangePassword responde POST /api/v1/auth/password (FR-020): cambia la
// contraseña de la identidad resuelta. Requiere sesión y CSRF, pero no el guard
// de cambio obligatorio: esta ruta es justamente la que permite resolverlo.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := session.IdentityFromContext(ctx)
	if !ok {
		httpserver.WriteError(ctx, w, h.logger, apperr.Unauthenticated(messageUnauthenticated))
		return
	}

	var in ChangePasswordInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	token, _ := session.SessionTokenFromRequest(r)
	if err := h.access.ChangeMyPassword(ctx, identity, token, in); err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, PasswordChangedResponse{PasswordChanged: true})
}

// Initialize responde POST /api/v1/setup/initialize (FR-007/US2): exige la
// cabecera X-Setup-Token con el valor de BOOTSTRAP_TOKEN, valida el DTO
// InitializeInput y delega en el servicio, que responde 201 con el UserItem de
// la cuenta creada. La comparación del token es en tiempo constante y los
// mensajes no revelan nunca el valor esperado ni si el sistema ya se inicializó
// (RG13).
//
// Códigos: sin cabecera → 401; valor distinto → 403; DTO inválido → 400;
// repetida con cuentas existentes → 409; por encima del umbral de IP → 429 (lo
// pone el rate-limit del grupo, P17).
func (h *Handler) Initialize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.setup == nil {
		// La ruta solo se publica cuando el servicio está cableado (routes.go);
		// si aun así se invoca, es un error de composición interno.
		httpserver.WriteError(ctx, w, h.logger,
			apperr.Internal(errors.New("inicialización: servicio no configurado")))
		return
	}

	provided := strings.TrimSpace(r.Header.Get(setupTokenHeader))
	if provided == "" {
		httpserver.WriteError(ctx, w, h.logger, apperr.Unauthenticated(messageSetupTokenMissing))
		return
	}
	// Tiempo constante: no filtra por dónde difieren el token probado y el
	// esperado. Un token vacío ya respondió 401 arriba, así que un esperado
	// vacío (config ausente) nunca autoriza.
	if subtle.ConstantTimeCompare([]byte(provided), []byte(h.setupToken)) != 1 {
		httpserver.WriteError(ctx, w, h.logger, apperr.Forbidden(messageSetupTokenInvalid))
		return
	}

	var in InitializeInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	item, err := h.setup.Initialize(ctx, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, item)
}

// sessionUserFromIdentity traduce la identidad resuelta por authn al DTO
// SessionUser del contrato. Reutiliza la misma forma que el login para que
// GET /auth/session y el login no puedan divergir.
func sessionUserFromIdentity(identity session.Identity) SessionUser {
	return SessionUser{
		ID:                 identity.UserID.String(),
		Email:              identity.Email,
		FirstName:          identity.FirstName,
		LastName:           identity.LastName,
		Phone:              identity.Phone,
		RoleID:             identity.RoleID.String(),
		RoleName:           identity.RoleName,
		Permissions:        ensureStrings(identity.Permissions),
		MustChangePassword: identity.MustChangePassword,
	}
}
