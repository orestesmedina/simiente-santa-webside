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

	"simiente-santa/backend/internal/agenda"
	"simiente-santa/backend/internal/media"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
	"simiente-santa/backend/internal/portada"
	"simiente-santa/backend/internal/status"
	"simiente-santa/backend/internal/usuarios"
)

// fakeRepository permite guionizar el estado de la conexión sin base de datos.
type fakeRepository struct{ err error }

func (f fakeRepository) Ping(context.Context) error { return f.err }

// stubPortadaRepo satisface portada.Repository para el humo de composición: las
// lecturas públicas devuelven "sin contenido publicado" y el resto (no
// ejercitado en el humo) queda como método no implementado del embedding.
type stubPortadaRepo struct{ portada.Repository }

func (stubPortadaRepo) GetHomeIdentityPublished(context.Context) (portada.Identity, bool, error) {
	return portada.Identity{}, false, nil
}

func (stubPortadaRepo) GetHomeAboutPublished(context.Context) (portada.About, bool, error) {
	return portada.About{}, false, nil
}

func (stubPortadaRepo) GetHomeContactPublished(context.Context) (portada.Contact, bool, error) {
	return portada.Contact{}, false, nil
}

func (stubPortadaRepo) ListHomeServicesPublished(context.Context) ([]portada.Service, error) {
	return nil, nil
}

func (stubPortadaRepo) ListHomeWhatsappChannelsPublished(context.Context) ([]portada.WhatsappChannel, error) {
	return nil, nil
}

func (stubPortadaRepo) ListHomeSocialLinksPublished(context.Context) ([]portada.SocialLink, error) {
	return nil, nil
}

// stubMediaRepo satisface media.Repository para el humo de composición: ningún
// archivo está publicado (la descarga responde 404 salvo por el nombre).
type stubMediaRepo struct{}

func (stubMediaRepo) IsFilePublished(context.Context, string, string, string) (bool, error) {
	return false, nil
}

// stubAgendaRepo satisface agenda.Repository para el humo de composición: sin
// nada publicado y próximo, «lo que viene» responde solo `lang` y los listados
// del panel vienen vacíos. El resto de métodos (no ejercitados en el humo) quedan
// como método no implementado del embedding. Con `actions` captura las filas de
// auditoría best-effort (denegaciones) que le llegan.
type stubAgendaRepo struct {
	agenda.Repository
	actions *[]audit.Action
}

func (s stubAgendaRepo) RecordAction(_ context.Context, action audit.Action) error {
	if s.actions != nil {
		*s.actions = append(*s.actions, action)
	}
	return nil
}

func (stubAgendaRepo) ListUpcomingEvents(context.Context, string, string, int) ([]agenda.Event, error) {
	return nil, nil
}

func (stubAgendaRepo) ListUpcomingActivities(context.Context, string, string, int) ([]agenda.Activity, error) {
	return nil, nil
}

func (stubAgendaRepo) ListEvents(context.Context, *agenda.PublicationState, paginate.Params) ([]agenda.Event, error) {
	return nil, nil
}

func (stubAgendaRepo) CountEvents(context.Context, *agenda.PublicationState) (int64, error) {
	return 0, nil
}

func (stubAgendaRepo) ListActivities(context.Context, *agenda.PublicationState, paginate.Params) ([]agenda.Activity, error) {
	return nil, nil
}

func (stubAgendaRepo) CountActivities(context.Context, *agenda.PublicationState) (int64, error) {
	return 0, nil
}

func (stubAgendaRepo) ListCategories(context.Context) ([]agenda.Category, error) { return nil, nil }

// ListNextUpcomingItems satisface el extracto de la portada del humo: sin ítems
// (la respuesta es solo `lang`).
func (stubAgendaRepo) ListNextUpcomingItems(context.Context, string, string, string, int) ([]agenda.NextItem, error) {
	return nil, nil
}

// liveSessionStore tiene una sesión viva para cualquier token: permite probar
// los permisos de los grupos de panel sin Redis.
type liveSessionStore struct{ userID uuid.UUID }

func (liveSessionStore) Create(context.Context, uuid.UUID) (string, error) { return "", nil }

func (s liveSessionStore) Resolve(context.Context, string) (session.Session, error) {
	now := time.Now()
	return session.Session{
		UserID:            s.userID,
		CreatedAt:         now,
		LastSeenAt:        now,
		AbsoluteExpiresAt: now.Add(time.Hour),
	}, nil
}

func (liveSessionStore) Revoke(context.Context, string) error                      { return nil }
func (liveSessionStore) RevokeUser(context.Context, uuid.UUID) error               { return nil }
func (liveSessionStore) RevokeUserExcept(context.Context, uuid.UUID, string) error { return nil }

// permissionsResolver resuelve la identidad con los permisos indicados.
type permissionsResolver struct{ permissions []string }

func (r permissionsResolver) Resolve(_ context.Context, userID uuid.UUID) (session.Identity, error) {
	return session.Identity{UserID: userID, Permissions: r.permissions}, nil
}

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

// fakeRoleService guioniza la gestión de roles para el humo de composición de
// las rutas /api/v1/admin/roles* y /api/v1/admin/permisos (T237).
type fakeRoleService struct{}

func (fakeRoleService) ListRoles(context.Context, paginate.Params) (usuarios.RoleList, error) {
	return usuarios.RoleList{}, nil
}

func (fakeRoleService) GetRole(context.Context, uuid.UUID) (usuarios.RoleItem, error) {
	return usuarios.RoleItem{}, nil
}

func (fakeRoleService) CreateRole(context.Context, uuid.UUID, usuarios.RoleCreateInput) (usuarios.RoleItem, error) {
	return usuarios.RoleItem{}, nil
}

func (fakeRoleService) UpdateRole(context.Context, uuid.UUID, uuid.UUID, usuarios.RoleUpdateInput) (usuarios.RoleItem, error) {
	return usuarios.RoleItem{}, nil
}

func (fakeRoleService) DeleteRole(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (fakeRoleService) ListPermissions(context.Context) (usuarios.PermissionList, error) {
	return usuarios.PermissionList{}, nil
}

// testOptions guioniza la cadena de sesión y permisos del humo de composición.
type testOptions struct {
	sessions session.Store
	resolver session.Resolver
	// agendaRepo sustituye al stub por defecto (permite capturar la auditoría).
	agendaRepo agenda.Repository
	// agendaRecorder sustituye al service de agenda como Recorder de la cadena
	// (por defecto se usa el propio service, como en producción).
	agendaRecorder audit.Recorder
}

// testDeps compone las dependencias del mux con dobles, sin base de datos ni
// Redis, y publica una ruta de sonda en el grupo de panel para comprobar que la
// cadena aprobada queda montada. Sin sesión viva (todas las peticiones del panel
// responden 401).
func testDeps(repo status.Repository, log *slog.Logger) apiDeps {
	return testDepsWith(testOptions{sessions: fakeSessionStore{}, resolver: fakeResolver{}}, repo, log)
}

// testDepsWith es testDeps con la cadena de sesión/permisos guionizada (para
// probar los 403 de los permisos independientes de F4).
func testDepsWith(opts testOptions, repo status.Repository, log *slog.Logger) apiDeps {
	handler := usuarios.NewHandler(usuarios.HandlerDeps{
		Access:     fakeAuthService{},
		Setup:      fakeSetupService{},
		Roles:      fakeRoleService{},
		SetupToken: "token-de-prueba",
		Logger:     log,
	})
	portadaHandler := portada.NewHandler(portada.HandlerDeps{
		Service:        portada.NewService(portada.ServiceDeps{Repository: stubPortadaRepo{}, Logger: log}),
		MaxUploadBytes: 1 << 20,
		Logger:         log,
	})
	mediaHandler := media.NewHandler(media.HandlerDeps{
		Service: media.NewService(media.ServiceDeps{Repository: stubMediaRepo{}, Logger: log}),
		Logger:  log,
	})
	agendaRepo := opts.agendaRepo
	if agendaRepo == nil {
		agendaRepo = stubAgendaRepo{}
	}
	agendaSvc := agenda.NewService(agenda.ServiceDeps{
		Repository: agendaRepo,
		Location:   time.UTC, // APP_TZ: el humo no depende de la zona del lugar
		Logger:     log,
	})
	agendaRecorder := opts.agendaRecorder
	if agendaRecorder == nil {
		// Como en producción: el service de agenda es el Recorder de su cadena y
		// resuelve método+ruta → acción (T422).
		agendaRecorder = agendaSvc
	}
	agendaHandler := agenda.NewHandler(agenda.HandlerDeps{
		Service:        agendaSvc,
		MaxUploadBytes: 1 << 20,
		Logger:         log,
	})
	return apiDeps{
		statusRepo:  repo,
		authHandler: handler,
		public: usuarios.PublicDeps{
			Login: nil,
			Session: []httpserver.Middleware{
				middleware.Authn(opts.sessions, opts.resolver, log),
				middleware.CSRF("secret", log),
			},
		},
		admin: usuarios.AdminDeps{
			Module: usuarios.PermissionAdminUsersRoles,
			Deps: middleware.AdminDeps{
				Sessions:   opts.sessions,
				Resolver:   opts.resolver,
				CSRFSecret: "secret",
				Logger:     log,
			},
			// El Handler publica las rutas de roles (T237); la sonda comprueba
			// que la cadena aprobada queda montada.
			Handler: handler,
			Routes: func(r httpserver.Registrar) {
				r.Handle(http.MethodGet, "/probe", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
		},
		portadaHandler: portadaHandler,
		portadaAdmin: portada.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions:   opts.sessions,
				Resolver:   opts.resolver,
				CSRFSecret: "secret",
				Logger:     log,
			},
		},
		mediaHandler:  mediaHandler,
		agendaHandler: agendaHandler,
		agendaAdmin: agenda.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions:   opts.sessions,
				Resolver:   opts.resolver,
				Recorder:   agendaRecorder,
				CSRFSecret: "secret",
				Logger:     log,
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

	t.Run("rutas de roles exigen sesión", func(t *testing.T) {
		for _, path := range []string{"/api/v1/admin/roles", "/api/v1/admin/roles/" + uuid.New().String(), "/api/v1/admin/permisos"} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusUnauthorized {
				t.Fatalf("GET %s: status = %d, se esperaba 401 (%s)", path, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
				t.Errorf("GET %s: cuerpo inesperado: %s", path, resp.Body)
			}
		}
	})

	t.Run("ruta inexistente responde 404 con sobre", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/noexiste", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
	})

	t.Run("portada pública responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"lang":"es"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
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

	t.Run("grupo de panel de la portada exige sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/portada", "", nil)
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("status = %d, se esperaba 401 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})
}

// TestAgendaRoutesSmoke comprueba el humo de composición de F4: la superficie
// pública de «lo que viene», los tres grupos de panel con su cadena y el 401/404
// con el sobre del contrato (§8.1.9).
func TestAgendaRoutesSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	srv := httptest.NewServer(httpserver.NewHandler(
		newMux(testDeps(fakeRepository{}, logger), logger),
		logger,
		middleware.Chain(logger, []string{"http://localhost:5173"})...,
	))
	t.Cleanup(srv.Close)

	t.Run("proximamente público responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/proximamente", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang (SC-012)", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("proximamente con lang inválido responde 400", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/proximamente?lang=fr", "", nil)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"invalid"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("los tres grupos de panel exigen sesión", func(t *testing.T) {
		paths := []string{
			"/api/v1/admin/eventos",
			"/api/v1/admin/actividades",
			"/api/v1/admin/categorias",
		}
		for _, path := range paths {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusUnauthorized {
				t.Fatalf("GET %s: status = %d, se esperaba 401 (%s)", path, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
				t.Errorf("GET %s: cuerpo inesperado: %s", path, resp.Body)
			}
		}
	})

	t.Run("las subidas de imágenes exigen sesión", func(t *testing.T) {
		for _, path := range []string{"/api/v1/admin/eventos/imagenes", "/api/v1/admin/actividades/imagenes"} {
			resp := testutil.Do(t, srv, http.MethodPost, path, "", nil)
			if resp.Status != http.StatusUnauthorized {
				t.Fatalf("POST %s: status = %d, se esperaba 401 (%s)", path, resp.Status, resp.Body)
			}
		}
	})

	t.Run("rutas de F4 inexistentes responden 404 con sobre", func(t *testing.T) {
		// Rutas que no coinciden con ningún patrón registrado (§8.1.9); las que
		// sí coinciden con un {id} exigen sesión (401) antes de llegar al handler.
		for _, path := range []string{
			"/api/v1/admin/agenda",
			"/api/v1/admin/eventos/x/y",
			"/api/v1/proximamente/extra",
		} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusNotFound {
				t.Fatalf("GET %s: status = %d, se esperaba 404 (%s)", path, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"not_found"`) {
				t.Errorf("GET %s: cuerpo inesperado: %s", path, resp.Body)
			}
		}
	})
}

// TestAgendaPermissionsSmoke fija SC-005/SC-011: los tres permisos son
// INDEPENDIENTES y se verifican en el servidor (FR-012). Una cuenta con `eventos`
// no entra en el catálogo de categorías ni en actividades, y al revés. El
// Recorder de la cadena es el service de F4 (T422), así que la denegación queda
// registrada con la acción de la ruta (`result='denied'`, R4-10); las lecturas
// del panel no dejan fila (el registro es de mutaciones).
func TestAgendaPermissionsSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	userID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name        string
		permission  string
		allowed     string
		deniedRead  string
		deniedWrite string
		method      string
		wantCode    string
		wantLabel   string
	}{
		{
			name:        "solo eventos",
			permission:  agenda.PermissionEventos,
			allowed:     "/api/v1/admin/eventos",
			deniedRead:  "/api/v1/admin/categorias",
			deniedWrite: "/api/v1/admin/categorias/" + itemID.String(),
			method:      http.MethodPatch,
			wantCode:    audit.ActionCategoryUpdate,
			wantLabel:   "Categorías · " + itemID.String(),
		},
		{
			name:        "solo actividades",
			permission:  agenda.PermissionActividades,
			allowed:     "/api/v1/admin/actividades",
			deniedRead:  "/api/v1/admin/eventos",
			deniedWrite: "/api/v1/admin/eventos/" + itemID.String(),
			method:      http.MethodDelete,
			wantCode:    audit.ActionEventDelete,
			wantLabel:   "Eventos · " + itemID.String(),
		},
		{
			name:        "solo categorias",
			permission:  agenda.PermissionCategorias,
			allowed:     "/api/v1/admin/categorias",
			deniedRead:  "/api/v1/admin/actividades",
			deniedWrite: "/api/v1/admin/actividades/" + itemID.String(),
			method:      http.MethodDelete,
			wantCode:    audit.ActionActivityDelete,
			wantLabel:   "Actividades · " + itemID.String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actions []audit.Action
			deps := testDepsWith(testOptions{
				sessions:   liveSessionStore{userID: userID},
				resolver:   permissionsResolver{permissions: []string{tt.permission}},
				agendaRepo: stubAgendaRepo{actions: &actions},
			}, fakeRepository{}, logger)

			srv := httptest.NewServer(httpserver.NewHandler(
				newMux(deps, logger),
				logger,
				middleware.Chain(logger, []string{"http://localhost:5173"})...,
			))
			t.Cleanup(srv.Close)

			sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

			// Con su permiso, el listado responde 200 (vacío).
			resp := testutil.Do(t, srv, http.MethodGet, tt.allowed, "", sessionCookie)
			if resp.Status != http.StatusOK {
				t.Fatalf("GET %s con permiso %s: status = %d, se esperaba 200 (%s)",
					tt.allowed, tt.permission, resp.Status, resp.Body)
			}

			// Una lectura de otro módulo: 403 genérico y sin fila (el registro es
			// de mutaciones).
			resp = testutil.Do(t, srv, http.MethodGet, tt.deniedRead, "", sessionCookie)
			if resp.Status != http.StatusForbidden {
				t.Fatalf("GET %s sin permiso: status = %d, se esperaba 403 (%s)", tt.deniedRead, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"forbidden"`) {
				t.Errorf("GET %s: cuerpo inesperado: %s", tt.deniedRead, resp.Body)
			}
			if len(actions) != 0 {
				t.Errorf("una lectura denegada no debe dejar fila: %+v", actions)
			}

			// Una mutación de otro módulo: 403 (authz va antes que CSRF) y la
			// denegación queda registrada con la acción de la ruta.
			resp = testutil.Do(t, srv, tt.method, tt.deniedWrite, "", sessionCookie)
			if resp.Status != http.StatusForbidden {
				t.Fatalf("%s %s sin permiso: status = %d, se esperaba 403 (%s)",
					tt.method, tt.deniedWrite, resp.Status, resp.Body)
			}
			if len(actions) != 1 {
				t.Fatalf("denegaciones registradas = %+v, se esperaba 1", actions)
			}
			action := actions[0]
			if action.Code != tt.wantCode || action.Result != audit.ResultDenied {
				t.Errorf("acción = %+v, se esperaba %s con result=denied", action, tt.wantCode)
			}
			if action.TargetLabel != tt.wantLabel {
				t.Errorf("targetLabel = %q, se esperaba %q", action.TargetLabel, tt.wantLabel)
			}
			if action.ActorUserID == nil || *action.ActorUserID != userID {
				t.Errorf("la denegación no lleva el actor de la sesión: %+v", action)
			}
		})
	}
}

// TestPortadaProximamenteSmoke comprueba el humo de composición del extracto del
// bloque «Lo que viene» de la portada (US7/FR-025, T446): la ruta pública
// responde sin sesión, con su sobre y sin caché, y una ruta inexistente del
// segmento sigue dando el 404 con sobre (§8.1.9).
func TestPortadaProximamenteSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	srv := httptest.NewServer(httpserver.NewHandler(
		newMux(testDeps(fakeRepository{}, logger), logger),
		logger,
		middleware.Chain(logger, []string{"http://localhost:5173"})...,
	))
	t.Cleanup(srv.Close)

	t.Run("el extracto responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/proximamente", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		// Sin ítems: la clave `items` se omite (la portada muestra su mensaje).
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("con lang=EN resuelve el idioma", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/proximamente?lang=EN", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"lang":"en"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("lang inválido responde 400", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/proximamente?lang=fr", "", nil)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"invalid"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("la portada de F3 sigue intacta", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"lang":"es"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("subruta inexistente responde 404 con sobre", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/desconocido", "", nil)
		if resp.Status != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"not_found"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})
}
