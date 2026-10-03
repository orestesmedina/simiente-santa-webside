package status

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/testutil"
)

// fakeService guioniza la respuesta del service para probar el handler aislado.
type fakeService struct {
	status SystemStatus
	err    error
}

func (f *fakeService) Status(context.Context) (SystemStatus, error) {
	return f.status, f.err
}

func TestGetSystemStatusSuccess(t *testing.T) {
	logger, _ := testutil.NewLogger()
	h := NewHandler(&fakeService{status: SystemStatus{Status: StatusOK, Database: DatabaseConnected}}, logger)

	rec := httptest.NewRecorder()
	h.GetSystemStatus(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", rec.Code)
	}
	// Sobre de éxito = DTO directo, sin wrapper {"data":…}.
	if got := rec.Body.String(); got != `{"status":"ok","database":"connected"}` {
		t.Errorf("cuerpo = %q, se esperaba el DTO directo del contrato", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, se esperaba application/json", ct)
	}
}

func TestGetSystemStatusDatabaseUnavailable(t *testing.T) {
	logger, _ := testutil.NewLogger()
	svcErr := apperr.DatabaseUnavailable(
		MessageDatabaseUnavailable,
		apperr.WithDetails(map[string]any{DetailDatabaseKey: DatabaseDisconnected}),
	)
	h := NewHandler(&fakeService{err: svcErr}, logger)

	rec := httptest.NewRecorder()
	h.GetSystemStatus(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, se esperaba 503", rec.Code)
	}
	const want = `{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}`
	if got := rec.Body.String(); got != want {
		t.Errorf("cuerpo = %q, se esperaba el sobre de error del contrato", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, se esperaba application/json", ct)
	}
}

func TestGetSystemStatusUnexpectedError(t *testing.T) {
	const secret = "sql: conexión a db-interna falló"
	logger, capture := testutil.NewLogger()
	h := NewHandler(&fakeService{err: errors.New(secret)}, logger)

	rec := httptest.NewRecorder()
	h.GetSystemStatus(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500", rec.Code)
	}
	const want = `{"error":{"code":"internal","message":"Error interno del servidor"}}`
	if got := rec.Body.String(); got != want {
		t.Errorf("cuerpo = %q, se esperaba el sobre internal genérico", got)
	}
	if body := rec.Body.String(); strings.Contains(body, "sql") || strings.Contains(body, "db-interna") {
		t.Errorf("la respuesta filtra información interna: %q", body)
	}

	// El detalle interno sí queda en el log estructurado para diagnóstico.
	record, ok := capture.Find("error interno")
	if !ok {
		t.Fatal("no se registró el error interno en el log")
	}
	if detail, _ := record.Attrs["error"].(string); !strings.Contains(detail, secret) {
		t.Errorf("el log no contiene el detalle interno: %v", record.Attrs)
	}
}
