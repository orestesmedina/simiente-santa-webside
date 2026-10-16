package donaciones

import (
	"context"
	"encoding/json"
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

// currencyOf es un atajo para construir un *Currency.
func currencyOf(c Currency) *Currency { return &c }

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
	if got := textPtr(text("Banco Nacional")); got == nil || *got != "Banco Nacional" {
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
	if got := pgText("Banco Nacional"); !got.Valid || got.String != "Banco Nacional" {
		t.Errorf("pgText(válido) = %+v", got)
	}
	if got := pgLabel(""); got.Valid {
		t.Errorf("pgLabel(\"\") = %+v, se esperaba NULL", got)
	}
	if got := pgLabel("Donaciones · Banco Nacional · CR05…"); !got.Valid {
		t.Errorf("pgLabel(válido) = %+v", got)
	}
}

// TestCurrencyHelpers verifica la traducción de la moneda (solo existe en IBAN):
// NULL ↔ nil y el código del contrato en ambos sentidos.
func TestCurrencyHelpers(t *testing.T) {
	if got := currencyPtr(pgtype.Text{}); got != nil {
		t.Errorf("currencyPtr(NULL) = %v, se esperaba nil", got)
	}
	if got := currencyPtr(text("CRC")); got == nil || *got != CurrencyCRC {
		t.Errorf("currencyPtr(CRC) = %v", got)
	}
	if got := currencyPtr(text("USD")); got == nil || *got != CurrencyUSD {
		t.Errorf("currencyPtr(USD) = %v", got)
	}
	if got := pgCurrencyPtr(nil); got.Valid {
		t.Errorf("pgCurrencyPtr(nil) = %+v, se esperaba NULL", got)
	}
	if got := pgCurrencyPtr(currencyOf(CurrencyUSD)); !got.Valid || got.String != "USD" {
		t.Errorf("pgCurrencyPtr(USD) = %+v", got)
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

// TestRegistriesValid fija los registros cerrados del contrato: estados, tipos
// de método, monedas y claves de texto.
func TestRegistriesValid(t *testing.T) {
	states := []struct {
		state PublicationState
		want  bool
	}{
		{StateDraft, true}, {StatePublished, true},
		{PublicationState("archived"), false}, {PublicationState(""), false},
	}
	for _, tt := range states {
		if got := tt.state.Valid(); got != tt.want {
			t.Errorf("PublicationState(%q).Valid() = %v, se esperaba %v", tt.state, got, tt.want)
		}
	}

	kinds := []struct {
		kind MetodoKind
		want bool
	}{
		{KindIban, true}, {KindSinpe, true},
		{MetodoKind("cash"), false}, {MetodoKind(""), false},
	}
	for _, tt := range kinds {
		if got := tt.kind.Valid(); got != tt.want {
			t.Errorf("MetodoKind(%q).Valid() = %v, se esperaba %v", tt.kind, got, tt.want)
		}
	}

	currencies := []struct {
		currency Currency
		want     bool
	}{
		{CurrencyCRC, true}, {CurrencyUSD, true},
		{Currency("EUR"), false}, {Currency(""), false},
	}
	for _, tt := range currencies {
		if got := tt.currency.Valid(); got != tt.want {
			t.Errorf("Currency(%q).Valid() = %v, se esperaba %v", tt.currency, got, tt.want)
		}
	}

	keys := []struct {
		key  TextKey
		want bool
	}{
		{TextKeyIntro, true}, {TextKeyUsage, true},
		{TextKey("other"), false}, {TextKey(""), false},
	}
	for _, tt := range keys {
		if got := tt.key.Valid(); got != tt.want {
			t.Errorf("TextKey(%q).Valid() = %v, se esperaba %v", tt.key, got, tt.want)
		}
	}
}

// TestMapMetodoDonacionMapsNullableColumns traduce una fila completa del panel a
// la entidad del dominio: las columnas anulables pasan a `*string`/`*Currency` y
// las fechas a time.Time (§8.1.6).
func TestMapMetodoDonacionMapsNullableColumns(t *testing.T) {
	id := uuid.New()
	created := time.Date(2026, 10, 14, 9, 30, 0, 0, time.UTC)

	got := mapMetodoDonacion(gendb.DonationMethod{
		ID:               pgUUID(id),
		Kind:             "iban",
		Number:           "CR05015202001234567890",
		Institution:      "Banco Nacional",
		HolderEs:         text("Iglesia Simiente Santa"),
		HolderEn:         pgtype.Text{},
		Currency:         text("CRC"),
		PublicationState: "published",
		SortOrder:        3,
		CreatedAt:        pgtype.Timestamptz{Time: created, Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: created, Valid: true},
	})

	if got.ID != id {
		t.Errorf("ID = %v, se esperaba %v", got.ID, id)
	}
	if got.Kind != KindIban || got.Number != "CR05015202001234567890" || got.Institution != "Banco Nacional" {
		t.Errorf("campos no traducidos: %+v", got)
	}
	if got.HolderEs == nil || *got.HolderEs != "Iglesia Simiente Santa" {
		t.Errorf("HolderEs = %v", got.HolderEs)
	}
	if got.HolderEn != nil {
		t.Errorf("HolderEn = %v, se esperaba nil (NULL)", got.HolderEn)
	}
	if got.Currency == nil || *got.Currency != CurrencyCRC {
		t.Errorf("Currency = %v", got.Currency)
	}
	if got.PublicationState != StatePublished || got.SortOrder != 3 {
		t.Errorf("estado/orden = %q/%d", got.PublicationState, got.SortOrder)
	}
	if !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(created) {
		t.Errorf("fechas = %v / %v, se esperaba %v", got.CreatedAt, got.UpdatedAt, created)
	}
}

// TestMapPublishedMetodoDonacionFixesPublishedState traduce la fila pública: la
// moneda nula se queda en nil (SINPE) y el estado se fija en `published` porque
// la consulta solo devuelve publicados.
func TestMapPublishedMetodoDonacionFixesPublishedState(t *testing.T) {
	got := mapPublishedMetodoDonacion(gendb.ListPublishedDonationMethodsRow{
		ID:          pgUUID(uuid.New()),
		Kind:        "sinpe",
		Number:      "+50688888888",
		Institution: "Banco Nacional",
		Currency:    pgtype.Text{},
		SortOrder:   1,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if got.PublicationState != StatePublished {
		t.Errorf("PublicationState = %q, se esperaba published", got.PublicationState)
	}
	if got.Currency != nil {
		t.Errorf("Currency = %v, se esperaba nil (SINPE sin moneda)", got.Currency)
	}
	if got.Kind != KindSinpe || got.Number != "+50688888888" {
		t.Errorf("campos no traducidos: %+v", got)
	}
}

// TestMapTextoDonacionMapsNullableColumns traduce la fila de un texto.
func TestMapTextoDonacionMapsNullableColumns(t *testing.T) {
	got := mapTextoDonacion(gendb.DonationText{
		ID:               pgUUID(uuid.New()),
		TextKey:          "intro",
		BodyEs:           "Puedes dar tu ofrenda…",
		BodyEn:           pgtype.Text{},
		PublicationState: "draft",
	})
	if got.TextKey != TextKeyIntro || got.BodyEs != "Puedes dar tu ofrenda…" {
		t.Errorf("campos no traducidos: %+v", got)
	}
	if got.BodyEn != nil {
		t.Errorf("BodyEn = %v, se esperaba nil (NULL)", got.BodyEn)
	}
	if got.PublicationState != StateDraft {
		t.Errorf("PublicationState = %q, se esperaba draft", got.PublicationState)
	}
}

// TestAuditParamsMapsDonationActions verifica el mapeo audit.Action →
// InsertAdminActionParams (R7-10): un objetivo `content` no lleva FK y conserva
// la etiqueta; los 8 códigos `donation.*` son válidos.
func TestAuditParamsMapsDonationActions(t *testing.T) {
	actor := uuid.New()
	tests := []struct {
		name string
		code string
	}{
		{name: "alta", code: audit.ActionDonationCreate},
		{name: "edición", code: audit.ActionDonationUpdate},
		{name: "borrado", code: audit.ActionDonationDelete},
		{name: "publicación", code: audit.ActionDonationPublish},
		{name: "retirada", code: audit.ActionDonationUnpublish},
		{name: "guardado de texto", code: audit.ActionDonationTextUpdate},
		{name: "publicación de texto", code: audit.ActionDonationTextPublish},
		{name: "retirada de texto", code: audit.ActionDonationTextUnpublish},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := audit.Action{
				ActorUserID: &actor,
				Code:        tt.code,
				TargetKind:  audit.TargetContent,
				TargetLabel: "Donaciones · Banco Nacional · CR05…",
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
			if !params.TargetLabel.Valid || params.TargetLabel.String != "Donaciones · Banco Nacional · CR05…" {
				t.Errorf("TargetLabel = %+v", params.TargetLabel)
			}
			if params.Result != "success" {
				t.Errorf("Result = %q", params.Result)
			}
		})
	}

	empty := auditParams(audit.Action{TargetKind: audit.TargetContent})
	if empty.TargetLabel.Valid {
		t.Errorf("TargetLabel vacío = %+v, se esperaba NULL", empty.TargetLabel)
	}
}

// TestInsertActionsValidatesBeforeTouchingTheDatabase fija el fail-closed del
// registro de auditoría (FR-017): un código fuera del registro cerrado se
// rechaza ANTES de ejecutar SQL —el `tx` es nil a propósito— y sin acciones no
// hay nada que hacer.
func TestInsertActionsValidatesBeforeTouchingTheDatabase(t *testing.T) {
	err := insertActions(context.Background(), nil, []audit.Action{{
		Code:        "donation.archive", // fuera del registro cerrado (R7-10)
		TargetKind:  audit.TargetContent,
		TargetLabel: "Donaciones · X",
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
// escritura (T708): si la transacción no se puede abrir, `withTx` devuelve el
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
// CHECK (23514) → 400 con el detalle por campo; y cualquier otro error se
// devuelve tal cual (nunca se disfraza).
func TestClassifyTranslatesPostgresErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want apperr.Kind
	}{
		{name: "sin fila", err: pgx.ErrNoRows, want: apperr.KindNotFound},
		{name: "duplicado", err: &pgconn.PgError{Code: codeUniqueViolation}, want: apperr.KindConflict},
		{
			name: "check de moneda por tipo",
			err:  &pgconn.PgError{Code: codeCheckViolation, ConstraintName: "donation_methods_check1"},
			want: apperr.KindInvalid,
		},
		{name: "otro error de PostgreSQL", err: &pgconn.PgError{Code: "57014"}, want: apperr.KindInternal},
		{name: "error cualquiera", err: errors.New("boom"), want: apperr.KindInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.err, "El método no existe", "duplicado")
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
// de CHECK de las tablas de F7: la BD es la última línea de defensa y su 400
// debe decir qué campo lo rompió. Los nombres son los que PostgreSQL asigna a
// los CHECK sin nombre de la migración `000012` (verificados en T711).
func TestCheckDetailsMapsInvariants(t *testing.T) {
	tests := []struct {
		constraint string
		field      string
	}{
		{constraint: "donation_methods_check", field: "number"},
		{constraint: "donation_methods_check1", field: "currency"},
		{constraint: "donation_methods_check2", field: "holderEs"},
		{constraint: "donation_texts_check", field: "textEs"},
		{constraint: "donation_methods_kind_check", field: "kind"},
		{constraint: "donation_methods_currency_check", field: "currency"},
		{constraint: "donation_methods_publication_state_check", field: "publicationState"},
		{constraint: "donation_texts_publication_state_check", field: "publicationState"},
		{constraint: "donation_texts_text_key_check", field: "textKey"},
		{constraint: "donation_methods_number_check", field: "number"},
		{constraint: "donation_methods_institution_check", field: "institution"},
		{constraint: "donation_methods_holder_es_check", field: "holderEs"},
		{constraint: "donation_methods_holder_en_check", field: "holderEn"},
		{constraint: "donation_texts_body_es_check", field: "textEs"},
		{constraint: "donation_texts_body_en_check", field: "textEn"},
		{constraint: "donation_methods_institution_check1", field: "institution"},
		{constraint: "donation_methods_holder_es_check1", field: "holderEs"},
		{constraint: "donation_methods_holder_en_check1", field: "holderEn"},
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

	if got := checkDetails("ministries_check"); got != nil {
		t.Errorf("checkDetails(desconocido) = %v, se esperaba nil", got)
	}
}

// TestOptionalPatchSemantics fija la semántica del PATCH (R7-15): `null`
// explícito borra, la clave presente con valor se aplica y la ausente no toca.
func TestOptionalPatchSemantics(t *testing.T) {
	var patch DonacionPatch
	if err := json.Unmarshal([]byte(`{"holderEs":null,"institution":"Banco"}`), &patch); err != nil {
		t.Fatalf("Unmarshal DonacionPatch: %v", err)
	}
	if !patch.HolderEs.Set() || !patch.HolderEs.Null() {
		t.Errorf("holderEs = %+v, se esperaba presente y null", patch.HolderEs)
	}
	if patch.HolderEn.Set() {
		t.Errorf("holderEn = %+v, se esperaba ausente", patch.HolderEn)
	}
	if !patch.Institution.Set() || patch.Institution.Null() || patch.Institution.Value() != "Banco" {
		t.Errorf("institution = %+v, se esperaba presente con valor", patch.Institution)
	}

	patch = DonacionPatch{SortOrder: OptionalOf(7), Number: OptionalNull[string]()}
	if !patch.SortOrder.Set() || patch.SortOrder.Value() != 7 {
		t.Errorf("sortOrder = %+v", patch.SortOrder)
	}
	if !patch.Number.Set() || !patch.Number.Null() {
		t.Errorf("number = %+v, se esperaba null explícito", patch.Number)
	}
}

// TestWrapKeepsOperationContext comprueba que wrap envuelve el error traducido
// con el nombre de la operación sin romper errors.Is/As.
func TestWrapKeepsOperationContext(t *testing.T) {
	err := wrap(&pgconn.PgError{Code: codeUniqueViolation}, "insert donation method", "", "duplicado")
	if err == nil {
		t.Fatal("wrap devolvió nil")
	}
	if got := err.Error(); !strings.HasPrefix(got, "insert donation method: ") {
		t.Errorf("wrap = %q, se esperaba el contexto de la operación", got)
	}
	if kind := errorKind(err); kind != apperr.KindConflict {
		t.Errorf("kind = %v, se esperaba conflict", kind)
	}
	if err := wrap(nil, "insert donation method", "", ""); err != nil {
		t.Errorf("wrap(nil) = %v, se esperaba nil", err)
	}

	wrapped := wrap(
		&pgconn.PgError{Code: codeCheckViolation, ConstraintName: "donation_methods_check1"},
		"insert donation method", "", "",
	)
	if details := errorDetails(wrapped); details == nil || details["currency"] == nil {
		t.Errorf("el detalle por campo se perdió: %v", details)
	}
}
