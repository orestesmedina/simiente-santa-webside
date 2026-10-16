//go:build integration

package donaciones

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
)

// Pruebas de integración de los textos de la sección contra PostgreSQL real
// (F7, T710): el upsert del singleton por clave (crea y luego reemplaza, sin
// duplicar por `UNIQUE (text_key)`), el registro cerrado de claves y las
// consultas del panel. Requieren DATABASE_URL_TEST (si no existe, se omiten
// solas).

// seedTexto guarda un texto por el repositorio y devuelve la fila; si falla, la
// prueba se detiene.
func seedTexto(t *testing.T, repo *repository, text TextoDonacion) TextoDonacion {
	t.Helper()
	saved, err := repo.UpsertDonationText(context.Background(), text)
	if err != nil {
		t.Fatalf("guardar texto %q: %v", text.TextKey, err)
	}
	return saved
}

// countTextos cuenta las filas de `donation_texts` con SQL directo.
func countTextos(t *testing.T, repo *repository) int {
	t.Helper()
	var total int
	if err := repo.pool.QueryRow(context.Background(), `SELECT count(*) FROM donation_texts`).Scan(&total); err != nil {
		t.Fatalf("contar filas de donation_texts: %v", err)
	}
	return total
}

// TestIntegrationUpsertDonationTextCreatesThenUpdates fija el patrón singleton
// (P7-17): la primera llamada crea la fila de la clave y la segunda la
// **reemplaza** (mismo id, sin duplicar) con `updated_at` renovado.
func TestIntegrationUpsertDonationTextCreatesThenUpdates(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	created := seedTexto(t, repo, TextoDonacion{
		TextKey:          TextKeyIntro,
		BodyEs:           "Puedes dar tu ofrenda en el culto.",
		BodyEn:           textOf("You can give your offering at the service."),
		PublicationState: StatePublished,
	})
	if created.ID == uuid.Nil {
		t.Fatal("el alta del texto no devolvió id")
	}
	if got := countTextos(t, repo); got != 1 {
		t.Fatalf("filas tras el alta = %d, se esperaba 1", got)
	}

	replaced := seedTexto(t, repo, TextoDonacion{
		TextKey:          TextKeyIntro,
		BodyEs:           "Puedes dar tu ofrenda o diezmo.",
		PublicationState: StateDraft,
	})
	if replaced.ID != created.ID {
		t.Errorf("el upsert duplicó la clave: id = %v, se esperaba %v", replaced.ID, created.ID)
	}
	if got := countTextos(t, repo); got != 1 {
		t.Errorf("filas tras el reemplazo = %d, se esperaba 1", got)
	}
	if replaced.BodyEs != "Puedes dar tu ofrenda o diezmo." || replaced.BodyEn != nil {
		t.Errorf("el reemplazo no escribió los campos: %+v", replaced)
	}
	if replaced.PublicationState != StateDraft {
		t.Errorf("estado = %q, se esperaba draft", replaced.PublicationState)
	}
	if replaced.UpdatedAt.Before(created.UpdatedAt) {
		t.Errorf("updated_at no se renovó: %v < %v", replaced.UpdatedAt, created.UpdatedAt)
	}

	// El detalle del panel devuelve lo guardado; una clave sin guardar no existe.
	got, found, err := repo.GetDonationTextByKey(ctx, TextKeyIntro)
	if err != nil || !found {
		t.Fatalf("GetDonationTextByKey(intro) = (found=%v, err=%v)", found, err)
	}
	if got.BodyEs != replaced.BodyEs {
		t.Errorf("detalle = %+v", got)
	}
	if _, found, err := repo.GetDonationTextByKey(ctx, TextKeyUsage); err != nil || found {
		t.Errorf("GetDonationTextByKey(usage) = (found=%v, err=%v), se esperaba (false, nil)", found, err)
	}
}

// TestIntegrationUpsertDonationTextRejectsUnknownKey fija el registro cerrado de
// claves (`text_key IN ('intro','usage')`): una clave fuera del catálogo falla.
func TestIntegrationUpsertDonationTextRejectsUnknownKey(t *testing.T) {
	repo := newDonationRepo(t)

	_, err := repo.UpsertDonationText(context.Background(), TextoDonacion{
		TextKey:          TextKey("otro"),
		BodyEs:           "Texto cualquiera",
		PublicationState: StatePublished,
	})
	if err == nil {
		t.Fatal("la BD aceptó una clave fuera del catálogo")
	}
	if kind := errorKind(err); kind != apperr.KindInvalid {
		t.Fatalf("kind = %v (%v), se esperaba invalid", kind, err)
	}
	if details := errorDetails(err); details == nil || details["textKey"] == nil {
		t.Fatalf("details = %v, se esperaba textKey", details)
	}
	if got := countTextos(t, repo); got != 0 {
		t.Errorf("se guardó algo a medias: %d filas", got)
	}
}

// TestIntegrationListDonationTextsOrders fija el listado del panel: ambos textos
// en orden de clave, con su estado propio (Q10).
func TestIntegrationListDonationTextsOrders(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	usage := seedTexto(t, repo, TextoDonacion{
		TextKey:          TextKeyUsage,
		BodyEs:           "Las donaciones sostienen la obra.",
		PublicationState: StateDraft,
	})
	intro := seedTexto(t, repo, TextoDonacion{
		TextKey:          TextKeyIntro,
		BodyEs:           "Puedes dar tu ofrenda.",
		BodyEn:           textOf("You can give your offering."),
		PublicationState: StatePublished,
	})

	items, err := repo.ListDonationTexts(ctx)
	if err != nil {
		t.Fatalf("ListDonationTexts: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("el panel devolvió %d textos, se esperaban 2", len(items))
	}
	if items[0].ID != intro.ID || items[1].ID != usage.ID {
		t.Errorf("orden = [%v, %v], se esperaba [intro, usage]", items[0].TextKey, items[1].TextKey)
	}
	if items[1].PublicationState != StateDraft {
		t.Errorf("el panel no ve el estado del borrador: %+v", items[1])
	}
}
