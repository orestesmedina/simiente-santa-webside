// suite de aceptación del sobre de respuestas (SC-008, SC-009).
//
// Monta el stack completo —router Registrar + cadena de middlewares + handlers,
// incluidos los reales del dominio status— y provoca todas las formas de
// respuesta para comprobar que el 100 % cumple el formato uniforme (D13,
// SC-008) y que un error inesperado no filtra información interna (FR-013,
// SC-009). Es un paquete de prueba externo (httpserver_test) para poder
// importar el dominio status sin crear un ciclo; el código de platform no
// conoce dominios (arq. R1).
package httpserver_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/testutil"
	"simiente-santa/backend/internal/status"
)

// secretInternal simula un error interno con datos sensibles (SQL, nombre de
// host, archivo) que NUNCA deben llegar al cliente.
const secretInternal = "sql: conexión a db-interna falló (repository.go)"

const (
	successBody          = `{"status":"ok","database":"connected"}`
	downBody             = `{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}`
	notFoundBody         = `{"error":{"code":"not_found","message":"Recurso no encontrado"}}`
	methodNotAllowedBody = `{"error":{"code":"method_not_allowed","message":"Método no permitido"}}`
	internalBody         = `{"error":{"code":"internal","message":"Error interno del servidor"}}`
)

// fakeRepo guioniza la salud de la base de datos del dominio status.
type fakeRepo struct{ err error }

func (f fakeRepo) Ping(context.Context) error { return f.err }

// newStack levanta el stack completo con el handler real de status (según
// repoErr) y dos handlers de prueba que provocan un error inesperado y un
// panic. Devuelve el servidor y la captura de logs para SC-009.
func newStack(t *testing.T, repoErr error) (*httptest.Server, *testutil.LogCapture) {
	t.Helper()
	logger, capture := testutil.NewLogger()

	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)

	// Handler real del dominio status: GET /healthz.
	status.RegisterPublic(root, status.NewHandler(status.NewService(fakeRepo{err: repoErr}), logger))

	// Handlers de prueba que provocan las respuestas que /healthz no puede
	// generar por sí solo (plan R12): error inesperado y panic.
	root.Handle(http.MethodGet, "/boom", func(w http.ResponseWriter, r *http.Request) {
		httpserver.WriteError(r.Context(), w, logger, errors.New(secretInternal))
	})
	root.Handle(http.MethodGet, "/panic", func(http.ResponseWriter, *http.Request) {
		panic("pánico de prueba")
	})

	srv := testutil.NewServerWithChain(t, mux, logger, []string{"http://localhost:5173"})
	return srv, capture
}

// TestEnvelopeSuccessIsDirectDTO (SC-008): 200 con el DTO directo, sin wrapper.
func TestEnvelopeSuccessIsDirectDTO(t *testing.T) {
	srv, _ := newStack(t, nil)

	resp := testutil.Do(t, srv, http.MethodGet, "/healthz", "", nil)

	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.Status)
	}
	if got := string(resp.Body); got != successBody {
		t.Errorf("cuerpo = %q, se esperaba %q", got, successBody)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}
}

// TestEnvelopeDatabaseUnavailable (SC-008): 503 con sobre de error y details.
func TestEnvelopeDatabaseUnavailable(t *testing.T) {
	srv, _ := newStack(t, errors.New("dial tcp 127.0.0.1:5432: connection refused"))

	resp := testutil.Do(t, srv, http.MethodGet, "/healthz", "", nil)

	if resp.Status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, se esperaba 503", resp.Status)
	}
	if got := string(resp.Body); got != downBody {
		t.Errorf("cuerpo = %q, se esperaba %q", got, downBody)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}
}

// TestEnvelopeUndocumentedRoute (SC-008): el fallback 404 responde sobre, no el
// texto plano de la stdlib.
func TestEnvelopeUndocumentedRoute(t *testing.T) {
	srv, _ := newStack(t, nil)

	resp := testutil.Do(t, srv, http.MethodGet, "/ruta-que-no-existe", "", nil)

	if resp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404", resp.Status)
	}
	if got := string(resp.Body); got != notFoundBody {
		t.Errorf("cuerpo = %q, se esperaba %q", got, notFoundBody)
	}
	if strings.Contains(strings.ToLower(string(resp.Body)), "page not found") {
		t.Error("la ruta no documentada devolvió el texto plano de la stdlib")
	}
}

// TestEnvelopeMethodNotAllowed (SC-008): método no documentado sobre una ruta
// existente responde el sobre 405.
func TestEnvelopeMethodNotAllowed(t *testing.T) {
	srv, _ := newStack(t, nil)

	resp := testutil.Do(t, srv, http.MethodPost, "/healthz", "", nil)

	if resp.Status != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, se esperaba 405", resp.Status)
	}
	if got := string(resp.Body); got != methodNotAllowedBody {
		t.Errorf("cuerpo = %q, se esperaba %q", got, methodNotAllowedBody)
	}
}

// TestEnvelopeUnexpectedErrorHidesInternals (SC-008, SC-009): 500 genérico sin
// información interna en el cuerpo y con el detalle registrado en el log,
// incluido el request_id que coincide con X-Request-ID.
func TestEnvelopeUnexpectedErrorHidesInternals(t *testing.T) {
	srv, capture := newStack(t, nil)

	resp := testutil.Do(t, srv, http.MethodGet, "/boom", "", nil)

	if resp.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500", resp.Status)
	}
	body := string(resp.Body)
	if body != internalBody {
		t.Errorf("cuerpo = %q, se esperaba %q", body, internalBody)
	}
	lower := strings.ToLower(body)
	for _, leak := range []string{"sql", "db-interna", "repository.go", ".go", "goroutine", "panic", "stack"} {
		if strings.Contains(lower, strings.ToLower(leak)) {
			t.Errorf("la respuesta filtra información interna (%q): %q", leak, body)
		}
	}

	// SC-009: el detalle interno sí queda en el log, con el request_id de la
	// respuesta.
	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		t.Fatal("la respuesta no llevó X-Request-ID")
	}
	record, ok := capture.Find("error interno")
	if !ok {
		t.Fatal("no se registró el error interno en el log")
	}
	if detail, _ := record.Attrs["error"].(string); !strings.Contains(detail, secretInternal) {
		t.Errorf("el log no contiene el detalle interno: %v", record.Attrs)
	}
	if got, _ := record.Attrs["request_id"].(string); got != requestID {
		t.Errorf("request_id del log = %q, se esperaba %q (el de X-Request-ID)", got, requestID)
	}
}

// TestEnvelopePanicIsRecovered (SC-009): un panic produce 500 con sobre y el
// proceso sigue atendiendo peticiones.
func TestEnvelopePanicIsRecovered(t *testing.T) {
	srv, _ := newStack(t, nil)

	resp := testutil.Do(t, srv, http.MethodGet, "/panic", "", nil)

	if resp.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500", resp.Status)
	}
	if got := string(resp.Body); got != internalBody {
		t.Errorf("cuerpo = %q, se esperaba %q", got, internalBody)
	}
	if strings.Contains(string(resp.Body), "pánico") {
		t.Errorf("la respuesta del panic filtra información interna: %q", resp.Body)
	}

	// El proceso sigue vivo: otra petición responde con normalidad.
	alive := testutil.Do(t, srv, http.MethodGet, "/healthz", "", nil)
	if alive.Status != http.StatusOK {
		t.Errorf("tras el panic, /healthz = %d, se esperaba 200 (proceso vivo)", alive.Status)
	}
}
