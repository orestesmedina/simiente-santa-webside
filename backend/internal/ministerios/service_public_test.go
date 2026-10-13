package ministerios

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Tests de la cara pública del service (T613): la sección `GetMinisterios` con
// la privacidad del encargado (FR-026/SC-014), el fallback `en → es` por campo
// (FR-008), el orden manual (FR-004), el tope de seguridad y la ausencia de
// borradores (FR-012/SC-002). El repositorio es el falso en memoria y NO
// enmascara la ficha del encargado: si el DTO no la lleva es porque el service
// no la publica.

// newTestService construye el service con el repositorio falso.
func newTestService(repo Repository) *service {
	return NewService(ServiceDeps{Repository: repo})
}

// publishedMinistry construye un ministerio publicado con datos completos; cada
// prueba ajusta lo que necesita.
func publishedMinistry(name string) Ministry {
	return Ministry{
		ID:               uuid.New(),
		NameEs:           name,
		DescriptionEs:    "Descripción de " + name,
		PublicationState: StatePublished,
	}
}

// mustJSON serializa un DTO público como viaja en la respuesta (el JSON crudo es
// lo que inspecciona SC-014: "0 apariciones").
func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("serializar el DTO: %v", err)
	}
	return string(data)
}

// TestGetMinisteriosOrdersByModuleOrder fija FR-004: el orden es el manual del
// módulo (`sortOrder`, `id`) y lo aplica la consulta; el service lo respeta.
func TestGetMinisteriosOrdersByModuleOrder(t *testing.T) {
	repo := newFakeRepository()
	first := publishedMinistry("Primero")
	first.SortOrder = 1
	second := publishedMinistry("Segundo")
	second.SortOrder = 2
	tieA := publishedMinistry("Empate A")
	tieA.ID = uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	tieB := publishedMinistry("Empate B")
	tieB.ID = uuid.MustParse("00000000-0000-0000-0000-00000000000b")
	// Se siembran desordenados a propósito.
	repo.seed(second, tieB, first, tieA)

	out, err := newTestService(repo).GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	got := make([]string, 0, len(out.Items))
	for _, item := range out.Items {
		got = append(got, item.Name)
	}
	want := []string{"Empate A", "Empate B", "Primero", "Segundo"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("orden = %v, se esperaba %v", got, want)
	}
}

// TestGetMinisteriosFallbackPerField fija FR-008/SC-006: cada campo se resuelve
// por separado (`en` si existe, si no `es`) y nunca queda un hueco; el enlace y
// el archivo de imagen NO se traducen.
func TestGetMinisteriosFallbackPerField(t *testing.T) {
	repo := newFakeRepository()
	ministry := publishedMinistry("Alabanza")
	ministry.NameEn = textOf("Worship")
	ministry.DescriptionEn = textOf("Worship team")
	ministry.LeaderNameEs = textOf("Ana Pérez")
	ministry.LeaderVisible = true
	ministry.WhatsappURL = textOf("https://chat.whatsapp.com/abc")
	ministry.ScheduleEs = textOf("Todos los viernes")
	ministry.ImageFile = textOf("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png")
	ministry.ImageAltEs = textOf("Grupo de alabanza")
	repo.seed(ministry)

	svc := newTestService(repo)

	es, err := svc.GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios(es) = %v", err)
	}
	item := es.Items[0]
	if item.Name != "Alabanza" || item.Description != "Descripción de Alabanza" {
		t.Errorf("es = %q / %q", item.Name, item.Description)
	}
	if item.Schedule == nil || *item.Schedule != "Todos los viernes" {
		t.Errorf("schedule = %v", item.Schedule)
	}

	en, err := svc.GetMinisterios(context.Background(), LangEN)
	if err != nil {
		t.Fatalf("GetMinisterios(en) = %v", err)
	}
	itemEn := en.Items[0]
	if itemEn.Name != "Worship" || itemEn.Description != "Worship team" {
		t.Errorf("en = %q / %q (se esperaba la traducción)", itemEn.Name, itemEn.Description)
	}
	// El horario no tiene traducción: cae al español (nunca vacío).
	if itemEn.Schedule == nil || *itemEn.Schedule != "Todos los viernes" {
		t.Errorf("schedule en = %v, se esperaba el fallback al español", itemEn.Schedule)
	}
	// El enlace y el alt de la imagen se resuelven por campo: el enlace no se
	// traduce (es el mismo destino en ambos idiomas).
	if itemEn.WhatsappURL == nil || *itemEn.WhatsappURL != "https://chat.whatsapp.com/abc" {
		t.Errorf("whatsappUrl en = %v", itemEn.WhatsappURL)
	}
	if itemEn.Image == nil || itemEn.Image.Alt != "Grupo de alabanza" {
		t.Errorf("image en = %+v, se esperaba el alt en español por fallback", itemEn.Image)
	}
}

// TestGetMinisteriosHidesUnpublishedLeader fija FR-026/SC-014 en la capa de
// servicio: con `leaderVisible` apagado el JSON crudo NO contiene el nombre ni
// el celular/correo del encargado **en ninguna forma**, en los dos idiomas, y
// con el interruptor encendido la ficha sale con un único medio de contacto.
func TestGetMinisteriosHidesUnpublishedLeader(t *testing.T) {
	const (
		name  = "Ana Pérez"
		phone = "+58 412 000 0000"
		email = "ana@ejemplo.com"
	)
	repo := newFakeRepository()
	hidden := publishedMinistry("Alabanza")
	hidden.ID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	hidden.LeaderNameEs = textOf(name)
	hidden.LeaderNameEn = textOf("Ana Perez")
	hidden.LeaderPhone = textOf(phone)
	// `LeaderVisible` queda en false: «Oculto por defecto» (P6-3).
	visible := publishedMinistry("Jóvenes")
	visible.ID = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	visible.LeaderNameEs = textOf("Luis Gómez")
	visible.LeaderEmail = textOf(email)
	visible.LeaderVisible = true
	repo.seed(hidden, visible)

	svc := newTestService(repo)

	for _, lang := range []string{LangES, LangEN} {
		out, err := svc.GetMinisterios(context.Background(), lang)
		if err != nil {
			t.Fatalf("GetMinisterios(%s) = %v", lang, err)
		}
		raw := mustJSON(t, out)
		for _, secret := range []string{name, "Ana Perez", phone, "412", "000 0000"} {
			if strings.Contains(raw, secret) {
				t.Errorf("lang=%s: el JSON contiene %q: %s", lang, secret, raw)
			}
		}

		if len(out.Items) != 2 {
			t.Fatalf("items = %d, se esperaban 2", len(out.Items))
		}
		hiddenItem := out.Items[0]
		if hiddenItem.Leader != nil {
			t.Errorf("un ministerio con la ficha oculta lleva leader: %+v", hiddenItem.Leader)
		}
		visibleItem := out.Items[1]
		if visibleItem.Leader == nil {
			t.Fatal("el ministerio con la ficha publicada no lleva leader")
		}
		if visibleItem.Leader.Name != "Luis Gómez" {
			t.Errorf("leader.name = %q", visibleItem.Leader.Name)
		}
		if visibleItem.Leader.Email == nil || *visibleItem.Leader.Email != email {
			t.Errorf("leader.email = %v", visibleItem.Leader.Email)
		}
		if visibleItem.Leader.Phone != nil {
			t.Errorf("leader.phone = %v, se esperaba solo el correo (Q8)", visibleItem.Leader.Phone)
		}
	}
}

// TestGetMinisteriosDefaultsToHidden fija «Oculto por defecto» en la capa de
// servicio (P6-3/RG6-10): un alta sin `leaderVisible` nace oculta y su DTO
// público no trae `leader` aunque haya datos guardados.
func TestGetMinisteriosDefaultsToHidden(t *testing.T) {
	ministry, err := ministryFromInput(MinisterioInput{
		NameEs:        "Alabanza",
		DescriptionEs: "Grupo de alabanza",
		LeaderNameEs:  "Ana Pérez",
		LeaderPhone:   "+58 412 000 0000",
	})
	if err != nil {
		t.Fatalf("ministryFromInput = %v", err)
	}
	if ministry.LeaderVisible {
		t.Fatal("LeaderVisible = true, un alta sin la clave nace oculta (P6-3)")
	}
	ministry.PublicationState = StatePublished
	ministry.ID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

	repo := newFakeRepository()
	repo.seed(ministry)

	out, err := newTestService(repo).GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	if out.Items[0].Leader != nil {
		t.Errorf("leader = %+v, se esperaba la ficha oculta", out.Items[0].Leader)
	}
	if raw := mustJSON(t, out); strings.Contains(raw, "Ana Pérez") || strings.Contains(raw, "412") {
		t.Errorf("el JSON filtra el encargado oculto: %s", raw)
	}
}

// TestGetMinisteriosLeaderWithNameOnly fija Q8: un encargado solo con nombre (y
// publicado) deja la ficha sin línea de contacto, no una línea vacía.
func TestGetMinisteriosLeaderWithNameOnly(t *testing.T) {
	repo := newFakeRepository()
	ministry := publishedMinistry("Alabanza")
	ministry.LeaderNameEs = textOf("Ana Pérez")
	ministry.LeaderVisible = true
	repo.seed(ministry)

	out, err := newTestService(repo).GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	leader := out.Items[0].Leader
	if leader == nil {
		t.Fatal("se esperaba la ficha del encargado")
	}
	if leader.Name != "Ana Pérez" || leader.Phone != nil || leader.Email != nil {
		t.Errorf("leader = %+v, se esperaba solo el nombre", leader)
	}
	if raw := mustJSON(t, out); strings.Contains(raw, `"phone"`) || strings.Contains(raw, `"email"`) {
		t.Errorf("la ficha sin contacto lleva claves vacías: %s", raw)
	}
}

// TestGetMinisteriosOnlyPublished fija FR-012/SC-002: ningún borrador sale por
// la vía pública (0 borradores).
func TestGetMinisteriosOnlyPublished(t *testing.T) {
	repo := newFakeRepository()
	draft := publishedMinistry("Borrador")
	draft.PublicationState = StateDraft
	repo.seed(draft, publishedMinistry("Publicado"))

	out, err := newTestService(repo).GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].Name != "Publicado" {
		t.Fatalf("items = %+v, se esperaba solo el publicado", out.Items)
	}
	if raw := mustJSON(t, out); strings.Contains(raw, "Borrador") || strings.Contains(raw, "publicationState") {
		t.Errorf("el DTO público expone un borrador o el estado: %s", raw)
	}
}

// TestGetMinisteriosOmitsItemsWhenEmpty fija SC-010: sin publicados la clave
// `items` se omite (nunca `null` ni `[]`); la API responde 200 y no redirige
// (P6-21).
func TestGetMinisteriosOmitsItemsWhenEmpty(t *testing.T) {
	out, err := newTestService(newFakeRepository()).GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	if out.Items != nil {
		t.Errorf("items = %+v, se esperaba nil", out.Items)
	}
	if raw := mustJSON(t, out); raw != `{"lang":"es"}` {
		t.Errorf("JSON = %s, se esperaba solo lang", raw)
	}
}

// TestGetMinisteriosCapsAt200 fija el tope de seguridad del servidor (RG6-12):
// con más de 200 publicados se sirven los 200 primeros ya ordenados.
func TestGetMinisteriosCapsAt200(t *testing.T) {
	repo := newFakeRepository()
	for i := 0; i < maxPublicItems+3; i++ {
		ministry := publishedMinistry("Ministerio")
		ministry.SortOrder = i
		repo.seed(ministry)
	}

	out, err := newTestService(repo).GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	if len(out.Items) != maxPublicItems {
		t.Fatalf("items = %d, se esperaban %d", len(out.Items), maxPublicItems)
	}
	if got := repo.publicLimits; len(got) != 1 || got[0] != maxPublicItems {
		t.Errorf("topes consultados = %v, se esperaba [%d]", got, maxPublicItems)
	}
}

// TestGetMinisteriosInvalidLang fija el 400 invalid con details.lang.
func TestGetMinisteriosInvalidLang(t *testing.T) {
	_, err := newTestService(newFakeRepository()).GetMinisterios(context.Background(), "fr")
	details := detailsOf(t, err)
	if _, ok := details["lang"]; !ok {
		t.Fatalf("details = %v, se esperaba la clave lang", details)
	}
}

// TestGetMinisteriosRepositoryError fija que un fallo de datos no se disfraza:
// sale envuelto y con contexto.
func TestGetMinisteriosRepositoryError(t *testing.T) {
	repo := newFakeRepository()
	repo.readErr = errors.New("boom")

	_, err := newTestService(repo).GetMinisterios(context.Background(), LangES)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, se esperaba el error del repositorio", err)
	}
}

// TestPublicImageRequiresAlt fija el fail-closed de la imagen pública: sin
// archivo no hay imagen y un archivo sin texto alternativo tampoco se publica
// (nunca un `alt` vacío, SC-017).
func TestPublicImageRequiresAlt(t *testing.T) {
	if got := publicImage(nil, nil, nil, LangES); got != nil {
		t.Errorf("sin archivo = %+v, se esperaba nil", got)
	}
	if got := publicImage(textOf("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png"), nil, nil, LangES); got != nil {
		t.Errorf("sin alt = %+v, se esperaba nil", got)
	}
	got := publicImage(
		textOf("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png"),
		textOf("Grupo"),
		textOf("Group"),
		LangEN,
	)
	if got == nil || got.URL != MediaPathPrefix+"img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png" || got.Alt != "Group" {
		t.Errorf("imagen = %+v", got)
	}
}
