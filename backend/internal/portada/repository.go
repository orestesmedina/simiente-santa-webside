package portada

import (
	"context"
	"errors"
	"fmt"

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

// repository implementa el acceso a datos del dominio portada sobre el
// *pgxpool.Pool compartido, con las consultas generadas por sqlc (T308). No
// conoce HTTP ni las reglas de negocio (arq. R4): solo SQL, transacciones, el
// mapeo de tipos y la inserción de la auditoría dentro del tx. Los tipos
// pgtype.* de pgx se traducen AQUÍ y solo aquí (§8.1.6); el dominio usa
// uuid.UUID, time.Time y string.
type repository struct {
	q    *gendb.Queries
	pool *pgxpool.Pool
}

// NewRepository construye el repositorio del dominio portada a partir del pool
// compartido que compone cmd/api. Devuelve el tipo concreto: la interfaz que
// consume el service la declara el propio service (arq. R3).
func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{q: gendb.New(pool), pool: pool}
}

// withTx ejecuta fn con un repository ligado a una transacción. Las mutaciones
// agrupan el cambio de contenido y su registro de auditoría en una única
// transacción (R3-11): o ambos o ninguno (FR-017/edge case).
func (r *repository) withTx(ctx context.Context, fn func(tx *repository) error) error {
	return database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&repository{q: r.q.WithTx(tx), pool: r.pool})
	})
}

// insertAdminAction registra una acción administrativa en la transacción
// vigente (r ya está ligado al tx cuando se llama desde una mutación). Reutiliza
// la consulta compartida `InsertAdminAction` de internal/db (R3-11): ningún
// dominio importa a otro.
func (r *repository) insertAdminAction(ctx context.Context, action audit.Action) error {
	if err := action.Validate(); err != nil {
		return fmt.Errorf("insert admin action: %w", err)
	}
	if _, err := r.q.InsertAdminAction(ctx, auditParams(action)); err != nil {
		return wrap(err, "insert admin action", "", "")
	}
	return nil
}

// auditParams traduce una audit.Action al parámetro generado por sqlc. Es el
// mapeo que R3-11 permite repetir por dominio (no se comparte código de
// dominio, solo la consulta SQL).
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

// pgText traduce un string ya normalizado: vacío se guarda como NULL (los
// campos `*_en` opcionales y las referencias de imagen, analyze I6).
func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// pgLabel traduce la etiqueta de un objetivo de acción: un valor vacío se
// guarda como NULL (target_label es anulable salvo para `content`, que el
// registro de audit exige y valida).
func pgLabel(label string) pgtype.Text {
	if label == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: label, Valid: true}
}

// mapHomeIdentity traduce la fila del singleton de identidad.
func mapHomeIdentity(row gendb.HomeIdentity) Identity {
	return Identity{
		ID:               uuidValue(row.ID),
		NameEs:           row.NameEs,
		NameEn:           textPtr(row.NameEn),
		TaglineEs:        textPtr(row.TaglineEs),
		TaglineEn:        textPtr(row.TaglineEn),
		MissionEs:        textPtr(row.MissionEs),
		MissionEn:        textPtr(row.MissionEn),
		VisionEs:         textPtr(row.VisionEs),
		VisionEn:         textPtr(row.VisionEn),
		LogoFile:         textPtr(row.LogoFile),
		LogoAltEs:        textPtr(row.LogoAltEs),
		LogoAltEn:        textPtr(row.LogoAltEn),
		CoverImageFile:   textPtr(row.CoverImageFile),
		CoverImageAltEs:  textPtr(row.CoverImageAltEs),
		CoverImageAltEn:  textPtr(row.CoverImageAltEn),
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapHomeAbout traduce la fila del singleton «quiénes somos».
func mapHomeAbout(row gendb.HomeAbout) About {
	return About{
		ID:               uuidValue(row.ID),
		TextEs:           row.TextEs,
		TextEn:           textPtr(row.TextEn),
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapHomeContact traduce la fila del singleton de contacto.
func mapHomeContact(row gendb.HomeContact) Contact {
	return Contact{
		ID:               uuidValue(row.ID),
		AddressEs:        row.AddressEs,
		AddressEn:        textPtr(row.AddressEn),
		Email:            row.Email,
		Phone:            row.Phone,
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapHomeService traduce una fila del horario.
func mapHomeService(row gendb.HomeService) Service {
	return Service{
		ID:               uuidValue(row.ID),
		DayOfWeek:        int(row.DayOfWeek),
		StartTime:        row.StartTime,
		EndTime:          textPtr(row.EndTime),
		NameEs:           row.NameEs,
		NameEn:           textPtr(row.NameEn),
		DescriptionEs:    textPtr(row.DescriptionEs),
		DescriptionEn:    textPtr(row.DescriptionEn),
		PlaceEs:          row.PlaceEs,
		PlaceEn:          textPtr(row.PlaceEn),
		PublicationState: PublicationState(row.PublicationState),
		SortOrder:        int(row.SortOrder),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapHomeWhatsappChannel traduce una fila de canal de WhatsApp.
func mapHomeWhatsappChannel(row gendb.HomeWhatsappChannel) WhatsappChannel {
	return WhatsappChannel{
		ID:               uuidValue(row.ID),
		Kind:             row.Kind,
		Destination:      row.Destination,
		NameEs:           row.NameEs,
		NameEn:           textPtr(row.NameEn),
		PublicationState: PublicationState(row.PublicationState),
		SortOrder:        int(row.SortOrder),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapHomeSocialLink traduce una fila de enlace de red social.
func mapHomeSocialLink(row gendb.HomeSocialLink) SocialLink {
	return SocialLink{
		ID:               uuidValue(row.ID),
		Network:          row.Network,
		URL:              row.Url,
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// --- Traducción de errores de PostgreSQL ---

// Códigos SQLSTATE que el dominio conoce (skill postgres-db): 23505
// unique_violation, 23503 foreign_key_violation, 23514 check_violation.
const (
	codeUniqueViolation = "23505"
	codeCheckViolation  = "23514"
)

// classify traduce un error de PostgreSQL a un error de dominio: `ErrNoRows`
// pasa a apperr.NotFound y las violaciones de integridad conocidas a
// apperr.Conflict/apperr.Invalid. El repository es el único punto que conoce
// pgx (arq. R4), así que la traducción ocurre aquí y no en el service.
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
			return apperr.Invalid(
				"Algún dato no cumple las reglas de la base de datos",
				apperr.WithCause(err),
			)
		}
	}
	return err
}

// wrap traduce el error y lo envuelve con contexto de la operación, de modo
// que errors.Is/As sigan encontrando el error de dominio.
func wrap(err error, operation, notFound, conflict string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, classify(err, notFound, conflict))
}
