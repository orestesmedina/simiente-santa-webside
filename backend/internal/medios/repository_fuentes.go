package medios

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	gendb "simiente-santa/backend/internal/db"
)

// repository_fuentes.go es el acceso a datos de las fuentes de importación
// (`media_episode_sources`): una fila por elemento de feed importado, con
// `UNIQUE (provider, external_id)` como garantía de idempotencia (R9-11/P9-11).
//
// La importación es una **instantánea** (P9-9/R9-6): la búsqueda nunca refresca
// título, fecha, carátula ni enlace de lo ya importado, así que aquí no hay
// ningún `UPDATE` —el único que existe es el repunte de la unión, en
// `repository_episodios.go`— y ninguna fuente se borra jamás (FR-019: conserva
// la procedencia).
//
// Ninguna de estas consultas escribe auditoría: el alta importada por la
// búsqueda se registra **solo** con `media.sync` (P9-14).

// errFuenteYaImportada es la señal INTERNA de la transacción del alta de un
// elemento (H3): el `INSERT … ON CONFLICT DO NOTHING` de la fuente no devolvió
// fila porque otra búsqueda ya había importado ese `(provider, external_id)`. No
// es un error de negocio —la idempotencia es la esperada (R9-11)— pero sí obliga
// a **revertir** la transacción para que la ficha recién creada no quede sin
// enlaces (nunca sale de este paquete: el repository lo traduce a `ok=false`).
var errFuenteYaImportada = errors.New("medios: la fuente ya estaba importada")

// errCandidatoConPar es la señal INTERNA de la misma transacción cuando el
// candidato de un par dudoso **ya tiene pareja** (otra búsqueda simultánea se le
// adelantó): la marca del par escribe con `duplicate_of_id IS NULL`, así que sin
// fila devuelta la transacción se revierte y la pareja recíproca no se rompe
// (M1). Tampoco sale de este paquete.
var errCandidatoConPar = errors.New("medios: el candidato del par ya tiene pareja")

// InsertMediaEpisodeSource importa un elemento de feed con
// `ON CONFLICT (provider, external_id) DO NOTHING` (FR-012/R9-11): si el
// elemento ya estaba importado la fila **no se duplica ni se actualiza** —la
// instantánea se conserva— y el segundo valor es false (no es un error). Con dos
// búsquedas simultáneas la base resuelve la carrera: exactamente una inserta.
func (r *repository) InsertMediaEpisodeSource(ctx context.Context, fuente FuenteEpisodio) (FuenteEpisodio, bool, error) {
	row, err := r.q.InsertMediaEpisodeSource(ctx, insertFuenteParams(fuente))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// `DO NOTHING` sin fila devuelta: el elemento ya estaba importado.
			return FuenteEpisodio{}, false, nil
		}
		return FuenteEpisodio{}, false, wrap(err, "insert media episode source", "", messageFuenteDuplicada)
	}
	return mapFuenteEpisodio(row), true, nil
}

// insertFuenteParams traduce una fuente del dominio al alta de sqlc. Lo usan el
// alta suelta (`InsertMediaEpisodeSource`) y la transacción del alta de un
// elemento (ficha + fuente, `repository_episodios.go`).
func insertFuenteParams(fuente FuenteEpisodio) gendb.InsertMediaEpisodeSourceParams {
	return gendb.InsertMediaEpisodeSourceParams{
		ItemID:     pgUUID(fuente.EpisodioID),
		Provider:   string(fuente.Proveedor),
		ExternalID: fuente.IDExterno,
		FeedTitle:  fuente.TituloFeed,
		Url:        fuente.Enlace,
	}
}

// ListMediaEpisodeSourcesByProvider devuelve las fuentes importadas de un
// proveedor en orden de importación (FIFO): es la lista contra la que la
// búsqueda detecta las que **ya no están en la ventana del feed** (FR-017). La
// cola de comprobación (solo fichas visibles no absorbidas y tope 30) la aplica
// el service.
func (r *repository) ListMediaEpisodeSourcesByProvider(ctx context.Context, proveedor Proveedor) ([]FuenteEpisodio, error) {
	rows, err := r.q.ListMediaEpisodeSourcesByProvider(ctx, string(proveedor))
	if err != nil {
		return nil, wrap(err, "list media episode sources by provider", "", "")
	}
	fuentes := make([]FuenteEpisodio, 0, len(rows))
	for _, row := range rows {
		fuentes = append(fuentes, mapFuenteEpisodio(row))
	}
	return fuentes, nil
}

// GetMediaEpisodeSource responde si un elemento ya está importado (salto de la
// importación, FR-012) y devuelve su procedencia para las comprobaciones de
// disponibilidad y el repunte de la unión. El segundo valor es false si no
// existe.
func (r *repository) GetMediaEpisodeSource(ctx context.Context, proveedor Proveedor, idExterno string) (FuenteEpisodio, bool, error) {
	row, err := r.q.GetMediaEpisodeSource(ctx, gendb.GetMediaEpisodeSourceParams{
		Provider:   string(proveedor),
		ExternalID: idExterno,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FuenteEpisodio{}, false, nil
		}
		return FuenteEpisodio{}, false, wrap(err, "get media episode source", "", "")
	}
	return mapFuenteEpisodio(row), true, nil
}

// ListMediaItemSourcesByItem devuelve las fuentes de una ficha en orden de
// importación: la procedencia que el panel muestra (`FuentesList`), el detalle
// que el equipo revisa antes de unir y la comprobación del resultado de la
// unión. Una ficha tiene 1 o 2 fuentes (una por proveedor, invariante del
// service).
func (r *repository) ListMediaItemSourcesByItem(ctx context.Context, itemID uuid.UUID) ([]FuenteEpisodio, error) {
	rows, err := r.q.ListMediaItemSourcesByItem(ctx, pgUUID(itemID))
	if err != nil {
		return nil, wrap(err, "list media item sources by item", "", "")
	}
	fuentes := make([]FuenteEpisodio, 0, len(rows))
	for _, row := range rows {
		fuentes = append(fuentes, mapFuenteEpisodio(row))
	}
	return fuentes, nil
}
