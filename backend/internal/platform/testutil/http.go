package testutil

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
)

// Chain arma la cadena de middlewares de F1 en el orden aprobado:
// request-id → recover → logging → CORS. Delega en el único constructor
// (middleware.Chain) para que las pruebas no diverjan de la cadena real.
func Chain(logger *slog.Logger, allowedOrigins []string) []httpserver.Middleware {
	return middleware.Chain(logger, allowedOrigins)
}

// NewServer levanta un httptest.Server sobre el mux con los middlewares
// indicados y lo cierra al terminar la prueba.
func NewServer(t *testing.T, mux *http.ServeMux, logger *slog.Logger, mws ...httpserver.Middleware) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(httpserver.NewHandler(mux, logger, mws...))
	t.Cleanup(srv.Close)
	return srv
}

// NewServerWithChain levanta un httptest.Server con la cadena completa de F1
// (Chain), lista para las pruebas que ejercitan el stack entero.
func NewServerWithChain(t *testing.T, mux *http.ServeMux, logger *slog.Logger, allowedOrigins []string) *httptest.Server {
	t.Helper()
	return NewServer(t, mux, logger, Chain(logger, allowedOrigins)...)
}

// Response agrupa lo que las pruebas suelen inspeccionar de una respuesta.
type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

// Do ejecuta una petición contra srv y devuelve la respuesta con el cuerpo ya
// leído y cerrado. Falla la prueba si la petición no se puede completar.
func Do(t *testing.T, srv *httptest.Server, method, path, body string, headers map[string]string) Response {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, srv.URL+path, reader)
	if err != nil {
		t.Fatalf("construir petición %s %s: %v", method, path, err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("ejecutar petición %s %s: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("leer cuerpo de %s %s: %v", method, path, err)
	}
	return Response{Status: resp.StatusCode, Header: resp.Header, Body: payload}
}
