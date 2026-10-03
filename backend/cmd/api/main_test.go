package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/testutil"
)

// fakeRepository permite guionizar el estado de la conexión sin base de datos.
type fakeRepository struct{ err error }

func (f fakeRepository) Ping(context.Context) error { return f.err }

// TestHealthzSmoke comprueba de punta a punta el mux ensamblado con la cadena de
// middlewares: el sobre de respuesta (200 o 503 según el repositorio) y que la
// trazabilidad (X-Request-ID) está montada.
func TestHealthzSmoke(t *testing.T) {
	const (
		okBody   = `{"status":"ok","database":"connected"}`
		downBody = `{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}`
	)

	tests := []struct {
		name       string
		repoErr    error
		wantStatus int
		wantBody   string
	}{
		{name: "base de datos conectada", repoErr: nil, wantStatus: http.StatusOK, wantBody: okBody},
		{
			name:       "base de datos no conectada",
			repoErr:    errors.New("dial tcp 127.0.0.1:5432: connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   downBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := testutil.NewLogger()
			srv := httptest.NewServer(httpserver.NewHandler(
				newMux(fakeRepository{err: tt.repoErr}, logger),
				logger,
				middlewareChain(logger, []string{"http://localhost:5173"})...,
			))
			t.Cleanup(srv.Close)

			resp := testutil.Do(t, srv, http.MethodGet, "/healthz", "", nil)

			if resp.Status != tt.wantStatus {
				t.Errorf("status = %d, se esperaba %d", resp.Status, tt.wantStatus)
			}
			if got := string(resp.Body); got != tt.wantBody {
				t.Errorf("cuerpo = %q, se esperaba %q", got, tt.wantBody)
			}
			if resp.Header.Get("X-Request-ID") == "" {
				t.Error("la cadena de middlewares no está montada: falta X-Request-ID")
			}
			if got := resp.Header.Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, se esperaba no-store", got)
			}
		})
	}
}
