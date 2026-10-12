package agenda

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de routes.go (T446): RegisterPublic publica la superficie pública del
// dominio en sus rutas del contrato, sobre un mux real. Las rutas de panel (con
// su cadena AdminChain) se cubren en el humo de cmd/api, que las alcanza a través
// de la cadena completa.

// newRoutesServer publica la superficie pública de la agenda sobre un mux real.
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

	t.Run("lo que viene responde 200", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/proximamente", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("el extracto de la portada responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/proximamente", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
		if body := string(resp.Body); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang sin ítems", body)
		}
	})

	t.Run("lang inválido en el extracto responde 400", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/proximamente?lang=fr", "", nil)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("subruta no registrada responde 404", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/desconocido", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
	})
}
