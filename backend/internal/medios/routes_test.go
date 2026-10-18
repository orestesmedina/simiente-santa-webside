package medios

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de routes.go (T928): RegisterPublic publica la superficie pública del
// dominio en sus rutas del contrato, sobre un mux real. Las rutas de panel (con su
// cadena AdminChain y el permiso `medios`) se cubren en el humo de cmd/api, que las
// alcanza a través de la cadena completa y con sesión viva.

// newRoutesServer publica la superficie pública de medios sobre un mux real.
func newRoutesServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	h := newTestHandler(t, newFakeRepository())

	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	RegisterPublic(root, h)
	return testutil.NewServer(t, mux, logger)
}

func TestRegisterPublicRoutes(t *testing.T) {
	srv := newRoutesServer(t)

	t.Run("la sección responde 200 sin sesión y sin caché", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/medios", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
		// Sin nada publicable: `items` es un arreglo vacío y el sobre está
		// completo (la redirección de la SPA es del frontend, P9-16).
		if body := string(resp.Body); body != `{"lang":"es","items":[],"total":0,"limit":12,"offset":0}` {
			t.Errorf("cuerpo = %s", body)
		}
	})

	t.Run("el bloque de la portada responde 200 sin sesión y sin caché", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/medios", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
		// Sin publicables: la clave `items` se omite y la portada no dibuja el
		// bloque (FR-027).
		if body := string(resp.Body); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang", body)
		}
	})

	t.Run("lang inválido responde 400 en las dos rutas", func(t *testing.T) {
		for _, path := range []string{"/api/v1/medios?lang=fr", "/api/v1/portada/medios?lang=fr"} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusBadRequest {
				t.Fatalf("GET %s: status = %d, se esperaba 400 (%s)", path, resp.Status, resp.Body)
			}
		}
	})

	t.Run("el bloque no admite límite", func(t *testing.T) {
		// El tope 5 es del servidor: `limit` en la query se ignora.
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/medios?limit=50", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("subruta no registrada responde 404", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/medios/extra", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("sin handler no publica nada", func(t *testing.T) {
		mux := http.NewServeMux()
		RegisterPublic(httpserver.NewMuxRegistrar(mux), nil)
		logger, _ := testutil.NewLogger()
		srv := testutil.NewServer(t, mux, logger)
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/medios", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404", resp.Status)
		}
	})
}

// TestNewRedSaliente fija la fábrica exportada de la red saliente (T928): los tres
// consumidores salen de ella con sus barreras y un tope imposible es un error de
// composición (nunca un consumidor a medias).
func TestNewRedSaliente(t *testing.T) {
	red, err := NewRedSaliente(&fakeStore{}, 1<<20)
	if err != nil {
		t.Fatalf("NewRedSaliente: %v", err)
	}
	if red.Feeds == nil || red.Probador == nil || red.Caratulas == nil {
		t.Fatalf("la fábrica dejó un consumidor sin construir: %+v", red)
	}

	if _, err := NewRedSaliente(&fakeStore{}, 0); err == nil {
		t.Error("un tope de carátulas imposible debía ser un error de composición")
	}
}
