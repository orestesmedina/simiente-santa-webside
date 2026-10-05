package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
	"simiente-santa/backend/internal/status"
	"simiente-santa/backend/internal/usuarios"
)

// fakeRepository permite guionizar el estado de la conexión sin base de datos.
type fakeRepository struct{ err error }

func (f fakeRepository) Ping(context.Context) error { return f.err }

// fakeAuthService guioniza el servicio de acceso para el humo de composición.
type fakeAuthService struct{}

func (fakeAuthService) Login(_ context.Context, in usuarios.LoginInput, _ string) (usuarios.SessionUser, []*http.Cookie, error) {
	return usuarios.SessionUser{
		ID:          uuid.New().String(),
		Email:       in.Email,
		FirstName:   "Ana",
		LastName:    "Responsable",
		Phone:       "+34 600 000 000",
		RoleID:      uuid.New().String(),
		RoleName:    "Administrador",
		Permissions: []string{usuarios.PermissionAdminUsersRoles},
	}, nil, nil
}

func (fakeAuthService) Logout(context.Context, string) ([]*http.Cookie, error) { return nil, nil }

func (fakeAuthService) ChangeMyPassword(context.Context, session.Identity, string, usuarios.ChangePasswordInput) error {
	return nil
}

// fakeSetupService guioniza la inicialización única para el humo de composición:
// devuelve la cuenta del DTO sin tocar PostgreSQL.
type fakeSetupService struct{}

// setupSmokeBody es un InitializeInput válido para el humo.
const setupSmokeBody = `{"firstName":"Ana","lastName":"Responsable","email":"ana@ejemplo.com","phone":"+34 612 345 678","password":"Semilla.2026"}`

func (fakeSetupService) Initialize(_ context.Context, in usuarios.InitializeInput) (usuarios.UserItem, error) {
	return usuarios.UserItem{
		ID:        uuid.New().String(),
		Email:     in.Email,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     in.Phone,
		RoleID:    uuid.New().String(),
		RoleName:  "Administrador",
		IsActive:  true,
		CreatedAt: time.Now(),
	}, nil
}

// fakeSessionStore no tiene sesiones: cualquier token es inexistente. Basta para
// el humo de composición (peticiones sin sesión).
type fakeSessionStore struct{}

func (fakeSessionStore) Create(context.Context, uuid.UUID) (string, error) { return "", nil }
func (fakeSessionStore) Resolve(context.Context, string) (session.Session, error) {
	return session.Session{}, session.ErrSessionNotFound
}
func (fakeSessionStore) Revoke(context.Context, string) error                      { return nil }
func (fakeSessionStore) RevokeUser(context.Context, uuid.UUID) error               { return nil }
func (fakeSessionStore) RevokeUserExcept(context.Context, uuid.UUID, string) error { return nil }

// fakeResolver implementa session.Resolver; nunca se llega a él sin cookie.
type fakeResolver struct{}

func (fakeResolver) Resolve(context.Context, uuid.UUID) (session.Identity, error) {
	return session.Identity{}, errors.New("sin identidad")
}

// testDeps compone las dependencias del mux con dobles, sin base de datos ni
// Redis, y publica una ruta de sonda en el grupo de panel para comprobar que la
// cadena aprobada queda montada.
func testDeps(repo status.Repository, log *slog.Logger) apiDeps {
	return apiDeps{
		statusRepo: repo,
		authHandler: usuarios.NewHandler(usuarios.HandlerDeps{
			Access:     fakeAuthService{},
			Setup:      fakeSetupService{},
			SetupToken: "token-de-prueba",
			Logger:     log,
		}),
		public: usuarios.PublicDeps{
			Login: nil,
			Session: []httpserver.Middleware{
				middleware.Authn(fakeSessionStore{}, fakeResolver{}, log),
				middleware.CSRF("secret", log),
			},
		},
		admin: usuarios.AdminDeps{
			Module: usuarios.PermissionAdminUsersRoles,
			Deps: middleware.AdminDeps{
				Sessions:   fakeSessionStore{},
				Resolver:   fakeResolver{},
				CSRFSecret: "secret",
				Logger:     log,
			},
			Routes: func(r httpserver.Registrar) {
				r.Handle(http.MethodGet, "/probe", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
		},
	}
}

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
				newMux(testDeps(fakeRepository{err: tt.repoErr}, logger), logger),
				logger,
				middleware.Chain(logger, []string{"http://localhost:5173"})...,
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

// TestNewRoutesSmoke comprueba que el mux publica la superficie de F2 con su
// sobre: el login responde el DTO directo, las rutas con sesión y el grupo de
// panel exigen sesión (401 unauthenticated) y una ruta inexistente sigue dando
// 404 con sobre de error.
func TestNewRoutesSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	srv := httptest.NewServer(httpserver.NewHandler(
		newMux(testDeps(fakeRepository{}, logger), logger),
		logger,
		middleware.Chain(logger, []string{"http://localhost:5173"})...,
	))
	t.Cleanup(srv.Close)

	t.Run("login público responde el DTO directo", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/auth/login",
			`{"email":"ana@ejemplo.com","password":"Semilla.2026"}`,
			map[string]string{"Content-Type": "application/json"})
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"email":"ana@ejemplo.com"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
		if strings.Contains(string(resp.Body), "password") {
			t.Errorf("la respuesta filtra credenciales: %s", resp.Body)
		}
	})

	t.Run("sesión sin cookie responde 401", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/auth/session", "", nil)
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("setup sin token responde 401", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", setupSmokeBody,
			map[string]string{"Content-Type": "application/json"})
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("setup con token responde 201", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodPost, "/api/v1/setup/initialize", setupSmokeBody,
			map[string]string{"Content-Type": "application/json", "X-Setup-Token": "token-de-prueba"})
		if resp.Status != http.StatusCreated {
			t.Fatalf("status = %d, se esperaba 201 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"roleName":"Administrador"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
		if strings.Contains(string(resp.Body), "password") {
			t.Errorf("la respuesta filtra credenciales: %s", resp.Body)
		}
	})

	t.Run("grupo de panel exige sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/probe", "", nil)
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("ruta inexistente responde 404 con sobre", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/noexiste", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
	})
}
