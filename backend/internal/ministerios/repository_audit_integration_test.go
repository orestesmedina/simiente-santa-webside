//go:build integration

package ministerios

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
)

// Pruebas de integración de la auditoría del dominio ministerios contra
// PostgreSQL real (T611): la mutación y su fila de `admin_actions` comparten
// transacción (FR-017, fail-closed: si el registro falla, la mutación NO se
// aplica) y el registro cerrado de códigos de la BD coincide con
// `platform/audit` (T606/R6-12), incluidos los dos códigos de la visibilidad de
// la ficha del encargado.
//
// La fila de auditoría se ejercita, como en las pruebas de F4, con el único
// código que la BD admite sin actor (`user.create`): la tabla `users` la truncan
// en paralelo las pruebas de integración del dominio usuarios y estas pruebas no
// la tocan a propósito.

// ministryActionCodes devuelve los códigos `ministry.*` del registro cerrado de
// platform/audit (T606), en el orden de la lista.
func ministryActionCodes() []string {
	codes := make([]string, 0, 8)
	for _, code := range audit.ActionCodes {
		if strings.HasPrefix(code, "ministry.") {
			codes = append(codes, code)
		}
	}
	return codes
}

// auditRowWithoutActor arma la fila de auditoría de una mutación de contenido
// sin actor (el único caso que la BD admite así es `user.create`).
func auditRowWithoutActor(label string) audit.Action {
	return audit.Action{
		Code:        audit.ActionUserCreate,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// TestIntegrationWriteTxRollsBackWhenAuditFails fija el edge case de FR-017: si
// el registro de auditoría falla en la base (aquí, un `target_label` que supera
// el CHECK de 254 caracteres), la mutación NO se aplica.
func TestIntegrationWriteTxRollsBackWhenAuditFails(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	names := []string{"No debe quedar", uuid.NewString()}
	name := names[1]
	tooLong := strings.Repeat("x", 300)

	err := repo.withTx(ctx, func(tx *repository) error {
		if _, err := tx.q.InsertMinistry(ctx, insertParams(baseMinistry(name))); err != nil {
			return err
		}
		return insertActions(ctx, tx, []audit.Action{auditRowWithoutActor(tooLong)})
	})
	if err == nil {
		t.Fatal("withTx no devolvió error con un registro inválido")
	}
	if kind := errorKind(err); kind != apperr.KindInvalid {
		t.Errorf("kind = %v, se esperaba invalid (CHECK de la base)", kind)
	}

	var stored int
	if err := repo.pool.QueryRow(ctx, `SELECT count(*) FROM ministries WHERE name_es = $1`, name).Scan(&stored); err != nil {
		t.Fatalf("contar ministerios: %v", err)
	}
	if stored != 0 {
		t.Errorf("la mutación se aplicó sin registro: %d ministerios", stored)
	}
	if got := countAdminActions(t, repo, tooLong); got != 0 {
		t.Errorf("quedó una fila de auditoría inválida: %d", got)
	}
}

// TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown comprueba el otro lado
// del fail-closed: un código fuera del registro cerrado de `platform/audit` se
// rechaza ANTES de tocar la base y la mutación tampoco se aplica.
func TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	name := "Ministerio con código desconocido " + uuid.NewString()

	err := repo.withTx(ctx, func(tx *repository) error {
		if _, err := tx.q.InsertMinistry(ctx, insertParams(baseMinistry(name))); err != nil {
			return err
		}
		return insertActions(ctx, tx, []audit.Action{{
			Code:        "ministry.archive", // fuera del registro cerrado (R6-12)
			TargetKind:  audit.TargetContent,
			TargetLabel: "Ministerios · X",
			Result:      audit.ResultSuccess,
		}})
	})
	if err == nil {
		t.Fatal("withTx no devolvió error con un código desconocido")
	}
	if !strings.Contains(err.Error(), audit.ErrUnknownActionCode.Error()) {
		t.Errorf("el error no identifica el registro cerrado: %v", err)
	}

	var stored int
	if err := repo.pool.QueryRow(ctx, `SELECT count(*) FROM ministries WHERE name_es = $1`, name).Scan(&stored); err != nil {
		t.Fatalf("contar ministerios: %v", err)
	}
	if stored != 0 {
		t.Errorf("la mutación se aplicó sin registro válido: %d ministerios", stored)
	}
}

// TestIntegrationAdminActionRegistryAdmitsMinistryCodes fija el registro cerrado
// de F6 (FR-017/R6-12/T606): los 8 códigos `ministry.*` que declara
// `platform/audit` coinciden uno a uno con el CHECK de la base, la base no admite
// ningún código de F6 de más, los 8 caben en el registro y un código fuera de él
// se rechaza.
func TestIntegrationAdminActionRegistryAdmitsMinistryCodes(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	// El CHECK que se compara es el que define la migración 000011: se aplica
	// dentro de la transacción de prueba porque el estado confirmado de la base
	// puede haber quedado en el de otra funcionalidad (las pruebas de integración
	// de F3/F4 reejecutan sus propios up/down sobre `admin_actions`).
	execMigrationInTx(t, ctx, tx, "000011_extend_admin_actions_for_ministries.up.sql")

	def := constraintDefTx(t, ctx, tx, "admin_actions_action_check", "admin_actions")

	codes := ministryActionCodes()
	if len(codes) != 8 {
		t.Fatalf("platform/audit declara %d códigos ministry.*, se esperaban 8", len(codes))
	}
	for _, code := range codes {
		if !strings.Contains(def, "'"+code+"'") {
			t.Fatalf("el CHECK de la BD no incluye %q", code)
		}
	}
	if inCheck := strings.Count(def, "'ministry."); inCheck != len(codes) {
		t.Fatalf("el CHECK tiene %d códigos ministry.* y platform/audit declara %d", inCheck, len(codes))
	}
	// Los dos códigos de la visibilidad de la ficha del encargado son propios de
	// F6 y forman parte del registro.
	for _, code := range []string{audit.ActionMinistryLeaderShow, audit.ActionMinistryLeaderHide} {
		if !strings.Contains(def, "'"+code+"'") {
			t.Fatalf("el registro no incluye el código de visibilidad %q", code)
		}
	}

	// Los 8 códigos CABEN en el registro. La regla del actor (F2: sin actor solo
	// `user.create`) exige un usuario real que esta prueba no crea a propósito
	// (la tabla `users` la truncan en paralelo las pruebas del dominio usuarios),
	// así que se retira DENTRO de esta transacción —se revierte al terminar y no
	// toca ninguna migración (§VI)— para aislar el registro de códigos.
	if _, err := tx.Exec(ctx, `ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check`); err != nil {
		t.Fatalf("retirar temporalmente la regla del actor: %v", err)
	}

	const insert = `INSERT INTO admin_actions (actor_user_id, action, target_kind, target_label, result)
		VALUES (NULL, $1, 'content', 'Ministerios · Registro', 'success')`
	for _, code := range codes {
		if _, err := tx.Exec(ctx, insert, code); err != nil {
			t.Fatalf("la base rechazó el código %q del registro: %v", code, err)
		}
	}

	// Un código fuera del registro se rechaza (fail-closed). Va dentro de un
	// savepoint: el fallo abortaría la transacción de prueba y no dejaría
	// comprobar el recuento de las filas válidas.
	sp, err := tx.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir savepoint: %v", err)
	}
	_, insertErr := sp.Exec(ctx, insert, "ministry.archive")
	if rollbackErr := sp.Rollback(ctx); rollbackErr != nil {
		t.Fatalf("revertir el savepoint: %v", rollbackErr)
	}
	if insertErr == nil {
		t.Fatal("la base aceptó un código fuera del registro")
	}
	if !strings.Contains(insertErr.Error(), "admin_actions_action_check") {
		t.Fatalf("el error no viene del CHECK del registro: %v", insertErr)
	}

	// Y las filas de los 8 códigos quedaron registradas en la transacción.
	if got := countTx(t, ctx, tx,
		`SELECT count(*) FROM admin_actions WHERE target_label = 'Ministerios · Registro'`); got != 8 {
		t.Errorf("filas de auditoría = %d, se esperaban 8", got)
	}
}
