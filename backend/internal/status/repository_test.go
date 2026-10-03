//go:build integration

package status

import (
	"context"
	"io"
	"net"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/database"
	"simiente-santa/backend/internal/platform/testutil"
)

// waitPing reintenta hasta 10 s: tolera una base de datos que aún está
// arrancando y, tras una recuperación, el descarte de la conexión muerta.
func waitPing(t *testing.T, repo Repository) error {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	var err error
	for time.Now().Before(deadline) {
		if err = repo.Ping(context.Background()); err == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return err
}

func TestIntegrationPingLiveDatabase(t *testing.T) {
	pool := testutil.Pool(t, context.Background())
	repo := NewRepository(pool)

	if err := waitPing(t, repo); err != nil {
		t.Fatalf("Ping() con la BD viva devolvió error: %v", err)
	}
}

func TestIntegrationPingReflectsDownAndRecoveryWithoutRestart(t *testing.T) {
	dbURL := testutil.DatabaseURL(t) // omite la prueba si no hay BD configurada

	// El proxy TCP permite cortar y restaurar la conexión contra el mismo
	// repositorio y pool, sin reiniciar nada (escenario 4 de US2).
	proxy := newTCPProxy(t, databaseHostPort(t, dbURL))
	proxyURL := replaceHost(t, dbURL, proxy.address())

	pool, err := database.NewPool(context.Background(), proxyURL)
	if err != nil {
		t.Fatalf("NewPool() error: %v", err)
	}
	t.Cleanup(pool.Close)
	repo := NewRepository(pool)

	// 1. BD viva a través del proxy.
	if err := waitPing(t, repo); err != nil {
		t.Fatalf("Ping() con la BD viva devolvió error: %v", err)
	}

	// 2. Se corta la conexión: ping falla en ≤ timeout + holgura.
	proxy.stop()
	start := time.Now()
	downErr := repo.Ping(context.Background())
	elapsed := time.Since(start)
	if downErr == nil {
		t.Fatal("Ping() con la BD caída no devolvió error")
	}
	if elapsed > database.PingTimeout+time.Second {
		t.Errorf("Ping() con la BD caída tardó %v, se esperaba ≤%v", elapsed, database.PingTimeout+time.Second)
	}

	// 3. Se restaura en la misma dirección: el mismo pool vuelve a responder.
	if err := proxy.start(); err != nil {
		t.Fatalf("restaurar proxy: %v", err)
	}
	if err := waitPing(t, repo); err != nil {
		t.Fatalf("Ping() tras recuperar la BD devolvió error: %v", err)
	}
}

// databaseHostPort extrae host:puerto de DATABASE_URL_TEST.
func databaseHostPort(t *testing.T, dbURL string) string {
	t.Helper()
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("parsear DATABASE_URL_TEST: %v", err)
	}
	return net.JoinHostPort(cfg.ConnConfig.Host, strconv.Itoa(int(cfg.ConnConfig.Port)))
}

// replaceHost devuelve dbURL apuntando a otro host:puerto (el proxy).
func replaceHost(t *testing.T, dbURL, hostPort string) string {
	t.Helper()
	u, err := url.Parse(dbURL)
	if err != nil {
		t.Fatalf("parsear DATABASE_URL_TEST: %v", err)
	}
	u.Host = hostPort
	return u.String()
}

// tcpProxy es un proxy TCP mínimo que reenvía a la base de datos y se puede
// detener y volver a levantar en la misma dirección. Simula la caída y la
// recuperación de la BD sin tocar el contenedor ni reiniciar el backend.
type tcpProxy struct {
	target string
	addr   string

	mu    sync.Mutex
	ln    net.Listener
	conns []net.Conn
}

func newTCPProxy(t *testing.T, target string) *tcpProxy {
	t.Helper()
	p := &tcpProxy{target: target, addr: "127.0.0.1:0"}
	if err := p.start(); err != nil {
		t.Fatalf("iniciar proxy TCP: %v", err)
	}
	t.Cleanup(p.stop)
	return p
}

func (p *tcpProxy) address() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.addr
}

// start escucha en la dirección actual (la primera vez, un puerto efímero; en
// las siguientes, el que ya quedó fijado).
func (p *tcpProxy) start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		return err
	}
	p.ln = ln
	p.addr = ln.Addr().String()
	go p.accept()
	return nil
}

// stop cierra el listener y todas las conexiones en curso, de modo que el pool
// deje de tener una conexión utilizable.
func (p *tcpProxy) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ln != nil {
		_ = p.ln.Close()
		p.ln = nil
	}
	for _, conn := range p.conns {
		_ = conn.Close()
	}
	p.conns = nil
}

func (p *tcpProxy) accept() {
	for {
		client, err := p.ln.Accept()
		if err != nil {
			return
		}
		upstream, err := net.Dial("tcp", p.target)
		if err != nil {
			_ = client.Close()
			continue
		}
		p.track(client, upstream)
		go p.pipe(client, upstream)
		go p.pipe(upstream, client)
	}
}

func (p *tcpProxy) track(conns ...net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conns = append(p.conns, conns...)
}

func (p *tcpProxy) pipe(dst, src net.Conn) {
	_, _ = io.Copy(dst, src)
	_ = dst.Close()
	_ = src.Close()
}
