package usuarios

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
)

// auditResult traduce el texto del resultado de una fila de auditoría al tipo
// del registro cerrado de platform/audit.
func auditResult(value string) audit.Result { return audit.Result(value) }

// repository implementa el acceso a datos del dominio usuarios sobre el
// *pgxpool.Pool compartido, con las consultas generadas por sqlc (T210/T211).
// No conoce HTTP ni las reglas de negocio (arq. R4): solo SQL, transacciones y
// el mapeo de tipos. Los tipos pgtype.* de pgx se traducen AQUÍ y solo aquí
// (§8.1.6); el dominio usa uuid.UUID, time.Time y string.
type repository struct {
	q    *gendb.Queries
	pool *pgxpool.Pool
}

// NewRepository construye el repositorio del dominio usuarios a partir del pool
// compartido que compone cmd/api. Devuelve el tipo concreto: la interfaz que
// consumen los servicios la declaran ellos (arq. R3).
func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{q: gendb.New(pool), pool: pool}
}

// --- Mapeo pgtype.* → dominio (§8.1.6) ---

// uuidValue traduce una columna UUID NOT NULL.
func uuidValue(v pgtype.UUID) uuid.UUID {
	return uuid.UUID(v.Bytes)
}

// uuidPtr traduce una columna UUID anulable a *uuid.UUID (nil si es NULL).
func uuidPtr(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	u := uuid.UUID(v.Bytes)
	return &u
}

// pgUUID traduce un uuid.UUID del dominio al parámetro pgtype.UUID.
func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// pgUUIDPtr traduce un *uuid.UUID al parámetro pgtype.UUID (NULL si es nil).
func pgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*id)
}

// timePtr traduce una columna TIMESTAMPTZ anulable a *time.Time (nil si NULL).
func timePtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

// pgTimestamptz traduce un time.Time al parámetro pgtype.Timestamptz.
func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// pgTimestamptzOpt traduce un *time.Time al parámetro pgtype.Timestamptz (NULL
// si es nil): los filtros de fecha de auditoría son opcionales.
func pgTimestamptzOpt(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgTimestamptz(*t)
}

// pgLabel traduce la etiqueta de un objetivo de acción al parámetro TEXT: un
// valor vacío se guarda como NULL (target_label es anulable).
func pgLabel(label string) pgtype.Text {
	if label == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: label, Valid: true}
}

// textPtr traduce una columna TEXT anulable a *string (nil si NULL).
func textPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

// pgTextPtr traduce un *string al parámetro pgtype.Text (NULL si es nil).
func pgTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// mapGetUserByIDRow traduce la fila de GetUserByID/GetUserByEmail.
func mapGetUserByIDRow(row gendb.GetUserByIDRow) User {
	return User{
		ID:                 uuidValue(row.ID),
		Email:              row.Email,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		Phone:              row.Phone,
		MustChangePassword: row.MustChangePassword,
		IsActive:           row.IsActive,
		RoleID:             uuidValue(row.RoleID),
		RoleName:           row.RoleName,
		LastLoginAt:        timePtr(row.LastLoginAt),
		LastLoginIP:        textPtr(row.LastLoginIp),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

// mapGetUserByEmailRow traduce la fila de GetUserByEmail.
func mapGetUserByEmailRow(row gendb.GetUserByEmailRow) User {
	return User{
		ID:                 uuidValue(row.ID),
		Email:              row.Email,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		Phone:              row.Phone,
		MustChangePassword: row.MustChangePassword,
		IsActive:           row.IsActive,
		RoleID:             uuidValue(row.RoleID),
		RoleName:           row.RoleName,
		LastLoginAt:        timePtr(row.LastLoginAt),
		LastLoginIP:        textPtr(row.LastLoginIp),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

// mapListUsersRow traduce una fila de ListUsers.
func mapListUsersRow(row gendb.ListUsersRow) User {
	return User{
		ID:                 uuidValue(row.ID),
		Email:              row.Email,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		Phone:              row.Phone,
		MustChangePassword: row.MustChangePassword,
		IsActive:           row.IsActive,
		RoleID:             uuidValue(row.RoleID),
		RoleName:           row.RoleName,
		LastLoginAt:        timePtr(row.LastLoginAt),
		LastLoginIP:        textPtr(row.LastLoginIp),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

// mapInsertUserRow traduce la fila de InsertUser. `RoleName` no viene en el
// RETURNING (no se puede unir en la misma consulta): el service relee con
// GetUserByID cuando necesita el DTO completo.
func mapInsertUserRow(row gendb.InsertUserRow) User {
	return User{
		ID:                 uuidValue(row.ID),
		Email:              row.Email,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		Phone:              row.Phone,
		MustChangePassword: row.MustChangePassword,
		IsActive:           row.IsActive,
		RoleID:             uuidValue(row.RoleID),
		LastLoginAt:        timePtr(row.LastLoginAt),
		LastLoginIP:        textPtr(row.LastLoginIp),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

// mapUpdateUserRow traduce la fila de UpdateUser.
func mapUpdateUserRow(row gendb.UpdateUserRow) User {
	return User{
		ID:                 uuidValue(row.ID),
		Email:              row.Email,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		Phone:              row.Phone,
		MustChangePassword: row.MustChangePassword,
		IsActive:           row.IsActive,
		RoleID:             uuidValue(row.RoleID),
		LastLoginAt:        timePtr(row.LastLoginAt),
		LastLoginIP:        textPtr(row.LastLoginIp),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

// mapUserAuth traduce la proyección de autenticación (FR-002/FR-018). Los
// permisos vacíos llegan como `[]` (nunca nil) para que la identidad resuelta
// serialice `[]`.
func mapUserAuth(row gendb.GetUserAuthByEmailRow) UserAuth {
	permissions := row.Permissions
	if permissions == nil {
		permissions = []string{}
	}
	return UserAuth{
		ID:                 uuidValue(row.ID),
		Email:              row.Email,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		PasswordHash:       row.PasswordHash,
		MustChangePassword: row.MustChangePassword,
		IsActive:           row.IsActive,
		RoleID:             uuidValue(row.RoleID),
		RoleName:           row.RoleName,
		Permissions:        permissions,
	}
}

// mapRole traduce una fila de roles sin permisos ni recuento (InsertRole,
// GetRoleByNameLower, UpdateRoleName).
func mapRole(row gendb.Role) Role {
	return Role{
		ID:        uuidValue(row.ID),
		Name:      row.Name,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

// mapRoleWithDetails traduce una fila de GetRoleByID (permisos y recuento
// embebidos) a la entidad de dominio.
func mapRoleWithDetails(row gendb.GetRoleByIDRow) Role {
	return Role{
		ID:          uuidValue(row.ID),
		Name:        row.Name,
		Permissions: ensureStrings(row.Permissions),
		UserCount:   row.UserCount,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

// mapListRolesRow traduce una fila de ListRoles.
func mapListRolesRow(row gendb.ListRolesRow) Role {
	return Role{
		ID:          uuidValue(row.ID),
		Name:        row.Name,
		Permissions: ensureStrings(row.Permissions),
		UserCount:   row.UserCount,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

// mapPermission traduce una fila del catálogo (FR-015).
func mapPermission(row gendb.Permission) Permission {
	return Permission{ID: uuidValue(row.ID), Code: row.Code, Label: row.Label}
}

// mapInsertLoginEventRow traduce la fila de InsertLoginEvent.
func mapInsertLoginEventRow(row gendb.InsertLoginEventRow) LoginEvent {
	return LoginEvent{
		ID:        uuidValue(row.ID),
		UserID:    uuidPtr(row.UserID),
		Result:    auditResult(row.Result),
		IP:        row.Ip,
		CreatedAt: row.CreatedAt.Time,
	}
}

// mapListLoginEventsRow traduce una fila de ListLoginEvents; compone el nombre
// derivado `firstName lastName` (F-01) a partir de las partes anulables.
func mapListLoginEventsRow(row gendb.ListLoginEventsRow) AccessEvent {
	return AccessEvent{
		ID:        uuidValue(row.ID),
		UserID:    uuidPtr(row.UserID),
		UserEmail: textPtr(row.UserEmail),
		UserName:  fullNamePtr(row.UserFirstName, row.UserLastName),
		Result:    auditResult(row.Result),
		IP:        row.Ip,
		CreatedAt: row.CreatedAt.Time,
	}
}

// mapInsertAdminActionRow traduce la fila de InsertAdminAction. El objetivo se
// aplana (`TargetID` = cuenta o rol) igual que en el listado.
func mapInsertAdminActionRow(row gendb.InsertAdminActionRow) AdminAction {
	return AdminAction{
		ID:           uuidValue(row.ID),
		ActorUserID:  uuidPtr(row.ActorUserID),
		Action:       row.Action,
		TargetKind:   audit.TargetKind(row.TargetKind),
		TargetUserID: uuidPtr(row.TargetUserID),
		TargetRoleID: uuidPtr(row.TargetRoleID),
		TargetLabel:  textPtr(row.TargetLabel),
		Result:       auditResult(row.Result),
		CreatedAt:    row.CreatedAt.Time,
	}
}

// mapListAdminActionsRow traduce una fila de ListAdminActions; compone el
// nombre derivado del actor (F-01).
func mapListAdminActionsRow(row gendb.ListAdminActionsRow) AdminActionEntry {
	return AdminActionEntry{
		ID:          uuidValue(row.ID),
		ActorUserID: uuidPtr(row.ActorUserID),
		ActorEmail:  textPtr(row.ActorEmail),
		ActorName:   fullNamePtr(row.ActorFirstName, row.ActorLastName),
		Action:      row.Action,
		TargetKind:  audit.TargetKind(row.TargetKind),
		TargetID:    uuidPtr(row.TargetID),
		TargetLabel: textPtr(row.TargetLabel),
		Result:      auditResult(row.Result),
		CreatedAt:   row.CreatedAt.Time,
	}
}

// --- Utilidades de mapeo ---

// ensureStrings devuelve `[]` en lugar de nil para que los DTO de salida
// serialicen un array vacío (§8.1.2).
func ensureStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// fullNamePtr compone `firstName lastName` (F-01) solo cuando la fila tiene
// nombre; si ambas partes son NULL devuelve nil (intento sin cuenta asociada).
func fullNamePtr(firstName, lastName pgtype.Text) *string {
	if !firstName.Valid && !lastName.Valid {
		return nil
	}
	name := ""
	if firstName.Valid {
		name = firstName.String
	}
	if lastName.Valid {
		if name != "" {
			name += " "
		}
		name += lastName.String
	}
	if name == "" {
		return nil
	}
	return &name
}

// --- Traducción de errores de PostgreSQL ---

// Códigos SQLSTATE que el dominio conoce. Se comparan como string para no
// añadir una dependencia adicional (skill postgres-db): 23505 unique_violation,
// 23503 foreign_key_violation, 23514 check_violation.
const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
	codeCheckViolation      = "23514"
)

// errTargetReference señala que una escritura del registro de auditoría
// referencia un objetivo (target_user_id/target_role_id) que no existe: la FK
// lo bloquea. La auditoría best-effort lo usa para reintentar la fila con el
// objetivo en nil y no perder el registro (FR-023).
var errTargetReference = errors.New("usuarios: el objetivo de la acción no existe")

// isForeignKeyViolation indica si err es una violación de clave foránea de
// PostgreSQL (SQLSTATE 23503).
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codeForeignKeyViolation
}

// classify traduce un error de PostgreSQL a un error de dominio. `ErrNoRows`
// pasa a apperr.NotFound (404) y las violaciones conocidas de integridad a
// apperr.Conflict/apperr.Invalid; el resto se devuelve envuelto por el llamador
// con %w. El repositorio es el único punto que conoce pgx (arq. R4), así que la
// traducción ocurre aquí y no en el service.
func classify(err error, notFound, conflict string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.NotFound(notFound, apperr.WithCause(err))
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case codeUniqueViolation:
			return apperr.Conflict(conflict, apperr.WithCause(err))
		case codeForeignKeyViolation:
			return apperr.Invalid(
				"Se referenció un registro que no existe",
				apperr.WithCause(err),
			)
		case codeCheckViolation:
			return apperr.Invalid(
				"Algún dato no cumple las reglas de la base de datos",
				apperr.WithCause(err),
			)
		}
	}
	return err
}

// wrap traduce el error y lo envuelve con contexto de la operación, de modo
// que errors.Is/As sigan encontrando el error de dominio.
func wrap(err error, operation, notFound, conflict string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, classify(err, notFound, conflict))
}

// wrapFKConflict es como wrap pero traduce la violación de clave foránea a
// apperr.Conflict en vez de apperr.Invalid. Lo usa `DeleteRole`: `users.role_id`
// es `ON DELETE RESTRICT` y un rol con cuentas asignadas no se puede eliminar
// (FR-017), que es un conflicto con el estado actual, no una entrada inválida.
func wrapFKConflict(err error, operation, conflict string) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codeForeignKeyViolation {
		return fmt.Errorf("%s: %w", operation, apperr.Conflict(conflict, apperr.WithCause(err)))
	}
	return fmt.Errorf("%s: %w", operation, classify(err, "", ""))
}
