package usuarios

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/database"
	"simiente-santa/backend/internal/platform/paginate"
)

// Consultas de roles y del catálogo de permisos (FR-014…FR-018) sobre sqlc, más
// el guard anti-bloqueo de FR-008 (P7). Errores traducidos con classify/wrap.

// InsertRole crea un rol (FR-014). Nombre duplicado (normalizado) →
// apperr.Conflict; los permisos se insertan aparte, en la misma transacción que
// el service orquesta.
func (r *repository) InsertRole(ctx context.Context, name string) (Role, error) {
	row, err := r.q.InsertRole(ctx, name)
	if err != nil {
		return Role{}, wrap(err, "insert role", "", "Ya existe un rol con ese nombre")
	}
	return mapRole(row), nil
}

// GetRoleByID devuelve un rol con sus permisos y el recuento de cuentas
// (FR-017). Inexistente → apperr.NotFound (404).
func (r *repository) GetRoleByID(ctx context.Context, id uuid.UUID) (Role, error) {
	row, err := r.q.GetRoleByID(ctx, pgUUID(id))
	if err != nil {
		return Role{}, wrap(err, "get role by id", "El rol no existe", "")
	}
	return mapRoleWithDetails(row), nil
}

// GetRoleByNameLower busca por nombre normalizado (Q5): dos nombres que solo
// difieren en mayúsculas o en espacios de los extremos son el mismo rol.
func (r *repository) GetRoleByNameLower(ctx context.Context, name string) (Role, error) {
	row, err := r.q.GetRoleByNameLower(ctx, name)
	if err != nil {
		return Role{}, wrap(err, "get role by name", "El rol no existe", "")
	}
	return mapRole(row), nil
}

// ListRoles devuelve una página de roles con permisos y recuento (FR-017), en
// el orden por defecto (§8.1.7).
func (r *repository) ListRoles(ctx context.Context, params paginate.Params) ([]Role, error) {
	rows, err := r.q.ListRoles(ctx, gendb.ListRolesParams{
		Off: int32(params.Offset),
		Lim: int32(params.Limit),
	})
	if err != nil {
		return nil, wrap(err, "list roles", "", "")
	}
	roles := make([]Role, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, mapListRolesRow(row))
	}
	return roles, nil
}

// CountRoles devuelve el total de roles (sobre del listado).
func (r *repository) CountRoles(ctx context.Context) (int64, error) {
	total, err := r.q.CountRoles(ctx)
	if err != nil {
		return 0, wrap(err, "count roles", "", "")
	}
	return total, nil
}

// UpdateRoleName renombra un rol conservando sus permisos (FR-017).
func (r *repository) UpdateRoleName(ctx context.Context, id uuid.UUID, name string) (Role, error) {
	row, err := r.q.UpdateRoleName(ctx, gendb.UpdateRoleNameParams{
		Name: name,
		ID:   pgUUID(id),
	})
	if err != nil {
		return Role{}, wrap(err, "update role name", "El rol no existe", "Ya existe un rol con ese nombre")
	}
	return mapRole(row), nil
}

// DeleteRolePermissions borra los permisos actuales de un rol; el service
// reinserta el conjunto nuevo en la misma transacción (FR-014). Un rol nunca
// queda sin permisos.
func (r *repository) DeleteRolePermissions(ctx context.Context, roleID uuid.UUID) error {
	err := r.q.DeleteRolePermissions(ctx, pgUUID(roleID))
	return wrap(err, "delete role permissions", "", "")
}

// InsertRolePermission añade un permiso a un rol (FR-014).
func (r *repository) InsertRolePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := r.q.InsertRolePermission(ctx, gendb.InsertRolePermissionParams{
		RoleID:       pgUUID(roleID),
		PermissionID: pgUUID(permissionID),
	})
	return wrap(err, "insert role permission", "", "")
}

// DeleteRole elimina un rol (FR-017). Devuelve el número de filas afectadas;
// con cuentas asignadas la FK `ON DELETE RESTRICT` devuelve apperr.Conflict
// (red de seguridad del service).
func (r *repository) DeleteRole(ctx context.Context, id uuid.UUID) (int64, error) {
	affected, err := r.q.DeleteRole(ctx, pgUUID(id))
	if err != nil {
		return 0, wrapFKConflict(err, "delete role", "No se puede eliminar un rol con cuentas asignadas")
	}
	return affected, nil
}

// CountRoleUsers cuenta las cuentas con un rol (FR-017): 0 = se puede eliminar.
func (r *repository) CountRoleUsers(ctx context.Context, roleID uuid.UUID) (int64, error) {
	total, err := r.q.CountRoleUsers(ctx, pgUUID(roleID))
	if err != nil {
		return 0, wrap(err, "count role users", "", "")
	}
	return total, nil
}

// ListPermissions devuelve el catálogo fijo de permisos (FR-015), ordenado por
// código. No se pagina: es un catálogo.
func (r *repository) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.q.ListPermissions(ctx)
	if err != nil {
		return nil, wrap(err, "list permissions", "", "")
	}
	permissions := make([]Permission, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, mapPermission(row))
	}
	return permissions, nil
}

// GetPermissionIDsByCodes traduce códigos del catálogo a sus ids. Devolver un
// mapa permite al service detectar qué códigos no existen (400 con details) y
// qué ids insertar en `role_permissions` (FR-014, FR-015).
func (r *repository) GetPermissionIDsByCodes(ctx context.Context, codes []string) (map[string]uuid.UUID, error) {
	rows, err := r.q.GetPermissionIDsByCodes(ctx, codes)
	if err != nil {
		return nil, wrap(err, "get permission ids by codes", "", "")
	}
	ids := make(map[string]uuid.UUID, len(rows))
	for _, row := range rows {
		ids[row.Code] = uuidValue(row.ID)
	}
	return ids, nil
}

// CountActiveAdmins cuenta las cuentas activas cuyo rol concede
// `admin_usuarios_roles` (FR-008). Es el recuento POST-mutación del guard.
func (r *repository) CountActiveAdmins(ctx context.Context) (int64, error) {
	total, err := r.q.CountActiveAdmins(ctx)
	if err != nil {
		return 0, wrap(err, "count active admins", "", "")
	}
	return total, nil
}

// withTx ejecuta fn con un repository ligado a una transacción. El service lo
// usa para agrupar una mutación y su registro de auditoría (o el guard) en una
// única transacción (R23/P7).
func (r *repository) withTx(ctx context.Context, fn func(tx *repository) error) error {
	return database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&repository{q: r.q.WithTx(tx), pool: r.pool})
	})
}

// GuardTx es la vista de datos disponible dentro de una transacción del guard
// anti-bloqueo (P7/P8). Se declara como interfaz para que los services que mutan
// dentro del guard dependan de un puerto (skill `go-backend`) y puedan probarse
// con fakes que no tocan PostgreSQL. Crece con cada operación que necesite
// ejecutarse dentro del guard.
type GuardTx interface {
	CountUsers(ctx context.Context) (int64, error)
	InsertRole(ctx context.Context, name string) (Role, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	InsertRolePermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	InsertUser(ctx context.Context, user NewUser) (User, error)
	InsertAdminAction(ctx context.Context, action audit.Action) (AdminAction, error)
	UpdateUser(ctx context.Context, update UserUpdate) (User, error)
}

// WithAdminGuard ejecuta la mutación dentro de una transacción serializada por
// el advisory lock del guard y comprueba DESPUÉS que sigue habiendo al menos un
// administrador activo (FR-008/P7). Si el recuento es 0, devuelve
// apperr.Conflict con `details.reason = "admin_required"` y la transacción se
// revierte (rollback). El lock hace imposible el estado prohibido incluso con
// dos administradores actuando a la vez (SC-004).
//
// El callback recibe GuardTx (interfaz), no el repositorio concreto: así el
// service no puede sacar la mutación de la transacción ni depender de pgx, y
// sus pruebas usan un fake. La inicialización única (FR-007) usa este mismo
// guard: el advisory lock serializa dos `Initialize` simultáneos (SC-003).
func (r *repository) WithAdminGuard(ctx context.Context, mutate func(tx GuardTx) error) error {
	return r.withTx(ctx, func(tx *repository) error {
		if err := tx.lockAdminGuard(ctx); err != nil {
			return err
		}
		if err := mutate(tx); err != nil {
			return err
		}
		return tx.ensureAdminRemains(ctx)
	})
}

// lockAdminGuard toma el advisory lock transaccional compartido con la
// inicialización única (FR-007). Se libera al cerrar la transacción.
func (r *repository) lockAdminGuard(ctx context.Context) error {
	if err := r.q.LockAdminGuard(ctx); err != nil {
		return fmt.Errorf("tomar cerradura anti-bloqueo: %w", err)
	}
	return nil
}

// ensureAdminRemains es el recuento post-mutación del guard (FR-008).
func (r *repository) ensureAdminRemains(ctx context.Context) error {
	total, err := r.CountActiveAdmins(ctx)
	if err != nil {
		return fmt.Errorf("contar administradores activos: %w", err)
	}
	if total == 0 {
		return apperr.Conflict(
			"No se puede dejar el panel sin administración",
			apperr.WithDetails(map[string]any{"reason": "admin_required"}),
		)
	}
	return nil
}
