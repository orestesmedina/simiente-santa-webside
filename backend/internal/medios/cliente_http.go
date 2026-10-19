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
	"time"
)

// cliente_http.go es la fábrica ÚNICA del http.Client con el que el dominio
// `medios` sale a internet (P9-20/R9-8, corrección A1 de la fase 5). Es la
// primera red saliente del backend y `seguridad` la audita entera: las cuatro
// barreras viven aquí, en un solo sitio imposible de olvidar en un consumidor
// nuevo, y las aplican los TRES consumidores —feeds (T908), probadores de
// disponibilidad (T909) y descarga de carátulas (T910)— con la MISMA tabla de
// pruebas.
//
// Barreras, aplicadas por construcción a toda petición (también tras cada
// redirect):
//
//  1. solo `https`, host de la allowlist DEL CONSUMIDOR (cada consumidor
//     declara la suya: feeds, oEmbed/episodio y carátulas) **y el puerto 443**
//     (sin puerto, `https` implica el 443): una URL con `:8443` no sale;
//  2. ninguna IP privada/loopback/link-local —incluido el servicio de metadatos
//     `169.254.169.254`—: el dialer resuelve el DNS y rechaza el destino en
//     CADA conexión, de modo que un feed comprometido no alcanza la red interna;
//  3. `CheckRedirect` valida cada `Location` contra la allowlist del consumidor
//     (fuera de ella → error; además net/http no sigue más de 10 saltos);
//  4. `io.LimitReader` en TODO cuerpo leído, con el tope del consumidor: nunca
//     se lee un cuerpo entero sin tope.
//
// Además: `User-Agent` propio en todas las peticiones y tiempos de espera de
// conexión, TLS, cabeceras y petición completa. Sin variables de entorno nuevas
// (P9-18): la allowlist, los topes y los tiempos son constantes revisables del
// dominio que cada consumidor pasa a la fábrica.

// Errores tipados de la red saliente (los consumidores los distinguen con
// errors.Is y el resumen de la búsqueda nunca expone su texto técnico).
var (
	errEsquemaNoPermitido      = errors.New("medios: la red saliente solo admite https")
	errHostNoPermitido         = errors.New("medios: host fuera de la allowlist del consumidor")
	errPuertoNoPermitido       = errors.New("medios: la red saliente solo admite el puerto 443")
	errIPNoPermitida           = errors.New("medios: el destino resuelve a una dirección no pública")
	errDemasiadasRedirecciones = errors.New("medios: demasiadas redirecciones")
	errConfiguracionCliente    = errors.New("medios: configuración de la red saliente inválida")
)

const (
	// usuarioAgente identifica al sitio en todas las peticiones salientes
	// (P9-20): los servicios externos ven quién les pide.
	usuarioAgente = "SimienteSanta-Medios/1.0 (+https://iglesiasimientasanta.org)"

	// maxRedirecciones es el tope de saltos que sigue el cliente (barrera 3);
	// `net/http` tampoco sigue más de 10.
	maxRedirecciones = 10

	// puertoPermitido es el ÚNICO puerto por el que sale la red saliente (B2):
	// la allowlist fija el host **y** el puerto, de modo que una carátula del
	// feed con `https://i.ytimg.com:8443/x.jpg` no puede hacer que el backend
	// conecte a un puerto arbitrario de un host permitido. Sin puerto en la URL,
	// `https` implica el 443.
	puertoPermitido = "443"

	// Tiempos de espera del transporte (P9-3/P9-18: 10/10/5 s por petición,
	// constantes revisables del dominio). El tiempo total por petición lo fija
	// cada consumidor en su configuración.
	tiempoEsperaConexion    = 10 * time.Second
	tiempoEsperaTLS         = 10 * time.Second
	tiempoEsperaCabeceras   = 10 * time.Second
	tiempoEsperaInactividad = 30 * time.Second
)

// configuracionCliente reúne las barreras de UN consumidor de la red saliente.
type configuracionCliente struct {
	// nombre identifica al consumidor en los mensajes de error ("feeds",
	// "probador", "caratulas").
	nombre string
	// hostsPermitidos es la allowlist EXACTA de hosts del consumidor (sin
	// puerto; en minúsculas). Un host fuera de ella no se pide jamás.
	hostsPermitidos []string
	// limiteCuerpo es el tope duro de bytes que se leen de cada cuerpo.
	limiteCuerpo int64
	// tiempoEspera es el tiempo máximo de una petición completa (P9-3).
	tiempoEspera time.Duration

	// transporte, resolver y dial solo se rellenan en las PRUEBAS (nunca en
	// producción): permiten ejercer las barreras sin tocar la red.
	transporte http.RoundTripper
	resolver   func(ctx context.Context, host string) ([]netip.Addr, error)
	dial       func(ctx context.Context, network, address string) (net.Conn, error)
}

// clienteSaliente es el cliente HTTP saliente del dominio: la fábrica única
// (P9-20). Los consumidores solo pueden pedir por `obtener`, que aplica las
// cuatro barreras.
type clienteSaliente struct {
	nombre       string
	hosts        map[string]struct{}
	limiteCuerpo int64
	cliente      *http.Client

	// Inyectables en pruebas.
	resolver func(ctx context.Context, host string) ([]netip.Addr, error)
	dial     func(ctx context.Context, network, address string) (net.Conn, error)
}

// respuestaSaliente es el resultado acotado de una petición saliente.
type respuestaSaliente struct {
	// Estado es el código HTTP de la respuesta.
	Estado int
	// Cuerpo son los bytes leídos, como máximo `limiteCuerpo`.
	Cuerpo []byte
	// Truncado indica que la lectura llegó al tope: la respuesta puede tener
	// más bytes y NUNCA se leyó entera (barrera 4). La descarga de carátulas lo
	// trata como rechazo.
	Truncado bool
}

// nuevoClienteSaliente construye el cliente de un consumidor. Falla si falta
// alguna de las barreras: sin allowlist, sin tope de cuerpo o sin tiempo de
// espera no hay cliente (es una configuración inválida del dominio).
func nuevoClienteSaliente(cfg configuracionCliente) (*clienteSaliente, error) {
	if strings.TrimSpace(cfg.nombre) == "" {
		return nil, fmt.Errorf("%w: falta el nombre del consumidor", errConfiguracionCliente)
	}
	if len(cfg.hostsPermitidos) == 0 {
		return nil, fmt.Errorf("%w: %s sin allowlist", errConfiguracionCliente, cfg.nombre)
	}
	hosts := make(map[string]struct{}, len(cfg.hostsPermitidos))
	for _, host := range cfg.hostsPermitidos {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			return nil, fmt.Errorf("%w: %s con un host vacío en la allowlist", errConfiguracionCliente, cfg.nombre)
		}
		hosts[host] = struct{}{}
	}
	if cfg.limiteCuerpo <= 0 {
		return nil, fmt.Errorf("%w: %s sin tope de cuerpo", errConfiguracionCliente, cfg.nombre)
	}
	if cfg.tiempoEspera <= 0 {
		return nil, fmt.Errorf("%w: %s sin tiempo de espera", errConfiguracionCliente, cfg.nombre)
	}

	c := &clienteSaliente{
		nombre:       cfg.nombre,
		hosts:        hosts,
		limiteCuerpo: cfg.limiteCuerpo,
		resolver:     cfg.resolver,
		dial:         cfg.dial,
	}

	transporte := cfg.transporte
	if transporte == nil {
		transporte = c.transporteSeguro()
	}
	c.cliente = &http.Client{
		Transport:     transporte,
		Timeout:       cfg.tiempoEspera,
		CheckRedirect: c.checkRedirect,
	}
	return c, nil
}

// transporteSeguro arma el transporte real: sin proxy (un proxy del entorno
// podría saltarse las barreras), con el dialer que resuelve y valida el destino
// y con tiempos de espera acotados.
func (c *clienteSaliente) transporteSeguro() *http.Transport {
	return &http.Transport{
		Proxy:                 nil,
		DialContext:           c.dialSeguro(),
		TLSHandshakeTimeout:   tiempoEsperaTLS,
		ResponseHeaderTimeout: tiempoEsperaCabeceras,
		ExpectContinueTimeout: tiempoEsperaConexion,
		IdleConnTimeout:       tiempoEsperaInactividad,
		MaxIdleConns:          4,
	}
}

// dialSeguro devuelve la barrera 2: en CADA conexión —también tras cada
// redirect— se comprueba el host contra la allowlist, se exige el puerto
// permitido, se resuelve su DNS y se rechaza cualquier dirección que no sea
// pública. Solo entonces se conecta a la IP resuelta (nunca al nombre) y al
// puerto permitido (nunca al que traiga la petición).
func (c *clienteSaliente) dialSeguro() func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, puerto, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("medios: dirección saliente inválida %q: %w", address, err)
		}
		if _, ok := c.hosts[strings.ToLower(host)]; !ok {
			return nil, fmt.Errorf("%w: %q", errHostNoPermitido, host)
		}
		if puerto != puertoPermitido {
			return nil, fmt.Errorf("%w: %q", errPuertoNoPermitido, puerto)
		}
		ips, err := c.verificarDestino(ctx, host)
		if err != nil {
			return nil, err
		}

		dial := c.dial
		if dial == nil {
			dialer := &net.Dialer{Timeout: tiempoEsperaConexion, KeepAlive: tiempoEsperaInactividad}
			dial = dialer.DialContext
		}
		return dial(ctx, network, net.JoinHostPort(ips[0].String(), puertoPermitido))
	}
}

// verificarDestino resuelve el host y exige que TODAS sus direcciones sean
// públicas (barrera 2). Devuelve las direcciones para que el dialer conecte a
// la primera.
func (c *clienteSaliente) verificarDestino(ctx context.Context, host string) ([]netip.Addr, error) {
	resolver := c.resolver
	if resolver == nil {
		resolver = func(ctx context.Context, host string) ([]netip.Addr, error) {
			return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		}
	}
	ips, err := resolver(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("medios: no se pudo resolver %q: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("medios: %q no resuelve a ninguna dirección", host)
	}
	for i, ip := range ips {
		ips[i] = ip.Unmap()
		if !ipPublica(ips[i]) {
			return nil, fmt.Errorf("%w: %s resuelve a %s", errIPNoPermitida, host, ips[i])
		}
	}
	return ips, nil
}

// ipPublica indica si la dirección es de uso público: unicast global y ni
// privada, ni loopback, ni link-local (metadatos), ni no especificada, ni
// multicast.
func ipPublica(ip netip.Addr) bool {
	if !ip.IsValid() || !ip.IsGlobalUnicast() {
		return false
	}
	switch {
	case ip.IsUnspecified(), ip.IsLoopback(), ip.IsPrivate(),
		ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(),
		ip.IsMulticast(), ip.IsInterfaceLocalMulticast():
		return false
	default:
		return true
	}
}

// checkRedirect es la barrera 3: cada `Location` se valida contra la allowlist
// del consumidor y se corta el seguimiento pasado el tope de saltos.
func (c *clienteSaliente) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirecciones {
		return fmt.Errorf("%w: más de %d saltos", errDemasiadasRedirecciones, maxRedirecciones)
	}
	return c.verificarURL(req.URL)
}

// verificarURL es la barrera 1: solo `https`, hosts de la allowlist del
// consumidor y el puerto permitido (vacío o 443: la allowlist fija el host **y**
// el puerto, B2).
func (c *clienteSaliente) verificarURL(u *url.URL) error {
	if u == nil {
		return fmt.Errorf("%w: sin URL", errHostNoPermitido)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("%w: esquema %q", errEsquemaNoPermitido, u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	if _, ok := c.hosts[host]; !ok {
		return fmt.Errorf("%w: %q", errHostNoPermitido, host)
	}
	if puerto := u.Port(); puerto != "" && puerto != puertoPermitido {
		return fmt.Errorf("%w: %q", errPuertoNoPermitido, puerto)
	}
	return nil
}

// obtener hace un GET aplicando las cuatro barreras y devuelve el estado y el
// cuerpo acotado. Es la ÚNICA vía de salida a internet del dominio: los tres
// consumidores pasan por aquí.
func (c *clienteSaliente) obtener(ctx context.Context, destino string) (*respuestaSaliente, error) {
	u, err := url.Parse(destino)
	if err != nil {
		return nil, fmt.Errorf("medios: URL saliente inválida %q: %w", destino, err)
	}
	if err := c.verificarURL(u); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("medios: construir la petición saliente: %w", err)
	}
	req.Header.Set("User-Agent", usuarioAgente)

	resp, err := c.cliente.Do(req)
	if err != nil {
		return nil, fmt.Errorf("medios: %s: %w", c.nombre, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Barrera 4: el cuerpo SIEMPRE se lee a través del tope del consumidor.
	cuerpo, err := io.ReadAll(io.LimitReader(resp.Body, c.limiteCuerpo))
	if err != nil {
		return nil, fmt.Errorf("medios: leer el cuerpo de %s: %w", c.nombre, err)
	}
	return &respuestaSaliente{
		Estado:   resp.StatusCode,
		Cuerpo:   cuerpo,
		Truncado: int64(len(cuerpo)) >= c.limiteCuerpo,
	}, nil
}
