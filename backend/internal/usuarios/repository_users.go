package usuarios

import (
	"context"
	"time"

	"github.com/google/uuid"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/paginate"
)

// Consultas de cuentas del panel (FR-009…FR-013, FR-019, FR-021) sobre sqlc.
// No hay ninguna operación de borrado: retirar el acceso es `is_active = false`
// (FR-013). Los errores de PostgreSQL se traducen con `classify`/`wrap` a
// apperr (404/409/400) para que el service no conozca pgx (arq. R4).

// GetUserAuthByEmail devuelve la proyección de autenticación de la cuenta
// (correo + hash + estado + rol + permisos) para el login y authn (FR-002,
// FR-018). Correo inexistente → apperr.NotFound, que el login traduce al 401
// genérico (FR-003).
func (r *repository) GetUserAuthByEmail(ctx context.Context, email string) (UserAuth, error) {
	row, err := r.q.GetUserAuthByEmail(ctx, email)
	if err != nil {
		return UserAuth{}, wrap(err, "get user auth by email", "La cuenta no existe", "")
	}
	return mapUserAuth(row), nil
}

// GetUserByID devuelve la ficha de una cuenta (FR-019/FR-021). Inexistente →
// apperr.NotFound (404).
func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row, err := r.q.GetUserByID(ctx, pgUUID(id))
	if err != nil {
		return User{}, wrap(err, "get user by id", "La cuenta no existe", "")
	}
	return mapGetUserByIDRow(row), nil
}

// GetUserByEmail devuelve la cuenta por correo normalizado (Q5). Se usa para la
// comprobación previa de duplicados; el `UNIQUE` de la base sigue siendo la red
// de seguridad (R9).
func (r *repository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, wrap(err, "get user by email", "La cuenta no existe", "")
	}
	return mapGetUserByEmailRow(row), nil
}

// ListUsers devuelve una página del listado de cuentas en el orden por defecto
// (`created_at DESC, id DESC`, §8.1.7). El service acota siempre limit/offset
// (P14).
func (r *repository) ListUsers(ctx context.Context, params paginate.Params) ([]User, error) {
	rows, err := r.q.ListUsers(ctx, gendb.ListUsersParams{
		Off: int32(params.Offset),
		Lim: int32(params.Limit),
	})
	if err != nil {
		return nil, wrap(err, "list users", "", "")
	}
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, mapListUsersRow(row))
	}
	return users, nil
}

// CountUsers devuelve el total de cuentas (para el sobre del listado y la
// inicialización única, FR-007).
func (r *repository) CountUsers(ctx context.Context) (int64, error) {
	total, err := r.q.CountUsers(ctx)
	if err != nil {
		return 0, wrap(err, "count users", "", "")
	}
	return total, nil
}

// CountUsersByRole cuenta las cuentas con un rol (FR-017).
func (r *repository) CountUsersByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	total, err := r.q.CountUsersByRole(ctx, pgUUID(roleID))
	if err != nil {
		return 0, wrap(err, "count users by role", "", "")
	}
	return total, nil
}

// InsertUser crea una cuenta (FR-009). El hash ya viene calculado por el
// service. Correo duplicado (normalizado) → apperr.Conflict; rol inexistente →
// apperr.Invalid (la FK lo detecta).
func (r *repository) InsertUser(ctx context.Context, user NewUser) (User, error) {
	row, err := r.q.InsertUser(ctx, gendb.InsertUserParams{
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Phone:              user.Phone,
		PasswordHash:       user.PasswordHash,
		MustChangePassword: user.MustChangePassword,
		IsActive:           user.IsActive,
		RoleID:             pgUUID(user.RoleID),
	})
	if err != nil {
		return User{}, wrap(err, "insert user", "", "Ya existe una cuenta con ese correo electrónico")
	}
	return mapInsertUserRow(row), nil
}

// UpdateUser edita datos, rol y estado (FR-011). Correo duplicado →
// apperr.Conflict; rol inexistente → apperr.Invalid.
func (r *repository) UpdateUser(ctx context.Context, update UserUpdate) (User, error) {
	row, err := r.q.UpdateUser(ctx, gendb.UpdateUserParams{
		Email:     update.Email,
		FirstName: update.FirstName,
		LastName:  update.LastName,
		Phone:     update.Phone,
		RoleID:    pgUUID(update.RoleID),
		IsActive:  update.IsActive,
		ID:        pgUUID(update.ID),
	})
	if err != nil {
		return User{}, wrap(err, "update user", "La cuenta no existe", "Ya existe una cuenta con ese correo electrónico")
	}
	return mapUpdateUserRow(row), nil
}

// UpdateUserPassword guarda el hash bcrypt (FR-010/§IV). El flag de cambio
// obligatorio se maneja aparte con SetUserMustChangePassword (FR-020).
func (r *repository) UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	err := r.q.UpdateUserPassword(ctx, gendb.UpdateUserPasswordParams{
		PasswordHash: passwordHash,
		ID:           pgUUID(id),
	})
	return wrap(err, "update user password", "", "")
}

// SetUserMustChangePassword marca o limpia la obligación de cambiar la
// contraseña al entrar (US7).
func (r *repository) SetUserMustChangePassword(ctx context.Context, id uuid.UUID, must bool) error {
	err := r.q.SetUserMustChangePassword(ctx, gendb.SetUserMustChangePasswordParams{
		MustChangePassword: must,
		ID:                 pgUUID(id),
	})
	return wrap(err, "set user must change password", "", "")
}

// UpdateUserLastLogin escribe la proyección del último acceso exitoso
// (`last_login_at` + `last_login_ip`) y es SOLO para el login correcto (FR-021).
// Ningún DTO de entrada las acepta.
func (r *repository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID, at time.Time, ip string) error {
	err := r.q.UpdateUserLastLogin(ctx, gendb.UpdateUserLastLoginParams{
		LastLoginAt: pgTimestamptz(at),
		LastLoginIp: pgTextPtr(&ip),
		ID:          pgUUID(id),
	})
	return wrap(err, "update user last login", "", "")
}
