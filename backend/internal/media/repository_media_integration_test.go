//go:build integration

package media

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración del dominio media contra PostgreSQL real (T415): la
// consulta compartida `IsFilePublished` sirve SOLO archivos referenciados por
// contenido publicado y visible hoy —la identidad de la portada (F3, regresión)
// y los eventos y actividades de agenda (F4, R4-2/R4-6)—.
//
// Aislamiento: cada prueba trabaja dentro de una TRANSACCIÓN que se revierte al
// terminar, en vez de truncar tablas compartidas (`home_identity` la truncan las
// pruebas de la portada; las `agenda_*` las de agenda). Así las pruebas de este
// paquete no interfieren con las de los otros dominios, que corren en paralelo.

// testToday/testNowTime son el corte FIJO de la regla R4-2 (nunca el reloj real).
const (
	testToday   = "2026-10-11"
	testNowTime = "12:00"
	testPast    = "2026-10-10"
	testFuture  = "2026-10-12"
)

// newMediaRepo abre una transacción de prueba y devuelve el repositorio ligado a
// ella; al terminar la prueba la transacción se revierte (nada queda escrito).
func newMediaRepo(t *testing.T) (*repository, *pgx.Tx, context.Context) {
	t.Helper()

	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir transacción de prueba: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	return &repository{q: gendb.New(tx)}, &tx, ctx
}

// seedHomeIdentity deja la identidad de la portada con el archivo indicado en el
// estado pedido, dentro de la transacción de la prueba.
func seedHomeIdentity(t *testing.T, tx *pgx.Tx, file, state string) {
	t.Helper()
	_, err := (*tx).Exec(context.Background(), `INSERT INTO home_identity
		(singleton, name_es, logo_file, logo_alt_es, publication_state)
		VALUES (TRUE, 'Iglesia Simiente', $1, 'Logotipo', $2)
		ON CONFLICT (singleton) DO UPDATE SET
			logo_file = EXCLUDED.logo_file,
			logo_alt_es = EXCLUDED.logo_alt_es,
			publication_state = EXCLUDED.publication_state`,
		pgtype.Text{String: file, Valid: true}, state)
	if err != nil {
		t.Fatalf("sembrar home_identity: %v", err)
	}
}

// seedEvent crea un evento con una fecha, dentro de la transacción de la prueba.
func seedEvent(t *testing.T, tx *pgx.Tx, file, state, day string, endTime *string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	if err := (*tx).QueryRow(ctx, `INSERT INTO agenda_events
		(title_es, detail_es, place_es, image_file, image_alt_es, publication_state)
		VALUES ('Evento de prueba', 'Detalle', 'Templo', $1, 'Alt', $2)
		RETURNING id`, pgtype.Text{String: file, Valid: true}, state).Scan(&id); err != nil {
		t.Fatalf("sembrar evento: %v", err)
	}
	if _, err := (*tx).Exec(ctx, `INSERT INTO agenda_event_dates (event_id, event_date, end_time)
		VALUES ($1, $2, $3)`, id, day, pgTextOrNull(endTime)); err != nil {
		t.Fatalf("sembrar fecha del evento: %v", err)
	}
	return id
}

// seedActivity crea una actividad con su fecha única, dentro de la transacción.
// El título lleva el nombre del archivo para que dos casos del mismo día no
// choquen con `UNIQUE (title_es, activity_date, place_es)`.
func seedActivity(t *testing.T, tx *pgx.Tx, file, state, day string, startTime, endTime *string) string {
	t.Helper()
	var id string
	if err := (*tx).QueryRow(context.Background(), `INSERT INTO agenda_activities
		(title_es, detail_es, place_es, image_file, image_alt_es, activity_date, start_time, end_time, publication_state)
		VALUES ('Actividad de prueba ' || $1, 'Detalle', 'Salón', $1, 'Alt', $2, $3, $4, $5)
		RETURNING id`,
		file, day, pgTextOrNull(startTime), pgTextOrNull(endTime), state).Scan(&id); err != nil {
		t.Fatalf("sembrar actividad: %v", err)
	}
	return id
}

// pgTextOrNull traduce un *string a pgtype.Text.
func pgTextOrNull(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

// published pregunta al repositorio por el archivo con el corte fijo.
func published(t *testing.T, repo *repository, ctx context.Context, file string) bool {
	t.Helper()
	ok, err := repo.IsFilePublished(ctx, file, testToday, testNowTime)
	if err != nil {
		t.Fatalf("IsFilePublished(%q): %v", file, err)
	}
	return ok
}

// TestIntegrationIsFilePublishedFollowsHomeState es la regresión de F3: el logo
// o la imagen de portada de una identidad publicada se sigue sirviendo y en
// borrador no (mismos casos que antes de mover la descarga a media).
func TestIntegrationIsFilePublishedFollowsHomeState(t *testing.T) {
	repo, tx, ctx := newMediaRepo(t)
	const file = "img_0123456789abcdef0123456789abcdef0123.png"

	seedHomeIdentity(t, tx, file, "draft")
	if published(t, repo, ctx, file) {
		t.Error("un archivo de portada en borrador no debe servirse")
	}

	seedHomeIdentity(t, tx, file, "published")
	if !published(t, repo, ctx, file) {
		t.Error("un archivo de portada publicada debe servirse")
	}

	// Un archivo huérfano (no referenciado) nunca se sirve.
	if published(t, repo, ctx, "img_000000000000000000000000000000000000.jpg") {
		t.Error("un archivo huérfano no debe servirse")
	}

	seedHomeIdentity(t, tx, file, "draft")
	if published(t, repo, ctx, file) {
		t.Error("al retirar la identidad, su archivo deja de servirse")
	}
}

// TestIntegrationIsFilePublishedFollowsEventDates fija la decisión firme de la
// puerta 2 en eventos (R4-6): un evento publicado y próximo sirve su imagen; al
// pasar la fecha deja de servirla **sin borrarse** (la referencia sigue en la
// fila) y al cargarle una fecha próxima vuelve a servirse.
func TestIntegrationIsFilePublishedFollowsEventDates(t *testing.T) {
	repo, tx, ctx := newMediaRepo(t)
	const file = "img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.jpg"

	// Borrador con fecha próxima: no se sirve.
	eventID := seedEvent(t, tx, file, "draft", testFuture, nil)
	if published(t, repo, ctx, file) {
		t.Error("un evento en borrador no debe servir su imagen")
	}

	// Publicado y próximo: se sirve.
	_, err := (*tx).Exec(ctx, `UPDATE agenda_events SET publication_state = 'published' WHERE id = $1`, eventID)
	if err != nil {
		t.Fatalf("publicar el evento: %v", err)
	}
	if !published(t, repo, ctx, file) {
		t.Error("un evento publicado y próximo debe servir su imagen")
	}

	// Solo con una fecha pasada: deja de servirse (el archivo no se borra).
	_, err = (*tx).Exec(ctx, `UPDATE agenda_event_dates SET event_date = $1 WHERE event_id = $2`, testPast, eventID)
	if err != nil {
		t.Fatalf("vencer el evento: %v", err)
	}
	if published(t, repo, ctx, file) {
		t.Error("un evento vencido no debe servir su imagen")
	}
	var storedFile pgtype.Text
	if err := (*tx).QueryRow(ctx, `SELECT image_file FROM agenda_events WHERE id = $1`, eventID).Scan(&storedFile); err != nil {
		t.Fatalf("leer el evento: %v", err)
	}
	if !storedFile.Valid {
		t.Error("el archivo se borró de la fila: solo debía dejar de servirse")
	}

	// Al cargarle una fecha próxima, la misma URL vuelve a responder.
	if _, err := (*tx).Exec(ctx, `INSERT INTO agenda_event_dates (event_id, event_date) VALUES ($1, $2)`, eventID, testFuture); err != nil {
		t.Fatalf("cargar fecha próxima: %v", err)
	}
	if !published(t, repo, ctx, file) {
		t.Error("al volver a tener fecha próxima, la imagen debe servirse otra vez")
	}
}

// TestIntegrationIsFilePublishedFollowsActivityDate cubre el mismo camino en
// actividades (fecha única) y los tres casos de la regla R4-2.
func TestIntegrationIsFilePublishedFollowsActivityDate(t *testing.T) {
	repo, tx, ctx := newMediaRepo(t)
	const file = "img_bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb.png"

	// Publicada con la fecha de hoy y fin ya terminado (10:00 < 12:00): no se sirve.
	started := "09:00"
	ended := "10:00"
	seedActivity(t, tx, file, "published", testToday, &started, &ended)
	if published(t, repo, ctx, file) {
		t.Error("una actividad de hoy ya terminada no debe servir su imagen")
	}

	// Publicada hoy sin hora de fin: se sirve todo el día.
	const openFile = "img_cccccccc-cccc-cccc-cccc-cccccccccccc.png"
	seedActivity(t, tx, openFile, "published", testToday, nil, nil)
	if !published(t, repo, ctx, openFile) {
		t.Error("una actividad de hoy sin hora de fin debe servir su imagen")
	}

	// Publicada con fecha futura: se sirve.
	const futureFile = "img_dddddddd-dddd-dddd-dddd-dddddddddddd.webp"
	seedActivity(t, tx, futureFile, "published", testFuture, nil, nil)
	if !published(t, repo, ctx, futureFile) {
		t.Error("una actividad publicada y próxima debe servir su imagen")
	}

	// En borrador con fecha futura: no se sirve.
	const draftFile = "img_eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee.jpg"
	seedActivity(t, tx, draftFile, "draft", testFuture, nil, nil)
	if published(t, repo, ctx, draftFile) {
		t.Error("una actividad en borrador no debe servir su imagen")
	}
}
