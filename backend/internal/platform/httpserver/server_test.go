package httpserver

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestNewAppliesDefaultTimeouts(t *testing.T) {
	srv := New(http.NewServeMux(), Options{}, discardLogger())

	if srv.httpServer.ReadHeaderTimeout != DefaultReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, se esperaba %v", srv.httpServer.ReadHeaderTimeout, DefaultReadHeaderTimeout)
	}
	if srv.httpServer.ReadTimeout != DefaultReadTimeout {
		t.Errorf("ReadTimeout = %v, se esperaba %v", srv.httpServer.ReadTimeout, DefaultReadTimeout)
	}
	if srv.httpServer.WriteTimeout != DefaultWriteTimeout {
		t.Errorf("WriteTimeout = %v, se esperaba %v", srv.httpServer.WriteTimeout, DefaultWriteTimeout)
	}
	if srv.httpServer.IdleTimeout != DefaultIdleTimeout {
		t.Errorf("IdleTimeout = %v, se esperaba %v", srv.httpServer.IdleTimeout, DefaultIdleTimeout)
	}
	if srv.shutdownTimeout != DefaultShutdownTimeout {
		t.Errorf("shutdownTimeout = %v, se esperaba %v", srv.shutdownTimeout, DefaultShutdownTimeout)
	}
}

func TestRunServesAndShutsDownCleanly(t *testing.T) {
	mux := http.NewServeMux()
	NewMuxRegistrar(mux).Handle(http.MethodGet, "/healthz", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})

	srv := New(mux, Options{Addr: "127.0.0.1:0", ShutdownTimeout: time.Second}, discardLogger())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	addr := waitForAddr(t, srv)

	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, se esperaba 200", resp.StatusCode)
	}
	if len(body) == 0 {
		t.Error("respuesta vacía del servidor en marcha")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() devolvió error en el apagado: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() no terminó tras cancelar el contexto")
	}

	if _, err := http.Get("http://" + addr + "/healthz"); err == nil {
		t.Error("el servidor sigue aceptando peticiones tras el Shutdown")
	}
}

// waitForAddr espera a que Run publique la dirección real del puerto efímero.
func waitForAddr(t *testing.T, srv *Server) string {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if addr := srv.Addr(); addr != "127.0.0.1:0" && addr != "" {
			return addr
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("el servidor no publicó su dirección a tiempo (Addr=%q)", srv.Addr())
	return ""
}
