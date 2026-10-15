//go:build integration

package donaciones

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
)

// Pruebas de integración de la auditoría del dominio donaciones contra
// PostgreSQL real (T711): la mutación y su fila de `admin_actions` comparten
// transacción (FE-017, fail-closed: si el registro falla, la mutación NO se
// aplica) y el registro cerrado de códigos de la BD coincide con
// `platform/audit` (T706/R7-10).
//
// Como en F4/F6, las filas de auditoría de actor nulo se ejercitan con el único
// código que la BD admite sin actor (`user.create`): estas pruebas no tocan la
// tabla `users`, que las pruebas de integración del dominio usuarios truncan en
// paralelo.

// donationActionCodes devuelve los códigos `donation.*` del registro cerrado de
// platform/audit (T706), en el orden de la lista.
func donationActionCodes() []string {
	codes := make([]string, 0, 8)
	for _, code := range audit.ActionCodes {
		if strings.HasPrefix(code, "donation.") {
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
	repo := newDonationRepo(t)
	ctx := context.Background()

	tooLong := strings.Repeat("x", 300)
	method := baseIban()
	method.Number = "CR05015202001234567897"

	err := repo.withTx(ctx, func(tx *repository) error {
		if _, err := tx.q.InsertDonationMethod(ctx, insertMethodParams(method)); err != nil {
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

	if got := countMetodos(t, repo); got != 0 {
		t.Errorf("la mutación se aplicó sin registro: %d métodos", got)
	}
}

// TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown comprueba el otro lado
// del fail-closed: un código fuera del registro cerrado de `platform/audit` se
// rechaza ANTES de tocar la base y la mutación tampoco se aplica.
func TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	method := baseIban()
	method.Number = "CR05015202001234567898"

	err := repo.withTx(ctx, func(tx *repository) error {
		if _, err := tx.q.InsertDonationMethod(ctx, insertMethodParams(method)); err != nil {
			return err
		}
		return insertActions(ctx, tx, []audit.Action{{
			Code:        "donation.archive", // fuera del registro cerrado (R7-10)
			TargetKind:  audit.TargetContent,
			TargetLabel: "Donaciones · X",
			Result:      audit.ResultSuccess,
		}})
	})
	if err == nil {
		t.Fatal("withTx no devolvió error con un código desconocido")
	}
	if !strings.Contains(err.Error(), audit.ErrUnknownActionCode.Error()) {
		t.Errorf("el error no identifica el registro cerrado: %v", err)
	}

	if got := countMetodos(t, repo); got != 0 {
		t.Errorf("la mutación se aplicó sin registro válido: %d métodos", got)
	}
}

// TestIntegrationAdminActionRegistryAdmitsDonationCodes fija el registro cerrado
// de F7 (FR-017/R7-10/T706): los 8 códigos `donation.*` que declara
// `platform/audit` coinciden uno a uno con el CHECK de la base, la base no admite
// ningún código de F7 de más, los 8 caben en el registro y un código fuera de él
// se rechaza.
func TestIntegrationAdminActionRegistryAdmitsDonationCodes(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	// El CHECK que se compara es el que define la migración 000013: se aplica
	// dentro de la transacción de prueba porque el estado confirmado de la base
	// puede haber quedado en el de otra funcionalidad.
	execMigrationInTx(t, ctx, tx, "000013_extend_admin_actions_for_donations.up.sql")

	def := constraintDefTx(t, ctx, tx, "admin_actions_action_check", "admin_actions")

	codes := donationActionCodes()
	if len(codes) != 8 {
		t.Fatalf("platform/audit declara %d códigos donation.*, se esperaban 8", len(codes))
	}
	for _, code := range codes {
		if !strings.Contains(def, "'"+code+"'") {
			t.Fatalf("el CHECK de la BD no incluye %q", code)
		}
	}
	if inCheck := strings.Count(def, "'donation."); inCheck != len(codes) {
		t.Fatalf("el CHECK tiene %d códigos donation.* y platform/audit declara %d", inCheck, len(codes))
	}

	// Los 8 códigos CABEN en el registro. Se retira DENTRO de esta transacción la
	// regla del actor (F2: sin actor solo `user.create`); se revierte al terminar
	// y no toca ninguna migración (§VI).
	if _, err := tx.Exec(ctx, `ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check`); err != nil {
		t.Fatalf("retirar temporalmente la regla del actor: %v", err)
	}

	const insert = `INSERT INTO admin_actions (actor_user_id, action, target_kind, target_label, result)
		VALUES (NULL, $1, 'content', 'Donaciones · Registro', 'success')`
	for _, code := range codes {
		if _, err := tx.Exec(ctx, insert, code); err != nil {
			t.Fatalf("la base rechazó el código %q del registro: %v", code, err)
		}
	}

	// Un código fuera del registro se rechaza (fail-closed), dentro de un
	// savepoint para no abortar la transacción de prueba.
	sp, err := tx.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir savepoint: %v", err)
	}
	_, insertErr := sp.Exec(ctx, insert, "donation.archive")
	if rollbackErr := sp.Rollback(ctx); rollbackErr != nil {
		t.Fatalf("revertir el savepoint: %v", rollbackErr)
	}
	if insertErr == nil {
		t.Fatal("la base aceptó un código fuera del registro")
	}
	if !strings.Contains(insertErr.Error(), "admin_actions_action_check") {
		t.Fatalf("el error no viene del CHECK del registro: %v", insertErr)
	}

	if got := countTx(t, ctx, tx,
		`SELECT count(*) FROM admin_actions WHERE target_label = 'Donaciones · Registro'`); got != 8 {
		t.Errorf("filas de auditoría = %d, se esperaban 8", got)
	}
}

// TestIntegrationAdminActionTargetLabelMax fija la forma de la fila de auditoría
// de F7 (FR-017): objetivo `content` (ambas FK en NULL), `target_label`
// descriptivo y el máximo de 254 caracteres del CHECK de F2.
func TestIntegrationAdminActionTargetLabelMax(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	// El CHECK que se ejercita es el que define la migración 000013: se aplica
	// dentro de la transacción de prueba porque el estado confirmado de la base
	// puede haber quedado en el de otra funcionalidad (las pruebas de integración
	// de otros dominios reejecutan sus propios up/down sobre `admin_actions`).
	execMigrationInTx(t, ctx, tx, "000013_extend_admin_actions_for_donations.up.sql")

	// La regla del actor (F2: sin actor solo `user.create`) se retira DENTRO de
	// la transacción de prueba para aislar la forma de la fila de contenido; se
	// revierte al terminar y no toca ninguna migración (§VI).
	if _, err := tx.Exec(ctx, `ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check`); err != nil {
		t.Fatalf("retirar temporalmente la regla del actor: %v", err)
	}

	insert := func(action, label string) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO admin_actions (actor_user_id, action, target_kind, target_label, result)
			 VALUES (NULL, $1, 'content', $2, 'success')`, action, label)
		return err
	}

	for _, code := range donationActionCodes() {
		if err := insert(code, "Donaciones · Banco Nacional · CR05…"); err != nil {
			t.Fatalf("la base rechazó el código %q como content: %v", code, err)
		}
	}

	// 254 caracteres caben; 255 no.
	if err := insert(audit.ActionDonationCreate, strings.Repeat("a", 254)); err != nil {
		t.Errorf("la base rechazó un target_label de 254 caracteres: %v", err)
	}
	err := insertInSavepoint(ctx, tx, "admin_actions", map[string]any{
		"actor_user_id": nil, "action": audit.ActionDonationCreate, "target_kind": "content",
		"target_label": strings.Repeat("a", 255), "result": "success",
	})
	if err == nil {
		t.Fatal("la base aceptó un target_label de 255 caracteres")
	}
	if got := constraintName(err); got != "admin_actions_target_label_check" {
		t.Errorf("restricción = %q, se esperaba admin_actions_target_label_check", got)
	}

	// Las dos FK de un objetivo `content` van en NULL (coherencia del objetivo).
	err = insertInSavepoint(ctx, tx, "admin_actions", map[string]any{
		"actor_user_id": nil, "action": audit.ActionDonationCreate, "target_kind": "content",
		"target_user_id": uuid.NewString(), "target_label": "Donaciones · X", "result": "success",
	})
	if err == nil {
		t.Fatal("la base aceptó un objetivo content con FK de usuario")
	}
	if got := constraintName(err); got != "admin_actions_target_coherence_check" {
		t.Errorf("restricción = %q, se esperaba admin_actions_target_coherence_check", got)
	}
}
