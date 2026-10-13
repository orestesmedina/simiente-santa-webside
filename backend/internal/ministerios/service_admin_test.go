package ministerios

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
)

// Tests de la gestión del panel (T615): alta, edición parcial con fusión
// (A6-3/P6-20), publicación, visibilidad del encargado, borrado, listado y
// auditoría con la regla exacta de códigos (R6-12/A6-10) y el ciclo de vida de
// los archivos (A6-11). El repositorio es el falso en memoria; el almacén, un
// falso que recuerda lo que se guardó y se borró.

// fakeStore es un almacén de imágenes en memoria para las pruebas.
type fakeStore struct {
	files     map[string][]byte
	saveErr   error
	deleteErr error
	saved     []string
	deleted   []string
}

func newFakeStore() *fakeStore { return &fakeStore{files: map[string][]byte{}} }

func (s *fakeStore) Save(_ context.Context, data []byte) (storage.File, error) {
	if s.saveErr != nil {
		return storage.File{}, s.saveErr
	}
	name := "img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.jpg"
	s.files[name] = data
	s.saved = append(s.saved, name)
	return storage.File{Name: name, ContentType: "image/jpeg", SizeBytes: int64(len(data))}, nil
}

func (s *fakeStore) Open(_ context.Context, name string) (io.ReadCloser, error) {
	data, ok := s.files[name]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return io.NopCloser(strings.NewReader(string(data))), nil
}

func (s *fakeStore) Delete(_ context.Context, name string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.files, name)
	s.deleted = append(s.deleted, name)
	return nil
}

// newAdminService construye el service del panel con el almacén indicado.
func newAdminService(repo Repository, store storage.Store) *service {
	return NewService(ServiceDeps{Repository: repo, Store: store})
}

// actorID es el actor de todas las mutaciones de las pruebas.
func actorID() uuid.UUID { return uuid.MustParse("22222222-2222-2222-2222-222222222222") }

// storedMinistry construye un ministerio ya guardado del panel.
func storedMinistry(name string) Ministry {
	return Ministry{
		ID:               uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		NameEs:           name,
		DescriptionEs:    "Descripción de " + name,
		PublicationState: StateDraft,
	}
}

// actionsOf devuelve los códigos de las acciones comprometidas con la mutación,
// en orden.
func actionsOf(repo *fakeRepository) []string {
	codes := make([]string, 0, len(repo.committed))
	for _, action := range repo.committed {
		codes = append(codes, action.Code)
	}
	return codes
}

// TestCreateMinisterioStoresAndAuditsOnlyCreate fija FR-013 y A6-10: el alta
// guarda con los defaults del contrato y registra **solo `ministry.create`**,
// aunque nazca publicado y con la ficha del encargado visible (cero filas
// `publish`/`unpublish`/`leader.show`/`leader.hide`).
func TestCreateMinisterioStoresAndAuditsOnlyCreate(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(repo, nil)

	visible := true
	ministry, err := svc.CreateMinisterio(context.Background(), actorID(), MinisterioInput{
		NameEs:           "  Alabanza  ",
		DescriptionEs:    "Grupo de alabanza",
		LeaderNameEs:     "Ana Pérez",
		LeaderPhone:      "+58 412 000 0000",
		LeaderVisible:    visible,
		PublicationState: "published",
		SortOrder:        3,
	})
	if err != nil {
		t.Fatalf("CreateMinisterio = %v", err)
	}

	if ministry.NameEs != "Alabanza" {
		t.Errorf("nameEs = %q, se esperaba normalizado", ministry.NameEs)
	}
	if ministry.PublicationState != string(StatePublished) {
		t.Errorf("publicationState = %q", ministry.PublicationState)
	}
	if !ministry.LeaderVisible {
		t.Error("leaderVisible = false, se esperaba true")
	}
	if ministry.SortOrder != 3 {
		t.Errorf("sortOrder = %d", ministry.SortOrder)
	}
	if len(repo.ministries) != 1 {
		t.Fatalf("no se guardó el ministerio: %+v", repo.ministries)
	}

	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionMinistryCreate {
		t.Fatalf("acciones = %v, se esperaba solo %s", got, audit.ActionMinistryCreate)
	}
	if label := repo.committed[0].TargetLabel; label != "Ministerios · Alabanza" {
		t.Errorf("targetLabel = %q", label)
	}
}

// TestCreateMinisterioDefaultsToHidden fija «Oculto por defecto» (P6-3): sin
// `leaderVisible` en el DTO, el ministerio nace con la ficha NO publicada y la
// vía pública no la lleva —aunque el encargado esté guardado—.
func TestCreateMinisterioDefaultsToHidden(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(repo, nil)

	created, err := svc.CreateMinisterio(context.Background(), actorID(), MinisterioInput{
		NameEs:           "Alabanza",
		DescriptionEs:    "Grupo de alabanza",
		LeaderNameEs:     "Ana Pérez",
		LeaderPhone:      "+58 412 000 0000",
		PublicationState: "published",
	})
	if err != nil {
		t.Fatalf("CreateMinisterio = %v", err)
	}
	if created.LeaderVisible {
		t.Fatal("leaderVisible = true, un alta sin la clave nace oculta (P6-3)")
	}
	if created.LeaderNameEs == nil || created.LeaderPhone == nil {
		t.Error("el panel debe ver los datos guardados del encargado (dato guardado ≠ publicado)")
	}

	public, err := svc.GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	if len(public.Items) != 1 || public.Items[0].Leader != nil {
		t.Fatalf("el DTO público lleva la ficha oculta: %+v", public.Items)
	}
	if raw := mustJSON(t, public); strings.Contains(raw, "Ana Pérez") || strings.Contains(raw, "412") {
		t.Errorf("el JSON público filtra el encargado oculto: %s", raw)
	}
}

// TestCreateMinisterioRejectsInvalidFields fija FR-014: cada campo inválido →
// `400 invalid` con `details` por campo y **sin guardar nada**.
func TestCreateMinisterioRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		input  MinisterioInput
		detail []string
	}{
		{
			name:   "sin nombre",
			input:  MinisterioInput{DescriptionEs: "D"},
			detail: []string{"nameEs"},
		},
		{
			name:   "sin descripción",
			input:  MinisterioInput{NameEs: "Alabanza"},
			detail: []string{"descriptionEs"},
		},
		{
			name: "inglés sin español",
			input: MinisterioInput{
				NameEs: "Alabanza", DescriptionEs: "D", ScheduleEn: "Every Friday",
			},
			detail: []string{"scheduleEs"},
		},
		{
			name: "los dos contactos",
			input: MinisterioInput{
				NameEs: "Alabanza", DescriptionEs: "D",
				LeaderNameEs: "Ana", LeaderPhone: "+58 412 000 0000", LeaderEmail: "ana@ejemplo.com",
			},
			detail: []string{"leaderEmail"},
		},
		{
			name: "contacto sin nombre",
			input: MinisterioInput{
				NameEs: "Alabanza", DescriptionEs: "D", LeaderEmail: "ana@ejemplo.com",
			},
			detail: []string{"leaderNameEs"},
		},
		{
			name: "enlace javascript",
			input: MinisterioInput{
				NameEs: "Alabanza", DescriptionEs: "D", WhatsappURL: "javascript:alert(1)",
			},
			detail: []string{"whatsappUrl"},
		},
		{
			name: "enlace de 501 caracteres",
			input: MinisterioInput{
				NameEs: "Alabanza", DescriptionEs: "D",
				WhatsappURL: "https://ejemplo.com/" + strings.Repeat("a", 481),
			},
			detail: []string{"whatsappUrl"},
		},
		{
			name: "imagen sin alt",
			input: MinisterioInput{
				NameEs: "Alabanza", DescriptionEs: "D",
				ImageFile: "img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png",
			},
			detail: []string{"imageAltEs"},
		},
		{
			name: "nombre de 161 caracteres",
			input: MinisterioInput{
				NameEs: strings.Repeat("a", MaxNameLength+1), DescriptionEs: "D",
			},
			detail: []string{"nameEs"},
		},
		{
			name: "estado fuera del registro",
			input: MinisterioInput{
				NameEs: "Alabanza", DescriptionEs: "D", PublicationState: "archived",
			},
			detail: []string{"publicationState"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			svc := newAdminService(repo, nil)

			_, err := svc.CreateMinisterio(context.Background(), actorID(), tt.input)
			details := detailsOf(t, err)
			for _, field := range tt.detail {
				if _, ok := details[field]; !ok {
					t.Fatalf("details = %v, se esperaba la clave %q", details, field)
				}
			}
			if len(repo.ministries) != 0 {
				t.Errorf("un DTO inválido guardó el ministerio: %+v", repo.ministries)
			}
			if len(repo.committed) != 0 {
				t.Errorf("un DTO inválido dejó filas de auditoría: %v", actionsOf(repo))
			}
			if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
				t.Errorf("se esperaba un rechazo registrado con result=failure: %+v", repo.recorded)
			}
		})
	}
}

// TestCreateMinisterioDuplicateIsConflict fija P6-12: el duplicado exacto →
// `409 conflict` con el mensaje del módulo, sin guardar nada.
func TestCreateMinisterioDuplicateIsConflict(t *testing.T) {
	repo := newFakeRepository()
	repo.duplicate = true
	svc := newAdminService(repo, nil)

	_, err := svc.CreateMinisterio(context.Background(), actorID(), MinisterioInput{
		NameEs: "Alabanza", DescriptionEs: "D",
	})
	if kind := errorKind(err); kind != apperr.KindConflict {
		t.Fatalf("kind = %v, se esperaba conflict (%v)", kind, err)
	}
	if len(repo.ministries) != 0 {
		t.Errorf("un duplicado no debe guardarse: %+v", repo.ministries)
	}
}

// TestCreateMinisterioSharedLinkIsAllowed fija RG6-11: compartir el enlace es un
// aviso (`whatsappUrlShared`), nunca un bloqueo.
func TestCreateMinisterioSharedLinkIsAllowed(t *testing.T) {
	const link = "https://chat.whatsapp.com/compartido"
	repo := newFakeRepository()
	existing := storedMinistry("Alabanza")
	existing.WhatsappURL = textOf(link)
	repo.seed(existing)
	svc := newAdminService(repo, nil)

	created, err := svc.CreateMinisterio(context.Background(), actorID(), MinisterioInput{
		NameEs: "Jóvenes", DescriptionEs: "Grupo de jóvenes", WhatsappURL: link,
	})
	if err != nil {
		t.Fatalf("CreateMinisterio = %v, el enlace compartido no debe bloquear", err)
	}
	if !created.WhatsappURLShared {
		t.Error("whatsappUrlShared = false, se esperaba el aviso")
	}

	// El aviso se recalcula en cada lectura: el existente también lo ve.
	got, err := svc.GetMinisterio(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("GetMinisterio = %v", err)
	}
	if !got.WhatsappURLShared {
		t.Error("el ministerio existente no ve el enlace compartido")
	}
}

// TestUpdateMinisterioMergesOnlyPresentFields fija A6-3/P6-20: el PATCH fusiona
// los campos presentes sobre la fila actual (los ausentes no cambian) y un
// `null` explícito borra el opcional.
func TestUpdateMinisterioMergesOnlyPresentFields(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	current.NameEn = textOf("Worship")
	current.PlaceEs = textOf("Templo")
	current.ScheduleEs = textOf("Todos los viernes")
	current.LeaderNameEs = textOf("Ana Pérez")
	current.LeaderPhone = textOf("+58 412 000 0000")
	current.LeaderVisible = true
	current.WhatsappURL = textOf("https://chat.whatsapp.com/abc")
	current.ImageFile = textOf("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png")
	current.ImageAltEs = textOf("Grupo de alabanza")
	current.SortOrder = 2
	repo.seed(current)
	svc := newAdminService(repo, nil)

	// Un solo campo: el resto queda intacto.
	updated, err := svc.UpdateMinisterio(context.Background(), actorID(), current.ID, MinisterioPatch{
		NameEs: textOf("  Alabanza y adoración  "),
	})
	if err != nil {
		t.Fatalf("UpdateMinisterio = %v", err)
	}
	if updated.NameEs != "Alabanza y adoración" {
		t.Errorf("nameEs = %q", updated.NameEs)
	}
	if updated.DescriptionEs != current.DescriptionEs || updated.SortOrder != 2 {
		t.Errorf("la fusión perdió datos: %+v", updated)
	}
	if updated.NameEn == nil || *updated.NameEn != "Worship" {
		t.Errorf("nameEn = %v, se esperaba intacto", updated.NameEn)
	}
	if updated.PlaceEs == nil || *updated.PlaceEs != "Templo" {
		t.Errorf("placeEs = %v, se esperaba intacto", updated.PlaceEs)
	}
	if updated.ImageFile == nil || *updated.ImageFile != *current.ImageFile {
		t.Errorf("imageFile = %v, se esperaba intacto", updated.ImageFile)
	}
	if updated.LeaderNameEs == nil || *updated.LeaderNameEs != "Ana Pérez" || !updated.LeaderVisible {
		t.Errorf("la ficha del encargado cambió: %+v", updated)
	}
	// La fila COMPLETA ya fusionada es lo que se escribe (última escritura
	// completa gana): el almacén tiene el resultado, no un parche.
	stored := repo.ministries[0]
	if stored.NameEn == nil || stored.DescriptionEs == "" || stored.SortOrder != 2 {
		t.Errorf("la fila escrita no es la fusionada: %+v", stored)
	}

	// `null` explícito borra los opcionales (y solo esos).
	cleared, err := svc.UpdateMinisterio(context.Background(), actorID(), current.ID, MinisterioPatch{
		WhatsappURL: OptionalNull[string](),
		ScheduleEs:  OptionalOf(""),
		ImageFile:   OptionalNull[string](),
	})
	if err != nil {
		t.Fatalf("UpdateMinisterio(null) = %v", err)
	}
	if cleared.WhatsappURL != nil {
		t.Errorf("whatsappUrl = %v, se esperaba borrado", cleared.WhatsappURL)
	}
	if cleared.ScheduleEs != nil {
		t.Errorf("scheduleEs = %v, se esperaba borrado", cleared.ScheduleEs)
	}
	if cleared.ImageFile != nil {
		t.Errorf("imageFile = %v, se esperaba borrado", cleared.ImageFile)
	}
	if cleared.NameEn == nil || cleared.PlaceEs == nil || cleared.LeaderPhone == nil {
		t.Errorf("un null explícito borró campos que no venían en el PATCH: %+v", cleared)
	}
}

// TestUpdateMinisterioNullExplicitFromJSON fija que un `null` **del JSON** (no
// construido en código) borra el opcional y que un campo ausente no cambia.
func TestUpdateMinisterioNullExplicitFromJSON(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	current.WhatsappURL = textOf("https://chat.whatsapp.com/abc")
	current.PlaceEs = textOf("Templo")
	repo.seed(current)

	body := `{"whatsappUrl":null,"placeEs":"  Templo nuevo  "}`
	var patch MinisterioPatch
	if err := json.Unmarshal([]byte(body), &patch); err != nil {
		t.Fatalf("deserializar el PATCH: %v", err)
	}
	if !patch.hasChanges() {
		t.Fatal("hasChanges = false, el PATCH trae dos claves")
	}

	updated, err := newAdminService(repo, nil).UpdateMinisterio(context.Background(), actorID(), current.ID, patch)
	if err != nil {
		t.Fatalf("UpdateMinisterio = %v", err)
	}
	if updated.WhatsappURL != nil {
		t.Errorf("whatsappUrl = %v, se esperaba borrado", updated.WhatsappURL)
	}
	if updated.PlaceEs == nil || *updated.PlaceEs != "Templo nuevo" {
		t.Errorf("placeEs = %v", updated.PlaceEs)
	}
	if updated.DescriptionEs != current.DescriptionEs {
		t.Errorf("descriptionEs = %q, se esperaba intacta", updated.DescriptionEs)
	}
}

// TestUpdateMinisterioLeaderToggleRegistersAndKeepsData fija FR-025/US2 esc.
// 10–11 y R6-12: apagar y encender el interruptor deja su fila
// (`leader.hide`/`leader.show`) y **conserva** nombre y contacto.
func TestUpdateMinisterioLeaderToggleRegistersAndKeepsData(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	current.LeaderNameEs = textOf("Ana Pérez")
	current.LeaderPhone = textOf("+58 412 000 0000")
	current.LeaderVisible = true
	current.PublicationState = StatePublished
	repo.seed(current)
	svc := newAdminService(repo, nil)

	hidden, err := svc.UpdateMinisterio(context.Background(), actorID(), current.ID, MinisterioPatch{
		LeaderVisible: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("UpdateMinisterio(hide) = %v", err)
	}
	if hidden.LeaderVisible {
		t.Error("leaderVisible = true, se esperaba apagado")
	}
	if hidden.LeaderNameEs == nil || hidden.LeaderPhone == nil {
		t.Error("apagar el interruptor borró los datos del encargado (dato guardado ≠ publicado)")
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionMinistryLeaderHide {
		t.Fatalf("acciones = %v, se esperaba solo leader.hide", got)
	}

	// Con la ficha oculta, el DTO público no la lleva.
	public, err := svc.GetMinisterios(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetMinisterios = %v", err)
	}
	if len(public.Items) != 1 || public.Items[0].Leader != nil {
		t.Fatalf("la ficha oculta salió al público: %+v", public.Items)
	}

	repo.committed = nil
	shown, err := svc.UpdateMinisterio(context.Background(), actorID(), current.ID, MinisterioPatch{
		LeaderVisible: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("UpdateMinisterio(show) = %v", err)
	}
	if !shown.LeaderVisible {
		t.Error("leaderVisible = false, se esperaba encendido")
	}
	if shown.LeaderNameEs == nil || *shown.LeaderNameEs != "Ana Pérez" {
		t.Error("encender el interruptor exigió reescribir los datos")
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionMinistryLeaderShow {
		t.Fatalf("acciones = %v, se esperaba solo leader.show", got)
	}
}

// TestUpdateMinisterioRegistersOneRowPerOperation fija R6-12: datos + estado +
// visibilidad en la misma operación → una fila por operación, en la misma
// transacción.
func TestUpdateMinisterioRegistersOneRowPerOperation(t *testing.T) {
	tests := []struct {
		name    string
		patch   MinisterioPatch
		want    []string
		initial Ministry
	}{
		{
			name:    "solo datos",
			initial: Ministry{PublicationState: StateDraft},
			patch:   MinisterioPatch{NameEs: textOf("Otro")},
			want:    []string{audit.ActionMinistryUpdate},
		},
		{
			name:    "solo estado",
			initial: Ministry{PublicationState: StateDraft},
			patch:   MinisterioPatch{PublicationState: textOf("published")},
			want:    []string{audit.ActionMinistryPublish},
		},
		{
			name:    "solo retirada",
			initial: Ministry{PublicationState: StatePublished},
			patch:   MinisterioPatch{PublicationState: textOf("draft")},
			want:    []string{audit.ActionMinistryUnpublish},
		},
		{
			name:    "datos y estado",
			initial: Ministry{PublicationState: StateDraft},
			patch:   MinisterioPatch{NameEs: textOf("Otro"), PublicationState: textOf("published")},
			want:    []string{audit.ActionMinistryUpdate, audit.ActionMinistryPublish},
		},
		{
			name:    "datos, estado y visibilidad",
			initial: Ministry{PublicationState: StateDraft, LeaderNameEs: textOf("Ana")},
			patch: MinisterioPatch{
				SortOrder:        intPtr(5),
				PublicationState: textOf("published"),
				LeaderVisible:    boolPtr(true),
			},
			want: []string{audit.ActionMinistryUpdate, audit.ActionMinistryPublish, audit.ActionMinistryLeaderShow},
		},
		{
			name:    "estado igual no registra",
			initial: Ministry{PublicationState: StatePublished},
			patch:   MinisterioPatch{PublicationState: textOf("published")},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			current := storedMinistry("Alabanza")
			current.PublicationState = tt.initial.PublicationState
			current.LeaderNameEs = tt.initial.LeaderNameEs
			if tt.initial.PublicationState == StatePublished {
				current.PublicationState = StatePublished
			}
			repo.seed(current)

			_, err := newAdminService(repo, nil).UpdateMinisterio(context.Background(), actorID(), current.ID, tt.patch)
			if err != nil {
				t.Fatalf("UpdateMinisterio = %v", err)
			}
			got := actionsOf(repo)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("acciones = %v, se esperaban %v", got, tt.want)
			}
		})
	}
}

// TestUpdateMinisterioValidationFailsWithoutSaving fija FR-014/A6-3: el
// resultado fusionado se revalida y un rechazo no guarda nada (fail-closed).
func TestUpdateMinisterioValidationFailsWithoutSaving(t *testing.T) {
	tests := []struct {
		name   string
		patch  MinisterioPatch
		detail string
	}{
		{
			name:   "nombre vacío",
			patch:  MinisterioPatch{NameEs: textOf("   ")},
			detail: "nameEs",
		},
		{
			name:   "horario en inglés sin español",
			patch:  MinisterioPatch{ScheduleEn: OptionalOf("Every Friday")},
			detail: "scheduleEs",
		},
		{
			name: "los dos contactos",
			patch: MinisterioPatch{
				LeaderNameEs: OptionalOf("Ana"),
				LeaderPhone:  OptionalOf("+58 412 000 0000"),
				LeaderEmail:  OptionalOf("ana@ejemplo.com"),
			},
			detail: "leaderEmail",
		},
		{
			name:   "enlace no https",
			patch:  MinisterioPatch{WhatsappURL: OptionalOf("javascript:alert(1)")},
			detail: "whatsappUrl",
		},
		{
			name:   "imagen sin alt",
			patch:  MinisterioPatch{ImageFile: OptionalOf("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png")},
			detail: "imageAltEs",
		},
		{
			name:   "texto alternativo sin imagen",
			patch:  MinisterioPatch{ImageAltEn: OptionalOf("Group")},
			detail: "imageAltEs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			current := storedMinistry("Alabanza")
			repo.seed(current)

			_, err := newAdminService(repo, nil).UpdateMinisterio(context.Background(), actorID(), current.ID, tt.patch)
			details := detailsOf(t, err)
			if _, ok := details[tt.detail]; !ok {
				t.Fatalf("details = %v, se esperaba la clave %q", details, tt.detail)
			}
			if repo.ministries[0].NameEs != "Alabanza" || repo.ministries[0].SortOrder != 0 {
				t.Errorf("un PATCH inválido cambió la fila: %+v", repo.ministries[0])
			}
			if len(repo.committed) != 0 {
				t.Errorf("un PATCH inválido dejó auditoría: %v", actionsOf(repo))
			}
		})
	}
}

// TestUpdateMinisterioEmptyPatchAndNotFound fija los dos rechazos de entrada:
// un PATCH sin campos → 400; un id inexistente → 404 (y en ambos casos queda el
// rechazo registrado, R6-12).
func TestUpdateMinisterioEmptyPatchAndNotFound(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(repo, nil)

	_, err := svc.UpdateMinisterio(context.Background(), actorID(), uuid.New(), MinisterioPatch{})
	if kind := errorKind(err); kind != apperr.KindInvalid {
		t.Fatalf("PATCH vacío: kind = %v, se esperaba invalid (%v)", kind, err)
	}

	_, err = svc.UpdateMinisterio(context.Background(), actorID(), uuid.New(), MinisterioPatch{NameEs: textOf("Otro")})
	if kind := errorKind(err); kind != apperr.KindNotFound {
		t.Fatalf("id inexistente: kind = %v, se esperaba not_found (%v)", kind, err)
	}

	if len(repo.committed) != 0 {
		t.Errorf("un rechazo no debe comprometer auditoría: %v", actionsOf(repo))
	}
	if len(repo.recorded) != 2 {
		t.Errorf("se esperaban dos rechazos registrados: %+v", repo.recorded)
	}
}

// TestUpdateMinisterioMutationFailureSavesNothing fija el fail-closed de la
// transacción: si el guardado falla, no queda fila escrita.
func TestUpdateMinisterioMutationFailureSavesNothing(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	repo.seed(current)
	repo.mutationErr = errors.New("la transacción falló")

	_, err := newAdminService(repo, nil).UpdateMinisterio(context.Background(), actorID(), current.ID, MinisterioPatch{
		NameEs: textOf("Otro"),
	})
	if err == nil || !strings.Contains(err.Error(), "la transacción falló") {
		t.Fatalf("err = %v", err)
	}
	if repo.ministries[0].NameEs != "Alabanza" {
		t.Errorf("la fila cambió pese al fallo: %+v", repo.ministries[0])
	}
}

// TestUpdateMinisterioReplacedImageIsCleaned fija A6-11: reemplazar o quitar la
// imagen limpia el archivo anterior best-effort; conservarla, no; y un fallo de
// limpieza no cambia la respuesta.
func TestUpdateMinisterioReplacedImageIsCleaned(t *testing.T) {
	const previous = "img_bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb.png"
	const replacement = "img_cccccccc-cccc-cccc-cccc-cccccccccccc.png"

	tests := []struct {
		name        string
		patch       MinisterioPatch
		wantDeleted []string
	}{
		{
			name:        "imagen reemplazada",
			patch:       MinisterioPatch{ImageFile: OptionalOf(replacement), ImageAltEs: OptionalOf("Nueva")},
			wantDeleted: []string{previous},
		},
		{
			name:        "imagen quitada",
			patch:       MinisterioPatch{ImageFile: OptionalNull[string]()},
			wantDeleted: []string{previous},
		},
		{
			name:        "imagen intacta",
			patch:       MinisterioPatch{NameEs: textOf("Otro nombre")},
			wantDeleted: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			current := storedMinistry("Alabanza")
			current.ImageFile = textOf(previous)
			current.ImageAltEs = textOf("Grupo")
			repo.seed(current)
			store := newFakeStore()
			store.files[previous] = []byte("png")

			_, err := newAdminService(repo, store).UpdateMinisterio(context.Background(), actorID(), current.ID, tt.patch)
			if err != nil {
				t.Fatalf("UpdateMinisterio = %v", err)
			}
			if strings.Join(store.deleted, ",") != strings.Join(tt.wantDeleted, ",") {
				t.Errorf("archivos borrados = %v, se esperaban %v", store.deleted, tt.wantDeleted)
			}
		})
	}

	t.Run("un fallo de limpieza no cambia la respuesta", func(t *testing.T) {
		repo := newFakeRepository()
		current := storedMinistry("Alabanza")
		current.ImageFile = textOf(previous)
		current.ImageAltEs = textOf("Grupo")
		repo.seed(current)
		store := newFakeStore()
		store.deleteErr = errors.New("no se pudo borrar")

		updated, err := newAdminService(repo, store).UpdateMinisterio(context.Background(), actorID(), current.ID, MinisterioPatch{
			ImageFile:  OptionalNull[string](),
			ImageAltEs: OptionalNull[string](),
		})
		if err != nil {
			t.Fatalf("UpdateMinisterio = %v, la limpieza es best-effort", err)
		}
		if updated.ImageFile != nil {
			t.Errorf("imageFile = %v, se esperaba borrado", updated.ImageFile)
		}
	})
}

// TestDeleteMinisterioKeepsTheFile fija A6-11: eliminar el ministerio **no**
// borra el archivo de su imagen (deja de servirse porque `IsFilePublished` ya no
// lo encuentra) y deja su fila `ministry.delete`.
func TestDeleteMinisterioKeepsTheFile(t *testing.T) {
	const file = "img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png"
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	current.ImageFile = textOf(file)
	repo.seed(current)
	store := newFakeStore()
	store.files[file] = []byte("png")

	if err := newAdminService(repo, store).DeleteMinisterio(context.Background(), actorID(), current.ID); err != nil {
		t.Fatalf("DeleteMinisterio = %v", err)
	}
	if len(repo.ministries) != 0 {
		t.Errorf("el ministerio sigue guardado: %+v", repo.ministries)
	}
	if len(store.deleted) != 0 {
		t.Errorf("se borró el archivo al eliminar el ministerio: %v", store.deleted)
	}
	if _, ok := store.files[file]; !ok {
		t.Error("el archivo desapareció del disco")
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionMinistryDelete {
		t.Fatalf("acciones = %v, se esperaba solo ministry.delete", got)
	}
	if label := repo.committed[0].TargetLabel; label != "Ministerios · Alabanza" {
		t.Errorf("targetLabel = %q", label)
	}
}

// TestDeleteMinisterioNotFound fija el 404 del borrado.
func TestDeleteMinisterioNotFound(t *testing.T) {
	err := newAdminService(newFakeRepository(), nil).DeleteMinisterio(context.Background(), actorID(), uuid.New())
	if kind := errorKind(err); kind != apperr.KindNotFound {
		t.Fatalf("kind = %v, se esperaba not_found (%v)", kind, err)
	}
}

// TestListMinisteriosIsPaginatedAndOrdered fija FR-010/P6-19: el listado incluye
// borradores, va en el orden del módulo (`sortOrder`, `id`), respeta la página y
// admite el filtro por estado.
func TestListMinisteriosIsPaginatedAndOrdered(t *testing.T) {
	repo := newFakeRepository()
	first := storedMinistry("Primero")
	first.SortOrder = 1
	second := storedMinistry("Segundo")
	second.ID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	second.SortOrder = 2
	second.PublicationState = StatePublished
	repo.seed(first, second)
	svc := newAdminService(repo, nil)

	all, err := svc.ListMinisterios(context.Background(), nil, paginate.Params{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListMinisterios = %v", err)
	}
	if all.Total != 2 || len(all.Items) != 2 || all.Limit != 10 || all.Offset != 0 {
		t.Fatalf("listado = %+v", all)
	}
	if all.Items[0].NameEs != "Primero" || all.Items[1].NameEs != "Segundo" {
		t.Errorf("orden = %s, %s", all.Items[0].NameEs, all.Items[1].NameEs)
	}

	page, err := svc.ListMinisterios(context.Background(), nil, paginate.Params{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListMinisterios(página) = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.Items[0].NameEs != "Segundo" {
		t.Errorf("página = %+v", page)
	}

	published := StatePublished
	filtered, err := svc.ListMinisterios(context.Background(), &published, paginate.Params{Limit: 10})
	if err != nil {
		t.Fatalf("ListMinisterios(filtro) = %v", err)
	}
	if filtered.Total != 1 || len(filtered.Items) != 1 || filtered.Items[0].NameEs != "Segundo" {
		t.Errorf("filtro = %+v", filtered)
	}
}

// TestGetMinisterioNotFound fija el 404 del detalle.
func TestGetMinisterioNotFound(t *testing.T) {
	_, err := newAdminService(newFakeRepository(), nil).GetMinisterio(context.Background(), uuid.New())
	if kind := errorKind(err); kind != apperr.KindNotFound {
		t.Fatalf("kind = %v, se esperaba not_found (%v)", kind, err)
	}
}

// TestUploadMinisterioImageAuditsAndFailsClosed fija R6-7: cada subida deja su
// fila `ministry.image.upload`; un tipo no permitido no deja archivo y un fallo
// de registro borra el archivo (fail-closed).
func TestUploadMinisterioImageAuditsAndFailsClosed(t *testing.T) {
	t.Run("subida correcta", func(t *testing.T) {
		repo := newFakeRepository()
		store := newFakeStore()
		result, err := newAdminService(repo, store).UploadMinisterioImage(context.Background(), actorID(), []byte("jpeg"))
		if err != nil {
			t.Fatalf("UploadMinisterioImage = %v", err)
		}
		if !storage.ValidName(result.FileName) || result.URL != MediaPathPrefix+result.FileName {
			t.Errorf("resultado = %+v", result)
		}
		if len(repo.recorded) != 1 || repo.recorded[0].Code != audit.ActionMinistryImageUpload {
			t.Fatalf("auditoría = %+v", repo.recorded)
		}
		if want := imageLabel(result.FileName); repo.recorded[0].TargetLabel != want {
			t.Errorf("targetLabel = %q, se esperaba %q", repo.recorded[0].TargetLabel, want)
		}
	})

	t.Run("formato no permitido", func(t *testing.T) {
		repo := newFakeRepository()
		store := newFakeStore()
		store.saveErr = storage.ErrInvalidFormat

		_, err := newAdminService(repo, store).UploadMinisterioImage(context.Background(), actorID(), []byte("svg"))
		details := detailsOf(t, err)
		if _, ok := details["file"]; !ok {
			t.Fatalf("details = %v, se esperaba la clave file", details)
		}
		if len(repo.recorded) != 0 {
			t.Errorf("un formato rechazado no debe auditarse: %+v", repo.recorded)
		}
	})

	t.Run("demasiado grande", func(t *testing.T) {
		store := newFakeStore()
		store.saveErr = storage.ErrTooLarge

		_, err := newAdminService(newFakeRepository(), store).UploadMinisterioImage(context.Background(), actorID(), []byte("x"))
		details := detailsOf(t, err)
		if _, ok := details["fileSize"]; !ok {
			t.Fatalf("details = %v, se esperaba la clave fileSize", details)
		}
	})

	t.Run("sin archivo", func(t *testing.T) {
		_, err := newAdminService(newFakeRepository(), newFakeStore()).UploadMinisterioImage(context.Background(), actorID(), nil)
		details := detailsOf(t, err)
		if _, ok := details["file"]; !ok {
			t.Fatalf("details = %v, se esperaba la clave file", details)
		}
	})

	t.Run("fallo de registro: fail-closed", func(t *testing.T) {
		repo := newFakeRepository()
		repo.recordErr = errors.New("no se pudo registrar")
		store := newFakeStore()

		_, err := newAdminService(repo, store).UploadMinisterioImage(context.Background(), actorID(), []byte("jpeg"))
		if err == nil {
			t.Fatal("se esperaba el error del registro")
		}
		if len(store.files) != 0 {
			t.Errorf("el archivo quedó en disco pese al fallo: %v", store.files)
		}
		if len(store.deleted) != 1 {
			t.Errorf("se esperaba la limpieza del archivo: %v", store.deleted)
		}
	})

	t.Run("sin almacén configurado", func(t *testing.T) {
		_, err := newAdminService(newFakeRepository(), nil).UploadMinisterioImage(context.Background(), actorID(), []byte("jpeg"))
		if kind := errorKind(err); kind != apperr.KindInternal {
			t.Fatalf("kind = %v, se esperaba internal (%v)", kind, err)
		}
	})
}

// TestMinisterioAdminDTOCarriesStoredLeader fija que el DTO del panel lleva los
// datos guardados del encargado con el interruptor apagado y el aviso
// `whatsappUrlShared` siempre presente (dato guardado ≠ dato publicado).
func TestMinisterioAdminDTOCarriesStoredLeader(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	current.LeaderNameEs = textOf("Ana Pérez")
	current.LeaderPhone = textOf("+58 412 000 0000")
	repo.seed(current)

	got, err := newAdminService(repo, nil).GetMinisterio(context.Background(), current.ID)
	if err != nil {
		t.Fatalf("GetMinisterio = %v", err)
	}
	raw := mustJSON(t, got)
	for _, key := range []string{
		`"leaderNameEs":"Ana Pérez"`, `"leaderPhone":"+58 412 000 0000"`, `"leaderVisible":false`,
		`"whatsappUrlShared":false`, `"publicationState":"draft"`, `"isPast"`,
	} {
		contains := strings.Contains(raw, key)
		if key == `"isPast"` {
			if contains {
				t.Errorf("el DTO del panel tiene isPast (en este módulo nada vence): %s", raw)
			}
			continue
		}
		if !contains {
			t.Errorf("falta %s en el DTO del panel: %s", key, raw)
		}
	}
}

// boolPtr e intPtr construyen punteros para los PATCH de las pruebas.
func boolPtr(v bool) *bool { return &v }

func intPtr(v int) *int { return &v }
