//go:build integration

package donaciones

import (
	"context"
	"testing"
)

// Pruebas de integración de la lectura pública contra PostgreSQL real (F7,
// T710, FR-012/SC-002): las consultas `ListPublished*` **nunca** devuelven
// borradores —ni métodos ni textos—, el orden es el de la sección y el `LIMIT`
// (5 en el bloque, 200 en la sección) se respeta. Requieren DATABASE_URL_TEST
// (si no existe, se omiten solas).

// TestIntegrationListPublishedMethodsExcludesDraftsAndOrders fija la única vía
// pública: solo publicados, en el orden `CASE kind` (iban primero), `sort_order`,
// `id`.
func TestIntegrationListPublishedMethodsExcludesDraftsAndOrders(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	ibanFive := baseIban()
	ibanFive.Number = "CR05015202001234567895"
	ibanFive.SortOrder = 5
	savedIbanFive := seedMetodo(t, repo, ibanFive)

	sinpeZero := baseSinpe()
	sinpeZero.SortOrder = 0
	savedSinpeZero := seedMetodo(t, repo, sinpeZero)

	ibanOne := baseIban()
	ibanOne.SortOrder = 1
	savedIbanOne := seedMetodo(t, repo, ibanOne)

	draft := baseIban()
	draft.Number = "CR05015202001234567899"
	draft.PublicationState = StateDraft
	seedMetodo(t, repo, draft)

	items, err := repo.ListPublishedDonationMethods(ctx, 200)
	if err != nil {
		t.Fatalf("ListPublishedDonationMethods: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("la lectura pública devolvió %d métodos, se esperaban 3 (sin el borrador)", len(items))
	}
	wantOrder := []string{savedIbanOne.Number, savedIbanFive.Number, savedSinpeZero.Number}
	for i, want := range wantOrder {
		if items[i].Number != want {
			t.Fatalf("orden público[%d] = %q, se esperaba %q (iban primero, sort_order, id)", i, items[i].Number, want)
		}
	}
	// El DTO/entidad pública no expone borradores: ninguno de los devueltos es el
	// borrador y ninguno declara otro estado que `published`.
	for _, item := range items {
		if item.Number == draft.Number {
			t.Errorf("se filtró un borrador: %+v", item)
		}
		if item.PublicationState != StatePublished {
			t.Errorf("estado público = %q, se esperaba published", item.PublicationState)
		}
	}
}

// TestIntegrationListPublishedMethodsRespectsLimit fija el tope del servidor: el
// `LIMIT` que pasa el service acota el resultado (5 en el bloque, 200 en la
// sección) devolviendo los primeros del orden.
func TestIntegrationListPublishedMethodsRespectsLimit(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	for i := 0; i < 8; i++ {
		m := baseIban()
		m.Number = "CR0501520200123456789" + string(rune('0'+i)) // 22 caracteres
		m.SortOrder = i
		seedMetodo(t, repo, m)
	}

	five, err := repo.ListPublishedDonationMethods(ctx, 5)
	if err != nil {
		t.Fatalf("ListPublishedDonationMethods(5): %v", err)
	}
	if len(five) != 5 {
		t.Fatalf("el bloque de la portada devolvió %d, se esperaban 5", len(five))
	}
	for i, item := range five {
		if item.SortOrder != i {
			t.Errorf("bloque[%d].sortOrder = %d, se esperaba %d", i, item.SortOrder, i)
		}
	}

	all, err := repo.ListPublishedDonationMethods(ctx, 200)
	if err != nil {
		t.Fatalf("ListPublishedDonationMethods(200): %v", err)
	}
	if len(all) != 8 {
		t.Errorf("la sección devolvió %d, se esperaban 8", len(all))
	}
}

// TestIntegrationListPublishedTextsExcludesDrafts fija que la lectura pública de
// textos solo devuelve publicados (Q10/P7-12).
func TestIntegrationListPublishedTextsExcludesDrafts(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	intro := seedTexto(t, repo, TextoDonacion{
		TextKey:          TextKeyIntro,
		BodyEs:           "Puedes dar tu ofrenda.",
		BodyEn:           textOf("You can give your offering."),
		PublicationState: StatePublished,
	})
	seedTexto(t, repo, TextoDonacion{
		TextKey:          TextKeyUsage,
		BodyEs:           "Las donaciones sostienen la obra.",
		PublicationState: StateDraft,
	})

	items, err := repo.ListPublishedDonationTexts(ctx)
	if err != nil {
		t.Fatalf("ListPublishedDonationTexts: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("la lectura pública devolvió %d textos, se esperaba 1 (sin el borrador)", len(items))
	}
	if items[0].TextKey != TextKeyIntro || items[0].ID != intro.ID {
		t.Errorf("texto público = %+v, se esperaba el introductorio", items[0])
	}
	if items[0].PublicationState != StatePublished {
		t.Errorf("estado público = %q, se esperaba published", items[0].PublicationState)
	}
}
