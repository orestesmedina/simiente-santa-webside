package usuarios

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/password"
)

// service_init.go implementa la inicialización única del administrador inicial
// (FR-007/US2, decisión P8): crea el rol "Administrador" con TODO el catálogo de
// permisos y la primera cuenta, en una sola transacción serializada por el
// mismo advisory lock que el guard anti-bloqueo (P7). No conoce HTTP ni SQL:
// depende de un puerto del repositorio (arq. R3) y de platform/password.

// adminRoleName es el nombre del rol que crea la inicialización (FR-007/US2).
const adminRoleName = "Administrador"

// messageAlreadyInitialized es el 409 de una inicialización repetida: el
// mensaje que pide US2 esc. 2 ("ya se hizo y no puede repetirse").
const messageAlreadyInitialized = "La inicialización ya se hizo y no puede repetirse"

// InitRepository es el puerto de datos que la inicialización necesita del
// repositorio (lo define quien lo consume, arq. R3). WithAdminGuard no expone el
// tipo concreto: la comprobación de "no existe ninguna cuenta", la creación del
// rol/cuenta y su registro ocurren dentro de la MISMA transacción, de modo que
// dos `Initialize` simultáneos no pueden crear dos administradores (SC-003).
type InitRepository interface {
	WithAdminGuard(ctx context.Context, mutate func(tx GuardTx) error) error
}

// InitServiceDeps agrupa las dependencias de la inicialización.
type InitServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository InitRepository
	// Logger registra fallos internos; nil usa el logger por defecto.
	Logger *slog.Logger
}

// initService implementa la inicialización única del administrador inicial.
type initService struct {
	repository InitRepository
	logger     *slog.Logger
}

// NewInitService construye el servicio de inicialización. Si logger es nil se
// usa el logger por defecto.
func NewInitService(deps InitServiceDeps) *initService {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &initService{repository: deps.Repository, logger: logger}
}

// Initialize crea el administrador inicial (FR-007/US2). Aplica la política de
// contraseñas FR-010 (la cuenta no necesita cambiarla: la eligió quien
// inicializa) y, en una única transacción con el advisory lock del guard:
//
//  1. comprueba que NO exista ninguna cuenta (CountUsers = 0); si ya hay, 409 y
//     nada se crea (la acción es única, US2 esc. 2);
//  2. crea el rol "Administrador" con los 9 permisos del catálogo;
//  3. crea la cuenta inicial activa, con el rol anterior; y
//  4. registra `user.create` SIN actor —el único caso permitido (FR-023)—.
//
// El token de despliegue `X-Setup-Token` lo valida el handler (T230). Devuelve
// el DTO UserItem del contrato (nunca credenciales, FR-003).
func (s *initService) Initialize(ctx context.Context, in InitializeInput) (UserItem, error) {
	email := normalizeEmail(in.Email)
	firstName := strings.TrimSpace(in.FirstName)
	lastName := strings.TrimSpace(in.LastName)
	phone := strings.TrimSpace(in.Phone)

	// La política FR-010 se valida ANTES de abrir la transacción: unos datos
	// inválidos responden 400 sin crear absolutamente nada.
	hash, err := password.Hash(in.Password, password.Context{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	})
	if err != nil {
		return UserItem{}, err
	}

	var created User
	err = s.repository.WithAdminGuard(ctx, func(tx GuardTx) error {
		total, err := tx.CountUsers(ctx)
		if err != nil {
			return fmt.Errorf("contar cuentas: %w", err)
		}
		if total != 0 {
			return apperr.Conflict(messageAlreadyInitialized)
		}

		role, err := tx.InsertRole(ctx, adminRoleName)
		if err != nil {
			return fmt.Errorf("crear el rol %q: %w", adminRoleName, err)
		}

		// El catálogo se lee de la base (T207), no se duplica en código: el rol
		// inicial recibe TODOS los permisos, también los módulos de F3–F9.
		permissions, err := tx.ListPermissions(ctx)
		if err != nil {
			return fmt.Errorf("leer el catálogo de permisos: %w", err)
		}
		for _, permission := range permissions {
			if err := tx.InsertRolePermission(ctx, role.ID, permission.ID); err != nil {
				return fmt.Errorf("asociar el permiso %q al rol inicial: %w", permission.Code, err)
			}
		}

		user, err := tx.InsertUser(ctx, NewUser{
			Email:              email,
			FirstName:          firstName,
			LastName:           lastName,
			Phone:              phone,
			PasswordHash:       hash,
			MustChangePassword: false,
			IsActive:           true,
			RoleID:             role.ID,
		})
		if err != nil {
			return fmt.Errorf("crear la cuenta inicial: %w", err)
		}
		// InsertUser no trae el nombre del rol (mapInsertUserRow §8.1.6); la
		// inicialización lo conoce porque acaba de crear el rol.
		user.RoleName = role.Name

		if _, err := tx.InsertAdminAction(ctx, audit.Action{
			ActorUserID:  nil, // único caso sin actor (FR-023)
			Code:         audit.ActionUserCreate,
			TargetKind:   audit.TargetUser,
			TargetUserID: &user.ID,
			TargetLabel:  user.Email,
			Result:       audit.ResultSuccess,
		}); err != nil {
			return fmt.Errorf("registrar la inicialización: %w", err)
		}

		created = user
		return nil
	})
	if err != nil {
		return UserItem{}, err
	}
	return UserItemFrom(created), nil
}

// initService implementa el puerto que publica el handler de inicialización.
var _ SetupService = (*initService)(nil)
