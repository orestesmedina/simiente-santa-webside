package medios

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Dobles COMPARTIDOS por las pruebas de la red saliente (T907A–T910). Viven
// aquí porque la tabla de barreras de P9-20 es la MISMA en los tres
// consumidores (corrección A1 de la fase 5). Ninguno toca la red real.
// ============================================================================

// transporteFalso es un http.RoundTripper programable: no toca la red y
// registra cada petición que sale por la fábrica única.
type transporteFalso struct {
	responder  func(*http.Request) (*http.Response, error)
	peticiones []*http.Request
}

func (f *transporteFalso) RoundTrip(req *http.Request) (*http.Response, error) {
	f.peticiones = append(f.peticiones, req)
	if f.responder == nil {
		return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader("ok")), nil), nil
	}
	return f.responder(req)
}

// ultima devuelve la última petición vista (nil si no hubo ninguna).
func (f *transporteFalso) ultima() *http.Request {
	if len(f.peticiones) == 0 {
		return nil
	}
	return f.peticiones[len(f.peticiones)-1]
}

// respuestaFalsa arma una respuesta HTTP con estado, cuerpo y cabeceras.
func respuestaFalsa(estado int, cuerpo io.ReadCloser, cabeceras map[string]string) *http.Response {
	header := make(http.Header, len(cabeceras))
	for clave, valor := range cabeceras {
		header.Set(clave, valor)
	}
	return &http.Response{
		StatusCode: estado,
		Header:     header,
		Body:       cuerpo,
	}
}

// cuerpoInfinito entrega exactamente `tope` bytes y falla si alguien intenta
// leer más allá: si un consumidor leyera el cuerpo sin io.LimitReader, la
// prueba falla en vez de colgarse. Es la prueba directa de la barrera 4.
type cuerpoInfinito struct {
	tope   int64
	leidos int64
}

func (c *cuerpoInfinito) Read(p []byte) (int, error) {
	if c.leidos >= c.tope {
		return 0, fmt.Errorf("cuerpoInfinito: se intentó leer más allá del tope (%d bytes)", c.tope)
	}
	n := len(p)
	if int64(n) > c.tope-c.leidos {
		n = int(c.tope - c.leidos)
	}
	for i := 0; i < n; i++ {
		p[i] = 'x'
	}
	c.leidos += int64(n)
	return n, nil
}

func (c *cuerpoInfinito) Close() error { return nil }

// resolverFalso devuelve las direcciones dadas para cualquier host, sin DNS.
func resolverFalso(ips ...string) func(context.Context, string) ([]netip.Addr, error) {
	direcciones := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		direcciones = append(direcciones, netip.MustParseAddr(ip))
	}
	return func(context.Context, string) ([]netip.Addr, error) {
		return direcciones, nil
	}
}

// dialFalso registra la dirección a la que se intentaría conectar (nunca
// conecta).
type dialFalso struct {
	direcciones []string
}

func (d *dialFalso) dial(_ context.Context, _, address string) (net.Conn, error) {
	d.direcciones = append(d.direcciones, address)
	return nil, errors.New("dialFalso: no conecta")
}

// casoBarreras describe un consumidor de la red saliente para ejecutar la tabla
// compartida de barreras (P9-20/A1).
type casoBarreras struct {
	// nombre del consumidor (subtítulos de la prueba).
	nombre string
	// cliente es la fábrica del consumidor, con transporte y resolver falsos.
	cliente *clienteSaliente
	// transporte es el doble programable que recibe sus peticiones.
	transporte *transporteFalso
	// destinoPermitido es una URL `https` de un host de la allowlist.
	destinoPermitido string
	// destinoAjeno es una URL `https` de un host FUERA de la allowlist.
	destinoAjeno string
}

// ejecutarBarrerasCompartidas corre la tabla de barreras de la red saliente
// sobre un consumidor: la MISMA tabla que T907A, T908, T909 y T910 ejecutan
// sobre cada uno (corrección A1 de la fase 5). Sin red real.
func ejecutarBarrerasCompartidas(t *testing.T, caso casoBarreras) {
	t.Helper()

	t.Run(caso.nombre+"/https de la allowlist con User-Agent propio", func(t *testing.T) {
		caso.transporte.responder = func(*http.Request) (*http.Response, error) {
			return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader("ok")), nil), nil
		}
		resp, err := caso.cliente.obtener(context.Background(), caso.destinoPermitido)
		if err != nil {
			t.Fatalf("obtener(%q) = %v, se esperaba nil", caso.destinoPermitido, err)
		}
		if resp.Estado != http.StatusOK {
			t.Fatalf("estado = %d, se esperaba 200", resp.Estado)
		}
		if got := caso.transporte.ultima().Header.Get("User-Agent"); got != usuarioAgente {
			t.Errorf("User-Agent = %q, se esperaba %q", got, usuarioAgente)
		}
	})

	t.Run(caso.nombre+"/http rechazado (barrera 1)", func(t *testing.T) {
		antes := len(caso.transporte.peticiones)
		destino := "http://" + strings.TrimPrefix(caso.destinoPermitido, "https://")
		if _, err := caso.cliente.obtener(context.Background(), destino); !errors.Is(err, errEsquemaNoPermitido) {
			t.Fatalf("obtener(%q) = %v, se esperaba %v", destino, err, errEsquemaNoPermitido)
		}
		if len(caso.transporte.peticiones) != antes {
			t.Errorf("se envió la petición http:// pese a rechazarla")
		}
	})

	t.Run(caso.nombre+"/host ajeno rechazado (barrera 1)", func(t *testing.T) {
		antes := len(caso.transporte.peticiones)
		if _, err := caso.cliente.obtener(context.Background(), caso.destinoAjeno); !errors.Is(err, errHostNoPermitido) {
			t.Fatalf("obtener(%q) = %v, se esperaba %v", caso.destinoAjeno, err, errHostNoPermitido)
		}
		if len(caso.transporte.peticiones) != antes {
			t.Errorf("se envió la petición a un host ajeno pese a rechazarla")
		}
	})

	t.Run(caso.nombre+"/puerto distinto de 443 rechazado (barrera 1)", func(t *testing.T) {
		antes := len(caso.transporte.peticiones)
		destino := conPuerto(t, caso.destinoPermitido, "8443")
		if _, err := caso.cliente.obtener(context.Background(), destino); !errors.Is(err, errPuertoNoPermitido) {
			t.Fatalf("obtener(%q) = %v, se esperaba %v", destino, err, errPuertoNoPermitido)
		}
		if len(caso.transporte.peticiones) != antes {
			t.Errorf("se envió la petición a un puerto no permitido pese a rechazarla")
		}

		// El dialer también lo exige (defensa en profundidad: el transporte
		// siempre trae puerto).
		u, err := url.Parse(caso.destinoPermitido)
		if err != nil {
			t.Fatalf("url.Parse(%q): %v", caso.destinoPermitido, err)
		}
		dialContext := caso.cliente.dialSeguro()
		if _, err := dialContext(context.Background(), "tcp", net.JoinHostPort(u.Hostname(), "8443")); !errors.Is(err, errPuertoNoPermitido) {
			t.Errorf("dialSeguro(puerto ajeno) = %v, se esperaba %v", err, errPuertoNoPermitido)
		}

		// El 443 explícito sí sale (y el dialer conecta al 443, no al que traiga
		// la petición).
		caso.transporte.responder = func(*http.Request) (*http.Response, error) {
			return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader("ok")), nil), nil
		}
		if _, err := caso.cliente.obtener(context.Background(), conPuerto(t, caso.destinoPermitido, "443")); err != nil {
			t.Fatalf("obtener(https://…:443) = %v, se esperaba nil", err)
		}
	})

	t.Run(caso.nombre+"/redirect a host ajeno rechazado (barrera 3)", func(t *testing.T) {
		caso.transporte.responder = func(*http.Request) (*http.Response, error) {
			return respuestaFalsa(http.StatusFound, nil, map[string]string{"Location": caso.destinoAjeno}), nil
		}
		if _, err := caso.cliente.obtener(context.Background(), caso.destinoPermitido); !errors.Is(err, errHostNoPermitido) {
			t.Fatalf("redirect ajeno = %v, se esperaba %v", err, errHostNoPermitido)
		}
	})

	t.Run(caso.nombre+"/redirect dentro de la allowlist seguido (barrera 3)", func(t *testing.T) {
		permitido := caso.destinoPermitido
		caso.transporte.responder = func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/inicio" {
				return respuestaFalsa(http.StatusFound, nil, map[string]string{"Location": permitido + "/final"}), nil
			}
			return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader("destino")), nil), nil
		}
		resp, err := caso.cliente.obtener(context.Background(), permitido+"/inicio")
		if err != nil {
			t.Fatalf("obtener() = %v, se esperaba nil", err)
		}
		if resp.Estado != http.StatusOK || string(resp.Cuerpo) != "destino" {
			t.Fatalf("respuesta = %d %q, se esperaba 200 \"destino\"", resp.Estado, resp.Cuerpo)
		}
	})

	t.Run(caso.nombre+"/IP privada, loopback y metadatos rechazadas (barrera 2)", func(t *testing.T) {
		for _, ip := range []string{"127.0.0.1", "10.0.0.5", "192.168.1.10", "172.16.4.4", "169.254.169.254", "::1", "fd00::1", "fe80::1", "0.0.0.0"} {
			caso.cliente.resolver = resolverFalso(ip)
			if _, err := caso.cliente.verificarDestino(context.Background(), "permitido.test"); !errors.Is(err, errIPNoPermitida) {
				t.Errorf("verificarDestino(%s) = %v, se esperaba %v", ip, err, errIPNoPermitida)
			}
		}
		caso.cliente.resolver = resolverFalso("93.184.216.34")
		if _, err := caso.cliente.verificarDestino(context.Background(), "permitido.test"); err != nil {
			t.Errorf("verificarDestino(IP pública) = %v, se esperaba nil", err)
		}
	})

	t.Run(caso.nombre+"/cuerpo por encima del tope cortado (barrera 4)", func(t *testing.T) {
		cuerpo := &cuerpoInfinito{tope: caso.cliente.limiteCuerpo}
		caso.transporte.responder = func(*http.Request) (*http.Response, error) {
			return respuestaFalsa(http.StatusOK, cuerpo, nil), nil
		}
		resp, err := caso.cliente.obtener(context.Background(), caso.destinoPermitido)
		if err != nil {
			t.Fatalf("obtener() = %v, se esperaba nil (el tope corta, no falla)", err)
		}
		if int64(len(resp.Cuerpo)) != caso.cliente.limiteCuerpo {
			t.Errorf("cuerpo = %d bytes, se esperaba el tope %d", len(resp.Cuerpo), caso.cliente.limiteCuerpo)
		}
		if !resp.Truncado {
			t.Errorf("Truncado = false, se esperaba true (la respuesta tenía más bytes)")
		}
		if cuerpo.leidos != cuerpo.tope {
			t.Errorf("se leyeron %d bytes del origen, se esperaba no pasar del tope %d", cuerpo.leidos, cuerpo.tope)
		}
	})

	t.Run(caso.nombre+"/petición que no responde falla por tiempo, no cuelga", func(t *testing.T) {
		caso.transporte.responder = func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		}
		inicio := time.Now()
		_, err := caso.cliente.obtener(context.Background(), caso.destinoPermitido)
		if err == nil {
			t.Fatalf("obtener() = nil, se esperaba un error por tiempo agotado")
		}
		if transcurrido := time.Since(inicio); transcurrido > 2*time.Second {
			t.Errorf("la petición tardó %v: el tiempo de espera no se aplicó", transcurrido)
		}
	})
}

// conPuerto devuelve la misma URL con el puerto indicado (barrera 1: la allowlist
// fija el host y el puerto).
func conPuerto(t *testing.T, destino, puerto string) string {
	t.Helper()
	u, err := url.Parse(destino)
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", destino, err)
	}
	u.Host = net.JoinHostPort(u.Hostname(), puerto)
	return u.String()
}

// clienteDePrueba arma el consumidor sintético de la tabla compartida.
func clienteDePrueba(t *testing.T, transporte *transporteFalso, hosts []string, limite int64, espera time.Duration) *clienteSaliente {
	t.Helper()
	cliente, err := nuevoClienteSaliente(configuracionCliente{
		nombre:          "prueba",
		hostsPermitidos: hosts,
		limiteCuerpo:    limite,
		tiempoEspera:    espera,
		transporte:      transporte,
		resolver:        resolverFalso("93.184.216.34"),
		dial:            (&dialFalso{}).dial,
	})
	if err != nil {
		t.Fatalf("nuevoClienteSaliente() = %v", err)
	}
	return cliente
}

// TestClienteSalienteBarreras fija la tabla compartida de barreras (P9-20/A1)
// sobre un consumidor sintético: es la misma que T908/T909/T910 ejecutan sobre
// feeds, probadores y carátulas.
func TestClienteSalienteBarreras(t *testing.T) {
	transporte := &transporteFalso{}
	cliente := clienteDePrueba(t, transporte, []string{"permitido.test"}, 1024, 50*time.Millisecond)

	ejecutarBarrerasCompartidas(t, casoBarreras{
		nombre:           "sintetico",
		cliente:          cliente,
		transporte:       transporte,
		destinoPermitido: "https://permitido.test/recurso",
		destinoAjeno:     "https://ajeno.test/recurso",
	})
}

// TestClienteSalienteDialResuelveYSoloConectaIPPublica fija la barrera 2 en la
// conexión real: el dialer resuelve el host y conecta a la IP pública resuelta
// (nunca al nombre ni a una IP privada).
func TestClienteSalienteDialResuelveYSoloConectaIPPublica(t *testing.T) {
	transporte := &transporteFalso{}
	cliente := clienteDePrueba(t, transporte, []string{"permitido.test"}, 1024, time.Second)

	dial := &dialFalso{}
	cliente.dial = dial.dial
	cliente.resolver = resolverFalso("93.184.216.34")

	transporte.responder = func(*http.Request) (*http.Response, error) {
		return respuestaFalsa(http.StatusOK, io.NopCloser(strings.NewReader("ok")), nil), nil
	}
	// El transporte falso no usa el dialer: se invoca la barrera directamente.
	dialContext := cliente.dialSeguro()
	if _, err := dialContext(context.Background(), "tcp", "permitido.test:443"); err == nil {
		t.Fatalf("dialSeguro() = nil, se esperaba el error del dialer falso")
	}
	if len(dial.direcciones) != 1 || dial.direcciones[0] != "93.184.216.34:443" {
		t.Fatalf("direcciones conectadas = %v, se esperaba [93.184.216.34:443]", dial.direcciones)
	}
}

// TestClienteSalienteDialRechazaHostAjeno fija la defensa en profundidad de la
// barrera 1 dentro del dialer: aunque el transporte no pase por CheckRedirect,
// no se abre una conexión a un host fuera de la allowlist.
func TestClienteSalienteDialRechazaHostAjeno(t *testing.T) {
	transporte := &transporteFalso{}
	cliente := clienteDePrueba(t, transporte, []string{"permitido.test"}, 1024, time.Second)

	dialContext := cliente.dialSeguro()
	if _, err := dialContext(context.Background(), "tcp", "ajeno.test:443"); !errors.Is(err, errHostNoPermitido) {
		t.Fatalf("dialSeguro(host ajeno) = %v, se esperaba %v", err, errHostNoPermitido)
	}
}

// TestClienteSalienteConfiguracionInvalida fija que la fábrica no se puede
// construir sin allowlist, sin tope de cuerpo ni sin tiempo de espera: las
// cuatro barreras son obligatorias (P9-20).
func TestClienteSalienteConfiguracionInvalida(t *testing.T) {
	casos := []configuracionCliente{
		{hostsPermitidos: []string{"a.test"}, limiteCuerpo: 1, tiempoEspera: time.Second},
		{nombre: "x", hostsPermitidos: nil, limiteCuerpo: 1, tiempoEspera: time.Second},
		{nombre: "x", hostsPermitidos: []string{""}, limiteCuerpo: 1, tiempoEspera: time.Second},
		{nombre: "x", hostsPermitidos: []string{"a.test"}, limiteCuerpo: 0, tiempoEspera: time.Second},
		{nombre: "x", hostsPermitidos: []string{"a.test"}, limiteCuerpo: 1},
	}
	for i, cfg := range casos {
		if _, err := nuevoClienteSaliente(cfg); err == nil {
			t.Errorf("caso %d: nuevoClienteSaliente() = nil, se esperaba un error de configuración", i)
		}
	}
}

// TestClienteSalienteAllowlistPorConsumidor fija que la allowlist es la DEL
// CONSUMIDOR: un host permitido para uno se rechaza en otro (así las carátulas
// no pueden leer los feeds ni al revés).
func TestClienteSalienteAllowlistPorConsumidor(t *testing.T) {
	transporte := &transporteFalso{}
	uno := clienteDePrueba(t, transporte, []string{"uno.test"}, 1024, time.Second)
	otro := clienteDePrueba(t, transporte, []string{"otro.test"}, 1024, time.Second)

	if _, err := uno.obtener(context.Background(), "https://otro.test/x"); !errors.Is(err, errHostNoPermitido) {
		t.Fatalf("host de otro consumidor = %v, se esperaba %v", err, errHostNoPermitido)
	}
	if _, err := otro.obtener(context.Background(), "https://uno.test/x"); !errors.Is(err, errHostNoPermitido) {
		t.Fatalf("host de otro consumidor = %v, se esperaba %v", err, errHostNoPermitido)
	}
}
