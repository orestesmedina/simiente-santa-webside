package portada

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
)

// --- Fake del repositorio (compartido por los tests del service) ---

// fakeRepository implementa el puerto Repository en memoria: guarda los
// singletons, las colecciones y las acciones de auditoría, y acumula las que
// recibe en cada mutación (las que el repository real insertaría en la misma
// transacción).
type fakeRepository struct {
	mu sync.Mutex

	identity *Identity
	about    *About
	contact  *Contact

	services []Service
	channels []WhatsappChannel
	socials  []SocialLink

	publishedFiles map[string]bool
	actions        []audit.Action

	getErr      error
	mutationErr error
	actionErr   error
	recordErr   error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{publishedFiles: map[string]bool{}}
}

func (f *fakeRepository) record(actions []audit.Action) {
	f.actions = append(f.actions, actions...)
}

func (f *fakeRepository) recorded() []audit.Action {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]audit.Action, len(f.actions))
	copy(out, f.actions)
	return out
}

// --- Singletons ---

func (f *fakeRepository) GetHomeIdentity(context.Context) (Identity, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return Identity{}, false, f.getErr
	}
	if f.identity == nil {
		return Identity{}, false, nil
	}
	return *f.identity, true, nil
}

func (f *fakeRepository) GetHomeIdentityPublished(ctx context.Context) (Identity, bool, error) {
	identity, found, err := f.GetHomeIdentity(ctx)
	if err != nil || !found || identity.PublicationState != StatePublished {
		return Identity{}, false, err
	}
	return identity, true, nil
}

func (f *fakeRepository) UpsertHomeIdentity(_ context.Context, in Identity, actions ...audit.Action) (Identity, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return Identity{}, f.mutationErr
	}
	if f.actionErr != nil {
		return Identity{}, f.actionErr
	}
	if f.identity != nil {
		in.ID = f.identity.ID
		in.CreatedAt = f.identity.CreatedAt
	} else if in.ID == uuid.Nil {
		in.ID = uuid.New()
	}
	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}
	in.UpdatedAt = time.Now()
	f.identity = &in
	f.record(actions)
	return in, nil
}

func (f *fakeRepository) GetHomeAbout(context.Context) (About, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return About{}, false, f.getErr
	}
	if f.about == nil {
		return About{}, false, nil
	}
	return *f.about, true, nil
}

func (f *fakeRepository) GetHomeAboutPublished(ctx context.Context) (About, bool, error) {
	about, found, err := f.GetHomeAbout(ctx)
	if err != nil || !found || about.PublicationState != StatePublished {
		return About{}, false, err
	}
	return about, true, nil
}

func (f *fakeRepository) UpsertHomeAbout(_ context.Context, in About, actions ...audit.Action) (About, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return About{}, f.mutationErr
	}
	if f.actionErr != nil {
		return About{}, f.actionErr
	}
	if f.about != nil {
		in.ID = f.about.ID
		in.CreatedAt = f.about.CreatedAt
	} else if in.ID == uuid.Nil {
		in.ID = uuid.New()
	}
	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}
	in.UpdatedAt = time.Now()
	f.about = &in
	f.record(actions)
	return in, nil
}

func (f *fakeRepository) GetHomeContact(context.Context) (Contact, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return Contact{}, false, f.getErr
	}
	if f.contact == nil {
		return Contact{}, false, nil
	}
	return *f.contact, true, nil
}

func (f *fakeRepository) GetHomeContactPublished(ctx context.Context) (Contact, bool, error) {
	contact, found, err := f.GetHomeContact(ctx)
	if err != nil || !found || contact.PublicationState != StatePublished {
		return Contact{}, false, err
	}
	return contact, true, nil
}

func (f *fakeRepository) UpsertHomeContact(_ context.Context, in Contact, actions ...audit.Action) (Contact, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return Contact{}, f.mutationErr
	}
	if f.actionErr != nil {
		return Contact{}, f.actionErr
	}
	if f.contact != nil {
		in.ID = f.contact.ID
		in.CreatedAt = f.contact.CreatedAt
	} else if in.ID == uuid.Nil {
		in.ID = uuid.New()
	}
	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}
	in.UpdatedAt = time.Now()
	f.contact = &in
	f.record(actions)
	return in, nil
}

// --- Archivos ---

func (f *fakeRepository) IsHomeFilePublished(_ context.Context, fileName string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return false, f.getErr
	}
	return f.publishedFiles[fileName], nil
}

// --- Horario ---

func (f *fakeRepository) InsertHomeService(_ context.Context, in Service, actions ...audit.Action) (Service, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return Service{}, f.mutationErr
	}
	if f.actionErr != nil {
		return Service{}, f.actionErr
	}
	in.ID = uuid.New()
	in.CreatedAt = time.Now()
	in.UpdatedAt = in.CreatedAt
	f.services = append(f.services, in)
	f.record(actions)
	return in, nil
}

func (f *fakeRepository) UpdateHomeService(_ context.Context, in Service, actions ...audit.Action) (Service, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return Service{}, f.mutationErr
	}
	if f.actionErr != nil {
		return Service{}, f.actionErr
	}
	for i := range f.services {
		if f.services[i].ID == in.ID {
			in.CreatedAt = f.services[i].CreatedAt
			in.UpdatedAt = time.Now()
			f.services[i] = in
			f.record(actions)
			return in, nil
		}
	}
	return Service{}, apperr.NotFound("El servicio no existe")
}

func (f *fakeRepository) DeleteHomeService(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	if f.actionErr != nil {
		return false, f.actionErr
	}
	for i := range f.services {
		if f.services[i].ID == id {
			f.services = append(f.services[:i], f.services[i+1:]...)
			f.record(actions)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepository) GetHomeServiceByID(_ context.Context, id uuid.UUID) (Service, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return Service{}, false, f.getErr
	}
	for _, service := range f.services {
		if service.ID == id {
			return service, true, nil
		}
	}
	return Service{}, false, nil
}

func (f *fakeRepository) ListHomeServices(context.Context) ([]Service, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	return append([]Service(nil), f.services...), nil
}

func (f *fakeRepository) ListHomeServicesPublished(context.Context) ([]Service, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	var out []Service
	for _, service := range f.services {
		if service.PublicationState == StatePublished {
			out = append(out, service)
		}
	}
	return out, nil
}

// --- WhatsApp ---

func (f *fakeRepository) InsertHomeWhatsappChannel(_ context.Context, in WhatsappChannel, actions ...audit.Action) (WhatsappChannel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return WhatsappChannel{}, f.mutationErr
	}
	if f.actionErr != nil {
		return WhatsappChannel{}, f.actionErr
	}
	for _, existing := range f.channels {
		if existing.Kind == in.Kind && existing.Destination == in.Destination && existing.NameEs == in.NameEs {
			return WhatsappChannel{}, apperr.Conflict(conflictWhatsappDuplicate)
		}
	}
	in.ID = uuid.New()
	in.CreatedAt = time.Now()
	in.UpdatedAt = in.CreatedAt
	f.channels = append(f.channels, in)
	f.record(actions)
	return in, nil
}

func (f *fakeRepository) UpdateHomeWhatsappChannel(_ context.Context, in WhatsappChannel, actions ...audit.Action) (WhatsappChannel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return WhatsappChannel{}, f.mutationErr
	}
	if f.actionErr != nil {
		return WhatsappChannel{}, f.actionErr
	}
	for _, existing := range f.channels {
		if existing.ID != in.ID && existing.Kind == in.Kind &&
			existing.Destination == in.Destination && existing.NameEs == in.NameEs {
			return WhatsappChannel{}, apperr.Conflict(conflictWhatsappDuplicate)
		}
	}
	for i := range f.channels {
		if f.channels[i].ID == in.ID {
			in.CreatedAt = f.channels[i].CreatedAt
			in.UpdatedAt = time.Now()
			f.channels[i] = in
			f.record(actions)
			return in, nil
		}
	}
	return WhatsappChannel{}, apperr.NotFound("El canal no existe")
}

func (f *fakeRepository) DeleteHomeWhatsappChannel(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	if f.actionErr != nil {
		return false, f.actionErr
	}
	for i := range f.channels {
		if f.channels[i].ID == id {
			f.channels = append(f.channels[:i], f.channels[i+1:]...)
			f.record(actions)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepository) GetHomeWhatsappChannelByID(_ context.Context, id uuid.UUID) (WhatsappChannel, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return WhatsappChannel{}, false, f.getErr
	}
	for _, channel := range f.channels {
		if channel.ID == id {
			return channel, true, nil
		}
	}
	return WhatsappChannel{}, false, nil
}

func (f *fakeRepository) ListHomeWhatsappChannels(context.Context) ([]WhatsappChannel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	return append([]WhatsappChannel(nil), f.channels...), nil
}

func (f *fakeRepository) ListHomeWhatsappChannelsPublished(context.Context) ([]WhatsappChannel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	var out []WhatsappChannel
	for _, channel := range f.channels {
		if channel.PublicationState == StatePublished {
			out = append(out, channel)
		}
	}
	return out, nil
}

// --- Redes ---

func (f *fakeRepository) InsertHomeSocialLink(_ context.Context, in SocialLink, actions ...audit.Action) (SocialLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return SocialLink{}, f.mutationErr
	}
	if f.actionErr != nil {
		return SocialLink{}, f.actionErr
	}
	for _, existing := range f.socials {
		if existing.Network == in.Network {
			return SocialLink{}, apperr.Conflict(conflictSocialDuplicate)
		}
	}
	in.ID = uuid.New()
	in.CreatedAt = time.Now()
	in.UpdatedAt = in.CreatedAt
	f.socials = append(f.socials, in)
	f.record(actions)
	return in, nil
}

func (f *fakeRepository) UpdateHomeSocialLink(_ context.Context, in SocialLink, actions ...audit.Action) (SocialLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return SocialLink{}, f.mutationErr
	}
	if f.actionErr != nil {
		return SocialLink{}, f.actionErr
	}
	for _, existing := range f.socials {
		if existing.ID != in.ID && existing.Network == in.Network {
			return SocialLink{}, apperr.Conflict(conflictSocialDuplicate)
		}
	}
	for i := range f.socials {
		if f.socials[i].ID == in.ID {
			in.CreatedAt = f.socials[i].CreatedAt
			in.UpdatedAt = time.Now()
			f.socials[i] = in
			f.record(actions)
			return in, nil
		}
	}
	return SocialLink{}, apperr.NotFound("El enlace no existe")
}

func (f *fakeRepository) DeleteHomeSocialLink(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	if f.actionErr != nil {
		return false, f.actionErr
	}
	for i := range f.socials {
		if f.socials[i].ID == id {
			f.socials = append(f.socials[:i], f.socials[i+1:]...)
			f.record(actions)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepository) GetHomeSocialLinkByID(_ context.Context, id uuid.UUID) (SocialLink, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return SocialLink{}, false, f.getErr
	}
	for _, link := range f.socials {
		if link.ID == id {
			return link, true, nil
		}
	}
	return SocialLink{}, false, nil
}

func (f *fakeRepository) ListHomeSocialLinks(context.Context) ([]SocialLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	return append([]SocialLink(nil), f.socials...), nil
}

func (f *fakeRepository) ListHomeSocialLinksPublished(context.Context) ([]SocialLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	var out []SocialLink
	for _, link := range f.socials {
		if link.PublicationState == StatePublished {
			out = append(out, link)
		}
	}
	return out, nil
}

// --- Auditoría best-effort ---

func (f *fakeRepository) RecordAction(_ context.Context, action audit.Action) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.recordErr != nil {
		return f.recordErr
	}
	f.actions = append(f.actions, action)
	return nil
}

// seedService, seedChannel y seedSocialLink registran elementos ya existentes.
func (f *fakeRepository) seedService(service Service) Service {
	if service.ID == uuid.Nil {
		service.ID = uuid.New()
	}
	service.CreatedAt = time.Now()
	service.UpdatedAt = service.CreatedAt
	f.services = append(f.services, service)
	return service
}

func (f *fakeRepository) seedChannel(channel WhatsappChannel) WhatsappChannel {
	if channel.ID == uuid.Nil {
		channel.ID = uuid.New()
	}
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = channel.CreatedAt
	f.channels = append(f.channels, channel)
	return channel
}

func (f *fakeRepository) seedSocialLink(link SocialLink) SocialLink {
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	link.CreatedAt = time.Now()
	link.UpdatedAt = link.CreatedAt
	f.socials = append(f.socials, link)
	return link
}

// --- Helpers de aserción ---

func errorKind(err error) apperr.Kind {
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		return domainErr.Kind
	}
	return apperr.KindInternal
}

func requireKind(t *testing.T, err error, kind apperr.Kind) *apperr.Error {
	t.Helper()
	if err == nil {
		t.Fatalf("se esperaba un error de kind %v, se obtuvo nil", kind)
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba *apperr.Error, se obtuvo %T: %v", err, err)
	}
	if domainErr.Kind != kind {
		t.Fatalf("kind = %v, se esperaba %v (mensaje %q)", domainErr.Kind, kind, domainErr.Message)
	}
	return domainErr
}

func requireDetail(t *testing.T, err error, field string) {
	t.Helper()
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba *apperr.Error, se obtuvo %T", err)
	}
	if _, ok := domainErr.Details[field]; !ok {
		t.Fatalf("details no incluye %q: %v", field, domainErr.Details)
	}
}

// --- T317: normalización e invariantes ---

func TestNormalizeText(t *testing.T) {
	cases := map[string]string{
		"  hola   mundo  ":   "hola mundo",
		"línea 1\n  línea 2": "línea 1\n línea 2",
		"\t domingo \t":      "domingo",
		"tildes ñ ¿qué? 😀":   "tildes ñ ¿qué? 😀",
		"a\t\tb":             "a b",
	}
	for in, want := range cases {
		if got := normalizeText(in); got != want {
			t.Errorf("normalizeText(%q) = %q, se esperaba %q", in, got, want)
		}
	}
}

func TestNormalizeOptional(t *testing.T) {
	for _, empty := range []string{"", "   ", "\t", " \t "} {
		if got := normalizeOptional(empty); got != nil {
			t.Errorf("normalizeOptional(%q) = %q, se esperaba nil", empty, *got)
		}
	}
	got := normalizeOptional("  x  ")
	if got == nil || *got != "x" {
		t.Fatalf("normalizeOptional(\"  x  \") = %v, se esperaba \"x\"", got)
	}
}

func TestNormalizePhone(t *testing.T) {
	cases := map[string]string{
		"+58 412-123 4567": "+584121234567",
		"(0412) 123.45.67": "04121234567",
		"0412 123 4567":    "04121234567",
	}
	for in, want := range cases {
		if got := normalizePhone(in); got != want {
			t.Errorf("normalizePhone(%q) = %q, se esperaba %q", in, got, want)
		}
	}
	if got := phoneDigits("+584121234567"); got != "584121234567" {
		t.Errorf("phoneDigits() = %q", got)
	}
}

func TestContentLabel(t *testing.T) {
	if got := contentLabel(sectionSchedule, "  Culto   dominical "); got != "Portada · Horario · Culto dominical" {
		t.Errorf("contentLabel() = %q", got)
	}
	if got := contentLabel(sectionAbout, ""); got != "Portada · Quiénes somos" {
		t.Errorf("contentLabel() = %q", got)
	}
}

func TestPublicationStateHelpers(t *testing.T) {
	if got := publicationStateOr(""); got != StateDraft {
		t.Errorf("publicationStateOr(\"\") = %q", got)
	}
	if got := publicationStateOr("published"); got != StatePublished {
		t.Errorf("publicationStateOr(published) = %q", got)
	}
	if code := stateActionCode(StatePublished); code != audit.ActionHomePublish {
		t.Errorf("stateActionCode(published) = %q", code)
	}
	if code := stateActionCode(StateDraft); code != audit.ActionHomeUnpublish {
		t.Errorf("stateActionCode(draft) = %q", code)
	}
}

func TestValidLangAndNetworks(t *testing.T) {
	if !validLang(LangES) || !validLang(LangEN) || validLang("fr") {
		t.Error("validLang no respeta el registro es|en")
	}
	if !validSocialNetwork("youtube") || validSocialNetwork("x") {
		t.Error("validSocialNetwork no respeta el catálogo")
	}
}

func TestValidSocialURL(t *testing.T) {
	cases := []struct {
		network string
		url     string
		want    bool
	}{
		{"facebook", "https://www.facebook.com/simiente", true},
		{"youtube", "https://youtu.be/abc", true},
		{"facebook", "https://twitter.com/simiente", false},
		{"facebook", "http://facebook.com/simiente", false},
		{"instagram", "javascript:alert(1)", false},
		{"tiktok", "https://tiktok.com/@simiente", true},
	}
	for _, tc := range cases {
		if got := validSocialURL(tc.network, tc.url); got != tc.want {
			t.Errorf("validSocialURL(%q, %q) = %v, se esperaba %v", tc.network, tc.url, got, tc.want)
		}
	}
}

func TestNormalizeWhatsappDestination(t *testing.T) {
	got, err := normalizeWhatsappDestination(KindDirect, " +58 412-1234567 ")
	if err != nil || got != "+584121234567" {
		t.Fatalf("direct: got (%q, %v)", got, err)
	}
	if _, err := normalizeWhatsappDestination(KindDirect, "abc"); err == nil {
		t.Error("direct inválido debería fallar")
	}
	got, err = normalizeWhatsappDestination(KindGroup, " https://chat.whatsapp.com/ABC ")
	if err != nil || got != "https://chat.whatsapp.com/ABC" {
		t.Fatalf("group: got (%q, %v)", got, err)
	}
	if _, err := normalizeWhatsappDestination(KindGroup, "https://example.com/grupo"); err == nil {
		t.Error("group de host ajeno debería fallar")
	}
	if _, err := normalizeWhatsappDestination("otro", "x"); err == nil {
		t.Error("kind desconocido debería fallar")
	}
}

func TestNormalizeScheduleTimes(t *testing.T) {
	start, end, err := normalizeScheduleTimes(" 10:00 ", " 12:00 ")
	if err != nil || start != "10:00" || end == nil || *end != "12:00" {
		t.Fatalf("rango válido: (%q, %v, %v)", start, end, err)
	}
	if _, _, err := normalizeScheduleTimes("10:00", ""); err != nil {
		t.Fatalf("fin opcional: %v", err)
	}
	if _, _, err := normalizeScheduleTimes("10:00", "09:00"); errorKind(err) != apperr.KindInvalid {
		t.Errorf("fin anterior al inicio debería ser invalid, fue %v", err)
	}
	if _, _, err := normalizeScheduleTimes("25:00", ""); errorKind(err) != apperr.KindInvalid {
		t.Errorf("inicio mal formado debería ser invalid, fue %v", err)
	}
	if _, _, err := normalizeScheduleTimes("10:00", "9:00"); errorKind(err) != apperr.KindInvalid {
		t.Errorf("fin mal formado debería ser invalid, fue %v", err)
	}
}

// La cadena vacía en un `*En` nunca debe llegar a la BD (analyze I6).
func TestNormalizeOptionalNeverEmptyString(t *testing.T) {
	if got := normalizeOptional("   "); got != nil {
		t.Fatalf("se esperaba nil, se obtuvo %q", *got)
	}
}

var _ Repository = (*fakeRepository)(nil)
