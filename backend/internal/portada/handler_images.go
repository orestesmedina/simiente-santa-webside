package portada

import (
	"context"
	"errors"
	"io"
	"net/http"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_images.go publica la subida de imágenes del panel:
// POST /api/v1/admin/portada/imagenes (T323, R3-8). El handler es fino: aplica
// el tope de tamaño en stream (http.MaxBytesReader), lee los bytes y delega en
// el service, que valida la firma binaria, genera el nombre y audita la subida
// con `home.image.upload` de forma fail-closed (analyze I8). El cuerpo va en
// multipart/form-data con el campo `file`.

// UploadImage responde POST /api/v1/admin/portada/imagenes (FR-002/FR-011):
// sube el logotipo o la imagen de portada. Solo JPEG/PNG/WebP por firma binaria
// (sin SVG ni GIF) y ≤ UPLOAD_MAX_BYTES (R3-8). La subida NO cambia contenido
// visible: la referencia se registra al guardar la identidad.
// Códigos: 201 con ImageUploadResult; sin archivo o tipo no permitido → 400
// invalid (details.file); archivo demasiado grande → 400 invalid
// (details.fileSize).
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.service.UploadImage(ctx, identity.UserID, data)
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
