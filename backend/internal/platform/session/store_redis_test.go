package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestNewRedisStoreValidation(t *testing.T) {
	if _, err := NewRedisStore(nil, StoreConfig{IdleTTL: time.Minute, AbsoluteTTL: time.Hour}); err == nil {
		t.Error("NewRedisStore(nil, ...) = nil, se esperaba error")
	}

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	t.Cleanup(func() { _ = client.Close() })

	if _, err := NewRedisStore(client, StoreConfig{}); err == nil {
		t.Error("NewRedisStore() con tiempos inválidos = nil, se esperaba error")
	}
	if _, err := NewRedisStore(client, StoreConfig{IdleTTL: time.Minute, AbsoluteTTL: time.Hour}); err != nil {
		t.Errorf("NewRedisStore() con configuración válida = %v, se esperaba nil", err)
	}
}

func TestNewThrottleValidation(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	t.Cleanup(func() { _ = client.Close() })

	if _, err := NewThrottle(nil, 5, time.Minute); err == nil {
		t.Error("NewThrottle(nil, ...) = nil, se esperaba error")
	}
	if _, err := NewThrottle(client, 0, time.Minute); err == nil {
		t.Error("NewThrottle() con maxAttempts=0 = nil, se esperaba error")
	}
	if _, err := NewThrottle(client, 5, 0); err == nil {
		t.Error("NewThrottle() con lockout=0 = nil, se esperaba error")
	}
	if _, err := NewThrottle(client, 5, time.Minute); err != nil {
		t.Errorf("NewThrottle() válido = %v, se esperaba nil", err)
	}
}

func TestNewRedisClient(t *testing.T) {
	client, err := NewRedisClient("redis://localhost:6379/0")
	if err != nil {
		t.Fatalf("NewRedisClient() error: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if got := client.Options().Addr; got != "localhost:6379" {
		t.Errorf("Addr = %q, se esperaba localhost:6379", got)
	}
	if got := client.Options().DB; got != 0 {
		t.Errorf("DB = %d, se esperaba 0", got)
	}

	if _, err := NewRedisClient("no-es-una-url"); err == nil {
		t.Error("NewRedisClient() con una URL inválida = nil, se esperaba error")
	}
}

func TestSessionKeys(t *testing.T) {
	token := "token-de-prueba"
	wantKey := sessionKeyPrefix + HashToken(token)

	if got := sessionKey(token); got != wantKey {
		t.Errorf("sessionKey() = %q, se esperaba %q", got, wantKey)
	}
	if got := sessionKeyFromHash(HashToken(token)); got != wantKey {
		t.Errorf("sessionKeyFromHash() = %q, se esperaba %q", got, wantKey)
	}

	userID := uuid.New()
	if got := userSessionsKey(userID); got != userSessionsKeyPrefix+userID.String() {
		t.Errorf("userSessionsKey() = %q, se esperaba %s", got, userSessionsKeyPrefix+userID.String())
	}
}

func TestThrottleKeys(t *testing.T) {
	throttle := &Throttle{maxAttempts: 5}
	identifier := "ana@ejemplo.com"

	if got := throttle.failKey(identifier); got != loginFailKeyPrefix+identifier {
		t.Errorf("failKey() = %q, se esperaba %s", got, loginFailKeyPrefix+identifier)
	}
	if got := throttle.blockKey(identifier); got != loginBlockKeyPrefix+identifier {
		t.Errorf("blockKey() = %q, se esperaba %s", got, loginBlockKeyPrefix+identifier)
	}
}
