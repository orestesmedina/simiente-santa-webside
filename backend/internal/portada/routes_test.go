package portada

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de routes.go (T326): comprueban que RegisterPublic y RegisterAdmin
// publican cada operación del contrato. Las rutas del panel se alcanzan a través
// de la cadena real: sin sesión responden 401 (no 404), lo que prueba que están
// registradas.

// routeSessionStore no tiene sesiones: cualquier token es inexistente.
type routeSessionStore struct{}

func (routeSessionStore) Create(context.Context, uuid.UUID) (string, error) { return "", nil }
func (routeSessionStore) Resolve(context.Context, string) (session.Session, error) {
	return session.Session{}, session.ErrSessionNotFound
}
func (routeSessionStore) Revoke(context.Context, string) error                      { return nil }
func (routeSessionStore) RevokeUser(context.Context, uuid.UUID) error               { return nil }
func (routeSessionStore) RevokeUserExcept(context.Context, uuid.UUID, string) error { return nil }

// routeResolver nunca se alcanza sin cookie válida.
type routeResolver struct{}

func (routeResolver) Resolve(context.Context, uuid.UUID) (session.Identity, error) {
	return session.Identity{}, session.ErrSessionNotFound
}

// newRoutesServer publica la superficie de la portada (y, opcionalmente, el
// grupo de panel) sobre un mux real.
func newRoutesServer(t *testing.T, withAdmin bool) *httptest.Server {
	t.Helper()
	logger, _ := testutil.NewLogger()
	svc := NewService(ServiceDeps{Repository: newFakeRepository(), Logger: logger})
	h := NewHandler(HandlerDeps{Service: svc, MaxUploadBytes: 1 << 20, Logger: logger})

	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	RegisterPublic(root, h)
	if withAdmin {
		RegisterAdmin(root, h, AdminDeps{Deps: middleware.AdminDeps{
			Sessions:   routeSessionStore{},
			Resolver:   routeResolver{},
			CSRFSecret: "secreto-de-prueba",
			Logger:     logger,
		}})
	}
	return testutil.NewServer(t, mux, logger)
}

func TestRegisterPublicRoutes(t *testing.T) {
	srv := newRoutesServer(t, false)

	t.Run("portada pública responde 200", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("media con nombre inválido responde 400", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/media/no-file", "", nil)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
	})
}

func TestRegisterAdminRoutes(t *testing.T) {
	srv := newRoutesServer(t, true)
	id := uuid.New().String()

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/admin/portada"},
		{http.MethodPut, "/api/v1/admin/portada/identidad"},
		{http.MethodPut, "/api/v1/admin/portada/quienes-somos"},
		{http.MethodPut, "/api/v1/admin/portada/contacto"},
		{http.MethodPost, "/api/v1/admin/portada/imagenes"},
		{http.MethodPost, "/api/v1/admin/portada/horario"},
		{http.MethodPatch, "/api/v1/admin/portada/horario/" + id},
		{http.MethodDelete, "/api/v1/admin/portada/horario/" + id},
		{http.MethodPost, "/api/v1/admin/portada/whatsapp"},
		{http.MethodPatch, "/api/v1/admin/portada/whatsapp/" + id},
		{http.MethodDelete, "/api/v1/admin/portada/whatsapp/" + id},
		{http.MethodPost, "/api/v1/admin/portada/redes"},
		{http.MethodPatch, "/api/v1/admin/portada/redes/" + id},
		{http.MethodDelete, "/api/v1/admin/portada/redes/" + id},
	}
	for _, route := range routes {
		resp := testutil.Do(t, srv, route.method, route.path, "", nil)
		if resp.Status != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, se esperaba 401 (%s)", route.method, route.path, resp.Status, resp.Body)
		}
	}

	t.Run("subruta no registrada responde 404", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/portada/desconocido", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
	})
}
