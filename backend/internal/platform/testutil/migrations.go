package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

// MigrationLockKey es la clave del lock de aviso (advisory lock) con el que las
// pruebas de migración destructivas se serializan entre sí. Es la palabra
// «SIMIENTE» en ASCII (0x53 0x49 0x4D 0x49 0x45 0x4E 0x54 0x45) para que se
// reconozca a simple vista en pg_locks. No la comparte producción: es una clave
// privada de las pruebas.
const MigrationLockKey int64 = 0x53494D49454E5445

// SerializeMigrations toma un lock de aviso TRANSACCIONAL de PostgreSQL
// (pg_advisory_xact_lock) para que dos pruebas de migración destructivas nunca
// se solapen, y debe ser la PRIMERA sentencia de la transacción de la prueba.
//
// Por qué existe: en CI hay UNA sola base compartida (`DATABASE_URL_TEST`) y
// todos los paquetes corren en paralelo contra ella. Los `down.sql` de
// `admin_actions` hacen `DELETE FROM admin_actions …` (RowExclusive) seguido de
// `ALTER TABLE admin_actions DROP/ADD CONSTRAINT` (AccessExclusive) dentro de la
// misma transacción. Con cuatro pruebas de migración a la vez (F3, F4, F6 y F7),
// dos de ellas se quedan cada una con el RowExclusive y piden el AccessExclusive:
// espera circular y PostgreSQL mata una —y en CI, la suite entera— con
// «deadlock detected (SQLSTATE 40P01)».
//
// El lock es de transacción: se libera solo al confirmar o revertir la
// transacción de la prueba, así que no hay nada que limpiar. Las pruebas que no
// ejecutan DDL destructivo (p. ej. las de auditoría) no lo necesitan: sus
// INSERT no entran en conflicto con el AccessExclusive de la migración.
func SerializeMigrations(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, MigrationLockKey); err != nil {
		t.Fatalf("tomar el lock de serialización de las pruebas de migración (%d): %v", MigrationLockKey, err)
	}
}
