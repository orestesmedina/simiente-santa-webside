package donaciones

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Tests de la cara pública del service (T713/T714): la sección `GetOfrendas` con
// los dos textos por clave, el orden de la sección (iban → sinpe), el fallback
// `en → es` por campo, el tope de seguridad, la ausencia de borradores y las
// claves omitidas; y el bloque `GetPortadaOfrendas` con su tope firme de 5. El
// repositorio es el falso en memoria.

// mustJSON serializa un DTO público como viaja en la respuesta (el JSON crudo es
// lo que inspecciona SC-002: "0 apariciones").
func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("serializar el DTO: %v", err)
	}
	return string(data)
}

// --- T713 · sección pública ---

// TestGetOfrendasReturnsOnlyPublished fija FR-012/SC-002: ningún borrador sale
// por la vía pública, ni métodos ni textos.
func TestGetOfrendasReturnsOnlyPublished(t *testing.T) {
	repo := newFakeRepository()
	draft := publishedMethod(KindIban, "CR05015202001234567890")
	draft.PublicationState = StateDraft
	draft.Institution = "Borrador"
	repo.seedMethods(draft, publishedMethod(KindSinpe, "+50688888888"))
	repo.seedTexts(TextoDonacion{
		TextKey: TextKeyIntro, BodyEs: "Texto en borrador", PublicationState: StateDraft,
	})

	out, err := newTestService(repo).GetOfrendas(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetOfrendas = %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].Kind != "sinpe" {
		t.Fatalf("items = %+v, se esperaba solo el método publicado", out.Items)
	}
	if out.Intro != nil {
		t.Errorf("intro = %+v, un texto en borrador no debe salir", out.Intro)
	}
	raw := mustJSON(t, out)
	if strings.Contains(raw, "Borrador") || strings.Contains(raw, "publicationState") {
		t.Errorf("el DTO público expone un borrador o el estado: %s", raw)
	}
}

// TestGetOfrendasOrdersBySectionOrder fija FR-004/Q1: primero los IBAN y después
// los SINPE; dentro de cada grupo por `sortOrder`, `id`.
func TestGetOfrendasOrdersBySectionOrder(t *testing.T) {
	repo := newFakeRepository()
	ibanA := publishedMethod(KindIban, "CR05015202001234567890")
	ibanA.SortOrder = 2
	ibanA.ID = uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	ibanB := publishedMethod(KindIban, "ES9121000418450200051332")
	ibanB.SortOrder = 1
	ibanB.ID = uuid.MustParse("00000000-0000-0000-0000-00000000000b")
	sinpeFirst := publishedMethod(KindSinpe, "+50688888888")
	sinpeFirst.SortOrder = 0
	// Sembrados desordenados a propósito.
	repo.seedMethods(sinpeFirst, ibanA, ibanB)

	out, err := newTestService(repo).GetOfrendas(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetOfrendas = %v", err)
	}
	got := make([]string, 0, len(out.Items))
	for _, item := range out.Items {
		got = append(got, item.Kind+"|"+item.Number)
	}
	want := []string{
		"iban|" + ibanB.Number,
		"iban|" + ibanA.Number,
		"sinpe|+50688888888",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("orden = %v, se esperaba %v", got, want)
	}
}

// TestGetOfrendasFallbackPerField fija FR-008/SC-006: el titular se resuelve por
// campo (`en` si existe, si no `es`, nunca vacío) y los campos no traducibles
// (`kind`, `number`, `institution`, `currency`) son idénticos en ambos idiomas.
func TestGetOfrendasFallbackPerField(t *testing.T) {
	repo := newFakeRepository()
	withTranslation := publishedMethod(KindIban, "CR05015202001234567890")
	withTranslation.Institution = "Banco Nacional"
	withTranslation.HolderEs = textOf("Ana Pérez")
	withTranslation.HolderEn = textOf("Ana Perez")
	esOnly := publishedMethod(KindIban, "ES9121000418450200051332")
	esOnly.HolderEs = textOf("José Núñez")
	esOnly.SortOrder = 1
	repo.seedMethods(withTranslation, esOnly)

	svc := newTestService(repo)

	es, err := svc.GetOfrendas(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetOfrendas(es) = %v", err)
	}
	en, err := svc.GetOfrendas(context.Background(), LangEN)
	if err != nil {
		t.Fatalf("GetOfrendas(en) = %v", err)
	}

	if es.Items[0].Holder != "Ana Pérez" {
		t.Errorf("holder es = %q", es.Items[0].Holder)
	}
	if en.Items[0].Holder != "Ana Perez" {
		t.Errorf("holder en = %q, se esperaba la traducción", en.Items[0].Holder)
	}
	// El segundo no tiene traducción: cae al español (nunca vacío).
	if en.Items[1].Holder != "José Núñez" {
		t.Errorf("holder en del segundo = %q, se esperaba fallback al español", en.Items[1].Holder)
	}
	// Los campos no traducibles son idénticos.
	for i := range es.Items {
		if es.Items[i].Kind != en.Items[i].Kind || es.Items[i].Number != en.Items[i].Number ||
			es.Items[i].Institution != en.Items[i].Institution || es.Items[i].Currency != en.Items[i].Currency {
			t.Errorf("campo no traducible cambió con el idioma: es=%+v en=%+v", es.Items[i], en.Items[i])
		}
	}
}

// TestGetOfrendasOmitsEmptyFields fija SC-010/US1 esc. 5: sin titular no hay
// clave `holder`, el SINPE no lleva `currency` y el DTO no expone
// `publicationState`; sin nada publicado la respuesta es solo `{lang}`.
func TestGetOfrendasOmitsEmptyFields(t *testing.T) {
	repo := newFakeRepository()
	repo.seedMethods(publishedMethod(KindSinpe, "+50688888888"))

	out, err := newTestService(repo).GetOfrendas(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetOfrendas = %v", err)
	}
	raw := mustJSON(t, out)
	for _, forbidden := range []string{"holder", "currency", "publicationState", "sortOrder"} {
		if strings.Contains(raw, forbidden) {
			t.Errorf("el DTO público lleva %q: %s", forbidden, raw)
		}
	}

	// Sin nada publicado (ni métodos ni textos) la respuesta es solo `{lang}`.
	empty, err := newTestService(newFakeRepository()).GetOfrendas(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetOfrendas(vacío) = %v", err)
	}
	if empty.Items != nil || empty.Intro != nil || empty.Usage != nil {
		t.Errorf("documento vacío = %+v", empty)
	}
	if got := mustJSON(t, empty); got != `{"lang":"es"}` {
		t.Errorf("JSON vacío = %s, se esperaba solo lang", got)
	}
}

// TestGetOfrendasLivesWithOnlyTexts fija P7-13 (pregunta #2): la sección vive si
// hay ≥1 método **o** ≥1 texto publicado; con solo textos publicados salen
// `intro`/`usage`.
func TestGetOfrendasLivesWithOnlyTexts(t *testing.T) {
	repo := newFakeRepository()
	repo.seedTexts(
		TextoDonacion{TextKey: TextKeyIntro, BodyEs: "Cómo donar", PublicationState: StatePublished},
		TextoDonacion{TextKey: TextKeyUsage, BodyEs: "En qué se usan", PublicationState: StatePublished},
	)

	out, err := newTestService(repo).GetOfrendas(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetOfrendas = %v", err)
	}
	if out.Items != nil {
		t.Errorf("items = %+v, se esperaba omitido", out.Items)
	}
	if out.Intro == nil || out.Intro.Text != "Cómo donar" {
		t.Errorf("intro = %+v", out.Intro)
	}
	if out.Usage == nil || out.Usage.Text != "En qué se usan" {
		t.Errorf("usage = %+v", out.Usage)
	}
}

// TestGetOfrendasCapsAt200 fija el tope de seguridad del servidor: con más de
// 200 publicados se sirven los 200 primeros ya ordenados.
func TestGetOfrendasCapsAt200(t *testing.T) {
	repo := newFakeRepository()
	for i := 0; i < maxPublicItems+3; i++ {
		method := publishedMethod(KindIban, "CR05015202001234567890")
		method.SortOrder = i
		repo.seedMethods(method)
	}

	out, err := newTestService(repo).GetOfrendas(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetOfrendas = %v", err)
	}
	if len(out.Items) != maxPublicItems {
		t.Fatalf("items = %d, se esperaban %d", len(out.Items), maxPublicItems)
	}
	if got := repo.publicLimits; len(got) != 1 || got[0] != maxPublicItems {
		t.Errorf("topes consultados = %v, se esperaba [%d]", got, maxPublicItems)
	}
}

// TestGetOfrendasInvalidLang fija el 400 invalid con details.lang.
func TestGetOfrendasInvalidLang(t *testing.T) {
	_, err := newTestService(newFakeRepository()).GetOfrendas(context.Background(), "fr")
	if _, ok := detailsOf(t, err)["lang"]; !ok {
		t.Fatalf("se esperaba details.lang: %v", err)
	}
}

// TestGetOfrendasRepositoryError fija que un fallo de datos no se disfraza.
func TestGetOfrendasRepositoryError(t *testing.T) {
	repo := newFakeRepository()
	repo.readErr = errors.New("boom")
	if _, err := newTestService(repo).GetOfrendas(context.Background(), LangES); err == nil ||
		!strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, se esperaba el error del repositorio", err)
	}
}
