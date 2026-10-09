package usuarios

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/password"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/validate"
)

// service_users.go implementa la gestión de cuentas del panel: listar y ver,
// crear, editar (datos, rol y estado) y restablecer la contraseña
// (FR-009…FR-013, FR-019, FR-021, US3/US5/US7). No conoce HTTP ni SQL: depende
// de un puerto del repositorio (arq. R3), del Store de sesiones (Redis, P9) y
// del registro de auditoría (T224/P20).
//
// Cada mutación que puede dejar el panel sin administración pasa por
// `WithAdminGuard` (FR-008/P7): la escritura, su registro `admin_actions` y el
// guard anti-bloqueo ocurren en la MISMA transacción (o todo o nada, R23). Los
// desenlaces que fallan antes o dentro de la transacción dejan su fila con
// `result='failure'` de forma best-effort (nunca cambian la respuesta, P20).

// Mensajes seguros para el cliente (contrato OpenAPI y ux.md §7).
const (
	// messageEmailInUse es el 409 del correo duplicado, también cuando solo
	// difiere en mayúsculas o espacios (Q5/SC-011).
	messageEmailInUse = "Ese correo ya está en uso"
	// messageRoleNotFound es el 400 de un rol inexistente (US3 esc. 5).
	messageRoleNotFound = "El rol seleccionado no existe"
	// messageNoChanges es el 400 de un PATCH sin ningún campo (minProperties: 1).
	messageNoChanges = "No hay cambios que guardar"
)

// UserRepository es el puerto de datos que la gestión de cuentas necesita del
// repositorio (lo define quien lo consume, arq. R3). Los errores de PostgreSQL
// llegan ya traducidos a apperr.
type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (Role, error)
	ListUsers(ctx context.Context, params paginate.Params) ([]User, error)
	CountUsers(ctx context.Context) (int64, error)
	// WithAdminGuard ejecuta la mutación en una transacción serializada por el
	// advisory lock y comprueba DESPUÉS que sigue habiendo al menos un
	// administrador activo (FR-008/P7). El callback recibe GuardTx.
	WithAdminGuard(ctx context.Context, mutate func(tx GuardTx) error) error
}

// ActionRecorder es el punto de escritura best-effort de las acciones que ya
// van a fallar (P20). Lo implementa *auditService (T224); su fallo nunca cambia
// la respuesta que ve la persona.
type ActionRecorder interface {
	RecordActionBestEffort(ctx context.Context, action audit.Action)
}

// UserServiceDeps agrupa las dependencias de la gestión de cuentas.
type UserServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository UserRepository
	// Sessions revoca las sesiones de una cuenta al desactivarla o restablecer
	// su contraseña (FR-012/R17). Puede ser nil en pruebas que no lo necesiten.
	Sessions session.Store
	// Audit registra best-effort los desenlaces fallidos (P20). Puede ser nil.
	Audit ActionRecorder
	// Logger registra los fallos best-effort; nil usa el logger por defecto.
	Logger *slog.Logger
}

// userService implementa la gestión de cuentas del panel.
type userService struct {
	repository UserRepository
	sessions   session.Store
	audit      ActionRecorder
	logger     *slog.Logger
}

// NewUserService construye el servicio de gestión de cuentas. Si logger es nil
// se usa el logger por defecto.
func NewUserService(deps UserServiceDeps) *userService {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &userService{
		repository: deps.Repository,
		sessions:   deps.Sessions,
		audit:      deps.Audit,
		logger:     logger,
	}
}

// ListUsers devuelve una página del listado de cuentas con su estado, correo,
// rol y último acceso (FR-019/FR-021). Los límites ya vienen normalizados por
// platform/paginate (P14).
func (s *userService) ListUsers(ctx context.Context, params paginate.Params) (UserList, error) {
	users, err := s.repository.ListUsers(ctx, params)
	if err != nil {
		return UserList{}, fmt.Errorf("listar cuentas: %w", err)
	}
	total, err := s.repository.CountUsers(ctx)
	if err != nil {
		return UserList{}, fmt.Errorf("contar cuentas: %w", err)
	}
	items := make([]UserItem, 0, len(users))
	for _, user := range users {
		items = append(items, UserItemFrom(user))
	}
	return UserList{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

// GetUser devuelve la ficha de una cuenta (FR-019/FR-021). Inexistente → 404.
// Una cuenta que nunca entró responde con `lastLoginAt`/`lastLoginIp` en nil,
// que el DTO serializa como `null` (US8 esc. 6).
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (UserItem, error) {
	user, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		return UserItem{}, err
	}
	return UserItemFrom(user), nil
}

// CreateUser crea una cuenta activa con la contraseña inicial definida por el
// administrador (FR-009/US3). Valida la forma (platform/validate), exige un rol
// existente (400 con details.roleId), normaliza el correo (Q5), rechaza
// duplicados con 409, aplica la política FR-010 a la contraseña y deja la
// cuenta con `mustChangePassword = true`. En una única transacción con el
// guard crea la cuenta y registra `user.create` con actor y objetivo (FR-023).
func (s *userService) CreateUser(ctx context.Context, actorID uuid.UUID, in CreateUserInput) (UserItem, error) {
	if err := validate.Struct(in); err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserCreate, nil, "")
		return UserItem{}, err
	}

	email := normalizeEmail(in.Email)
	firstName := strings.TrimSpace(in.FirstName)
	lastName := strings.TrimSpace(in.LastName)
	phone := strings.TrimSpace(in.Phone)

	roleID, err := uuid.Parse(strings.TrimSpace(in.RoleID))
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserCreate, nil, email)
		return UserItem{}, roleNotFoundError()
	}

	// La política FR-010 se valida antes de abrir la transacción: un rechazo no
	// escribe nada.
	hash, err := password.Hash(in.Password, password.Context{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	})
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserCreate, nil, email)
		return UserItem{}, err
	}

	role, err := s.repository.GetRoleByID(ctx, roleID)
	if err != nil {
		if isNotFound(err) {
			s.recordFailure(ctx, actorID, audit.ActionUserCreate, nil, email)
			return UserItem{}, roleNotFoundError()
		}
		return UserItem{}, fmt.Errorf("comprobar el rol: %w", err)
	}

	if err := s.ensureEmailAvailable(ctx, email, uuid.Nil); err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserCreate, nil, email)
		return UserItem{}, err
	}

	var created User
	err = s.repository.WithAdminGuard(ctx, func(tx GuardTx) error {
		user, err := tx.InsertUser(ctx, NewUser{
			Email:              email,
			FirstName:          firstName,
			LastName:           lastName,
			Phone:              phone,
			PasswordHash:       hash,
			MustChangePassword: true,
			IsActive:           true,
			RoleID:             roleID,
		})
		if err != nil {
			return err
		}
		// InsertUser no trae el nombre del rol (mapInsertUserRow §8.1.6); se
		// conoce porque se acaba de comprobar.
		user.RoleName = role.Name

		if _, err := tx.InsertAdminAction(ctx, audit.Action{
			ActorUserID:  &actorID,
			Code:         audit.ActionUserCreate,
			TargetKind:   audit.TargetUser,
			TargetUserID: &user.ID,
			TargetLabel:  user.Email,
			Result:       audit.ResultSuccess,
		}); err != nil {
			return fmt.Errorf("registrar la creación: %w", err)
		}
		created = user
		return nil
	})
	if err != nil {
		err = normalizeConflict(err)
		s.recordFailure(ctx, actorID, audit.ActionUserCreate, nil, email)
		return UserItem{}, err
	}
	return UserItemFrom(created), nil
}

// UpdateUser edita los datos, el rol (el nuevo reemplaza al anterior, Q4) y el
// estado de una cuenta (FR-011/US5). Al desactivar revoca todas sus sesiones
// (FR-012/P9) y conserva los datos (FR-013). Toda mutación pasa por el guard
// anti-bloqueo (FR-008/P7): desactivar o cambiar el rol de forma que quede 0
// cuentas activas con `admin_usuarios_roles` → 409 con
// details.reason="admin_required". Registra user.update / user.activate /
// user.deactivate dentro de la misma transacción (FR-023).
func (s *userService) UpdateUser(ctx context.Context, actorID uuid.UUID, id uuid.UUID, in UpdateUserInput) (UserItem, error) {
	if err := validate.Struct(in); err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserUpdate, nil, "")
		return UserItem{}, err
	}
	if !updateHasChanges(in) {
		s.recordFailure(ctx, actorID, audit.ActionUserUpdate, nil, "")
		return UserItem{}, apperr.Invalid(messageNoChanges)
	}

	current, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		// La cuenta no existe: no se referencia como objetivo (la FK de
		// admin_actions lo rechazaría); el fallo se registra igual.
		s.recordFailure(ctx, actorID, audit.ActionUserUpdate, nil, "")
		return UserItem{}, err
	}

	update, role, err := s.mergeUpdate(ctx, current, in)
	if err != nil {
		s.recordFailure(ctx, actorID, updateActionCode(current, in), &id, current.Email)
		return UserItem{}, err
	}
	code := updateActionCode(current, in)

	var updated User
	err = s.repository.WithAdminGuard(ctx, func(tx GuardTx) error {
		user, err := tx.UpdateUser(ctx, update)
		if err != nil {
			return err
		}
		user.RoleName = role.Name

		if _, err := tx.InsertAdminAction(ctx, audit.Action{
			ActorUserID:  &actorID,
			Code:         code,
			TargetKind:   audit.TargetUser,
			TargetUserID: &user.ID,
			TargetLabel:  user.Email,
			Result:       audit.ResultSuccess,
		}); err != nil {
			return fmt.Errorf("registrar la edición: %w", err)
		}
		updated = user
		return nil
	})
	if err != nil {
		err = normalizeConflict(err)
		s.recordFailure(ctx, actorID, code, &id, current.Email)
		return UserItem{}, err
	}

	// Desactivar corta el acceso de inmediato (FR-012/P9). Es best-effort: aunque
	// Redis falle, la cuenta ya está inactiva y authn la rechaza en cada
	// petición; devolver un error solo confundiría (la baja ya se completó).
	if !updated.IsActive {
		s.revokeUserSessions(ctx, id)
	}
	return UserItemFrom(updated), nil
}

// ResetUserPassword define una contraseña nueva para una cuenta (FR-010/US7
// esc. 5): aplica la política, guarda el hash bcrypt, marca
// `mustChangePassword = true` (su titular debe cambiarla al entrar) y revoca
// todas sus sesiones (R17). Registra `user.password_reset` con actor, objetivo
// y fecha —nunca la contraseña (FR-026)—. Cuenta inexistente → 404.
func (s *userService) ResetUserPassword(
	ctx context.Context,
	actorID uuid.UUID,
	id uuid.UUID,
	in ResetPasswordInput,
) error {
	if err := validate.Struct(in); err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserPasswordReset, nil, "")
		return err
	}

	current, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		// Cuenta inexistente: se registra el fallo sin referencia al objetivo
		// (la FK de admin_actions no admite un user_id que no existe).
		s.recordFailure(ctx, actorID, audit.ActionUserPasswordReset, nil, "")
		return err
	}

	// La política se valida (y el hash se calcula) fuera de la transacción: el
	// coste de bcrypt no debe retener el advisory lock del guard.
	hash, err := password.Hash(in.Password, password.Context{
		FirstName: current.FirstName,
		LastName:  current.LastName,
		Email:     current.Email,
	})
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserPasswordReset, &id, current.Email)
		return err
	}

	err = s.repository.WithAdminGuard(ctx, func(tx GuardTx) error {
		if err := tx.UpdateUserPassword(ctx, id, hash); err != nil {
			return err
		}
		if err := tx.SetUserMustChangePassword(ctx, id, true); err != nil {
			return err
		}
		if _, err := tx.InsertAdminAction(ctx, audit.Action{
			ActorUserID:  &actorID,
			Code:         audit.ActionUserPasswordReset,
			TargetKind:   audit.TargetUser,
			TargetUserID: &id,
			TargetLabel:  current.Email,
			Result:       audit.ResultSuccess,
		}); err != nil {
			return fmt.Errorf("registrar el restablecimiento: %w", err)
		}
		return nil
	})
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionUserPasswordReset, &id, current.Email)
		return err
	}

	// R17: el titular debe cambiar la contraseña y no conserva ninguna sesión.
	s.revokeUserSessions(ctx, id)
	return nil
}

// mergeUpdate combina la cuenta actual con los campos presentes en la entrada
// (un campo vacío significa "no se toca") y resuelve el rol nuevo, si lo hay.
// Aplica las mismas validaciones de negocio que la creación (rol existente,
// correo normalizado y sin duplicados).
func (s *userService) mergeUpdate(ctx context.Context, current User, in UpdateUserInput) (UserUpdate, Role, error) {
	update := UserUpdate{
		ID:        current.ID,
		Email:     current.Email,
		FirstName: current.FirstName,
		LastName:  current.LastName,
		Phone:     current.Phone,
		RoleID:    current.RoleID,
		IsActive:  current.IsActive,
	}
	if value := strings.TrimSpace(in.FirstName); value != "" {
		update.FirstName = value
	}
	if value := strings.TrimSpace(in.LastName); value != "" {
		update.LastName = value
	}
	if value := strings.TrimSpace(in.Phone); value != "" {
		update.Phone = value
	}
	if value := strings.TrimSpace(in.Email); value != "" {
		update.Email = normalizeEmail(value)
	}
	if in.IsActive != nil {
		update.IsActive = *in.IsActive
	}

	role := Role{ID: current.RoleID, Name: current.RoleName}
	if rawRole := strings.TrimSpace(in.RoleID); rawRole != "" {
		roleID, err := uuid.Parse(rawRole)
		if err != nil {
			return UserUpdate{}, Role{}, roleNotFoundError()
		}
		resolved, err := s.repository.GetRoleByID(ctx, roleID)
		if err != nil {
			if isNotFound(err) {
				return UserUpdate{}, Role{}, roleNotFoundError()
			}
			return UserUpdate{}, Role{}, fmt.Errorf("comprobar el rol: %w", err)
		}
		role = resolved
		update.RoleID = resolved.ID
	}

	if !strings.EqualFold(update.Email, current.Email) {
		if err := s.ensureEmailAvailable(ctx, update.Email, current.ID); err != nil {
			return UserUpdate{}, Role{}, err
		}
	}
	return update, role, nil
}

// ensureEmailAvailable comprueba que el correo normalizado no esté en uso por
// OTRA cuenta (exceptID). Devuelve apperr.Conflict con el mensaje del contrato
// y nil si está libre.
func (s *userService) ensureEmailAvailable(ctx context.Context, email string, exceptID uuid.UUID) error {
	existing, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if isNotFound(err) {
			return nil // libre
		}
		return fmt.Errorf("comprobar el correo: %w", err)
	}
	if existing.ID != exceptID {
		return apperr.Conflict(messageEmailInUse)
	}
	return nil
}

// updateHasChanges indica si el PATCH trae al menos un campo (minProperties: 1).
func updateHasChanges(in UpdateUserInput) bool {
	return strings.TrimSpace(in.FirstName) != "" ||
		strings.TrimSpace(in.LastName) != "" ||
		strings.TrimSpace(in.Email) != "" ||
		strings.TrimSpace(in.Phone) != "" ||
		strings.TrimSpace(in.RoleID) != "" ||
		in.IsActive != nil
}

// updateActionCode elige el código de FR-023 según el efecto real de la edición.
func updateActionCode(current User, in UpdateUserInput) string {
	if in.IsActive != nil {
		switch {
		case !*in.IsActive && current.IsActive:
			return audit.ActionUserDeactivate
		case *in.IsActive && !current.IsActive:
			return audit.ActionUserActivate
		}
	}
	return audit.ActionUserUpdate
}

// revokeUserSessions corta las sesiones abiertas de una cuenta (FR-012/R17). Es
// best-effort: authn ya rechaza la cuenta inactiva en cada petición, así que un
// fallo de Redis queda en el log y no cambia la respuesta.
func (s *userService) revokeUserSessions(ctx context.Context, id uuid.UUID) {
	if s.sessions == nil {
		return
	}
	if err := s.sessions.RevokeUser(ctx, id); err != nil {
		s.logError(ctx, "no se pudieron revocar las sesiones de la cuenta", err)
	}
}

// recordFailure deja la fila `result='failure'` de una operación de gestión que
// no se completó (P20). Best-effort: nunca cambia la respuesta.
func (s *userService) recordFailure(ctx context.Context, actorID uuid.UUID, code string, targetID *uuid.UUID, label string) {
	if s.audit == nil {
		return
	}
	action := audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetUser,
		TargetLabel: label,
		Result:      audit.ResultFailure,
	}
	if targetID != nil {
		action.TargetUserID = targetID
	}
	s.audit.RecordActionBestEffort(ctx, action)
}

// logError registra un fallo best-effort con el logger de la petición (con
// request_id) o con el logger base del servicio.
func (s *userService) logError(ctx context.Context, message string, err error) {
	logger := s.logger
	if requestLogger := httpserver.RequestLoggerFromContext(ctx); requestLogger != nil {
		logger = requestLogger
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Error(message, slog.Any("error", err))
}

// roleNotFoundError es el 400 uniforme de un rol inexistente, con details.roleId.
func roleNotFoundError() error {
	return apperr.Invalid(
		messageRoleNotFound,
		apperr.WithDetails(map[string]any{"roleId": messageRoleNotFound}),
	)
}

// normalizeConflict traduce el 409 de un correo duplicado al mensaje del
// contrato, SIN tocar el 409 del guard anti-bloqueo (FR-008), que lleva
// details.reason="admin_required" y debe propagarse tal cual.
func normalizeConflict(err error) error {
	if !isConflict(err) || isAdminRequired(err) {
		return err
	}
	return apperr.Conflict(messageEmailInUse, apperr.WithCause(err))
}

// isConflict indica si el error es un apperr.Conflict.
func isConflict(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindConflict
}

// isAdminRequired reconoce el 409 del guard anti-bloqueo (FR-008/P7).
func isAdminRequired(err error) bool {
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		return false
	}
	return domainErr.Kind == apperr.KindConflict && domainErr.Details["reason"] == "admin_required"
}

// userService implementa el puerto que publican los handlers de cuentas.
var _ UsersService = (*userService)(nil)
