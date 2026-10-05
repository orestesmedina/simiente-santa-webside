// Package usuarios es el único dominio de F2 (P15): cuentas del panel, roles y
// permisos, sesión e identidad, y el registro de auditoría. Sigue las capas
// handler → service → repository (§II): los handlers no tocan SQL y el service
// no conoce net/http ni pgx.
//
// Este archivo reúne las entidades del dominio (con tipos de
// github.com/google/uuid, P4) y los DTOs del contrato OpenAPI con sus etiquetas
// `validate` (T213/§8.1.4). No tiene lógica de negocio.
package usuarios

import (
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// Los estados de una cuenta son `is_active` (activa/desactivada, FR-012) y
// `must_change_password` (cambio obligatorio al entrar, US7). Los resultados de
// un intento o acción y los códigos de acción administrativa viven en
// platform/audit (audit.Result, audit.TargetKind, audit.ActionCodes) y no se
// duplican aquí: es el registro cerrado de FR-022/FR-023.

// --- Entidades ---

// User es una cuenta del panel (FR-009…FR-013, FR-019). Nunca se borra
// (FR-013): retirar el acceso es IsActive=false. PasswordHash no forma parte de
// ningún DTO de salida (FR-003): solo lo devuelve la consulta de autenticación,
// en UserAuth.
type User struct {
	ID                 uuid.UUID
	Email              string
	FirstName          string
	LastName           string
	Phone              string
	MustChangePassword bool
	IsActive           bool
	RoleID             uuid.UUID
	RoleName           string
	// LastLoginAt y LastLoginIP son la proyección del último acceso exitoso
	// (FR-021); nil mientras la cuenta no haya iniciado sesión.
	LastLoginAt *time.Time
	LastLoginIP *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserAuth es la proyección de credenciales que consumen el login y authn:
// correo, hash, estado, rol y los códigos de permiso efectivos (FR-002, FR-018).
// El hash nunca sale del dominio como DTO (§IV, FR-003).
type UserAuth struct {
	ID                 uuid.UUID
	Email              string
	FirstName          string
	LastName           string
	PasswordHash       string
	MustChangePassword bool
	IsActive           bool
	RoleID             uuid.UUID
	RoleName           string
	Permissions        []string
}

// Role es un conjunto de permisos por módulo definido por el administrador
// (FR-014/FR-017). Un rol tiene al menos un permiso y un nombre único
// normalizado.
type Role struct {
	ID          uuid.UUID
	Name        string
	Permissions []string
	// UserCount es cuántas cuentas tienen el rol (0 = se puede eliminar,
	// FR-017).
	UserCount int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Permission es una entrada del catálogo fijo de módulos (FR-015).
type Permission struct {
	ID    uuid.UUID
	Code  string
	Label string
}

// LoginEvent es un intento de inicio de sesión registrado (FR-022). UserID es
// nil cuando el correo no corresponde a ninguna cuenta: el intento queda sin
// asociación (FR-003/FR-026).
type LoginEvent struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Result    audit.Result
	IP        string
	CreatedAt time.Time
}

// AccessEvent es un intento de acceso con los datos derivados de su cuenta
// (F-01): no se guardan en el registro, se resuelven con LEFT JOIN al consultar.
// Los tres campos de cuenta son nil cuando el intento no se asoció a ninguna.
type AccessEvent struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	UserEmail *string
	UserName  *string
	Result    audit.Result
	IP        string
	CreatedAt time.Time
}

// AdminAction es una acción administrativa sensible registrada (FR-023): quién,
// qué, sobre qué, cuándo y con qué resultado.
type AdminAction struct {
	ID           uuid.UUID
	ActorUserID  *uuid.UUID
	Action       string
	TargetKind   audit.TargetKind
	TargetUserID *uuid.UUID
	TargetRoleID *uuid.UUID
	TargetLabel  *string
	Result       audit.Result
	CreatedAt    time.Time
}

// AdminActionEntry es una acción administrativa con el actor derivado de su
// cuenta (F-01) y el objetivo aplanado (`TargetID` = cuenta o rol). Es la forma
// que devuelven los listados; el email y el nombre no se guardan duplicados.
type AdminActionEntry struct {
	ID          uuid.UUID
	ActorUserID *uuid.UUID
	ActorEmail  *string
	ActorName   *string
	Action      string
	TargetKind  audit.TargetKind
	TargetID    *uuid.UUID
	TargetLabel *string
	Result      audit.Result
	CreatedAt   time.Time
}

// --- Parámetros de escritura (los construye el service) ---

// NewUser son los datos con los que se crea una cuenta. El hash ya viene
// calculado por el service (platform/password); el repository solo lo guarda.
type NewUser struct {
	Email              string
	FirstName          string
	LastName           string
	Phone              string
	PasswordHash       string
	MustChangePassword bool
	IsActive           bool
	RoleID             uuid.UUID
}

// UserUpdate son los datos editables de una cuenta (FR-011): nombre, apellidos,
// correo, teléfono, rol y estado. Nunca incluye `last_login_*` (FR-021): no hay
// vía para editarlas.
type UserUpdate struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
	Phone     string
	RoleID    uuid.UUID
	IsActive  bool
}

// NewRole es la creación de un rol con sus permisos (FR-014). Los permisos ya
// vienen resueltos a ids por el service.
type NewRole struct {
	Name          string
	PermissionIDs []uuid.UUID
}

// RoleUpdate es la edición de un rol (FR-017): nombre y/o permisos. Un campo
// vacío/`nil` significa "no se toca".
type RoleUpdate struct {
	ID            uuid.UUID
	Name          string
	PermissionIDs []uuid.UUID
}

// AuditFilter son los filtros de los listados de auditoría (FR-024/P22):
// cuenta, rango de fechas semirango `[From, To)` y paginación.
type AuditFilter struct {
	UserID *uuid.UUID
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// --- DTOs de entrada del contrato (etiquetas `validate`, T213) ---

// LoginInput es el cuerpo de POST /api/v1/auth/login.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,max=64"`
}

// InitializeInput son los datos del primer administrador (FR-007). No lleva
// rol: el rol "Administrador" nace con la inicialización.
type InitializeInput struct {
	FirstName string `json:"firstName" validate:"required,min=1,max=80"`
	LastName  string `json:"lastName" validate:"required,min=1,max=120"`
	Email     string `json:"email" validate:"required,email,max=254"`
	Phone     string `json:"phone" validate:"required,phone,min=7,max=32"`
	Password  string `json:"password" validate:"required,min=8,max=64"`
}

// CreateUserInput es el cuerpo de POST /api/v1/admin/usuarios (FR-009). RoleID
// viaja como UUID en texto; el service lo parsea y comprueba que exista.
type CreateUserInput struct {
	FirstName string `json:"firstName" validate:"required,min=1,max=80"`
	LastName  string `json:"lastName" validate:"required,min=1,max=120"`
	Email     string `json:"email" validate:"required,email,max=254"`
	Phone     string `json:"phone" validate:"required,phone,min=7,max=32"`
	RoleID    string `json:"roleId" validate:"required"`
	Password  string `json:"password" validate:"required,min=8,max=64"`
}

// UpdateUserInput es el cuerpo de PATCH /api/v1/admin/usuarios/{id} (FR-011).
// Un campo vacío significa "no se toca"; el service exige que haya al menos uno
// (minProperties: 1 del contrato). IsActive es puntero para distinguir "no
// enviado" de `false` (desactivar).
type UpdateUserInput struct {
	FirstName string `json:"firstName,omitempty" validate:"omitempty,min=1,max=80"`
	LastName  string `json:"lastName,omitempty" validate:"omitempty,min=1,max=120"`
	Email     string `json:"email,omitempty" validate:"omitempty,email,max=254"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,phone,min=7,max=32"`
	RoleID    string `json:"roleId,omitempty"`
	IsActive  *bool  `json:"isActive,omitempty"`
}

// ResetPasswordInput es el cuerpo de POST /api/v1/admin/usuarios/{id}/password
// (FR-010): el administrador define la contraseña nueva.
type ResetPasswordInput struct {
	Password string `json:"password" validate:"required,min=8,max=64"`
}

// ChangePasswordInput es el cuerpo de POST /api/v1/auth/password (FR-020).
type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword" validate:"required,max=64"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=64"`
}

// RoleCreateInput es el cuerpo de POST /api/v1/admin/roles (FR-014).
type RoleCreateInput struct {
	Name        string   `json:"name" validate:"required,min=1,max=80"`
	Permissions []string `json:"permissions" validate:"required"`
}

// RoleUpdateInput es el cuerpo de PATCH /api/v1/admin/roles/{id} (FR-017).
type RoleUpdateInput struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,min=1,max=80"`
	Permissions []string `json:"permissions,omitempty"`
}

// --- DTOs de salida del contrato ---

// SessionUser es la identidad de la sesión iniciada con sus permisos efectivos
// (FR-018). No contiene credenciales (FR-003).
type SessionUser struct {
	ID                 string   `json:"id"`
	Email              string   `json:"email"`
	FirstName          string   `json:"firstName"`
	LastName           string   `json:"lastName"`
	Phone              string   `json:"phone"`
	RoleID             string   `json:"roleId"`
	RoleName           string   `json:"roleName"`
	Permissions        []string `json:"permissions"`
	MustChangePassword bool     `json:"mustChangePassword"`
}

// UserItem es una cuenta en listados y detalle (FR-019/FR-021). Nunca incluye
// credenciales.
type UserItem struct {
	ID                 string     `json:"id"`
	Email              string     `json:"email"`
	FirstName          string     `json:"firstName"`
	LastName           string     `json:"lastName"`
	Phone              string     `json:"phone"`
	RoleID             string     `json:"roleId"`
	RoleName           string     `json:"roleName"`
	IsActive           bool       `json:"isActive"`
	MustChangePassword bool       `json:"mustChangePassword"`
	LastLoginAt        *time.Time `json:"lastLoginAt"`
	LastLoginIP        *string    `json:"lastLoginIp"`
	CreatedAt          time.Time  `json:"createdAt"`
}

// UserList es el sobre de un listado de cuentas (§8.1.2).
type UserList struct {
	Items  []UserItem `json:"items"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// RoleItem es un rol en listados y detalle (FR-017).
type RoleItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	UserCount   int64     `json:"userCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

// RoleList es el sobre de un listado de roles.
type RoleList struct {
	Items  []RoleItem `json:"items"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// PermissionItem es una entrada del catálogo de permisos (FR-015).
type PermissionItem struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// PermissionList es el sobre del catálogo (no se pagina: es fijo).
type PermissionList struct {
	Items []PermissionItem `json:"items"`
}

// AccessEventItem es un intento de acceso en el historial (FR-022). Los campos
// de la cuenta son derivados y anulables (F-01).
type AccessEventItem struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"userId"`
	UserEmail *string   `json:"userEmail"`
	UserName  *string   `json:"userName"`
	Result    string    `json:"result"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"createdAt"`
}

// AccessEventList es el sobre del historial de accesos.
type AccessEventList struct {
	Items  []AccessEventItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

// AdminActionItem es una acción administrativa en el historial (FR-023).
type AdminActionItem struct {
	ID          string    `json:"id"`
	ActorID     *string   `json:"actorId"`
	ActorEmail  *string   `json:"actorEmail"`
	ActorName   *string   `json:"actorName"`
	Action      string    `json:"action"`
	TargetKind  string    `json:"targetKind"`
	TargetID    *string   `json:"targetId"`
	TargetLabel *string   `json:"targetLabel"`
	Result      string    `json:"result"`
	CreatedAt   time.Time `json:"createdAt"`
}

// AdminActionList es el sobre del historial de acciones.
type AdminActionList struct {
	Items  []AdminActionItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

// LogoutResponse es la respuesta de POST /api/v1/auth/logout.
type LogoutResponse struct {
	LoggedOut bool `json:"loggedOut"`
}

// PasswordChangedResponse es la respuesta de POST /api/v1/auth/password.
type PasswordChangedResponse struct {
	PasswordChanged bool `json:"passwordChanged"`
}

// PasswordResetResponse es la respuesta del restablecimiento por un
// administrador.
type PasswordResetResponse struct {
	PasswordReset bool `json:"passwordReset"`
}

// RoleDeletedResponse es la respuesta de DELETE /api/v1/admin/roles/{id}.
type RoleDeletedResponse struct {
	Deleted bool `json:"deleted"`
}

// --- Conversiones entidad → DTO de salida ---

// UserItemFrom traduce una cuenta a su DTO de listado/detalle.
func UserItemFrom(u User) UserItem {
	return UserItem{
		ID:                 u.ID.String(),
		Email:              u.Email,
		FirstName:          u.FirstName,
		LastName:           u.LastName,
		Phone:              u.Phone,
		RoleID:             u.RoleID.String(),
		RoleName:           u.RoleName,
		IsActive:           u.IsActive,
		MustChangePassword: u.MustChangePassword,
		LastLoginAt:        u.LastLoginAt,
		LastLoginIP:        u.LastLoginIP,
		CreatedAt:          u.CreatedAt,
	}
}

// SessionUserFrom construye la identidad de sesión (FR-018) a partir de la
// cuenta y sus permisos efectivos.
func SessionUserFrom(u User, permissions []string) SessionUser {
	return SessionUser{
		ID:                 u.ID.String(),
		Email:              u.Email,
		FirstName:          u.FirstName,
		LastName:           u.LastName,
		Phone:              u.Phone,
		RoleID:             u.RoleID.String(),
		RoleName:           u.RoleName,
		Permissions:        permissions,
		MustChangePassword: u.MustChangePassword,
	}
}

// RoleItemFrom traduce un rol a su DTO.
func RoleItemFrom(r Role) RoleItem {
	return RoleItem{
		ID:          r.ID.String(),
		Name:        r.Name,
		Permissions: r.Permissions,
		UserCount:   r.UserCount,
		CreatedAt:   r.CreatedAt,
	}
}

// PermissionItemFrom traduce una entrada del catálogo a su DTO.
func PermissionItemFrom(p Permission) PermissionItem {
	return PermissionItem{Code: p.Code, Label: p.Label}
}

// AccessEventItemFrom traduce un intento de acceso a su DTO de historial.
func AccessEventItemFrom(e AccessEvent) AccessEventItem {
	item := AccessEventItem{
		ID:        e.ID.String(),
		UserEmail: e.UserEmail,
		UserName:  e.UserName,
		Result:    string(e.Result),
		IP:        e.IP,
		CreatedAt: e.CreatedAt,
	}
	if e.UserID != nil {
		id := e.UserID.String()
		item.UserID = &id
	}
	return item
}

// AdminActionItemFrom traduce una acción administrativa a su DTO de historial.
func AdminActionItemFrom(a AdminActionEntry) AdminActionItem {
	item := AdminActionItem{
		ID:          a.ID.String(),
		ActorEmail:  a.ActorEmail,
		ActorName:   a.ActorName,
		Action:      a.Action,
		TargetKind:  string(a.TargetKind),
		TargetLabel: a.TargetLabel,
		Result:      string(a.Result),
		CreatedAt:   a.CreatedAt,
	}
	if a.ActorUserID != nil {
		id := a.ActorUserID.String()
		item.ActorID = &id
	}
	if a.TargetID != nil {
		id := a.TargetID.String()
		item.TargetID = &id
	}
	return item
}
