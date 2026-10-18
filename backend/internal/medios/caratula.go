package medios

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"simiente-santa/backend/internal/platform/storage"
)

// caratula.go copia al sitio la carátula de un elemento importado (P9-7/R9-8).
// Es la razón por la que F9 necesita la rama nueva de `IsFilePublished`: la
// carátula se sirve como `/api/v1/media/<file>` y el navegador del visitante
// **no** pide nada a YouTube ni a Spotify (SC-005).
//
// La descarga es **acotada** (primera red saliente del backend, RG9-4):
//
//   - las cuatro barreras las aplica el cliente de la **fábrica única**
//     (T907A/P9-20): solo `https` y host de la allowlist (`i.ytimg.com`,
//     `d3t3ozftmdmh3i.cloudfront.net`), sin IPs privadas/loopback tras resolver
//     DNS, sin seguir redirecciones fuera de la allowlist y `io.LimitReader` con
//     el tope `UPLOAD_MAX_BYTES`;
//   - la **firma binaria** la valida `storage.Store.Save` (`storage.Detect`:
//     solo JPEG/PNG/WebP; SVG, GIF y HTML renombrado se rechazan), con el
//     nombre `img_<uuid>.<ext>` generado por el servidor y escritura atómica.
//
// Un fallo de descarga **nunca** tira la importación: el elemento entra **sin
// carátula** y el cliente pinta el marcador neutro (default revisable #4).

// Errores tipados de la descarga de carátulas (el resumen de la búsqueda nunca
// expone su texto técnico).
var (
	errCaratulaNoDisponible    = errors.New("medios: la carátula no se pudo descargar")
	errCaratulaDemasiadoGrande = errors.New("medios: la carátula supera el tope de tamaño")
)

const (
	// hostCaratulaYouTube y hostCaratulaPodcast son la allowlist de carátulas
	// (P9-7/R9-8): `i.ytimg.com` (YouTube) y el CDN del podcast (verificado en
	// R9-5). Si un feed trae otro host: sin carátula, marcador neutro.
	hostCaratulaYouTube = "i.ytimg.com"
	hostCaratulaPodcast = "d3t3ozftmdmh3i.cloudfront.net"

	// tiempoEsperaCaratula es el tiempo máximo por descarga (P9-3: 10 s).
	tiempoEsperaCaratula = 10 * time.Second
	// maxCaratulasPorBusqueda es el tope duro de descargas de una búsqueda
	// (P9-3/P9-18: 30, constante revisable). Lo aplica la búsqueda (T921).
	maxCaratulasPorBusqueda = 30
)

// hostsCaratula es la allowlist del descargador de carátulas.
var hostsCaratula = []string{hostCaratulaYouTube, hostCaratulaPodcast}

// CoverFetcher descarga la carátula de un elemento y la guarda en el sitio
// (P9-7/R9-8). Es el puerto que consume la búsqueda (T921) y el punto donde las
// pruebas inyectan un doble: **se declara aquí** y T915 no lo redeclara (M1).
type CoverFetcher interface {
	// FetchCover descarga `caratulaURL` y la guarda con `platform/storage`;
	// devuelve el nombre generado por el servidor. Un error significa que la
	// ficha entra **sin carátula** (marcador neutro): la importación nunca se
	// aborta por una carátula.
	FetchCover(ctx context.Context, caratulaURL string) (string, error)
}

// descargadorCaratulas es la implementación real de CoverFetcher.
type descargadorCaratulas struct {
	cliente *clienteSaliente
	store   storage.Store
}

// nuevoDescargadorCaratulas construye el descargador real (P9-7) con el cliente
// de la fábrica única y el almacén de archivos existente (F3/F4/F6).
// `maxBytes` es `UPLOAD_MAX_BYTES`: el mismo tope que aplica `storage.Store`.
func nuevoDescargadorCaratulas(store storage.Store, maxBytes int64) (*descargadorCaratulas, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("%w: el tope de carátulas (UPLOAD_MAX_BYTES) no puede ser 0", errConfiguracionCliente)
	}
	cliente, err := nuevoClienteSaliente(configuracionCliente{
		nombre:          "caratulas",
		hostsPermitidos: hostsCaratula,
		limiteCuerpo:    maxBytes,
		tiempoEspera:    tiempoEsperaCaratula,
	})
	if err != nil {
		return nil, err
	}
	return nuevoDescargadorCaratulasConCliente(cliente, store), nil
}

// nuevoDescargadorCaratulasConCliente permite inyectar la fábrica en las
// pruebas.
func nuevoDescargadorCaratulasConCliente(cliente *clienteSaliente, store storage.Store) *descargadorCaratulas {
	return &descargadorCaratulas{cliente: cliente, store: store}
}

// FetchCover descarga la carátula con las barreras de la fábrica única y la
// guarda con `platform/storage` (firma binaria y nombre generado por el
// servidor). Nunca devuelve un nombre a medias: o hay archivo guardado, o hay
// error y la ficha entra sin carátula.
func (d *descargadorCaratulas) FetchCover(ctx context.Context, caratulaURL string) (string, error) {
	resp, err := d.cliente.obtener(ctx, caratulaURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errCaratulaNoDisponible, err)
	}
	if resp.Estado != http.StatusOK {
		return "", fmt.Errorf("%w: el origen respondió %d", errCaratulaNoDisponible, resp.Estado)
	}
	if resp.Truncado {
		// El cuerpo llegó al tope: no se leyó entero y jamás se guarda un
		// archivo cortado (R9-8).
		return "", fmt.Errorf("%w: %d bytes (tope %d)", errCaratulaDemasiadoGrande, len(resp.Cuerpo), d.cliente.limiteCuerpo)
	}

	archivo, err := d.store.Save(ctx, resp.Cuerpo)
	if err != nil {
		// Aquí se rechazan las firmas no permitidas (SVG, GIF, HTML renombrado)
		// y cualquier fallo del almacén: la ficha entra sin carátula.
		return "", fmt.Errorf("%w: %w", errCaratulaNoDisponible, err)
	}
	return archivo.Name, nil
}
