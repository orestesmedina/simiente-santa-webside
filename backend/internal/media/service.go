// Package media es el dominio de archivos de imagen publicados (F4, promovido
// desde `portada`, R4-6/P4-1): sirve la descarga pública
// `GET /api/v1/media/{fileName}`.
//
// Deja de ser un asunto de la portada porque con F4 (y con F5–F9) la descarga es
// transversal: la regla es siempre la misma —**solo archivos referenciados por
// contenido publicado y visible hoy**— y la resuelve la consulta compartida
// `IsFilePublished` (internal/db/queries/media.sql), que cubre la identidad de la
// portada (F3) y los eventos y actividades de agenda (F4). El cambio respecto a
// F3 es de **ubicación** (y de que la comprobación añade "y próximo"): el
// handler, las cabeceras y los errores son los mismos.
//
// Capas handler → service → repository (§II): el handler solo fija las cabeceras
// seguras, el service aplica la política (400/404) y el repository es el único
// que conoce pgx.
package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/storage"
)

// Formatos del corte de "hoy"/"ahora" que viajan parametrizados al SQL (R4-2).
const (
	dateFormat = "2006-01-02"
	timeFormat = "15:04"
)

// Repository es el puerto de datos que el service de media necesita. Lo declara
// quien lo consume (arq. R3) y lo implementa *repository (repository.go).
type Repository interface {
	// IsFilePublished indica si el nombre está referenciado por contenido
	// publicado Y visible hoy: identidad de portada publicada (F3) o un
	// evento/actividad publicados con fecha próxima (F4). `today`/`nowTime` son
	// el corte del lugar, calculado por el service en APP_TZ (R4-2).
	IsFilePublished(ctx context.Context, fileName, today, nowTime string) (bool, error)
}

// ServiceDeps agrupa las dependencias del service.
type ServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository Repository
	// Store es el almacén de imágenes (platform/storage). Puede ser nil en
	// pruebas que no abran archivos.
	Store storage.Store
	// Location es la zona horaria del lugar (APP_TZ, T409) con la que se
	// calcula el corte "hoy"/"ahora" (R4-2). nil usa UTC (cmd/api siempre pasa
	// la zona validada de la configuración).
	Location *time.Location
	// Now es el reloj (inyectable en pruebas). nil usa time.Now.
	Now func() time.Time
	// Logger registra los fallos best-effort; nil usa el logger por defecto.
	Logger *slog.Logger
}

// service implementa la política de descarga pública de imágenes.
type service struct {
	repository Repository
	store      storage.Store
	location   *time.Location
	now        func() time.Time
	logger     *slog.Logger
}

// NewService construye el service de media. Si Logger es nil se usa el logger
// por defecto; si Location es nil se usa UTC.
func NewService(deps ServiceDeps) *service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	location := deps.Location
	if location == nil {
		location = time.UTC
	}
	now := deps.Now
	if now == nil {
		now = time.Now
	}
	return &service{
		repository: deps.Repository,
		store:      deps.Store,
		location:   location,
		now:        now,
		logger:     logger,
	}
}

// OpenMedia aplica la política de descarga (analyze C4/M6 y R4-6): un nombre que
// no cumple el patrón generado por el servidor es `400 invalid`; un nombre
// válido pero inexistente o NO referenciado por contenido publicado y próximo es
// `404 not_found`; solo entonces se abre el archivo del Store. La imagen de un
// elemento vencido deja de servirse **sin borrar el archivo** (decisión firme de
// la puerta 2) y vuelve a servirse al cargarle una fecha próxima.
func (s *service) OpenMedia(ctx context.Context, fileName string) (io.ReadCloser, error) {
	if !storage.ValidName(fileName) {
		return nil, apperr.Invalid(
			"El nombre del archivo no es válido",
			apperr.WithDetails(map[string]any{"fileName": "formato no reconocido"}),
		)
	}
	today, nowTime := s.cutoff()
	published, err := s.repository.IsFilePublished(ctx, fileName, today, nowTime)
	if err != nil {
		return nil, fmt.Errorf("comprobar si la imagen está publicada: %w", err)
	}
	if !published {
		return nil, apperr.NotFound("La imagen no está disponible")
	}
	if s.store == nil {
		return nil, apperr.NotFound("La imagen no está disponible")
	}
	reader, err := s.store.Open(ctx, fileName)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, apperr.NotFound("La imagen no está disponible")
		}
		return nil, fmt.Errorf("abrir la imagen: %w", err)
	}
	return reader, nil
}

// cutoff devuelve el "hoy" y el "ahora" del corte de la regla R4-2 en la zona
// del lugar (APP_TZ): son parámetros del SQL, nunca SQL construido.
func (s *service) cutoff() (string, string) {
	in := s.now().In(s.location)
	return in.Format(dateFormat), in.Format(timeFormat)
}
