package media

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de routes.go (T415): RegisterPublic publica la descarga pública en su
// ruta del contrato (`GET /api/v1/media/{fileName}`) sobre un mux real.

// newRoutesServer publica la superficie pública de media sobre un mux real.
func newRoutesServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	repo := newFakeRepository()
	store := &fakeStore{files: map[string]string{}}
	svc := NewService(ServiceDeps{Repository: repo, Store: store, Logger: logger})
	h := NewHandler(HandlerDeps{Service: svc, Logger: logger})

	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	RegisterPublic(root, h)
	return testutil.NewServer(t, mux, logger)
}

func TestRegisterPublicRoutes(t *testing.T) {
	srv := newRoutesServer(t)

	t.Run("nombre inválido responde 400", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/media/no-file", "", nil)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("nombre válido no publicado responde 404", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/media/"+validName, "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("subruta no registrada responde 404", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/media/", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
	})
}
