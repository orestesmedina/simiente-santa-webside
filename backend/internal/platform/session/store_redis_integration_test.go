//go:build integration

package session

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"simiente-santa/backend/internal/platform/testutil/containers"
)

// newTestClient levanta un Redis real (testcontainers, T219) y lo deja limpio.
func newTestClient(t *testing.T) *redis.Client {
	t.Helper()

	client, err := NewRedisClient(containers.RedisURL(t))
	if err != nil {
		t.Fatalf("NewRedisClient() error: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if err := client.FlushAll(context.Background()).Err(); err != nil {
		t.Fatalf("FlushAll() error: %v", err)
	}
	return client
}

// newTestStore construye el Store sobre un Redis real y devuelve también el
// cliente para inspeccionar claves y TTL.
func newTestStore(t *testing.T, cfg StoreConfig) (*redis.Client, *redisStore) {
	t.Helper()

	client := newTestClient(t)
	store, err := NewRedisStore(client, cfg)
	if err != nil {
		t.Fatalf("NewRedisStore() error: %v", err)
	}
	concrete, ok := store.(*redisStore)
	if !ok {
		t.Fatalf("NewRedisStore() devolvió %T, se esperaba *redisStore", store)
	}
	return client, concrete
}

func TestIntegrationStoreCreateResolveRevoke(t *testing.T) {
	client, store := newTestStore(t, StoreConfig{IdleTTL: 30 * time.Minute, AbsoluteTTL: time.Hour})
	ctx := context.Background()
	userID := uuid.New()

	token, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if token == "" {
		t.Fatal("Create() devolvió un token vacío")
	}

	// El token en claro nunca se guarda (P1): solo su SHA-256 como clave.
	key := sessionKey(token)
	raw, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("Get(%s) error: %v", key, err)
	}
	if strings.Contains(raw, token) {
		t.Errorf("el token en claro aparece en Redis: %q", raw)
	}
	if !strings.Contains(raw, userID.String()) {
		t.Errorf("el valor de la sesión no contiene el userId: %q", raw)
	}

	// TTL = inactividad (30 min), no la vida absoluta.
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("TTL() error: %v", err)
	}
	if ttl <= 0 || ttl > 30*time.Minute {
		t.Errorf("TTL de la sesión = %v, se esperaba (0, 30m]", ttl)
	}

	// Índice por cuenta para revocar todas sus sesiones (FR-012).
	member, err := client.SIsMember(ctx, userSessionsKey(userID), HashToken(token)).Result()
	if err != nil {
		t.Fatalf("SIsMember() error: %v", err)
	}
	if !member {
		t.Error("el hash del token no quedó en user_sessions:<userId>")
	}

	got, err := store.Resolve(ctx, token)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, se esperaba %v", got.UserID, userID)
	}
	if got.AbsoluteExpiresAt.Sub(got.CreatedAt) != time.Hour {
		t.Errorf("vida absoluta = %v, se esperaba 1h", got.AbsoluteExpiresAt.Sub(got.CreatedAt))
	}

	if _, err := store.Resolve(ctx, "token-inexistente"); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Resolve(token) desconocido = %v, se esperaba ErrSessionNotFound", err)
	}

	if err := store.Revoke(ctx, token); err != nil {
		t.Fatalf("Revoke() error: %v", err)
	}
	if _, err := store.Resolve(ctx, token); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Resolve() tras Revoke = %v, se esperaba ErrSessionNotFound", err)
	}
	if member, err := client.SIsMember(ctx, userSessionsKey(userID), HashToken(token)).Result(); err != nil {
		t.Fatalf("SIsMember() error: %v", err)
	} else if member {
		t.Error("el índice conserva el hash tras revocar la sesión")
	}
}

func TestIntegrationStoreIdleTTLRefresh(t *testing.T) {
	_, store := newTestStore(t, StoreConfig{
		IdleTTL:          2 * time.Second,
		AbsoluteTTL:      time.Hour,
		LastSeenThrottle: 50 * time.Millisecond,
	})
	ctx := context.Background()

	token, err := store.Create(ctx, uuid.New())
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)
	if _, err := store.Resolve(ctx, token); err != nil {
		t.Fatalf("Resolve() a los 1.2s (dentro de la inactividad): %v", err)
	}

	// La actividad de los 1.2s refrescó el TTL: a los 2.4s la sesión sigue viva.
	time.Sleep(1200 * time.Millisecond)
	if _, err := store.Resolve(ctx, token); err != nil {
		t.Fatalf("Resolve() a los 2.4s (TTL refrescado a los 1.2s): %v", err)
	}

	// 2.2s sin actividad desde el último refresco → caduca por inactividad.
	time.Sleep(2200 * time.Millisecond)
	if _, err := store.Resolve(ctx, token); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("Resolve() tras 2.2s de inactividad = %v, se esperaba ErrSessionNotFound", err)
	}
}

func TestIntegrationStoreAbsoluteCutoff(t *testing.T) {
	_, store := newTestStore(t, StoreConfig{
		IdleTTL:          time.Hour,
		AbsoluteTTL:      2 * time.Second,
		LastSeenThrottle: 50 * time.Millisecond,
	})
	ctx := context.Background()

	token, err := store.Create(ctx, uuid.New())
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	start := time.Now()

	// La actividad es continua (cada 400ms), pero la vida absoluta de 2s corta
	// la sesión igualmente (R15).
	var lastErr error
	for time.Since(start) < 5*time.Second {
		time.Sleep(400 * time.Millisecond)
		if _, err := store.Resolve(ctx, token); err != nil {
			lastErr = err
			break
		}
	}
	if !errors.Is(lastErr, ErrSessionNotFound) {
		t.Fatalf("no se cortó la sesión tras la vida absoluta: err = %v", lastErr)
	}
	if elapsed := time.Since(start); elapsed < 2*time.Second {
		t.Errorf("la sesión se cortó a los %v, antes de los 2s de vida absoluta", elapsed)
	}
}

func TestIntegrationStoreRevokeUser(t *testing.T) {
	client, store := newTestStore(t, StoreConfig{IdleTTL: 30 * time.Minute, AbsoluteTTL: time.Hour})
	ctx := context.Background()
	userA, userB := uuid.New(), uuid.New()

	tokenA1, err := store.Create(ctx, userA)
	if err != nil {
		t.Fatalf("Create(A1) error: %v", err)
	}
	tokenA2, err := store.Create(ctx, userA)
	if err != nil {
		t.Fatalf("Create(A2) error: %v", err)
	}
	tokenB, err := store.Create(ctx, userB)
	if err != nil {
		t.Fatalf("Create(B) error: %v", err)
	}

	if err := store.RevokeUser(ctx, userA); err != nil {
		t.Fatalf("RevokeUser() error: %v", err)
	}

	for _, token := range []string{tokenA1, tokenA2} {
		if _, err := store.Resolve(ctx, token); !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("Resolve() de una sesión revocada = %v, se esperaba ErrSessionNotFound", err)
		}
		if n, err := client.Exists(ctx, sessionKey(token)).Result(); err != nil {
			t.Fatalf("Exists() error: %v", err)
		} else if n != 0 {
			t.Errorf("la clave sess de una sesión revocada sigue existiendo")
		}
	}
	if _, err := store.Resolve(ctx, tokenB); err != nil {
		t.Errorf("Resolve() de otra cuenta = %v, no debía verse afectada", err)
	}
	if n, err := client.Exists(ctx, userSessionsKey(userA)).Result(); err != nil {
		t.Fatalf("Exists() error: %v", err)
	} else if n != 0 {
		t.Error("user_sessions de la cuenta revocada sigue existiendo")
	}
}

func TestIntegrationStoreRevokeUserExcept(t *testing.T) {
	_, store := newTestStore(t, StoreConfig{IdleTTL: 30 * time.Minute, AbsoluteTTL: time.Hour})
	ctx := context.Background()
	userID := uuid.New()

	keep, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("Create(keep) error: %v", err)
	}
	other, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("Create(other) error: %v", err)
	}

	if err := store.RevokeUserExcept(ctx, userID, keep); err != nil {
		t.Fatalf("RevokeUserExcept() error: %v", err)
	}

	if _, err := store.Resolve(ctx, keep); err != nil {
		t.Errorf("Resolve(keep) = %v, la sesión actual debía conservarse", err)
	}
	if _, err := store.Resolve(ctx, other); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Resolve(other) = %v, se esperaba ErrSessionNotFound", err)
	}
}
