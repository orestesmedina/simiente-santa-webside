package portada

import (
	"io"
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/storage"
)

// handler_media.go publica la descarga pública de imágenes:
// GET /api/v1/media/{fileName} (T323). La política (analyze C4/M6) vive en el
// service: un nombre que no cumple el patrón generado por el servidor es 400
// invalid; un nombre válido pero inexistente o NO referenciado por contenido
// publicado es 404 not_found. El handler solo fija las cabeceras seguras.

// GetMedia responde GET /api/v1/media/{fileName} (FR-013/FR-019): sirve una
// imagen subida desde el panel SOLO si está referenciada por contenido
// publicado. Cabeceras: X-Content-Type-Options: nosniff, Content-Disposition:
// inline y Cache-Control: no-store (la descarga depende del estado de
// publicación, RG3-8). Sin autenticación.
// Códigos: 200 (binario); nombre fuera del patrón → 400 invalid; inexistente o
// no publicado → 404 not_found.
func (h *Handler) GetMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.serviceConfigured(ctx, w) {
		return
	}

	fileName := r.PathValue("fileName")
	reader, err := h.service.OpenMedia(ctx, fileName)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	defer func() { _ = reader.Close() }()

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", storage.ContentTypeFor(fileName))
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, reader); err != nil {
		// La cabecera ya viajó: solo se deja constancia en el log.
		logger := httpserver.RequestLoggerFromContext(ctx)
		if logger == nil {
			logger = h.logger
		}
		if logger != nil {
			logger.Error("no se pudo escribir la imagen en la respuesta", "error", err)
		}
	}
}
