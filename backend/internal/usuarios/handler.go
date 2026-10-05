package usuarios

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/validate"
)

// handler.go reúne el constructor del Handler y los helpers HTTP comunes a
// todos los handlers del dominio (arq. §5.4): decodificación del cuerpo,
// validación con platform/validate y extracción de la IP de origen. Los
// handlers concretos viven en handler_auth.go (T228) y en los archivos de las
// tareas siguientes.

// messageUnauthenticated es el 401 genérico de la capa HTTP del dominio. Coincide
// a propósito con el de la cadena de sesión (middleware.authn): el mismo texto
// para «no hay cookie», «sesión expirada» y «cuenta desactivada». Se emite
// cuando el handler no encuentra la identidad ya resuelta en el contexto.
const messageUnauthenticated = "Necesitas iniciar sesión para continuar"

// maxRequestBytes acota el cuerpo de una petición (1 MiB): evita leer entradas
// ilimitadas antes de decodificarlas (CWE-400).
const maxRequestBytes = 1 << 20

// AccessService es la interfaz que los handlers de acceso necesitan del
// servicio. La define quien la consume (arq. R3) y la implementa *authService
// (service_auth.go).
type AccessService interface {
	// Login autentica con correo y contraseña y devuelve la sesión y las
	// cookies que el handler escribe (FR-002, FR-003, FR-006).
	Login(ctx context.Context, in LoginInput, ip string) (SessionUser, []*http.Cookie, error)
	// Logout cierra la sesión del token y devuelve las cookies de borrado
	// (FR-004).
	Logout(ctx context.Context, token string) ([]*http.Cookie, error)
	// ChangeMyPassword cambia la contraseña de la identidad y revoca las demás
	// sesiones (FR-020).
	ChangeMyPassword(ctx context.Context, identity session.Identity, currentToken string, in ChangePasswordInput) error
}

// SetupService es el puerto que el handler de inicialización única necesita del
// servicio (FR-007). Lo implementa *initService (service_init.go).
type SetupService interface {
	// Initialize crea el administrador inicial y devuelve su DTO; repetirla con
	// cuentas existentes devuelve apperr.Conflict (409).
	Initialize(ctx context.Context, in InitializeInput) (UserItem, error)
}

// Handler expone las operaciones del dominio usuarios. Es solo HTTP: decodifica
// y valida, delega en el service y responde con el sobre uniforme de éxito (el
// DTO directo) o deriva cualquier error a httpserver.WriteError, el único punto
// de traducción (arq. R7). No conoce SQL ni las reglas de negocio.
type Handler struct {
	access     AccessService
	setup      SetupService
	users      UsersService
	roles      RolesService
	setupToken string
	logger     *slog.Logger
}

// HandlerDeps reúne las dependencias del Handler del dominio. Es un struct para
// no encadenar parámetros y para que el cableado (cmd/api) sea explícito.
type HandlerDeps struct {
	// Access es el servicio de acceso al panel (login, logout, sesión y cambio
	// de la propia contraseña). Obligatorio.
	Access AccessService
	// Setup es el servicio de inicialización única. Opcional: si es nil, la ruta
	// /api/v1/setup/initialize no se publica.
	Setup SetupService
	// Users es el servicio de gestión de cuentas. Opcional: si es nil, las rutas
	// /api/v1/admin/usuarios* no se publican (T234).
	Users UsersService
	// Roles es el servicio de gestión de roles y del catálogo de permisos.
	// Opcional: si es nil, las rutas /api/v1/admin/roles* y /api/v1/admin/permisos
	// no se publican (T237).
	Roles RolesService
	// SetupToken es el valor esperado de la cabecera X-Setup-Token
	// (BOOTSTRAP_TOKEN). Nunca se registra ni se devuelve (RG13).
	SetupToken string
	// Logger registra errores; nil usa el logger por defecto.
	Logger *slog.Logger
}

// NewHandler construye el Handler del dominio usuarios. Si Logger es nil se usa
// el logger por defecto.
func NewHandler(deps HandlerDeps) *Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		access:     deps.Access,
		setup:      deps.Setup,
		users:      deps.Users,
		roles:      deps.Roles,
		setupToken: deps.SetupToken,
		logger:     logger,
	}
}

// decodeAndValidate decodifica el cuerpo JSON de la petición en dto y lo valida
// con platform/validate (CWE-20). Devuelve true si el DTO es utilizable; si no,
// ya escribió el 400 invalid (con `details` por campo cuando la validación lo
// produce) y devuelve false. Se rechazan campos desconocidos
// (additionalProperties: false del contrato) y más de un valor JSON.
//
// Todo rechazo pasa por recordRejected: es el punto de escritura de P20 para
// «JSON inválido o DTO no válido». En T228 deja la traza con el request_id; T239
// conecta el registro duradero en admin_actions (best-effort, nunca cambia la
// respuesta — R23).
func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dto any) bool {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dto); err != nil {
		h.rejectInvalid(ctx, w, r, err)
		return false
	}
	// Un cuerpo con un segundo valor JSON es una entrada mal formada.
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
// inválido (P20). Es best-effort y nunca cambia la respuesta; usa el logger por
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
}

// writeCookies escribe las cookies que devuelve el service (sesión y CSRF en el
// login; borrado en el logout).
func writeCookies(w http.ResponseWriter, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
	}
}

// clientIP extrae la IP de origen del intento (sin puerto) para la auditoría y
// el bloqueo de FR-006. No usa X-Forwarded-For: sin un proxy de confianza esa
// cabecera es falsificable (misma decisión que middleware.ratelimit).
func clientIP(r *http.Request) string {
	addr := strings.TrimSpace(r.RemoteAddr)
	if addr == "" {
		return "unknown"
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
