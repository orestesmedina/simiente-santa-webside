//go:build integration

package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// codeLockNotAvailable es el SQLSTATE que PostgreSQL devuelve cuando un lock no
// se puede tomar dentro de lock_timeout (55P03 lock_not_available).
const codeLockNotAvailable = "55P03"

// TestIntegrationSerializeMigrationsExcludesConcurrentMigrations es la prueba de
// regresión del helper (hallazgo bloqueante del CI del PR #10: 40P01 entre las
// pruebas de migración del CI, donde una sola base compartida recibe a todos los
// paquetes en paralelo). No depende del azar ni del planificador: comprueba en
// tres pasos que el helper serializa de verdad.
//
//  1. tras llamarlo, la conexión de la prueba tiene tomado un lock de aviso
//     (sin el helper, pg_locks no tendría ninguno y esta aserción falla);
//  2. otra transacción que pida la MISMA clave espera en vez de obtenerla
//     (lock_timeout corto → 55P03), que es justo lo que impide el solape de los
//     `down` destructivos;
//  3. al revertirse la transacción de la prueba, el lock se libera solo, sin
//     limpieza manual (es lo que permite que la siguiente prueba lo obtenga).
func TestIntegrationSerializeMigrationsExcludesConcurrentMigrations(t *testing.T) {
	ctx := context.Background()
	pool := Pool(t, ctx)

	// Sesión A: la transacción de una prueba de migración.
	txA, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transacción A: %v", err)
	}
	t.Cleanup(func() { _ = txA.Rollback(ctx) })

	SerializeMigrations(t, ctx, txA)

	// 1) El lock existe y lo tiene la conexión de A.
	var lockerPID int
	if err := txA.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&lockerPID); err != nil {
		t.Fatalf("leer el pid de la transacción A: %v", err)
	}
	var held int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_locks
		WHERE locktype = 'advisory' AND pid = $1`, lockerPID).Scan(&held); err != nil {
		t.Fatalf("consultar pg_locks: %v", err)
	}
	if held != 1 {
		t.Fatalf("locks de aviso tomados por la conexión de la prueba = %d, se esperaba 1 "+
			"(¿SerializeMigrations no está tomando el lock?)", held)
	}

	// 2) Una segunda transacción que pide la misma clave espera y se rinde con
	// lock_timeout (si el helper no serializara, la obtendría de inmediato).
	txB, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transacción B: %v", err)
	}
	t.Cleanup(func() { _ = txB.Rollback(ctx) })
	if _, err := txB.Exec(ctx, `SET LOCAL lock_timeout = '500ms'`); err != nil {
		t.Fatalf("fijar lock_timeout: %v", err)
	}
	if _, err := txB.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, MigrationLockKey); err == nil {
		t.Fatal("una segunda transacción obtuvo el mismo lock: las pruebas de migración no quedan serializadas")
	} else {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != codeLockNotAvailable {
			t.Fatalf("error = %v, se esperaba SQLSTATE %s (lock_not_available)", err, codeLockNotAvailable)
		}
	}
	// La transacción B quedó abortada por el timeout: se descarta y su conexión
	// vuelve al pool.
	if err := txB.Rollback(ctx); err != nil {
		t.Fatalf("revertir la transacción B: %v", err)
	}

	// 3) Al revertir A, el lock de transacción se libera solo: no hay nada que
	// limpiar a mano. No se comprueba pidiendo la clave otra vez (las pruebas de
	// migración de los OTROS paquetes compiten por la misma clave y pueden
	// tenerla tomada en este instante de forma legítima), sino que la conexión de
	// A ya no retiene ningún lock de aviso.
	if err := txA.Rollback(ctx); err != nil {
		t.Fatalf("revertir la transacción A: %v", err)
	}
	var stillHeld int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_locks
		WHERE locktype = 'advisory' AND pid = $1`, lockerPID).Scan(&stillHeld); err != nil {
		t.Fatalf("consultar pg_locks tras el rollback: %v", err)
	}
	if stillHeld != 0 {
		t.Fatalf("la conexión de la prueba sigue con %d locks de aviso tras revertir: "+
			"el lock debería liberarse con la transacción", stillHeld)
	}
}
