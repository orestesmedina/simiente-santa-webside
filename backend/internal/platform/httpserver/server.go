package httpserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"simiente-santa/backend/internal/platform/apperr"
)

// Timeouts por defecto del servidor. Se aplican cuando Options los deja a
// cero; el llamador (cmd/api) los puede sobrescribir.
const (
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultReadTimeout       = 10 * time.Second
	DefaultWriteTimeout      = 15 * time.Second
	DefaultIdleTimeout       = 60 * time.Second
	DefaultShutdownTimeout   = 10 * time.Second
)

// Options configura el servidor HTTP.
type Options struct {
	// Addr es la dirección de escucha (":8080" en producción, "127.0.0.1:0"
	// para un puerto efímero en pruebas).
	Addr string
	// ReadHeaderTimeout acota la lectura de cabeceras (mitiga Slowloris).
	ReadHeaderTimeout time.Duration
	// ReadTimeout acota la lectura completa de la petición.
	ReadTimeout time.Duration
	// WriteTimeout acota la escritura de la respuesta.
	WriteTimeout time.Duration
	// IdleTimeout acota las conexiones keep-alive inactivas.
	IdleTimeout time.Duration
	// ShutdownTimeout es el periodo de gracia del apagado ordenado.
	ShutdownTimeout time.Duration
}

// Server envuelve el *http.Server con el apagado ordenado y el registro del
// ciclo de vida.
type Server struct {
	httpServer      *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration

	mu   sync.Mutex
	addr string
}

// New construye el servidor sobre mux, con la cadena de middlewares en el orden
// recibido (el primero, el más externo) y el fallback 404/405 convertido a
// sobre de error. No abre el socket: eso ocurre en Run.
func New(mux *http.ServeMux, opts Options, logger *slog.Logger, mws ...Middleware) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	opts = withDefaults(opts)

	srv := &http.Server{
		Addr:              opts.Addr,
		Handler:           NewHandler(mux, logger, mws...),
		ReadHeaderTimeout: opts.ReadHeaderTimeout,
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout,
		IdleTimeout:       opts.IdleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	return &Server{
		httpServer:      srv,
		logger:          logger,
		shutdownTimeout: opts.ShutdownTimeout,
		addr:            opts.Addr,
	}
}

// NewHandler compone el handler final: el mux con su fallback 404/405 en sobre
// de error, envuelto por los middlewares en el orden recibido (el primero, el
// más externo). Lo usan New y las pruebas del stack HTTP.
func NewHandler(mux *http.ServeMux, logger *slog.Logger, mws ...Middleware) http.Handler {
	var handler = envelopeFallback(mux, logger)
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// envelopeFallback convierte el 404/405 que la stdlib responde en texto plano
// en el sobre de error (SC-008). Solo actúa cuando el ServeMux no casó ninguna
// ruta (pattern == ""): las respuestas de los handlers —incluidos sus
// WriteError— pasan intactas.
//
// Cuando la ruta SÍ casa se delega en mux.ServeHTTP y no en el handler devuelto
// por mux.Handler: es ServeHTTP quien rellena los valores de ruta
// (`r.PathValue("id")` de los patrones `/{id}`). Llamar al handler directamente
// dejaría `PathValue` vacío.
func envelopeFallback(mux *http.ServeMux, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}

		rec := &fallbackRecorder{header: make(http.Header)}
		h.ServeHTTP(rec, r)

		ctx := r.Context()
		if rec.status == http.StatusMethodNotAllowed {
			if allow := rec.header.Get("Allow"); allow != "" {
				w.Header().Set("Allow", allow)
			}
			WriteError(ctx, w, logger, apperr.MethodNotAllowed("Método no permitido"))
			return
		}
		WriteError(ctx, w, logger, apperr.NotFound("Recurso no encontrado"))
	})
}

// fallbackRecorder captura la respuesta de texto de la stdlib para decidir el
// sobre de error sin escribir nada al cliente.
type fallbackRecorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *fallbackRecorder) Header() http.Header { return r.header }

func (r *fallbackRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
}

func (r *fallbackRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(b)
}

// Addr devuelve la dirección real de escucha una vez arrancado (útil cuando se
// pidió un puerto efímero "127.0.0.1:0"), o la configurada antes de arrancar.
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

func (s *Server) setAddr(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addr = addr
}

// Run arranca el servidor y lo mantiene hasta que el contexto se cancela o llega
// SIGINT/SIGTERM. Entonces hace un apagado ordenado (Shutdown) con el periodo de
// gracia configurado y devuelve nil. Un fallo al escuchar o servir se devuelve
// como error.
func (s *Server) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen %q: %w", s.httpServer.Addr, err)
	}
	s.setAddr(ln.Addr().String())

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("servidor escuchando", slog.String("addr", s.Addr()))
		if err := s.httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	s.logger.Info("apagando servidor", slog.Duration("timeout", s.shutdownTimeout))
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	s.logger.Info("servidor detenido")
	return nil
}

// withDefaults completa con los valores por defecto los campos sin configurar.
func withDefaults(opts Options) Options {
	if opts.ReadHeaderTimeout == 0 {
		opts.ReadHeaderTimeout = DefaultReadHeaderTimeout
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = DefaultReadTimeout
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = DefaultWriteTimeout
	}
	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = DefaultIdleTimeout
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = DefaultShutdownTimeout
	}
	return opts
}
