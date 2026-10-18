package medios

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// probadorDePrueba arma el probador con un transporte falso (sin red).
func probadorDePrueba(t *testing.T, transporte *transporteFalso) *probadorDisponibilidad {
	t.Helper()
	cliente := clienteDeConsumidor(t, transporte, "probador", hostsProbador, limiteCuerpoProbador)
	return nuevoProbadorDisponibilidadConCliente(cliente)
}

// TestProbeYouTubeTabla fija la tabla de casos del probador de YouTube
// (R9-2/P9-6): `200` → disponible, `404` → **no disponible confirmado** y todo
// lo demás —`400`, `401`, `403`, `429`, `5xx`, tiempo agotado o error de red— →
// inconcluso. La regla a prueba de fallos de FR-017: solo el confirmado oculta.
func TestProbeYouTubeTabla(t *testing.T) {
	casos := []struct {
		nombre string
		estado int
		red    bool
		cuelga bool
		want   ResultadoDisponibilidad
	}{
		{nombre: "200 disponible", estado: http.StatusOK, want: DisponibilidadDisponible},
		{nombre: "404 confirmado", estado: http.StatusNotFound, want: DisponibilidadNoDisponible},
		{nombre: "400 inconcluso", estado: http.StatusBadRequest, want: DisponibilidadInconclusa},
		{nombre: "401 inconcluso", estado: http.StatusUnauthorized, want: DisponibilidadInconclusa},
		{nombre: "403 inconcluso", estado: http.StatusForbidden, want: DisponibilidadInconclusa},
		{nombre: "429 inconcluso", estado: http.StatusTooManyRequests, want: DisponibilidadInconclusa},
		{nombre: "500 inconcluso", estado: http.StatusInternalServerError, want: DisponibilidadInconclusa},
		{nombre: "503 inconcluso", estado: http.StatusServiceUnavailable, want: DisponibilidadInconclusa},
		{nombre: "error de red inconcluso", red: true, want: DisponibilidadInconclusa},
		{nombre: "tiempo agotado inconcluso", cuelga: true, want: DisponibilidadInconclusa},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			transporte := &transporteFalso{}
			transporte.responder = func(req *http.Request) (*http.Response, error) {
				switch {
				case caso.red:
					return nil, errors.New("red caída")
				case caso.cuelga:
					<-req.Context().Done()
					return nil, req.Context().Err()
				default:
					return respuestaFalsa(caso.estado, io.NopCloser(strings.NewReader("{}")), nil), nil
				}
			}
			probador := probadorDePrueba(t, transporte)

			got, err := probador.ProbeYouTube(context.Background(), "dQw4w9WgXcQ")
			if got != caso.want {
				t.Fatalf("ProbeYouTube() = %q, se esperaba %q (err = %v)", got, caso.want, err)
			}
			if caso.want != DisponibilidadInconclusa && err != nil {
				t.Errorf("err = %v, se esperaba nil en un resultado concluyente", err)
			}
		})
	}
}

// TestProbePodcastTabla fija la tabla de casos del probador del podcast
// (R9-2/P9-6): se comprueba el **enlace del episodio** que trae el feed;
// `200` → disponible, `404`/`410` → **no disponible confirmado** y el resto
// (`504` verificado con los enlaces reales, `5xx`, tiempo agotado, red) →
// inconcluso.
func TestProbePodcastTabla(t *testing.T) {
	casos := []struct {
		nombre string
		estado int
		red    bool
		cuelga bool
		want   ResultadoDisponibilidad
	}{
		{nombre: "200 disponible", estado: http.StatusOK, want: DisponibilidadDisponible},
		{nombre: "404 confirmado", estado: http.StatusNotFound, want: DisponibilidadNoDisponible},
		{nombre: "410 confirmado", estado: http.StatusGone, want: DisponibilidadNoDisponible},
		{nombre: "500 inconcluso", estado: http.StatusInternalServerError, want: DisponibilidadInconclusa},
		{nombre: "504 inconcluso", estado: http.StatusGatewayTimeout, want: DisponibilidadInconclusa},
		{nombre: "error de red inconcluso", red: true, want: DisponibilidadInconclusa},
		{nombre: "tiempo agotado inconcluso", cuelga: true, want: DisponibilidadInconclusa},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			transporte := &transporteFalso{}
			transporte.responder = func(req *http.Request) (*http.Response, error) {
				switch {
				case caso.red:
					return nil, errors.New("red caída")
				case caso.cuelga:
					<-req.Context().Done()
					return nil, req.Context().Err()
				default:
					return respuestaFalsa(caso.estado, io.NopCloser(strings.NewReader("")), nil), nil
				}
			}
			probador := probadorDePrueba(t, transporte)

			got, err := probador.ProbePodcast(context.Background(), enlaceEpisodio)
			if got != caso.want {
				t.Fatalf("ProbePodcast() = %q, se esperaba %q (err = %v)", got, caso.want, err)
			}
			if caso.want != DisponibilidadInconclusa && err != nil {
				t.Errorf("err = %v, se esperaba nil en un resultado concluyente", err)
			}
		})
	}
}

// TestProbeSoloConfirma404Y410 fija la parte crítica de SC-014: **ningún** otro
// código HTTP produce `unavailable`. Ni el 403 (bloqueo del servicio), ni el
// 429 (cuota), ni el 5xx, ni un 200 raro pueden ocultar una ficha.
func TestProbeSoloConfirma404Y410(t *testing.T) {
	codigos := []int{
		http.StatusOK, http.StatusCreated, http.StatusMovedPermanently, http.StatusFound,
		http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusMethodNotAllowed, http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout,
	}
	for _, codigo := range codigos {
		t.Run(http.StatusText(codigo), func(t *testing.T) {
			transporte := &transporteFalso{}
			transporte.responder = func(*http.Request) (*http.Response, error) {
				return respuestaFalsa(codigo, io.NopCloser(strings.NewReader("")), nil), nil
			}
			probador := probadorDePrueba(t, transporte)

			if got, _ := probador.ProbeYouTube(context.Background(), "dQw4w9WgXcQ"); got == DisponibilidadNoDisponible {
				t.Fatalf("YouTube: el código %d ocultó una ficha", codigo)
			}
			if got, _ := probador.ProbePodcast(context.Background(), enlaceEpisodio); got == DisponibilidadNoDisponible {
				t.Fatalf("podcast: el código %d ocultó una ficha", codigo)
			}
		})
	}

	// Los únicos códigos que confirman la indisponibilidad.
	for _, codigo := range []int{http.StatusNotFound, http.StatusGone} {
		transporte := &transporteFalso{}
		transporte.responder = func(*http.Request) (*http.Response, error) {
			return respuestaFalsa(codigo, io.NopCloser(strings.NewReader("")), nil), nil
		}
		probador := probadorDePrueba(t, transporte)
		if got, _ := probador.ProbePodcast(context.Background(), enlaceEpisodio); got != DisponibilidadNoDisponible {
			t.Errorf("podcast: el código %d = %q, se esperaba %q", codigo, got, DisponibilidadNoDisponible)
		}
	}
}

// TestProbeYouTubePideOEmbed fija que el probador de YouTube usa el oEmbed
// público (R9-2: gratis, sin claves y sin cuota) con la forma verificada
// `oembed?url=<video>&format=json`, y no la página del video.
func TestProbeYouTubePideOEmbed(t *testing.T) {
	transporte := &transporteFalso{}
	transporte.responder = func(*http.Request) (*http.Response, error) {
		return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader(`{"title":"x"}`)), nil), nil
	}
	probador := probadorDePrueba(t, transporte)

	if _, err := probador.ProbeYouTube(context.Background(), "dQw4w9WgXcQ"); err != nil {
		t.Fatalf("ProbeYouTube() = %v", err)
	}
	pedida := transporte.ultima()
	if pedida == nil {
		t.Fatalf("no se hizo ninguna petición")
	}
	if pedida.URL.Host != hostOEmbedYouTube || pedida.URL.Path != "/oembed" {
		t.Errorf("destino = %s, se esperaba el oEmbed de YouTube", pedida.URL)
	}
	if got := pedida.URL.Query().Get("url"); got != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Errorf("url del video = %q", got)
	}
	if got := pedida.URL.Query().Get("format"); got != "json" {
		t.Errorf("format = %q, se esperaba json", got)
	}
}

// TestProbePodcastPideElEnlaceDelEpisodio fija que el probador del podcast
// comprueba el enlace que trae el feed (y no un oEmbed de Spotify: 504
// verificado, R9-2).
func TestProbePodcastPideElEnlaceDelEpisodio(t *testing.T) {
	transporte := &transporteFalso{}
	transporte.responder = func(*http.Request) (*http.Response, error) {
		return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader("")), nil), nil
	}
	probador := probadorDePrueba(t, transporte)

	if _, err := probador.ProbePodcast(context.Background(), enlaceEpisodio); err != nil {
		t.Fatalf("ProbePodcast() = %v", err)
	}
	if got := transporte.ultima().URL.String(); got != enlaceEpisodio {
		t.Errorf("destino = %q, se esperaba %q", got, enlaceEpisodio)
	}
}

// TestProbePodcastRechazaEnlaceAjeno fija que el probador no sale del host del
// proveedor: un enlace que no sea de `podcasters.spotify.com` no se pide jamás
// (barrera 1 de la fábrica única) y queda inconcluso.
func TestProbePodcastRechazaEnlaceAjeno(t *testing.T) {
	transporte := &transporteFalso{}
	probador := probadorDePrueba(t, transporte)

	got, err := probador.ProbePodcast(context.Background(), "https://evil.example/episodio")
	if got != DisponibilidadInconclusa {
		t.Errorf("ProbePodcast(enlace ajeno) = %q, se esperaba %q", got, DisponibilidadInconclusa)
	}
	if err == nil {
		t.Errorf("se esperaba el error informativo del rechazo de host")
	}
	if len(transporte.peticiones) != 0 {
		t.Errorf("se pidió un host fuera de la allowlist: %v", transporte.peticiones)
	}
}

// TestProbadorBarreras ejecuta la tabla compartida de barreras (T907A/P9-20)
// sobre el probador de disponibilidad: la MISMA tabla que los otros dos
// consumidores.
func TestProbadorBarreras(t *testing.T) {
	transporte := &transporteFalso{}
	cliente := clienteDeConsumidor(t, transporte, "probador", hostsProbador, limiteCuerpoProbador)
	_ = nuevoProbadorDisponibilidadConCliente(cliente)

	ejecutarBarrerasCompartidas(t, casoBarreras{
		nombre:           "probador",
		cliente:          cliente,
		transporte:       transporte,
		destinoPermitido: "https://www.youtube.com/oembed?url=x&format=json",
		destinoAjeno:     "https://ajeno.test/oembed",
	})
}

// TestNuevoProbadorDisponibilidadConfiguracion fija la configuración REAL del
// probador (R9-2/P9-18): el endpoint público de oEmbed, el host del episodio del
// podcast, el tope de cuerpo y el tiempo de espera de 5 s.
func TestNuevoProbadorDisponibilidadConfiguracion(t *testing.T) {
	probador, err := nuevoProbadorDisponibilidad()
	if err != nil {
		t.Fatalf("nuevoProbadorDisponibilidad() = %v", err)
	}
	if urlOEmbedYouTube != "https://www.youtube.com/oembed" {
		t.Errorf("endpoint de oEmbed = %q, se esperaba el público de YouTube", urlOEmbedYouTube)
	}
	if len(probador.cliente.hosts) != 2 {
		t.Fatalf("allowlist = %v, se esperaban el oEmbed y el host del episodio", probador.cliente.hosts)
	}
	for _, host := range []string{hostOEmbedYouTube, hostEpisodioPodcast} {
		if _, ok := probador.cliente.hosts[host]; !ok {
			t.Errorf("falta %q en la allowlist del probador", host)
		}
	}
	if probador.cliente.limiteCuerpo != limiteCuerpoProbador {
		t.Errorf("tope de cuerpo = %d, se esperaba %d", probador.cliente.limiteCuerpo, limiteCuerpoProbador)
	}
	if probador.cliente.cliente.Timeout != tiempoEsperaProbador {
		t.Errorf("tiempo de espera = %v, se esperaba %v", probador.cliente.cliente.Timeout, tiempoEsperaProbador)
	}
}

// enlaceEpisodio es el enlace real de un episodio (research R9-5).
const enlaceEpisodio = "https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/Potencial-sin-Resultados-La-Leccin-de-la-Higuera-e3jsj3n"

// fakeAvailabilityProber es el doble de `AvailabilityProber` que usan las
// pruebas del dominio (T921: la disponibilidad se prueba sin red).
type fakeAvailabilityProber struct {
	youtube    map[string]ResultadoDisponibilidad
	podcast    map[string]ResultadoDisponibilidad
	errYoutube error
	errPodcast error
}

func (f *fakeAvailabilityProber) ProbeYouTube(_ context.Context, videoID string) (ResultadoDisponibilidad, error) {
	if resultado, ok := f.youtube[videoID]; ok {
		return resultado, f.errYoutube
	}
	return DisponibilidadInconclusa, f.errYoutube
}

func (f *fakeAvailabilityProber) ProbePodcast(_ context.Context, enlace string) (ResultadoDisponibilidad, error) {
	if resultado, ok := f.podcast[enlace]; ok {
		return resultado, f.errPodcast
	}
	return DisponibilidadInconclusa, f.errPodcast
}

var _ AvailabilityProber = (*fakeAvailabilityProber)(nil)

// TestFakeAvailabilityProber fija el doble que usarán las pruebas del servicio.
func TestFakeAvailabilityProber(t *testing.T) {
	fake := &fakeAvailabilityProber{
		youtube: map[string]ResultadoDisponibilidad{"vivo": DisponibilidadDisponible, "muerto": DisponibilidadNoDisponible},
		podcast: map[string]ResultadoDisponibilidad{enlaceEpisodio: DisponibilidadNoDisponible},
	}
	if got, _ := fake.ProbeYouTube(context.Background(), "vivo"); got != DisponibilidadDisponible {
		t.Errorf("ProbeYouTube(vivo) = %q", got)
	}
	if got, _ := fake.ProbeYouTube(context.Background(), "muerto"); got != DisponibilidadNoDisponible {
		t.Errorf("ProbeYouTube(muerto) = %q", got)
	}
	if got, _ := fake.ProbePodcast(context.Background(), enlaceEpisodio); got != DisponibilidadNoDisponible {
		t.Errorf("ProbePodcast() = %q", got)
	}
	if got, _ := fake.ProbeYouTube(context.Background(), "desconocido"); got != DisponibilidadInconclusa {
		t.Errorf("ProbeYouTube(desconocido) = %q, se esperaba inconcluso", got)
	}
}

// TestProbeRespetaElTiempoDeEspera fija que una comprobación que no responde
// termina por tiempo agotado (5 s en producción) y nunca cuelga la búsqueda.
func TestProbeRespetaElTiempoDeEspera(t *testing.T) {
	transporte := &transporteFalso{}
	transporte.responder = func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	}
	probador := probadorDePrueba(t, transporte)

	inicio := time.Now()
	if got, _ := probador.ProbeYouTube(context.Background(), "dQw4w9WgXcQ"); got != DisponibilidadInconclusa {
		t.Fatalf("ProbeYouTube() = %q, se esperaba inconcluso", got)
	}
	if transcurrido := time.Since(inicio); transcurrido > 2*time.Second {
		t.Errorf("tardó %v: no se aplicó el tiempo de espera del consumidor", transcurrido)
	}
}
