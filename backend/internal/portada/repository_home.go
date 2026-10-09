package portada

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/database"
)

// Claves de las cerraduras de asesoramiento transaccionales (R3-4). Cada
// singleton tiene la suya (distinta de la del guard anti-bloqueo de F2,
// 726112001): serializan dos primeras escrituras simultáneas para que el upsert
// deje una única fila. Se liberan al cerrar la transacción.
const (
	lockHomeIdentity int64 = 730011001
	lockHomeAbout    int64 = 730011002
	lockHomeContact  int64 = 730011003
)

// withSingletonTx ejecuta fn dentro de una transacción con la cerradura de
// asesoramiento del singleton: el upsert y sus filas de auditoría comparten tx
// (R3-11/FR-017), de modo que o ambos o ninguno. La cerradura se toma con SQL
// parametrizado sobre el tx (no en la consulta del upsert: R3-4).
func (r *repository) withSingletonTx(ctx context.Context, lockKey int64, fn func(tx *repository) error) error {
	return database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey); err != nil {
			return fmt.Errorf("advisory lock: %w", err)
		}
		return fn(&repository{q: gendb.New(tx), pool: r.pool})
	})
}

// --- Identidad ---

// UpsertHomeIdentity crea o reemplaza el singleton de identidad (FR-002). Las
// acciones de auditoría (opcionales) se insertan en la MISMA transacción
// (FR-017): si una falla, el contenido no se aplica.
func (r *repository) UpsertHomeIdentity(ctx context.Context, identity Identity, actions ...audit.Action) (Identity, error) {
	var result Identity
	err := r.withSingletonTx(ctx, lockHomeIdentity, func(tx *repository) error {
		row, err := tx.q.UpsertHomeIdentity(ctx, gendb.UpsertHomeIdentityParams{
			NameEs:           identity.NameEs,
			NameEn:           pgTextPtr(identity.NameEn),
			TaglineEs:        pgTextPtr(identity.TaglineEs),
			TaglineEn:        pgTextPtr(identity.TaglineEn),
			MissionEs:        pgTextPtr(identity.MissionEs),
			MissionEn:        pgTextPtr(identity.MissionEn),
			VisionEs:         pgTextPtr(identity.VisionEs),
			VisionEn:         pgTextPtr(identity.VisionEn),
			LogoFile:         pgTextPtr(identity.LogoFile),
			LogoAltEs:        pgTextPtr(identity.LogoAltEs),
			LogoAltEn:        pgTextPtr(identity.LogoAltEn),
			CoverImageFile:   pgTextPtr(identity.CoverImageFile),
			CoverImageAltEs:  pgTextPtr(identity.CoverImageAltEs),
			CoverImageAltEn:  pgTextPtr(identity.CoverImageAltEn),
			PublicationState: string(identity.PublicationState),
		})
		if err != nil {
			return wrap(err, "upsert home identity", "", "")
		}
		for _, action := range actions {
			if err := tx.insertAdminAction(ctx, action); err != nil {
				return err
			}
		}
		result = mapHomeIdentity(row)
		return nil
	})
	if err != nil {
		return Identity{}, err
	}
	return result, nil
}

// GetHomeIdentity devuelve el singleton para el panel (borradores incluidos);
// el segundo valor es false si todavía no se ha guardado (el contrato lo tipa
// como null).
func (r *repository) GetHomeIdentity(ctx context.Context) (Identity, bool, error) {
	row, err := r.q.GetHomeIdentity(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, false, nil
		}
		return Identity{}, false, wrap(err, "get home identity", "", "")
	}
	return mapHomeIdentity(row), true, nil
}

// GetHomeIdentityPublished devuelve la identidad SOLO si está publicada; el
// segundo valor es false cuando no hay fila o está en borrador (FR-013).
func (r *repository) GetHomeIdentityPublished(ctx context.Context) (Identity, bool, error) {
	row, err := r.q.GetHomeIdentityPublished(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, false, nil
		}
		return Identity{}, false, wrap(err, "get home identity published", "", "")
	}
	return mapHomeIdentity(row), true, nil
}

// --- «Quiénes somos» ---

// UpsertHomeAbout crea o reemplaza el singleton «quiénes somos» (FR-003) con su
// auditoría en la misma transacción (FR-017).
func (r *repository) UpsertHomeAbout(ctx context.Context, about About, actions ...audit.Action) (About, error) {
	var result About
	err := r.withSingletonTx(ctx, lockHomeAbout, func(tx *repository) error {
		row, err := tx.q.UpsertHomeAbout(ctx, gendb.UpsertHomeAboutParams{
			TextEs:           about.TextEs,
			TextEn:           pgTextPtr(about.TextEn),
			PublicationState: string(about.PublicationState),
		})
		if err != nil {
			return wrap(err, "upsert home about", "", "")
		}
		for _, action := range actions {
			if err := tx.insertAdminAction(ctx, action); err != nil {
				return err
			}
		}
		result = mapHomeAbout(row)
		return nil
	})
	if err != nil {
		return About{}, err
	}
	return result, nil
}

// GetHomeAbout devuelve el singleton para el panel; false si no se ha guardado.
func (r *repository) GetHomeAbout(ctx context.Context) (About, bool, error) {
	row, err := r.q.GetHomeAbout(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return About{}, false, nil
		}
		return About{}, false, wrap(err, "get home about", "", "")
	}
	return mapHomeAbout(row), true, nil
}

// GetHomeAboutPublished devuelve «quiénes somos» solo si está publicado.
func (r *repository) GetHomeAboutPublished(ctx context.Context) (About, bool, error) {
	row, err := r.q.GetHomeAboutPublished(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return About{}, false, nil
		}
		return About{}, false, wrap(err, "get home about published", "", "")
	}
	return mapHomeAbout(row), true, nil
}

// --- Contacto ---

// UpsertHomeContact crea o reemplaza el singleton de contacto (FR-007) con su
// auditoría en la misma transacción (FR-017).
func (r *repository) UpsertHomeContact(ctx context.Context, contact Contact, actions ...audit.Action) (Contact, error) {
	var result Contact
	err := r.withSingletonTx(ctx, lockHomeContact, func(tx *repository) error {
		row, err := tx.q.UpsertHomeContact(ctx, gendb.UpsertHomeContactParams{
			AddressEs:        contact.AddressEs,
			AddressEn:        pgTextPtr(contact.AddressEn),
			Email:            contact.Email,
			Phone:            contact.Phone,
			PublicationState: string(contact.PublicationState),
		})
		if err != nil {
			return wrap(err, "upsert home contact", "", "")
		}
		for _, action := range actions {
			if err := tx.insertAdminAction(ctx, action); err != nil {
				return err
			}
		}
		result = mapHomeContact(row)
		return nil
	})
	if err != nil {
		return Contact{}, err
	}
	return result, nil
}

// GetHomeContact devuelve el singleton para el panel; false si no se ha guardado.
func (r *repository) GetHomeContact(ctx context.Context) (Contact, bool, error) {
	row, err := r.q.GetHomeContact(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Contact{}, false, nil
		}
		return Contact{}, false, wrap(err, "get home contact", "", "")
	}
	return mapHomeContact(row), true, nil
}

// GetHomeContactPublished devuelve el contacto solo si está publicado.
func (r *repository) GetHomeContactPublished(ctx context.Context) (Contact, bool, error) {
	row, err := r.q.GetHomeContactPublished(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Contact{}, false, nil
		}
		return Contact{}, false, wrap(err, "get home contact published", "", "")
	}
	return mapHomeContact(row), true, nil
}

// --- Archivos ---

// IsHomeFilePublished indica si un nombre de archivo está referenciado por
// contenido PUBLICADO (analyze C4): la descarga pública solo sirve estos
// archivos (el resto responde 404, FR-013).
func (r *repository) IsHomeFilePublished(ctx context.Context, fileName string) (bool, error) {
	published, err := r.q.IsHomeFilePublished(ctx, pgtype.Text{String: fileName, Valid: fileName != ""})
	if err != nil {
		return false, wrap(err, "is home file published", "", "")
	}
	return published, nil
}

// --- Horario (servicios) ---

// insertActions ejecuta las filas de auditoría de una mutación de colección
// dentro de su misma transacción (FR-017, fail-closed).
func insertActions(ctx context.Context, tx *repository, actions []audit.Action) error {
	for _, action := range actions {
		if err := tx.insertAdminAction(ctx, action); err != nil {
			return err
		}
	}
	return nil
}

// InsertHomeService da de alta un servicio del horario (FR-004) con su
// auditoría en la misma transacción (FR-017).
func (r *repository) InsertHomeService(ctx context.Context, service Service, actions ...audit.Action) (Service, error) {
	var result Service
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertHomeService(ctx, gendb.InsertHomeServiceParams{
			DayOfWeek:        int16(service.DayOfWeek),
			StartTime:        service.StartTime,
			EndTime:          pgTextPtr(service.EndTime),
			NameEs:           service.NameEs,
			NameEn:           pgTextPtr(service.NameEn),
			DescriptionEs:    pgTextPtr(service.DescriptionEs),
			DescriptionEn:    pgTextPtr(service.DescriptionEn),
			PlaceEs:          service.PlaceEs,
			PlaceEn:          pgTextPtr(service.PlaceEn),
			PublicationState: string(service.PublicationState),
			SortOrder:        int32(service.SortOrder),
		})
		if err != nil {
			return wrap(err, "insert home service", "", "")
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapHomeService(row)
		return nil
	})
	if err != nil {
		return Service{}, err
	}
	return result, nil
}

// UpdateHomeService reemplaza los campos mutables de un servicio (FR-004). Un
// id inexistente → apperr.NotFound; la auditoría va en la misma transacción.
func (r *repository) UpdateHomeService(ctx context.Context, service Service, actions ...audit.Action) (Service, error) {
	var result Service
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateHomeService(ctx, gendb.UpdateHomeServiceParams{
			DayOfWeek:        int16(service.DayOfWeek),
			StartTime:        service.StartTime,
			EndTime:          pgTextPtr(service.EndTime),
			NameEs:           service.NameEs,
			NameEn:           pgTextPtr(service.NameEn),
			DescriptionEs:    pgTextPtr(service.DescriptionEs),
			DescriptionEn:    pgTextPtr(service.DescriptionEn),
			PlaceEs:          service.PlaceEs,
			PlaceEn:          pgTextPtr(service.PlaceEn),
			PublicationState: string(service.PublicationState),
			SortOrder:        int32(service.SortOrder),
			ID:               pgUUID(service.ID),
		})
		if err != nil {
			return wrap(err, "update home service", "El servicio no existe", "")
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapHomeService(row)
		return nil
	})
	if err != nil {
		return Service{}, err
	}
	return result, nil
}

// DeleteHomeService elimina un servicio (borrado físico) y registra su
// auditoría en la misma transacción. Devuelve false si el id no existía.
func (r *repository) DeleteHomeService(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteHomeService(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete home service", "", "")
		}
		if rows == 0 {
			return nil
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}

// GetHomeServiceByID devuelve un servicio por id; false si no existe.
func (r *repository) GetHomeServiceByID(ctx context.Context, id uuid.UUID) (Service, bool, error) {
	row, err := r.q.GetHomeServiceByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Service{}, false, nil
		}
		return Service{}, false, wrap(err, "get home service", "", "")
	}
	return mapHomeService(row), true, nil
}

// ListHomeServices devuelve todos los servicios (panel, borradores incluidos)
// en el orden del contrato (`sort_order, id`).
func (r *repository) ListHomeServices(ctx context.Context) ([]Service, error) {
	rows, err := r.q.ListHomeServices(ctx)
	if err != nil {
		return nil, wrap(err, "list home services", "", "")
	}
	services := make([]Service, 0, len(rows))
	for _, row := range rows {
		services = append(services, mapHomeService(row))
	}
	return services, nil
}

// ListHomeServicesPublished devuelve solo los servicios publicados (FR-013).
func (r *repository) ListHomeServicesPublished(ctx context.Context) ([]Service, error) {
	rows, err := r.q.ListHomeServicesPublished(ctx)
	if err != nil {
		return nil, wrap(err, "list home services published", "", "")
	}
	services := make([]Service, 0, len(rows))
	for _, row := range rows {
		services = append(services, mapHomeService(row))
	}
	return services, nil
}

// --- Canales de WhatsApp ---

// InsertHomeWhatsappChannel da de alta un canal (FR-005). El duplicado exacto
// (kind + destino + nombre) → apperr.Conflict (UNIQUE de la BD); la auditoría
// va en la misma transacción.
func (r *repository) InsertHomeWhatsappChannel(ctx context.Context, channel WhatsappChannel, actions ...audit.Action) (WhatsappChannel, error) {
	var result WhatsappChannel
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertHomeWhatsappChannel(ctx, gendb.InsertHomeWhatsappChannelParams{
			NameEs:           channel.NameEs,
			NameEn:           pgTextPtr(channel.NameEn),
			Kind:             channel.Kind,
			Destination:      channel.Destination,
			PublicationState: string(channel.PublicationState),
			SortOrder:        int32(channel.SortOrder),
		})
		if err != nil {
			return wrap(err, "insert home whatsapp channel", "", conflictWhatsappDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapHomeWhatsappChannel(row)
		return nil
	})
	if err != nil {
		return WhatsappChannel{}, err
	}
	return result, nil
}

// UpdateHomeWhatsappChannel reemplaza los campos mutables de un canal (FR-005).
// Duplicado exacto → apperr.Conflict; id inexistente → apperr.NotFound.
func (r *repository) UpdateHomeWhatsappChannel(ctx context.Context, channel WhatsappChannel, actions ...audit.Action) (WhatsappChannel, error) {
	var result WhatsappChannel
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateHomeWhatsappChannel(ctx, gendb.UpdateHomeWhatsappChannelParams{
			NameEs:           channel.NameEs,
			NameEn:           pgTextPtr(channel.NameEn),
			Kind:             channel.Kind,
			Destination:      channel.Destination,
			PublicationState: string(channel.PublicationState),
			SortOrder:        int32(channel.SortOrder),
			ID:               pgUUID(channel.ID),
		})
		if err != nil {
			return wrap(err, "update home whatsapp channel", "El canal no existe", conflictWhatsappDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapHomeWhatsappChannel(row)
		return nil
	})
	if err != nil {
		return WhatsappChannel{}, err
	}
	return result, nil
}

// DeleteHomeWhatsappChannel elimina un canal y registra su auditoría en la
// misma transacción. Devuelve false si el id no existía.
func (r *repository) DeleteHomeWhatsappChannel(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteHomeWhatsappChannel(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete home whatsapp channel", "", "")
		}
		if rows == 0 {
			return nil
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}

// GetHomeWhatsappChannelByID devuelve un canal por id; false si no existe.
func (r *repository) GetHomeWhatsappChannelByID(ctx context.Context, id uuid.UUID) (WhatsappChannel, bool, error) {
	row, err := r.q.GetHomeWhatsappChannelByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WhatsappChannel{}, false, nil
		}
		return WhatsappChannel{}, false, wrap(err, "get home whatsapp channel", "", "")
	}
	return mapHomeWhatsappChannel(row), true, nil
}

// ListHomeWhatsappChannels devuelve todos los canales (panel).
func (r *repository) ListHomeWhatsappChannels(ctx context.Context) ([]WhatsappChannel, error) {
	rows, err := r.q.ListHomeWhatsappChannels(ctx)
	if err != nil {
		return nil, wrap(err, "list home whatsapp channels", "", "")
	}
	channels := make([]WhatsappChannel, 0, len(rows))
	for _, row := range rows {
		channels = append(channels, mapHomeWhatsappChannel(row))
	}
	return channels, nil
}

// ListHomeWhatsappChannelsPublished devuelve solo los canales publicados.
func (r *repository) ListHomeWhatsappChannelsPublished(ctx context.Context) ([]WhatsappChannel, error) {
	rows, err := r.q.ListHomeWhatsappChannelsPublished(ctx)
	if err != nil {
		return nil, wrap(err, "list home whatsapp channels published", "", "")
	}
	channels := make([]WhatsappChannel, 0, len(rows))
	for _, row := range rows {
		channels = append(channels, mapHomeWhatsappChannel(row))
	}
	return channels, nil
}

// --- Redes sociales ---

// InsertHomeSocialLink da de alta el enlace de una red (FR-006). Un segundo
// enlace para la misma red → apperr.Conflict (UNIQUE (network)); la auditoría va
// en la misma transacción.
func (r *repository) InsertHomeSocialLink(ctx context.Context, link SocialLink, actions ...audit.Action) (SocialLink, error) {
	var result SocialLink
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertHomeSocialLink(ctx, gendb.InsertHomeSocialLinkParams{
			Network:          link.Network,
			Url:              link.URL,
			PublicationState: string(link.PublicationState),
		})
		if err != nil {
			return wrap(err, "insert home social link", "", conflictSocialDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapHomeSocialLink(row)
		return nil
	})
	if err != nil {
		return SocialLink{}, err
	}
	return result, nil
}

// UpdateHomeSocialLink reemplaza los campos mutables de un enlace (FR-006).
// Duplicado → apperr.Conflict; id inexistente → apperr.NotFound.
func (r *repository) UpdateHomeSocialLink(ctx context.Context, link SocialLink, actions ...audit.Action) (SocialLink, error) {
	var result SocialLink
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateHomeSocialLink(ctx, gendb.UpdateHomeSocialLinkParams{
			Network:          link.Network,
			Url:              link.URL,
			PublicationState: string(link.PublicationState),
			ID:               pgUUID(link.ID),
		})
		if err != nil {
			return wrap(err, "update home social link", "El enlace no existe", conflictSocialDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapHomeSocialLink(row)
		return nil
	})
	if err != nil {
		return SocialLink{}, err
	}
	return result, nil
}

// DeleteHomeSocialLink elimina un enlace y registra su auditoría en la misma
// transacción. Devuelve false si el id no existía.
func (r *repository) DeleteHomeSocialLink(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteHomeSocialLink(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete home social link", "", "")
		}
		if rows == 0 {
			return nil
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}

// GetHomeSocialLinkByID devuelve un enlace por id; false si no existe.
func (r *repository) GetHomeSocialLinkByID(ctx context.Context, id uuid.UUID) (SocialLink, bool, error) {
	row, err := r.q.GetHomeSocialLinkByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SocialLink{}, false, nil
		}
		return SocialLink{}, false, wrap(err, "get home social link", "", "")
	}
	return mapHomeSocialLink(row), true, nil
}

// ListHomeSocialLinks devuelve todos los enlaces (panel) ordenados por red.
func (r *repository) ListHomeSocialLinks(ctx context.Context) ([]SocialLink, error) {
	rows, err := r.q.ListHomeSocialLinks(ctx)
	if err != nil {
		return nil, wrap(err, "list home social links", "", "")
	}
	links := make([]SocialLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, mapHomeSocialLink(row))
	}
	return links, nil
}

// ListHomeSocialLinksPublished devuelve solo los enlaces publicados.
func (r *repository) ListHomeSocialLinksPublished(ctx context.Context) ([]SocialLink, error) {
	rows, err := r.q.ListHomeSocialLinksPublished(ctx)
	if err != nil {
		return nil, wrap(err, "list home social links published", "", "")
	}
	links := make([]SocialLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, mapHomeSocialLink(row))
	}
	return links, nil
}

// Mensajes de conflicto por restricción UNIQUE (los usa el service para el 409).
const (
	conflictWhatsappDuplicate = "Ya existe un canal de WhatsApp con ese tipo, destino y nombre"
	conflictSocialDuplicate   = "Ya existe un enlace para esa red social"
)
