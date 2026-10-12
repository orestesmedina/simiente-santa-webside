package agenda

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de la subida de imágenes por módulo (T427): el handler real sobre el
// service real, con un LocalStore sobre t.TempDir() y el repositorio falso (sin
// PostgreSQL). Comprueban la firma binaria, el tope de tamaño, la auditoría por
// módulo, el fail-closed y la limpieza del archivo.

// pngSignature son los 8 bytes de firma que http.DetectContentType reconoce como
// image/png.
var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// jpegSignature son los 3 bytes de firma de un JPEG.
var jpegSignature = []byte("\xff\xd8\xff")

// newImageHandler construye el handler sobre un service real con el repositorio
// falso y un LocalStore en dir.
func newImageHandler(t *testing.T, repo *fakeRepository, dir string, maxBytes int64) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	location, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	svc := NewService(ServiceDeps{
		Repository: repo,
		Store:      storage.NewLocalStore(dir, maxBytes),
		Location:   location,
		Now:        func() time.Time { return time.Date(2026, time.October, 11, 18, 0, 0, 0, time.UTC) },
		Logger:     logger,
	})
	return NewHandler(HandlerDeps{Service: svc, MaxUploadBytes: maxBytes, Logger: logger})
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

// doUpload lanza la subida indicada con identidad del panel.
func doUpload(t *testing.T, h *Handler, path string, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", contentType)
	req = withIdentity(req, session.Identity{UserID: uuid.New(), Permissions: []string{"eventos"}})
	rec := httptest.NewRecorder()
	switch path {
	case "/api/v1/admin/eventos/imagenes":
		h.UploadEventoImage(rec, req)
	default:
		h.UploadActividadImage(rec, req)
	}
	return rec
}

// TestUploadImageHandlerOK fija el 201 con el resultado de la subida y la fila
// de auditoría del módulo correspondiente (R4-10).
func TestUploadImageHandlerOK(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		code    string
		label   string
		content []byte
	}{
		{
			name:    "imagen de evento",
			path:    "/api/v1/admin/eventos/imagenes",
			code:    audit.ActionEventImageUpload,
			label:   "Eventos · Imagen",
			content: pngSignature,
		},
		{
			name:    "imagen de actividad",
			path:    "/api/v1/admin/actividades/imagenes",
			code:    audit.ActionActivityImageUpload,
			label:   "Actividades · Imagen",
			content: jpegSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			repo := newFakeRepository()
			h := newImageHandler(t, repo, dir, 1<<20)

			body, contentType := multipartFile(t, "file", "foto.png", tt.content)
			rec := doUpload(t, h, tt.path, body, contentType)

			if rec.Code != http.StatusCreated {
				t.Fatalf("status = %d, se esperaba 201 (%s)", rec.Code, rec.Body)
			}
			var got ImageUploadResult
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("cuerpo no es ImageUploadResult: %v (%s)", err, rec.Body)
			}
			if !storage.ValidName(got.FileName) {
				t.Errorf("fileName = %q, se esperaba el nombre generado por el servidor", got.FileName)
			}
			if got.URL != MediaPathPrefix+got.FileName || got.SizeBytes != int64(len(tt.content)) {
				t.Errorf("resultado = %+v", got)
			}

			if len(repo.recorded) != 1 {
				t.Fatalf("filas de auditoría = %+v, se esperaba 1", repo.recorded)
			}
			action := repo.recorded[0]
			if action.Code != tt.code || action.TargetKind != audit.TargetContent {
				t.Errorf("acción = %+v", action)
			}
			if action.TargetLabel != tt.label+" · "+got.FileName {
				t.Errorf("targetLabel = %q", action.TargetLabel)
			}
		})
	}
}

// TestUploadImageHandlerMissingFile fija el 400 (details.file) cuando no hay
// archivo o el campo no es `file`.
func TestUploadImageHandlerMissingFile(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	h := newImageHandler(t, repo, dir, 1<<20)

	t.Run("sin archivo", func(t *testing.T) {
		body, contentType := multipartFile(t, "file", "foto.png", nil)
		rec := doUpload(t, h, "/api/v1/admin/eventos/imagenes", body, contentType)
		// Un archivo vacío llega al service, que responde 400 con details.file.
		if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte(`"file"`)) {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("campo erróneo", func(t *testing.T) {
		body, contentType := multipartFile(t, "imagen", "foto.png", pngSignature)
		rec := doUpload(t, h, "/api/v1/admin/eventos/imagenes", body, contentType)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid"`)) || !bytes.Contains(rec.Body.Bytes(), []byte(`"file"`)) {
			t.Errorf("sobre inesperado: %s", rec.Body)
		}
	})

	t.Run("sin multipart", func(t *testing.T) {
		body := bytes.NewBufferString("no es multipart")
		rec := doUpload(t, h, "/api/v1/admin/eventos/imagenes", body, "text/plain")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	// Ninguna subida rechazada dejó fila de auditoría.
	if len(repo.recorded) != 0 {
		t.Errorf("se auditó una subida rechazada: %+v", repo.recorded)
	}
}

// TestUploadImageHandlerTooLarge fija el 400 (details.fileSize) con el tope
// aplicado en stream.
func TestUploadImageHandlerTooLarge(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	const maxBytes = 1 << 10
	h := newImageHandler(t, repo, dir, maxBytes)

	big := append(append([]byte{}, pngSignature...), bytes.Repeat([]byte("a"), maxBytes+1)...)
	body, contentType := multipartFile(t, "file", "grande.png", big)
	rec := doUpload(t, h, "/api/v1/admin/eventos/imagenes", body, contentType)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"fileSize"`)) {
		t.Errorf("sin details.fileSize: %s", rec.Body)
	}
	if len(repo.recorded) != 0 {
		t.Errorf("se auditó una subida rechazada: %+v", repo.recorded)
	}
}

// TestUploadImageHandlerRejectsUnsupportedTypes fija que la firma binaria manda:
// SVG, HTML renombrado o texto no son imágenes válidas.
func TestUploadImageHandlerRejectsUnsupportedTypes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  []byte
	}{
		{name: "svg renombrado a png", filename: "logo.png", content: []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)},
		{name: "html renombrado a jpg", filename: "foto.jpg", content: []byte("<html><body>hola</body></html>")},
		{name: "gif", filename: "anim.gif", content: []byte("GIF89a")},
		{name: "texto plano", filename: "notas.png", content: []byte("solo texto")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			repo := newFakeRepository()
			h := newImageHandler(t, repo, dir, 1<<20)

			body, contentType := multipartFile(t, "file", tt.filename, tt.content)
			rec := doUpload(t, h, "/api/v1/admin/eventos/imagenes", body, contentType)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !bytes.Contains(rec.Body.Bytes(), []byte(`"file"`)) {
				t.Errorf("sin details.file: %s", rec.Body)
			}
			if len(repo.recorded) != 0 {
				t.Errorf("se auditó una subida rechazada: %+v", repo.recorded)
			}
		})
	}
}

// TestUploadImageHandlerFailClosed fija el fail-closed (analyze I8/R4-10): si el
// registro de auditoría falla, la subida se aborta y el archivo se elimina.
func TestUploadImageHandlerFailClosed(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeRepository()
	repo.recordErr = errors.New("auditoría caída")
	h := newImageHandler(t, repo, dir, 1<<20)

	body, contentType := multipartFile(t, "file", "foto.png", pngSignature)
	rec := doUpload(t, h, "/api/v1/admin/eventos/imagenes", body, contentType)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	// El archivo no queda en disco: el directorio del almacén se vació al fallar
	// el registro (fail-closed).
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("leer el directorio del almacén: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("quedaron %d archivos tras el fallo del registro: %v", len(entries), entries)
	}
}
