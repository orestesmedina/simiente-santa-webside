package agenda

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

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

// repository.go es la base del acceso a datos del dominio agenda: el
// constructor sobre el *pgxpool.Pool compartido, la transacción de escritura
// (`withTx`) y el mapeo de tipos entre pgtype.* y el dominio.
//
// No conoce HTTP ni las reglas de negocio (arq. R4): solo SQL (las consultas
// generadas por sqlc de T404), transacciones, el mapeo de tipos y la inserción
// de la auditoría dentro del tx. Los tipos pgtype.* de pgx se traducen AQUÍ y
// solo aquí (§8.1.6); el dominio usa uuid.UUID, time.Time y string.
//
// Las mutaciones reciben las filas de auditoría que deben insertarse en la MISMA
// transacción (variádicas): o el cambio y su registro, o nada (FR-017/R4-10).
// Ningún dominio importa a otro: la consulta `InsertAdminAction` se reutiliza de
// internal/db (R2).
type repository struct {
	q    *gendb.Queries
	pool *pgxpool.Pool
}

// NewRepository construye el repositorio del dominio agenda a partir del pool
// compartido que compone cmd/api. Devuelve el tipo concreto: la interfaz que
// consume el service la declara el propio service (arq. R3).
func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{q: gendb.New(pool), pool: pool}
}

// withTx ejecuta fn con un repository ligado a una transacción. Las mutaciones
// agrupan el cambio de contenido y su registro de auditoría en una única
// transacción (R4-10): o ambos o ninguno (FR-017/edge case).
func (r *repository) withTx(ctx context.Context, fn func(tx *repository) error) error {
	return database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&repository{q: gendb.New(tx), pool: r.pool})
	})
}

// insertAdminAction registra una acción administrativa en la transacción
// vigente (r ya está ligado al tx cuando se llama desde una mutación). Reutiliza
// la consulta compartida `InsertAdminAction` de internal/db (R4-10).
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
// mapeo que R4-10 permite repetir por dominio (no se comparte código de dominio,
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

// clockPattern es el patrón "HH:MM" en 24 h que exige el contrato, el `CHECK`
// de la base y la validación del service (T417). Vive aquí porque el mapeo de
// tipos lo necesita para no colar una hora que la BD rechazaría.
var clockPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

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

// pgDate traduce una fecha de calendario "YYYY-MM-DD" (DateFormat) al parámetro
// DATE. Una fecha mal formada devuelve NULL —el service la valida antes de
// llamar (T417)— y la base la rechaza si la columna es NOT NULL.
func pgDate(date string) pgtype.Date {
	parsed, err := time.Parse(DateFormat, strings.TrimSpace(date))
	if err != nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: parsed, Valid: true}
}

// dateString traduce una columna DATE a la fecha de calendario del dominio
// ("YYYY-MM-DD"); NULL devuelve cadena vacía.
func dateString(v pgtype.Date) string {
	if !v.Valid {
		return ""
	}
	return v.Time.Format(DateFormat)
}

// pgTime traduce una hora "HH:MM" (TimeFormat) al parámetro TIME (pgtype.Time
// cuenta microsegundos desde la medianoche). Una hora mal formada devuelve NULL
// —el service la valida antes de llamar (T417)—.
func pgTime(clock string) pgtype.Time {
	clock = strings.TrimSpace(clock)
	if !clockPattern.MatchString(clock) {
		return pgtype.Time{}
	}
	parsed, err := time.Parse(TimeFormat, clock)
	if err != nil {
		return pgtype.Time{}
	}
	microseconds := (int64(parsed.Hour())*3600 + int64(parsed.Minute())*60) * int64(time.Second/time.Microsecond)
	return pgtype.Time{Microseconds: microseconds, Valid: true}
}

// --- Traducción de errores de PostgreSQL ---

// Códigos SQLSTATE que el dominio conoce (skill postgres-db): 23505
// unique_violation, 23503 foreign_key_violation, 23514 check_violation.
const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
	codeCheckViolation      = "23514"
)

// classify traduce un error de PostgreSQL a un error de dominio: `ErrNoRows`
// pasa a apperr.NotFound; las violaciones de integridad conocidas a
// apperr.Conflict (duplicado del módulo y categoría en uso: la FK de
// `category_id` es ON DELETE RESTRICT, R4-5) o apperr.Invalid (un CHECK). El
// repository es el único punto que conoce pgx (arq. R4), así que la traducción
// ocurre aquí y no en el service.
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
		case codeUniqueViolation, codeForeignKeyViolation:
			return apperr.Conflict(conflict, apperr.WithCause(err))
		case codeCheckViolation:
			return apperr.Invalid(
				"Algún dato no cumple las reglas de la base de datos",
				apperr.WithCause(err),
			)
		}
	}
	return err
}

// wrap traduce el error y lo envuelve con contexto de la operación, de modo que
// errors.Is/As sigan encontrando el error de dominio.
func wrap(err error, operation, notFound, conflict string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, classify(err, notFound, conflict))
}
