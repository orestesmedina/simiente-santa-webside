package httpserver

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want string
	}{
		{name: "base vacía", a: "", b: "/healthz", want: "/healthz"},
		{name: "prefijo simple", a: "/api", b: "/healthz", want: "/api/healthz"},
		{name: "prefijo con barra final", a: "/api/", b: "/healthz", want: "/api/healthz"},
		{name: "ruta sin barra inicial", a: "/api", b: "healthz", want: "/api/healthz"},
		{name: "ruta vacía", a: "/api", b: "", want: "/api"},
		{name: "ambos vacíos", a: "", b: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinPath(tt.a, tt.b); got != tt.want {
				t.Errorf("joinPath(%q, %q) = %q, se esperaba %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestHandlePublishesMethodAndPath(t *testing.T) {
	mux := http.NewServeMux()
	NewMuxRegistrar(mux).Handle(http.MethodGet, "/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	rec := httptest.NewRecorder()
	NewHandler(mux, discardLogger()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Errorf("respuesta = %d %q, se esperaba 200 ok", rec.Code, rec.Body.String())
	}
}

func TestGroupAccumulatesPrefixAndMiddleware(t *testing.T) {
	var order []string
	tag := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	mux := http.NewServeMux()
	root := NewMuxRegistrar(mux)

	t.Run("grupo simple: el primero de la lista es el más externo", func(t *testing.T) {
		order = nil
		g := root.Group("/api/v1", tag("g1"), tag("g2"))
		g.Handle(http.MethodGet, "/muestra", func(w http.ResponseWriter, _ *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

		rec := httptest.NewRecorder()
		NewHandler(mux, discardLogger()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/muestra", nil))

		want := []string{"g1", "g2", "handler"}
		if !reflect.DeepEqual(order, want) {
			t.Errorf("orden = %v, se esperaba %v", order, want)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, se esperaba 200", rec.Code)
		}
	})

	t.Run("grupo anidado acumula prefijo y middlewares heredados", func(t *testing.T) {
		order = nil
		admin := root.Group("/api/v1", tag("outer")).Group("/admin", tag("inner"))
		admin.Handle(http.MethodGet, "/items", func(w http.ResponseWriter, _ *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

		rec := httptest.NewRecorder()
		NewHandler(mux, discardLogger()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/items", nil))

		want := []string{"outer", "inner", "handler"}
		if !reflect.DeepEqual(order, want) {
			t.Errorf("orden = %v, se esperaba %v", order, want)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, se esperaba 200", rec.Code)
		}
	})
}

func TestEnvelopeFallbackConvertsStdlibFallbacks(t *testing.T) {
	mux := http.NewServeMux()
	NewMuxRegistrar(mux).Handle(http.MethodGet, "/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := NewHandler(mux, discardLogger())

	t.Run("ruta no documentada -> 404 con sobre", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-existe", nil))

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type = %q, se esperaba JSON (la stdlib responde texto)", ct)
		}
		env := decodeEnvelope(t, rec.Body.Bytes())
		if env.Error.Code != "not_found" {
			t.Errorf("code = %q, se esperaba not_found", env.Error.Code)
		}
	})

	t.Run("método no documentado -> 405 con sobre", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/healthz", nil))

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, se esperaba 405", rec.Code)
		}
		env := decodeEnvelope(t, rec.Body.Bytes())
		if env.Error.Code != "method_not_allowed" {
			t.Errorf("code = %q, se esperaba method_not_allowed", env.Error.Code)
		}
		if allow := rec.Header().Get("Allow"); allow == "" {
			t.Error("se perdió la cabecera Allow del 405 de la stdlib")
		}
	})
}
