package medios

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// disponibilidad.go comprueba si el enlace de un elemento que salió de la
// ventana de su feed sigue disponible (FR-017, research R9-2/P9-6). Son dos
// comprobaciones **gratuitas y sin claves**, ambas sobre el cliente de la
// fábrica única (T907A/P9-20): el enlace del episodio lo trae el feed, así que
// las barreras de red son obligatorias (corrección A1 de la fase 5).
//
//   - **YouTube**: `GET https://www.youtube.com/oembed?url=<video>&format=json`
//     (`200` → disponible; `404` → **no disponible confirmado**).
//   - **Podcast**: `GET` del enlace del episodio devuelto por el feed
//     (`200` → disponible; `404`/`410` → **no disponible confirmado**). El
//     oEmbed de Spotify **no** sirve para nuestros enlaces (`podcasters.
//     spotify.com`): responde `504` (verificado).
//
// La regla es **a prueba de fallos** (FR-017/SC-014): **solo** una confirmación
// (`404`/`410`) dice `unavailable` y solo ella puede ocultar una ficha;
// `inconclusive` —tiempo agotado, `5xx`, `429`, `403`, error de red o cualquier
// duda— **nunca** oculta nada. Un probador caído no vacía el sitio.

const (
	// urlOEmbedYouTube es el endpoint público de oEmbed de YouTube (R9-2).
	urlOEmbedYouTube = "https://www.youtube.com/oembed"
	// hostOEmbedYouTube es el host de la allowlist del probador de YouTube.
	hostOEmbedYouTube = "www.youtube.com"
	// hostEpisodioPodcast es el host de los enlaces del podcast (ya validado al
	// importar; el probador solo sale a ese host).
	hostEpisodioPodcast = "podcasters.spotify.com"

	// tiempoEsperaProbador es el tiempo máximo por comprobación (P9-3: 5 s).
	tiempoEsperaProbador = 5 * time.Second
	// limiteCuerpoProbador es el tope de bytes del cuerpo (barrera 4 de P9-20):
	// de la comprobación solo interesa el código de estado.
	limiteCuerpoProbador = 64 << 10 // 64 KiB
)

// hostsProbador es la allowlist del probador: el oEmbed de YouTube y el host de
// los episodios del podcast.
var hostsProbador = []string{hostOEmbedYouTube, hostEpisodioPodcast}

// ResultadoDisponibilidad es el resultado de comprobar un enlace (FR-017).
//
// T911 (`model.go`) **reutiliza** este tipo y no lo redeclara (M1).
type ResultadoDisponibilidad string

const (
	// DisponibilidadDisponible: el enlace responde `200`.
	DisponibilidadDisponible ResultadoDisponibilidad = "available"
	// DisponibilidadNoDisponible: la comprobación **confirmó** que el enlace ya
	// no está (`404`/`410`). Es el único resultado que oculta una ficha.
	DisponibilidadNoDisponible ResultadoDisponibilidad = "unavailable"
	// DisponibilidadInconclusa: no se pudo comprobar (tiempo agotado, `5xx`,
	// error de red, duda). **Nunca** oculta nada.
	DisponibilidadInconclusa ResultadoDisponibilidad = "inconclusive"
)

// AvailabilityProber comprueba si el enlace de un elemento sigue disponible
// (FR-017/P9-6). Es el puerto que consume la búsqueda (T921) y el punto donde
// las pruebas inyectan un doble: **se declara aquí** y T915 no lo redeclara
// (M1).
//
// El error de retorno es **solo informativo** (el servicio lo deja en el log):
// el resultado ya es utilizable y la regla a prueba de fallos manda, de modo que
// un error **nunca** cambia el desenlace de la búsqueda.
type AvailabilityProber interface {
	ProbeYouTube(ctx context.Context, videoID string) (ResultadoDisponibilidad, error)
	ProbePodcast(ctx context.Context, enlace string) (ResultadoDisponibilidad, error)
}

// probadorDisponibilidad es la implementación real de AvailabilityProber.
type probadorDisponibilidad struct {
	cliente *clienteSaliente
}

// nuevoProbadorDisponibilidad construye el probador real con el cliente de la
// fábrica única (P9-20).
func nuevoProbadorDisponibilidad() (*probadorDisponibilidad, error) {
	cliente, err := nuevoClienteSaliente(configuracionCliente{
		nombre:          "probador",
		hostsPermitidos: hostsProbador,
		limiteCuerpo:    limiteCuerpoProbador,
		tiempoEspera:    tiempoEsperaProbador,
	})
	if err != nil {
		return nil, err
	}
	return nuevoProbadorDisponibilidadConCliente(cliente), nil
}

// probadorDisponibilidad implementa el puerto del dominio (M1).
var _ AvailabilityProber = (*probadorDisponibilidad)(nil)

// nuevoProbadorDisponibilidadConCliente permite inyectar la fábrica en las
// pruebas.
func nuevoProbadorDisponibilidadConCliente(cliente *clienteSaliente) *probadorDisponibilidad {
	return &probadorDisponibilidad{cliente: cliente}
}

// ProbeYouTube comprueba el video con el oEmbed público de YouTube (R9-2):
// `200` → disponible; `404` → **no disponible confirmado**; cualquier otra
// respuesta (`400`, `401`, `403`, `429`, `5xx`, tiempo agotado o error de red) →
// inconcluso.
func (p *probadorDisponibilidad) ProbeYouTube(ctx context.Context, videoID string) (ResultadoDisponibilidad, error) {
	video := "https://" + hostOEmbedYouTube + "/watch?v=" + url.QueryEscape(videoID)
	destino := urlOEmbedYouTube + "?url=" + url.QueryEscape(video) + "&format=json"

	resultado, err := p.obtener(ctx, destino)
	if err != nil {
		return DisponibilidadInconclusa, fmt.Errorf("medios: comprobar el video: %w", err)
	}
	return resultado, nil
}

// ProbePodcast comprueba el enlace del episodio que trae el feed (R9-2):
// `200` → disponible; `404`/`410` → **no disponible confirmado**; el resto →
// inconcluso.
func (p *probadorDisponibilidad) ProbePodcast(ctx context.Context, enlace string) (ResultadoDisponibilidad, error) {
	resultado, err := p.obtener(ctx, enlace)
	if err != nil {
		return DisponibilidadInconclusa, fmt.Errorf("medios: comprobar el episodio: %w", err)
	}
	return resultado, nil
}

// obtener hace la petición y traduce el código de estado a resultado. El error
// de la fábrica (host rechazado, red, tiempo agotado) se devuelve junto al
// resultado inconcluso: es informativo y nunca oculta.
func (p *probadorDisponibilidad) obtener(ctx context.Context, destino string) (ResultadoDisponibilidad, error) {
	resp, err := p.cliente.obtener(ctx, destino)
	if err != nil {
		return DisponibilidadInconclusa, err
	}
	switch resp.Estado {
	case http.StatusOK:
		return DisponibilidadDisponible, nil
	case http.StatusNotFound, http.StatusGone:
		// Confirmado: el enlace ya no está. Es el único caso que oculta
		// (FR-017/SC-014) y aplica a los dos proveedores: YouTube devuelve 404
		// y el podcast 404/410.
		return DisponibilidadNoDisponible, nil
	default:
		return DisponibilidadInconclusa, nil
	}
}
