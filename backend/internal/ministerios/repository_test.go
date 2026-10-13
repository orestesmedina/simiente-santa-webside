package ministerios

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

	gendb "simiente-santa/backend/internal/db"
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

// errorDetails devuelve el detalle seguro del error de dominio (nil si no es un
// error de dominio o no lleva detalle).
func errorDetails(err error) map[string]any {
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		return domainErr.Details
	}
	return nil
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
	if got := pgText("Alabanza"); !got.Valid || got.String != "Alabanza" {
		t.Errorf("pgText(válido) = %+v", got)
	}
	if got := pgLabel(""); got.Valid {
		t.Errorf("pgLabel(\"\") = %+v, se esperaba NULL", got)
	}
	if got := pgLabel("Ministerios · Alabanza"); !got.Valid {
		t.Errorf("pgLabel(válido) = %+v", got)
	}
}

// TestMaskedTextHelper fija el mapeo de las cuatro columnas del encargado que
// sqlc tipa como `interface{}` (no sabe inferir el tipo de un `CASE WHEN
// leader_visible`): `nil` (la fila oculta) y cualquier valor que no sea texto se
// traducen a "sin dato" y un texto se conserva. El tipo `interface{}` no sale
// del paquete de datos (§8.1.6, R6-2).
func TestMaskedTextHelper(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  *string
	}{
		{name: "nil (fila con la ficha oculta)", value: nil, want: nil},
		{name: "texto", value: "Ana Pérez", want: textOf("Ana Pérez")},
		{name: "texto vacío", value: "", want: nil},
		{name: "otro tipo", value: []byte("Ana"), want: nil},
		{name: "número", value: 42, want: nil},
		{name: "booleano", value: true, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskedText(tt.value)
			switch {
			case tt.want == nil && got != nil:
				t.Fatalf("maskedText(%v) = %q, se esperaba sin dato", tt.value, *got)
			case tt.want != nil && got == nil:
				t.Fatalf("maskedText(%v) = sin dato, se esperaba %q", tt.value, *tt.want)
			case tt.want != nil && *got != *tt.want:
				t.Fatalf("maskedText(%v) = %q, se esperaba %q", tt.value, *got, *tt.want)
			}
		})
	}
}

// TestStateFilter verifica el filtro opcional de estado del panel: nil es "sin
// filtro" (NULL en la consulta) y un estado viaja como texto.
func TestStateFilter(t *testing.T) {
	if got := pgStatePtr(nil); got.Valid {
		t.Errorf("pgStatePtr(nil) = %+v, se esperaba NULL", got)
	}
	published := StatePublished
	if got := pgStatePtr(&published); !got.Valid || got.String != "published" {
		t.Errorf("pgStatePtr(published) = %+v", got)
	}
	draft := StateDraft
	if got := pgStatePtr(&draft); !got.Valid || got.String != "draft" {
		t.Errorf("pgStatePtr(draft) = %+v", got)
	}
}

// TestPublicationStateValid fija el registro cerrado de estados (FR-012).
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

// TestMapMinistryMapsNullableColumns traduce una fila completa del panel a la
// entidad del dominio: las columnas anulables (incluida la ficha del encargado
// íntegra, que el panel sí ve con el interruptor apagado) pasan a `*string` y
// las fechas a time.Time (§8.1.6).
func TestMapMinistryMapsNullableColumns(t *testing.T) {
	id := uuid.New()
	created := time.Date(2026, 10, 13, 9, 30, 0, 0, time.UTC)

	got := mapMinistry(gendb.Ministry{
		ID:               pgUUID(id),
		NameEs:           "Alabanza",
		NameEn:           text("Worship"),
		DescriptionEs:    "Grupo de alabanza",
		DescriptionEn:    pgtype.Text{},
		LeaderNameEs:     text("Ana Pérez"),
		LeaderNameEn:     pgtype.Text{},
		LeaderPhone:      text("+58 412 000 0000"),
		LeaderEmail:      pgtype.Text{},
		LeaderVisible:    false,
		WhatsappUrl:      text("https://chat.whatsapp.com/abc"),
		ScheduleEs:       text("Todos los viernes, 7:00 p. m."),
		ScheduleEn:       pgtype.Text{},
		PlaceEs:          text("Templo"),
		PlaceEn:          text("Temple"),
		ImageFile:        text("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png"),
		ImageAltEs:       text("Grupo de alabanza"),
		ImageAltEn:       pgtype.Text{},
		PublicationState: "published",
		SortOrder:        3,
		CreatedAt:        pgtype.Timestamptz{Time: created, Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: created, Valid: true},
	})

	if got.ID != id {
		t.Errorf("ID = %v, se esperaba %v", got.ID, id)
	}
	if got.NameEs != "Alabanza" || got.NameEn == nil || *got.NameEn != "Worship" {
		t.Errorf("nombre = %q / %v", got.NameEs, got.NameEn)
	}
	if got.DescriptionEn != nil {
		t.Errorf("DescriptionEn = %v, se esperaba nil (NULL)", got.DescriptionEn)
	}
	if got.LeaderNameEs == nil || *got.LeaderNameEs != "Ana Pérez" {
		t.Errorf("LeaderNameEs = %v, se esperaba «Ana Pérez»", got.LeaderNameEs)
	}
	if got.LeaderPhone == nil || *got.LeaderPhone != "+58 412 000 0000" {
		t.Errorf("LeaderPhone = %v", got.LeaderPhone)
	}
	if got.LeaderEmail != nil {
		t.Errorf("LeaderEmail = %v, se esperaba nil", got.LeaderEmail)
	}
	if got.LeaderVisible {
		t.Error("LeaderVisible = true, se esperaba false (dato guardado ≠ publicado)")
	}
	if got.PlaceEn == nil || *got.PlaceEn != "Temple" {
		t.Errorf("PlaceEn = %v", got.PlaceEn)
	}
	if got.PublicationState != StatePublished {
		t.Errorf("PublicationState = %q", got.PublicationState)
	}
	if got.SortOrder != 3 {
		t.Errorf("SortOrder = %d, se esperaba 3", got.SortOrder)
	}
	if !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(created) {
		t.Errorf("fechas = %v / %v, se esperaba %v", got.CreatedAt, got.UpdatedAt, created)
	}
}

// TestAuditParamsMapsMinistryActions verifica el mapeo audit.Action →
// InsertAdminActionParams (R6-12): un objetivo `content` no lleva FK y conserva
// la etiqueta; los códigos de la visibilidad del encargado son válidos.
func TestAuditParamsMapsMinistryActions(t *testing.T) {
	actor := uuid.New()
	tests := []struct {
		name string
		code string
	}{
		{name: "alta", code: audit.ActionMinistryCreate},
		{name: "edición", code: audit.ActionMinistryUpdate},
		{name: "borrado", code: audit.ActionMinistryDelete},
		{name: "publicación", code: audit.ActionMinistryPublish},
		{name: "retirada", code: audit.ActionMinistryUnpublish},
		{name: "ficha del encargado publicada", code: audit.ActionMinistryLeaderShow},
		{name: "ficha del encargado oculta", code: audit.ActionMinistryLeaderHide},
		{name: "subida de imagen", code: audit.ActionMinistryImageUpload},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := audit.Action{
				ActorUserID: &actor,
				Code:        tt.code,
				TargetKind:  audit.TargetContent,
				TargetLabel: "Ministerios · Alabanza",
				Result:      audit.ResultSuccess,
			}
			if err := action.Validate(); err != nil {
				t.Fatalf("la acción %q no valida: %v", tt.code, err)
			}

			params := auditParams(action)
			if !params.ActorUserID.Valid || uuid.UUID(params.ActorUserID.Bytes) != actor {
				t.Errorf("ActorUserID = %+v", params.ActorUserID)
			}
			if params.Action != tt.code {
				t.Errorf("Action = %q, se esperaba %q", params.Action, tt.code)
			}
			if params.TargetKind != "content" {
				t.Errorf("TargetKind = %q, se esperaba content", params.TargetKind)
			}
			if params.TargetUserID.Valid || params.TargetRoleID.Valid {
				t.Errorf("las FK de contenido deben ir NULL: %+v / %+v", params.TargetUserID, params.TargetRoleID)
			}
			if !params.TargetLabel.Valid || params.TargetLabel.String != "Ministerios · Alabanza" {
				t.Errorf("TargetLabel = %+v", params.TargetLabel)
			}
			if params.Result != "success" {
				t.Errorf("Result = %q", params.Result)
			}
		})
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
		Code:        "ministry.archive",
		TargetKind:  audit.TargetContent,
		TargetLabel: "Ministerios · X",
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
// escritura (T607): si la transacción no se puede abrir, `withTx` devuelve el
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
// PostgreSQL a errores de dominio: sin fila → 404; duplicado (23505) → 409;
// CHECK (23514) → 400 con el detalle por campo del invariante que se rompió; y
// cualquier otro error se devuelve tal cual (nunca se disfraza).
func TestClassifyTranslatesPostgresErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want apperr.Kind
	}{
		{name: "sin fila", err: pgx.ErrNoRows, want: apperr.KindNotFound},
		{name: "duplicado", err: &pgconn.PgError{Code: codeUniqueViolation}, want: apperr.KindConflict},
		{
			name: "check de contacto excluyente",
			err:  &pgconn.PgError{Code: codeCheckViolation, ConstraintName: "ministries_check"},
			want: apperr.KindInvalid,
		},
		{name: "otro error de PostgreSQL", err: &pgconn.PgError{Code: "57014"}, want: apperr.KindInternal},
		{name: "error cualquiera", err: errors.New("boom"), want: apperr.KindInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.err, "El ministerio no existe", "duplicado")
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

// TestCheckDetailsMapsInvariants fija el detalle por campo de las violaciones
// de CHECK de la tabla `ministries`: la BD es la última línea de defensa de los
// invariantes de FR-014/FR-015/FR-020 y su 400 debe decir qué campo lo rompió.
// Los CHECK sin nombre van en el orden de la migración `000010` (nombres
// propios `ministries_check1`, `ministries_check2`…).
func TestCheckDetailsMapsInvariants(t *testing.T) {
	tests := []struct {
		constraint string
		field      string
	}{
		{constraint: "ministries_check", field: "leaderEmail"},
		{constraint: "ministries_check1", field: "leaderNameEs"},
		{constraint: "ministries_check2", field: "nameEs"},
		{constraint: "ministries_check3", field: "descriptionEs"},
		{constraint: "ministries_check4", field: "leaderNameEs"},
		{constraint: "ministries_check5", field: "scheduleEs"},
		{constraint: "ministries_check6", field: "placeEs"},
		{constraint: "ministries_check7", field: "imageAltEs"},
		{constraint: "ministries_check8", field: "imageAltEs"},
		{constraint: "ministries_whatsapp_url_check1", field: "whatsappUrl"},
		{constraint: "ministries_name_es_check1", field: "nameEs"},
		{constraint: "ministries_leader_phone_check1", field: "leaderPhone"},
		{constraint: "ministries_leader_email_check1", field: "leaderEmail"},
		{constraint: "ministries_schedule_es_check1", field: "scheduleEs"},
		{constraint: "ministries_place_es_check1", field: "placeEs"},
		{constraint: "ministries_leader_name_es_check1", field: "leaderNameEs"},
		{constraint: "ministries_whatsapp_url_check2", field: "whatsappUrl"},
	}

	for _, tt := range tests {
		details := checkDetails(tt.constraint)
		if details == nil {
			t.Errorf("checkDetails(%q) = nil, se esperaba el detalle del campo", tt.constraint)
			continue
		}
		if _, ok := details[tt.field]; !ok {
			t.Errorf("checkDetails(%q) = %v, se esperaba la clave %q", tt.constraint, details, tt.field)
		}
	}

	// Un CHECK del que no sabemos nada (otra tabla, otra migración) no inventa
	// detalle: el 400 sale con el mensaje seguro y sin `details`.
	if got := checkDetails("agenda_events_check"); got != nil {
		t.Errorf("checkDetails(desconocido) = %v, se esperaba nil", got)
	}
}

// TestWrapKeepsOperationContext comprueba que wrap envuelve el error traducido
// con el nombre de la operación sin romper errors.Is/As (FR-013: el log tiene el
// detalle, el cliente no).
func TestWrapKeepsOperationContext(t *testing.T) {
	err := wrap(&pgconn.PgError{Code: codeUniqueViolation}, "insert ministry", "", "duplicado")
	if err == nil {
		t.Fatal("wrap devolvió nil")
	}
	if got := err.Error(); !strings.HasPrefix(got, "insert ministry: ") {
		t.Errorf("wrap = %q, se esperaba el contexto de la operación", got)
	}
	if kind := errorKind(err); kind != apperr.KindConflict {
		t.Errorf("kind = %v, se esperaba conflict", kind)
	}
	if err := wrap(nil, "insert ministry", "", ""); err != nil {
		t.Errorf("wrap(nil) = %v, se esperaba nil", err)
	}

	// El detalle por campo del CHECK sobrevive al envoltorio (lo lee el handler).
	wrapped := wrap(
		&pgconn.PgError{Code: codeCheckViolation, ConstraintName: "ministries_check8"},
		"insert ministry", "", "",
	)
	if details := errorDetails(wrapped); details == nil || details["imageAltEs"] == nil {
		t.Errorf("el detalle por campo se perdió: %v", details)
	}
}
