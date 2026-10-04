package testutil

import (
	"context"
	"net/http"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
)

func TestPoolSkipsWithoutDatabaseURL(t *testing.T) {
	t.Setenv(DatabaseURLEnv, "")

	reached := false
	t.Run("sin DATABASE_URL_TEST", func(t *testing.T) {
		_ = Pool(t, context.Background())
		reached = true
	})

	if reached {
		t.Error("Pool no omitió la prueba sin DATABASE_URL_TEST")
	}
}

func TestLogCapture(t *testing.T) {
	logger, capture := NewLogger()

	logger.With("service", "api").Info("petición atendida", "status", 200, "request_id", "req-1")

	if capture.Count() != 1 {
		t.Fatalf("registros capturados = %d, se esperaba 1", capture.Count())
	}
	record, ok := capture.Find("petición atendida")
	if !ok {
		t.Fatal("Find no encontró el registro")
	}
	want := map[string]any{
		"service":    "api",
		"status":     int64(200),
		"request_id": "req-1",
	}
	for key, value := range want {
		if record.Attrs[key] != value {
			t.Errorf("atributo %q = %v, se esperaba %v", key, record.Attrs[key], value)
		}
	}
}

func TestHTTPServerWithChain(t *testing.T) {
	logger, _ := NewLogger()
	mux := http.NewServeMux()
	httpserver.NewMuxRegistrar(mux).Handle(http.MethodGet, "/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpserver.WriteJSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})
	srv := NewServerWithChain(t, mux, logger, []string{"http://localhost:5173"})

	t.Run("éxito con la cadena completa", func(t *testing.T) {
		resp := Do(t, srv, http.MethodGet, "/healthz", "", nil)

		if resp.Status != http.StatusOK {
			t.Errorf("status = %d, se esperaba 200", resp.Status)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Error("la cadena no propagó X-Request-ID")
		}
		if got := string(resp.Body); got != `{"status":"ok"}` {
			t.Errorf("cuerpo = %q, se esperaba el sobre de éxito", got)
		}
	})

	t.Run("ruta no documentada con la cadena completa", func(t *testing.T) {
		resp := Do(t, srv, http.MethodGet, "/no-existe", "", nil)

		if resp.Status != http.StatusNotFound {
			t.Errorf("status = %d, se esperaba 404", resp.Status)
		}
	})
}
