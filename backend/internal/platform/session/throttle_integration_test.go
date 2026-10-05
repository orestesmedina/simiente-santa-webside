//go:build integration

package session

import (
	"context"
	"testing"
	"time"
)

func TestIntegrationThrottleCountsAndFlag(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	const maxAttempts = 5
	lockout := 15 * time.Minute
	throttle, err := NewThrottle(client, maxAttempts, lockout)
	if err != nil {
		t.Fatalf("NewThrottle() error: %v", err)
	}

	identifier := "ana@ejemplo.com"

	if blocked, _, err := throttle.Blocked(ctx, identifier); err != nil {
		t.Fatalf("Blocked() error: %v", err)
	} else if blocked {
		t.Fatal("Blocked() = true sin fallos previos")
	}

	// Los primeros cuatro fallos no bloquean.
	for i := 1; i < maxAttempts; i++ {
		count, err := throttle.RegisterFailure(ctx, identifier)
		if err != nil {
			t.Fatalf("RegisterFailure() error: %v", err)
		}
		if count != int64(i) {
			t.Fatalf("RegisterFailure() = %d en el fallo %d", count, i)
		}
		if blocked, _, err := throttle.Blocked(ctx, identifier); err != nil {
			t.Fatalf("Blocked() error: %v", err)
		} else if blocked {
			t.Fatalf("Blocked() = true tras %d fallos, se esperaba false", i)
		}
	}

	// El 5.º fallo crea la bandera de bloqueo (la respuesta 401 la decide el
	// dominio, no este mecanismo).
	count, err := throttle.RegisterFailure(ctx, identifier)
	if err != nil {
		t.Fatalf("RegisterFailure() error: %v", err)
	}
	if count != maxAttempts {
		t.Fatalf("RegisterFailure() = %d, se esperaba %d", count, maxAttempts)
	}

	blocked, retryAfter, err := throttle.Blocked(ctx, identifier)
	if err != nil {
		t.Fatalf("Blocked() error: %v", err)
	}
	if !blocked {
		t.Fatal("Blocked() = false tras el 5.º fallo, se esperaba true")
	}
	if retryAfter <= 0 || retryAfter > lockout {
		t.Errorf("Retry-After = %v, se esperaba (0, %v]", retryAfter, lockout)
	}

	// Ambas claves existen con su TTL.
	if n, err := client.Exists(ctx, loginFailKeyPrefix+identifier, loginBlockKeyPrefix+identifier).Result(); err != nil {
		t.Fatalf("Exists() error: %v", err)
	} else if n != 2 {
		t.Errorf("claves de contador/bandera = %d, se esperaban 2", n)
	}

	if err := throttle.Reset(ctx, identifier); err != nil {
		t.Fatalf("Reset() error: %v", err)
	}
	if blocked, _, err := throttle.Blocked(ctx, identifier); err != nil {
		t.Fatalf("Blocked() error: %v", err)
	} else if blocked {
		t.Error("Blocked() = true tras Reset()")
	}
	if n, err := client.Exists(ctx, loginFailKeyPrefix+identifier, loginBlockKeyPrefix+identifier).Result(); err != nil {
		t.Fatalf("Exists() error: %v", err)
	} else if n != 0 {
		t.Errorf("quedan %d claves tras Reset()", n)
	}
}

// TestIntegrationThrottleFailureKeyAlwaysHasTTL fija M4: el TTL del contador
// `login:fail:*` se fija de forma atómica con el incremento, incluso si la clave
// ya existía sin expiración (residuo de una versión anterior).
func TestIntegrationThrottleFailureKeyAlwaysHasTTL(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	lockout := 15 * time.Minute
	throttle, err := NewThrottle(client, 5, lockout)
	if err != nil {
		t.Fatalf("NewThrottle() error: %v", err)
	}

	identifier := "ttl@ejemplo.com"
	failKey := loginFailKeyPrefix + identifier

	// Primera falla: la clave nace con TTL.
	if _, err := throttle.RegisterFailure(ctx, identifier); err != nil {
		t.Fatalf("RegisterFailure() error: %v", err)
	}
	ttl, err := client.TTL(ctx, failKey).Result()
	if err != nil {
		t.Fatalf("TTL() error: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("la clave de fallos se creó sin TTL: %v", ttl)
	}

	// Residuo sin TTL: el siguiente fallo le repone la expiración de forma
	// atómica.
	if err := client.Set(ctx, failKey, "4", 0).Err(); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	count, err := throttle.RegisterFailure(ctx, identifier)
	if err != nil {
		t.Fatalf("RegisterFailure() error: %v", err)
	}
	if count != 5 {
		t.Fatalf("contador = %d, se esperaba 5", count)
	}
	ttl, err = client.TTL(ctx, failKey).Result()
	if err != nil {
		t.Fatalf("TTL() error: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("la clave de fallos quedó sin TTL: %v", ttl)
	}
}

func TestIntegrationThrottleLockoutExpires(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	const maxAttempts = 5
	lockout := 2 * time.Second
	throttle, err := NewThrottle(client, maxAttempts, lockout)
	if err != nil {
		t.Fatalf("NewThrottle() error: %v", err)
	}

	identifier := "bloqueo@ejemplo.com"
	for i := 0; i < maxAttempts; i++ {
		if _, err := throttle.RegisterFailure(ctx, identifier); err != nil {
			t.Fatalf("RegisterFailure() error: %v", err)
		}
	}

	blocked, retryAfter, err := throttle.Blocked(ctx, identifier)
	if err != nil {
		t.Fatalf("Blocked() error: %v", err)
	}
	if !blocked {
		t.Fatal("Blocked() = false tras el bloqueo")
	}
	if retryAfter <= 0 || retryAfter > lockout {
		t.Errorf("Retry-After = %v, se esperaba (0, %v]", retryAfter, lockout)
	}

	// El TTL de la bandera vence y se puede volver a intentar (sin limpieza
	// manual).
	time.Sleep(lockout + 500*time.Millisecond)
	if blocked, _, err := throttle.Blocked(ctx, identifier); err != nil {
		t.Fatalf("Blocked() error: %v", err)
	} else if blocked {
		t.Error("Blocked() = true pasado el TTL de la bandera")
	}
}
