package medios

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"simiente-santa/backend/internal/platform/storage"
)

// Imágenes mínimas con firma binaria real (http.DetectContentType).
var (
	jpegMinimo = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xD9}
	pngMinimo  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R', 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89}
	webpMinimo = []byte{'R', 'I', 'F', 'F', 0x24, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P', 'V', 'P', '8', ' ', 0x18, 0x00, 0x00, 0x00, 0x30, 0x01, 0x00, 0x9D, 0x01, 0x2A, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00}
	gifMinimo  = []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xFF\xFF\xFF!\xF9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")
	svgMinimo  = []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"><rect width="1" height="1"/></svg>`)
	htmlFalso  = []byte("<!DOCTYPE html><html><body>no soy una imagen</body></html>")
)

const (
	caratulaYouTube = "https://i.ytimg.com/vi/SimienteAA1/hqdefault.jpg"
	caratulaPodcast = "https://d3t3ozftmdmh3i.cloudfront.net/staging/podcast_uploaded_nologo/x.jpg"
	maxBytesPrueba  = 1 << 20
)

// fakeStore es el doble de `storage.Store` para forzar fallos de guardado y
// registrar lo que recibe. El camino feliz de las pruebas usa el `LocalStore`
// real sobre un directorio temporal: hermético y con la firma binaria de verdad.
type fakeStore struct {
	ultimo []byte
	err    error
	veces  int
}

func (f *fakeStore) Save(_ context.Context, data []byte) (storage.File, error) {
	f.veces++
	f.ultimo = data
	if f.err != nil {
		return storage.File{}, f.err
	}
	return storage.File{Name: "img_11111111-1111-1111-1111-111111111111.jpg", ContentType: "image/jpeg", SizeBytes: int64(len(data))}, nil
}

func (f *fakeStore) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, storage.ErrNotFound
}

func (f *fakeStore) Delete(context.Context, string) error { return nil }

var _ storage.Store = (*fakeStore)(nil)

// almacenDePrueba devuelve el LocalStore real sobre un directorio temporal y su
// ruta (para comprobar en disco lo que se guardó).
func almacenDePrueba(t *testing.T) (*storage.LocalStore, string) {
	t.Helper()
	dir := t.TempDir()
	return storage.NewLocalStore(dir, maxBytesPrueba), dir
}

// descargadorDePrueba arma el descargador con un transporte falso y el almacén
// indicado (sin red).
func descargadorDePrueba(t *testing.T, transporte *transporteFalso, almacen storage.Store) *descargadorCaratulas {
	t.Helper()
	cliente := clienteDeConsumidorConLimite(t, transporte, "caratulas", hostsCaratula, maxBytesPrueba)
	return nuevoDescargadorCaratulasConCliente(cliente, almacen)
}

// clienteDeConsumidorConLimite es `clienteDeConsumidor` con un tope de cuerpo
// explícito (el de las carátulas es `UPLOAD_MAX_BYTES`).
func clienteDeConsumidorConLimite(t *testing.T, transporte http.RoundTripper, nombre string, hosts []string, limite int64) *clienteSaliente {
	t.Helper()
	cliente, err := nuevoClienteSaliente(configuracionCliente{
		nombre:          nombre,
		hostsPermitidos: hosts,
		limiteCuerpo:    limite,
		tiempoEspera:    50 * time.Millisecond,
		transporte:      transporte,
		resolver:        resolverFalso("93.184.216.34"),
		dial:            (&dialFalso{}).dial,
	})
	if err != nil {
		t.Fatalf("nuevoClienteSaliente(%s) = %v", nombre, err)
	}
	return cliente
}

// TestFetchCoverDescargaYGuarda fija el camino feliz (P9-7/R9-8): la carátula se
// descarga por la fábrica única, se guarda con `platform/storage` (nombre
// `img_<uuid>.<ext>` generado por el servidor) y los bytes llegan **intactos**.
func TestFetchCoverDescargaYGuarda(t *testing.T) {
	transporte := &transporteFalso{}
	transporte.responder = func(*http.Request) (*http.Response, error) {
		return respuestaFalsa(http.StatusOK, io.NopCloser(bytes.NewReader(jpegMinimo)), nil), nil
	}
	almacen, dir := almacenDePrueba(t)
	descargador := descargadorDePrueba(t, transporte, almacen)

	nombre, err := descargador.FetchCover(context.Background(), caratulaYouTube)
	if err != nil {
		t.Fatalf("FetchCover() = %v, se esperaba nil", err)
	}
	if !storage.ValidName(nombre) || !strings.HasSuffix(nombre, ".jpg") {
		t.Fatalf("nombre = %q, se esperaba un nombre generado por el servidor", nombre)
	}
	guardado, err := os.ReadFile(dir + "/" + nombre)
	if err != nil {
		t.Fatalf("leer el archivo guardado: %v", err)
	}
	if !bytes.Equal(guardado, jpegMinimo) {
		t.Errorf("los bytes guardados no son idénticos a los descargados")
	}
	if got := transporte.ultima().Header.Get("User-Agent"); got != usuarioAgente {
		t.Errorf("User-Agent = %q, se esperaba %q", got, usuarioAgente)
	}
}

// TestFetchCoverFormatosValidos fija que PNG y WebP también se copian (R9-8: la
// carátula del podcast es JPEG y las de YouTube pueden ser WebP).
func TestFetchCoverFormatosValidos(t *testing.T) {
	casos := map[string][]byte{"jpg": jpegMinimo, "png": pngMinimo, "webp": webpMinimo}
	for extension, bytesImagen := range casos {
		t.Run(extension, func(t *testing.T) {
			transporte := &transporteFalso{}
			transporte.responder = func(*http.Request) (*http.Response, error) {
				return respuestaFalsa(http.StatusOK, io.NopCloser(bytes.NewReader(bytesImagen)), nil), nil
			}
			almacen, _ := almacenDePrueba(t)
			descargador := descargadorDePrueba(t, transporte, almacen)

			nombre, err := descargador.FetchCover(context.Background(), caratulaYouTube)
			if err != nil {
				t.Fatalf("FetchCover() = %v", err)
			}
			if !strings.HasSuffix(nombre, "."+extension) {
				t.Errorf("nombre = %q, se esperaba extensión %s", nombre, extension)
			}
		})
	}
}

// TestFetchCoverRechazaFirmaNoPermitida fija la firma binaria (R9-8): SVG, GIF,
// HTML renombrado o cualquier otro contenido **no** se guarda (y el archivo no
// llega a existir), aunque el origen responda `200`.
func TestFetchCoverRechazaFirmaNoPermitida(t *testing.T) {
	casos := map[string][]byte{"svg": svgMinimo, "gif": gifMinimo, "html": htmlFalso, "vacio": {}}
	for nombre, bytesImagen := range casos {
		t.Run(nombre, func(t *testing.T) {
			transporte := &transporteFalso{}
			transporte.responder = func(*http.Request) (*http.Response, error) {
				return respuestaFalsa(http.StatusOK, io.NopCloser(bytes.NewReader(bytesImagen)), nil), nil
			}
			almacen, dir := almacenDePrueba(t)
			descargador := descargadorDePrueba(t, transporte, almacen)

			if _, err := descargador.FetchCover(context.Background(), caratulaYouTube); !errors.Is(err, storage.ErrInvalidFormat) {
				t.Fatalf("FetchCover(%s) = %v, se esperaba %v", nombre, err, storage.ErrInvalidFormat)
			}
			archivos, err := os.ReadDir(dir)
			if err == nil && len(archivos) != 0 {
				t.Errorf("se guardó un archivo con firma no permitida: %v", archivos)
			}
		})
	}
}

// TestFetchCoverNoDisponible fija que una descarga que no responde `200` es un
// error de la descarga: la ficha entra **sin carátula** (marcador neutro) y la
// importación nunca se aborta por eso (default #4/R9-8).
func TestFetchCoverNoDisponible(t *testing.T) {
	for _, estado := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError, http.StatusTooManyRequests} {
		transporte := &transporteFalso{}
		transporte.responder = func(*http.Request) (*http.Response, error) {
			return respuestaFalsa(estado, io.NopCloser(strings.NewReader("")), nil), nil
		}
		almacen, _ := almacenDePrueba(t)
		descargador := descargadorDePrueba(t, transporte, almacen)

		nombre, err := descargador.FetchCover(context.Background(), caratulaYouTube)
		if err == nil {
			t.Fatalf("estado %d: FetchCover() = %q, se esperaba un error", estado, nombre)
		}
		if nombre != "" {
			t.Errorf("estado %d: nombre = %q, se esperaba vacío", estado, nombre)
		}
	}
}

// TestFetchCoverDemasiadoGrande fija el tope de tamaño (R9-8): una carátula por
// encima de `UPLOAD_MAX_BYTES` se rechaza **sin leer el cuerpo entero** y sin
// guardar nada.
func TestFetchCoverDemasiadoGrande(t *testing.T) {
	transporte := &transporteFalso{}
	cuerpo := &cuerpoInfinito{tope: maxBytesPrueba}
	transporte.responder = func(*http.Request) (*http.Response, error) {
		return respuestaFalsa(http.StatusOK, cuerpo, nil), nil
	}
	almacen, dir := almacenDePrueba(t)
	descargador := descargadorDePrueba(t, transporte, almacen)

	if _, err := descargador.FetchCover(context.Background(), caratulaYouTube); err == nil {
		t.Fatalf("FetchCover(cuerpo enorme) = nil, se esperaba un error")
	}
	if cuerpo.leidos != cuerpo.tope {
		t.Errorf("se leyeron %d bytes, se esperaba no pasar del tope %d", cuerpo.leidos, cuerpo.tope)
	}
	archivos, err := os.ReadDir(dir)
	if err == nil && len(archivos) != 0 {
		t.Errorf("se guardó una carátula por encima del tope: %v", archivos)
	}
}

// TestFetchCoverRechazaHostHttpYRedirectAjeno fija las barreras 1 y 3 de la
// fábrica única sobre la descarga: solo `https` y hosts de la allowlist
// (`i.ytimg.com`, `d3t3ozftmdmh3i.cloudfront.net`), también tras un redirect.
func TestFetchCoverRechazaHostHttpYRedirectAjeno(t *testing.T) {
	casos := []struct {
		nombre   string
		destino  string
		location string
		wantErr  error
	}{
		{nombre: "host no permitido", destino: "https://evil.example/caratula.jpg", wantErr: errHostNoPermitido},
		{nombre: "http rechazado", destino: "http://i.ytimg.com/vi/x/hqdefault.jpg", wantErr: errEsquemaNoPermitido},
		{nombre: "redirect a host ajeno", destino: caratulaYouTube, location: "https://evil.example/caratula.jpg", wantErr: errHostNoPermitido},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			transporte := &transporteFalso{}
			transporte.responder = func(*http.Request) (*http.Response, error) {
				if caso.location != "" {
					return respuestaFalsa(http.StatusFound, nil, map[string]string{"Location": caso.location}), nil
				}
				return respuestaFalsa(http.StatusOK, io.NopCloser(bytes.NewReader(jpegMinimo)), nil), nil
			}
			almacen, _ := almacenDePrueba(t)
			descargador := descargadorDePrueba(t, transporte, almacen)

			if _, err := descargador.FetchCover(context.Background(), caso.destino); !errors.Is(err, caso.wantErr) {
				t.Fatalf("FetchCover(%q) = %v, se esperaba %v", caso.destino, err, caso.wantErr)
			}
		})
	}
}

// TestFetchCoverRechazaIPPrivada fija la barrera 2 sobre la descarga: aunque el
// host esté en la allowlist, una resolución a una IP privada, loopback o de
// metadatos (`169.254.169.254`) no se conecta.
func TestFetchCoverRechazaIPPrivada(t *testing.T) {
	for _, ip := range []string{"169.254.169.254", "127.0.0.1", "10.1.2.3", "::1"} {
		transporte := &transporteFalso{}
		almacen, _ := almacenDePrueba(t)
		descargador := descargadorDePrueba(t, transporte, almacen)
		descargador.cliente.resolver = resolverFalso(ip)

		if _, err := descargador.cliente.verificarDestino(context.Background(), hostCaratulaYouTube); !errors.Is(err, errIPNoPermitida) {
			t.Errorf("verificarDestino(%s) = %v, se esperaba %v", ip, err, errIPNoPermitida)
		}
	}
}

// TestFetchCoverGuardadoFalla fija que un fallo del almacén también deja la
// ficha sin carátula (nunca aborta la importación).
func TestFetchCoverGuardadoFalla(t *testing.T) {
	transporte := &transporteFalso{}
	transporte.responder = func(*http.Request) (*http.Response, error) {
		return respuestaFalsa(http.StatusOK, io.NopCloser(bytes.NewReader(jpegMinimo)), nil), nil
	}
	almacen := &fakeStore{err: errors.New("disco lleno")}
	descargador := descargadorDePrueba(t, transporte, almacen)

	nombre, err := descargador.FetchCover(context.Background(), caratulaYouTube)
	if err == nil {
		t.Fatalf("FetchCover() = %q, se esperaba un error del almacén", nombre)
	}
	if almacen.veces != 1 || !bytes.Equal(almacen.ultimo, jpegMinimo) {
		t.Errorf("el almacén recibió %d bytes intactos: %v", len(almacen.ultimo), almacen.veces)
	}
}

// TestFetchCoverDescargadorBarreras ejecuta la tabla compartida de barreras
// (T907A/P9-20) sobre la descarga de carátulas: la MISMA tabla que los otros dos
// consumidores.
func TestFetchCoverDescargadorBarreras(t *testing.T) {
	transporte := &transporteFalso{}
	cliente := clienteDeConsumidorConLimite(t, transporte, "caratulas", hostsCaratula, maxBytesPrueba)
	almacen, _ := almacenDePrueba(t)
	_ = nuevoDescargadorCaratulasConCliente(cliente, almacen)

	ejecutarBarrerasCompartidas(t, casoBarreras{
		nombre:           "caratulas",
		cliente:          cliente,
		transporte:       transporte,
		destinoPermitido: caratulaPodcast,
		destinoAjeno:     "https://evil.example/caratula.jpg",
	})
}

// TestDescargadorImplementaCoverFetcher deja explícita la interfaz del dominio
// (M1: se declara aquí y T915 no la redeclara).
func TestDescargadorImplementaCoverFetcher(t *testing.T) {
	var descargador CoverFetcher = (*descargadorCaratulas)(nil)
	if descargador == nil {
		t.Fatalf("el descargador no implementa CoverFetcher")
	}
}

// fakeCoverFetcher es el doble de `CoverFetcher` que usarán las pruebas del
// servicio (T921: las carátulas se prueban sin red).
type fakeCoverFetcher struct {
	nombres map[string]string
	err     error
	pedidas []string
}

func (f *fakeCoverFetcher) FetchCover(_ context.Context, caratulaURL string) (string, error) {
	f.pedidas = append(f.pedidas, caratulaURL)
	if f.err != nil {
		return "", f.err
	}
	return f.nombres[caratulaURL], nil
}

var _ CoverFetcher = (*fakeCoverFetcher)(nil)

// TestFakeCoverFetcher fija el doble que usarán las pruebas del servicio.
func TestFakeCoverFetcher(t *testing.T) {
	fake := &fakeCoverFetcher{nombres: map[string]string{caratulaYouTube: "img_x.jpg"}}
	if got, err := fake.FetchCover(context.Background(), caratulaYouTube); err != nil || got != "img_x.jpg" {
		t.Errorf("FetchCover() = %q, %v", got, err)
	}
	if got, _ := fake.FetchCover(context.Background(), "https://otra/x.jpg"); got != "" {
		t.Errorf("FetchCover(desconocida) = %q, se esperaba vacío", got)
	}
}
