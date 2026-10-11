//go:build integration

package agenda

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración del dominio agenda contra PostgreSQL real (F4).
// Requieren DATABASE_URL_TEST con las migraciones aplicadas (lo aporta el CI o
// el compose local; si no existe, se omiten solas). Estas pruebas truncan las
// tablas `agenda_*`: DATABASE_URL_TEST debe apuntar a una base de pruebas.
//
// NO dependen de las tablas `users`/`roles` (las truncan en paralelo las pruebas
// del dominio usuarios): la fila de auditoría se ejercita por el único camino
// sin actor que la BD admite (`action='user.create'` con `target_kind='content'`,
// igual que la prueba de migraciones de F3).

// newAgendaRepo construye el repositorio contra la base de pruebas y vacía las
// tablas de agenda para partir de un estado conocido.
func newAgendaRepo(t *testing.T) (*repository, *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	const truncate = `TRUNCATE agenda_activities, agenda_event_dates, agenda_events, agenda_categories`
	if _, err := pool.Exec(ctx, truncate); err != nil {
		t.Fatalf("vaciar tablas de agenda: %v", err)
	}
	return NewRepository(pool), pool
}

// countCategoriesByName cuenta las categorías con un nombre exacto (una consulta
// fija y parametrizada: nunca SQL construido con datos variables).
func countCategoriesByName(t *testing.T, pool *pgxpool.Pool, name string) int {
	t.Helper()
	var total int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM agenda_categories WHERE name_es = $1`, name).Scan(&total); err != nil {
		t.Fatalf("contar categorías: %v", err)
	}
	return total
}

// countAdminActionsByLabel cuenta las filas de auditoría con una etiqueta exacta.
func countAdminActionsByLabel(t *testing.T, pool *pgxpool.Pool, label string) int {
	t.Helper()
	var total int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM admin_actions WHERE target_label = $1`, label).Scan(&total); err != nil {
		t.Fatalf("contar acciones administrativas: %v", err)
	}
	return total
}

// auditRowWithoutActor arma la fila de auditoría de una mutación de contenido.
// Usa `user.create` porque es el único código que la BD admite sin actor
// (`CHECK (actor_user_id IS NOT NULL OR action = 'user.create')`, F2): así la
// prueba no depende de la tabla `users`.
func auditRowWithoutActor(label string) audit.Action {
	return audit.Action{
		Code:        audit.ActionUserCreate,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// insertTestCategory inserta una categoría por la consulta generada (la misma
// que usará el repository de categorías en T414).
func insertTestCategory(t *testing.T, ctx context.Context, tx *repository, name string) error {
	t.Helper()
	_, err := tx.q.InsertCategory(ctx, gendb.InsertCategoryParams{
		NameEs:    name,
		NameEn:    pgtype.Text{},
		SortOrder: 0,
	})
	return err
}

// TestIntegrationWriteTxCommitsMutationAndAudit verifica la base de escritura de
// T411: la mutación y su fila de auditoría comparten la MISMA transacción y, si
// todo va bien, quedan ambas (FR-017).
func TestIntegrationWriteTxCommitsMutationAndAudit(t *testing.T) {
	repo, pool := newAgendaRepo(t)
	ctx := context.Background()

	name := "Categoría de tx " + uuid.NewString()
	label := "Eventos · tx de prueba " + uuid.NewString()

	err := repo.withTx(ctx, func(tx *repository) error {
		if err := insertTestCategory(t, ctx, tx, name); err != nil {
			return err
		}
		return insertActions(ctx, tx, []audit.Action{auditRowWithoutActor(label)})
	})
	if err != nil {
		t.Fatalf("withTx: %v", err)
	}

	if got := countCategoriesByName(t, pool, name); got != 1 {
		t.Errorf("categorías con el nombre de la prueba = %d, se esperaba 1", got)
	}
	if got := countAdminActionsByLabel(t, pool, label); got != 1 {
		t.Errorf("filas de auditoría = %d, se esperaba 1", got)
	}
}

// TestIntegrationWriteTxRollsBackWhenAuditFails fija el edge case de FR-017: si
// el registro de auditoría falla en la base (aquí, un `target_label` que supera
// el CHECK de 254 caracteres), la mutación NO se aplica.
func TestIntegrationWriteTxRollsBackWhenAuditFails(t *testing.T) {
	repo, pool := newAgendaRepo(t)
	ctx := context.Background()

	name := "Categoría que no debe quedar " + uuid.NewString()
	tooLong := strings.Repeat("x", 300)

	err := repo.withTx(ctx, func(tx *repository) error {
		if err := insertTestCategory(t, ctx, tx, name); err != nil {
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

	if got := countCategoriesByName(t, pool, name); got != 0 {
		t.Errorf("la mutación se aplicó sin registro: %d categorías", got)
	}
	if got := countAdminActionsByLabel(t, pool, tooLong); got != 0 {
		t.Errorf("quedó una fila de auditoría inválida: %d", got)
	}
}

// TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown comprueba el otro lado
// del fail-closed: un código fuera del registro cerrado de `platform/audit` se
// rechaza ANTES de tocar la base y la mutación tampoco se aplica.
func TestIntegrationWriteTxRollsBackWhenAuditCodeIsUnknown(t *testing.T) {
	repo, pool := newAgendaRepo(t)
	ctx := context.Background()

	name := "Categoría con código desconocido " + uuid.NewString()

	err := repo.withTx(ctx, func(tx *repository) error {
		if err := insertTestCategory(t, ctx, tx, name); err != nil {
			return err
		}
		return insertActions(ctx, tx, []audit.Action{{
			Code:        "event.archive", // fuera del registro cerrado (R4-10)
			TargetKind:  audit.TargetContent,
			TargetLabel: "Eventos · X",
			Result:      audit.ResultSuccess,
		}})
	})
	if err == nil {
		t.Fatal("withTx no devolvió error con un código desconocido")
	}
	if !strings.Contains(err.Error(), audit.ErrUnknownActionCode.Error()) {
		t.Errorf("el error no identifica el registro cerrado: %v", err)
	}
	if got := countCategoriesByName(t, pool, name); got != 0 {
		t.Errorf("la mutación se aplicó sin registro válido: %d categorías", got)
	}
}
