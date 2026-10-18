package medios

import (
	"context"
	"errors"
	"fmt"
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

// repository.go es la base del acceso a datos del dominio medios: el
// constructor sobre el *pgxpool.Pool compartido, la transacción de escritura
// (`withTx`), el mapeo de tipos entre pgtype.* y el dominio (§8.1.6) y la
// inserción de la auditoría DENTRO de la misma transacción de la mutación
// (FR-022/P9-14).
//
// No conoce HTTP ni las reglas de negocio (arq. R4): solo SQL (las consultas
// generadas por sqlc de T903), transacciones, el mapeo de tipos y la inserción
// de la auditoría. Los tipos pgtype.* de pgx se traducen AQUÍ y solo aquí
// (§8.1.6): el dominio usa uuid.UUID, time.Time, string y sus enums; **ningún
// pgtype sale de este paquete de repository** (R4).
//
// Las mutaciones reciben las filas de auditoría que deben insertarse en la MISMA
// transacción (variádicas): o el cambio y su registro, o nada (FR-022, edge case
// «ninguna operación se aplica sin registro»). Ningún dominio importa a otro: la
// consulta `InsertAdminAction` se reutiliza de internal/db (R7-10/R2).
type repository struct {
	q    *gendb.Queries
	pool *pgxpool.Pool
}

// NewRepository construye el repositorio del dominio medios a partir del pool
// compartido que compone cmd/api. Devuelve el tipo concreto: la interfaz que
// consume el service la declara el propio service (arq. R3, T915).
func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{q: gendb.New(pool), pool: pool}
}

// withTx ejecuta fn con un repository ligado a una transacción. Las mutaciones
// agrupan el cambio de contenido y su registro de auditoría en una única
// transacción (P9-14/FR-022): o ambos o ninguno.
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
// misma transacción (FR-022, fail-closed): si una falla, la mutación no se
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
// mapeo que P9-14/R7-10 permite repetir por dominio (no se comparte código de
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

// textValue traduce una columna TEXT anulable a string ("" si NULL): es el
// caso de los enlaces que resuelve el `LEFT JOIN` por proveedor.
func textValue(v pgtype.Text) string {
	return v.String
}

// pgTextPtr traduce un *string del dominio al parámetro pgtype.Text.
func pgTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// pgText traduce un string ya normalizado: vacío se guarda como NULL (los
// campos opcionales, analyze I6).
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

// pgTime traduce una fecha del dominio al parámetro pgtype.Timestamptz.
func pgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// timePtr traduce una columna TIMESTAMPTZ anulable a *time.Time (nil si NULL):
// es el caso de las marcas sin resolver y de la unión.
func timePtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

// pgTimePtr traduce un *time.Time del dominio al parámetro pgtype.Timestamptz
// (NULL si es nil).
func pgTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgTime(*t)
}

// enumValue traduce una columna TEXT NOT NULL al enum del dominio.
func enumValue[T ~string](v string) T {
	return T(v)
}

// enumPtr traduce una columna TEXT anulable al enum del dominio (nil si NULL):
// una columna sin valor **nunca** se convierte en el valor cero del enum.
func enumPtr[T ~string](v pgtype.Text) *T {
	if !v.Valid {
		return nil
	}
	e := T(v.String)
	return &e
}

// pgEnumPtr traduce un enum opcional del dominio al parámetro pgtype.Text (NULL
// si es nil).
func pgEnumPtr[T ~string](v *T) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*v), Valid: true}
}

// --- Mapeo de filas sqlc → dominio ---

// mapEpisodio traduce la fila de una ficha devuelta por una escritura
// (`InsertMediaItem`/`UpdateMediaItem`), que no trae los enlaces.
func mapEpisodio(row gendb.MediaItem) Episodio {
	return Episodio{
		ID:                            uuidValue(row.ID),
		Titulo:                        row.Title,
		PublicadoEn:                   row.PublishedAt.Time,
		DescripcionEs:                 textPtr(row.DescriptionEs),
		DescripcionEn:                 textPtr(row.DescriptionEn),
		DescripcionEditada:            row.DescriptionEdited,
		Caratula:                      textPtr(row.CoverFile),
		Origen:                        enumValue[OrigenMedio](row.Origin),
		Visibilidad:                   enumValue[Visibilidad](row.Visibility),
		DuplicadoDeID:                 uuidPtr(row.DuplicateOfID),
		MarcaDuplicadoResueltaEn:      timePtr(row.DuplicateMarkResolvedAt),
		DisponibilidadComprobadaEn:    timePtr(row.AvailabilityCheckedAt),
		ResultadoDisponibilidad:       enumPtr[ResultadoDisponibilidad](row.AvailabilityResult),
		MarcaDisponibilidad:           enumPtr[MarcaDisponibilidad](row.AvailabilityMark),
		MarcaDisponibilidadResueltaEn: timePtr(row.AvailabilityMarkResolvedAt),
		UnidoAID:                      uuidPtr(row.MergedIntoID),
		CreadoEn:                      row.CreatedAt.Time,
		ActualizadoEn:                 row.UpdatedAt.Time,
	}
}

// mapEpisodioListado traduce la fila del listado del panel (misma forma que
// `mapEpisodio` más los dos enlaces del `LEFT JOIN`).
func mapEpisodioListado(row gendb.ListMediaItemsRow) Episodio {
	return Episodio{
		ID:                            uuidValue(row.ID),
		Titulo:                        row.Title,
		PublicadoEn:                   row.PublishedAt.Time,
		DescripcionEs:                 textPtr(row.DescriptionEs),
		DescripcionEn:                 textPtr(row.DescriptionEn),
		DescripcionEditada:            row.DescriptionEdited,
		Caratula:                      textPtr(row.CoverFile),
		Origen:                        enumValue[OrigenMedio](row.Origin),
		Visibilidad:                   enumValue[Visibilidad](row.Visibility),
		DuplicadoDeID:                 uuidPtr(row.DuplicateOfID),
		MarcaDuplicadoResueltaEn:      timePtr(row.DuplicateMarkResolvedAt),
		DisponibilidadComprobadaEn:    timePtr(row.AvailabilityCheckedAt),
		ResultadoDisponibilidad:       enumPtr[ResultadoDisponibilidad](row.AvailabilityResult),
		MarcaDisponibilidad:           enumPtr[MarcaDisponibilidad](row.AvailabilityMark),
		MarcaDisponibilidadResueltaEn: timePtr(row.AvailabilityMarkResolvedAt),
		UnidoAID:                      uuidPtr(row.MergedIntoID),
		VideoURL:                      textValue(row.VideoUrl),
		AudioURL:                      textValue(row.AudioUrl),
		CreadoEn:                      row.CreatedAt.Time,
		ActualizadoEn:                 row.UpdatedAt.Time,
	}
}

// mapEpisodioDetalle traduce la fila del detalle del panel (la misma forma que
// el listado; se declara aparte porque sqlc genera un tipo de fila por consulta).
func mapEpisodioDetalle(row gendb.GetMediaItemByIDRow) Episodio {
	return Episodio{
		ID:                            uuidValue(row.ID),
		Titulo:                        row.Title,
		PublicadoEn:                   row.PublishedAt.Time,
		DescripcionEs:                 textPtr(row.DescriptionEs),
		DescripcionEn:                 textPtr(row.DescriptionEn),
		DescripcionEditada:            row.DescriptionEdited,
		Caratula:                      textPtr(row.CoverFile),
		Origen:                        enumValue[OrigenMedio](row.Origin),
		Visibilidad:                   enumValue[Visibilidad](row.Visibility),
		DuplicadoDeID:                 uuidPtr(row.DuplicateOfID),
		MarcaDuplicadoResueltaEn:      timePtr(row.DuplicateMarkResolvedAt),
		DisponibilidadComprobadaEn:    timePtr(row.AvailabilityCheckedAt),
		ResultadoDisponibilidad:       enumPtr[ResultadoDisponibilidad](row.AvailabilityResult),
		MarcaDisponibilidad:           enumPtr[MarcaDisponibilidad](row.AvailabilityMark),
		MarcaDisponibilidadResueltaEn: timePtr(row.AvailabilityMarkResolvedAt),
		UnidoAID:                      uuidPtr(row.MergedIntoID),
		VideoURL:                      textValue(row.VideoUrl),
		AudioURL:                      textValue(row.AudioUrl),
		CreadoEn:                      row.CreatedAt.Time,
		ActualizadoEn:                 row.UpdatedAt.Time,
	}
}

// mapEpisodioPublico traduce la fila de la sección pública: la consulta no
// selecciona `visibility` ni las marcas (el predicado ya está en el `WHERE` y el
// DTO público nunca las expone), así que la ficha se completa con lo que la
// consulta garantiza —visible, sin absorción y sin par abierto—.
func mapEpisodioPublico(row gendb.ListPublishedMediaItemsRow) Episodio {
	return Episodio{
		ID:            uuidValue(row.ID),
		Titulo:        row.Title,
		PublicadoEn:   row.PublishedAt.Time,
		DescripcionEs: textPtr(row.DescriptionEs),
		DescripcionEn: textPtr(row.DescriptionEn),
		Caratula:      textPtr(row.CoverFile),
		Origen:        enumValue[OrigenMedio](row.Origin),
		Visibilidad:   VisibilidadVisible,
		VideoURL:      textValue(row.VideoUrl),
		AudioURL:      textValue(row.AudioUrl),
		CreadoEn:      row.CreatedAt.Time,
		ActualizadoEn: row.UpdatedAt.Time,
	}
}

// mapEpisodioPortada traduce la fila del bloque de la portada: mismo predicado y
// mismas columnas que la sección (sqlc genera su propio tipo de fila).
func mapEpisodioPortada(row gendb.ListPortadaMediaItemsRow) Episodio {
	return Episodio{
		ID:            uuidValue(row.ID),
		Titulo:        row.Title,
		PublicadoEn:   row.PublishedAt.Time,
		DescripcionEs: textPtr(row.DescriptionEs),
		DescripcionEn: textPtr(row.DescriptionEn),
		Caratula:      textPtr(row.CoverFile),
		Origen:        enumValue[OrigenMedio](row.Origin),
		Visibilidad:   VisibilidadVisible,
		VideoURL:      textValue(row.VideoUrl),
		AudioURL:      textValue(row.AudioUrl),
		CreadoEn:      row.CreatedAt.Time,
		ActualizadoEn: row.UpdatedAt.Time,
	}
}

// mapFuenteEpisodio traduce la fila de una fuente de importación.
func mapFuenteEpisodio(row gendb.MediaEpisodeSource) FuenteEpisodio {
	return FuenteEpisodio{
		ID:            uuidValue(row.ID),
		EpisodioID:    uuidValue(row.ItemID),
		Proveedor:     enumValue[Proveedor](row.Provider),
		IDExterno:     row.ExternalID,
		TituloFeed:    row.FeedTitle,
		Enlace:        row.Url,
		CreadoEn:      row.CreatedAt.Time,
		ActualizadoEn: row.UpdatedAt.Time,
	}
}

// mapSyncRun traduce la fila de un registro de búsqueda.
func mapSyncRun(row gendb.MediaSyncRun) SyncRun {
	return SyncRun{
		ID:            uuidValue(row.ID),
		ActorUserID:   uuidPtr(row.ActorUserID),
		IniciadoEn:    row.StartedAt.Time,
		TerminadoEn:   row.FinishedAt.Time,
		Importados:    int(row.ImportedCount),
		Unidos:        int(row.JoinedCount),
		Marcados:      int(row.MarkedCount),
		Ocultos:       int(row.HiddenCount),
		Comprobados:   int(row.AvailabilityCheckedCount),
		Inconclusos:   int(row.AvailabilityInconclusiveCount),
		EstadoYouTube: enumValue[EstadoFeed](row.YoutubeStatus),
		EstadoPodcast: enumValue[EstadoFeed](row.PodcastStatus),
		CreadoEn:      row.CreatedAt.Time,
		ActualizadoEn: row.UpdatedAt.Time,
	}
}

// --- Traducción de errores de PostgreSQL ---

// Códigos SQLSTATE que el dominio conoce (skill postgres-db): 23505
// unique_violation y 23514 check_violation.
const (
	codeUniqueViolation = "23505"
	codeCheckViolation  = "23514"
)

// Mensajes de los errores de dominio del módulo (seguros para el cliente).
const (
	// messageEpisodioNoExiste es el 404 de una ficha inexistente.
	messageEpisodioNoExiste = "El episodio no existe"
	// messageFuenteDuplicada es el 409 de un elemento de feed ya importado (la
	// idempotencia normal la absorbe el `ON CONFLICT`: esto es la última línea
	// de defensa de la BD, R9-11).
	messageFuenteDuplicada = "Ese episodio ya está importado"
)

// classify traduce un error de PostgreSQL a un error de dominio: `ErrNoRows`
// pasa a apperr.NotFound; un duplicado a apperr.Conflict; una violación de
// `CHECK` a apperr.Invalid con el detalle por campo del invariante que se rompió
// (la BD es la última línea de defensa y su 400 debe decir qué campo falla). El
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
		case codeUniqueViolation:
			return apperr.Conflict(conflict, apperr.WithCause(err))
		case codeCheckViolation:
			opts := []apperr.Option{apperr.WithCause(err)}
			if details := checkDetails(pgErr.ConstraintName); details != nil {
				opts = append(opts, apperr.WithDetails(details))
			}
			return apperr.Invalid(
				"Algún dato no cumple las reglas del episodio",
				opts...,
			)
		}
	}
	return err
}

// checkDetails traduce el nombre de la restricción que se rompió al detalle por
// campo del contrato (el mismo campo que nombra la validación del service). Los
// nombres son los que PostgreSQL asigna a las restricciones sin nombre de la
// migración `000014` (`<tabla>_<columna>_check`, con sufijo numérico cuando el
// nombre ya está tomado; verificado en T914): el nombre no es bonito pero sí
// estable, porque **ninguna migración aplicada se edita** (§VI). Un `CHECK` de
// otro módulo —o uno nuevo— devuelve nil y el error sale con el mensaje seguro y
// sin `details` (nunca se inventa el campo).
func checkDetails(constraint string) map[string]any {
	switch constraint {
	// Invariantes de coherencia de `media_items` (CHECK sin nombre, en el orden
	// de 000014).
	case "media_items_check": // description_en exige description_es
		return detail("descriptionEs", "la descripción en inglés exige la descripción en español")
	case "media_items_check1": // resolución del duplicado sin par
		return detail("duplicateOfId", "la resolución del duplicado exige el par")
	case "media_items_check2": // resolución de disponibilidad sin marca
		return detail("availabilityMark", "la resolución exige la marca de disponibilidad")

	// Registros cerrados del contrato.
	case "media_items_origin_check":
		return detail("origin", "el origen debe ser video, audio o both")
	case "media_items_visibility_check":
		return detail("visibility", "la visibilidad debe ser visible u oculta")
	case "media_items_availability_mark_check":
		return detail("availabilityMark", "la marca debe ser unavailable o missing_from_feed")
	case "media_items_availability_result_check":
		return detail("availabilityResult", "el resultado debe ser available, unavailable o inconclusive")
	case "media_episode_sources_provider_check":
		return detail("provider", "el proveedor debe ser youtube o podcast")
	case "media_sync_runs_youtube_status_check":
		return detail("youtube", "el estado del feed debe ser ok o failed")
	case "media_sync_runs_podcast_status_check":
		return detail("podcast", "el estado del feed debe ser ok o failed")

	// Longitudes de `media_items`.
	case "media_items_title_check":
		return detail("title", "el nombre admite entre 1 y 300 caracteres")
	case "media_items_description_es_check":
		return detail("descriptionEs", "la descripción admite entre 1 y 300 caracteres")
	case "media_items_description_en_check":
		return detail("descriptionEn", "la descripción admite entre 1 y 300 caracteres")
	case "media_items_cover_file_check":
		return detail("coverFile", "el nombre del archivo admite entre 1 y 120 caracteres")

	// Textos recortados: el dato normalizado es el que se guarda (F2/P13).
	case "media_items_title_check1":
		return detail("title", "quita los espacios sobrantes del principio y del final")
	case "media_items_description_es_check1":
		return detail("descriptionEs", "quita los espacios sobrantes del principio y del final")
	case "media_items_description_en_check1":
		return detail("descriptionEn", "quita los espacios sobrantes del principio y del final")

	// `media_episode_sources`: forma del enlace por proveedor (R9-8).
	case "media_episode_sources_url_check":
		return detail("url", "el enlace admite entre 1 y 500 caracteres")
	case "media_episode_sources_url_check1":
		return detail("url", "el enlace debe empezar por https://")
	case "media_episode_sources_check":
		return detail("url", "el enlace no tiene la forma del proveedor")
	case "media_episode_sources_external_id_check":
		return detail("externalId", "el identificador admite entre 1 y 120 caracteres")
	case "media_episode_sources_external_id_check1":
		return detail("externalId", "quita los espacios sobrantes del principio y del final")
	case "media_episode_sources_feed_title_check":
		return detail("feedTitle", "el título del feed admite entre 1 y 300 caracteres")
	case "media_episode_sources_feed_title_check1":
		return detail("feedTitle", "quita los espacios sobrantes del principio y del final")

	// `media_sync_runs`: la búsqueda no puede terminar antes de empezar.
	case "media_sync_runs_check":
		return detail("finishedAt", "la búsqueda no puede terminar antes de empezar")
	case "media_sync_runs_imported_count_check":
		return detail("importedCount", "el conteo no puede ser negativo")
	case "media_sync_runs_joined_count_check":
		return detail("joinedCount", "el conteo no puede ser negativo")
	case "media_sync_runs_marked_count_check":
		return detail("markedCount", "el conteo no puede ser negativo")
	case "media_sync_runs_hidden_count_check":
		return detail("hiddenCount", "el conteo no puede ser negativo")
	case "media_sync_runs_availability_checked_count_check":
		return detail("availabilityCheckedCount", "el conteo no puede ser negativo")
	case "media_sync_runs_availability_inconclusive_count_check":
		return detail("availabilityInconclusiveCount", "el conteo no puede ser negativo")

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
