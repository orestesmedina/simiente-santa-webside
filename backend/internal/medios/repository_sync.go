package medios

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
)

// repository_sync.go es el acceso a datos del registro de búsquedas
// (`media_sync_runs`, FR-018/R9-9): una fila por «Buscar episodios nuevos»
// completada, con su resumen numérico y el estado de cada feed.
//
// El registro y su fila de auditoría (`media.sync`, una por búsqueda completa)
// van en la MISMA transacción (P9-14/FR-022): la búsqueda no puede quedar
// registrada a medias.

// InsertMediaSyncRun persiste el resumen de una búsqueda terminada y registra su
// fila de auditoría en la misma transacción (FR-018/FR-022). Los conteos son
// parciales cuando el presupuesto de tiempo se agotó o un feed falló: los fallos
// se informan, no se esconden (FR-016).
func (r *repository) InsertMediaSyncRun(ctx context.Context, run SyncRun, actions ...audit.Action) (SyncRun, error) {
	var result SyncRun
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertMediaSyncRun(ctx, insertMediaSyncRunParams(run))
		if err != nil {
			return wrap(err, "insert media sync run", "", "")
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapSyncRun(row)
		return nil
	})
	if err != nil {
		return SyncRun{}, err
	}
	return result, nil
}

// GetLastMediaSyncRun devuelve la búsqueda más reciente por `started_at` (con
// desempate por `id`) para el bloque «última búsqueda» del panel (FR-018). El
// segundo valor es false cuando todavía no se ha buscado nunca: no es un error,
// el panel muestra «nunca se ha buscado».
func (r *repository) GetLastMediaSyncRun(ctx context.Context) (SyncRun, bool, error) {
	row, err := r.q.GetLastMediaSyncRun(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SyncRun{}, false, nil
		}
		return SyncRun{}, false, wrap(err, "get last media sync run", "", "")
	}
	return mapSyncRun(row), true, nil
}

// insertMediaSyncRunParams traduce el resumen del dominio al alta de sqlc.
func insertMediaSyncRunParams(run SyncRun) gendb.InsertMediaSyncRunParams {
	return gendb.InsertMediaSyncRunParams{
		ActorUserID:                   pgUUIDPtr(run.ActorUserID),
		StartedAt:                     pgTime(run.IniciadoEn),
		FinishedAt:                    pgTime(run.TerminadoEn),
		ImportedCount:                 int32(run.Importados),
		JoinedCount:                   int32(run.Unidos),
		MarkedCount:                   int32(run.Marcados),
		HiddenCount:                   int32(run.Ocultos),
		AvailabilityCheckedCount:      int32(run.Comprobados),
		AvailabilityInconclusiveCount: int32(run.Inconclusos),
		YoutubeStatus:                 string(run.EstadoYouTube),
		PodcastStatus:                 string(run.EstadoPodcast),
	}
}
