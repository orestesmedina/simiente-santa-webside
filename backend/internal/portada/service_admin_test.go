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
	requireKind(t, err, apperr.KindInvalid)
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
	requireKind(t, err, apperr.KindInvalid)
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
