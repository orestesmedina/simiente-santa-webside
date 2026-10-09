package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// newStore construye un LocalStore sobre un directorio temporal (la prueba no
// toca el disco del desarrollo). El límite por defecto es holgado salvo que la
// prueba lo ajuste.
func newStore(t *testing.T) *LocalStore {
	t.Helper()
	return NewLocalStore(t.TempDir(), 8*1024*1024)
}

func TestLocalStoreSaveGeneratesValidUniqueName(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	first, err := store.Save(ctx, jpegBytes)
	if err != nil {
		t.Fatalf("Save(jpeg): %v", err)
	}
	if !ValidName(first.Name) {
		t.Fatalf("Save() nombre = %q, no cumple el patrón del servidor", first.Name)
	}
	if first.ContentType != "image/jpeg" || first.SizeBytes != int64(len(jpegBytes)) {
		t.Fatalf("Save() = %+v, se esperaba jpeg de %d bytes", first, len(jpegBytes))
	}

	second, err := store.Save(ctx, pngBytes)
	if err != nil {
		t.Fatalf("Save(png): %v", err)
	}
	if first.Name == second.Name {
		t.Fatalf("Save() generó el mismo nombre dos veces: %q", first.Name)
	}
	if second.ContentType != "image/png" {
		t.Fatalf("Save(png) ContentType = %q", second.ContentType)
	}

	// El archivo está realmente en disco y con el contenido subido.
	onDisk, err := os.ReadFile(filepath.Join(store.root, first.Name))
	if err != nil {
		t.Fatalf("leer archivo guardado: %v", err)
	}
	if !bytes.Equal(onDisk, jpegBytes) {
		t.Fatal("el contenido guardado no coincide con el subido")
	}
}

func TestLocalStoreSaveRejectsUnsupportedAndTooLarge(t *testing.T) {
	t.Run("firma no permitida", func(t *testing.T) {
		store := newStore(t)
		if _, err := store.Save(context.Background(), svgBytes); !errors.Is(err, ErrInvalidFormat) {
			t.Fatalf("Save(svg) = %v, se esperaba ErrInvalidFormat", err)
		}
		entries, err := os.ReadDir(store.root)
		if err != nil {
			t.Fatalf("leer directorio: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("Save(svg) dejó %d archivos, se esperaba ninguno", len(entries))
		}
	})

	t.Run("demasiado grande", func(t *testing.T) {
		store := NewLocalStore(t.TempDir(), int64(len(jpegBytes)-1))
		if _, err := store.Save(context.Background(), jpegBytes); !errors.Is(err, ErrTooLarge) {
			t.Fatalf("Save(oversize) = %v, se esperaba ErrTooLarge", err)
		}
	})
}

func TestLocalStoreOpenRoundTrips(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	saved, err := store.Save(ctx, webpBytes)
	if err != nil {
		t.Fatalf("Save(webp): %v", err)
	}

	reader, err := store.Open(ctx, saved.Name)
	if err != nil {
		t.Fatalf("Open(%q): %v", saved.Name, err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("leer: %v", err)
	}
	if !bytes.Equal(got, webpBytes) {
		t.Fatal("Open() devolvió un contenido distinto al guardado")
	}
}

func TestLocalStoreOpenRejectsUnsafeNames(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	if _, err := store.Save(ctx, jpegBytes); err != nil {
		t.Fatalf("Save: %v", err)
	}

	for _, name := range []string{
		"../../etc/passwd",
		"../" + filepath.Base(store.root),
		"/etc/passwd",
		"img_abc.jpg",
		"img_0123456789abcdef0123456789abcdef0123.sh",
		"",
	} {
		t.Run(name, func(t *testing.T) {
			reader, err := store.Open(ctx, name)
			if !errors.Is(err, ErrInvalidName) {
				t.Fatalf("Open(%q) = %v, se esperaba ErrInvalidName", name, err)
			}
			if reader != nil {
				t.Fatalf("Open(%q) devolvió un reader con error", name)
			}
		})
	}
}

func TestLocalStoreOpenMissingReturnsErrNotFound(t *testing.T) {
	store := newStore(t)
	// Nombre válido pero inexistente.
	const missing = "img_0123456789abcdef0123456789abcdef0123.jpg"
	if _, err := store.Open(context.Background(), missing); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Open(inexistente) = %v, se esperaba ErrNotFound", err)
	}
}

func TestLocalStoreDeleteIsIdempotent(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	saved, err := store.Save(ctx, pngBytes)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := store.Delete(ctx, saved.Name); err != nil {
		t.Fatalf("Delete(existente): %v", err)
	}
	if _, err := store.Open(ctx, saved.Name); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Open tras Delete = %v, se esperaba ErrNotFound", err)
	}
	// Borrar de nuevo no falla (borrado best-effort de la imagen anterior).
	if err := store.Delete(ctx, saved.Name); err != nil {
		t.Fatalf("Delete(inexistente) = %v, se esperaba nil", err)
	}
}
