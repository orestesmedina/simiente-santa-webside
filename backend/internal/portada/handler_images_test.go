package portada

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests del handler de subida POST /api/v1/admin/portada/imagenes (T323): el
// handler real sobre el service real, con un LocalStore sobre t.TempDir() y el
// repositorio falso (sin PostgreSQL). Comprueban la firma binaria, el tope, la
// auditoría fail-closed y la limpieza del archivo.

// pngSignature son los 8 bytes de firma que http.DetectContentType reconoce
// como image/png.
var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// withAdminIdentity deja la identidad en el contexto como haría authn.
func withAdminIdentity(req *http.Request, identity session.Identity) *http.Request {
	return req.WithContext(session.ContextWithIdentity(req.Context(), identity))
}

// multipartFile construye un cuerpo multipart/form-data con el campo y el
// archivo indicados.
func multipartFile(t *testing.T, field, filename string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("crear parte multipart: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("escribir parte multipart: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("cerrar multipart: %v", err)
	}
	return &buf, mw.FormDataContentType()
}

// newImageHandler construye el handler sobre un service real con el
// repositorio falso y un LocalStore en dir.
func newImageHandler(t *testing.T, repo *fakeRepository, dir string, maxBytes int64) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	svc := NewService(ServiceDeps{
		Repository: repo,
		Store:      storage.NewLocalStore(dir, maxBytes),
		Logger:     logger,
	})
	return NewHandler(HandlerDeps{Service: svc, MaxUploadBytes: maxBytes, Logger: logger})
}

func doUpload(t *testing.T, h *Handler, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/portada/imagenes", body)
	req.Header.Set("Content-Type", contentType)
	req = withAdminIdentity(req, session.Identity{UserID: uuid.New(), Permissions: []string{"portada"}})
	rec := httptest.NewRecorder()
	h.UploadImage(rec, req)
	return rec
}

func TestUploadImageHandlerOK(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	h := newImageHandler(t, repo, dir, 1<<20)

	body, ct := multipartFile(t, "file", "logo.png", pngSignature)
	rec := doUpload(t, h, body, ct)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201 (%s)", rec.Code, rec.Body)
	}
	var got ImageUploadResult
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es ImageUploadResult: %v (%s)", err, rec.Body)
	}
	if !storage.ValidName(got.FileName) {
		t.Errorf("fileName = %q, no cumple el patrón del servidor", got.FileName)
	}
	if got.MimeType != "image/png" {
		t.Errorf("mimeType = %q, se esperaba image/png", got.MimeType)
	}
	if got.URL != MediaPathPrefix+got.FileName {
		t.Errorf("url = %q, se esperaba %q", got.URL, MediaPathPrefix+got.FileName)
	}
	if got.SizeBytes != int64(len(pngSignature)) {
		t.Errorf("sizeBytes = %d, se esperaba %d", got.SizeBytes, len(pngSignature))
	}

	// El archivo quedó en disco.
	if _, err := os.Stat(filepath.Join(dir, got.FileName)); err != nil {
		t.Errorf("el archivo no está en disco: %v", err)
	}
	// La subida dejó su fila de auditoría home.image.upload (analyze I8).
	actions := repo.recorded()
	if len(actions) != 1 {
		t.Fatalf("se esperaba 1 acción de auditoría, hay %d: %+v", len(actions), actions)
	}
	if actions[0].Code != audit.ActionHomeImageUpload {
		t.Errorf("código = %q, se esperaba %q", actions[0].Code, audit.ActionHomeImageUpload)
	}
	if actions[0].TargetKind != audit.TargetContent {
		t.Errorf("targetKind = %q, se esperaba content", actions[0].TargetKind)
	}
	if actions[0].Result != audit.ResultSuccess {
		t.Errorf("result = %q, se esperaba success", actions[0].Result)
	}
	if want := contentLabel(sectionImage, got.FileName); actions[0].TargetLabel != want {
		t.Errorf("targetLabel = %q, se esperaba %q", actions[0].TargetLabel, want)
	}
}

func TestUploadImageHandlerRejectsMissingFile(t *testing.T) {
	dir := t.TempDir()
	h := newImageHandler(t, newFakeRepository(), dir, 1<<20)

	// Cuerpo multipart sin el campo `file`.
	body, ct := multipartFile(t, "otro", "logo.png", pngSignature)
	rec := doUpload(t, h, body, ct)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid"`)) ||
		!bytes.Contains(rec.Body.Bytes(), []byte(`"file"`)) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

func TestUploadImageHandlerRejectsBySignature(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	h := newImageHandler(t, repo, dir, 1<<20)

	// Un SVG (o un HTML renombrado) no tiene firma de imagen permitida.
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`)
	body, ct := multipartFile(t, "file", "logo.svg", svg)
	rec := doUpload(t, h, body, ct)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid"`)) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
	if len(repo.recorded()) != 0 {
		t.Errorf("un formato rechazado no debe auditarse: %+v", repo.recorded())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("un formato rechazado no debe dejar archivos: %v", entries)
	}
}

func TestUploadImageHandlerRejectsTooLarge(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	h := newImageHandler(t, repo, dir, 16)

	big := append(append([]byte{}, pngSignature...), bytes.Repeat([]byte{0x00}, 200)...)
	body, ct := multipartFile(t, "file", "logo.png", big)
	rec := doUpload(t, h, body, ct)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"fileSize"`)) {
		t.Errorf("se esperaba details.fileSize: %s", rec.Body)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("una subida demasiado grande no debe dejar archivos: %v", entries)
	}
}

func TestUploadImageHandlerAuditFailureIsClosed(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	repo.recordErr = context.DeadlineExceeded
	h := newImageHandler(t, repo, dir, 1<<20)

	body, ct := multipartFile(t, "file", "logo.png", pngSignature)
	rec := doUpload(t, h, body, ct)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	// Fail-closed: sin registro, la subida se aborta y se elimina el archivo.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("un fallo de registro debe eliminar el archivo: %v", entries)
	}
}
