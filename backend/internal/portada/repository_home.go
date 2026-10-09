package portada

import (
	"context"
	"errors"
	"fmt"

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
