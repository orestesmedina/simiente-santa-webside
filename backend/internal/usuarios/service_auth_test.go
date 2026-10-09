package usuarios

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/password"
	"simiente-santa/backend/internal/platform/session"
)

// Tests unitarios de service_auth.go con fakes de repositorio, Store de
// sesiones, contadores FR-006 y registro de accesos. No tocan PostgreSQL ni
// Redis. Las contraseñas de prueba se hashean con bcrypt de coste mínimo para
// que la suite sea rápida (la producción usa cost 12, que es lo que valida la
// verificación real).

// --- Fakes ---

type fakeAuthRepo struct {
	mu sync.Mutex

	authByEmail map[string]UserAuth
	users       map[uuid.UUID]User
	roles       map[uuid.UUID]Role

	authErr   error
	userErr   error
	roleErr   error
	changeErr error
	recordErr error

	seenEmails         []string
	recordSuccessCalls []loginSuccessCall
	updatedHashes      map[uuid.UUID]string
	mustUpdates        map[uuid.UUID]bool
	changeCalls        int
	adminActions       int
}

type loginSuccessCall struct {
	userID uuid.UUID
	ip     string
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		authByEmail:   map[string]UserAuth{},
		users:         map[uuid.UUID]User{},
		roles:         map[uuid.UUID]Role{},
		updatedHashes: map[uuid.UUID]string{},
		mustUpdates:   map[uuid.UUID]bool{},
	}
}

func (f *fakeAuthRepo) GetUserAuthByEmail(_ context.Context, email string) (UserAuth, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seenEmails = append(f.seenEmails, email)
	if f.authErr != nil {
		return UserAuth{}, f.authErr
	}
	auth, ok := f.authByEmail[email]
	if !ok {
		return UserAuth{}, apperr.NotFound("La cuenta no existe")
	}
	return auth, nil
}

func (f *fakeAuthRepo) GetUserByID(_ context.Context, id uuid.UUID) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.userErr != nil {
		return User{}, f.userErr
	}
	user, ok := f.users[id]
	if !ok {
		return User{}, apperr.NotFound("La cuenta no existe")
	}
	return user, nil
}

func (f *fakeAuthRepo) GetRoleByID(_ context.Context, id uuid.UUID) (Role, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.roleErr != nil {
		return Role{}, f.roleErr
	}
	role, ok := f.roles[id]
	if !ok {
		return Role{}, apperr.NotFound("El rol no existe")
	}
	return role, nil
}

func (f *fakeAuthRepo) ChangeUserPassword(_ context.Context, id uuid.UUID, hash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.changeErr != nil {
		return f.changeErr
	}
	f.changeCalls++
	f.updatedHashes[id] = hash
	f.mustUpdates[id] = false
	return nil
}

func (f *fakeAuthRepo) RecordLoginSuccess(_ context.Context, userID uuid.UUID, ip string) (LoginEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.recordErr != nil {
		return LoginEvent{}, f.recordErr
	}
	f.recordSuccessCalls = append(f.recordSuccessCalls, loginSuccessCall{userID: userID, ip: ip})
	return LoginEvent{UserID: &userID, Result: audit.ResultSuccess, IP: ip}, nil
}

type revokeExceptCall struct {
	userID     uuid.UUID
	keepTokens []string
}

type fakeSessionStore struct {
	mu sync.Mutex

	createdToken string
	createErr    error
	created      []uuid.UUID

	revoked   []string
	revokeErr error

	revokeUserExceptCalls []revokeExceptCall
	revokeUserExceptErr   error
}

func (f *fakeSessionStore) Create(_ context.Context, userID uuid.UUID) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.createErr != nil {
		return "", f.createErr
	}
	f.created = append(f.created, userID)
	return f.createdToken, nil
}

func (f *fakeSessionStore) Resolve(_ context.Context, _ string) (session.Session, error) {
	return session.Session{}, session.ErrSessionNotFound
}

func (f *fakeSessionStore) Revoke(_ context.Context, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.revokeErr != nil {
		return f.revokeErr
	}
	f.revoked = append(f.revoked, token)
	return nil
}

func (f *fakeSessionStore) RevokeUser(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeSessionStore) RevokeUserExcept(_ context.Context, userID uuid.UUID, keepToken string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.revokeUserExceptErr != nil {
		return f.revokeUserExceptErr
	}
	f.revokeUserExceptCalls = append(f.revokeUserExceptCalls, revokeExceptCall{userID: userID, keepTokens: []string{keepToken}})
	return nil
}

type fakeThrottle struct {
	mu sync.Mutex

	maxAttempts int64
	blocked     bool
	retryAfter  time.Duration
	failures    map[string]int64
	resets      map[string]int
	blockedErr  error
	registerErr error
	resetErr    error
}

func newFakeThrottle() *fakeThrottle {
	return &fakeThrottle{
		maxAttempts: MaxFailedAttempts,
		failures:    map[string]int64{},
		resets:      map[string]int{},
	}
}

func (f *fakeThrottle) Blocked(_ context.Context, _ string) (bool, time.Duration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.blockedErr != nil {
		return false, 0, f.blockedErr
	}
	return f.blocked, f.retryAfter, nil
}

func (f *fakeThrottle) RegisterFailure(_ context.Context, identifier string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.registerErr != nil {
		return 0, f.registerErr
	}
	f.failures[identifier]++
	if f.failures[identifier] >= f.maxAttempts {
		f.blocked = true
		if f.retryAfter <= 0 {
			f.retryAfter = LockoutDuration
		}
	}
	return f.failures[identifier], nil
}

func (f *fakeThrottle) Reset(_ context.Context, identifier string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.resetErr != nil {
		return f.resetErr
	}
	delete(f.failures, identifier)
	f.blocked = false
	f.retryAfter = 0
	f.resets[identifier]++
	return nil
}

type fakeLoginRecorder struct {
	mu     sync.Mutex
	events []audit.Event
}

func (f *fakeLoginRecorder) RecordLoginEventBestEffort(_ context.Context, event audit.Event) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
}

func (f *fakeLoginRecorder) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

// --- Helpers ---

func hashFast(t *testing.T, value string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash de prueba: %v", err)
	}
	return string(hash)
}

func newTestAuthService(repo *fakeAuthRepo, store *fakeSessionStore, throttle *fakeThrottle, recorder *fakeLoginRecorder) *authService {
	return NewAuthService(AuthServiceDeps{
		Repository:  repo,
		Sessions:    store,
		Throttle:    throttle,
		LoginEvents: recorder,
		Cookies:     session.NewCookieConfig(false, time.Hour),
		CSRFSecret:  "secreto-de-prueba",
	})
}

func activeAuth(t *testing.T, email, plainPassword string) UserAuth {
	t.Helper()
	return UserAuth{
		ID:                 uuid.New(),
		Email:              email,
		FirstName:          "Ana",
		LastName:           "Pérez",
		PasswordHash:       hashFast(t, plainPassword),
		MustChangePassword: false,
		IsActive:           true,
		RoleID:             uuid.New(),
		RoleName:           "Administrador",
		Permissions:        []string{"admin_usuarios_roles"},
	}
}

func userFromAuth(auth UserAuth, phone string, active bool) User {
	return User{
		ID:                 auth.ID,
		Email:              auth.Email,
		FirstName:          auth.FirstName,
		LastName:           auth.LastName,
		Phone:              phone,
		MustChangePassword: auth.MustChangePassword,
		IsActive:           active,
		RoleID:             auth.RoleID,
		RoleName:           auth.RoleName,
	}
}

func requireKind(t *testing.T, err error, kind apperr.Kind) *apperr.Error {
	t.Helper()
	if err == nil {
		t.Fatalf("se esperaba un error de tipo %s, se recibió nil", kind)
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba *apperr.Error, se recibió %T: %v", err, err)
	}
	if domainErr.Kind != kind {
		t.Fatalf("kind = %s (%d), se esperaba %s", domainErr.Kind, domainErr.Kind, kind)
	}
	return domainErr
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// --- Login ---

func TestLoginSuccess(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.authByEmail["ana@ejemplo.com"] = auth
	repo.users[auth.ID] = userFromAuth(auth, "+34 600 000 000", true)

	store := &fakeSessionStore{createdToken: "token-de-sesion"}
	throttle := newFakeThrottle()
	recorder := &fakeLoginRecorder{}
	service := newTestAuthService(repo, store, throttle, recorder)

	got, cookies, err := service.Login(context.Background(),
		LoginInput{Email: "  Ana@Ejemplo.COM  ", Password: "Corriente1!"}, "10.0.0.1")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// Correo normalizado (Q5): el repositorio recibe la forma canónica.
	if len(repo.seenEmails) != 1 || repo.seenEmails[0] != "ana@ejemplo.com" {
		t.Fatalf("correo consultado = %v, se esperaba la forma normalizada", repo.seenEmails)
	}
	if got.Email != "ana@ejemplo.com" || got.Phone != "+34 600 000 000" {
		t.Fatalf("DTO de sesión = %+v", got)
	}
	if len(got.Permissions) != 1 || got.Permissions[0] != "admin_usuarios_roles" {
		t.Fatalf("permisos = %v", got.Permissions)
	}
	if got.MustChangePassword {
		t.Fatal("mustChangePassword no debería estar activo")
	}

	if len(store.created) != 1 || store.created[0] != auth.ID {
		t.Fatalf("sesiones creadas = %v", store.created)
	}
	if len(repo.recordSuccessCalls) != 1 || repo.recordSuccessCalls[0].ip != "10.0.0.1" {
		t.Fatalf("registro de éxito = %+v", repo.recordSuccessCalls)
	}
	if throttle.resets["ana@ejemplo.com"] != 1 {
		t.Fatalf("el contador no se limpió: %v", throttle.resets)
	}
	if recorder.count() != 0 {
		t.Fatalf("un login correcto no deja intento fallido: %d", recorder.count())
	}

	if len(cookies) != 2 {
		t.Fatalf("cookies = %d, se esperaban 2", len(cookies))
	}
	sessionCookie := findCookie(cookies, session.CookieSession)
	if sessionCookie == nil || sessionCookie.Value != "token-de-sesion" || !sessionCookie.HttpOnly {
		t.Fatalf("cookie de sesión = %+v", sessionCookie)
	}
	if csrf := findCookie(cookies, session.CookieCSRF); csrf == nil || csrf.Value == "" {
		t.Fatalf("cookie CSRF = %+v", csrf)
	}
}

func TestLoginGenericFailureIsIdenticalForUnknownAndWrongPassword(t *testing.T) {
	t.Run("correo inexistente", func(t *testing.T) {
		repo := newFakeAuthRepo()
		throttle := newFakeThrottle()
		recorder := &fakeLoginRecorder{}
		service := newTestAuthService(repo, &fakeSessionStore{}, throttle, recorder)

		_, _, err := service.Login(context.Background(),
			LoginInput{Email: "nadie@ejemplo.com", Password: "LoQueSea1!"}, "10.0.0.9")
		domainErr := requireKind(t, err, apperr.KindUnauthenticated)
		if domainErr.Message != loginFailureMessage {
			t.Fatalf("mensaje = %q, se esperaba el genérico", domainErr.Message)
		}
		if domainErr.Code() != "unauthenticated" {
			t.Fatalf("code = %q", domainErr.Code())
		}
		if recorder.count() != 1 || recorder.events[0].UserID != nil {
			t.Fatalf("el intento sin cuenta debe registrarse sin asociación: %+v", recorder.events)
		}
		if recorder.events[0].IP != "10.0.0.9" || recorder.events[0].Result != audit.ResultFailure {
			t.Fatalf("evento = %+v", recorder.events[0])
		}
	})

	t.Run("contraseña incorrecta", func(t *testing.T) {
		repo := newFakeAuthRepo()
		auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
		repo.authByEmail["ana@ejemplo.com"] = auth
		recorder := &fakeLoginRecorder{}
		service := newTestAuthService(repo, &fakeSessionStore{}, newFakeThrottle(), recorder)

		_, _, err := service.Login(context.Background(),
			LoginInput{Email: "ana@ejemplo.com", Password: "OtraCosa9#"}, "10.0.0.9")
		domainErr := requireKind(t, err, apperr.KindUnauthenticated)
		if domainErr.Message != loginFailureMessage {
			t.Fatalf("mensaje = %q, se esperaba el genérico", domainErr.Message)
		}
		// El correo inexistente y la contraseña errónea comparten mensaje y
		// código (FR-003/SC-008).
		if recorder.count() != 1 || recorder.events[0].UserID == nil || *recorder.events[0].UserID != auth.ID {
			t.Fatalf("el intento debe asociarse a la cuenta: %+v", recorder.events)
		}
	})
}

func TestLoginInactiveAccountForbidden(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	auth.IsActive = false
	auth.MustChangePassword = true
	repo.authByEmail["ana@ejemplo.com"] = auth

	throttle := newFakeThrottle()
	recorder := &fakeLoginRecorder{}
	service := newTestAuthService(repo, &fakeSessionStore{}, throttle, recorder)

	_, _, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Corriente1!"}, "10.0.0.2")
	domainErr := requireKind(t, err, apperr.KindForbidden)
	if domainErr.Message != accountInactiveMessage {
		t.Fatalf("mensaje = %q", domainErr.Message)
	}
	// El contrato exige `details.reason = "access_disabled"` para que la UI
	// muestre el aviso dedicado en lugar del error genérico (US1 esc. 3).
	if domainErr.Details["reason"] != accessDisabledReason {
		t.Fatalf("details.reason = %v, se esperaba %q", domainErr.Details["reason"], accessDisabledReason)
	}
	if recorder.count() != 1 || recorder.events[0].UserID == nil {
		t.Fatalf("la cuenta inactiva debe dejar su intento: %+v", recorder.events)
	}
	if throttle.failures["ana@ejemplo.com"] != 1 {
		t.Fatalf("el intento inactivo es un fallo de FR-006: %v", throttle.failures)
	}
}

func TestLoginLockoutFifthRespondsGenericAndSixthRateLimited(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.authByEmail["ana@ejemplo.com"] = auth
	throttle := newFakeThrottle()
	recorder := &fakeLoginRecorder{}
	service := newTestAuthService(repo, &fakeSessionStore{}, throttle, recorder)

	// Los cinco primeros fallos responden el 401 genérico (FR-006).
	for attempt := 1; attempt <= MaxFailedAttempts; attempt++ {
		_, _, err := service.Login(context.Background(),
			LoginInput{Email: "ana@ejemplo.com", Password: "Mala1!x"}, "10.0.0.3")
		domainErr := requireKind(t, err, apperr.KindUnauthenticated)
		if domainErr.Message != loginFailureMessage {
			t.Fatalf("fallo %d: mensaje = %q", attempt, domainErr.Message)
		}
	}
	if !throttle.blocked {
		t.Fatal("el 5.º fallo debe crear el bloqueo")
	}

	// El 6.º intento ya responde 429 con Retry-After, sin verificar la cuenta.
	_, _, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Mala1!x"}, "10.0.0.3")
	domainErr := requireKind(t, err, apperr.KindRateLimited)
	if domainErr.Message != lockoutMessage {
		t.Fatalf("mensaje de bloqueo = %q", domainErr.Message)
	}
	if domainErr.RetryAfter() != int(LockoutDuration.Seconds()) {
		t.Fatalf("Retry-After = %d, se esperaba %d", domainErr.RetryAfter(), int(LockoutDuration.Seconds()))
	}
	if domainErr.Details[apperr.RetryAfterDetail] != int(LockoutDuration.Seconds()) {
		t.Fatalf("details = %+v", domainErr.Details)
	}
	if recorder.count() != MaxFailedAttempts+1 {
		t.Fatalf("intentos registrados = %d", recorder.count())
	}

	// Pasados los 15 min se permite de nuevo: con la contraseña correcta entra.
	throttle.blocked = false
	throttle.retryAfter = 0
	repo.users[auth.ID] = userFromAuth(auth, "+34 600 000 000", true)
	if _, _, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Corriente1!"}, "10.0.0.3"); err != nil {
		t.Fatalf("pasado el bloqueo debe permitir entrar: %v", err)
	}
}

func TestLoginSuccessClearsCounter(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.authByEmail["ana@ejemplo.com"] = auth
	repo.users[auth.ID] = userFromAuth(auth, "+34 600 000 000", true)
	throttle := newFakeThrottle()
	service := newTestAuthService(repo, &fakeSessionStore{createdToken: "t"}, throttle, &fakeLoginRecorder{})

	if _, _, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Mala1!x"}, "10.0.0.4"); err == nil {
		t.Fatal("el primer intento debía fallar")
	}
	if throttle.failures["ana@ejemplo.com"] != 1 {
		t.Fatalf("fallos = %v", throttle.failures)
	}
	if _, _, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Corriente1!"}, "10.0.0.4"); err != nil {
		t.Fatalf("el intento correcto debía entrar: %v", err)
	}
	if throttle.resets["ana@ejemplo.com"] != 1 || throttle.failures["ana@ejemplo.com"] != 0 {
		t.Fatalf("el contador no se limpió: resets=%v fallos=%v", throttle.resets, throttle.failures)
	}
}

func TestLoginFailsClosedWhenRegisterFails(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.authByEmail["ana@ejemplo.com"] = auth
	repo.users[auth.ID] = userFromAuth(auth, "+34 600 000 000", true)
	repo.recordErr = errors.New("base de datos caída")

	store := &fakeSessionStore{createdToken: "t"}
	service := newTestAuthService(repo, store, newFakeThrottle(), &fakeLoginRecorder{})

	_, cookies, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Corriente1!"}, "10.0.0.5")
	if err == nil {
		t.Fatal("sin registro no debe haber acceso (fail-closed)")
	}
	if len(store.created) != 0 || cookies != nil {
		t.Fatalf("no debe crearse sesión ni cookies: %v / %v", store.created, cookies)
	}
}

func TestLoginBestEffortRegistrationNeverChangesResponse(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.authByEmail["ana@ejemplo.com"] = auth

	// Un fallo del contador (Redis) no convierte el 401 en 500.
	throttle := newFakeThrottle()
	throttle.registerErr = errors.New("redis caído")
	service := newTestAuthService(repo, &fakeSessionStore{}, throttle, &fakeLoginRecorder{})
	_, _, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Mala1!x"}, "10.0.0.6")
	_ = requireKind(t, err, apperr.KindUnauthenticated)

	// Un fallo al limpiar el contador tampoco impide el acceso correcto.
	successRepo := newFakeAuthRepo()
	successRepo.authByEmail["ana@ejemplo.com"] = auth
	successRepo.users[auth.ID] = userFromAuth(auth, "+34 600 000 000", true)
	resetThrottle := newFakeThrottle()
	resetThrottle.resetErr = errors.New("redis caído")
	successService := newTestAuthService(successRepo, &fakeSessionStore{createdToken: "t"}, resetThrottle, &fakeLoginRecorder{})
	if _, _, err := successService.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Corriente1!"}, "10.0.0.6"); err != nil {
		t.Fatalf("un fallo best-effort no debe impedir el acceso: %v", err)
	}
}

func TestLoginFailsClosedWhenThrottleUnavailable(t *testing.T) {
	repo := newFakeAuthRepo()
	throttle := newFakeThrottle()
	throttle.blockedErr = errors.New("redis caído")
	service := newTestAuthService(repo, &fakeSessionStore{}, throttle, &fakeLoginRecorder{})

	// Sin poder comprobar el bloqueo no se autentica (fail-closed).
	if _, _, err := service.Login(context.Background(),
		LoginInput{Email: "ana@ejemplo.com", Password: "Corriente1!"}, "10.0.0.7"); err == nil {
		t.Fatal("un Redis caído no debe permitir el acceso")
	}
}

func TestLoginResponseNeverContainsPassword(t *testing.T) {
	payload, err := json.Marshal(SessionUser{MustChangePassword: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := strings.ToLower(string(payload))
	if strings.Contains(body, "passwordhash") || strings.Contains(body, `"password"`) {
		t.Fatalf("el DTO de sesión expone credenciales: %s", payload)
	}
	// Las cookies de sesión solo llevan el token opaco y el CSRF firmado.
	cookies := []*http.Cookie{
		session.NewSessionCookie(session.CookieConfig{}, "token"),
		session.NewCSRFCookie(session.CookieConfig{}, "nonce.firma"),
	}
	for _, cookie := range cookies {
		if strings.Contains(cookie.Value, "Corriente1!") {
			t.Fatalf("la cookie %s parece contener una credencial", cookie.Name)
		}
	}
}

// --- Logout ---

func TestLogoutRevokesSessionAndClearsCookies(t *testing.T) {
	store := &fakeSessionStore{}
	service := newTestAuthService(newFakeAuthRepo(), store, newFakeThrottle(), &fakeLoginRecorder{})

	cookies, err := service.Logout(context.Background(), "token-actual")
	if err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if len(store.revoked) != 1 || store.revoked[0] != "token-actual" {
		t.Fatalf("sesión revocada = %v", store.revoked)
	}
	if len(cookies) != 2 {
		t.Fatalf("cookies = %d", len(cookies))
	}
	for _, name := range []string{session.CookieSession, session.CookieCSRF} {
		cookie := findCookie(cookies, name)
		if cookie == nil || cookie.MaxAge >= 0 {
			t.Fatalf("la cookie %s no se borra: %+v", name, cookie)
		}
	}
}

// --- Resolve ---

func TestResolveReturnsFreshIdentityPerRequest(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.users[auth.ID] = userFromAuth(auth, "+34 600 000 000", true)
	role := Role{ID: auth.RoleID, Name: "Administrador", Permissions: []string{"admin_usuarios_roles"}}
	repo.roles[auth.RoleID] = role

	service := newTestAuthService(repo, &fakeSessionStore{}, newFakeThrottle(), &fakeLoginRecorder{})

	identity, err := service.Resolve(context.Background(), auth.ID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if identity.UserID != auth.ID || identity.Email != "ana@ejemplo.com" || identity.RoleName != "Administrador" {
		t.Fatalf("identidad = %+v", identity)
	}
	if !identity.HasPermission("admin_usuarios_roles") {
		t.Fatalf("permisos = %v", identity.Permissions)
	}

	// Un cambio de permisos del rol se refleja en la siguiente petición
	// (FR-018/SC-009).
	role.Permissions = []string{"eventos"}
	repo.roles[auth.RoleID] = role
	identity, err = service.Resolve(context.Background(), auth.ID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if identity.HasPermission("admin_usuarios_roles") || !identity.HasPermission("eventos") {
		t.Fatalf("los permisos no se refrescaron: %v", identity.Permissions)
	}
}

func TestResolveInactiveAccountIsRejected(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.users[auth.ID] = userFromAuth(auth, "+34 600 000 000", false)

	service := newTestAuthService(repo, &fakeSessionStore{}, newFakeThrottle(), &fakeLoginRecorder{})
	_, err := service.Resolve(context.Background(), auth.ID)
	if !errors.Is(err, ErrAccountInactive) {
		t.Fatalf("err = %v, se esperaba ErrAccountInactive", err)
	}
}

func TestResolveUnknownAccount(t *testing.T) {
	service := newTestAuthService(newFakeAuthRepo(), &fakeSessionStore{}, newFakeThrottle(), &fakeLoginRecorder{})
	_, err := service.Resolve(context.Background(), uuid.New())
	_ = requireKind(t, err, apperr.KindNotFound)
}

// --- ChangeMyPassword ---

func changeIdentity(auth UserAuth) session.Identity {
	return session.Identity{
		UserID:    auth.ID,
		Email:     auth.Email,
		FirstName: auth.FirstName,
		LastName:  auth.LastName,
	}
}

func TestChangeMyPasswordSuccess(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	auth.MustChangePassword = true
	repo.authByEmail["ana@ejemplo.com"] = auth
	store := &fakeSessionStore{}
	service := newTestAuthService(repo, store, newFakeThrottle(), &fakeLoginRecorder{})

	err := service.ChangeMyPassword(context.Background(), changeIdentity(auth), "token-actual",
		ChangePasswordInput{CurrentPassword: "Corriente1!", NewPassword: "Nueva9#Aa"})
	if err != nil {
		t.Fatalf("ChangeMyPassword: %v", err)
	}

	hash, ok := repo.updatedHashes[auth.ID]
	if !ok {
		t.Fatal("no se guardó el hash nuevo")
	}
	if err := password.Verify(hash, "Nueva9#Aa"); err != nil {
		t.Fatalf("el hash no corresponde a la contraseña nueva: %v", err)
	}
	if must, ok := repo.mustUpdates[auth.ID]; !ok || must {
		t.Fatalf("mustChangePassword debe quedar en false: %v", repo.mustUpdates)
	}
	// El hash y la obligación se resuelven en una sola operación (atómica).
	if repo.changeCalls != 1 {
		t.Fatalf("cambios de contraseña = %d, se esperaba 1", repo.changeCalls)
	}
	if len(store.revokeUserExceptCalls) != 1 ||
		store.revokeUserExceptCalls[0].userID != auth.ID ||
		store.revokeUserExceptCalls[0].keepTokens[0] != "token-actual" {
		t.Fatalf("revocación de las demás sesiones = %+v", store.revokeUserExceptCalls)
	}
	// El cambio propio NO es una acción administrativa (FR-023).
	if repo.adminActions != 0 {
		t.Fatalf("admin_actions escritas = %d, se esperaban 0", repo.adminActions)
	}
}

func TestChangeMyPasswordWrongCurrentPassword(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.authByEmail["ana@ejemplo.com"] = auth
	store := &fakeSessionStore{}
	service := newTestAuthService(repo, store, newFakeThrottle(), &fakeLoginRecorder{})

	err := service.ChangeMyPassword(context.Background(), changeIdentity(auth), "token",
		ChangePasswordInput{CurrentPassword: "NoEsLa1!", NewPassword: "Nueva9#Aa"})
	domainErr := requireKind(t, err, apperr.KindInvalid)
	if _, ok := domainErr.Details[currentPasswordDetail]; !ok {
		t.Fatalf("details = %+v, se esperaba el campo de la contraseña actual", domainErr.Details)
	}
	if len(repo.updatedHashes) != 0 || len(store.revokeUserExceptCalls) != 0 {
		t.Fatal("una contraseña actual incorrecta no debe cambiar nada")
	}
}

func TestChangeMyPasswordPolicyViolations(t *testing.T) {
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")

	cases := []struct {
		name string
		pass string
	}{
		{"igual al nombre", "Ana"},
		{"igual a los apellidos", "Pérez"},
		{"igual al correo", "ana@ejemplo.com"},
		{"demasiado corta", "Ab1!"},
		{"demasiado larga", strings.Repeat("Aa1!", 20)},
		{"sin mayúscula", "minusculas1!"},
		{"sin minúscula", "MAYUSCULAS1!"},
		{"sin número", "SinNumeros!"},
		{"sin especial", "SinEspecial1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeAuthRepo()
			repo.authByEmail["ana@ejemplo.com"] = auth
			store := &fakeSessionStore{}
			service := newTestAuthService(repo, store, newFakeThrottle(), &fakeLoginRecorder{})

			err := service.ChangeMyPassword(context.Background(), changeIdentity(auth), "token",
				ChangePasswordInput{CurrentPassword: "Corriente1!", NewPassword: tc.pass})
			domainErr := requireKind(t, err, apperr.KindInvalid)
			if detail, ok := domainErr.Details["newPassword"]; !ok || detail == "" {
				t.Fatalf("details = %+v, se esperaba el requisito en newPassword", domainErr.Details)
			}
			if len(repo.updatedHashes) != 0 {
				t.Fatal("una contraseña fuera de política no debe guardarse")
			}
		})
	}
}

func TestChangeMyPasswordRevokesOtherSessionsOnly(t *testing.T) {
	repo := newFakeAuthRepo()
	auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
	repo.authByEmail["ana@ejemplo.com"] = auth
	store := &fakeSessionStore{}
	service := newTestAuthService(repo, store, newFakeThrottle(), &fakeLoginRecorder{})

	if err := service.ChangeMyPassword(context.Background(), changeIdentity(auth), "sesion-actual",
		ChangePasswordInput{CurrentPassword: "Corriente1!", NewPassword: "Nueva9#Aa"}); err != nil {
		t.Fatalf("ChangeMyPassword: %v", err)
	}
	call := store.revokeUserExceptCalls[0]
	if call.keepTokens[0] != "sesion-actual" {
		t.Fatalf("se debe conservar la sesión actual: %+v", call)
	}
}

func TestChangeMyPasswordPropagatesErrors(t *testing.T) {
	change := func(repo *fakeAuthRepo, store *fakeSessionStore, auth UserAuth) error {
		service := newTestAuthService(repo, store, newFakeThrottle(), &fakeLoginRecorder{})
		return service.ChangeMyPassword(context.Background(), changeIdentity(auth), "token",
			ChangePasswordInput{CurrentPassword: "Corriente1!", NewPassword: "Nueva9#Aa"})
	}

	cases := []struct {
		name  string
		setup func(*fakeAuthRepo, *fakeSessionStore)
	}{
		{"leer credenciales", func(r *fakeAuthRepo, _ *fakeSessionStore) { r.authErr = errors.New("bd caída") }},
		{"guardar contraseña", func(r *fakeAuthRepo, _ *fakeSessionStore) { r.changeErr = errors.New("bd caída") }},
		{"revocar sesiones", func(_ *fakeAuthRepo, s *fakeSessionStore) { s.revokeUserExceptErr = errors.New("redis caído") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeAuthRepo()
			auth := activeAuth(t, "ana@ejemplo.com", "Corriente1!")
			repo.authByEmail["ana@ejemplo.com"] = auth
			store := &fakeSessionStore{}
			tc.setup(repo, store)
			if err := change(repo, store, auth); err == nil {
				t.Fatal("el error del repositorio debe propagarse")
			}
		})
	}
}

func TestLogoutPropagatesRevokeError(t *testing.T) {
	store := &fakeSessionStore{revokeErr: errors.New("redis caído")}
	service := newTestAuthService(newFakeAuthRepo(), store, newFakeThrottle(), &fakeLoginRecorder{})

	if _, err := service.Logout(context.Background(), "token"); err == nil {
		t.Fatal("el fallo al revocar la sesión debe propagarse")
	}
}

func TestRetryAfterSecondsRoundsUp(t *testing.T) {
	// Mismo mínimo de 1 s que middleware.retryAfterSeconds (M5): nunca se anuncia
	// un reintento inmediato.
	if got := retryAfterSeconds(0); got != 1 {
		t.Fatalf("retryAfterSeconds(0) = %d, se esperaba 1", got)
	}
	if got := retryAfterSeconds(-time.Second); got != 1 {
		t.Fatalf("retryAfterSeconds(negativo) = %d, se esperaba 1", got)
	}
	if got := retryAfterSeconds(899 * time.Second); got != 899 {
		t.Fatalf("retryAfterSeconds(899s) = %d", got)
	}
	if got := retryAfterSeconds(899*time.Second + 200*time.Millisecond); got != 900 {
		t.Fatalf("retryAfterSeconds(899.2s) = %d, se esperaba redondeo hacia arriba", got)
	}
}
