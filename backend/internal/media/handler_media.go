package media

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/storage"
)

// handler_media.go publica la descarga pública de imágenes:
// GET /api/v1/media/{fileName} (misma ruta, cabeceras y errores que F3; R4-6).
// La política vive en el service: un nombre que no cumple el patrón generado por
// el servidor es 400 invalid; un nombre válido pero inexistente o NO
// referenciado por contenido publicado y próximo es 404 not_found. El handler
// solo fija las cabeceras seguras.

// Service es el puerto que el handler necesita del service. Lo declara quien lo
// consume (arq. R3) y lo implementa *service (service.go).
type Service interface {
	// OpenMedia aplica la política de descarga pública y abre el archivo.
	OpenMedia(ctx context.Context, fileName string) (io.ReadCloser, error)
}

// Handler expone la descarga pública de archivos. Es solo HTTP.
type Handler struct {
	service Service
	logger  *slog.Logger
}

// HandlerDeps reúne las dependencias del Handler.
type HandlerDeps struct {
	// Service es el servicio del dominio. Obligatorio.
	Service Service
	// Logger registra errores; nil usa el logger por defecto.
	Logger *slog.Logger
}

// NewHandler construye el Handler del dominio media. Si Logger es nil se usa el
// logger por defecto.
func NewHandler(deps HandlerDeps) *Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{service: deps.Service, logger: logger}
}

// GetMedia responde GET /api/v1/media/{fileName}: sirve una imagen subida desde
// el panel SOLO si está referenciada por contenido publicado y próximo.
// Cabeceras: X-Content-Type-Options: nosniff, Content-Disposition: inline y
// Cache-Control: no-store (la descarga depende del estado de publicación,
// RG3-8). Sin autenticación.
// Códigos: 200 (binario); nombre fuera del patrón → 400 invalid; inexistente o
// no publicado → 404 not_found.
func (h *Handler) GetMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.service == nil {
		httpserver.WriteError(ctx, w, h.logger,
			apperr.Internal(errors.New("media: servicio no configurado")))
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

// compile-time check: *service must satisfy the port the handler declares.
var _ Service = (*service)(nil)
