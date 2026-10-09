package portada

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/storage"
)

// fakeStore es un Store en memoria para probar la política de descarga.
type fakeStore struct {
	files     map[string][]byte
	openErr   error
	deleteErr error
	deleted   []string
}

func (s *fakeStore) Save(context.Context, []byte) (storage.File, error) {
	return storage.File{}, nil
}

func (s *fakeStore) Open(_ context.Context, name string) (io.ReadCloser, error) {
	if s.openErr != nil {
		return nil, s.openErr
	}
	data, ok := s.files[name]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *fakeStore) Delete(_ context.Context, name string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, name)
	return nil
}

func strptr(s string) *string { return &s }

// T318: fallback `en → es` por campo, nunca un hueco (SC-006).
func TestGetPortadaFallbackPerField(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{
		ID:               uuid.New(),
		NameEs:           "Iglesia Simiente",
		NameEn:           nil, // sin traducción → fallback
		TaglineEs:        strptr("Una familia"),
		TaglineEn:        nil,
		MissionEn:        strptr("Our mission"),
		LogoFile:         strptr("img_11111111-1111-1111-1111-111111111111.jpg"),
		PublicationState: StatePublished,
	}
	repo.about = &About{TextEs: "Somos una iglesia", PublicationState: StatePublished}
	repo.contact = &Contact{AddressEs: "Calle 1", Email: "hola@simiente.org", Phone: "04121234567", PublicationState: StatePublished}

	service := NewService(ServiceDeps{Repository: repo})

	portada, err := service.GetPortada(context.Background(), LangEN)
	if err != nil {
		t.Fatalf("GetPortada(en) = %v", err)
	}
	if portada.Lang != LangEN {
		t.Fatalf("lang = %q", portada.Lang)
	}
	if portada.Identity == nil || portada.Identity.Name != "Iglesia Simiente" {
		t.Fatalf("name sin fallback: %+v", portada.Identity)
	}
	if portada.Identity.Tagline == nil || *portada.Identity.Tagline != "Una familia" {
		t.Fatalf("tagline sin fallback: %+v", portada.Identity.Tagline)
	}
	if portada.Identity.Mission == nil || *portada.Identity.Mission != "Our mission" {
		t.Fatalf("mission debería usar la traducción: %+v", portada.Identity.Mission)
	}
	wantLogo := MediaPathPrefix + "img_11111111-1111-1111-1111-111111111111.jpg"
	if portada.Identity.LogoURL != wantLogo {
		t.Fatalf("logoUrl = %q, se esperaba %q", portada.Identity.LogoURL, wantLogo)
	}
	if portada.About == nil || portada.About.Text != "Somos una iglesia" {
		t.Fatalf("about sin fallback: %+v", portada.About)
	}
}

// T318: secciones sin elementos publicados se omiten por completo (SC-012).
func TestGetPortadaOmitsEmptySections(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{NameEs: "Borrador", PublicationState: StateDraft}
	repo.about = &About{TextEs: "Borrador", PublicationState: StateDraft}
	service := NewService(ServiceDeps{Repository: repo})

	portada, err := service.GetPortada(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetPortada = %v", err)
	}
	if portada.Identity != nil || portada.About != nil || portada.Contact != nil ||
		portada.Schedule != nil || portada.Whatsapp != nil || portada.Socials != nil {
		t.Fatalf("con todo en borrador solo debe quedar lang: %+v", portada)
	}
}

// T318: sección mixta publica solo lo publicado.
func TestGetPortadaMixedState(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{NameEs: "Publicada", PublicationState: StatePublished}
	repo.services = []Service{
		{ID: uuid.New(), NameEs: "Publicado", PlaceEs: "Sede", StartTime: "10:00", PublicationState: StatePublished},
		{ID: uuid.New(), NameEs: "Borrador", PlaceEs: "Sede", StartTime: "11:00", PublicationState: StateDraft},
	}
	service := NewService(ServiceDeps{Repository: repo})

	portada, err := service.GetPortada(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetPortada = %v", err)
	}
	if len(portada.Schedule) != 1 || portada.Schedule[0].Name != "Publicado" {
		t.Fatalf("se esperaba solo el servicio publicado: %+v", portada.Schedule)
	}
}

// T318: idioma fuera de es|en → 400 invalid con details.lang.
func TestGetPortadaInvalidLang(t *testing.T) {
	service := NewService(ServiceDeps{Repository: newFakeRepository()})
	_, err := service.GetPortada(context.Background(), "fr")
	requireKind(t, err, apperr.KindInvalid)
	requireDetail(t, err, "lang")
}

// T318: el orden de las colecciones se conserva y los enlaces se arman (SC-009).
func TestGetPortadaOrderingAndLinks(t *testing.T) {
	repo := newFakeRepository()
	first := uuid.New()
	second := uuid.New()
	repo.services = []Service{
		{ID: first, NameEs: "Primero", PlaceEs: "Sede", StartTime: "08:00", PublicationState: StatePublished},
		{ID: second, NameEs: "Segundo", PlaceEs: "Sede", StartTime: "09:00", PublicationState: StatePublished},
	}
	repo.channels = []WhatsappChannel{
		{ID: uuid.New(), NameEs: "Directo", Kind: KindDirect, Destination: "+584121234567", PublicationState: StatePublished},
		{ID: uuid.New(), NameEs: "Grupo", Kind: KindGroup, Destination: "https://chat.whatsapp.com/ABC", PublicationState: StatePublished},
	}
	repo.socials = []SocialLink{
		{ID: uuid.New(), Network: "facebook", URL: "https://facebook.com/x", PublicationState: StatePublished},
		{ID: uuid.New(), Network: "instagram", URL: "https://instagram.com/x", PublicationState: StatePublished},
	}
	service := NewService(ServiceDeps{Repository: repo})

	portada, err := service.GetPortada(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetPortada = %v", err)
	}
	if len(portada.Schedule) != 2 || portada.Schedule[0].ID != first.String() || portada.Schedule[1].ID != second.String() {
		t.Fatalf("orden del horario alterado: %+v", portada.Schedule)
	}
	if portada.Whatsapp[0].URL != "https://wa.me/584121234567" {
		t.Fatalf("url direct = %q", portada.Whatsapp[0].URL)
	}
	if portada.Whatsapp[1].URL != "https://chat.whatsapp.com/ABC" {
		t.Fatalf("url group = %q", portada.Whatsapp[1].URL)
	}
}

// T318: política de descarga (analyze C4/M6).
func TestOpenMediaPolicy(t *testing.T) {
	const valid = "img_22222222-2222-2222-2222-222222222222.png"
	repo := newFakeRepository()
	store := &fakeStore{files: map[string][]byte{valid: []byte("PNG")}}
	service := NewService(ServiceDeps{Repository: repo, Store: store})

	// Nombre fuera del patrón → 400 invalid.
	if _, err := service.OpenMedia(context.Background(), "logo.png"); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("nombre inválido debería ser invalid: %v", err)
	}

	// Nombre válido pero no referenciado por contenido publicado → 404.
	if _, err := service.OpenMedia(context.Background(), valid); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("archivo no publicado debería ser not_found: %v", err)
	}

	// Publicado y existente → se abre.
	repo.publishedFiles[valid] = true
	reader, err := service.OpenMedia(context.Background(), valid)
	if err != nil {
		t.Fatalf("OpenMedia publicado = %v", err)
	}
	data, _ := io.ReadAll(reader)
	_ = reader.Close()
	if string(data) != "PNG" {
		t.Fatalf("contenido = %q", data)
	}

	// Publicado pero el archivo físico no existe → 404.
	missing := "img_33333333-3333-3333-3333-333333333333.webp"
	repo.publishedFiles[missing] = true
	if _, err := service.OpenMedia(context.Background(), missing); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("archivo físico ausente debería ser not_found: %v", err)
	}

	// Error del repositorio se propaga.
	repo.getErr = errors.New("boom")
	if _, err := service.OpenMedia(context.Background(), valid); err == nil {
		t.Fatal("se esperaba error del repositorio")
	}
}
