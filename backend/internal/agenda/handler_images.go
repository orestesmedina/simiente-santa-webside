package agenda

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_images.go publica la subida de imágenes del panel por módulo (T427,
// R3-8 de F3):
//
//	POST /api/v1/admin/eventos/imagenes
//	POST /api/v1/admin/actividades/imagenes
//
// El handler es fino: aplica el tope de tamaño en stream (http.MaxBytesReader),
// lee los bytes y delega en el service del módulo, que valida la firma binaria
// (solo JPEG/PNG/WebP: sin SVG ni GIF), genera el nombre `img_<uuid>.<ext>` con
// platform/storage y audita la subida con `event.image.upload` /
// `activity.image.upload` de forma fail-closed (analyze I8/R4-10). El cuerpo va
// en multipart/form-data con el campo `file`.
//
// Cada endpoint va tras SU permiso de módulo (routes.go, T428): una cuenta solo
// de actividades no sube al área de eventos, ni al revés (FR-012).

// multipartOverhead es el margen que se suma al tope de la imagen para el
// envoltorio multipart antes de aplicar http.MaxBytesReader: el tamaño real del
// archivo se comprueba después con el límite exacto.
const multipartOverhead = 1 << 20

// UploadEventoImage responde POST /api/v1/admin/eventos/imagenes (FR-020):
// sube la imagen de un evento. Solo JPEG/PNG/WebP por firma binaria (sin SVG ni
// GIF) y ≤ UPLOAD_MAX_BYTES (R3-8). La subida NO cambia contenido visible: la
// referencia se registra al guardar el evento (`imageFile`).
// Códigos: 201 con ImageUploadResult; sin archivo o tipo no permitido → 400
// invalid (details.file); archivo demasiado grande → 400 invalid
// (details.fileSize); sin sesión → 401; sin el permiso `eventos` → 403 (cadena).
func (h *Handler) UploadEventoImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, ModuleEvents)
}

// UploadActividadImage responde POST /api/v1/admin/actividades/imagenes
// (FR-020): igual que la de eventos, tras el permiso `actividades`.
// Códigos: los mismos que UploadEventoImage.
func (h *Handler) UploadActividadImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, ModuleActivities)
}

// uploadImage unifica el preludio de las subidas de los módulos: exige sesión y
// servicio, aplica el tope en stream, lee el archivo y delega en el service del
// módulo, que audita la subida (fail-closed).
func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request, module Module) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	// El tope se aplica en stream (antes de leer el archivo) con un margen para
	// el envoltorio multipart; el tamaño real se comprueba después.
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes+multipartOverhead)

	file, _, err := r.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			h.writeFileTooLarge(ctx, w)
			return
		}
		httpserver.WriteError(ctx, w, h.logger, apperr.Invalid(
			"No se recibió ninguna imagen",
			apperr.WithDetails(map[string]any{"file": "adjunta un archivo de imagen"}),
		))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			h.writeFileTooLarge(ctx, w)
			return
		}
		httpserver.WriteError(ctx, w, h.logger, apperr.Internal(err))
		return
	}
	if int64(len(data)) > h.maxUploadBytes {
		h.writeFileTooLarge(ctx, w)
		return
	}

	result, err := h.uploadForModule(ctx, module, identity.UserID, data)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, result)
}

// uploadForModule delega la subida en el método del service del módulo (una
// cuenta solo de actividades no sube al área de eventos: cada ruta lleva su
// permiso, T428).
func (h *Handler) uploadForModule(ctx context.Context, module Module, actorID uuid.UUID, data []byte) (ImageUploadResult, error) {
	switch module {
	case ModuleEvents:
		return h.service.UploadEventoImage(ctx, actorID, data)
	case ModuleActivities:
		return h.service.UploadActividadImage(ctx, actorID, data)
	default:
		return ImageUploadResult{}, apperr.Internal(
			errors.New("agenda: módulo de subida no soportado"),
		)
	}
}

// writeFileTooLarge responde el 400 del contrato por una imagen que supera el
// tope configurado (details.fileSize; no se añade un error.code nuevo).
func (h *Handler) writeFileTooLarge(ctx context.Context, w http.ResponseWriter) {
	httpserver.WriteError(ctx, w, h.logger, apperr.Invalid(
		"La imagen es demasiado grande",
		apperr.WithDetails(map[string]any{"fileSize": "la imagen supera el tamaño máximo permitido"}),
	))
}
