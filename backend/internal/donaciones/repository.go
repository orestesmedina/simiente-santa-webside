package donaciones

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/database"
)

// repository.go es la base del acceso a datos del dominio donaciones: el
// constructor sobre el *pgxpool.Pool compartido, la transacción de escritura
// (`withTx`) y el mapeo de tipos entre pgtype.* y el dominio.
//
// No conoce HTTP ni las reglas de negocio (arq. R4): solo SQL (las consultas
// generadas por sqlc de T703), transacciones, el mapeo de tipos y la inserción
// de la auditoría dentro del tx. Los tipos pgtype.* de pgx se traducen AQUÍ y
// solo aquí (§8.1.6); el dominio usa uuid.UUID, time.Time y string.
//
// Las mutaciones reciben las filas de auditoría que deben insertarse en la MISMA
// transacción (variádicas): o el cambio y su registro, o nada (FR-017/R7-10).
// Ningún dominio importa a otro: la consulta `InsertAdminAction` se reutiliza de
// internal/db (R2).
type repository struct {
	q    *gendb.Queries
	pool *pgxpool.Pool
}

// NewRepository construye el repositorio del dominio donaciones a partir del
// pool compartido que compone cmd/api. Devuelve el tipo concreto: la interfaz
// que consume el service la declara el propio service (arq. R3).
func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{q: gendb.New(pool), pool: pool}
}

// withTx ejecuta fn con un repository ligado a una transacción. Las mutaciones
// agrupan el cambio de contenido y su registro de auditoría en una única
// transacción (R7-10): o ambos o ninguno (FR-017/edge case).
func (r *repository) withTx(ctx context.Context, fn func(tx *repository) error) error {
	return database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&repository{q: gendb.New(tx), pool: r.pool})
	})
}

// insertAdminAction registra una acción administrativa en la transacción
// vigente (r ya está ligado al tx cuando se llama desde una mutación). Reutiliza
// la consulta compartida `InsertAdminAction` de internal/db (R7-10).
func (r *repository) insertAdminAction(ctx context.Context, action audit.Action) error {
	if err := action.Validate(); err != nil {
		return fmt.Errorf("insert admin action: %w", err)
	}
	if _, err := r.q.InsertAdminAction(ctx, auditParams(action)); err != nil {
		return wrap(err, "insert admin action", "", "")
	}
	return nil
}

// insertActions ejecuta las filas de auditoría de una mutación dentro de su
// misma transacción (FR-017, fail-closed): si una falla, la mutación no se
// aplica.
func insertActions(ctx context.Context, tx *repository, actions []audit.Action) error {
	for _, action := range actions {
		if err := tx.insertAdminAction(ctx, action); err != nil {
			return err
		}
	}
	return nil
}

// auditParams traduce una audit.Action al parámetro generado por sqlc. Es el
// mapeo que R7-10 permite repetir por dominio (no se comparte código de dominio,
// solo la consulta SQL).
func auditParams(action audit.Action) gendb.InsertAdminActionParams {
	return gendb.InsertAdminActionParams{
		ActorUserID:  pgUUIDPtr(action.ActorUserID),
		Action:       action.Code,
		TargetKind:   string(action.TargetKind),
		TargetUserID: pgUUIDPtr(action.TargetUserID),
		TargetRoleID: pgUUIDPtr(action.TargetRoleID),
		TargetLabel:  pgLabel(action.TargetLabel),
		Result:       string(action.Result),
	}
}

// --- Mapeo pgtype.* → dominio (§8.1.6) ---

// uuidValue traduce una columna UUID NOT NULL.
func uuidValue(v pgtype.UUID) uuid.UUID {
	return uuid.UUID(v.Bytes)
}

// uuidPtr traduce una columna UUID anulable a *uuid.UUID (nil si es NULL).
func uuidPtr(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	u := uuid.UUID(v.Bytes)
	return &u
}

// pgUUID traduce un uuid.UUID del dominio al parámetro pgtype.UUID.
func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// pgUUIDPtr traduce un *uuid.UUID al parámetro pgtype.UUID (NULL si es nil).
func pgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*id)
}

// textPtr traduce una columna TEXT anulable a *string (nil si NULL).
func textPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

// pgTextPtr traduce un *string del dominio al parámetro pgtype.Text.
func pgTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// pgText traduce un string ya normalizado: vacío se guarda como NULL (los campos
// `*_en` opcionales, analyze I6).
func pgText(s string) pgtype.Text {
	if strings.TrimSpace(s) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// pgLabel traduce la etiqueta de un objetivo de acción: un valor vacío se guarda
// como NULL (target_label es anulable salvo para `content`, que el registro de
// audit exige y valida).
func pgLabel(label string) pgtype.Text {
	if label == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: label, Valid: true}
}

// pgStatePtr traduce el filtro opcional de estado del listado del panel (nil =
// sin filtro, que es un NULL en la consulta).
func pgStatePtr(state *PublicationState) pgtype.Text {
	if state == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*state), Valid: true}
}

// currencyPtr traduce una columna TEXT anulable a *Currency (nil si NULL): la
// moneda solo existe en las cuentas IBAN.
func currencyPtr(v pgtype.Text) *Currency {
	if !v.Valid {
		return nil
	}
	c := Currency(v.String)
	return &c
}

// pgCurrencyPtr traduce un *Currency del dominio al parámetro pgtype.Text (NULL
// si es nil: el SINPE no lleva moneda, Q2/FR-014).
func pgCurrencyPtr(c *Currency) pgtype.Text {
	if c == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*c), Valid: true}
}

// mapMetodoDonacion traduce la fila completa de un método (sqlc) a la entidad
// del dominio. La usan las consultas del panel, que devuelven todos los campos
// incluido el estado (dato guardado ≠ dato publicado).
func mapMetodoDonacion(row gendb.DonationMethod) MetodoDonacion {
	return MetodoDonacion{
		ID:               uuidValue(row.ID),
		Kind:             MetodoKind(row.Kind),
		Number:           row.Number,
		Institution:      row.Institution,
		HolderEs:         textPtr(row.HolderEs),
		HolderEn:         textPtr(row.HolderEn),
		Currency:         currencyPtr(row.Currency),
		PublicationState: PublicationState(row.PublicationState),
		SortOrder:        int(row.SortOrder),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapPublishedMetodoDonacion traduce la fila pública (sqlc) a la entidad del
// dominio. `publication_state` no viaja en la consulta: el estado se fija en
// `published` porque la consulta solo devuelve filas publicadas.
func mapPublishedMetodoDonacion(row gendb.ListPublishedDonationMethodsRow) MetodoDonacion {
	return MetodoDonacion{
		ID:               uuidValue(row.ID),
		Kind:             MetodoKind(row.Kind),
		Number:           row.Number,
		Institution:      row.Institution,
		HolderEs:         textPtr(row.HolderEs),
		HolderEn:         textPtr(row.HolderEn),
		Currency:         currencyPtr(row.Currency),
		PublicationState: StatePublished,
		SortOrder:        int(row.SortOrder),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapTextoDonacion traduce la fila de un texto (sqlc) a la entidad del dominio.
// La usan las consultas del panel y del público (misma forma de fila).
func mapTextoDonacion(row gendb.DonationText) TextoDonacion {
	return TextoDonacion{
		ID:               uuidValue(row.ID),
		TextKey:          TextKey(row.TextKey),
		BodyEs:           row.BodyEs,
		BodyEn:           textPtr(row.BodyEn),
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// --- Traducción de errores de PostgreSQL ---

// Códigos SQLSTATE que el dominio conoce (skill postgres-db): 23505
// unique_violation y 23514 check_violation. La tabla `donation_methods` no
// tiene `UNIQUE` (el duplicado lo detecta `ExistsDonationMethodDuplicate`,
// R7-9); `donation_texts` sí tiene `UNIQUE (text_key)` (el upsert lo absorbe) y
// el duplicado se traduce igual porque la base es la última línea de defensa.
const (
	codeUniqueViolation = "23505"
	codeCheckViolation  = "23514"
)

// classify traduce un error de PostgreSQL a un error de dominio: `ErrNoRows`
// pasa a apperr.NotFound; un duplicado a apperr.Conflict; una violación de
// `CHECK` a apperr.Invalid con el detalle por campo del invariante que se rompió
// (FR-014/FR-015/FR-016: la BD es la última línea de defensa y su 400 debe decir
// qué campo falla). El repository es el único punto que conoce pgx (arq. R4),
// así que la traducción ocurre aquí y no en el service.
func classify(err error, notFound, conflict string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.NotFound(notFound, apperr.WithCause(err))
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case codeUniqueViolation:
			return apperr.Conflict(conflict, apperr.WithCause(err))
		case codeCheckViolation:
			opts := []apperr.Option{apperr.WithCause(err)}
			if details := checkDetails(pgErr.ConstraintName); details != nil {
				opts = append(opts, apperr.WithDetails(details))
			}
			return apperr.Invalid(
				"Algún dato no cumple las reglas del método de donación",
				opts...,
			)
		}
	}
	return err
}

// checkDetails traduce el nombre de la restricción que se rompió al detalle por
// campo del contrato (el mismo campo que nombra la validación del service). Los
// `CHECK` que la migración `000012` declara sin nombre reciben de PostgreSQL los
// nombres en el orden en que aparecen en la tabla (`donation_methods_check`,
// `donation_methods_check1`, …): el nombre no es bonito pero sí estable, porque
// **ninguna migración aplicada se edita** (§VI). Un `CHECK` de otro módulo —o
// uno nuevo— devuelve nil y el error sale con el mensaje seguro y sin `details`
// (nunca se inventa el campo).
func checkDetails(constraint string) map[string]any {
	switch constraint {
	// Invariantes de coherencia (CHECK sin nombre, en el orden de 000012).
	case "donation_methods_check": // formato del número por tipo
		return detail("number", "el número no tiene el formato de su tipo")
	case "donation_methods_check1": // moneda por tipo
		return detail("currency", "la moneda es obligatoria en IBAN y no aplica en SINPE")
	case "donation_methods_check2": // holder_en exige holder_es
		return detail("holderEs", "el titular en inglés exige el titular en español")
	case "donation_texts_check": // body_en exige body_es (redundante con body_es NOT NULL)
		return detail("textEs", "el texto en inglés exige el texto en español")

	// Registros cerrados (mismo detalle que el contrato).
	case "donation_methods_kind_check":
		return detail("kind", "el tipo debe ser iban o sinpe")
	case "donation_methods_currency_check":
		return detail("currency", "la moneda debe ser CRC o USD")
	case "donation_methods_publication_state_check", "donation_texts_publication_state_check":
		return detail("publicationState", "el estado debe ser draft o published")
	case "donation_texts_text_key_check":
		return detail("textKey", "la clave debe ser intro o usage")

	// Longitudes.
	case "donation_methods_number_check":
		return detail("number", "el número admite entre 1 y 34 caracteres")
	case "donation_methods_institution_check":
		return detail("institution", "la entidad financiera admite entre 1 y 160 caracteres")
	case "donation_methods_holder_es_check":
		return detail("holderEs", "el titular admite entre 1 y 160 caracteres")
	case "donation_methods_holder_en_check":
		return detail("holderEn", "el titular admite entre 1 y 160 caracteres")
	case "donation_texts_body_es_check":
		return detail("textEs", "el texto admite entre 1 y 2.000 caracteres")
	case "donation_texts_body_en_check":
		return detail("textEn", "el texto admite entre 1 y 2.000 caracteres")

	// Textos recortados: el dato normalizado es el que se guarda (F2/P13).
	case "donation_methods_institution_check1":
		return detail("institution", "quita los espacios sobrantes del principio y del final")
	case "donation_methods_holder_es_check1":
		return detail("holderEs", "quita los espacios sobrantes del principio y del final")
	case "donation_methods_holder_en_check1":
		return detail("holderEn", "quita los espacios sobrantes del principio y del final")

	default:
		return nil
	}
}

// detail arma el detalle seguro de un campo con un solo motivo.
func detail(field, reason string) map[string]any {
	return map[string]any{field: reason}
}

// wrap traduce el error y lo envuelve con contexto de la operación, de modo que
// errors.Is/As sigan encontrando el error de dominio.
func wrap(err error, operation, notFound, conflict string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, classify(err, notFound, conflict))
}
