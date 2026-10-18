package medios

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// feeds.go es la lectura de los dos feeds del cliente (FR-010) detrás de la
// interfaz falsable `FeedSource` (P9-4, M1: se declara AQUÍ y T915 no la
// redeclara). La red saliente va por el cliente de la fábrica única
// (T907A/P9-20) y el parseo es propio, con `encoding/xml` (R9-3): **cero
// dependencias nuevas**.
//
// Los dos formatos están fijados por los feeds reales (research R9-5):
//
//   - **Atom de YouTube**: `yt:videoId` (identificador externo), `title`,
//     `link rel="alternate"` (`watch?v=<11>`), `published` RFC3339 y
//     `media:group` con `media:thumbnail` (carátula en `i.ytimg.com`) y
//     `media:description` (texto, entidades ya decodificadas por el parser de
//     XML).
//   - **RSS 2.0 del podcast**: `guid` (identificador externo), `title`,
//     `description` con **HTML dentro de CDATA**, `link` (el enlace del
//     episodio), `pubDate` RFC1123 con `GMT` (o RFC1123Z con `+0000`) e
//     `itunes:image`.
//
// El contenido de un feed es **dato, nunca código** (constitución §IV/FR-023):
// las descripciones se publican como texto plano, sin etiquetas y con las
// entidades decodificadas. Un elemento con una fecha imposible, un enlace fuera
// del host del proveedor o sin título **se descarta con aviso** (nunca se
// inventa un dato); un XML roto es un **error claro** y nunca deja elementos a
// medias.

// Errores tipados de la lectura de feeds. `errFeedNoDisponible` es el que el
// resumen de la búsqueda convierte en `failed` (FR-016: un feed caído no oculta
// ni marca nada).
var (
	errXMLFeed          = errors.New("medios: el feed no es un XML válido")
	errFeedNoDisponible = errors.New("medios: el feed no respondió")
)

const (
	// Los dos orígenes de FR-010 (P9-18): públicos, gratuitos y sin claves.
	urlFeedYouTube = "https://www.youtube.com/feeds/videos.xml?channel_id=UCXXJbgowcGnjI7BDsPnKOTw"
	urlFeedPodcast = "https://anchor.fm/s/1116351e8/podcast/rss"

	// Hosts de la allowlist del lector de feeds (barrera 1 de P9-20).
	hostFeedYouTube = "www.youtube.com"
	hostFeedPodcast = "anchor.fm"

	// tiempoEsperaFeeds es el tiempo máximo por lectura de feed (P9-3).
	tiempoEsperaFeeds = 10 * time.Second
	// limiteCuerpoFeeds es el tope de bytes de un feed (barrera 4 de P9-20).
	limiteCuerpoFeeds = 2 << 20 // 2 MiB: un feed real ronda los 25 KiB.
)

// Formas de enlace admitidas, espejo exacto de los `CHECK` de la migración
// 000014 (defensa en profundidad: el parser valida y la BD lo garantiza).
var (
	patronEnlaceYouTube = regexp.MustCompile(`^https://(www\.)?youtube\.com/watch\?v=[A-Za-z0-9_-]{11}$`)
	patronEnlacePodcast = regexp.MustCompile(`^https://podcasters\.spotify\.com/`)
)

// MotivoAviso es la causa por la que un elemento del feed no se importa.
type MotivoAviso string

const (
	// MotivoFechaInvalida: la fecha del elemento no se pudo interpretar (nunca
	// se inventa una fecha: sin ella no hay orden ni unión segura).
	MotivoFechaInvalida MotivoAviso = "fecha_invalida"
	// MotivoEnlaceAjeno: el enlace no pertenece al host del proveedor.
	MotivoEnlaceAjeno MotivoAviso = "enlace_ajeno"
	// MotivoTituloVacio: el elemento no trae título (la ficha exige uno).
	MotivoTituloVacio MotivoAviso = "titulo_vacio"
	// MotivoIdentificadorVacio: el elemento no trae identificador externo (sin
	// él no hay idempotencia: R9-11).
	MotivoIdentificadorVacio MotivoAviso = "identificador_vacio"
)

// AvisoFeed describe un elemento del feed descartado y por qué: la búsqueda
// nunca pierde lo que pasó, aunque el elemento no se importe (RG9-11).
type AvisoFeed struct {
	// IDExterno del elemento descartado ("" si no se pudo leer).
	IDExterno string
	// Motivo es una causa estable, sin datos técnicos de red.
	Motivo MotivoAviso
}

// EntradaFeed es un elemento de feed ya validado (P9-4/R9-3): lo que la
// importación necesita para crear la ficha y su fuente. El proveedor lo conoce
// quien llamó a `FetchYouTube`/`FetchPodcast` (una entrada no viaja sin origen).
type EntradaFeed struct {
	// IDExterno es el identificador del elemento en su feed: `yt:videoId` en
	// YouTube y `guid` en el podcast (R9-5/R9-11). Es la llave de idempotencia.
	IDExterno string
	// Titulo es el título del feed, tal cual (FR-003: se publica sin normalizar
	// ni alterar; la normalización es solo la llave de la unión).
	Titulo string
	// Enlace es la URL del elemento, validada contra la forma del proveedor
	// (R9-8): `youtube.com/watch?v=<11>` o `podcasters.spotify.com/...`.
	Enlace string
	// Publicado es la fecha de publicación del feed.
	Publicado time.Time
	// CaratulaURL es la carátula del feed ("" si el feed no la trae: la ficha
	// entra sin carátula y el cliente pinta el marcador neutro).
	CaratulaURL string
	// Descripcion es la descripción del feed en TEXTO PLANO ("" si el feed no
	// trae descripción: la ficha omite la línea, precisión 3 de la puerta 2).
	// Los límites de longitud (300/120/500) los aplica la importación (T920).
	Descripcion string
}

// ResultadoFeed es el resultado de leer un feed: los elementos aprovechables y
// los avisos de los descartados. Un XML roto NO llega aquí: es un error.
type ResultadoFeed struct {
	Entradas []EntradaFeed
	Avisos   []AvisoFeed
}

// FeedSource lee los dos feeds del cliente (FR-010). Es el puerto que consume
// el servicio de importación (T920/T921) y el punto donde las pruebas inyectan
// un doble: sin red real.
type FeedSource interface {
	FetchYouTube(ctx context.Context) (ResultadoFeed, error)
	FetchPodcast(ctx context.Context) (ResultadoFeed, error)
}

// lectorFeeds es la implementación real de FeedSource (P9-4): lee los dos
// feeds con el cliente de la fábrica única y los parsea con `encoding/xml`.
type lectorFeeds struct {
	cliente    *clienteSaliente
	urlYouTube string
	urlPodcast string
}

// nuevoLectorFeeds construye el lector real de los dos feeds (FR-010) con el
// cliente de la fábrica única (P9-20: https + allowlist, sin IPs privadas,
// redirects validados, cuerpo acotado y User-Agent propio).
func nuevoLectorFeeds() (*lectorFeeds, error) {
	cliente, err := nuevoClienteSaliente(configuracionCliente{
		nombre:          "feeds",
		hostsPermitidos: []string{hostFeedYouTube, hostFeedPodcast},
		limiteCuerpo:    limiteCuerpoFeeds,
		tiempoEspera:    tiempoEsperaFeeds,
	})
	if err != nil {
		return nil, err
	}
	return nuevoLectorFeedsConCliente(cliente), nil
}

// nuevoLectorFeedsConCliente permite inyectar la fábrica en las pruebas.
func nuevoLectorFeedsConCliente(cliente *clienteSaliente) *lectorFeeds {
	return &lectorFeeds{cliente: cliente, urlYouTube: urlFeedYouTube, urlPodcast: urlFeedPodcast}
}

// FetchYouTube lee y parsea el feed Atom del canal de YouTube.
func (l *lectorFeeds) FetchYouTube(ctx context.Context) (ResultadoFeed, error) {
	return l.leer(ctx, l.urlYouTube, "YouTube", parsearAtomYouTube)
}

// FetchPodcast lee y parsea el RSS 2.0 del podcast.
func (l *lectorFeeds) FetchPodcast(ctx context.Context) (ResultadoFeed, error) {
	return l.leer(ctx, l.urlPodcast, "el podcast", parsearRSSPodcast)
}

// leer aplica el camino común de los dos orígenes: pedir por la fábrica única,
// exigir un `200` y un cuerpo no truncado (un feed cortado no se parsea jamás) y
// delegar en el parser del formato.
func (l *lectorFeeds) leer(
	ctx context.Context,
	destino string,
	nombre string,
	parsear func([]byte) (ResultadoFeed, error),
) (ResultadoFeed, error) {
	resp, err := l.cliente.obtener(ctx, destino)
	if err != nil {
		return ResultadoFeed{}, fmt.Errorf("medios: leer el feed de %s: %w", nombre, err)
	}
	if resp.Estado != http.StatusOK {
		return ResultadoFeed{}, fmt.Errorf("%w: el feed de %s respondió %d", errFeedNoDisponible, nombre, resp.Estado)
	}
	if resp.Truncado {
		return ResultadoFeed{}, fmt.Errorf("%w: el feed de %s supera el tope de lectura", errFeedNoDisponible, nombre)
	}
	return parsear(resp.Cuerpo)
}

// ---------------------------------------------------------------------------
// Parser Atom (YouTube)
// ---------------------------------------------------------------------------

// atomFeed es la parte del feed Atom que F9 usa. `encoding/xml` casa por nombre
// local, de modo que el prefijo del espacio de nombres (`yt:`, `media:`) es
// indiferente.
type atomFeed struct {
	XMLName  xml.Name      `xml:"feed"`
	Entradas []atomEntrada `xml:"entry"`
}

type atomEntrada struct {
	VideoID   string       `xml:"videoId"`
	Titulo    string       `xml:"title"`
	Enlaces   []atomEnlace `xml:"link"`
	Publicado string       `xml:"published"`
	Grupo     atomGrupo    `xml:"group"`
}

type atomEnlace struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type atomGrupo struct {
	Caratulas   []atomCaratula `xml:"thumbnail"`
	Descripcion string         `xml:"description"`
}

type atomCaratula struct {
	URL string `xml:"url,attr"`
}

// parsearAtomYouTube convierte el XML del feed de YouTube en entradas
// validadas (función pura: bytes → datos, sin red).
func parsearAtomYouTube(datos []byte) (ResultadoFeed, error) {
	var feed atomFeed
	if err := xml.Unmarshal(datos, &feed); err != nil {
		return ResultadoFeed{}, fmt.Errorf("%w: %v", errXMLFeed, err)
	}

	resultado := ResultadoFeed{Entradas: make([]EntradaFeed, 0, len(feed.Entradas))}
	for _, entrada := range feed.Entradas {
		id := strings.TrimSpace(entrada.VideoID)
		titulo := strings.TrimSpace(entrada.Titulo)
		enlace := enlaceAtom(entrada.Enlaces)

		switch {
		case id == "":
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: "", Motivo: MotivoIdentificadorVacio})
			continue
		case titulo == "":
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: id, Motivo: MotivoTituloVacio})
			continue
		case !patronEnlaceYouTube.MatchString(enlace):
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: id, Motivo: MotivoEnlaceAjeno})
			continue
		}

		publicado, err := time.Parse(time.RFC3339, strings.TrimSpace(entrada.Publicado))
		if err != nil {
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: id, Motivo: MotivoFechaInvalida})
			continue
		}

		resultado.Entradas = append(resultado.Entradas, EntradaFeed{
			IDExterno:   id,
			Titulo:      titulo,
			Enlace:      enlace,
			Publicado:   publicado,
			CaratulaURL: caratulaAtom(entrada.Grupo.Caratulas),
			// `media:description` es texto: el parser de XML ya decodificó sus
			// entidades; solo se colapsan los espacios.
			Descripcion: colapsarEspacios(entrada.Grupo.Descripcion),
		})
	}
	return resultado, nil
}

// enlaceAtom elige el enlace del elemento: el `rel="alternate"` (el que apunta
// al video) y, si no lo trae, el primero con `href`.
func enlaceAtom(enlaces []atomEnlace) string {
	for _, enlace := range enlaces {
		if enlace.Rel == "alternate" && enlace.Href != "" {
			return strings.TrimSpace(enlace.Href)
		}
	}
	for _, enlace := range enlaces {
		if enlace.Href != "" {
			return strings.TrimSpace(enlace.Href)
		}
	}
	return ""
}

// caratulaAtom toma la primera carátula del grupo (`media:thumbnail`); "" si el
// feed no la trae (la ficha entra sin carátula: marcador neutro).
func caratulaAtom(caratulas []atomCaratula) string {
	for _, caratula := range caratulas {
		if url := strings.TrimSpace(caratula.URL); url != "" {
			return url
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Parser RSS 2.0 (podcast)
// ---------------------------------------------------------------------------

// rssFeed es la parte del RSS 2.0 que F9 usa. La carátula del podcast es el
// `itunes:image` del elemento (research R9-5: el arte del programa se repite en
// todos los episodios; sin `itunes:image` la ficha entra sin carátula y el
// cliente pinta el marcador neutro).
type rssFeed struct {
	XMLName xml.Name `xml:"rss"`
	Canal   rssCanal `xml:"channel"`
}

type rssCanal struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Titulo      string    `xml:"title"`
	Descripcion string    `xml:"description"`
	Enlace      string    `xml:"link"`
	GUID        string    `xml:"guid"`
	PubDate     string    `xml:"pubDate"`
	Imagen      rssImagen `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd image"`
}

type rssImagen struct {
	Href string `xml:"href,attr"`
}

// parsearRSSPodcast convierte el XML del feed del podcast en entradas
// validadas (función pura: bytes → datos, sin red).
func parsearRSSPodcast(datos []byte) (ResultadoFeed, error) {
	var feed rssFeed
	if err := xml.Unmarshal(datos, &feed); err != nil {
		return ResultadoFeed{}, fmt.Errorf("%w: %v", errXMLFeed, err)
	}

	resultado := ResultadoFeed{Entradas: make([]EntradaFeed, 0, len(feed.Canal.Items))}

	for _, item := range feed.Canal.Items {
		id := strings.TrimSpace(item.GUID)
		titulo := strings.TrimSpace(item.Titulo)
		enlace := strings.TrimSpace(item.Enlace)

		switch {
		case id == "":
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: "", Motivo: MotivoIdentificadorVacio})
			continue
		case titulo == "":
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: id, Motivo: MotivoTituloVacio})
			continue
		case !patronEnlacePodcast.MatchString(enlace):
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: id, Motivo: MotivoEnlaceAjeno})
			continue
		}

		publicado, err := parsearFechaRSS(item.PubDate)
		if err != nil {
			resultado.Avisos = append(resultado.Avisos, AvisoFeed{IDExterno: id, Motivo: MotivoFechaInvalida})
			continue
		}

		caratula := strings.TrimSpace(item.Imagen.Href)

		resultado.Entradas = append(resultado.Entradas, EntradaFeed{
			IDExterno:   id,
			Titulo:      titulo,
			Enlace:      enlace,
			Publicado:   publicado,
			CaratulaURL: caratula,
			Descripcion: textoPlanoDescripcion(item.Descripcion),
		})
	}
	return resultado, nil
}

// parsearFechaRSS interpreta el `pubDate` de RSS 2.0: RFC1123 con `GMT`
// (el caso real verificado) y, si falla, RFC1123Z con desplazamiento numérico
// (`+0000`).
func parsearFechaRSS(valor string) (time.Time, error) {
	valor = strings.TrimSpace(valor)
	if fecha, err := time.Parse(time.RFC1123, valor); err == nil {
		return fecha, nil
	}
	return time.Parse(time.RFC1123Z, valor)
}

// ---------------------------------------------------------------------------
// Texto plano (FR-023: el contenido del feed es dato, nunca código)
// ---------------------------------------------------------------------------

// textoPlanoDescripcion convierte la descripción del podcast —HTML dentro de
// CDATA— en texto plano: quita las etiquetas, decodifica las entidades
// (`&quot;`…) y colapsa los espacios. El HTML del feed **nunca** se publica ni
// se interpreta (constitución §IV).
func textoPlanoDescripcion(descripcion string) string {
	var b strings.Builder
	b.Grow(len(descripcion))
	dentroDeEtiqueta := false
	for _, r := range descripcion {
		switch {
		case r == '<':
			dentroDeEtiqueta = true
			// Cada etiqueta separa palabras (`</p><p>` → espacio, no "ab").
			b.WriteRune(' ')
		case r == '>':
			dentroDeEtiqueta = false
			b.WriteRune(' ')
		case dentroDeEtiqueta:
			continue
		default:
			b.WriteRune(r)
		}
	}
	// Las entidades se decodifican DESPUÉS de quitar las etiquetas: así un
	// `&lt;script&gt;` literal se publica como texto (`<script>`) y no se
	// confunde con una etiqueta del feed.
	return colapsarEspacios(html.UnescapeString(b.String()))
}

// colapsarEspacios convierte cualquier secuencia de espacios en blanco en un
// único espacio y recorta los bordes.
func colapsarEspacios(texto string) string {
	return strings.Join(strings.Fields(texto), " ")
}
