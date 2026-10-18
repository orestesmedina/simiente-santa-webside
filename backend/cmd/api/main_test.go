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
	"simiente-santa/backend/internal/donaciones"
	"simiente-santa/backend/internal/media"
	"simiente-santa/backend/internal/medios"
	"simiente-santa/backend/internal/ministerios"
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

// stubMinisteriosRepo satisface ministerios.Repository para el humo de
// composición: sin nada publicado la sección y el bloque responden solo `lang`, y
// los listados del panel vienen vacíos. El resto de métodos (no ejercitados en el
// humo) quedan como método no implementado del embedding. Con `actions` captura
// las filas de auditoría best-effort (denegaciones) que le llegan.
type stubMinisteriosRepo struct {
	ministerios.Repository
	actions *[]audit.Action
}

func (s stubMinisteriosRepo) RecordAction(_ context.Context, action audit.Action) error {
	if s.actions != nil {
		*s.actions = append(*s.actions, action)
	}
	return nil
}

func (stubMinisteriosRepo) ListPublishedMinistries(context.Context, int) ([]ministerios.Ministry, error) {
	return nil, nil
}

func (stubMinisteriosRepo) ListMinistries(context.Context, *ministerios.PublicationState, paginate.Params) ([]ministerios.MinistryListItem, error) {
	return nil, nil
}

func (stubMinisteriosRepo) CountMinistries(context.Context, *ministerios.PublicationState) (int64, error) {
	return 0, nil
}

// stubDonacionesRepo satisface donaciones.Repository para el humo de composición:
// sin nada publicado la sección y el bloque responden solo `lang`, y los listados
// del panel vienen vacíos. El resto de métodos (no ejercitados en el humo) quedan
// como método no implementado del embedding. Con `actions` captura las filas de
// auditoría best-effort (denegaciones) que le llegan.
type stubDonacionesRepo struct {
	donaciones.Repository
	actions *[]audit.Action
}

func (s stubDonacionesRepo) RecordAction(_ context.Context, action audit.Action) error {
	if s.actions != nil {
		*s.actions = append(*s.actions, action)
	}
	return nil
}

func (stubDonacionesRepo) ListPublishedDonationMethods(context.Context, int) ([]donaciones.MetodoDonacion, error) {
	return nil, nil
}

func (stubDonacionesRepo) ListPublishedDonationTexts(context.Context) ([]donaciones.TextoDonacion, error) {
	return nil, nil
}

func (stubDonacionesRepo) ListDonationMethods(context.Context, *donaciones.PublicationState, paginate.Params) ([]donaciones.MetodoDonacion, error) {
	return nil, nil
}

func (stubDonacionesRepo) CountDonationMethods(context.Context, *donaciones.PublicationState) (int64, error) {
	return 0, nil
}

func (stubDonacionesRepo) ListDonationTexts(context.Context) ([]donaciones.TextoDonacion, error) {
	return nil, nil
}

// stubMediosRepo satisface medios.Repository para el humo de composición: sin
// nada publicable la sección responde con `items` vacío y el bloque solo `lang`, y
// los listados del panel vienen vacíos. El resto de métodos (no ejercitados en el
// humo) quedan como método no implementado del embedding. Con `actions` captura
// las filas de auditoría best-effort (denegaciones) que le llegan.
type stubMediosRepo struct {
	medios.Repository
	actions *[]audit.Action
}

func (s stubMediosRepo) RecordAction(_ context.Context, action audit.Action) error {
	if s.actions != nil {
		*s.actions = append(*s.actions, action)
	}
	return nil
}

func (stubMediosRepo) ListPublishedMediaItems(context.Context, paginate.Params) ([]medios.Episodio, error) {
	return nil, nil
}

func (stubMediosRepo) CountPublishedMediaItems(context.Context) (int64, error) { return 0, nil }

func (stubMediosRepo) ListPortadaMediaItems(context.Context, int) ([]medios.Episodio, error) {
	return nil, nil
}

func (stubMediosRepo) ListMediaItems(context.Context, medios.FiltroPanel, paginate.Params) ([]medios.Episodio, error) {
	return nil, nil
}

func (stubMediosRepo) CountMediaItems(context.Context, medios.FiltroPanel) (int64, error) {
	return 0, nil
}

func (stubMediosRepo) GetLastMediaSyncRun(context.Context) (medios.SyncRun, bool, error) {
	return medios.SyncRun{}, false, nil
}

func (stubMediosRepo) ListMediaEpisodeSourcesByProvider(context.Context, medios.Proveedor) ([]medios.FuenteEpisodio, error) {
	return nil, nil
}

// InsertMediaSyncRun cierra la búsqueda del humo (sin red real: los dos orígenes
// quedan `failed`) y captura la fila de auditoría `media.sync` que la acompaña.
func (s stubMediosRepo) InsertMediaSyncRun(_ context.Context, run medios.SyncRun, actions ...audit.Action) (medios.SyncRun, error) {
	run.ID = uuid.New()
	if s.actions != nil {
		*s.actions = append(*s.actions, actions...)
	}
	return run, nil
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
	// ministeriosRepo sustituye al stub por defecto (permite capturar la auditoría).
	ministeriosRepo ministerios.Repository
	// donacionesRepo sustituye al stub por defecto (permite capturar la auditoría).
	donacionesRepo donaciones.Repository
	// mediosRepo sustituye al stub por defecto (permite capturar la auditoría).
	mediosRepo medios.Repository
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
	ministeriosRepo := opts.ministeriosRepo
	if ministeriosRepo == nil {
		ministeriosRepo = stubMinisteriosRepo{}
	}
	// Como en producción: el service de ministerios es el Recorder de su cadena y
	// resuelve método+ruta → acción (T616).
	ministeriosSvc := ministerios.NewService(ministerios.ServiceDeps{
		Repository: ministeriosRepo,
		Logger:     log,
	})
	ministeriosHandler := ministerios.NewHandler(ministerios.HandlerDeps{
		Service:        ministeriosSvc,
		MaxUploadBytes: 1 << 20,
		Logger:         log,
	})
	donacionesRepo := opts.donacionesRepo
	if donacionesRepo == nil {
		donacionesRepo = stubDonacionesRepo{}
	}
	// Como en producción: el service de donaciones es el Recorder de su cadena y
	// resuelve método+ruta → acción (T716).
	donacionesSvc := donaciones.NewService(donaciones.ServiceDeps{
		Repository: donacionesRepo,
		Logger:     log,
	})
	donacionesHandler := donaciones.NewHandler(donaciones.HandlerDeps{
		Service: donacionesSvc,
		Logger:  log,
	})
	mediosRepo := opts.mediosRepo
	if mediosRepo == nil {
		mediosRepo = stubMediosRepo{}
	}
	// Como en producción: el service de medios es el Recorder de su cadena y
	// resuelve método+ruta → acción (T922). La red saliente real no se construye
	// en el humo: la búsqueda no se ejercita sin sesión y los feeds reales no
	// deben tocarse en las pruebas.
	mediosSvc := medios.NewService(medios.ServiceDeps{
		Repository: mediosRepo,
		Location:   time.UTC, // APP_TZ: el humo no depende de la zona del lugar
		Logger:     log,
	})
	mediosHandler := medios.NewHandler(medios.HandlerDeps{
		Service: mediosSvc,
		Logger:  log,
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
		ministeriosHandler: ministeriosHandler,
		ministeriosAdmin: ministerios.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions:   opts.sessions,
				Resolver:   opts.resolver,
				Recorder:   ministeriosSvc,
				CSRFSecret: "secret",
				Logger:     log,
			},
		},
		donacionesHandler: donacionesHandler,
		donacionesAdmin: donaciones.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions:   opts.sessions,
				Resolver:   opts.resolver,
				Recorder:   donacionesSvc,
				CSRFSecret: "secret",
				Logger:     log,
			},
		},
		mediosHandler: mediosHandler,
		mediosAdmin: medios.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions:   opts.sessions,
				Resolver:   opts.resolver,
				Recorder:   mediosSvc,
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

// TestMinisteriosRoutesSmoke comprueba el humo de composición de F6 (T621): la
// superficie pública (sección y bloque de la portada) responde sin sesión y sin
// caché, el grupo de panel exige sesión, `/healthz` sigue intacto (§8.1.9) y una
// ruta inexistente responde 404 con sobre.
func TestMinisteriosRoutesSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	srv := httptest.NewServer(httpserver.NewHandler(
		newMux(testDeps(fakeRepository{}, logger), logger),
		logger,
		middleware.Chain(logger, []string{"http://localhost:5173"})...,
	))
	t.Cleanup(srv.Close)

	t.Run("la sección pública responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/ministerios", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang (items omitida, SC-010)", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("el bloque de la portada responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/ministerios", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("lang inválido responde 400", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/ministerios?lang=fr", "", nil)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"invalid"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("el grupo de panel exige sesión", func(t *testing.T) {
		for _, probe := range []struct {
			method string
			path   string
		}{
			{method: http.MethodGet, path: "/api/v1/admin/ministerios"},
			{method: http.MethodPost, path: "/api/v1/admin/ministerios"},
			{method: http.MethodGet, path: "/api/v1/admin/ministerios/" + uuid.New().String()},
			{method: http.MethodPatch, path: "/api/v1/admin/ministerios/" + uuid.New().String()},
			{method: http.MethodDelete, path: "/api/v1/admin/ministerios/" + uuid.New().String()},
			{method: http.MethodPost, path: "/api/v1/admin/ministerios/imagenes"},
		} {
			resp := testutil.Do(t, srv, probe.method, probe.path, "", nil)
			if resp.Status != http.StatusUnauthorized {
				t.Fatalf("%s %s: status = %d, se esperaba 401 (%s)", probe.method, probe.path, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
				t.Errorf("%s %s: cuerpo inesperado: %s", probe.method, probe.path, resp.Body)
			}
		}
	})

	t.Run("healthz sigue intacto y las rutas inexistentes dan 404 con sobre", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/healthz", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("/healthz: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		for _, path := range []string{
			"/api/v1/admin/ministerios/x/y",
			"/api/v1/ministerios/extra",
			"/api/v1/portada/ministerios/extra",
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

	t.Run("la portada de F3 y la agenda siguen intactas", func(t *testing.T) {
		for _, path := range []string{"/api/v1/portada", "/api/v1/proximamente"} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusOK {
				t.Fatalf("GET %s: status = %d, se esperaba 200 (%s)", path, resp.Status, resp.Body)
			}
		}
	})
}

// TestMinisteriosPermissionsSmoke fija SC-005/SC-011 (FR-011): el permiso
// `ministerios` es independiente de los demás y se verifica **en el servidor**.
// Una cuenta con `ministerios` no entra en los grupos de F4 y al revés; las
// denegaciones quedan registradas con la acción de la ruta (`result='denied'`,
// R6-12) y las lecturas del panel no dejan fila.
func TestMinisteriosPermissionsSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	userID := uuid.New()
	itemID := uuid.New()

	t.Run("con ministerios", func(t *testing.T) {
		var actions []audit.Action
		deps := testDepsWith(testOptions{
			sessions: liveSessionStore{userID: userID},
			resolver: permissionsResolver{permissions: []string{ministerios.PermissionMinisterios}},
			// La denegación de una mutación de OTRO módulo la registra el Recorder
			// de ESA cadena (el service de F4, T422): aquí se captura desde el
			// repositorio de agenda para poder comprobarla.
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
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/ministerios", "", sessionCookie)
		if resp.Status != http.StatusOK {
			t.Fatalf("GET con permiso: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if len(actions) != 0 {
			t.Errorf("una lectura del panel no debe dejar fila: %+v", actions)
		}

		// Otro módulo (F4): 403 genérico y sin fila (la lectura no se audita).
		resp = testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/eventos", "", sessionCookie)
		if resp.Status != http.StatusForbidden {
			t.Fatalf("GET eventos: status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"forbidden"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
		if len(actions) != 0 {
			t.Errorf("una lectura denegada no debe dejar fila: %+v", actions)
		}

		// Una mutación de otro módulo: 403 y la denegación queda registrada con la
		// acción de ESA ruta (el Recorder de la cadena es el service de F6).
		resp = testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/eventos/"+itemID.String(), "", sessionCookie)
		if resp.Status != http.StatusForbidden {
			t.Fatalf("PATCH eventos: status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if len(actions) != 1 {
			t.Fatalf("denegaciones = %+v, se esperaba 1", actions)
		}
		if actions[0].Code != "event.update" || actions[0].Result != audit.ResultDenied {
			t.Errorf("acción = %+v, se esperaba event.update con result=denied", actions[0])
		}
		if actions[0].ActorUserID == nil || *actions[0].ActorUserID != userID {
			t.Errorf("la denegación no lleva el actor de la sesión: %+v", actions[0])
		}
	})

	t.Run("sin ministerios", func(t *testing.T) {
		for _, permission := range []string{
			agenda.PermissionEventos,
			agenda.PermissionActividades,
			agenda.PermissionCategorias,
		} {
			t.Run(permission, func(t *testing.T) {
				var actions []audit.Action
				deps := testDepsWith(testOptions{
					sessions:        liveSessionStore{userID: userID},
					resolver:        permissionsResolver{permissions: []string{permission}},
					ministeriosRepo: stubMinisteriosRepo{actions: &actions},
				}, fakeRepository{}, logger)

				srv := httptest.NewServer(httpserver.NewHandler(
					newMux(deps, logger),
					logger,
					middleware.Chain(logger, []string{"http://localhost:5173"})...,
				))
				t.Cleanup(srv.Close)
				sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

				resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/ministerios", "", sessionCookie)
				if resp.Status != http.StatusForbidden {
					t.Fatalf("GET ministerios con %s: status = %d, se esperaba 403 (%s)", permission, resp.Status, resp.Body)
				}
				if !strings.Contains(string(resp.Body), `"code":"forbidden"`) {
					t.Errorf("cuerpo inesperado: %s", resp.Body)
				}

				resp = testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/ministerios/"+itemID.String(), "", sessionCookie)
				if resp.Status != http.StatusForbidden {
					t.Fatalf("PATCH ministerios con %s: status = %d, se esperaba 403 (%s)", permission, resp.Status, resp.Body)
				}
				if len(actions) != 1 {
					t.Fatalf("denegaciones = %+v, se esperaba 1", actions)
				}
				action := actions[0]
				if action.Code != audit.ActionMinistryUpdate || action.Result != audit.ResultDenied {
					t.Errorf("acción = %+v, se esperaba ministry.update con result=denied", action)
				}
				if want := "Ministerios · " + itemID.String(); action.TargetLabel != want {
					t.Errorf("targetLabel = %q, se esperaba %q", action.TargetLabel, want)
				}
			})
		}
	})

	t.Run("y al revés: su propio módulo sigue funcionando", func(t *testing.T) {
		deps := testDepsWith(testOptions{
			sessions: liveSessionStore{userID: userID},
			resolver: permissionsResolver{permissions: []string{agenda.PermissionEventos}},
		}, fakeRepository{}, logger)

		srv := httptest.NewServer(httpserver.NewHandler(
			newMux(deps, logger),
			logger,
			middleware.Chain(logger, []string{"http://localhost:5173"})...,
		))
		t.Cleanup(srv.Close)
		sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/eventos", "", sessionCookie)
		if resp.Status != http.StatusOK {
			t.Fatalf("GET eventos con su permiso: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
	})
}

// TestDonacionesRoutesSmoke comprueba el humo de composición de F7 (T721): la
// superficie pública (sección de ofrendas y bloque de la portada) responde sin
// sesión y sin caché, el grupo de panel exige sesión, `/healthz` sigue intacto
// (§8.1.9) y una ruta inexistente responde 404 con sobre.
func TestDonacionesRoutesSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	srv := httptest.NewServer(httpserver.NewHandler(
		newMux(testDeps(fakeRepository{}, logger), logger),
		logger,
		middleware.Chain(logger, []string{"http://localhost:5173"})...,
	))
	t.Cleanup(srv.Close)

	t.Run("la sección pública responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/ofrendas", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		// Sin nada publicado: las claves de contenido se omiten (SC-010) y es el
		// frontend el que redirige (P7-21).
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("el bloque de la portada responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/ofrendas", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("lang inválido responde 400", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/ofrendas?lang=fr", "", nil)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"invalid"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
	})

	t.Run("el grupo de panel exige sesión", func(t *testing.T) {
		for _, probe := range []struct {
			method string
			path   string
		}{
			{method: http.MethodGet, path: "/api/v1/admin/donaciones"},
			{method: http.MethodPost, path: "/api/v1/admin/donaciones"},
			{method: http.MethodGet, path: "/api/v1/admin/donaciones/" + uuid.New().String()},
			{method: http.MethodPatch, path: "/api/v1/admin/donaciones/" + uuid.New().String()},
			{method: http.MethodDelete, path: "/api/v1/admin/donaciones/" + uuid.New().String()},
			{method: http.MethodGet, path: "/api/v1/admin/donaciones/textos"},
			{method: http.MethodPut, path: "/api/v1/admin/donaciones/textos/intro"},
		} {
			resp := testutil.Do(t, srv, probe.method, probe.path, "", nil)
			if resp.Status != http.StatusUnauthorized {
				t.Fatalf("%s %s: status = %d, se esperaba 401 (%s)", probe.method, probe.path, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
				t.Errorf("%s %s: cuerpo inesperado: %s", probe.method, probe.path, resp.Body)
			}
		}
	})

	t.Run("healthz sigue intacto y las rutas inexistentes dan 404 con sobre", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/healthz", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("/healthz: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		// Las rutas que coinciden con un {id} exigen sesión (401) antes de llegar
		// al handler; estas no coinciden con ningún patrón registrado (§8.1.9).
		for _, path := range []string{
			"/api/v1/admin/donaciones/x/y",
			"/api/v1/ofrendas/extra",
			"/api/v1/portada/ofrendas/extra",
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

	t.Run("la portada de F3 y los demás módulos siguen intactos", func(t *testing.T) {
		for _, path := range []string{"/api/v1/portada", "/api/v1/proximamente", "/api/v1/ministerios"} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusOK {
				t.Fatalf("GET %s: status = %d, se esperaba 200 (%s)", path, resp.Status, resp.Body)
			}
		}
	})
}

// TestDonacionesPermissionsSmoke fija SC-005/SC-011 (FR-011): el permiso
// `donaciones` es independiente de los demás y se verifica **en el servidor**. Una
// cuenta con `donaciones` no entra en los grupos de F4/F6 y al revés; las
// denegaciones quedan registradas con la acción de la ruta (`result='denied'`,
// R7-10) y las lecturas del panel no dejan fila.
func TestDonacionesPermissionsSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	userID := uuid.New()
	itemID := uuid.New()

	t.Run("con donaciones", func(t *testing.T) {
		var agendaActions []audit.Action
		deps := testDepsWith(testOptions{
			sessions: liveSessionStore{userID: userID},
			resolver: permissionsResolver{permissions: []string{donaciones.PermissionDonaciones}},
			// La denegación de una mutación de OTRO módulo la registra el Recorder
			// de ESA cadena (el service de F4): aquí se captura desde el
			// repositorio de agenda para poder comprobarla.
			agendaRepo: stubAgendaRepo{actions: &agendaActions},
		}, fakeRepository{}, logger)

		srv := httptest.NewServer(httpserver.NewHandler(
			newMux(deps, logger),
			logger,
			middleware.Chain(logger, []string{"http://localhost:5173"})...,
		))
		t.Cleanup(srv.Close)
		sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

		// Con su permiso, el listado responde 200 (vacío).
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/donaciones", "", sessionCookie)
		if resp.Status != http.StatusOK {
			t.Fatalf("GET con permiso: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if len(agendaActions) != 0 {
			t.Errorf("una lectura del panel no debe dejar fila: %+v", agendaActions)
		}

		// Otro módulo (F4): 403 genérico y sin fila (la lectura no se audita).
		resp = testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/eventos", "", sessionCookie)
		if resp.Status != http.StatusForbidden {
			t.Fatalf("GET eventos: status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"forbidden"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
		if len(agendaActions) != 0 {
			t.Errorf("una lectura denegada no debe dejar fila: %+v", agendaActions)
		}

		// Una mutación de otro módulo: 403 y la denegación queda registrada con la
		// acción de ESA ruta (el Recorder de la cadena es el service de F4).
		resp = testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/eventos/"+itemID.String(), "", sessionCookie)
		if resp.Status != http.StatusForbidden {
			t.Fatalf("PATCH eventos: status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if len(agendaActions) != 1 {
			t.Fatalf("denegaciones = %+v, se esperaba 1", agendaActions)
		}
		if agendaActions[0].Code != "event.update" || agendaActions[0].Result != audit.ResultDenied {
			t.Errorf("acción = %+v, se esperaba event.update con result=denied", agendaActions[0])
		}
	})

	t.Run("sin donaciones", func(t *testing.T) {
		for _, permission := range []string{
			agenda.PermissionEventos,
			agenda.PermissionActividades,
			agenda.PermissionCategorias,
			ministerios.PermissionMinisterios,
		} {
			t.Run(permission, func(t *testing.T) {
				var actions []audit.Action
				deps := testDepsWith(testOptions{
					sessions:       liveSessionStore{userID: userID},
					resolver:       permissionsResolver{permissions: []string{permission}},
					donacionesRepo: stubDonacionesRepo{actions: &actions},
				}, fakeRepository{}, logger)

				srv := httptest.NewServer(httpserver.NewHandler(
					newMux(deps, logger),
					logger,
					middleware.Chain(logger, []string{"http://localhost:5173"})...,
				))
				t.Cleanup(srv.Close)
				sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

				resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/donaciones", "", sessionCookie)
				if resp.Status != http.StatusForbidden {
					t.Fatalf("GET donaciones con %s: status = %d, se esperaba 403 (%s)", permission, resp.Status, resp.Body)
				}
				if !strings.Contains(string(resp.Body), `"code":"forbidden"`) {
					t.Errorf("cuerpo inesperado: %s", resp.Body)
				}

				resp = testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/donaciones/"+itemID.String(), "", sessionCookie)
				if resp.Status != http.StatusForbidden {
					t.Fatalf("PATCH donaciones con %s: status = %d, se esperaba 403 (%s)", permission, resp.Status, resp.Body)
				}
				if len(actions) != 1 {
					t.Fatalf("denegaciones = %+v, se esperaba 1", actions)
				}
				action := actions[0]
				if action.Code != audit.ActionDonationUpdate || action.Result != audit.ResultDenied {
					t.Errorf("acción = %+v, se esperaba donation.update con result=denied", action)
				}
				if want := "Donaciones · " + itemID.String(); action.TargetLabel != want {
					t.Errorf("targetLabel = %q, se esperaba %q", action.TargetLabel, want)
				}
			})
		}
	})

	t.Run("y al revés: su propio módulo sigue funcionando", func(t *testing.T) {
		deps := testDepsWith(testOptions{
			sessions: liveSessionStore{userID: userID},
			resolver: permissionsResolver{permissions: []string{agenda.PermissionEventos}},
		}, fakeRepository{}, logger)

		srv := httptest.NewServer(httpserver.NewHandler(
			newMux(deps, logger),
			logger,
			middleware.Chain(logger, []string{"http://localhost:5173"})...,
		))
		t.Cleanup(srv.Close)
		sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/eventos", "", sessionCookie)
		if resp.Status != http.StatusOK {
			t.Fatalf("GET eventos con su permiso: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
	})
}

// TestMediosRoutesSmoke comprueba el humo de composición de F9 (T928): la
// superficie pública (sección y bloque de la portada) responde sin sesión y sin
// caché, el grupo de panel exige sesión en sus cinco operaciones, `/healthz` sigue
// intacto (§8.1.9) y una ruta inexistente responde 404 con sobre.
func TestMediosRoutesSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	srv := httptest.NewServer(httpserver.NewHandler(
		newMux(testDeps(fakeRepository{}, logger), logger),
		logger,
		middleware.Chain(logger, []string{"http://localhost:5173"})...,
	))
	t.Cleanup(srv.Close)

	t.Run("la sección pública responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/medios", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		// Sin nada publicable: `items` vacío y el sobre completo (la redirección
		// de la SPA es del frontend, P9-16).
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es","items":[],"total":0,"limit":12,"offset":0}` {
			t.Errorf("cuerpo = %s", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("el bloque de la portada responde 200 sin sesión", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/portada/medios", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if body := strings.TrimSpace(string(resp.Body)); body != `{"lang":"es"}` {
			t.Errorf("cuerpo = %s, se esperaba solo lang (items omitida)", body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, se esperaba no-store", got)
		}
	})

	t.Run("lang inválido responde 400", func(t *testing.T) {
		for _, path := range []string{"/api/v1/medios?lang=fr", "/api/v1/portada/medios?lang=fr"} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusBadRequest {
				t.Fatalf("GET %s: status = %d, se esperaba 400 (%s)", path, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"invalid"`) {
				t.Errorf("GET %s: cuerpo inesperado: %s", path, resp.Body)
			}
		}
	})

	t.Run("el grupo de panel exige sesión", func(t *testing.T) {
		for _, probe := range []struct {
			method string
			path   string
		}{
			{method: http.MethodGet, path: "/api/v1/admin/medios"},
			{method: http.MethodPost, path: "/api/v1/admin/medios/buscar-nuevos"},
			{method: http.MethodPatch, path: "/api/v1/admin/medios/" + uuid.New().String()},
			{method: http.MethodPost, path: "/api/v1/admin/medios/" + uuid.New().String() + "/unir"},
			{method: http.MethodPost, path: "/api/v1/admin/medios/" + uuid.New().String() + "/marcas/resolver"},
		} {
			resp := testutil.Do(t, srv, probe.method, probe.path, "", nil)
			if resp.Status != http.StatusUnauthorized {
				t.Fatalf("%s %s: status = %d, se esperaba 401 (%s)", probe.method, probe.path, resp.Status, resp.Body)
			}
			if !strings.Contains(string(resp.Body), `"code":"unauthenticated"`) {
				t.Errorf("%s %s: cuerpo inesperado: %s", probe.method, probe.path, resp.Body)
			}
		}
	})

	t.Run("healthz sigue intacto y las rutas inexistentes dan 404 con sobre", func(t *testing.T) {
		resp := testutil.Do(t, srv, http.MethodGet, "/healthz", "", nil)
		if resp.Status != http.StatusOK {
			t.Fatalf("/healthz: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		for _, path := range []string{
			"/api/v1/admin/medios/x/y",
			"/api/v1/medios/extra",
			"/api/v1/portada/medios/extra",
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

	t.Run("la portada de F3 y los demás módulos siguen intactos", func(t *testing.T) {
		for _, path := range []string{
			"/api/v1/portada",
			"/api/v1/proximamente",
			"/api/v1/ministerios",
			"/api/v1/ofrendas",
			"/api/v1/portada/ofrendas",
		} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", nil)
			if resp.Status != http.StatusOK {
				t.Fatalf("GET %s: status = %d, se esperaba 200 (%s)", path, resp.Status, resp.Body)
			}
		}
	})
}

// TestMediosPermissionsSmoke fija SC-009/FR-021 (T928): el permiso `medios` es
// independiente de los demás y se verifica **en el servidor** a través de la
// cadena `AdminChain("medios")`. Una cuenta con `medios` no entra en los grupos de
// F4/F6/F7 y al revés; las denegaciones quedan registradas con la acción de la ruta
// (`result='denied'`, T922) y las lecturas del panel no dejan fila.
func TestMediosPermissionsSmoke(t *testing.T) {
	logger, _ := testutil.NewLogger()
	userID := uuid.New()
	itemID := uuid.New()

	t.Run("con medios", func(t *testing.T) {
		var accionesMedios []audit.Action
		var accionesAjenas []audit.Action
		deps := testDepsWith(testOptions{
			sessions:   liveSessionStore{userID: userID},
			resolver:   permissionsResolver{permissions: []string{medios.PermissionMedios}},
			mediosRepo: stubMediosRepo{actions: &accionesMedios},
			// La denegación de una mutación de OTRO módulo la registra el Recorder
			// de ESA cadena (el service de F4): aquí se captura desde el
			// repositorio de agenda para poder comprobarla.
			agendaRepo: stubAgendaRepo{actions: &accionesAjenas},
		}, fakeRepository{}, logger)

		srv := httptest.NewServer(httpserver.NewHandler(
			newMux(deps, logger),
			logger,
			middleware.Chain(logger, []string{"http://localhost:5173"})...,
		))
		t.Cleanup(srv.Close)
		sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

		// Con su permiso, el listado responde 200 (vacío) y una lectura no deja
		// fila (el registro es de mutaciones).
		resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/medios", "", sessionCookie)
		if resp.Status != http.StatusOK {
			t.Fatalf("GET con permiso: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if len(accionesMedios) != 0 {
			t.Errorf("una lectura del panel no debe dejar fila: %+v", accionesMedios)
		}

		// La búsqueda con permiso, sesión y CSRF responde 200 con el resumen. El
		// humo no construye la red saliente real, así que los dos orígenes quedan
		// `failed`: es exactamente FR-016 (no es un error y nada se oculta).
		csrf, err := session.NewCSRFToken("secret")
		if err != nil {
			t.Fatalf("generar el token CSRF: %v", err)
		}
		mutacion := map[string]string{
			"Cookie":       session.CookieSession + "=token-de-prueba; " + session.CookieCSRF + "=" + csrf,
			"X-CSRF-Token": csrf,
			"Content-Type": "application/json",
		}
		resp = testutil.Do(t, srv, http.MethodPost, "/api/v1/admin/medios/buscar-nuevos", "", mutacion)
		if resp.Status != http.StatusOK {
			t.Fatalf("POST buscar-nuevos: status = %d, se esperaba 200 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"youtube":"failed"`) {
			t.Errorf("el resumen debe informar del feed caído: %s", resp.Body)
		}
		if len(accionesMedios) != 1 {
			t.Fatalf("auditoría = %+v, se esperaba una fila media.sync", accionesMedios)
		}
		if accionesMedios[0].Code != audit.ActionMediaSync || accionesMedios[0].Result != audit.ResultSuccess {
			t.Errorf("acción = %+v, se esperaba media.sync con result=success", accionesMedios[0])
		}
		if accionesMedios[0].ActorUserID == nil || *accionesMedios[0].ActorUserID != userID {
			t.Errorf("la búsqueda no lleva el actor de la sesión: %+v", accionesMedios[0])
		}

		// Otro módulo (F4): 403 genérico y sin fila (la lectura no se audita).
		resp = testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/eventos", "", sessionCookie)
		if resp.Status != http.StatusForbidden {
			t.Fatalf("GET eventos: status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if !strings.Contains(string(resp.Body), `"code":"forbidden"`) {
			t.Errorf("cuerpo inesperado: %s", resp.Body)
		}
		if len(accionesAjenas) != 0 {
			t.Errorf("una lectura denegada no debe dejar fila: %+v", accionesAjenas)
		}

		// Una mutación de otro módulo: 403 y la denegación queda registrada con la
		// acción de ESA ruta (el Recorder de la cadena es el service de F4).
		resp = testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/eventos/"+itemID.String(), "", sessionCookie)
		if resp.Status != http.StatusForbidden {
			t.Fatalf("PATCH eventos: status = %d, se esperaba 403 (%s)", resp.Status, resp.Body)
		}
		if len(accionesAjenas) != 1 || accionesAjenas[0].Code != "event.update" ||
			accionesAjenas[0].Result != audit.ResultDenied {
			t.Fatalf("denegaciones ajenas = %+v, se esperaba event.update con result=denied", accionesAjenas)
		}
	})

	t.Run("sin medios", func(t *testing.T) {
		for _, permission := range []string{
			agenda.PermissionEventos,
			agenda.PermissionActividades,
			agenda.PermissionCategorias,
			ministerios.PermissionMinisterios,
			donaciones.PermissionDonaciones,
		} {
			t.Run(permission, func(t *testing.T) {
				var acciones []audit.Action
				deps := testDepsWith(testOptions{
					sessions:   liveSessionStore{userID: userID},
					resolver:   permissionsResolver{permissions: []string{permission}},
					mediosRepo: stubMediosRepo{actions: &acciones},
				}, fakeRepository{}, logger)

				srv := httptest.NewServer(httpserver.NewHandler(
					newMux(deps, logger),
					logger,
					middleware.Chain(logger, []string{"http://localhost:5173"})...,
				))
				t.Cleanup(srv.Close)
				sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

				resp := testutil.Do(t, srv, http.MethodGet, "/api/v1/admin/medios", "", sessionCookie)
				if resp.Status != http.StatusForbidden {
					t.Fatalf("GET medios con %s: status = %d, se esperaba 403 (%s)", permission, resp.Status, resp.Body)
				}
				if !strings.Contains(string(resp.Body), `"code":"forbidden"`) {
					t.Errorf("cuerpo inesperado: %s", resp.Body)
				}

				// Una mutación denegada: 403 (authz va antes que CSRF) y la
				// denegación queda registrada con la acción de la ruta.
				resp = testutil.Do(t, srv, http.MethodPatch, "/api/v1/admin/medios/"+itemID.String(), "", sessionCookie)
				if resp.Status != http.StatusForbidden {
					t.Fatalf("PATCH medios con %s: status = %d, se esperaba 403 (%s)", permission, resp.Status, resp.Body)
				}
				if len(acciones) != 1 {
					t.Fatalf("denegaciones = %+v, se esperaba 1", acciones)
				}
				action := acciones[0]
				if action.Code != audit.ActionMediaDescriptionUpdate || action.Result != audit.ResultDenied {
					t.Errorf("acción = %+v, se esperaba media.description.update con result=denied", action)
				}
				if want := "Medios · " + itemID.String(); action.TargetLabel != want {
					t.Errorf("targetLabel = %q, se esperaba %q", action.TargetLabel, want)
				}
				if action.ActorUserID == nil || *action.ActorUserID != userID {
					t.Errorf("la denegación no lleva el actor de la sesión: %+v", action)
				}
			})
		}
	})

	t.Run("y al revés: una cuenta solo de medios no entra en los demás módulos", func(t *testing.T) {
		deps := testDepsWith(testOptions{
			sessions: liveSessionStore{userID: userID},
			resolver: permissionsResolver{permissions: []string{medios.PermissionMedios}},
		}, fakeRepository{}, logger)

		srv := httptest.NewServer(httpserver.NewHandler(
			newMux(deps, logger),
			logger,
			middleware.Chain(logger, []string{"http://localhost:5173"})...,
		))
		t.Cleanup(srv.Close)
		sessionCookie := map[string]string{"Cookie": session.CookieSession + "=token-de-prueba"}

		for _, path := range []string{
			"/api/v1/admin/eventos",
			"/api/v1/admin/ministerios",
			"/api/v1/admin/donaciones",
		} {
			resp := testutil.Do(t, srv, http.MethodGet, path, "", sessionCookie)
			if resp.Status != http.StatusForbidden {
				t.Fatalf("GET %s con solo medios: status = %d, se esperaba 403 (%s)", path, resp.Status, resp.Body)
			}
		}
	})
}
