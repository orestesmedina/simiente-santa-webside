package session

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// Prefijos de las claves de los contadores de FR-006 (data-model.md).
const (
	loginFailKeyPrefix  = "login:fail:"
	loginBlockKeyPrefix = "login:block:"
	blockFlagValue      = "1"
)

// registerFailureScript incrementa el contador de fallos y fija su TTL en una
// sola operación atómica de Redis (M4): no puede quedar una clave de
// `login:fail:*` sin expiración. Además crea la bandera de bloqueo con su TTL
// cuando el contador alcanza el máximo. Argumentos: [ttlSegundos, maxAttempts,
// valorBandera]; devuelve el contador acumulado.
var registerFailureScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
redis.call('EXPIRE', KEYS[1], ARGV[1])
if count >= tonumber(ARGV[2]) then
    redis.call('SET', KEYS[2], ARGV[3], 'EX', ARGV[1])
end
return count
`)

// Throttle es el MECANISMO de los contadores de intentos fallidos de FR-006
// (reparto F-13): sabe incrementar el contador, crear la bandera de bloqueo con
// su TTL y devolver el Retry-After. La SEMÁNTICA y las constantes (cuántos
// fallos, cuánto dura el bloqueo, que el 5.º fallo aún responde 401) las pone
// el dominio usuarios (T225) al construirlo; este paquete no fija valores.
//
// Las claves se indexan por el identificador normalizado (correo) exista o no
// la cuenta, de modo que el comportamiento no revela existencia (FR-003).
type Throttle struct {
	client      redis.UniversalClient
	maxAttempts int
	lockout     time.Duration
}

// NewThrottle construye el mecanismo de contadores. maxAttempts es el número de
// fallos a partir del cual se crea la bandera de bloqueo; lockout es el TTL de
// la bandera y del contador (ambos decididos por el dominio).
func NewThrottle(client redis.UniversalClient, maxAttempts int, lockout time.Duration) (*Throttle, error) {
	if client == nil {
		return nil, errors.New("session: cliente de Redis nulo")
	}
	if maxAttempts <= 0 {
		return nil, fmt.Errorf("session: maxAttempts debe ser positivo (recibido %d)", maxAttempts)
	}
	if lockout <= 0 {
		return nil, fmt.Errorf("session: lockout debe ser positivo (recibido %v)", lockout)
	}
	return &Throttle{client: client, maxAttempts: maxAttempts, lockout: lockout}, nil
}

// Blocked indica si el identificador está bloqueado y cuánto queda hasta poder
// reintentar (Retry-After). No está bloqueado si la bandera no existe (su TTL
// cumplido la borra sola, sin limpieza manual).
func (t *Throttle) Blocked(ctx context.Context, identifier string) (blocked bool, retryAfter time.Duration, err error) {
	ttl, err := t.client.TTL(ctx, t.blockKey(identifier)).Result()
	if err != nil {
		return false, 0, fmt.Errorf("consultar bloqueo: %w", err)
	}
	// go-redis devuelve -2 si la clave no existe y -1 si no tiene expiración.
	if ttl <= 0 {
		return false, 0, nil
	}
	return true, ttl, nil
}

// RegisterFailure incrementa el contador de fallos consecutivos del
// identificador y, al alcanzar maxAttempts, crea la bandera de bloqueo con su
// TTL. Devuelve el número de fallos acumulados para que el dominio decida la
// respuesta (el intento que alcanza el máximo todavía responde 401; el bloqueo
// se aplica a partir del siguiente, que ya ve la bandera).
func (t *Throttle) RegisterFailure(ctx context.Context, identifier string) (int64, error) {
	ttlSeconds := int64(math.Ceil(t.lockout.Seconds()))
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}
	count, err := registerFailureScript.Run(ctx, t.client,
		[]string{t.failKey(identifier), t.blockKey(identifier)},
		ttlSeconds, t.maxAttempts, blockFlagValue,
	).Int64()
	if err != nil {
		return 0, fmt.Errorf("registrar intento fallido: %w", err)
	}
	return count, nil
}

// Reset borra el contador y la bandera de bloqueo del identificador: lo llama
// un inicio de sesión correcto.
func (t *Throttle) Reset(ctx context.Context, identifier string) error {
	if err := t.client.Del(ctx, t.failKey(identifier), t.blockKey(identifier)).Err(); err != nil {
		return fmt.Errorf("limpiar contadores de intentos: %w", err)
	}
	return nil
}

// failKey es la clave del contador de fallos.
func (t *Throttle) failKey(identifier string) string {
	return loginFailKeyPrefix + identifier
}

// blockKey es la clave de la bandera de bloqueo.
func (t *Throttle) blockKey(identifier string) string {
	return loginBlockKeyPrefix + identifier
}
