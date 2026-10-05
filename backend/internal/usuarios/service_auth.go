package usuarios

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/password"
	"simiente-santa/backend/internal/platform/session"
)

// service_auth.go implementa el acceso al panel: iniciar y cerrar sesión,
// resolver la identidad de una petición y el cambio de la propia contraseña
// (FR-002…FR-006, FR-012, FR-018, FR-020/FR-022). Vive en el dominio porque
// conoce las reglas de negocio; la sesión, el hash y las cookies son plumbing
// de platform (session/password). No conoce pgx ni SQL: depende de una interfaz
// del repositorio (arq. R3) y de puertos del mismo estilo para el Store de
// sesiones, los contadores de intentos (FR-006) y el registro de accesos.

// Constantes de FR-006: la SEMÁNTICA (cuántos fallos, cuánto dura el bloqueo)
// vive en el dominio; el mecanismo (INCR/TTL/bandera) en platform/session
// (data-model.md). El 5.º fallo responde el 401 genérico y crea el bloqueo; el
// 429 aplica desde el 6.º intento.
const (
	// MaxFailedAttempts es el número de fallos que activa el bloqueo.
	MaxFailedAttempts = 5
	// LockoutDuration es la duración del bloqueo temporal.
	LockoutDuration = 15 * time.Minute
)

// dummyPasswordHash es un hash bcrypt válido (cost 12) de una contraseña que
// nadie conoce. Cuando el correo no corresponde a ninguna cuenta se verifica
// contra él para gastar el mismo tiempo que una verificación real y no revelar
// la existencia por el tiempo de respuesta (R18, FR-003/SC-008).
const dummyPasswordHash = "$2a$12$ZQKpuNJ4cr1wjM88iA/W9uukm7AcpYvEh.Nd3wm0wl5P2gd86n32W"

// Mensajes seguros para el cliente (ux.md §7, contrato OpenAPI).
const (
	// loginFailureMessage es el 401 GENÉRICO: idéntico para credenciales
	// incorrectas y para un correo inexistente (FR-003/SC-008).
	loginFailureMessage = "Correo o contraseña incorrectos"
	// accountInactiveMessage acompaña al 403 de una cuenta desactivada con
	// credenciales correctas (US1 esc. 3).
	accountInactiveMessage = "Ese acceso está desactivado. Pide a un administrador de la iglesia que lo reactive para poder entrar"
	// lockoutMessage acompaña al 429 del bloqueo temporal (FR-006); es idéntico
	// exista o no la cuenta.
	lockoutMessage = "Demasiados intentos fallidos. El acceso queda bloqueado temporalmente durante 15 minutos"
	// currentPasswordDetail es la clave de Details del error de contraseña actual
	// incorrecta (el campo del contrato al que corresponde).
	currentPasswordDetail = "currentPassword"
)

// ErrAccountInactive indica que la cuenta de una identidad que se resuelve ya
// no está activa: FR-012 exige cortar el acceso de inmediato, también con una
// sesión abierta. authn lo traduce a 401.
var ErrAccountInactive = errors.New("usuarios: la cuenta está desactivada")

// AuthRepository es el puerto de datos que el servicio de acceso necesita del
// repositorio (lo define quien lo consume, arq. R3). Los errores de PostgreSQL
// ya llegan traducidos a apperr (NotFound, etc.).
type AuthRepository interface {
	GetUserAuthByEmail(ctx context.Context, email string) (UserAuth, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (Role, error)
	UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	SetUserMustChangePassword(ctx context.Context, id uuid.UUID, must bool) error
	RecordLoginSuccess(ctx context.Context, userID uuid.UUID, ip string) (LoginEvent, error)
}

// LoginThrottle es el mecanismo de los contadores de FR-006. Lo implementa
// *session.Throttle; el dominio pone las constantes al construirlo. Las claves
// se indexan por el identificador normalizado exista o no la cuenta, de modo
// que el comportamiento no revela existencia.
type LoginThrottle interface {
	Blocked(ctx context.Context, identifier string) (blocked bool, retryAfter time.Duration, err error)
	RegisterFailure(ctx context.Context, identifier string) (int64, error)
	Reset(ctx context.Context, identifier string) error
}

// LoginEventRecorder es el punto de escritura best-effort de los intentos que
// fallan (FR-022/R23). Lo implementa el servicio de auditoría; su fallo nunca
// cambia la respuesta que ve la persona. El éxito NO pasa por aquí: usa
// RecordLoginSuccess (fail-closed, en la misma transacción que last_login_*).
type LoginEventRecorder interface {
	RecordLoginEventBestEffort(ctx context.Context, event audit.Event)
}

// AuthServiceDeps agrupa las dependencias del servicio de acceso para no
// encadenar una lista larga de parámetros en el constructor.
type AuthServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository AuthRepository
	// Sessions gestiona las sesiones (Redis): crear, revocar y resolver.
	Sessions session.Store
	// Throttle es el mecanismo de los contadores de FR-006.
	Throttle LoginThrottle
	// LoginEvents registra best-effort los intentos fallidos (FR-022).
	LoginEvents LoginEventRecorder
	// Cookies son los atributos de las cookies de sesión y CSRF (Max-Age = vida
	// absoluta, Secure según el entorno).
	Cookies session.CookieConfig
	// CSRFSecret firma la cookie CSRF de doble envío (P10).
	CSRFSecret string
	// Logger registra los fallos best-effort que no cambian la respuesta.
	Logger *slog.Logger
}

// authService implementa el acceso al panel. Es el Resolver que consume authn
// (session.Resolver): resuelve la identidad con los permisos vigentes de su rol
// en cada petición (FR-018/SC-009).
type authService struct {
	repository  AuthRepository
	sessions    session.Store
	throttle    LoginThrottle
	loginEvents LoginEventRecorder
	cookies     session.CookieConfig
	csrfSecret  string
	logger      *slog.Logger
}

// NewAuthService construye el servicio de acceso. Si logger es nil se usa el
// logger por defecto.
func NewAuthService(deps AuthServiceDeps) *authService {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &authService{
		repository:  deps.Repository,
		sessions:    deps.Sessions,
		throttle:    deps.Throttle,
		loginEvents: deps.LoginEvents,
		cookies:     deps.Cookies,
		csrfSecret:  deps.CSRFSecret,
		logger:      logger,
	}
}

// Login autentica con correo y contraseña (FR-002/US1). Normaliza el correo
// (Q5) y aplica, en este orden: bloqueo FR-006, verificación bcrypt (con
// verificación dummy si la cuenta no existe, R18), rechazo de la cuenta
// inactiva y, en éxito, limpieza del contador, registro del acceso —sin
// registro, sin acceso, FR-021/FR-022—, creación de la sesión en Redis y
// emisión de las cookies. Devuelve el DTO de la sesión y las cookies que el
// handler escribe.
func (s *authService) Login(ctx context.Context, in LoginInput, ip string) (SessionUser, []*http.Cookie, error) {
	identifier := normalizeEmail(in.Email)

	blocked, retryAfter, err := s.throttle.Blocked(ctx, identifier)
	if err != nil {
		return SessionUser{}, nil, fmt.Errorf("comprobar bloqueo de intentos: %w", err)
	}
	if blocked {
		// El intento durante un bloqueo también se registra (Edge Case FR-006/
		// FR-022). No se toca la cuenta para no alargar el bloqueo ni verificarla.
		s.recordFailure(ctx, nil, ip)
		return SessionUser{}, nil, apperr.RateLimited(
			lockoutMessage,
			apperr.WithRetryAfter(retryAfterSeconds(retryAfter)),
		)
	}

	auth, err := s.repository.GetUserAuthByEmail(ctx, identifier)
	if err != nil {
		if isNotFound(err) {
			// Verificación dummy: mismo coste que una real aunque la cuenta no
			// exista (R18). El resultado se descarta.
			_ = password.Verify(dummyPasswordHash, in.Password)
			s.recordFailure(ctx, nil, ip)
			return SessionUser{}, nil, loginFailureError()
		}
		return SessionUser{}, nil, fmt.Errorf("buscar credenciales: %w", err)
	}

	if err := password.Verify(auth.PasswordHash, in.Password); err != nil {
		// Cada fallo incrementa el contador; el 5.º crea el bloqueo pero aún
		// responde el 401 genérico (FR-006).
		s.registerFailure(ctx, identifier)
		s.recordFailure(ctx, &auth.ID, ip)
		return SessionUser{}, nil, loginFailureError()
	}

	if !auth.IsActive {
		// Credenciales correctas pero acceso desactivado (US1 esc. 3). Es un
		// inicio de sesión fallido y cuenta para FR-006.
		s.registerFailure(ctx, identifier)
		s.recordFailure(ctx, &auth.ID, ip)
		return SessionUser{}, nil, apperr.Forbidden(accountInactiveMessage)
	}

	// Éxito. Se relee la ficha porque UserAuth no trae el teléfono que exige el
	// DTO de sesión.
	user, err := s.repository.GetUserByID(ctx, auth.ID)
	if err != nil {
		return SessionUser{}, nil, fmt.Errorf("leer la cuenta: %w", err)
	}

	// Limpiar el contador es best-effort: un fallo de Redis no debe impedir el
	// acceso a quien ya demostró sus credenciales.
	if err := s.throttle.Reset(ctx, identifier); err != nil {
		s.logError(ctx, "no se pudo limpiar el contador de intentos", err)
	}

	// Fail-closed: sin registro del acceso, no hay sesión (R23).
	if _, err := s.repository.RecordLoginSuccess(ctx, auth.ID, ip); err != nil {
		return SessionUser{}, nil, fmt.Errorf("registrar el acceso: %w", err)
	}

	token, err := s.sessions.Create(ctx, auth.ID)
	if err != nil {
		return SessionUser{}, nil, fmt.Errorf("crear sesión: %w", err)
	}
	csrf, err := session.NewCSRFToken(s.csrfSecret)
	if err != nil {
		return SessionUser{}, nil, fmt.Errorf("generar token CSRF: %w", err)
	}

	return SessionUserFrom(user, auth.Permissions), s.sessionCookies(token, csrf), nil
}

// Logout termina la sesión del token (FR-004) y devuelve las cookies de
// borrado. Si el token viene vacío solo limpia las cookies.
func (s *authService) Logout(ctx context.Context, token string) ([]*http.Cookie, error) {
	if token != "" {
		if err := s.sessions.Revoke(ctx, token); err != nil {
			return nil, fmt.Errorf("cerrar sesión: %w", err)
		}
	}
	return s.clearCookies(), nil
}

// Resolve implementa session.Resolver (authn): resuelve la identidad de una
// cuenta a partir de su id con los permisos VIGENTES de su rol en cada petición
// (FR-018/SC-009) y exige que la cuenta siga activa (FR-012). Un cambio de rol,
// de permisos o de estado se refleja desde la primera petición posterior.
func (s *authService) Resolve(ctx context.Context, userID uuid.UUID) (session.Identity, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return session.Identity{}, err
	}
	if !user.IsActive {
		return session.Identity{}, fmt.Errorf("%w", ErrAccountInactive)
	}
	role, err := s.repository.GetRoleByID(ctx, user.RoleID)
	if err != nil {
		return session.Identity{}, err
	}
	return session.Identity{
		UserID:             user.ID,
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Phone:              user.Phone,
		RoleID:             user.RoleID,
		RoleName:           user.RoleName,
		Permissions:        ensureStrings(role.Permissions),
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// ChangeMyPassword cambia la contraseña de la cuenta de la identidad (FR-020):
// exige la contraseña actual, aplica la política FR-010 (también "distinta del
// nombre/apellidos/correo"), guarda el hash nuevo, resuelve la obligación de
// cambio y revoca las DEMÁS sesiones conservando la actual (R17). No es una
// acción administrativa: no deja fila en admin_actions (FR-023).
func (s *authService) ChangeMyPassword(
	ctx context.Context,
	identity session.Identity,
	currentToken string,
	in ChangePasswordInput,
) error {
	auth, err := s.repository.GetUserAuthByEmail(ctx, identity.Email)
	if err != nil {
		return fmt.Errorf("leer credenciales: %w", err)
	}
	if err := password.Verify(auth.PasswordHash, in.CurrentPassword); err != nil {
		return apperr.Invalid(
			"Tu contraseña actual no coincide",
			apperr.WithDetails(map[string]any{
				currentPasswordDetail: "La contraseña actual no es correcta",
			}),
		)
	}

	hash, err := password.Hash(in.NewPassword, password.Context{
		FirstName: auth.FirstName,
		LastName:  auth.LastName,
		Email:     auth.Email,
	})
	if err != nil {
		return err
	}

	if err := s.repository.UpdateUserPassword(ctx, auth.ID, hash); err != nil {
		return fmt.Errorf("guardar la contraseña: %w", err)
	}
	if err := s.repository.SetUserMustChangePassword(ctx, auth.ID, false); err != nil {
		return fmt.Errorf("resolver el cambio obligatorio: %w", err)
	}
	if err := s.sessions.RevokeUserExcept(ctx, auth.ID, currentToken); err != nil {
		return fmt.Errorf("revocar las demás sesiones: %w", err)
	}
	return nil
}

// recordFailure registra best-effort un intento fallido (FR-022). El fallo del
// registro no cambia la respuesta (R23).
func (s *authService) recordFailure(ctx context.Context, userID *uuid.UUID, ip string) {
	s.loginEvents.RecordLoginEventBestEffort(ctx, audit.Event{
		UserID: userID,
		Result: audit.ResultFailure,
		IP:     ip,
	})
}

// registerFailure incrementa el contador de FR-006. Un fallo de Redis se queda
// en el log: no convierte un 401 en un 500 ni revela estado interno.
func (s *authService) registerFailure(ctx context.Context, identifier string) {
	if _, err := s.throttle.RegisterFailure(ctx, identifier); err != nil {
		s.logError(ctx, "no se pudo registrar el intento fallido", err)
	}
}

// sessionCookies construye las cookies de sesión y CSRF de un login correcto.
func (s *authService) sessionCookies(token, csrf string) []*http.Cookie {
	return []*http.Cookie{
		session.NewSessionCookie(s.cookies, token),
		session.NewCSRFCookie(s.cookies, csrf),
	}
}

// clearCookies construye las cookies de borrado del logout (FR-004).
func (s *authService) clearCookies() []*http.Cookie {
	return []*http.Cookie{
		session.ClearSessionCookie(s.cookies),
		session.ClearCSRFCookie(s.cookies),
	}
}

// logError registra un fallo con el logger de la petición (que lleva el
// request_id) o con el logger base del servicio.
func (s *authService) logError(ctx context.Context, message string, err error) {
	logger := s.logger
	if requestLogger := httpserver.RequestLoggerFromContext(ctx); requestLogger != nil {
		logger = requestLogger
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Error(message, slog.Any("error", err))
}

// normalizeEmail aplica la normalización de Q5 (trim + minúsculas) al
// identificador de login. El correo se guarda normalizado, así que la igualdad
// exacta del índice UNIQUE es la misma.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// isNotFound indica si el error es un apperr.NotFound (correo inexistente).
func isNotFound(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindNotFound
}

// loginFailureError es el 401 genérico e idéntico para todo fallo de
// credenciales (FR-003/SC-008).
func loginFailureError() error {
	return apperr.Unauthenticated(loginFailureMessage)
}

// retryAfterSeconds redondea hacia arriba el tiempo restante del bloqueo a
// segundos para la cabecera Retry-After y Details["retryAfterSeconds"].
func retryAfterSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int(math.Ceil(d.Seconds()))
}

// authService implementa el Resolver que consume authn (T227).
var _ session.Resolver = (*authService)(nil)
