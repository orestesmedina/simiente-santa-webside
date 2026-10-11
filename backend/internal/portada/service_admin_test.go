package portada

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
)

func actionCodes(actions []audit.Action) []string {
	codes := make([]string, 0, len(actions))
	for _, action := range actions {
		codes = append(codes, action.Code)
	}
	return codes
}

func hasCode(actions []audit.Action, code string) bool {
	for _, action := range actions {
		if action.Code == code {
			return true
		}
	}
	return false
}

func validIdentityInput() IdentityInput {
	return IdentityInput{
		NameEs:           "  Iglesia Simiente Santa ",
		NameEn:           "   ",
		TaglineEs:        "Una familia",
		LogoFile:         "img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.jpg",
		LogoAltEs:        "Logotipo de la iglesia",
		PublicationState: string(StatePublished),
	}
}

// T319: guardado de identidad con normalización analyze I6.
func TestSaveIdentityNormalizesAndAudits(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	saved, err := service.SaveIdentity(context.Background(), uuid.New(), validIdentityInput())
	if err != nil {
		t.Fatalf("SaveIdentity = %v", err)
	}
	if saved.NameEs != "Iglesia Simiente Santa" {
		t.Fatalf("nameEs = %q", saved.NameEs)
	}
	if saved.NameEn != nil {
		t.Fatalf("nameEn debería ser nil, fue %q", *saved.NameEn)
	}
	actions := repo.recorded()
	if len(actions) != 1 || actions[0].Code != audit.ActionHomeIdentityUpdate {
		t.Fatalf("acciones = %v", actionCodes(actions))
	}
	if actions[0].TargetKind != audit.TargetContent || actions[0].TargetLabel == "" {
		t.Fatalf("objetivo mal formado: %+v", actions[0])
	}
}

// T319: cambiar el estado deja DOS filas (analyze M5).
func TestSaveIdentityStateChangeTwoRows(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{ID: uuid.New(), NameEs: "Vieja", PublicationState: StatePublished}
	service := NewService(ServiceDeps{Repository: repo})

	in := validIdentityInput()
	in.PublicationState = string(StateDraft)
	if _, err := service.SaveIdentity(context.Background(), uuid.New(), in); err != nil {
		t.Fatalf("SaveIdentity = %v", err)
	}
	actions := repo.recorded()
	if len(actions) != 2 || !hasCode(actions, audit.ActionHomeIdentityUpdate) || !hasCode(actions, audit.ActionHomeUnpublish) {
		t.Fatalf("se esperaban update+unpublish, fueron %v", actionCodes(actions))
	}
}

// T319: una imagen exige texto alternativo en español (FR-019) y no guarda nada.
func TestSaveIdentityRequiresImageAlt(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	in := validIdentityInput()
	in.LogoAltEs = "   "
	_, err := service.SaveIdentity(context.Background(), uuid.New(), in)
	_ = requireKind(t, err, apperr.KindInvalid)
	requireDetail(t, err, "logoAltEs")
	if repo.identity != nil {
		t.Fatal("no debía guardarse nada")
	}
}

// T319: si el registro falla, la mutación no se aplica (edge case FR-017).
func TestSaveIdentityRollsBackOnAuditFailure(t *testing.T) {
	repo := newFakeRepository()
	repo.actionErr = apperr.Internal(nil)
	service := NewService(ServiceDeps{Repository: repo})

	if _, err := service.SaveIdentity(context.Background(), uuid.New(), validIdentityInput()); err == nil {
		t.Fatal("se esperaba error")
	}
	if repo.identity != nil {
		t.Fatal("la mutación no debía aplicarse")
	}
}

// T319: al reemplazar una imagen se borra la anterior (R3-8).
func TestSaveIdentityDeletesReplacedImage(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{
		ID:               uuid.New(),
		NameEs:           "Vieja",
		LogoFile:         strptr("img_bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb.jpg"),
		LogoAltEs:        strptr("Viejo alt"),
		PublicationState: StatePublished,
	}
	store := &fakeStore{files: map[string][]byte{}}
	service := NewService(ServiceDeps{Repository: repo, Store: store})

	if _, err := service.SaveIdentity(context.Background(), uuid.New(), validIdentityInput()); err != nil {
		t.Fatalf("SaveIdentity = %v", err)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "img_bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb.jpg" {
		t.Fatalf("borrados = %v", store.deleted)
	}
}

// T319: «quiénes somos» valida y audita.
func TestSaveAbout(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	if _, err := service.SaveAbout(context.Background(), uuid.New(), AboutInput{TextEs: "   "}); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("texto vacío debería ser invalid: %v", err)
	}

	saved, err := service.SaveAbout(context.Background(), uuid.New(), AboutInput{TextEs: "  Somos  iglesia ", TextEn: " ", PublicationState: string(StateDraft)})
	if err != nil {
		t.Fatalf("SaveAbout = %v", err)
	}
	if saved.TextEs != "Somos iglesia" || saved.TextEn != nil {
		t.Fatalf("normalización: %+v", saved)
	}
	if !hasCode(repo.recorded(), audit.ActionHomeAboutUpdate) {
		t.Fatalf("falta home.about.update: %v", actionCodes(repo.recorded()))
	}
}

// T319: el contacto valida correo/teléfono y normaliza el teléfono.
func TestSaveContact(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	_, err := service.SaveContact(context.Background(), uuid.New(), ContactInput{
		AddressEs: "Calle 1", Email: "no-es-correo", Phone: "0412-1234567", PublicationState: string(StateDraft),
	})
	_ = requireKind(t, err, apperr.KindInvalid)
	requireDetail(t, err, "email")

	saved, err := service.SaveContact(context.Background(), uuid.New(), ContactInput{
		AddressEs: " Calle  1 ", Email: " hola@simiente.org ", Phone: "(0412) 123-4567", PublicationState: string(StatePublished),
	})
	if err != nil {
		t.Fatalf("SaveContact = %v", err)
	}
	if saved.Phone != "04121234567" || saved.Email != "hola@simiente.org" {
		t.Fatalf("normalización del contacto: %+v", saved)
	}
}

// T320: el alta publicada registra `create` y nunca `publish` (analyze M5).
func TestCreateServicePublishedRegistersCreateOnly(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	saved, err := service.CreateService(context.Background(), uuid.New(), ScheduleItemInput{
		DayOfWeek: 0, StartTime: "10:00", EndTime: "12:00",
		NameEs: "Culto dominical", PlaceEs: "Templo", PublicationState: string(StatePublished),
	})
	if err != nil {
		t.Fatalf("CreateService = %v", err)
	}
	if saved.EndTime == nil || *saved.EndTime != "12:00" {
		t.Fatalf("endTime = %v", saved.EndTime)
	}
	actions := repo.recorded()
	if len(actions) != 1 || actions[0].Code != audit.ActionHomeScheduleCreate {
		t.Fatalf("acciones = %v", actionCodes(actions))
	}
	if actions[0].TargetLabel != "Portada · Horario · Culto dominical" {
		t.Fatalf("targetLabel = %q", actions[0].TargetLabel)
	}
}

// T320: fin anterior al inicio → 400 details.endTime.
func TestCreateServiceEndBeforeStart(t *testing.T) {
	service := NewService(ServiceDeps{Repository: newFakeRepository()})
	_, err := service.CreateService(context.Background(), uuid.New(), ScheduleItemInput{
		StartTime: "12:00", EndTime: "10:00", NameEs: "Culto", PlaceEs: "Templo",
	})
	_ = requireKind(t, err, apperr.KindInvalid)
	requireDetail(t, err, "endTime")
}

// T320: límite de colección (≤50).
func TestCreateServiceLimit(t *testing.T) {
	repo := newFakeRepository()
	for i := 0; i < MaxScheduleItems; i++ {
		repo.seedService(Service{NameEs: "Culto", PlaceEs: "Templo", StartTime: "10:00", PublicationState: StateDraft})
	}
	service := NewService(ServiceDeps{Repository: repo})
	_, err := service.CreateService(context.Background(), uuid.New(), ScheduleItemInput{
		StartTime: "10:00", NameEs: "Nuevo", PlaceEs: "Templo",
	})
	_ = requireKind(t, err, apperr.KindInvalid)
}

// T320: PATCH que cambia datos y estado deja dos filas (analyze M5).
func TestUpdateServiceDataAndStateTwoRows(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedService(Service{NameEs: "Culto", PlaceEs: "Templo", StartTime: "10:00", PublicationState: StatePublished})
	service := NewService(ServiceDeps{Repository: repo})

	name := "Culto nuevo"
	state := string(StateDraft)
	if _, err := service.UpdateService(context.Background(), uuid.New(), current.ID, ScheduleItemPatch{NameEs: &name, PublicationState: &state}); err != nil {
		t.Fatalf("UpdateService = %v", err)
	}
	actions := repo.recorded()
	if len(actions) != 2 || !hasCode(actions, audit.ActionHomeScheduleUpdate) || !hasCode(actions, audit.ActionHomeUnpublish) {
		t.Fatalf("se esperaban update+unpublish, fueron %v", actionCodes(actions))
	}
}

// T320: un PATCH que solo cambia el estado registra solo publish/unpublish.
func TestUpdateServiceStateOnly(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedService(Service{NameEs: "Culto", PlaceEs: "Templo", StartTime: "10:00", PublicationState: StateDraft})
	service := NewService(ServiceDeps{Repository: repo})

	state := string(StatePublished)
	if _, err := service.UpdateService(context.Background(), uuid.New(), current.ID, ScheduleItemPatch{PublicationState: &state}); err != nil {
		t.Fatalf("UpdateService = %v", err)
	}
	actions := repo.recorded()
	if len(actions) != 1 || actions[0].Code != audit.ActionHomePublish {
		t.Fatalf("acciones = %v", actionCodes(actions))
	}
}

// T320: id inexistente → 404.
func TestUpdateServiceNotFound(t *testing.T) {
	service := NewService(ServiceDeps{Repository: newFakeRepository()})
	name := "X"
	_, err := service.UpdateService(context.Background(), uuid.New(), uuid.New(), ScheduleItemPatch{NameEs: &name})
	_ = requireKind(t, err, apperr.KindNotFound)
}

// T320: borrado conserva el nombre del elemento en la auditoría.
func TestDeleteServiceKeepsLabel(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedService(Service{NameEs: "Culto dominical", PlaceEs: "Templo", StartTime: "10:00"})
	service := NewService(ServiceDeps{Repository: repo})

	if err := service.DeleteService(context.Background(), uuid.New(), current.ID); err != nil {
		t.Fatalf("DeleteService = %v", err)
	}
	actions := repo.recorded()
	if len(actions) != 1 || actions[0].Code != audit.ActionHomeScheduleDelete {
		t.Fatalf("acciones = %v", actionCodes(actions))
	}
	if actions[0].TargetLabel != "Portada · Horario · Culto dominical" {
		t.Fatalf("targetLabel = %q", actions[0].TargetLabel)
	}
}

// T320: canal directo normalizado + duplicado exacto → 409.
func TestCreateWhatsappDirectAndDuplicate(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	in := WhatsappChannelInput{NameEs: "Canal", Kind: KindDirect, Destination: " +58 412-1234567 ", PublicationState: string(StatePublished)}
	saved, err := service.CreateWhatsappChannel(context.Background(), uuid.New(), in)
	if err != nil {
		t.Fatalf("CreateWhatsappChannel = %v", err)
	}
	if saved.Destination != "+584121234567" {
		t.Fatalf("destination = %q", saved.Destination)
	}
	if !hasCode(repo.recorded(), audit.ActionHomeWhatsappCreate) {
		t.Fatalf("falta home.whatsapp.create: %v", actionCodes(repo.recorded()))
	}

	_, err = service.CreateWhatsappChannel(context.Background(), uuid.New(), in)
	_ = requireKind(t, err, apperr.KindConflict)
}

// T320: grupo con host ajeno → 400 details.destination.
func TestCreateWhatsappGroupHost(t *testing.T) {
	service := NewService(ServiceDeps{Repository: newFakeRepository()})
	_, err := service.CreateWhatsappChannel(context.Background(), uuid.New(), WhatsappChannelInput{
		NameEs: "Grupo", Kind: KindGroup, Destination: "https://example.com/grupo",
	})
	_ = requireKind(t, err, apperr.KindInvalid)
	requireDetail(t, err, "destination")
}

// T320: límite de canales (≤20).
func TestCreateWhatsappLimit(t *testing.T) {
	repo := newFakeRepository()
	for i := 0; i < MaxWhatsappChannels; i++ {
		repo.seedChannel(WhatsappChannel{NameEs: "Canal", Kind: KindDirect, Destination: "0412000000", PublicationState: StateDraft})
	}
	service := NewService(ServiceDeps{Repository: repo})
	_, err := service.CreateWhatsappChannel(context.Background(), uuid.New(), WhatsappChannelInput{
		NameEs: "Nuevo", Kind: KindDirect, Destination: "04121111111",
	})
	_ = requireKind(t, err, apperr.KindInvalid)
}

// T320: redes: red fuera de catálogo, host incorrecto y duplicado.
func TestCreateSocialLinkValidation(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	_, err := service.CreateSocialLink(context.Background(), uuid.New(), SocialLinkInput{Network: "x", URL: "https://x.com/simiente"})
	_ = requireKind(t, err, apperr.KindInvalid)
	requireDetail(t, err, "network")

	_, err = service.CreateSocialLink(context.Background(), uuid.New(), SocialLinkInput{Network: "facebook", URL: "https://instagram.com/simiente"})
	_ = requireKind(t, err, apperr.KindInvalid)
	requireDetail(t, err, "url")

	if _, err := service.CreateSocialLink(context.Background(), uuid.New(), SocialLinkInput{Network: "facebook", URL: "https://facebook.com/simiente", PublicationState: string(StatePublished)}); err != nil {
		t.Fatalf("CreateSocialLink = %v", err)
	}
	_, err = service.CreateSocialLink(context.Background(), uuid.New(), SocialLinkInput{Network: "facebook", URL: "https://facebook.com/otro"})
	_ = requireKind(t, err, apperr.KindConflict)
}

// T320: editar el enlace revalida red y URL.
func TestUpdateSocialLink(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedSocialLink(SocialLink{Network: "facebook", URL: "https://facebook.com/simiente", PublicationState: StatePublished})
	service := NewService(ServiceDeps{Repository: repo})

	network := "instagram"
	url := "https://instagram.com/simiente"
	if _, err := service.UpdateSocialLink(context.Background(), uuid.New(), current.ID, SocialLinkPatch{Network: &network, URL: &url}); err != nil {
		t.Fatalf("UpdateSocialLink = %v", err)
	}
	if !hasCode(repo.recorded(), audit.ActionHomeSocialUpdate) {
		t.Fatalf("falta home.social.update: %v", actionCodes(repo.recorded()))
	}
}

// T320: PATCH de canal: datos + estado (dos filas) y revalidación del destino.
func TestUpdateWhatsappChannel(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedChannel(WhatsappChannel{NameEs: "Canal", Kind: KindDirect, Destination: "04121234567", PublicationState: StateDraft})
	other := repo.seedChannel(WhatsappChannel{NameEs: "Otro", Kind: KindDirect, Destination: "04129999999", PublicationState: StateDraft})
	service := NewService(ServiceDeps{Repository: repo})

	name := "Canal editado"
	state := string(StatePublished)
	if _, err := service.UpdateWhatsappChannel(context.Background(), uuid.New(), current.ID, WhatsappChannelPatch{NameEs: &name, PublicationState: &state}); err != nil {
		t.Fatalf("UpdateWhatsappChannel = %v", err)
	}
	actions := repo.recorded()
	if len(actions) != 2 || !hasCode(actions, audit.ActionHomeWhatsappUpdate) || !hasCode(actions, audit.ActionHomePublish) {
		t.Fatalf("se esperaban update+publish, fueron %v", actionCodes(actions))
	}

	// Duplicado exacto al editar → 409.
	otherName := name
	otherDestination := "04121234567"
	kind := KindDirect
	_, err := service.UpdateWhatsappChannel(context.Background(), uuid.New(), other.ID, WhatsappChannelPatch{
		NameEs: &otherName, Destination: &otherDestination, Kind: &kind,
	})
	_ = requireKind(t, err, apperr.KindConflict)

	// Sin cambios → 400.
	if _, err := service.UpdateWhatsappChannel(context.Background(), uuid.New(), current.ID, WhatsappChannelPatch{}); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("patch vacío debería ser invalid: %v", err)
	}

	// Id inexistente → 404.
	if _, err := service.UpdateWhatsappChannel(context.Background(), uuid.New(), uuid.New(), WhatsappChannelPatch{NameEs: &name}); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("id inexistente debería ser not_found: %v", err)
	}

	// Cambio de tipo con destino inválido → 400.
	group := KindGroup
	badDest := "https://example.com/g"
	if _, err := service.UpdateWhatsappChannel(context.Background(), uuid.New(), current.ID, WhatsappChannelPatch{Kind: &group, Destination: &badDest}); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("destino incoherente debería ser invalid: %v", err)
	}
}

// T320: borrado de canal y enlace con su auditoría.
func TestDeleteWhatsappChannelAndSocialLink(t *testing.T) {
	repo := newFakeRepository()
	channel := repo.seedChannel(WhatsappChannel{NameEs: "Canal", Kind: KindDirect, Destination: "04121234567"})
	link := repo.seedSocialLink(SocialLink{Network: "facebook", URL: "https://facebook.com/x"})
	service := NewService(ServiceDeps{Repository: repo})

	if err := service.DeleteWhatsappChannel(context.Background(), uuid.New(), channel.ID); err != nil {
		t.Fatalf("DeleteWhatsappChannel = %v", err)
	}
	if err := service.DeleteSocialLink(context.Background(), uuid.New(), link.ID); err != nil {
		t.Fatalf("DeleteSocialLink = %v", err)
	}
	if !hasCode(repo.recorded(), audit.ActionHomeWhatsappDelete) || !hasCode(repo.recorded(), audit.ActionHomeSocialDelete) {
		t.Fatalf("faltan borrados: %v", actionCodes(repo.recorded()))
	}
	if err := service.DeleteWhatsappChannel(context.Background(), uuid.New(), channel.ID); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("segundo borrado debería ser not_found: %v", err)
	}
	if err := service.DeleteSocialLink(context.Background(), uuid.New(), link.ID); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("segundo borrado debería ser not_found: %v", err)
	}
	if err := service.DeleteService(context.Background(), uuid.New(), uuid.New()); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("servicio inexistente debería ser not_found: %v", err)
	}
}

// T320: variantes del merge de servicios (campos opcionales, fin y día).
func TestUpdateServiceMergeVariants(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedService(Service{NameEs: "Culto", PlaceEs: "Templo", StartTime: "10:00", EndTime: strptr("12:00"), PublicationState: StateDraft})
	service := NewService(ServiceDeps{Repository: repo})

	empty := ""
	placeEn := "Temple"
	descriptionEs := "  descripción  "
	sortOrder := 5
	saved, err := service.UpdateService(context.Background(), uuid.New(), current.ID, ScheduleItemPatch{
		EndTime: OptionalOf(empty), PlaceEn: OptionalOf(placeEn), DescriptionEs: OptionalOf(descriptionEs), SortOrder: &sortOrder,
	})
	if err != nil {
		t.Fatalf("UpdateService = %v", err)
	}
	if saved.EndTime != nil || saved.PlaceEn == nil || *saved.PlaceEn != "Temple" || saved.SortOrder != 5 {
		t.Fatalf("merge incorrecto: %+v", saved)
	}

	// Día fuera de rango → 400.
	badDay := 9
	if _, err := service.UpdateService(context.Background(), uuid.New(), current.ID, ScheduleItemPatch{DayOfWeek: &badDay}); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("día inválido debería ser invalid: %v", err)
	}

	// Estado igual al actual y sin más cambios → no deja fila.
	state := string(StateDraft)
	before := len(repo.recorded())
	if _, err := service.UpdateService(context.Background(), uuid.New(), current.ID, ScheduleItemPatch{PublicationState: &state}); err != nil {
		t.Fatalf("UpdateService = %v", err)
	}
	if len(repo.recorded()) != before {
		t.Fatal("un PATCH sin cambio efectivo no debe dejar fila")
	}

	// Nombre en español vacío → 400.
	emptyname := "   "
	if _, err := service.UpdateService(context.Background(), uuid.New(), current.ID, ScheduleItemPatch{NameEs: &emptyname}); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("nombre vacío debería ser invalid: %v", err)
	}
}

// B1 (corrección): limpiar un campo opcional con `null` es un cambio de datos y
// deja su fila `home.schedule.update` (analyze M5: solo datos). Decodifica el
// cuerpo JSON real del panel para ejercitar la presencia del campo.
func TestPatchScheduleNullClearsAndAudits(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedService(Service{
		NameEs: "Culto", PlaceEs: "Templo", StartTime: "10:00",
		EndTime: strptr("12:00"), PublicationState: StatePublished,
	})
	service := NewService(ServiceDeps{Repository: repo})

	patch := decodePatch[ScheduleItemPatch](t, `{"endTime": null}`)
	saved, err := service.UpdateService(context.Background(), uuid.New(), current.ID, patch)
	if err != nil {
		t.Fatalf("UpdateService = %v", err)
	}
	if saved.EndTime != nil {
		t.Fatalf("endTime: null debe quitar la hora de fin; quedó %v", *saved.EndTime)
	}
	actions := repo.recorded()
	if len(actions) != 1 || actions[0].Code != audit.ActionHomeScheduleUpdate {
		t.Fatalf("limpiar la hora de fin debe dejar solo home.schedule.update: %v", actionCodes(actions))
	}
}

// T319/T320: el cambio de estado de about/contact deja dos filas.
func TestSaveAboutAndContactStateChange(t *testing.T) {
	repo := newFakeRepository()
	repo.about = &About{ID: uuid.New(), TextEs: "Viejo", PublicationState: StatePublished}
	repo.contact = &Contact{ID: uuid.New(), AddressEs: "Calle", Email: "a@b.com", Phone: "04121234567", PublicationState: StatePublished}
	service := NewService(ServiceDeps{Repository: repo})

	if _, err := service.SaveAbout(context.Background(), uuid.New(), AboutInput{TextEs: "Nuevo", PublicationState: string(StateDraft)}); err != nil {
		t.Fatalf("SaveAbout = %v", err)
	}
	if _, err := service.SaveContact(context.Background(), uuid.New(), ContactInput{AddressEs: "Calle", Email: "a@b.com", Phone: "04121234567", PublicationState: string(StateDraft)}); err != nil {
		t.Fatalf("SaveContact = %v", err)
	}
	actions := repo.recorded()
	if len(actions) != 4 {
		t.Fatalf("se esperaban 4 filas, fueron %v", actionCodes(actions))
	}
}

// T320: update de red sin cambios efectivos y cambio de estado.
func TestUpdateSocialLinkStateOnly(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedSocialLink(SocialLink{Network: "facebook", URL: "https://facebook.com/x", PublicationState: StateDraft})
	service := NewService(ServiceDeps{Repository: repo})

	state := string(StatePublished)
	if _, err := service.UpdateSocialLink(context.Background(), uuid.New(), current.ID, SocialLinkPatch{PublicationState: &state}); err != nil {
		t.Fatalf("UpdateSocialLink = %v", err)
	}
	if len(repo.recorded()) != 1 || repo.recorded()[0].Code != audit.ActionHomePublish {
		t.Fatalf("acciones = %v", actionCodes(repo.recorded()))
	}
	if _, err := service.UpdateSocialLink(context.Background(), uuid.New(), current.ID, SocialLinkPatch{}); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("patch vacío debería ser invalid: %v", err)
	}
	if _, err := service.UpdateSocialLink(context.Background(), uuid.New(), uuid.New(), SocialLinkPatch{URL: strptr("https://facebook.com/y")}); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("id inexistente debería ser not_found: %v", err)
	}
}

// T340: agregado del panel GET /api/v1/admin/portada (service).
func TestGetPortadaAdminEmpty(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	got, err := service.GetPortadaAdmin(context.Background())
	if err != nil {
		t.Fatalf("GetPortadaAdmin = %v", err)
	}
	if got.Identity != nil || got.About != nil || got.Contact != nil {
		t.Errorf("los singletons deben ser null sin datos: %+v", got)
	}
	if got.Schedule.Items == nil || got.Whatsapp.Items == nil || got.Socials.Items == nil {
		t.Fatalf("las colecciones nunca deben ser null: %+v", got)
	}
	if len(got.Schedule.Items)+len(got.Whatsapp.Items)+len(got.Socials.Items) != 0 {
		t.Errorf("colecciones inesperadas: %+v", got)
	}
}

func TestGetPortadaAdminWithData(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{NameEs: "Iglesia", NameEn: strptr("Church"), PublicationState: StateDraft}
	repo.about = &About{TextEs: "Texto", PublicationState: StatePublished}
	repo.contact = &Contact{AddressEs: "Calle 1", Email: "a@b.com", Phone: "+584121234567", PublicationState: StateDraft}
	repo.seedService(Service{DayOfWeek: 0, StartTime: "10:00", NameEs: "Culto", PlaceEs: "Sede", PublicationState: StateDraft})
	repo.seedChannel(WhatsappChannel{Kind: KindDirect, Destination: "+584121234567", NameEs: "General", PublicationState: StateDraft})
	repo.seedSocialLink(SocialLink{Network: "facebook", URL: "https://facebook.com/x", PublicationState: StateDraft})
	service := NewService(ServiceDeps{Repository: repo})

	got, err := service.GetPortadaAdmin(context.Background())
	if err != nil {
		t.Fatalf("GetPortadaAdmin = %v", err)
	}
	if got.Identity == nil || got.About == nil || got.Contact == nil {
		t.Fatalf("faltan singletons: %+v", got)
	}
	// Los borradores se ven en el panel (analyze C1) y los pares Es/En crudos.
	if got.Identity.PublicationState != string(StateDraft) {
		t.Errorf("publicationState = %q, se esperaba draft", got.Identity.PublicationState)
	}
	if got.Identity.NameEn == nil || *got.Identity.NameEn != "Church" {
		t.Errorf("nameEn = %v", got.Identity.NameEn)
	}
	if got.Contact.PublicationState != string(StateDraft) {
		t.Errorf("contact publicationState = %q", got.Contact.PublicationState)
	}
	if len(got.Schedule.Items) != 1 || got.Schedule.Items[0].PublicationState != string(StateDraft) {
		t.Errorf("schedule = %+v", got.Schedule.Items)
	}
	if len(got.Whatsapp.Items) != 1 || len(got.Socials.Items) != 1 {
		t.Errorf("colecciones = %+v / %+v", got.Whatsapp.Items, got.Socials.Items)
	}
}
