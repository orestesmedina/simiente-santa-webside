//go:build integration

package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"simiente-santa/backend/internal/platform/database"
)

func TestRedisURL(t *testing.T) {
	url := RedisURL(t)
	if url == "" {
		t.Fatal("RedisURL() devolvió una URL vacía")
	}

	options, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("parsear la URL de Redis %q: %v", url, err)
	}
	client := redis.NewClient(options)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping a Redis: %v", err)
	}
	if err := client.Set(ctx, "clave-de-prueba", "valor", time.Minute).Err(); err != nil {
		t.Fatalf("set en Redis: %v", err)
	}
	value, err := client.Get(ctx, "clave-de-prueba").Result()
	if err != nil {
		t.Fatalf("get en Redis: %v", err)
	}
	if value != "valor" {
		t.Errorf("valor = %q, se esperaba %q", value, "valor")
	}
}

func TestPostgresURL(t *testing.T) {
	url := PostgresURL(t)
	if url == "" {
		t.Fatal("PostgresURL() devolvió una URL vacía")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("construir el pool de PostgreSQL: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping a PostgreSQL: %v", err)
	}
}

func TestPostgresURLReusesDatabaseURLTest(t *testing.T) {
	const reused = "postgres://reutilizada:5432/simiente_test?sslmode=disable"
	t.Setenv(DatabaseURLEnv, reused)

	if got := PostgresURL(t); got != reused {
		t.Errorf("PostgresURL() = %q, se esperaba reutilizar %q", got, reused)
	}
}
