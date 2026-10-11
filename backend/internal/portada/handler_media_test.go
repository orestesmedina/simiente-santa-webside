package portada

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests del handler de descarga GET /api/v1/media/{fileName} (T323): el handler
// real sobre el service real, con un LocalStore sobre t.TempDir() y el
// repositorio falso. Comprueban la política de publicación (analyze C4/M6) y las
// cabeceras seguras.

// newMediaHandler construye el handler y devuelve también el Store para sembrar
// archivos reales.
func newMediaHandler(t *testing.T, repo *fakeRepository, dir string, maxBytes int64) (*Handler, storage.Store) {
	t.Helper()
	logger, _ := testutil.NewLogger()
	store := storage.NewLocalStore(dir, maxBytes)
	svc := NewService(ServiceDeps{Repository: repo, Store: store, Logger: logger})
	return NewHandler(HandlerDeps{Service: svc, MaxUploadBytes: maxBytes, Logger: logger}), store
}

func doMedia(t *testing.T, h *Handler, name string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+name, nil)
	req.SetPathValue("fileName", name)
	rec := httptest.NewRecorder()
	h.GetMedia(rec, req)
	return rec
}

func TestGetMediaHandlerPublishedFile(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	h, store := newMediaHandler(t, repo, dir, 1<<20)

	file, err := store.Save(context.Background(), pngSignature)
	if err != nil {
		t.Fatalf("sembrar archivo: %v", err)
	}
	repo.publishedFiles[file.Name] = true

	rec := doMedia(t, h, file.Name)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if !bytes.Equal(rec.Body.Bytes(), pngSignature) {
		t.Errorf("cuerpo = %v, se esperaba la imagen", rec.Body.Bytes())
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, se esperaba nosniff", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != "inline" {
		t.Errorf("Content-Disposition = %q, se esperaba inline", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, se esperaba image/png", got)
	}
}

func TestGetMediaHandlerInvalidName(t *testing.T) {
	dir := t.TempDir()
	h, _ := newMediaHandler(t, newFakeRepository(), dir, 1<<20)

	// Un nombre fuera del patrón se rechaza ANTES de tocar el disco (400).
	for _, name := range []string{"../../etc/passwd", "img_abc.jpg", "logo.png"} {
		rec := doMedia(t, h, name)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("GET %q: status = %d, se esperaba 400 (%s)", name, rec.Code, rec.Body)
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid"`)) {
			t.Errorf("GET %q: sobre inesperado: %s", name, rec.Body)
		}
	}
}

func TestGetMediaHandlerNotPublished(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	h, store := newMediaHandler(t, repo, dir, 1<<20)

	file, err := store.Save(context.Background(), pngSignature)
	if err != nil {
		t.Fatalf("sembrar archivo: %v", err)
	}
	// El archivo existe en disco pero NO está referenciado por contenido
	// publicado: 404 (analyze C4).
	rec := doMedia(t, h, file.Name)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"not_found"`)) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

func TestGetMediaHandlerMissingFile(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	h, _ := newMediaHandler(t, repo, dir, 1<<20)

	name := "img_11111111-1111-1111-1111-111111111111.png"
	repo.publishedFiles[name] = true // publicado pero el archivo no existe

	rec := doMedia(t, h, name)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
	}
}
