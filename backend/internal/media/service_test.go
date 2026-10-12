package media

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/storage"
)

// fakeRepository guioniza la política de publicación y recuerda el corte
// ("hoy"/"ahora") con el que se le consultó.
type fakeRepository struct {
	published map[string]bool
	err       error
	calls     []cutoffCall
}

// cutoffCall es una llamada al repositorio con su corte.
type cutoffCall struct {
	fileName string
	today    string
	nowTime  string
}

func newFakeRepository(published ...string) *fakeRepository {
	fake := &fakeRepository{published: map[string]bool{}}
	for _, name := range published {
		fake.published[name] = true
	}
	return fake
}

func (f *fakeRepository) IsFilePublished(_ context.Context, fileName, today, nowTime string) (bool, error) {
	f.calls = append(f.calls, cutoffCall{fileName: fileName, today: today, nowTime: nowTime})
	if f.err != nil {
		return false, f.err
	}
	return f.published[fileName], nil
}

// fakeStore es un almacén en memoria para las pruebas de la política.
type fakeStore struct {
	files   map[string]string
	openErr error
}

func (f *fakeStore) Save(context.Context, []byte) (storage.File, error) {
	return storage.File{}, errors.New("no se usa en estas pruebas")
}

func (f *fakeStore) Open(_ context.Context, name string) (io.ReadCloser, error) {
	if f.openErr != nil {
		return nil, f.openErr
	}
	content, ok := f.files[name]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func (f *fakeStore) Delete(context.Context, string) error { return nil }

// errorKind clasifica el error de dominio (la traducción a HTTP vive solo en
// httpserver).
func errorKind(err error) apperr.Kind {
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		return domainErr.Kind
	}
	return apperr.KindInternal
}

const validName = "img_22222222-2222-2222-2222-222222222222.png"

// TestOpenMediaPolicy fija la política de descarga (analyze C4/M6 y R4-6).
func TestOpenMediaPolicy(t *testing.T) {
	repo := newFakeRepository()
	store := &fakeStore{files: map[string]string{validName: "PNG"}}
	service := NewService(ServiceDeps{Repository: repo, Store: store})

	// Nombre fuera del patrón → 400 invalid, sin consultar la base.
	if _, err := service.OpenMedia(context.Background(), "logo.png"); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("nombre inválido debería ser invalid: %v", err)
	}
	if len(repo.calls) != 0 {
		t.Errorf("un nombre inválido no debe llegar al repositorio: %+v", repo.calls)
	}

	// Nombre válido pero no referenciado por contenido publicado y próximo → 404.
	if _, err := service.OpenMedia(context.Background(), validName); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("archivo no publicado debería ser not_found: %v", err)
	}

	// Publicado y existente → se abre.
	repo.published[validName] = true
	reader, err := service.OpenMedia(context.Background(), validName)
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
	repo.published[missing] = true
	if _, err := service.OpenMedia(context.Background(), missing); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("archivo físico ausente debería ser not_found: %v", err)
	}

	// Error del repositorio se propaga (nunca se disfraza de 404).
	repo.err = errors.New("boom")
	if _, err := service.OpenMedia(context.Background(), validName); err == nil || errorKind(err) != apperr.KindInternal {
		t.Fatalf("se esperaba el error del repositorio: %v", err)
	}
}

// TestOpenMediaWithoutStore responde 404 cuando el almacén no está cableado
// (nunca un 500: el archivo simplemente no está disponible).
func TestOpenMediaWithoutStore(t *testing.T) {
	repo := newFakeRepository(validName)
	service := NewService(ServiceDeps{Repository: repo})

	if _, err := service.OpenMedia(context.Background(), validName); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("sin almacén debería ser not_found: %v", err)
	}
}

// TestOpenMediaCutoffUsesAppTimeZone fija R4-2: el "hoy"/"ahora" del corte se
// calculan en la zona del lugar (APP_TZ) y viajan como parámetros al SQL. Un
// instante de medianoche UTC ya es "ayer" en Costa Rica (UTC−06:00).
func TestOpenMediaCutoffUsesAppTimeZone(t *testing.T) {
	costaRica, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	instant := time.Date(2026, time.October, 17, 0, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		location  *time.Location
		wantToday string
		wantNow   string
	}{
		{name: "zona del lugar", location: costaRica, wantToday: "2026-10-16", wantNow: "18:30"},
		{name: "UTC", location: time.UTC, wantToday: "2026-10-17", wantNow: "00:30"},
		{name: "sin zona configurada", location: nil, wantToday: "2026-10-17", wantNow: "00:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository(validName)
			store := &fakeStore{files: map[string]string{validName: "PNG"}}
			service := NewService(ServiceDeps{
				Repository: repo,
				Store:      store,
				Location:   tt.location,
				Now:        func() time.Time { return instant },
			})

			if _, err := service.OpenMedia(context.Background(), validName); err != nil {
				t.Fatalf("OpenMedia: %v", err)
			}
			if len(repo.calls) != 1 {
				t.Fatalf("llamadas al repositorio = %d, se esperaba 1", len(repo.calls))
			}
			call := repo.calls[0]
			if call.today != tt.wantToday || call.nowTime != tt.wantNow {
				t.Errorf("corte = %s %s, se esperaba %s %s", call.today, call.nowTime, tt.wantToday, tt.wantNow)
			}
		})
	}
}
