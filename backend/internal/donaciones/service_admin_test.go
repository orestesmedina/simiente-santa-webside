package donaciones

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// Tests de la gestión del panel (T715): alta, edición parcial con fusión
// (R7-15), duplicados (P7-10/R7-9), publicación, borrado, listado y los textos
// (singleton por clave) con la regla exacta de códigos de auditoría (R7-10). El
// repositorio es el falso en memoria.

// actorID es el actor de todas las mutaciones de las pruebas.
func actorID() uuid.UUID { return uuid.MustParse("22222222-2222-2222-2222-222222222222") }

// storedMethod construye un método ya guardado del panel (borrador, IBAN).
func storedMethod() MetodoDonacion {
	currency := CurrencyCRC
	return MetodoDonacion{
		ID:               uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Kind:             KindIban,
		Number:           "CR05015202001234567890",
		Institution:      "Banco Nacional",
		HolderEs:         textOf("Ana Pérez"),
		Currency:         &currency,
		PublicationState: StateDraft,
	}
}

// actionsOf devuelve los códigos de las acciones comprometidas con la mutación.
func actionsOf(repo *fakeRepository) []string {
	codes := make([]string, 0, len(repo.committed))
	for _, action := range repo.committed {
		codes = append(codes, action.Code)
	}
	return codes
}

// --- Alta ---

// TestCreateDonacionStoresAndAuditsOnlyCreate fija FR-013/A6-10 homólogo de F7:
// el alta normaliza, guarda con los defaults y registra **solo `donation.create`**
// aunque nazca publicada (cero filas `publish`).
func TestCreateDonacionStoresAndAuditsOnlyCreate(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	created, err := svc.CreateDonacion(context.Background(), actorID(), DonacionInput{
		Kind:             "iban",
		Number:           "  cr05 0152-0200-1234-5678-90  ",
		Institution:      "  Banco Nacional  ",
		HolderEs:         "  Ana Pérez  ",
		Currency:         "CRC",
		PublicationState: "published",
		SortOrder:        3,
	})
	if err != nil {
		t.Fatalf("CreateDonacion = %v", err)
	}
	if created.Number != "CR05015202001234567890" {
		t.Errorf("number = %q, se esperaba normalizado", created.Number)
	}
	if created.Institution != "Banco Nacional" {
		t.Errorf("institution = %q", created.Institution)
	}
	if created.PublicationState != string(StatePublished) {
		t.Errorf("publicationState = %q", created.PublicationState)
	}
	if len(repo.methods) != 1 {
		t.Fatalf("no se guardó el método: %+v", repo.methods)
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionDonationCreate {
		t.Fatalf("acciones = %v, se esperaba solo %s", got, audit.ActionDonationCreate)
	}
	if label := repo.committed[0].TargetLabel; label != "Donaciones · Banco Nacional · CR05015202001234567890" {
		t.Errorf("targetLabel = %q", label)
	}
}

// TestCreateDonacionRejectsInvalidFields fija FR-014: cada campo inválido → `400
// invalid` con `details` por campo y **sin guardar nada**.
func TestCreateDonacionRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		input  DonacionInput
		detail []string
	}{
		{
			name:   "sin tipo",
			input:  DonacionInput{Number: "CR05015202001234567890", Institution: "Banco", Currency: "CRC"},
			detail: []string{"kind"},
		},
		{
			name:   "IBAN mal formado",
			input:  DonacionInput{Kind: "iban", Number: "CR0501520200123456789", Institution: "Banco", Currency: "CRC"},
			detail: []string{"number"},
		},
		{
			name:   "SINPE no celular",
			input:  DonacionInput{Kind: "sinpe", Number: "22222222", Institution: "Banco"},
			detail: []string{"number"},
		},
		{
			name:   "sin entidad",
			input:  DonacionInput{Kind: "iban", Number: "CR05015202001234567890", Currency: "CRC"},
			detail: []string{"institution"},
		},
		{
			name:   "IBAN sin moneda",
			input:  DonacionInput{Kind: "iban", Number: "CR05015202001234567890", Institution: "Banco"},
			detail: []string{"currency"},
		},
		{
			name:   "SINPE con moneda",
			input:  DonacionInput{Kind: "sinpe", Number: "88888888", Institution: "Banco", Currency: "CRC"},
			detail: []string{"currency"},
		},
		{
			name: "titular en inglés sin español",
			input: DonacionInput{
				Kind: "iban", Number: "CR05015202001234567890", Institution: "Banco", Currency: "CRC",
				HolderEn: "Ana Perez",
			},
			detail: []string{"holderEs"},
		},
		{
			name: "entidad de 161 caracteres",
			input: DonacionInput{
				Kind: "iban", Number: "CR05015202001234567890", Institution: strings.Repeat("a", MaxInstitutionLength+1), Currency: "CRC",
			},
			detail: []string{"institution"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			svc := newTestService(repo)
			_, err := svc.CreateDonacion(context.Background(), actorID(), tt.input)
			details := detailsOf(t, err)
			for _, field := range tt.detail {
				if _, ok := details[field]; !ok {
					t.Fatalf("details = %v, se esperaba la clave %q", details, field)
				}
			}
			if len(repo.methods) != 0 {
				t.Errorf("se guardó algo a medias: %+v", repo.methods)
			}
			// Un rechazo deja su fila de auditoría best-effort (`failure`).
			if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
				t.Errorf("rechazo sin registrar: %+v", repo.recorded)
			}
		})
	}
}

// TestCreateDonacionDuplicateConflicts fija P7-10/R7-9: un duplicado exacto por
// tipo + número normalizado → `409 conflict` sin guardar.
func TestCreateDonacionDuplicateConflicts(t *testing.T) {
	repo := newFakeRepository()
	repo.seedMethods(storedMethod())
	svc := newTestService(repo)

	_, err := svc.CreateDonacion(context.Background(), actorID(), DonacionInput{
		Kind: "iban", Number: "cr05 0152-0200-1234-5678-90", Institution: "Otro Banco", Currency: "CRC",
	})
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) || domainErr.Kind != apperr.KindConflict {
		t.Fatalf("err = %v, se esperaba conflict", err)
	}
	if len(repo.methods) != 1 {
		t.Errorf("el duplicado se guardó: %+v", repo.methods)
	}
}

// --- Edición (fusión) ---

// TestUpdateDonacionKeepsNumberDoesNotConflict fija P7-10: editar un método sin
// cambiar su número responde `200` (el chequeo excluye la fila que se edita).
func TestUpdateDonacionKeepsNumberDoesNotConflict(t *testing.T) {
	repo := newFakeRepository()
	repo.seedMethods(storedMethod())
	svc := newTestService(repo)

	updated, err := svc.UpdateDonacion(context.Background(), actorID(), storedMethod().ID, DonacionPatch{
		Institution: OptionalOf("Banco de Costa Rica"),
		SortOrder:   OptionalOf(7),
	})
	if err != nil {
		t.Fatalf("UpdateDonacion = %v", err)
	}
	if updated.Institution != "Banco de Costa Rica" || updated.SortOrder != 7 {
		t.Errorf("updated = %+v", updated)
	}
	// El resto queda intacto.
	if updated.Number != "CR05015202001234567890" || updated.Kind != "iban" || updated.HolderEs == nil || *updated.HolderEs != "Ana Pérez" {
		t.Errorf("campos no presentes cambiaron: %+v", updated)
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionDonationUpdate {
		t.Errorf("acciones = %v, se esperaba donation.update", got)
	}
}

// TestUpdateDonacionNumberToAnotherConflicts fija R7-9: un PATCH que cambia el
// número al de **otro** método → `409 conflict`.
func TestUpdateDonacionNumberToAnotherConflicts(t *testing.T) {
	repo := newFakeRepository()
	other := storedMethod()
	other.ID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	other.Number = "ES9121000418450200051332"
	repo.seedMethods(storedMethod(), other)
	svc := newTestService(repo)

	_, err := svc.UpdateDonacion(context.Background(), actorID(), storedMethod().ID, DonacionPatch{
		Number: OptionalOf("ES9121000418450200051332"),
	})
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) || domainErr.Kind != apperr.KindConflict {
		t.Fatalf("err = %v, se esperaba conflict", err)
	}
}

// TestUpdateDonacionNullClearsOptional fija R7-15: un `null` explícito borra el
// opcional; los campos ausentes quedan intactos.
func TestUpdateDonacionNullClearsOptional(t *testing.T) {
	repo := newFakeRepository()
	repo.seedMethods(storedMethod())
	svc := newTestService(repo)

	updated, err := svc.UpdateDonacion(context.Background(), actorID(), storedMethod().ID, DonacionPatch{
		HolderEs: OptionalNull[string](),
	})
	if err != nil {
		t.Fatalf("UpdateDonacion = %v", err)
	}
	if updated.HolderEs != nil {
		t.Errorf("holderEs = %v, se esperaba nil", updated.HolderEs)
	}
	if updated.Institution != "Banco Nacional" || updated.Number != "CR05015202001234567890" {
		t.Errorf("otros campos cambiaron: %+v", updated)
	}
}

// TestUpdateDonacionAuditsStateChanges fija R7-10: solo estado → una fila
// `publish`/`unpublish`; datos + estado → dos filas (datos primero).
func TestUpdateDonacionAuditsStateChanges(t *testing.T) {
	// Solo estado: publicado.
	onlyState := newFakeRepository()
	onlyState.seedMethods(storedMethod())
	if _, err := newTestService(onlyState).UpdateDonacion(context.Background(), actorID(), storedMethod().ID,
		DonacionPatch{PublicationState: OptionalOf("published")}); err != nil {
		t.Fatalf("UpdateDonacion(solo estado) = %v", err)
	}
	if got := actionsOf(onlyState); len(got) != 1 || got[0] != audit.ActionDonationPublish {
		t.Errorf("acciones (solo estado) = %v, se esperaba donation.publish", got)
	}

	// Datos + estado: dos filas.
	both := newFakeRepository()
	both.seedMethods(storedMethod())
	if _, err := newTestService(both).UpdateDonacion(context.Background(), actorID(), storedMethod().ID,
		DonacionPatch{
			Institution:      OptionalOf("Banco de Costa Rica"),
			PublicationState: OptionalOf("published"),
		}); err != nil {
		t.Fatalf("UpdateDonacion(datos+estado) = %v", err)
	}
	if got := actionsOf(both); len(got) != 2 || got[0] != audit.ActionDonationUpdate || got[1] != audit.ActionDonationPublish {
		t.Errorf("acciones (datos+estado) = %v, se esperaban [donation.update donation.publish]", got)
	}
}

// TestUpdateDonacionNotFound fija el 404 de un id inexistente.
func TestUpdateDonacionNotFound(t *testing.T) {
	repo := newFakeRepository()
	_, err := newTestService(repo).UpdateDonacion(context.Background(), actorID(), uuid.New(), DonacionPatch{
		Institution: OptionalOf("Banco"),
	})
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) || domainErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, se esperaba not_found", err)
	}
}

// TestUpdateDonacionNoChangesReturnsCurrent fija que un PATCH sin campos devuelve
// la fila actual sin escribir ni auditar (el contrato no declara minProperties).
func TestUpdateDonacionNoChangesReturnsCurrent(t *testing.T) {
	repo := newFakeRepository()
	repo.seedMethods(storedMethod())
	svc := newTestService(repo)

	out, err := svc.UpdateDonacion(context.Background(), actorID(), storedMethod().ID, DonacionPatch{})
	if err != nil {
		t.Fatalf("UpdateDonacion = %v", err)
	}
	if out.Institution != "Banco Nacional" {
		t.Errorf("out = %+v", out)
	}
	if len(repo.committed) != 0 {
		t.Errorf("se auditó un cambio inexistente: %+v", repo.committed)
	}
}

// --- Borrado, detalle y listado ---

// TestDeleteDonacionAuditsDelete fija que el borrado registra `donation.delete` y
// un id inexistente es 404.
func TestDeleteDonacionAuditsDelete(t *testing.T) {
	repo := newFakeRepository()
	repo.seedMethods(storedMethod())
	svc := newTestService(repo)

	if err := svc.DeleteDonacion(context.Background(), actorID(), storedMethod().ID); err != nil {
		t.Fatalf("DeleteDonacion = %v", err)
	}
	if len(repo.methods) != 0 {
		t.Errorf("no se borró: %+v", repo.methods)
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionDonationDelete {
		t.Errorf("acciones = %v, se esperaba donation.delete", got)
	}

	err := svc.DeleteDonacion(context.Background(), actorID(), uuid.New())
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) || domainErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, se esperaba not_found", err)
	}
}

// TestGetDonacionAndList fija el detalle (404 incluido) y el listado paginado con
// filtro por estado y el orden del visitante.
func TestGetDonacionAndList(t *testing.T) {
	repo := newFakeRepository()
	published := storedMethod()
	published.ID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	published.PublicationState = StatePublished
	draft := storedMethod()
	draft.ID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	draft.Kind = KindSinpe
	draft.Number = "+50688888888"
	draft.Currency = nil
	repo.seedMethods(draft, published)
	svc := newTestService(repo)

	got, err := svc.GetDonacion(context.Background(), published.ID)
	if err != nil || got.ID != published.ID.String() {
		t.Fatalf("GetDonacion = %+v / %v", got, err)
	}

	list, err := svc.ListDonaciones(context.Background(), nil, paginate.Params{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListDonaciones = %v", err)
	}
	if list.Total != 2 || len(list.Items) != 2 || list.Limit != 10 {
		t.Fatalf("listado = %+v", list)
	}
	// Orden del visitante: iban primero.
	if list.Items[0].Kind != "iban" || list.Items[1].Kind != "sinpe" {
		t.Errorf("orden = %+v", list.Items)
	}

	state := StatePublished
	filtered, err := svc.ListDonaciones(context.Background(), &state, paginate.Params{Limit: 10})
	if err != nil {
		t.Fatalf("ListDonaciones(filtro) = %v", err)
	}
	if filtered.Total != 1 || filtered.Items[0].Kind != "iban" {
		t.Errorf("filtro = %+v", filtered)
	}
}

// --- Textos (singleton por clave, sin borrado) ---

// TestGetTextosReturnsSingletons fija que `intro`/`usage` vienen en `null` hasta
// el primer guardado y con su propio estado.
func TestGetTextosReturnsSingletons(t *testing.T) {
	repo := newFakeRepository()
	repo.seedTexts(TextoDonacion{TextKey: TextKeyIntro, BodyEs: "Cómo donar", PublicationState: StatePublished})
	svc := newTestService(repo)

	out, err := svc.GetTextos(context.Background())
	if err != nil {
		t.Fatalf("GetTextos = %v", err)
	}
	if out.Intro == nil || out.Intro.TextEs != "Cómo donar" || out.Intro.PublicationState != string(StatePublished) {
		t.Errorf("intro = %+v", out.Intro)
	}
	if out.Usage != nil {
		t.Errorf("usage = %+v, se esperaba null", out.Usage)
	}

	empty, err := newTestService(newFakeRepository()).GetTextos(context.Background())
	if err != nil {
		t.Fatalf("GetTextos(vacío) = %v", err)
	}
	if empty.Intro != nil || empty.Usage != nil {
		t.Errorf("textos vacíos = %+v", empty)
	}
}

// TestPutTextoCreatesAndReplaces fija FR-021/P7-17: el primer guardado es
// `donation.text.update` y no duplica el singleton; un estado nuevo registra
// además `text.publish`/`unpublish`; datos + estado → dos filas.
func TestPutTextoCreatesAndReplaces(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	// Primer guardado en borrador: solo text.update.
	if _, err := svc.PutTexto(context.Background(), actorID(), TextKeyIntro, DonacionTextoInput{
		TextEs: "Cómo donar",
	}); err != nil {
		t.Fatalf("PutTexto(create) = %v", err)
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionDonationTextUpdate {
		t.Fatalf("acciones (primer guardado) = %v", got)
	}
	if len(repo.texts) != 1 {
		t.Fatalf("no se creó el singleton: %+v", repo.texts)
	}

	// Reemplazo cambiando datos y estado: dos filas.
	repo.committed = nil
	out, err := svc.PutTexto(context.Background(), actorID(), TextKeyIntro, DonacionTextoInput{
		TextEs:           "Cómo donar 🙌🏽",
		PublicationState: "published",
	})
	if err != nil {
		t.Fatalf("PutTexto(replace) = %v", err)
	}
	if out.TextEs != "Cómo donar 🙌🏽" || out.PublicationState != string(StatePublished) {
		t.Errorf("out = %+v", out)
	}
	if len(repo.texts) != 1 {
		t.Fatalf("el upsert duplicó el singleton: %+v", repo.texts)
	}
	if got := actionsOf(repo); len(got) != 2 || got[0] != audit.ActionDonationTextUpdate || got[1] != audit.ActionDonationTextPublish {
		t.Errorf("acciones (datos+estado) = %v", got)
	}

	// Solo estado (retirar): text.unpublish.
	repo.committed = nil
	if _, err := svc.PutTexto(context.Background(), actorID(), TextKeyIntro, DonacionTextoInput{
		TextEs: "Cómo donar 🙌🏽", PublicationState: "draft",
	}); err != nil {
		t.Fatalf("PutTexto(unpublish) = %v", err)
	}
	if got := actionsOf(repo); len(got) != 1 || got[0] != audit.ActionDonationTextUnpublish {
		t.Errorf("acciones (retirar) = %v", got)
	}
}

// TestPutTextoRejectsInvalid fija el 400 con `details.textKey`/`textEs` y sin
// guardar nada.
func TestPutTextoRejectsInvalid(t *testing.T) {
	// Clave fuera del catálogo.
	repo := newFakeRepository()
	_, err := newTestService(repo).PutTexto(context.Background(), actorID(), TextKey("otro"), DonacionTextoInput{TextEs: "Ok"})
	if _, ok := detailsOf(t, err)["textKey"]; !ok {
		t.Fatalf("se esperaba details.textKey: %v", err)
	}
	if len(repo.texts) != 0 {
		t.Errorf("se guardó con clave inválida: %+v", repo.texts)
	}

	// textEn sin textEs.
	repo2 := newFakeRepository()
	_, err = newTestService(repo2).PutTexto(context.Background(), actorID(), TextKeyUsage, DonacionTextoInput{
		TextEs: "  ", TextEn: "How",
	})
	if _, ok := detailsOf(t, err)["textEs"]; !ok {
		t.Fatalf("se esperaba details.textEs: %v", err)
	}
}

// TestMutationFailureDoesNotRecordSuccess fija el fail-closed en el service: si
// la mutación falla no se deja ninguna fila de éxito (el repository garantiza la
// atomicidad real en T711).
func TestMutationFailureDoesNotRecordSuccess(t *testing.T) {
	repo := newFakeRepository()
	repo.mutationErr = errors.New("boom")
	_, err := newTestService(repo).CreateDonacion(context.Background(), actorID(), DonacionInput{
		Kind: "iban", Number: "CR05015202001234567890", Institution: "Banco", Currency: "CRC",
	})
	if err == nil {
		t.Fatal("se esperaba el error de la mutación")
	}
	for _, action := range repo.committed {
		if action.Result == audit.ResultSuccess {
			t.Errorf("se registró un éxito tras un fallo: %+v", action)
		}
	}
}
