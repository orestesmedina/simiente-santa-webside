package ministerios

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

// repository.go es la base del acceso a datos del dominio ministerios: el
// constructor sobre el *pgxpool.Pool compartido, la transacción de escritura
// (`withTx`) y el mapeo de tipos entre pgtype.* y el dominio.
//
// No conoce HTTP ni las reglas de negocio (arq. R4): solo SQL (las consultas
// generadas por sqlc de T603), transacciones, el mapeo de tipos y la inserción
// de la auditoría dentro del tx. Los tipos pgtype.* de pgx se traducen AQUÍ y
// solo aquí (§8.1.6); el dominio usa uuid.UUID, time.Time y string.
//
// Las mutaciones reciben las filas de auditoría que deben insertarse en la MISMA
// transacción (variádicas): o el cambio y su registro, o nada (FR-017/R6-12).
// Ningún dominio importa a otro: la consulta `InsertAdminAction` se reutiliza de
// internal/db (R2).
type repository struct {
	q    *gendb.Queries
	pool *pgxpool.Pool
}

// NewRepository construye el repositorio del dominio ministerios a partir del
// pool compartido que compone cmd/api. Devuelve el tipo concreto: la interfaz
// que consume el service la declara el propio service (arq. R3).
func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{q: gendb.New(pool), pool: pool}
}

// withTx ejecuta fn con un repository ligado a una transacción. Las mutaciones
// agrupan el cambio de contenido y su registro de auditoría en una única
// transacción (R6-12): o ambos o ninguno (FR-017/edge case).
func (r *repository) withTx(ctx context.Context, fn func(tx *repository) error) error {
	return database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&repository{q: gendb.New(tx), pool: r.pool})
	})
}

// insertAdminAction registra una acción administrativa en la transacción
// vigente (r ya está ligado al tx cuando se llama desde una mutación). Reutiliza
// la consulta compartida `InsertAdminAction` de internal/db (R6-12).
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
// mapeo que R6-12 permite repetir por dominio (no se comparte código de dominio,
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

// maskedText traduce una de las cuatro columnas del encargado que sqlc tipa como
// `interface{}`: no sabe inferir el tipo de un `CASE WHEN leader_visible THEN …
// END` (sqlc v1.31.1) y el `CASE` es literal a propósito (R6-2/P6-2, anotado en
// queries/ministries.sql). Con la ficha oculta la fila llega con NULL y aquí se
// convierte en "sin dato"; cualquier valor que no sea texto tampoco puede
// publicarse (fail-closed) y una cadena vacía es "sin dato". El tipo
// `interface{}` no sale de este paquete (§8.1.6, R4).
func maskedText(v interface{}) *string {
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

// mapMinistry traduce la fila completa de un ministerio (sqlc) a la entidad del
// dominio. La usan las consultas del panel, que sí devuelven la ficha del
// encargado íntegra aunque no esté publicada (dato guardado ≠ dato publicado).
func mapMinistry(row gendb.Ministry) Ministry {
	return Ministry{
		ID:               uuidValue(row.ID),
		NameEs:           row.NameEs,
		NameEn:           textPtr(row.NameEn),
		DescriptionEs:    row.DescriptionEs,
		DescriptionEn:    textPtr(row.DescriptionEn),
		LeaderNameEs:     textPtr(row.LeaderNameEs),
		LeaderNameEn:     textPtr(row.LeaderNameEn),
		LeaderPhone:      textPtr(row.LeaderPhone),
		LeaderEmail:      textPtr(row.LeaderEmail),
		LeaderVisible:    row.LeaderVisible,
		WhatsappURL:      textPtr(row.WhatsappUrl),
		ScheduleEs:       textPtr(row.ScheduleEs),
		ScheduleEn:       textPtr(row.ScheduleEn),
		PlaceEs:          textPtr(row.PlaceEs),
		PlaceEn:          textPtr(row.PlaceEn),
		ImageFile:        textPtr(row.ImageFile),
		ImageAltEs:       textPtr(row.ImageAltEs),
		ImageAltEn:       textPtr(row.ImageAltEn),
		PublicationState: PublicationState(row.PublicationState),
		SortOrder:        int(row.SortOrder),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// --- Traducción de errores de PostgreSQL ---

// Códigos SQLSTATE que el dominio conoce (skill postgres-db): 23505
// unique_violation y 23514 check_violation. La tabla `ministries` no tiene
// `UNIQUE` (el duplicado lo detecta `ExistsMinistryDuplicate`, R6-11) ni claves
// foráneas (el encargado es un campo del ministerio, R6-1), pero el duplicado se
// traduce igual porque el service lo comprueba con una consulta y la base es la
// última línea de defensa.
const (
	codeUniqueViolation = "23505"
	codeCheckViolation  = "23514"
)

// classify traduce un error de PostgreSQL a un error de dominio: `ErrNoRows`
// pasa a apperr.NotFound; un duplicado a apperr.Conflict (R6-11); una violación
// de `CHECK` a apperr.Invalid con el detalle por campo del invariante que se
// rompió (FR-014/FR-015/FR-020: la BD es la última línea de defensa y su 400
// debe decir qué campo falla). El repository es el único punto que conoce pgx
// (arq. R4), así que la traducción ocurre aquí y no en el service.
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
				"Algún dato no cumple las reglas del ministerio",
				opts...,
			)
		}
	}
	return err
}

// checkDetails traduce el nombre de la restricción que se rompió al detalle por
// campo del contrato (el mismo campo que nombra la validación del service). Los
// `CHECK` que la migración `000010` declara sin nombre reciben de PostgreSQL los
// nombres en el orden en que aparecen en la tabla (`ministries_check`,
// `ministries_check1`, …, `ministries_check8`): el nombre no es bonito pero sí
// estable, porque **ninguna migración aplicada se edita** (§VI). Un `CHECK` de
// otro módulo —o uno nuevo— devuelve nil y el error sale con el mensaje seguro y
// sin `details` (nunca se inventa el campo).
func checkDetails(constraint string) map[string]any {
	switch constraint {
	// Invariantes de coherencia (CHECK sin nombre, en el orden de 000010).
	case "ministries_check":
		return detail("leaderEmail", "el celular y el correo del encargado no pueden ir juntos")
	case "ministries_check1":
		return detail("leaderNameEs", "el contacto del encargado exige su nombre")
	case "ministries_check2":
		return detail("nameEs", "el nombre en inglés exige el nombre en español")
	case "ministries_check3":
		return detail("descriptionEs", "la descripción en inglés exige la descripción en español")
	case "ministries_check4":
		return detail("leaderNameEs", "el nombre del encargado en inglés exige el español")
	case "ministries_check5":
		return detail("scheduleEs", "el horario en inglés exige el horario en español")
	case "ministries_check6":
		return detail("placeEs", "el lugar en inglés exige el lugar en español")
	case "ministries_check7":
		return detail("imageAltEs", "el texto alternativo en inglés exige el español")
	case "ministries_check8":
		return detail("imageAltEs", "toda imagen lleva su texto alternativo en español")

	// Enlace: solo `https` sin espacios (FR-015).
	case "ministries_whatsapp_url_check1":
		return detail("whatsappUrl", "el enlace debe empezar por https:// y no llevar espacios")

	// Estado de publicación (el registro del contrato).
	case "ministries_publication_state_check":
		return detail("publicationState", "el estado debe ser draft o published")

	// Textos recortados: el dato normalizado es el que se guarda (A6-9).
	case "ministries_name_es_check1":
		return detail("nameEs", "quita los espacios sobrantes del principio y del final")
	case "ministries_leader_name_es_check1":
		return detail("leaderNameEs", "quita los espacios sobrantes del principio y del final")
	case "ministries_schedule_es_check1":
		return detail("scheduleEs", "quita los espacios sobrantes del principio y del final")
	case "ministries_place_es_check1":
		return detail("placeEs", "quita los espacios sobrantes del principio y del final")
	case "ministries_whatsapp_url_check2":
		return detail("whatsappUrl", "quita los espacios sobrantes del principio y del final")
	case "ministries_leader_phone_check1":
		return detail("leaderPhone", "quita los espacios sobrantes del principio y del final")
	case "ministries_leader_email_check1":
		return detail("leaderEmail", "quita los espacios sobrantes del principio y del final")

	// Longitudes (los mismos límites del contrato y del service, FR-014).
	case "ministries_name_es_check", "ministries_name_en_check":
		return detail("nameEs", "el nombre admite entre 1 y 160 caracteres")
	case "ministries_description_es_check", "ministries_description_en_check":
		return detail("descriptionEs", "la descripción admite entre 1 y 2.000 caracteres")
	case "ministries_leader_name_es_check", "ministries_leader_name_en_check":
		return detail("leaderNameEs", "el nombre del encargado admite entre 1 y 160 caracteres")
	case "ministries_leader_phone_check":
		return detail("leaderPhone", "el celular admite entre 7 y 32 caracteres")
	case "ministries_leader_email_check":
		return detail("leaderEmail", "el correo admite entre 3 y 254 caracteres")
	case "ministries_schedule_es_check", "ministries_schedule_en_check":
		return detail("scheduleEs", "el horario admite entre 1 y 200 caracteres")
	case "ministries_place_es_check", "ministries_place_en_check":
		return detail("placeEs", "el lugar admite entre 1 y 200 caracteres")
	case "ministries_image_file_check":
		return detail("imageFile", "el nombre del archivo admite entre 1 y 120 caracteres")
	case "ministries_image_alt_es_check", "ministries_image_alt_en_check":
		return detail("imageAltEs", "el texto alternativo admite entre 1 y 300 caracteres")
	case "ministries_whatsapp_url_check":
		return detail("whatsappUrl", "el enlace admite entre 1 y 500 caracteres")

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
