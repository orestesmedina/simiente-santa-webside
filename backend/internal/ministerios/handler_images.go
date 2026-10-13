package ministerios

import (
	"context"
	"errors"
	"io"
	"net/http"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_images.go publica la subida de imágenes del panel:
// POST /api/v1/admin/ministerios/imagenes (T620, R6-7). El handler es fino:
// aplica el tope de tamaño en stream (http.MaxBytesReader), lee los bytes y
// delega en el service, que valida la firma binaria (JPEG/PNG/WebP, sin SVG ni
// GIF), genera el nombre `img_<uuid>.<ext>` y audita la subida con
// `ministry.image.upload` de forma fail-closed. El cuerpo va en
// multipart/form-data con el campo `file`.

// multipartOverhead es el margen que se suma al tope de la imagen para el
// envoltorio multipart antes de aplicar http.MaxBytesReader: el tamaño real del
// archivo se comprueba después, ya con los bytes en memoria.
const multipartOverhead = 1 << 20

// UploadMinisterioImage responde POST /api/v1/admin/ministerios/imagenes
// (FR-020): sube la imagen de un ministerio. Solo JPEG/PNG/WebP por firma
// binaria (sin SVG ni GIF) y ≤ UPLOAD_MAX_BYTES. La subida NO cambia contenido
// visible: la referencia se registra al guardar el ministerio (`imageFile`).
// Códigos: 201 con ImageUploadResult; sin archivo o tipo no permitido → 400
// invalid (details.file); archivo demasiado grande → 400 invalid
// (details.fileSize).
func (h *Handler) UploadMinisterioImage(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.service.UploadMinisterioImage(ctx, identity.UserID, data)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, result)
}

// writeFileTooLarge responde el 400 del contrato por una imagen que supera el
// tope configurado (details.fileSize; no se añade un error.code nuevo).
func (h *Handler) writeFileTooLarge(ctx context.Context, w http.ResponseWriter) {
	httpserver.WriteError(ctx, w, h.logger, apperr.Invalid(
		"La imagen es demasiado grande",
		apperr.WithDetails(map[string]any{"fileSize": "la imagen supera el tamaño máximo permitido"}),
	))
}
