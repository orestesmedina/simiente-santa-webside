package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// LocalStore guarda los archivos en un directorio local (`UPLOAD_DIR`) con
// nombres generados por el servidor. No conoce HTTP ni SQL y no usa estado
// global: el directorio y el límite de tamaño entran por constructor.
type LocalStore struct {
	root     string
	maxBytes int64
}

// NewLocalStore construye el almacén sobre root. maxBytes es el tamaño máximo
// aceptado por subida (UPLOAD_MAX_BYTES).
func NewLocalStore(root string, maxBytes int64) *LocalStore {
	return &LocalStore{root: root, maxBytes: maxBytes}
}

// Save valida la firma y el tamaño, genera `img_<uuid>.<ext>` y escribe el
// archivo de forma atómica (temp + rename) para no dejar una imagen a medias si
// el proceso muere durante la escritura.
func (s *LocalStore) Save(ctx context.Context, data []byte) (File, error) {
	if err := ctx.Err(); err != nil {
		return File{}, err
	}
	if int64(len(data)) > s.maxBytes {
		return File{}, fmt.Errorf("save: %d bytes superan el límite de %d: %w", len(data), s.maxBytes, ErrTooLarge)
	}

	mediaType, ext, ok := Detect(data)
	if !ok {
		return File{}, fmt.Errorf("save: firma binaria no permitida: %w", ErrInvalidFormat)
	}

	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return File{}, fmt.Errorf("save: crear directorio %q: %w", s.root, err)
	}

	name := fmt.Sprintf("img_%s.%s", uuid.NewString(), ext)
	if err := writeAtomic(s.root, name, data); err != nil {
		return File{}, fmt.Errorf("save %q: %w", name, err)
	}

	return File{Name: name, ContentType: mediaType, SizeBytes: int64(len(data))}, nil
}

// Open devuelve el contenido de un archivo con nombre válido. Valida el patrón
// ANTES de componer la ruta: `../`, rutas absolutas y nombres de cliente se
// rechazan con ErrInvalidName sin tocar el disco.
func (s *LocalStore) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !ValidName(name) {
		return nil, fmt.Errorf("open %q: %w", name, ErrInvalidName)
	}

	f, err := os.Open(filepath.Join(s.root, name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("open %q: %w", name, ErrNotFound)
		}
		return nil, fmt.Errorf("open %q: %w", name, err)
	}
	return f, nil
}

// Delete elimina un archivo con nombre válido. Es idempotente: si el archivo ya
// no existe, no devuelve error (la reposición de imagen borra la anterior
// best-effort, R3-8).
func (s *LocalStore) Delete(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !ValidName(name) {
		return fmt.Errorf("delete %q: %w", name, ErrInvalidName)
	}

	if err := os.Remove(filepath.Join(s.root, name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete %q: %w", name, err)
	}
	return nil
}

// writeAtomic escribe data en un archivo temporal del mismo directorio y lo
// renombra a name: el rename dentro del mismo sistema de archivos es atómico,
// así que nunca se observa un archivo parcial.
func writeAtomic(root, name string, data []byte) error {
	tmp, err := os.CreateTemp(root, ".tmp-upload-*")
	if err != nil {
		return fmt.Errorf("crear temporal: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op si el rename tuvo éxito

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("escribir temporal: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cerrar temporal: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(root, name)); err != nil {
		return fmt.Errorf("renombrar: %w", err)
	}
	return nil
}
