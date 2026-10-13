//go:build integration

package ministerios

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// Pruebas de integración de la lectura pública del dominio ministerios contra
// PostgreSQL real (F6, T609): `ListPublishedMinistries` es la ÚNICA vía pública y
// la privacidad del encargado se garantiza **en la consulta** (FR-026/SC-014,
// R6-2/RG6-1): con `leader_visible` apagado la fila sale con NULL —ni el nombre
// ni el celular/correo—, así que el dato no tiene por dónde viajar; con el
// interruptor encendido sale la ficha completa (y solo el nombre si no hay
// contacto). Fija además el «Oculto por defecto» del `DEFAULT FALSE` de la
// migración `000010` (P6-3/RG6-10), el orden `sort_order, id` y el `LIMIT`.
//
// Comparten los ayudantes de repository_ministerios_integration_test.go (mismo
// paquete y misma etiqueta de compilación).

// TestIntegrationListPublishedMinistriesHidesTheLeaderByDefault es la prueba más
// importante del módulo en la capa de datos: una fila con nombre y celular de
// encargado guardados correctamente y el interruptor en su DEFAULT sale de la
// lectura pública **sin** ningún dato del encargado (0 apariciones), mientras el
// panel los sigue viendo.
func TestIntegrationListPublishedMinistriesHidesTheLeaderByDefault(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	// Alta que OMITE `leader_visible`: la columna toma su DEFAULT.
	var id uuid.UUID
	if err := repo.pool.QueryRow(ctx, `INSERT INTO ministries
		(name_es, description_es, leader_name_es, leader_phone, publication_state)
		VALUES ($1, 'Grupo de alabanza', 'Ana Pérez', '+584120000000', 'published')
		RETURNING id`, "Alabanza "+uuid.NewString()).Scan(&id); err != nil {
		t.Fatalf("sembrar ministerio sin leader_visible: %v", err)
	}

	// El panel ve los datos guardados y el interruptor apagado: el dato guardado
	// no es el dato publicado (FR-025).
	detail, found, err := repo.GetMinistryByID(ctx, id)
	if err != nil || !found {
		t.Fatalf("GetMinistryByID = (found=%v, err=%v)", found, err)
	}
	if detail.LeaderVisible {
		t.Error("el DEFAULT de `leader_visible` no es FALSE («Oculto por defecto», P6-3)")
	}
	if detail.LeaderNameEs == nil || *detail.LeaderNameEs != "Ana Pérez" ||
		detail.LeaderPhone == nil || *detail.LeaderPhone != "+584120000000" {
		t.Fatalf("la fila no guardó los datos del encargado: %+v", detail.Ministry)
	}

	// La lectura pública no transporta NADA del encargado.
	items, err := repo.ListPublishedMinistries(ctx, 200)
	if err != nil {
		t.Fatalf("ListPublishedMinistries: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("la sección pública devolvió %d ministerios, se esperaba 1", len(items))
	}
	assertLeaderHidden(t, "ficha no publicada", items[0])

	// Y sigue oculta por más que el contenido sí sea público.
	if items[0].NameEs == "" || items[0].DescriptionEs == "" {
		t.Errorf("el contenido público se perdió: %+v", items[0])
	}
}

// TestIntegrationListPublishedMinistriesShowsThePublishedLeader fija el otro
// lado del interruptor: con `leader_visible = TRUE` la ficha se publica tal cual
// (nombre y, si existe, su único medio de contacto) y sin contacto sale solo el
// nombre (Q8: «solo nombre» es válido y no deja líneas vacías).
func TestIntegrationListPublishedMinistriesShowsThePublishedLeader(t *testing.T) {
	repo := newMinistryRepo(t)

	withPhone := baseMinistry("Alabanza")
	withPhone.LeaderVisible = true
	seedMinistry(t, repo, withPhone)

	withEmail := baseMinistry("Intercesión")
	withEmail.LeaderVisible = true
	withEmail.LeaderPhone = nil
	withEmail.LeaderEmail = ptr("ana@example.com")
	seedMinistry(t, repo, withEmail)

	onlyName := baseMinistry("Niños")
	onlyName.LeaderVisible = true
	onlyName.LeaderPhone = nil
	onlyName.LeaderEmail = nil
	seedMinistry(t, repo, onlyName)

	items, err := repo.ListPublishedMinistries(context.Background(), 200)
	if err != nil {
		t.Fatalf("ListPublishedMinistries: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("la sección pública devolvió %d ministerios, se esperaban 3", len(items))
	}

	byName := make(map[string]Ministry, len(items))
	for _, item := range items {
		byName[item.NameEs] = item
	}

	phone := byName["Alabanza"]
	if phone.LeaderNameEs == nil || *phone.LeaderNameEs != "Ana Pérez" {
		t.Errorf("con la ficha publicada falta el nombre: %+v", phone)
	}
	if phone.LeaderPhone == nil || *phone.LeaderPhone != "+584120000000" {
		t.Errorf("con la ficha publicada falta el celular: %+v", phone)
	}
	if phone.LeaderEmail != nil {
		t.Errorf("el correo viajó junto al celular: %+v", phone.LeaderEmail)
	}

	email := byName["Intercesión"]
	if email.LeaderEmail == nil || *email.LeaderEmail != "ana@example.com" {
		t.Errorf("con la ficha publicada falta el correo: %+v", email)
	}
	if email.LeaderPhone != nil {
		t.Errorf("el celular viajó junto al correo: %+v", email.LeaderPhone)
	}

	name := byName["Niños"]
	if name.LeaderNameEs == nil || *name.LeaderNameEs != "Ana Pérez" {
		t.Errorf("con la ficha publicada falta el nombre: %+v", name)
	}
	if name.LeaderPhone != nil || name.LeaderEmail != nil {
		t.Errorf("un encargado sin contacto no debe transportar ninguno: %+v", name)
	}
}

// TestIntegrationListPublishedMinistriesNeverReturnsDrafts fija SC-002 en la
// única vía pública: los borradores no salen del servidor, con datos de encargado
// o sin ellos.
func TestIntegrationListPublishedMinistriesNeverReturnsDrafts(t *testing.T) {
	repo := newMinistryRepo(t)

	published := baseMinistry("Alabanza")
	published.PublicationState = StatePublished
	seedMinistry(t, repo, published)

	draft := baseMinistry("Intercesión")
	draft.PublicationState = StateDraft
	seedMinistry(t, repo, draft)

	items, err := repo.ListPublishedMinistries(context.Background(), 200)
	if err != nil {
		t.Fatalf("ListPublishedMinistries: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("la sección pública devolvió %d ministerios, se esperaba 1 (sin borradores)", len(items))
	}
	if items[0].NameEs != "Alabanza" {
		t.Errorf("el visitante recibió %q, se esperaba el publicado", items[0].NameEs)
	}
	if items[0].PublicationState != StatePublished {
		t.Errorf("estado de la lectura pública = %q, se esperaba published", items[0].PublicationState)
	}

	// Un borrador nunca es contenido público por más que el resto de sus datos
	// estén completos y publicados.
	var drafts int
	if err := repo.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM ministries WHERE publication_state = 'draft'`).Scan(&drafts); err != nil {
		t.Fatalf("contar borradores: %v", err)
	}
	if drafts != 1 {
		t.Errorf("borradores sembrados = %d, se esperaba 1", drafts)
	}

	// Sin publicados, la lectura devuelve la colección vacía (el service omite
	// `items`; la redirección es del frontend, P6-21).
	if _, err := repo.pool.Exec(context.Background(), `DELETE FROM ministries`); err != nil {
		t.Fatalf("vaciar ministries: %v", err)
	}
	empty, err := repo.ListPublishedMinistries(context.Background(), 200)
	if err != nil {
		t.Fatalf("ListPublishedMinistries(sin publicados): %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("la sección pública devolvió %d ministerios sin publicados", len(empty))
	}
}

// TestIntegrationListPublishedMinistriesOrdersAndLimits fija FR-004/FR-027: el
// orden es `sort_order, id` (los 5 PRIMEROS del orden manual del módulo) y el
// `LIMIT` se respeta (5 en el bloque de la portada, 200 en la sección).
func TestIntegrationListPublishedMinistriesOrdersAndLimits(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	// Ocho publicados; los dos primeros comparten `sort_order` para fijar el
	// desempate por `id`.
	for i, order := range []int{2, 1, 1, 3, 4, 5, 6, 7} {
		ministry := baseMinistry(fmt.Sprintf("Ministerio %d", i))
		ministry.SortOrder = order
		seedMinistry(t, repo, ministry)
	}

	all, err := repo.ListPublishedMinistries(ctx, 200)
	if err != nil {
		t.Fatalf("ListPublishedMinistries(200): %v", err)
	}
	if len(all) != 8 {
		t.Fatalf("la sección devolvió %d ministerios, se esperaban 8", len(all))
	}
	assertOrderedBySortOrderThenID(t, all)

	// El bloque de la portada son los 5 PRIMEROS del mismo orden (tope firme del
	// servidor, FR-027): ni uno distinto ni uno de menos.
	block, err := repo.ListPublishedMinistries(ctx, 5)
	if err != nil {
		t.Fatalf("ListPublishedMinistries(5): %v", err)
	}
	if len(block) != 5 {
		t.Fatalf("el bloque devolvió %d ministerios, se esperaban 5", len(block))
	}
	for i := range block {
		if block[i].ID != all[i].ID {
			t.Fatalf("el bloque no son los 5 primeros del orden: posición %d = %v, se esperaba %v",
				i, block[i].ID, all[i].ID)
		}
	}
}

// assertLeaderHidden comprueba que la entidad pública no transporta ningún dato
// del encargado (ni el nombre ni el contacto), que es la garantía de FR-026 en
// la capa de datos.
func assertLeaderHidden(t *testing.T, context string, item Ministry) {
	t.Helper()
	if item.LeaderNameEs != nil || item.LeaderNameEn != nil {
		t.Errorf("%s: la lectura pública transportó el nombre del encargado: %v / %v",
			context, item.LeaderNameEs, item.LeaderNameEn)
	}
	if item.LeaderPhone != nil || item.LeaderEmail != nil {
		t.Errorf("%s: la lectura pública transportó el contacto del encargado: %v / %v",
			context, item.LeaderPhone, item.LeaderEmail)
	}
	if item.LeaderVisible {
		t.Errorf("%s: la lectura pública dice que la ficha está publicada", context)
	}
}

// assertOrderedBySortOrderThenID comprueba el orden manual del módulo
// (`sort_order`, `id`), el mismo que ve el visitante y el panel (P6-19).
func assertOrderedBySortOrderThenID(t *testing.T, items []Ministry) {
	t.Helper()
	for i := 1; i < len(items); i++ {
		prev, current := items[i-1], items[i]
		if prev.SortOrder > current.SortOrder ||
			(prev.SortOrder == current.SortOrder && prev.ID.String() > current.ID.String()) {
			t.Fatalf("orden roto en la posición %d: (sort %d, id %v) antes de (sort %d, id %v)",
				i, prev.SortOrder, prev.ID, current.SortOrder, current.ID)
		}
	}
}
