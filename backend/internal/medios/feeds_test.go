package medios

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// fixture lee un archivo de testdata (datos de prueba commiteados, nunca red).
func fixture(t *testing.T, nombre string) []byte {
	t.Helper()
	datos, err := os.ReadFile("testdata/" + nombre)
	if err != nil {
		t.Fatalf("leer testdata/%s: %v", nombre, err)
	}
	return datos
}

// TestParsearAtomYouTubeFixture fija el parseo del feed Atom de YouTube
// (P9-4/R9-3) contra el fixture de 9 entradas: identificador externo
// (`yt:videoId`), título, enlace `watch?v=`, fecha RFC3339, carátula
// (`media:thumbnail`) y descripción a texto plano, con tildes, ñ, ¿¡ y emojis
// intactos.
func TestParsearAtomYouTubeFixture(t *testing.T) {
	resultado, err := parsearAtomYouTube(fixture(t, "youtube_feed.xml"))
	if err != nil {
		t.Fatalf("parsearAtomYouTube() = %v, se esperaba nil", err)
	}
	if len(resultado.Entradas) != 9 {
		t.Fatalf("entradas = %d, se esperaban 9", len(resultado.Entradas))
	}
	if len(resultado.Avisos) != 0 {
		t.Errorf("avisos = %v, se esperaba ninguno", resultado.Avisos)
	}

	porID := map[string]EntradaFeed{}
	for _, entrada := range resultado.Entradas {
		porID[entrada.IDExterno] = entrada
	}

	video, ok := porID["SimienteAA6"]
	if !ok {
		t.Fatalf("no se encontró el elemento del par real; ids = %v", ids(resultado.Entradas))
	}
	// El par real de research R9-4: el título del video va SIN los dos puntos.
	if video.Titulo != "Potencial sin Resultados La Lección de la Higuera" {
		t.Errorf("título del video = %q", video.Titulo)
	}
	if video.Enlace != "https://www.youtube.com/watch?v=SimienteAA6" {
		t.Errorf("enlace = %q", video.Enlace)
	}
	if got := video.Publicado.UTC().Format(time.RFC3339); got != "2026-05-25T15:00:00Z" {
		t.Errorf("publicado = %s, se esperaba 2026-05-25T15:00:00Z", got)
	}
	if video.CaratulaURL != "https://i.ytimg.com/vi/SimienteAA6/hqdefault.jpg" {
		t.Errorf("carátula = %q", video.CaratulaURL)
	}
	// Entidades del XML decodificadas (&#39; → ') y sin etiquetas.
	if !strings.Contains(video.Descripcion, "'hojas'") {
		t.Errorf("la descripción no decodificó la entidad: %q", video.Descripcion)
	}
	if strings.ContainsAny(video.Descripcion, "<>") {
		t.Errorf("la descripción conserva etiquetas: %q", video.Descripcion)
	}

	// Tildes, ñ, ¿¡ y emojis intactos en el título y el identificador externo
	// del elemento exclusivo de YouTube (spec §Datos de origen).
	exclusivo, ok := porID["SimienteAA9"]
	if !ok || exclusivo.Titulo != "A Mayor Fe, Mayor Presión" {
		t.Fatalf("el elemento exclusivo de YouTube no se parseó: %+v", exclusivo)
	}
	if _, ok := porID["SimienteAA2"]; !ok {
		t.Errorf("falta un elemento con ñ y tildes en el título")
	}
}

// TestParsearRSSPodcastFixture fija el parseo del RSS 2.0 del podcast
// (P9-4/R9-3) contra el fixture real de 9 elementos: `guid` como identificador
// externo, `pubDate` con `GMT`, el enlace del episodio, `itunes:image` y la
// descripción con HTML dentro de CDATA → texto plano con entidades
// decodificadas.
func TestParsearRSSPodcastFixture(t *testing.T) {
	resultado, err := parsearRSSPodcast(fixture(t, "podcast_feed.xml"))
	if err != nil {
		t.Fatalf("parsearRSSPodcast() = %v, se esperaba nil", err)
	}
	if len(resultado.Entradas) != 9 {
		t.Fatalf("entradas = %d, se esperaban 9", len(resultado.Entradas))
	}
	if len(resultado.Avisos) != 0 {
		t.Errorf("avisos = %v, se esperaba ninguno", resultado.Avisos)
	}

	porID := map[string]EntradaFeed{}
	for _, entrada := range resultado.Entradas {
		porID[entrada.IDExterno] = entrada
	}
	audio, ok := porID["bc42e0cd-d585-40ab-ac37-b67695a1d545"]
	if !ok {
		t.Fatalf("no se encontró el elemento del par real; ids = %v", ids(resultado.Entradas))
	}
	if audio.Titulo != "Potencial sin Resultados: La Lección de la Higuera" {
		t.Errorf("título del podcast = %q", audio.Titulo)
	}
	if !strings.HasPrefix(audio.Enlace, "https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/") {
		t.Errorf("enlace = %q", audio.Enlace)
	}
	if got := audio.Publicado.UTC().Format(time.RFC3339); got != "2026-05-25T17:58:50Z" {
		t.Errorf("publicado = %s, se esperaba 2026-05-25T17:58:50Z", got)
	}
	if audio.CaratulaURL != "https://d3t3ozftmdmh3i.cloudfront.net/staging/podcast_uploaded_nologo/45766890/45766890-1776097995866-18d4d70c5410a.jpg" {
		t.Errorf("carátula = %q", audio.CaratulaURL)
	}
	if strings.ContainsAny(audio.Descripcion, "<>") {
		t.Errorf("la descripción conserva etiquetas HTML: %q", audio.Descripcion)
	}
	if !strings.Contains(audio.Descripcion, `"hojas"`) {
		t.Errorf("la descripción no decodificó &quot;: %q", audio.Descripcion)
	}
	if !strings.Contains(audio.Descripcion, "¡Escúchalo") {
		t.Errorf("la descripción perdió los signos del español: %q", audio.Descripcion)
	}

	// El par real de títulos normaliza igual (la llave de la unión, FR-013).
	video := EntradaFeed{Titulo: "Potencial sin Resultados La Lección de la Higuera"}
	if normalizarTitulo(video.Titulo) != normalizarTitulo(audio.Titulo) {
		t.Errorf("el par real no normaliza igual: %q vs %q",
			normalizarTitulo(video.Titulo), normalizarTitulo(audio.Titulo))
	}
}

// TestParsearAtomYouTubeDescartaConAviso fija que un enlace fuera del host del
// proveedor, una fecha imposible y un título vacío **descartan el elemento con
// aviso** (nunca se inventa una fecha ni se importa un enlace ajeno) y que los
// elementos sanos del mismo feed sí entran.
func TestParsearAtomYouTubeDescartaConAviso(t *testing.T) {
	resultado, err := parsearAtomYouTube(fixture(t, "youtube_feed_patologico.xml"))
	if err != nil {
		t.Fatalf("parsearAtomYouTube() = %v, se esperaba nil", err)
	}
	if len(resultado.Entradas) != 1 || resultado.Entradas[0].IDExterno != "SimienteCC1" {
		t.Fatalf("entradas = %v, se esperaba solo SimienteCC1", ids(resultado.Entradas))
	}
	if resultado.Entradas[0].Titulo != "Sermón correcto con tildes, ñ y ¡signos!" {
		t.Errorf("título = %q", resultado.Entradas[0].Titulo)
	}

	avisos := map[string]MotivoAviso{}
	for _, aviso := range resultado.Avisos {
		avisos[aviso.IDExterno] = aviso.Motivo
	}
	if len(avisos) != 3 {
		t.Fatalf("avisos = %v, se esperaban 3", resultado.Avisos)
	}
	if avisos["SimienteCC2"] != MotivoEnlaceAjeno {
		t.Errorf("aviso de SimienteCC2 = %q, se esperaba %q", avisos["SimienteCC2"], MotivoEnlaceAjeno)
	}
	if avisos["SimienteCC3"] != MotivoFechaInvalida {
		t.Errorf("aviso de SimienteCC3 = %q, se esperaba %q", avisos["SimienteCC3"], MotivoFechaInvalida)
	}
	if avisos["SimienteCC4"] != MotivoTituloVacio {
		t.Errorf("aviso de SimienteCC4 = %q, se esperaba %q", avisos["SimienteCC4"], MotivoTituloVacio)
	}
}

// TestParsearRSSPodcastDescartaConAviso fija los defaults documentados de un
// elemento sin carátula ni descripción ("" en ambos: la ficha omite la línea,
// precisión 3 de la puerta 2) y el descarte con aviso de un enlace fuera de
// `podcasters.spotify.com`; las fechas `+0000` (RFC1123Z) se parsean igual que
// las `GMT`.
func TestParsearRSSPodcastDescartaConAviso(t *testing.T) {
	resultado, err := parsearRSSPodcast(fixture(t, "podcast_feed_patologico.xml"))
	if err != nil {
		t.Fatalf("parsearRSSPodcast() = %v, se esperaba nil", err)
	}
	if len(resultado.Entradas) != 2 {
		t.Fatalf("entradas = %v, se esperaban 2", ids(resultado.Entradas))
	}

	porID := map[string]EntradaFeed{}
	for _, entrada := range resultado.Entradas {
		porID[entrada.IDExterno] = entrada
	}

	completo := porID["guid-patologico-1"]
	if got := completo.Publicado.UTC().Format(time.RFC3339); got != "2026-05-26T03:00:00Z" {
		t.Errorf("fecha con +0000 = %s, se esperaba 2026-05-26T03:00:00Z", got)
	}
	if !strings.Contains(completo.Titulo, "🙌") || !strings.Contains(completo.Titulo, "¡Qué día!") {
		t.Errorf("título con emoji y signos alterado: %q", completo.Titulo)
	}
	if strings.ContainsAny(completo.Descripcion, "<>") {
		t.Errorf("la descripción conserva etiquetas: %q", completo.Descripcion)
	}
	if !strings.Contains(completo.Descripcion, `"completo"`) {
		t.Errorf("la descripción no decodificó &quot;: %q", completo.Descripcion)
	}
	if !strings.Contains(completo.Descripcion, "Primero: mira la higuera.") {
		t.Errorf("la descripción perdió el texto de la lista: %q", completo.Descripcion)
	}

	sinCampos := porID["guid-patologico-2"]
	if sinCampos.CaratulaURL != "" {
		t.Errorf("carátula = %q, se esperaba vacía (sin itunes:image)", sinCampos.CaratulaURL)
	}
	if sinCampos.Descripcion != "" {
		t.Errorf("descripción = %q, se esperaba vacía", sinCampos.Descripcion)
	}

	if len(resultado.Avisos) != 1 || resultado.Avisos[0].Motivo != MotivoEnlaceAjeno {
		t.Fatalf("avisos = %v, se esperaba un enlace ajeno", resultado.Avisos)
	}
	if resultado.Avisos[0].IDExterno != "guid-patologico-3" {
		t.Errorf("aviso sobre %q, se esperaba guid-patologico-3", resultado.Avisos[0].IDExterno)
	}
}

// TestParsearFeedRotoDevuelveError fija que un XML roto devuelve un error claro
// y **nunca** elementos a medias (RG9-11).
func TestParsearFeedRotoDevuelveError(t *testing.T) {
	roto := fixture(t, "feed_roto.xml")

	resultado, err := parsearAtomYouTube(roto)
	if err == nil {
		t.Fatalf("parsearAtomYouTube(xml roto) = nil, se esperaba un error")
	}
	if !errors.Is(err, errXMLFeed) {
		t.Errorf("error = %v, se esperaba %v", err, errXMLFeed)
	}
	if len(resultado.Entradas) != 0 {
		t.Errorf("entradas = %d, se esperaban 0 (nunca items a medias)", len(resultado.Entradas))
	}

	if _, err := parsearRSSPodcast(roto); !errors.Is(err, errXMLFeed) {
		t.Errorf("parsearRSSPodcast(xml roto) = %v, se esperaba %v", err, errXMLFeed)
	}
}

// TestParsearFeedNoEsUnFeed fija que un documento XML bien formado pero ajeno
// (p. ej. la página de error que devuelve el origen) es un error claro.
func TestParsearFeedNoEsUnFeed(t *testing.T) {
	html := []byte(`<html><head><title>Error 404</title></head><body>No encontrado</body></html>`)
	if _, err := parsearAtomYouTube(html); !errors.Is(err, errXMLFeed) {
		t.Errorf("parsearAtomYouTube(página ajena) = %v, se esperaba %v", err, errXMLFeed)
	}
	if _, err := parsearRSSPodcast(html); !errors.Is(err, errXMLFeed) {
		t.Errorf("parsearRSSPodcast(página ajena) = %v, se esperaba %v", err, errXMLFeed)
	}
}

// TestLectorFeedsBarreras ejecuta la tabla compartida de barreras de la red
// saliente (T907A/P9-20) sobre el lector de feeds: la MISMA tabla que los otros
// dos consumidores.
func TestLectorFeedsBarreras(t *testing.T) {
	transporte := &transporteFalso{}
	cliente := clienteDeConsumidor(t, transporte, "feeds", []string{hostFeedYouTube, hostFeedPodcast}, limiteCuerpoFeeds)
	_ = nuevoLectorFeedsConCliente(cliente)

	ejecutarBarrerasCompartidas(t, casoBarreras{
		nombre:           "feeds",
		cliente:          cliente,
		transporte:       transporte,
		destinoPermitido: urlFeedYouTube,
		destinoAjeno:     "https://ajeno.test/feed.xml",
	})
}

// TestLectorFeedsLeeLosDosFeeds fija que el lector pide las dos URLs de FR-010
// (P9-18) por el cliente de la fábrica única y parsea los dos formatos.
func TestLectorFeedsLeeLosDosFeeds(t *testing.T) {
	transporte := &transporteFalso{}
	transporte.responder = func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case hostFeedYouTube:
			return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader(string(fixture(t, "youtube_feed.xml")))), nil), nil
		case hostFeedPodcast:
			return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader(string(fixture(t, "podcast_feed.xml")))), nil), nil
		default:
			return respuestaFalsa(http.StatusNotFound, io.NopCloser(strings.NewReader("")), nil), nil
		}
	}
	cliente := clienteDeConsumidor(t, transporte, "feeds", []string{hostFeedYouTube, hostFeedPodcast}, limiteCuerpoFeeds)
	lector := nuevoLectorFeedsConCliente(cliente)

	youtube, err := lector.FetchYouTube(context.Background())
	if err != nil {
		t.Fatalf("FetchYouTube() = %v", err)
	}
	podcast, err := lector.FetchPodcast(context.Background())
	if err != nil {
		t.Fatalf("FetchPodcast() = %v", err)
	}
	if len(youtube.Entradas) != 9 || len(podcast.Entradas) != 9 {
		t.Fatalf("entradas = %d y %d, se esperaban 9 y 9", len(youtube.Entradas), len(podcast.Entradas))
	}

	destinos := map[string]bool{}
	for _, req := range transporte.peticiones {
		destinos[req.URL.String()] = true
	}
	if !destinos[urlFeedYouTube] || !destinos[urlFeedPodcast] {
		t.Errorf("no se pidieron las dos URLs de FR-010: %v", destinos)
	}
}

// TestLectorFeedsFeedNoDisponible fija que un feed que no responde (404, 5xx o
// cuerpo por encima del tope) es un ERROR del lector —el resumen lo marcará
// `failed` y no se ocultará ni marcará nada por esa causa (FR-016)— y nunca un
// catálogo a medias.
func TestLectorFeedsFeedNoDisponible(t *testing.T) {
	casos := []struct {
		nombre   string
		estado   int
		truncado bool
	}{
		{"404 del origen", http.StatusNotFound, false},
		{"500 del origen", http.StatusInternalServerError, false},
		{"cuerpo por encima del tope", http.StatusOK, true},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			transporte := &transporteFalso{}
			transporte.responder = func(*http.Request) (*http.Response, error) {
				if caso.truncado {
					return respuestaFalsa(caso.estado, &cuerpoInfinito{tope: limiteCuerpoFeeds}, nil), nil
				}
				return respuestaFalsa(caso.estado, io.NopCloser(strings.NewReader("")), nil), nil
			}
			cliente := clienteDeConsumidor(t, transporte, "feeds", []string{hostFeedYouTube, hostFeedPodcast}, limiteCuerpoFeeds)
			lector := nuevoLectorFeedsConCliente(cliente)

			resultado, err := lector.FetchYouTube(context.Background())
			if !errors.Is(err, errFeedNoDisponible) {
				t.Fatalf("FetchYouTube() = %v, se esperaba %v", err, errFeedNoDisponible)
			}
			if len(resultado.Entradas) != 0 || len(resultado.Avisos) != 0 {
				t.Errorf("resultado = %+v, se esperaba vacío", resultado)
			}
		})
	}
}

// TestLectorFeedsImplementaFeedSource deja explícita la interfaz del dominio
// (M1: se declara aquí y T915 no la redeclara).
func TestLectorFeedsImplementaFeedSource(t *testing.T) {
	var fuente FeedSource = (*lectorFeeds)(nil)
	if fuente == nil {
		t.Fatalf("el lector no implementa FeedSource")
	}
}

// clienteDeConsumidor arma el cliente de un consumidor con transporte y
// resolución falsos (sin red).
func clienteDeConsumidor(t *testing.T, transporte http.RoundTripper, nombre string, hosts []string, limite int64) *clienteSaliente {
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

// ids resume los identificadores externos de un resultado (mensajes de error).
func ids(entradas []EntradaFeed) []string {
	salida := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		salida = append(salida, entrada.IDExterno)
	}
	return salida
}
