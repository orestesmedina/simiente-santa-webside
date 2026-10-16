//go:build integration

package donaciones

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración de la capa de datos de F7 contra PostgreSQL real
// (T711): la verificación ejecutable de `data-model.md` §«Validación del
// modelo» en lo que toca al esquema.
//
//   1. las migraciones `000012` y `000013` aplican up → down → up sin errores
//      (incluida la restauración del CHECK de `admin_actions`);
//   2. el catálogo de `permissions` sigue en 10 filas con `donaciones` entre
//      ellas (F7 no lo toca; los `down` no lo rozan);
//   3. los `CHECK` de las tablas se disparan en la BD y su nombre coincide con
//      el que traduce el repository al detalle por campo (T708).
//
// Aislamiento: cada prueba trabaja dentro de una TRANSACCIÓN que se revierte al
// terminar (el DDL de PostgreSQL es transaccional), así que los `down` —que son
// destructivos a propósito— no tocan el estado que ven las pruebas de
// integración de los otros dominios, que corren en paralelo. La atomicidad de
// mutación + auditoría y el registro cerrado de códigos se verifican en
// repository_audit_integration_test.go.

// execMigrationInTx ejecuta un archivo .sql completo de migración dentro de la
// transacción indicada.
func execMigrationInTx(t *testing.T, ctx context.Context, tx pgx.Tx, name string) {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", name)
	sql, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leer %s: %v", name, err)
	}
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		t.Fatalf("ejecutar %s: %v", name, err)
	}
}

// testTx abre una transacción de prueba que se revierte al terminar.
func testTx(t *testing.T, ctx context.Context) (*pgxpool.Pool, pgx.Tx) {
	t.Helper()
	pool := testutil.Pool(t, ctx)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir transacción de prueba: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	return pool, tx
}

// constraintDefTx devuelve la definición textual de una restricción.
func constraintDefTx(t *testing.T, ctx context.Context, tx pgx.Tx, name, table string) string {
	t.Helper()
	var def string
	if err := tx.QueryRow(ctx, `SELECT pg_get_constraintdef(oid) FROM pg_constraint
		WHERE conname = $1 AND conrelid = $2::regclass`, name, table).Scan(&def); err != nil {
		t.Fatalf("definición de la restricción %s: %v", name, err)
	}
	return def
}

// countTx cuenta filas con una consulta fija.
func countTx(t *testing.T, ctx context.Context, tx pgx.Tx, query string, args ...any) int {
	t.Helper()
	var total int
	if err := tx.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		t.Fatalf("contar (%s): %v", query, err)
	}
	return total
}

// tableExistsTx indica si una tabla existe en la transacción.
func tableExistsTx(t *testing.T, ctx context.Context, tx pgx.Tx, table string) bool {
	t.Helper()
	return countTx(t, ctx, tx,
		`SELECT count(*) FROM pg_class WHERE relname = $1 AND relkind = 'r'`, table) != 0
}

// insertRawDonationTx inserta en la tabla indicada las columnas del mapa
// (nombres literales del esquema) con valores parametrizados y devuelve el error
// de PostgreSQL sin abortar la prueba: las columnas que no aparecen toman su
// DEFAULT (o NULL), que es justo lo que algunas pruebas necesitan.
func insertRawDonationTx(ctx context.Context, tx pgx.Tx, table string, values map[string]any) error {
	columns := make([]string, 0, len(values))
	for name := range values {
		columns = append(columns, name)
	}
	sort.Strings(columns)

	placeholders := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	for i, name := range columns {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, values[name])
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	_, err := tx.Exec(ctx, query, args...)
	return err
}

// insertInSavepoint ejecuta el INSERT dentro de un savepoint y lo revierte
// siempre: en PostgreSQL un error aborta la transacción entera, así que sin el
// savepoint el primer caso inválido dejaría la transacción de prueba inservible
// para los siguientes.
func insertInSavepoint(ctx context.Context, tx pgx.Tx, table string, values map[string]any) error {
	sp, err := tx.Begin(ctx)
	if err != nil {
		return err
	}
	insertErr := insertRawDonationTx(ctx, sp, table, values)
	if rollbackErr := sp.Rollback(ctx); rollbackErr != nil && insertErr == nil {
		return rollbackErr
	}
	return insertErr
}

// constraintName devuelve el nombre de la restricción que rompió un INSERT.
func constraintName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// sqlState devuelve el código SQLSTATE del error de PostgreSQL ("" si no lo es).
func sqlState(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// TestIntegrationMigration000012UpDownUp aplica
// `000012_create_donation_content` up → down → up y comprueba que las tablas
// quedan como las describe `data-model.md`: columnas obligatorias, defaults de
// estado y orden, y los índices del orden y del filtro público.
func TestIntegrationMigration000012UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	const (
		up   = "000012_create_donation_content.up.sql"
		down = "000012_create_donation_content.down.sql"
	)

	// Punto de partida normalizado: sin las tablas (la migración ya está
	// aplicada en la base de pruebas).
	execMigrationInTx(t, ctx, tx, down)
	if tableExistsTx(t, ctx, tx, "donation_methods") || tableExistsTx(t, ctx, tx, "donation_texts") {
		t.Fatal("las tablas siguen existiendo tras el down")
	}

	execMigrationInTx(t, ctx, tx, up) // 1) up
	assertDonationSchema(t, ctx, tx)

	execMigrationInTx(t, ctx, tx, down) // 2) down
	if tableExistsTx(t, ctx, tx, "donation_methods") || tableExistsTx(t, ctx, tx, "donation_texts") {
		t.Fatal("las tablas siguen existiendo tras el segundo down")
	}

	execMigrationInTx(t, ctx, tx, up) // 3) up
	assertDonationSchema(t, ctx, tx)
}

// assertDonationSchema comprueba el estado esperado de las tablas de F7.
func assertDonationSchema(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	// Obligatorios de un método (FR-014/FR-019): kind, number, institution.
	for _, column := range []string{"kind", "number", "institution"} {
		var nullable string
		if err := tx.QueryRow(ctx, `SELECT is_nullable FROM information_schema.columns
			WHERE table_name = 'donation_methods' AND column_name = $1`, column).Scan(&nullable); err != nil {
			t.Fatalf("leer is_nullable de %s: %v", column, err)
		}
		if nullable != "NO" {
			t.Errorf("%s es anulable, se esperaba NOT NULL", column)
		}
	}

	// El estado de publicación nace en borrador y el orden en 0.
	var stateDefault, sortDefault *string
	if err := tx.QueryRow(ctx, `SELECT
			(SELECT column_default FROM information_schema.columns
			  WHERE table_name = 'donation_methods' AND column_name = 'publication_state'),
			(SELECT column_default FROM information_schema.columns
			  WHERE table_name = 'donation_methods' AND column_name = 'sort_order')`).Scan(&stateDefault, &sortDefault); err != nil {
		t.Fatalf("leer los defaults de estado y orden: %v", err)
	}
	if stateDefault == nil || !strings.Contains(*stateDefault, "draft") {
		t.Errorf("publication_state DEFAULT = %v, se esperaba 'draft'", stateDefault)
	}
	if sortDefault == nil || !strings.HasPrefix(*sortDefault, "0") {
		t.Errorf("sort_order DEFAULT = %v, se esperaba 0", sortDefault)
	}

	// Los dos índices del módulo: el orden público/panel y el filtro por estado.
	indexes := countTx(t, ctx, tx, `SELECT count(*) FROM pg_indexes
		WHERE tablename = 'donation_methods'
		  AND indexname IN ('donation_methods_kind_sort_order_id_idx', 'donation_methods_publication_state_idx')`)
	if indexes != 2 {
		t.Errorf("índices de donation_methods = %d, se esperaban 2 (orden y estado)", indexes)
	}

	// `donation_texts` tiene el `UNIQUE (text_key)` del singleton por clave.
	if uniques := countTx(t, ctx, tx,
		`SELECT count(*) FROM pg_constraint WHERE conrelid = 'donation_texts'::regclass AND contype = 'u'`); uniques != 1 {
		t.Errorf("UNIQUE de donation_texts = %d, se esperaba 1 (text_key)", uniques)
	}
}

// TestIntegrationMigration000013UpDownUp aplica
// `000013_extend_admin_actions_for_donations` up → down → up y comprueba que el
// registro cerrado de códigos gana los 8 `donation.*` y que el `down` restaura
// el de F6 sin tocar los códigos de F2/F3/F4/F6.
func TestIntegrationMigration000013UpDownUp(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	const (
		up   = "000013_extend_admin_actions_for_donations.up.sql"
		down = "000013_extend_admin_actions_for_donations.down.sql"
	)

	// Punto de partida normalizado: el CHECK previo a F7.
	execMigrationInTx(t, ctx, tx, down)
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 1) up
	assertActionCheck(t, ctx, tx, true)

	execMigrationInTx(t, ctx, tx, down) // 2) down
	assertActionCheck(t, ctx, tx, false)

	execMigrationInTx(t, ctx, tx, up) // 3) up
	assertActionCheck(t, ctx, tx, true)

	// El tipo de objetivo NO cambia: los métodos y los textos son `content`.
	kindDef := constraintDefTx(t, ctx, tx, "admin_actions_target_kind_check", "admin_actions")
	for _, kind := range []string{"'user'", "'role'", "'content'"} {
		if !strings.Contains(kindDef, kind) {
			t.Fatalf("target_kind no admite %s: %s", kind, kindDef)
		}
	}
}

// assertActionCheck comprueba que el CHECK de códigos incluye (o no) los 8
// códigos de F7 y que los de F2/F3/F4/F6 siguen intactos (regresión).
func assertActionCheck(t *testing.T, ctx context.Context, tx pgx.Tx, wantF7 bool) {
	t.Helper()
	def := constraintDefTx(t, ctx, tx, "admin_actions_action_check", "admin_actions")

	codes := donationActionCodes()
	if len(codes) != 8 {
		t.Fatalf("platform/audit declara %d códigos donation.*, se esperaban 8", len(codes))
	}
	for _, code := range codes {
		if got := strings.Contains(def, "'"+code+"'"); got != wantF7 {
			t.Fatalf("el CHECK incluye %q = %v, se esperaba %v", code, got, wantF7)
		}
	}
	if inCheck := strings.Count(def, "'donation."); inCheck != 8 && wantF7 {
		t.Fatalf("el CHECK tiene %d códigos donation.*, se esperaban 8", inCheck)
	}
	if inCheck := strings.Count(def, "'donation."); inCheck != 0 && !wantF7 {
		t.Fatalf("el CHECK de F6 tiene %d códigos donation.*, se esperaban 0", inCheck)
	}

	// Los códigos de las funcionalidades anteriores no se pierden.
	for _, code := range []string{"user.create", "home.publish", "event.create", "category.delete", "ministry.publish"} {
		if !strings.Contains(def, "'"+code+"'") {
			t.Fatalf("el CHECK perdió %q: %s", code, def)
		}
	}
}

// TestIntegrationPermissionsCatalogUnchanged fija que F7 **no** toca el catálogo
// de permisos de F2 (FR-011/P7-6): sigue con 10 permisos, `donaciones` entre
// ellos, y los `down` de F7 tampoco lo rozan.
func TestIntegrationPermissionsCatalogUnchanged(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	const catalog = `SELECT count(*) FROM permissions`

	if got := countTx(t, ctx, tx, catalog); got != 10 {
		t.Fatalf("permisos = %d, se esperaban 10", got)
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM permissions WHERE code = 'donaciones'`); got != 1 {
		t.Fatalf("el permiso `donaciones` no está en el catálogo: %d", got)
	}
	var label string
	if err := tx.QueryRow(ctx, `SELECT label FROM permissions WHERE code = 'donaciones'`).Scan(&label); err != nil {
		t.Fatalf("leer la etiqueta de `donaciones`: %v", err)
	}
	if label != "Donaciones" {
		t.Errorf("etiqueta de `donaciones` = %q, se esperaba «Donaciones»", label)
	}

	// Los `down` de F7 no tocan el catálogo.
	execMigrationInTx(t, ctx, tx, "000013_extend_admin_actions_for_donations.down.sql")
	execMigrationInTx(t, ctx, tx, "000012_create_donation_content.down.sql")
	if got := countTx(t, ctx, tx, catalog); got != 10 {
		t.Fatalf("permisos tras los down de F7 = %d, se esperaban 10", got)
	}
	if got := countTx(t, ctx, tx, `SELECT count(*) FROM permissions WHERE code = 'donaciones'`); got != 1 {
		t.Fatalf("los down de F7 retiraron el permiso `donaciones`: %d", got)
	}
}

// TestIntegrationDonationChecksFireInTheDatabase comprueba que cada invariante
// del modelo está protegido por la BD y que el nombre de la restricción que
// PostgreSQL reporta coincide con el que el repository traduce al detalle por
// campo (T708/FR-014/FR-015/FR-016/FR-019).
func TestIntegrationDonationChecksFireInTheDatabase(t *testing.T) {
	ctx := context.Background()
	_, tx := testTx(t, ctx)

	ibanBase := map[string]any{
		"kind": "iban", "number": "CR05015202001234567890",
		"institution": "Banco Nacional", "currency": "CRC",
	}
	savepoint := func(values map[string]any) error {
		return insertInSavepoint(ctx, tx, "donation_methods", values)
	}
	withBase := func(overrides map[string]any) map[string]any {
		values := map[string]any{}
		for k, v := range ibanBase {
			values[k] = v
		}
		for k, v := range overrides {
			values[k] = v
		}
		return values
	}

	methodTests := []struct {
		name       string
		values     map[string]any
		constraint string
		field      string
	}{
		{
			name:       "IBAN con espacios",
			values:     withBase(map[string]any{"number": "CR05 0152 0200 1234 5678 90"}),
			constraint: "donation_methods_check", field: "number",
		},
		{
			name:       "SINPE sin +506",
			values:     withBase(map[string]any{"kind": "sinpe", "number": "88888888", "currency": nil}),
			constraint: "donation_methods_check", field: "number",
		},
		{
			name:       "IBAN sin moneda",
			values:     withBase(map[string]any{"currency": nil}),
			constraint: "donation_methods_check1", field: "currency",
		},
		{
			name:       "SINPE con moneda",
			values:     withBase(map[string]any{"kind": "sinpe", "number": "+50688888888"}),
			constraint: "donation_methods_check1", field: "currency",
		},
		{
			name:       "titular inglés sin español",
			values:     withBase(map[string]any{"holder_en": "Ana"}),
			constraint: "donation_methods_check2", field: "holderEs",
		},
		{
			// Un `kind` fuera del catálogo también incumple el formato por tipo y
			// PostgreSQL reporta ese CHECK (el de `kind` es una defensa redundante
			// no alcanzable por sí sola, como en F6).
			name:       "tipo fuera del catálogo",
			values:     withBase(map[string]any{"kind": "cash"}),
			constraint: "donation_methods_check", field: "number",
		},
		{
			name:       "moneda fuera del catálogo",
			values:     withBase(map[string]any{"currency": "EUR"}),
			constraint: "donation_methods_currency_check", field: "currency",
		},
		{
			name:       "estado fuera del catálogo",
			values:     withBase(map[string]any{"publication_state": "archived"}),
			constraint: "donation_methods_publication_state_check", field: "publicationState",
		},
		{
			name:       "entidad con espacios sobrantes",
			values:     withBase(map[string]any{"institution": " Banco Nacional "}),
			constraint: "donation_methods_institution_check1", field: "institution",
		},
		{
			name:       "titular español con espacios sobrantes",
			values:     withBase(map[string]any{"holder_es": " Ana "}),
			constraint: "donation_methods_holder_es_check1", field: "holderEs",
		},
		{
			name:       "titular inglés con espacios sobrantes",
			values:     withBase(map[string]any{"holder_es": "Ana", "holder_en": " Ana "}),
			constraint: "donation_methods_holder_en_check1", field: "holderEn",
		},
	}

	for _, tt := range methodTests {
		t.Run("metodo/"+tt.name, func(t *testing.T) {
			err := savepoint(tt.values)
			if err == nil {
				t.Fatal("la BD aceptó una fila que viola un invariante")
			}
			if got := constraintName(err); got != tt.constraint {
				t.Fatalf("restricción = %q, se esperaba %q (%v)", got, tt.constraint, err)
			}
			if details := checkDetails(constraintName(err)); details == nil || details[tt.field] == nil {
				t.Fatalf("checkDetails(%q) = %v, se esperaba el campo %q", tt.constraint, details, tt.field)
			}
		})
	}

	textBase := map[string]any{"text_key": "intro", "body_es": "Puedes dar tu ofrenda."}
	textTests := []struct {
		name       string
		values     map[string]any
		constraint string
		field      string
	}{
		{
			name:       "clave fuera del catálogo",
			values:     map[string]any{"text_key": "otro", "body_es": "Texto"},
			constraint: "donation_texts_text_key_check", field: "textKey",
		},
		{
			name:       "cuerpo español vacío",
			values:     map[string]any{"text_key": "usage", "body_es": ""},
			constraint: "donation_texts_body_es_check", field: "textEs",
		},
		{
			name:       "cuerpo español de más de 2.000",
			values:     map[string]any{"text_key": "usage", "body_es": strings.Repeat("x", 2001)},
			constraint: "donation_texts_body_es_check", field: "textEs",
		},
		{
			name:       "estado fuera del catálogo",
			values:     map[string]any{"text_key": "usage", "body_es": "Texto", "publication_state": "archived"},
			constraint: "donation_texts_publication_state_check", field: "publicationState",
		},
	}
	for _, tt := range textTests {
		t.Run("texto/"+tt.name, func(t *testing.T) {
			err := insertInSavepoint(ctx, tx, "donation_texts", tt.values)
			if err == nil {
				t.Fatal("la BD aceptó una fila que viola un invariante")
			}
			if got := constraintName(err); got != tt.constraint {
				t.Fatalf("restricción = %q, se esperaba %q (%v)", got, tt.constraint, err)
			}
			if details := checkDetails(constraintName(err)); details == nil || details[tt.field] == nil {
				t.Fatalf("checkDetails(%q) = %v, se esperaba el campo %q", tt.constraint, details, tt.field)
			}
		})
	}

	// `UNIQUE (text_key)` impide dos renglones de la misma clave: el segundo
	// INSERT es una violación de unicidad (23505), no un CHECK.
	if err := insertRawDonationTx(ctx, tx, "donation_texts", textBase); err != nil {
		t.Fatalf("primer texto: %v", err)
	}
	err := insertInSavepoint(ctx, tx, "donation_texts", map[string]any{
		"text_key": "intro", "body_es": "Otro introductorio",
	})
	if err == nil {
		t.Fatal("la BD aceptó dos textos con la misma clave")
	}
	if got := sqlState(err); got != codeUniqueViolation {
		t.Fatalf("SQLSTATE = %q, se esperaba %q (unique_violation) (%v)", got, codeUniqueViolation, err)
	}

	// Los valores válidos sí entran (defensa en profundidad, no bloqueo).
	if err := savepoint(withBase(map[string]any{"number": "ES9121000418450200051332"})); err != nil {
		t.Errorf("la BD rechazó un IBAN válido: %v", err)
	}
	if err := savepoint(map[string]any{
		"kind": "sinpe", "number": "+50688888888", "institution": "Banco Nacional",
	}); err != nil {
		t.Errorf("la BD rechazó un SINPE válido: %v", err)
	}
}
