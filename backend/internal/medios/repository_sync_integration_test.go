//go:build integration

package medios

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Pruebas de integración del registro de búsquedas contra PostgreSQL real
// (T913): el resumen de FR-018 se persiste completo, «la última búsqueda» del
// panel se resuelve por `started_at DESC, id` y los `CHECK` de los conteos se
// disparan en la base.
//
// Requieren DATABASE_URL_TEST con las migraciones aplicadas (lo aporta el CI o
// el compose local; si no existe, se omiten solas).

// resumenBase arma el resumen de una búsqueda terminada: es el punto de partida
// de las pruebas, que cambian un solo aspecto cada vez.
func resumenBase() SyncRun {
	inicio := time.Now().UTC().Truncate(time.Millisecond)
	return SyncRun{
		IniciadoEn:    inicio,
		TerminadoEn:   inicio.Add(12 * time.Second),
		Importados:    3,
		Unidos:        1,
		Marcados:      2,
		Ocultos:       1,
		Comprobados:   4,
		Inconclusos:   1,
		EstadoYouTube: FeedOK,
		EstadoPodcast: FeedFallido,
	}
}

// TestIntegrationInsertMediaSyncRunPersistsTheSummary comprueba que el resumen
// se persiste completo —conteos, fechas y estado de cada feed— y que «la última
// búsqueda» lo devuelve.
//
// El actor viaja en NULL a propósito: crear una cuenta real tocaría la tabla
// `users`, que las pruebas de integración del dominio usuarios truncan en
// paralelo (mismo criterio que F7/T711), y la FK `ON DELETE RESTRICT` haría
// fallar ese TRUNCATE. La traducción del actor está cubierta por las pruebas
// unitarias del mapeo y el actor real lo ejercita el service (T921).
func TestIntegrationInsertMediaSyncRunPersistsTheSummary(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	run := resumenBase()

	guardada, err := repo.InsertMediaSyncRun(ctx, run)
	if err != nil {
		t.Fatalf("InsertMediaSyncRun: %v", err)
	}
	if guardada.ID == uuid.Nil {
		t.Fatal("el registro no devolvió id")
	}
	if guardada.ActorUserID != nil {
		t.Errorf("ActorUserID = %v, se esperaba NULL", guardada.ActorUserID)
	}
	if !guardada.IniciadoEn.Equal(run.IniciadoEn) || !guardada.TerminadoEn.Equal(run.TerminadoEn) {
		t.Errorf("fechas = %v / %v, se esperaban %v / %v",
			guardada.IniciadoEn, guardada.TerminadoEn, run.IniciadoEn, run.TerminadoEn)
	}
	if guardada.Importados != 3 || guardada.Unidos != 1 || guardada.Marcados != 2 || guardada.Ocultos != 1 {
		t.Errorf("conteos = %+v", guardada)
	}
	if guardada.Comprobados != 4 || guardada.Inconclusos != 1 {
		t.Errorf("conteos de disponibilidad = %+v", guardada)
	}
	if guardada.EstadoYouTube != FeedOK || guardada.EstadoPodcast != FeedFallido {
		t.Errorf("estados = %q/%q", guardada.EstadoYouTube, guardada.EstadoPodcast)
	}

	ultima, found, err := repo.GetLastMediaSyncRun(ctx)
	if err != nil || !found {
		t.Fatalf("GetLastMediaSyncRun = (%v, %v)", found, err)
	}
	if ultima.ID != guardada.ID {
		t.Errorf("la última búsqueda = %v, se esperaba %v", ultima.ID, guardada.ID)
	}
}

// TestIntegrationGetLastMediaSyncRunReturnsTheMostRecent fija el orden de «la
// última búsqueda» (`started_at DESC, id`, §8.1.7): gana la que empezó más
// tarde, no la que se guardó al final.
func TestIntegrationGetLastMediaSyncRunReturnsTheMostRecent(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	base := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	nueva := resumenBase()
	nueva.IniciadoEn = base.Add(2 * time.Hour)
	nueva.TerminadoEn = nueva.IniciadoEn.Add(30 * time.Second)

	antigua := resumenBase()
	antigua.IniciadoEn = base
	antigua.TerminadoEn = antigua.IniciadoEn.Add(30 * time.Second)

	// Se guarda primero la nueva: si el orden fuese por inserción, ganaría la
	// antigua.
	guardadaNueva, err := repo.InsertMediaSyncRun(ctx, nueva)
	if err != nil {
		t.Fatalf("InsertMediaSyncRun (nueva): %v", err)
	}
	if _, err := repo.InsertMediaSyncRun(ctx, antigua); err != nil {
		t.Fatalf("InsertMediaSyncRun (antigua): %v", err)
	}

	ultima, found, err := repo.GetLastMediaSyncRun(ctx)
	if err != nil || !found {
		t.Fatalf("GetLastMediaSyncRun = (%v, %v)", found, err)
	}
	if ultima.ID != guardadaNueva.ID {
		t.Errorf("la última búsqueda = %v, se esperaba la más reciente %v", ultima.ID, guardadaNueva.ID)
	}
}

// TestIntegrationGetLastMediaSyncRunWithoutAnyRun fija el estado «nunca se ha
// buscado» (FR-018): no es un error, el panel muestra el bloque sin resumen.
func TestIntegrationGetLastMediaSyncRunWithoutAnyRun(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	if _, found, err := repo.GetLastMediaSyncRun(ctx); err != nil || found {
		t.Errorf("GetLastMediaSyncRun sin búsquedas = (%v, %v), se esperaba (false, nil)", found, err)
	}
}

// TestIntegrationInsertMediaSyncRunRejectsInvalidData comprueba que la BD es la
// última línea de defensa del registro: un conteo negativo y un fin anterior al
// inicio se traducen a `invalid` con el campo en el detalle.
func TestIntegrationInsertMediaSyncRunRejectsInvalidData(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	negativo := resumenBase()
	negativo.Importados = -1
	if _, err := repo.InsertMediaSyncRun(ctx, negativo); err == nil {
		t.Error("la BD aceptó un conteo negativo")
	} else if details := errorDetails(err); details == nil || details["importedCount"] == nil {
		t.Errorf("el detalle por campo se perdió: %v (%v)", details, err)
	}

	invertida := resumenBase()
	invertida.TerminadoEn = invertida.IniciadoEn.Add(-time.Minute)
	if _, err := repo.InsertMediaSyncRun(ctx, invertida); err == nil {
		t.Error("la BD aceptó una búsqueda que termina antes de empezar")
	} else if details := errorDetails(err); details == nil || details["finishedAt"] == nil {
		t.Errorf("el detalle por campo se perdió: %v (%v)", details, err)
	}
}
