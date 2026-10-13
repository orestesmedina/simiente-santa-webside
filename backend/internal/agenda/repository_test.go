package agenda

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/database"
)

// text devuelve un pgtype.Text válido (columna con valor).
func text(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }

// textOf es un atajo para construir un *string.
func textOf(s string) *string { return &s }

// errorKind clasifica el error de dominio (mismo helper que usan las pruebas de
// los otros dominios: la traducción a HTTP vive solo en httpserver).
func errorKind(err error) apperr.Kind {
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		return domainErr.Kind
	}
	return apperr.KindInternal
}

// TestUUIDHelpers verifica la traducción de UUID anulables (§8.1.6).
func TestUUIDHelpers(t *testing.T) {
	id := uuid.New()
	if got := uuidValue(pgtype.UUID{Bytes: [16]byte(id), Valid: true}); got != id {
		t.Errorf("uuidValue = %v, se esperaba %v", got, id)
	}
	if got := uuidPtr(pgtype.UUID{}); got != nil {
		t.Errorf("uuidPtr(NULL) = %v, se esperaba nil", got)
	}
	if got := uuidPtr(pgtype.UUID{Bytes: [16]byte(id), Valid: true}); got == nil || *got != id {
		t.Errorf("uuidPtr(válido) = %v", got)
	}
	if got := pgUUID(id); !got.Valid || uuid.UUID(got.Bytes) != id {
		t.Errorf("pgUUID = %+v", got)
	}
	if got := pgUUIDPtr(nil); got.Valid {
		t.Errorf("pgUUIDPtr(nil) = %+v, se esperaba NULL", got)
	}
	if got := pgUUIDPtr(&id); !got.Valid || uuid.UUID(got.Bytes) != id {
		t.Errorf("pgUUIDPtr(válido) = %+v", got)
	}
}

// TestTextHelpers verifica la traducción de textos opcionales: "" (y solo
// espacios) se guarda como NULL (analyze I6) y un valor no vacío como TEXT.
func TestTextHelpers(t *testing.T) {
	if got := textPtr(pgtype.Text{}); got != nil {
		t.Errorf("textPtr(NULL) = %v, se esperaba nil", got)
	}
	if got := textPtr(text("hola")); got == nil || *got != "hola" {
		t.Errorf("textPtr(válido) = %v", got)
	}
	if got := pgTextPtr(nil); got.Valid {
		t.Errorf("pgTextPtr(nil) = %+v, se esperaba NULL", got)
	}
	if got := pgTextPtr(textOf("x")); !got.Valid || got.String != "x" {
		t.Errorf("pgTextPtr = %+v", got)
	}
	if got := pgText(""); got.Valid {
		t.Errorf("pgText(\"\") = %+v, se esperaba NULL", got)
	}
	if got := pgText("   "); got.Valid {
		t.Errorf("pgText(espacios) = %+v, se esperaba NULL", got)
	}
	if got := pgText("Culto"); !got.Valid || got.String != "Culto" {
		t.Errorf("pgText(válido) = %+v", got)
	}
	if got := pgLabel(""); got.Valid {
		t.Errorf("pgLabel(\"\") = %+v, se esperaba NULL", got)
	}
	if got := pgLabel("Eventos · Noche de aposentos"); !got.Valid {
		t.Errorf("pgLabel(válido) = %+v", got)
	}
}

// TestDateHelpers verifica la traducción de fechas de calendario: el dominio
// usa "YYYY-MM-DD" y no convierte de zona horaria (R4-3).
func TestDateHelpers(t *testing.T) {
	got := pgDate("2026-10-17")
	if !got.Valid {
		t.Fatalf("pgDate(válida) = %+v, se esperaba válida", got)
	}
	if formatted := got.Time.Format(DateFormat); formatted != "2026-10-17" {
		t.Errorf("pgDate guarda %q, se esperaba 2026-10-17", formatted)
	}
	if formatted := dateString(got); formatted != "2026-10-17" {
		t.Errorf("dateString = %q, se esperaba 2026-10-17", formatted)
	}

	for _, invalid := range []string{"", "  ", "2026-13-32", "17/10/2026", "2026-10-17T00:00:00Z"} {
		if v := pgDate(invalid); v.Valid {
			t.Errorf("pgDate(%q) = %+v, se esperaba NULL", invalid, v)
		}
	}
	if got := dateString(pgtype.Date{}); got != "" {
		t.Errorf("dateString(NULL) = %q, se esperaba vacío", got)
	}
}

// TestTimeHelpers verifica la traducción de horas "HH:MM" al parámetro TIME de
// PostgreSQL (microsegundos desde la medianoche).
func TestTimeHelpers(t *testing.T) {
	tests := []struct {
		clock string
		want  int64
		valid bool
	}{
		{clock: "00:00", want: 0, valid: true},
		{clock: "07:30", want: (7*3600 + 30*60) * 1_000_000, valid: true},
		{clock: "23:59", want: (23*3600 + 59*60) * 1_000_000, valid: true},
		{clock: "", valid: false},
		{clock: "24:00", valid: false},
		{clock: "7:30", valid: false},
		{clock: "07:60", valid: false},
	}
	for _, tt := range tests {
		got := pgTime(tt.clock)
		if got.Valid != tt.valid {
			t.Errorf("pgTime(%q).Valid = %v, se esperaba %v", tt.clock, got.Valid, tt.valid)
			continue
		}
		if tt.valid && got.Microseconds != tt.want {
			t.Errorf("pgTime(%q) = %d microsegundos, se esperaban %d", tt.clock, got.Microseconds, tt.want)
		}
	}
}

// TestPublicationStateValid fija el registro cerrado de estados (FR-013).
func TestPublicationStateValid(t *testing.T) {
	tests := []struct {
		state PublicationState
		want  bool
	}{
		{StateDraft, true},
		{StatePublished, true},
		{PublicationState("archived"), false},
		{PublicationState(""), false},
	}
	for _, tt := range tests {
		if got := tt.state.Valid(); got != tt.want {
			t.Errorf("PublicationState(%q).Valid() = %v, se esperaba %v", tt.state, got, tt.want)
		}
	}
}

// TestAuditParamsMapsContentAction verifica el mapeo audit.Action →
// InsertAdminActionParams (R4-10): un objetivo `content` no lleva FK y conserva
// la etiqueta; una etiqueta vacía se guarda como NULL.
func TestAuditParamsMapsContentAction(t *testing.T) {
	actor := uuid.New()
	params := auditParams(audit.Action{
		ActorUserID: &actor,
		Code:        audit.ActionEventPublish,
		TargetKind:  audit.TargetContent,
		TargetLabel: "Eventos · Noche de aposentos",
		Result:      audit.ResultSuccess,
	})

	if !params.ActorUserID.Valid || uuid.UUID(params.ActorUserID.Bytes) != actor {
		t.Errorf("ActorUserID = %+v", params.ActorUserID)
	}
	if params.Action != audit.ActionEventPublish {
		t.Errorf("Action = %q", params.Action)
	}
	if params.TargetKind != "content" {
		t.Errorf("TargetKind = %q, se esperaba content", params.TargetKind)
	}
	if params.TargetUserID.Valid || params.TargetRoleID.Valid {
		t.Errorf("las FK de contenido deben ir NULL: %+v / %+v", params.TargetUserID, params.TargetRoleID)
	}
	if !params.TargetLabel.Valid || params.TargetLabel.String != "Eventos · Noche de aposentos" {
		t.Errorf("TargetLabel = %+v", params.TargetLabel)
	}
	if params.Result != "success" {
		t.Errorf("Result = %q", params.Result)
	}

	// Etiqueta vacía (no es el caso de `content`, pero el mapeo es genérico).
	empty := auditParams(audit.Action{TargetKind: audit.TargetContent})
	if empty.TargetLabel.Valid {
		t.Errorf("TargetLabel vacío = %+v, se esperaba NULL", empty.TargetLabel)
	}
}

// TestInsertActionsValidatesBeforeTouchingTheDatabase fija el fail-closed del
// registro de auditoría (FR-017): un código fuera del registro cerrado se
// rechaza ANTES de ejecutar SQL —el `tx` es nil a propósito: si `insertActions`
// intentara ejecutar la consulta, la prueba fallaría con un panic— y sin
// acciones no hay nada que hacer. El commit y el rollback reales se verifican
// contra PostgreSQL en repository_audit_integration_test.go.
func TestInsertActionsValidatesBeforeTouchingTheDatabase(t *testing.T) {
	err := insertActions(context.Background(), nil, []audit.Action{{
		Code:        "event.archive",
		TargetKind:  audit.TargetContent,
		TargetLabel: "Eventos · X",
		Result:      audit.ResultSuccess,
	}})
	if !errors.Is(err, audit.ErrUnknownActionCode) {
		t.Fatalf("insertActions(código desconocido) = %v, se esperaba %v", err, audit.ErrUnknownActionCode)
	}

	if err := insertActions(context.Background(), nil, nil); err != nil {
		t.Errorf("insertActions(sin acciones) = %v, se esperaba nil", err)
	}
}

// TestWithTxFailsWhenTransactionCannotStart comprueba el contrato de la base de
// escritura (T411): si la transacción no se puede abrir, `withTx` devuelve el
// error y NO ejecuta la mutación. Se usa un pool perezoso apuntando a un puerto
// cerrado (ninguna conexión posible) para no depender de una base real.
func TestWithTxFailsWhenTransactionCannotStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, "postgres://app:app@127.0.0.1:1/app?sslmode=disable")
	if err != nil {
		t.Fatalf("crear pool: %v", err)
	}
	defer pool.Close()

	called := false
	err = NewRepository(pool).withTx(ctx, func(*repository) error {
		called = true
		return nil
	})
	if err == nil {
		t.Fatal("withTx debería fallar si no puede abrir la transacción")
	}
	if called {
		t.Error("la mutación no debe ejecutarse sin transacción")
	}
}

// TestClassifyTranslatesPostgresErrors fija la traducción de errores de
// PostgreSQL a errores de dominio: sin fila → 404; duplicado (23505) y FK en uso
// (23503, la categoría con ON DELETE RESTRICT) → 409; CHECK (23514) → 400; y
// cualquier otro error se devuelve tal cual (nunca se disfraza).
func TestClassifyTranslatesPostgresErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want apperr.Kind
	}{
		{name: "sin fila", err: pgx.ErrNoRows, want: apperr.KindNotFound},
		{name: "duplicado", err: &pgconn.PgError{Code: codeUniqueViolation}, want: apperr.KindConflict},
		{name: "clave foránea en uso", err: &pgconn.PgError{Code: codeForeignKeyViolation}, want: apperr.KindConflict},
		{name: "check", err: &pgconn.PgError{Code: codeCheckViolation}, want: apperr.KindInvalid},
		{name: "otro error de PostgreSQL", err: &pgconn.PgError{Code: "57014"}, want: apperr.KindInternal},
		{name: "error cualquiera", err: errors.New("boom"), want: apperr.KindInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.err, "no existe", "duplicado")
			if kind := errorKind(got); kind != tt.want {
				t.Fatalf("classify = %v (%v), se esperaba %v", got, kind, tt.want)
			}
			if tt.want == apperr.KindInternal && !errors.Is(got, tt.err) {
				t.Fatalf("classify perdió el error original: %v", got)
			}
		})
	}

	if err := classify(nil, "", ""); err != nil {
		t.Errorf("classify(nil) = %v, se esperaba nil", err)
	}
}

// TestWrapKeepsOperationContext comprueba que wrap envuelve el error traducido
// con el nombre de la operación sin romper errors.Is/As (FR-013: el log tiene el
// detalle, el cliente no).
func TestWrapKeepsOperationContext(t *testing.T) {
	err := wrap(&pgconn.PgError{Code: codeUniqueViolation}, "insert event", "", "duplicado")
	if err == nil {
		t.Fatal("wrap devolvió nil")
	}
	if got := err.Error(); !strings.HasPrefix(got, "insert event: ") {
		t.Errorf("wrap = %q, se esperaba el contexto de la operación", got)
	}
	if kind := errorKind(err); kind != apperr.KindConflict {
		t.Errorf("kind = %v, se esperaba conflict", kind)
	}
	if err := wrap(nil, "insert event", "", ""); err != nil {
		t.Errorf("wrap(nil) = %v, se esperaba nil", err)
	}
}
