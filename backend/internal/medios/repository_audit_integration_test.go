//go:build integration

package medios

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
)

// Pruebas de integración de la auditoría del dominio medios contra PostgreSQL
// real (T914, bloque 4): la mutación y su fila de `admin_actions` comparten
// transacción (FR-022, fail-closed: si el registro falla, la mutación NO se
// aplica) y el registro cerrado de códigos de la BD coincide con
// `platform/audit` (T906/P9-14).
//
// Como en F3/F4/F6/F7, estas pruebas **no** dependen de las tablas `users`/
// `roles` (las truncan en paralelo las pruebas de integración del dominio
// usuarios): el camino de tx del repositorio se ejercita con el único código que
// la BD admite sin actor (`user.create`) y los 6 códigos `media.*` se ejercitan
// dentro de la transacción de la prueba, con la regla del actor retirada
// temporalmente (se revierte al terminar y no toca ninguna migración, §VI).

// mediaActionCodes devuelve los códigos `media.*` del registro cerrado de
// platform/audit (T906), en el orden de la lista.
func mediaActionCodes() []string {
	codes := make([]string, 0, 6)
	for _, code := range audit.ActionCodes {
		if strings.HasPrefix(code, "media.") {
			codes = append(codes, code)
		}
	}
	return codes
}

// auditRowWithoutActor arma la fila de auditoría de una mutación sin actor (el
// único caso que la BD admite así es `user.create`): así la prueba no depende de
// la tabla `users`.
func auditRowWithoutActor(label string) audit.Action {
	return audit.Action{
		Code:        audit.ActionUserCreate,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// auditRowMedia arma la fila de auditoría de una acción de medios sobre una
// ficha.
func auditRowMedia(code, label string) audit.Action {
	return audit.Action{
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// mediaItemDePrueba arma una ficha válida para las mutaciones de la prueba.
func mediaItemDePrueba(titulo string) Episodio {
	return Episodio{
		Titulo:      titulo,
		PublicadoEn: time.Now(),
		Origen:      OrigenVideo,
		Visibilidad: VisibilidadVisible,
	}
}

// countFichas cuenta las fichas con SQL directo (independiente del código bajo
// prueba) para comprobar que nada se guardó a medias.
func countFichas(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var total int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM media_items`).Scan(&total); err != nil {
		t.Fatalf("contar fichas: %v", err)
	}
	return total
}

// countAdminActionsByLabel cuenta las filas de auditoría con una etiqueta exacta
// (la tabla es compartida y las pruebas de integración de otros dominios cuentan
// sus propias filas).
func countAdminActionsByLabel(t *testing.T, pool *pgxpool.Pool, label string) int {
	t.Helper()
	var total int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM admin_actions WHERE target_label = $1`, label).Scan(&total); err != nil {
		t.Fatalf("contar acciones administrativas: %v", err)
	}
	return total
}

// cleanupAdminActionsByLabel borra al terminar la prueba las filas de auditoría
// con la etiqueta indicada, para no ensuciar una tabla compartida.
func cleanupAdminActionsByLabel(t *testing.T, pool *pgxpool.Pool, label string) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(),
			`DELETE FROM admin_actions WHERE target_label = $1`, label); err != nil {
			t.Logf("no se pudieron limpiar las filas de auditoría de la prueba: %v", err)
		}
	})
}

// TestIntegrationMutationAndAuditCommitTogether comprueba el camino normal del
// repositorio: el alta de una ficha y su fila de auditoría quedan escritas, y la
// mutación y el registro comparten transacción (si el registro no cabe, nada se
// escribe).
func TestIntegrationMutationAndAuditCommitTogether(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	label := "Medios · tx de prueba " + uuid.NewString()
	cleanupAdminActionsByLabel(t, repo.pool, label)

	guardada, err := repo.InsertMediaItem(ctx, mediaItemDePrueba("Sermón con auditoría"), auditRowWithoutActor(label))
	if err != nil {
		t.Fatalf("InsertMediaItem con auditoría: %v", err)
	}
	if got := countFichas(t, repo.pool); got != 1 {
		t.Errorf("fichas = %d, se esperaba 1", got)
	}
	if got := countAdminActionsByLabel(t, repo.pool, label); got != 1 {
		t.Errorf("filas de auditoría = %d, se esperaba 1", got)
	}

	// La mutación por el repositorio (fila completa ya fusionada) también deja
	// su registro en la misma transacción.
	oculta := guardada
	oculta.Visibilidad = VisibilidadOculta
	if _, err := repo.UpdateMediaItem(ctx, oculta, auditRowWithoutActor(label)); err != nil {
		t.Fatalf("UpdateMediaItem con auditoría: %v", err)
	}
	if got := countAdminActionsByLabel(t, repo.pool, label); got != 2 {
		t.Errorf("filas de auditoría = %d, se esperaban 2", got)
	}
}

// TestIntegrationWriteTxRollsBackWhenAuditFails fija el edge case de FR-022: si
// el registro de auditoría falla en la base (aquí, un `target_label` que supera
// el CHECK de 254 caracteres), la mutación NO se aplica.
func TestIntegrationWriteTxRollsBackWhenAuditFails(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	demasiadoLargo := strings.Repeat("x", 300)
	err := repo.withTx(ctx, func(tx *repository) error {
		if _, err := tx.q.InsertMediaItem(ctx, insertMediaItemParams(mediaItemDePrueba("Sermón sin registro"))); err != nil {
			return err
		}
		return insertActions(ctx, tx, []audit.Action{auditRowWithoutActor(demasiadoLargo)})
	})
	if err == nil {
		t.Fatal("withTx no devolvió error con un registro inválido")
	}
	if kind := errorKind(err); kind != apperr.KindInvalid {
		t.Errorf("kind = %v, se esperaba invalid (CHECK de la base)", kind)
	}

	if got := countFichas(t, repo.pool); got != 0 {
		t.Errorf("la mutación se aplicó sin registro: %d fichas", got)
	}
}

// TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown comprueba el otro lado
// del fail-closed: un código fuera del registro cerrado de `platform/audit` se
// rechaza ANTES de tocar la base y la mutación tampoco se aplica.
func TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	err := repo.withTx(ctx, func(tx *repository) error {
		if _, err := tx.q.InsertMediaItem(ctx, insertMediaItemParams(mediaItemDePrueba("Sermón sin registro"))); err != nil {
			return err
		}
		return insertActions(ctx, tx, []audit.Action{auditRowMedia("media.archive", "Medios · X")})
	})
	if err == nil {
		t.Fatal("withTx no devolvió error con un código desconocido")
	}
	if !strings.Contains(err.Error(), audit.ErrUnknownActionCode.Error()) {
		t.Errorf("el error no identifica el registro cerrado: %v", err)
	}

	if got := countFichas(t, repo.pool); got != 0 {
		t.Errorf("la mutación se aplicó sin registro válido: %d fichas", got)
	}
}

// TestIntegrationMediaMutationsRecordTheirAuditRow fija la forma de la fila de
// auditoría de cada uno de los 6 códigos `media.*` sobre una mutación de la
// ficha: objetivo `content` (ambas FK en NULL), `target_label` descriptivo y
// ≤ 254 caracteres, y mutación + registro en la MISMA transacción.
//
// Se trabaja dentro de la transacción de la prueba (que se revierte) con la
// regla del actor de F2 retirada temporalmente: la tabla `users` la truncan en
// paralelo las pruebas del dominio usuarios, así que las filas `media.*` no
// pueden llevar actor real aquí (el actor real lo ejercita el service, T918/T921).
func TestIntegrationMediaMutationsRecordTheirAuditRow(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	if _, err := tx.Exec(ctx, `ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check`); err != nil {
		t.Fatalf("retirar temporalmente la regla del actor: %v", err)
	}

	// El repositorio ligado a la transacción de la prueba: la mutación y su
	// registro van juntos y todo se revierte al terminar.
	repo := &repository{q: gendb.New(tx)}

	codes := mediaActionCodes()
	if len(codes) != 6 {
		t.Fatalf("platform/audit declara %d códigos media.*, se esperaban 6", len(codes))
	}
	for _, code := range codes {
		titulo := "Sermón " + code
		row, err := repo.q.InsertMediaItem(ctx, insertMediaItemParams(mediaItemDePrueba(titulo)))
		if err != nil {
			t.Fatalf("InsertMediaItem (%s): %v", code, err)
		}
		label := "Medios · " + row.Title
		if err := insertActions(ctx, repo, []audit.Action{auditRowMedia(code, label)}); err != nil {
			t.Fatalf("registrar %s: %v", code, err)
		}
	}

	if got := countTx(t, ctx, tx, `SELECT count(*) FROM media_items`); got != 6 {
		t.Errorf("fichas = %d, se esperaban 6", got)
	}
	if got := countTx(t, ctx, tx,
		`SELECT count(*) FROM admin_actions WHERE target_kind = 'content'
		   AND target_user_id IS NULL AND target_role_id IS NULL
		   AND target_label LIKE 'Medios · %' AND char_length(target_label) <= 254`); got != 6 {
		t.Errorf("filas de auditoría bien formadas = %d, se esperaban 6", got)
	}

	// Fail-closed dentro de la misma transacción: un registro que no cabe en el
	// CHECK del target_label deja la mutación revertida. El error aborta la
	// transacción, así que el caso se mide con un savepoint.
	sp, err := tx.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir savepoint: %v", err)
	}
	_, insertErr := sp.Exec(ctx, `INSERT INTO media_items (title, published_at, origin) VALUES ($1, now(), 'video')`,
		"Sermón sin registro")
	if insertErr == nil {
		_, insertErr = sp.Exec(ctx, `INSERT INTO admin_actions (actor_user_id, action, target_kind, target_label, result)
			VALUES (NULL, $1, 'content', $2, 'success')`, audit.ActionMediaSync, strings.Repeat("x", 300))
	}
	if rollbackErr := sp.Rollback(ctx); rollbackErr != nil {
		t.Fatalf("revertir el savepoint: %v", rollbackErr)
	}
	if insertErr == nil {
		t.Fatal("la BD aceptó un registro con target_label de 300 caracteres")
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM media_items WHERE title = 'Sermón sin registro'`); got != 0 {
		t.Errorf("la mutación sobrevivió al fallo del registro: %d fichas", got)
	}
}

// TestIntegrationAdminActionRegistryAdmitsMediaCodes fija el registro cerrado de
// F9 (P9-14/T906): los 6 códigos `media.*` que declara `platform/audit`
// coinciden uno a uno con el CHECK de la base, los 6 caben en el registro y un
// código fuera de él se rechaza.
func TestIntegrationAdminActionRegistryAdmitsMediaCodes(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	// El CHECK que se compara es el que define la migración 000015: se aplica
	// dentro de la transacción de prueba porque el estado confirmado de la base
	// puede haber quedado en el de otra funcionalidad.
	execMigrationInTx(t, ctx, tx, "000015_extend_admin_actions_for_media.up.sql")

	def := constraintDefTx(t, ctx, tx, "admin_actions_action_check", "admin_actions")

	codes := mediaActionCodes()
	if len(codes) != 6 {
		t.Fatalf("platform/audit declara %d códigos media.*, se esperaban 6", len(codes))
	}
	for _, code := range codes {
		if !strings.Contains(def, "'"+code+"'") {
			t.Fatalf("el CHECK de la BD no incluye %q", code)
		}
	}
	if inCheck := strings.Count(def, "'media."); inCheck != len(codes) {
		t.Fatalf("el CHECK tiene %d códigos media.* y platform/audit declara %d", inCheck, len(codes))
	}

	// Los 6 códigos CABEN en el registro. Se retira DENTRO de esta transacción la
	// regla del actor (F2: sin actor solo `user.create`); se revierte al terminar
	// y no toca ninguna migración (§VI).
	if _, err := tx.Exec(ctx, `ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check`); err != nil {
		t.Fatalf("retirar temporalmente la regla del actor: %v", err)
	}

	const insert = `INSERT INTO admin_actions (actor_user_id, action, target_kind, target_label, result)
		VALUES (NULL, $1, 'content', 'Medios · Registro', 'success')`
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
	_, insertErr := sp.Exec(ctx, insert, "media.archive")
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
		`SELECT count(*) FROM admin_actions WHERE target_label = 'Medios · Registro'`); got != 6 {
		t.Errorf("filas de auditoría = %d, se esperaban 6", got)
	}
}

// TestIntegrationAdminActionTargetLabelMax fija la forma de la fila de auditoría
// de F9 (FR-022): objetivo `content` (ambas FK en NULL), `target_label`
// descriptivo y el máximo de 254 caracteres del CHECK de F2.
func TestIntegrationAdminActionTargetLabelMax(t *testing.T) {
	ctx := context.Background()
	_, tx := migrationTx(t, ctx)

	execMigrationInTx(t, ctx, tx, "000015_extend_admin_actions_for_media.up.sql")

	// La regla del actor (F2: sin actor solo `user.create`) se retira DENTRO de
	// la transacción de prueba para aislar la forma de la fila de contenido.
	if _, err := tx.Exec(ctx, `ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check`); err != nil {
		t.Fatalf("retirar temporalmente la regla del actor: %v", err)
	}

	insert := func(action, label string) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO admin_actions (actor_user_id, action, target_kind, target_label, result)
			 VALUES (NULL, $1, 'content', $2, 'success')`, action, label)
		return err
	}

	for _, code := range mediaActionCodes() {
		if err := insert(code, "Medios · Potencial sin Resultados"); err != nil {
			t.Fatalf("la base rechazó el código %q como content: %v", code, err)
		}
	}

	// 254 caracteres caben; 255 no.
	if err := insert(audit.ActionMediaSync, strings.Repeat("a", 254)); err != nil {
		t.Errorf("la base rechazó un target_label de 254 caracteres: %v", err)
	}
	err := insertInSavepoint(ctx, tx, "admin_actions", map[string]any{
		"actor_user_id": nil, "action": audit.ActionMediaSync, "target_kind": "content",
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
		"actor_user_id": nil, "action": audit.ActionMediaSync, "target_kind": "content",
		"target_user_id": uuid.NewString(), "target_label": "Medios · X", "result": "success",
	})
	if err == nil {
		t.Fatal("la base aceptó un objetivo content con FK de usuario")
	}
	if got := constraintName(err); got != "admin_actions_target_coherence_check" {
		t.Errorf("restricción = %q, se esperaba admin_actions_target_coherence_check", got)
	}
}
