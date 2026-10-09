package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DefaultLastSeenThrottle estrangula la escritura de lastSeenAt a una vez por
// minuto (R1/R15): la actividad refresca el TTL en cada petición, pero no
// reescribe el valor de la sesión en cada una.
const DefaultLastSeenThrottle = time.Minute

// ErrSessionNotFound indica que no hay sesión viva para el token: no existía,
// caducó por inactividad (el TTL de Redis la borró) o caducó por vida absoluta.
// authn la traduce a 401 unauthenticated.
var ErrSessionNotFound = errors.New("session: sesión no encontrada o expirada")

// Session es el estado efímero de una sesión en Redis (R1). Se serializa a JSON
// en `sess:<sha256(token)>`; el token en claro nunca se guarda.
type Session struct {
	UserID            uuid.UUID `json:"userId"`
	CreatedAt         time.Time `json:"createdAt"`
	LastSeenAt        time.Time `json:"lastSeenAt"`
	AbsoluteExpiresAt time.Time `json:"absoluteExpiresAt"`
}

// ExpiredAt indica si la sesión superó su vida absoluta en now (R15).
func (s Session) ExpiredAt(now time.Time) bool {
	return !now.Before(s.AbsoluteExpiresAt)
}

// Store gestiona las sesiones del panel. Lo implementa la variante Redis
// (NewRedisStore); el dominio usuarios lo consume para crear, resolver y
// revocar sesiones (FR-001/FR-004/FR-005/FR-012).
type Store interface {
	// Create abre una sesión para la cuenta y devuelve el token en claro que
	// viaja en la cookie. El token no se persiste: en Redis queda su hash.
	Create(ctx context.Context, userID uuid.UUID) (string, error)
	// Resolve devuelve la sesión viva del token y refresca su TTL de
	// inactividad (acotado a la vida absoluta) y, de forma estrangulada,
	// lastSeenAt. Devuelve ErrSessionNotFound si no hay sesión viva.
	Resolve(ctx context.Context, token string) (Session, error)
	// Revoke cierra la sesión del token (logout, FR-004).
	Revoke(ctx context.Context, token string) error
	// RevokeUser cierra todas las sesiones de una cuenta (desactivación o
	// restablecimiento de contraseña, FR-012/R17).
	RevokeUser(ctx context.Context, userID uuid.UUID) error
	// RevokeUserExcept cierra todas las sesiones de una cuenta salvo la del
	// token indicado (cambio de contraseña propio: se conserva la sesión
	// actual, R17).
	RevokeUserExcept(ctx context.Context, userID uuid.UUID, keepToken string) error
}

// StoreConfig son los tiempos de la sesión, tomados de platform/config:
// SESSION_IDLE_TTL_MINUTES (30) y SESSION_ABSOLUTE_TTL_MINUTES (60). La vida
// absoluta es inmóvil; el TTL de la clave es la inactividad, acotada a lo que
// quede de vida absoluta (R15).
type StoreConfig struct {
	// IdleTTL es la vida por inactividad; refrescada en cada actividad.
	IdleTTL time.Duration
	// AbsoluteTTL es la vida absoluta desde el login; nunca se refresca.
	AbsoluteTTL time.Duration
	// LastSeenThrottle acota cada cuánto se reescribe lastSeenAt. Si es <= 0 se
	// usa DefaultLastSeenThrottle.
	LastSeenThrottle time.Duration
}

// Validate comprueba que los tiempos de sesión son utilizables.
func (c StoreConfig) Validate() error {
	if c.IdleTTL <= 0 {
		return fmt.Errorf("session: IdleTTL debe ser positiva (recibido %v)", c.IdleTTL)
	}
	if c.AbsoluteTTL <= 0 {
		return fmt.Errorf("session: AbsoluteTTL debe ser positiva (recibido %v)", c.AbsoluteTTL)
	}
	return nil
}

// lastSeenThrottle devuelve el estrangulamiento efectivo de lastSeenAt.
func (c StoreConfig) lastSeenThrottle() time.Duration {
	if c.LastSeenThrottle <= 0 {
		return DefaultLastSeenThrottle
	}
	return c.LastSeenThrottle
}

// nextTTL devuelve el TTL con el que se refresca la clave de sesión:
// min(IdleTTL, tiempo restante hasta AbsoluteExpiresAt) (R15), de modo que
// Redis expire la clave como muy tarde en la vida absoluta aunque haya
// actividad continua.
func nextTTL(now time.Time, s Session, idleTTL time.Duration) time.Duration {
	remaining := s.AbsoluteExpiresAt.Sub(now)
	if remaining < idleTTL {
		return remaining
	}
	return idleTTL
}
