package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/session"
)

// withIdentity simula lo que hace authn (T227): deja la identidad resuelta en
// el contexto de la petición para que el guard pueda leerla.
func withIdentity(identity session.Identity, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(session.ContextWithIdentity(r.Context(), identity)))
	})
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// adminYAuthMux monta el guard SOLO en el grupo /api/v1/admin, como lo cablea
// routes.go (T228). Las rutas /api/v1/auth no lo montan: son las rutas
// blanqueadas de una cuenta con mustChangePassword.
func adminYAuthMux(identity session.Identity) *http.ServeMux {
	mux := http.NewServeMux()
	admin := PasswordGuard(nil)(okHandler())
	mux.Handle("/api/v1/admin/", withIdentity(identity, admin))
	blanqueadas := withIdentity(identity, okHandler())
	mux.Handle("/api/v1/auth/session", blanqueadas)
	mux.Handle("/api/v1/auth/logout", blanqueadas)
	mux.Handle("/api/v1/auth/password", blanqueadas)
	return mux
}

func doRequest(t *testing.T, mux *http.ServeMux, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestPasswordGuardDeniesAdminRoutesWhenChangeRequired(t *testing.T) {
	identity := session.Identity{UserID: uuid.New(), MustChangePassword: true}
	mux := adminYAuthMux(identity)

	rec := doRequest(t, mux, http.MethodGet, "/api/v1/admin/usuarios")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, se esperaba 403", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"reason":"password_change_required"`) {
		t.Fatalf("el detalle no señala el cambio obligatorio: %s", body)
	}
	if !strings.Contains(body, `"forbidden"`) {
		t.Fatalf("el sobre de error no es el esperado: %s", body)
	}
}

func TestPasswordGuardAllowsAdminRoutesWhenNoChangeRequired(t *testing.T) {
	identity := session.Identity{UserID: uuid.New(), MustChangePassword: false}
	mux := adminYAuthMux(identity)

	rec := doRequest(t, mux, http.MethodGet, "/api/v1/admin/usuarios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", rec.Code)
	}
}

func TestPasswordGuardLeavesAuthRoutesAvailable(t *testing.T) {
	identity := session.Identity{UserID: uuid.New(), MustChangePassword: true}
	mux := adminYAuthMux(identity)

	for _, path := range []string{"/api/v1/auth/session", "/api/v1/auth/logout", "/api/v1/auth/password"} {
		rec := doRequest(t, mux, http.MethodPost, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, se esperaba 200 (ruta blanqueada)", path, rec.Code)
		}
	}
}

func TestPasswordGuardWithoutIdentityIsUnauthenticated(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/admin/", PasswordGuard(nil)(okHandler()))

	rec := doRequest(t, mux, http.MethodGet, "/api/v1/admin/usuarios")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 sin identidad", rec.Code)
	}
}
