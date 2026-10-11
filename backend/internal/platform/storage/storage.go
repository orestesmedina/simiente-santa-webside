// Package storage es el plumbing de archivos de la aplicación: la interfaz
// Store, la detección de imagen por firma binaria y los errores tipados que
// consumen los dominios (F3: imágenes de la portada; reutilizable por F8/F9).
//
// No hace SQL, ni HTTP (más allá del sniffing de bytes de net/http), ni guarda
// estado global: el directorio entra por constructor (R6). Los nombres de
// archivo los genera SIEMPRE el servidor (`img_<uuid>.<ext>`); ningún camino
// acepta un nombre controlado por el usuario, de modo que el path traversal es
// imposible por diseño (R3-8/§IV).
package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Errores tipados del plumbing. Los consumidores los distinguen con
// errors.Is (p. ej. la descarga pública traduce ErrInvalidName a 400 y
// ErrNotFound a 404, analyze M6).
var (
	// ErrNotFound indica que el archivo (con nombre válido) no existe.
	ErrNotFound = errors.New("storage: archivo no encontrado")
	// ErrInvalidName indica que el nombre no cumple el patrón generado por el
	// servidor; se rechaza antes de tocar el disco.
	ErrInvalidName = errors.New("storage: nombre de archivo inválido")
	// ErrInvalidFormat indica que los bytes no tienen la firma de un tipo
	// permitido (JPEG/PNG/WebP). SVG y GIF se rechazan aquí.
	ErrInvalidFormat = errors.New("storage: formato de imagen no permitido")
	// ErrTooLarge indica que los bytes superan el límite configurado.
	ErrTooLarge = errors.New("storage: archivo demasiado grande")
)

// Tipos MIME permitidos y su extensión canónica (R3-8: solo JPEG, PNG y WebP;
// sin SVG ni GIF).
const (
	contentTypeJPEG = "image/jpeg"
	contentTypePNG  = "image/png"
	contentTypeWEBP = "image/webp"

	extJPEG = "jpg"
	extPNG  = "png"
	extWEBP = "webp"
)

// namePattern es el patrón EXACTO de los nombres generados por el servidor:
// `img_` + UUID en minúsculas (36 caracteres hex/dash) + extensión permitida.
// Open/Delete lo exigen antes de componer la ruta en disco.
var namePattern = regexp.MustCompile(`^img_[0-9a-f-]{36}\.(jpg|png|webp)$`)

// File describe un archivo guardado por el servidor.
type File struct {
	// Name es el nombre generado (`img_<uuid>.<ext>`); nunca el del cliente.
	Name string
	// ContentType es el tipo detectado por firma binaria, seguro para servir.
	ContentType string
	// SizeBytes es el tamaño real del contenido guardado.
	SizeBytes int64
}

// Store es el puerto de almacenamiento de archivos. Lo implementa LocalStore
// (disco local) y lo pueden implementar otros backends sin cambiar a los
// consumidores (R3-8, plan B con URL versionada o S3).
type Store interface {
	// Save valida la firma y el tamaño de data, genera un nombre único y
	// persiste el archivo. Devuelve ErrInvalidFormat o ErrTooLarge si no pasa
	// la política, sin escribir nada.
	Save(ctx context.Context, data []byte) (File, error)
	// Open devuelve el contenido de un archivo con nombre válido. Un nombre que
	// no cumple el patrón devuelve ErrInvalidName; uno válido e inexistente,
	// ErrNotFound.
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	// Delete elimina un archivo con nombre válido. Es idempotente: un archivo
	// que ya no existe no es un error.
	Delete(ctx context.Context, name string) error
}

// ValidName indica si name cumple el patrón generado por el servidor.
func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

// Detect clasifica los bytes por su firma binaria (http.DetectContentType) y
// devuelve el tipo MIME y la extensión cuando es JPEG, PNG o WebP. Cualquier
// otro contenido (SVG, GIF, HTML renombrado, etc.) devuelve ok=false.
func Detect(data []byte) (mediaType, ext string, ok bool) {
	switch http.DetectContentType(data) {
	case contentTypeJPEG:
		return contentTypeJPEG, extJPEG, true
	case contentTypePNG:
		return contentTypePNG, extPNG, true
	case contentTypeWEBP:
		return contentTypeWEBP, extWEBP, true
	default:
		return "", "", false
	}
}

// ContentTypeFor deduce el tipo MIME seguro a partir de la extensión de un
// nombre válido. Un nombre inválido o de extensión desconocida devuelve el tipo
// genérico binario (nunca se sirve como ejecutable/HTML).
func ContentTypeFor(name string) string {
	switch {
	case strings.HasSuffix(name, ".jpg"):
		return contentTypeJPEG
	case strings.HasSuffix(name, ".png"):
		return contentTypePNG
	case strings.HasSuffix(name, ".webp"):
		return contentTypeWEBP
	default:
		return "application/octet-stream"
	}
}
