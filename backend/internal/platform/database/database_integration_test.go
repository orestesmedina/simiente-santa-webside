//go:build integration

package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Helpers propios y mínimos de este archivo: no usa platform/testutil para no
// crear la dependencia circular T010 ↔ T013 (nota 3 de tasks.md).

// testDatabaseURL devuelve DATABASE_URL_TEST o salta la prueba si no existe.
func testDatabaseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("DATABASE_URL_TEST")
	if url == "" {
		t.Skip("DATABASE_URL_TEST no está definida; se omiten las pruebas de integración")
	}
	return url
}

// testPool construye un pool contra DATABASE_URL_TEST y lo cierra al terminar.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := NewPool(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("NewPool() error: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// waitPing espera a que la base de datos responda (tolera un arranque lento).
func waitPing(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if err := Ping(context.Background(), pool); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("la base de datos no respondió en 10 s")
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestIntegrationNewPoolAndPing(t *testing.T) {
	pool := testPool(t)
	waitPing(t, pool)

	if err := Ping(context.Background(), pool); err != nil {
		t.Fatalf("Ping() con la BD viva devolvió error: %v", err)
	}
}

func TestIntegrationPingUnreachableHost(t *testing.T) {
	testDatabaseURL(t) // salta si no hay BD configurada

	pool, err := NewPool(context.Background(), "postgres://app:app@10.255.255.1:5432/app?sslmode=disable")
	if err != nil {
		t.Fatalf("NewPool() error: %v", err)
	}
	defer pool.Close()

	start := time.Now()
	err = Ping(context.Background(), pool)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Ping() con un host inalcanzable no devolvió error")
	}
	if elapsed > PingTimeout+time.Second {
		t.Errorf("Ping() tardó %v, se esperaba ≤%v", elapsed, PingTimeout+time.Second)
	}
}

func TestIntegrationWithTxCommitAndRollback(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	waitPing(t, pool)

	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS _database_tx_test (id integer PRIMARY KEY)`); err != nil {
		t.Fatalf("crear tabla de prueba: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DROP TABLE IF EXISTS _database_tx_test`)
	})
	if _, err := pool.Exec(ctx, `TRUNCATE _database_tx_test`); err != nil {
		t.Fatalf("limpiar tabla de prueba: %v", err)
	}

	// Commit: fn devuelve nil.
	err := WithTx(ctx, pool, func(tx pgx.Tx) error {
		_, execErr := tx.Exec(ctx, `INSERT INTO _database_tx_test (id) VALUES (1)`)
		return execErr
	})
	if err != nil {
		t.Fatalf("WithTx (commit) error: %v", err)
	}
	if got := countTxRows(t, pool); got != 1 {
		t.Fatalf("filas tras el commit = %d, se esperaba 1", got)
	}

	// Rollback: fn inserta y luego devuelve un error.
	sentinel := errors.New("rollback a propósito")
	err = WithTx(ctx, pool, func(tx pgx.Tx) error {
		if _, execErr := tx.Exec(ctx, `INSERT INTO _database_tx_test (id) VALUES (2)`); execErr != nil {
			return execErr
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx (rollback) error = %v, se esperaba %v", err, sentinel)
	}
	if got := countTxRows(t, pool); got != 1 {
		t.Fatalf("filas tras el rollback = %d, se esperaba 1 (la fila 2 no debió persistir)", got)
	}
}

func countTxRows(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM _database_tx_test`).Scan(&n); err != nil {
		t.Fatalf("contar filas: %v", err)
	}
	return n
}
