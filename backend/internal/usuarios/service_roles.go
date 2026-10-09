package usuarios

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/validate"
)

// service_roles.go implementa la gestión de roles y el catálogo de permisos del
// panel (FR-014…FR-018, US4/US6): listar y ver roles, crearlos con permisos por
// módulo (combinables libremente, ≥ 1 permiso y nombre único normalizado),
// editarlos (los cambios se reflejan de inmediato porque los permisos se
// resuelven por petición) y eliminarlos solo cuando ninguna cuenta los usa. No
// conoce HTTP ni SQL: depende de un puerto del repositorio (arq. R3) y del
// registro de auditoría (T224/P20).
//
// Invariantes que garantiza este service (data-model.md):
//   - Todo rol conserva AL MENOS un permiso (FR-014): las mutaciones reemplazan
//     el conjunto completo (DELETE + INSERT) y rechazan el vacío antes de tocar
//     la base.
//   - Un rol solo se elimina si CountRoleUsers = 0 (FR-017); la FK
//     `users.role_id ON DELETE RESTRICT` es la red de seguridad ante carreras.
//   - Editar/eliminar un rol pasa por el guard anti-bloqueo (FR-008/P7): quitar
//     `admin_usuarios_roles` al único rol que lo tiene deja el panel sin
//     administración → 409 y rollback.
//
// La creación usa una transacción SIN guard (WithTx): crear un rol no puede
// retirar acceso, así que no necesita el advisory lock ni exige que ya exista un
// administrador. El registro `role.create` va en la MISMA transacción que el rol
// y sus permisos (FR-023).

// Mensajes seguros para el cliente (contrato OpenAPI y ux.md).
const (
	// messageRoleNeedsPermission es el 400 de un rol sin permisos (FR-014/US4
	// esc. 4).
	messageRoleNeedsPermission = "un rol debe tener al menos un permiso"
	// messageRoleNameInUse es el 409 del nombre duplicado normalizado
	// (Q5/SC-011): también cuando solo difiere en mayúsculas o espacios.
	messageRoleNameInUse = "Ya existe un rol con ese nombre"
	// messageRoleInUse es el 409 al eliminar un rol con cuentas asignadas
	// (FR-017/US6 esc. 5).
	messageRoleInUse = "No se puede eliminar un rol con cuentas asignadas; reasigna esas cuentas primero"
	// messageUnknownPermission es el 400 de un permiso que no está en el
	// catálogo (FR-015).
	messageUnknownPermission = "Alguno de los permisos no existe en el catálogo"
)

// RoleRepository es el puerto de datos que la gestión de roles necesita del
// repositorio (lo define quien lo consume, arq. R3). Los errores de PostgreSQL
// llegan ya traducidos a apperr.
type RoleRepository interface {
	GetRoleByID(ctx context.Context, id uuid.UUID) (Role, error)
	GetRoleByNameLower(ctx context.Context, name string) (Role, error)
	ListRoles(ctx context.Context, params paginate.Params) ([]Role, error)
	CountRoles(ctx context.Context) (int64, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetPermissionIDsByCodes(ctx context.Context, codes []string) (map[string]uuid.UUID, error)
	CountRoleUsers(ctx context.Context, roleID uuid.UUID) (int64, error)
	// WithTx ejecuta la creación de un rol y su registro en una transacción
	// (sin guard anti-bloqueo: crear no puede dejar el panel sin admin).
	WithTx(ctx context.Context, mutate func(tx GuardTx) error) error
	// WithAdminGuard ejecuta la edición/eliminación con su registro y el guard
	// anti-bloqueo (FR-008/P7).
	WithAdminGuard(ctx context.Context, mutate func(tx GuardTx) error) error
}

// RoleServiceDeps agrupa las dependencias de la gestión de roles.
type RoleServiceDeps struct {
	// Repository es el acceso a datos del dominio (arq. R3).
	Repository RoleRepository
	// Audit registra best-effort los desenlaces fallidos (P20). Puede ser nil.
	Audit ActionRecorder
}

// roleService implementa la gestión de roles y el catálogo de permisos.
type roleService struct {
	repository RoleRepository
	audit      ActionRecorder
}

// NewRoleService construye el servicio de gestión de roles.
func NewRoleService(deps RoleServiceDeps) *roleService {
	return &roleService{repository: deps.Repository, audit: deps.Audit}
}

// ListRoles devuelve una página de roles con sus permisos y el recuento de
// cuentas que los usan (FR-017). Los límites ya vienen normalizados por
// platform/paginate (P14).
func (s *roleService) ListRoles(ctx context.Context, params paginate.Params) (RoleList, error) {
	roles, err := s.repository.ListRoles(ctx, params)
	if err != nil {
		return RoleList{}, fmt.Errorf("listar roles: %w", err)
	}
	total, err := s.repository.CountRoles(ctx)
	if err != nil {
		return RoleList{}, fmt.Errorf("contar roles: %w", err)
	}
	items := make([]RoleItem, 0, len(roles))
	for _, role := range roles {
		items = append(items, RoleItemFrom(role))
	}
	return RoleList{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

// GetRole devuelve la ficha de un rol con sus permisos y el recuento de cuentas
// (FR-017). Inexistente → 404.
func (s *roleService) GetRole(ctx context.Context, id uuid.UUID) (RoleItem, error) {
	role, err := s.repository.GetRoleByID(ctx, id)
	if err != nil {
		return RoleItem{}, err
	}
	return RoleItemFrom(role), nil
}

// CreateRole crea un rol con permisos por módulo combinables libremente
// (FR-014/US4): normaliza el nombre (trim + colapso de espacios), rechaza
// duplicados sin distinguir mayúsculas (409/Q5), exige al menos un permiso
// (400) y comprueba que todos existan en el catálogo (FR-015). En una única
// transacción crea el rol, asocia sus permisos y registra `role.create` con
// actor y objetivo (FR-023).
func (s *roleService) CreateRole(ctx context.Context, actorID uuid.UUID, in RoleCreateInput) (RoleItem, error) {
	if err := validate.Struct(in); err != nil {
		s.recordFailure(ctx, actorID, audit.ActionRoleCreate, nil, "")
		return RoleItem{}, err
	}
	name := normalizeRoleName(in.Name)

	codes, err := rolePermissionCodes(in.Permissions)
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionRoleCreate, nil, "")
		return RoleItem{}, err
	}

	ids, err := s.resolvePermissionIDs(ctx, codes)
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionRoleCreate, nil, "")
		return RoleItem{}, err
	}

	if err := s.ensureRoleNameAvailable(ctx, name, uuid.Nil); err != nil {
		s.recordFailure(ctx, actorID, audit.ActionRoleCreate, nil, name)
		return RoleItem{}, err
	}

	var created Role
	err = s.repository.WithTx(ctx, func(tx GuardTx) error {
		role, err := tx.InsertRole(ctx, name)
		if err != nil {
			return err
		}
		for _, code := range codes {
			if err := tx.InsertRolePermission(ctx, role.ID, ids[code]); err != nil {
				return fmt.Errorf("asociar el permiso %q: %w", code, err)
			}
		}
		if _, err := tx.InsertAdminAction(ctx, audit.Action{
			ActorUserID:  &actorID,
			Code:         audit.ActionRoleCreate,
			TargetKind:   audit.TargetRole,
			TargetRoleID: &role.ID,
			TargetLabel:  role.Name,
			Result:       audit.ResultSuccess,
		}); err != nil {
			return fmt.Errorf("registrar la creación del rol: %w", err)
		}
		created = role
		return nil
	})
	if err != nil {
		err = normalizeRoleConflict(err)
		s.recordFailure(ctx, actorID, audit.ActionRoleCreate, nil, name)
		return RoleItem{}, err
	}
	created.Permissions = codes
	return RoleItemFrom(created), nil
}

// UpdateRole edita el nombre y/o los permisos de un rol (FR-017/FR-018/US6).
// Cambiar el nombre conserva las cuentas y sus permisos; reemplazar los permisos
// se hace con DELETE + INSERT del conjunto completo y nunca deja el rol sin
// ninguno (400). El cambio se refleja de inmediato en las cuentas que tienen el
// rol porque sus permisos se resuelven por petición (SC-009). La mutación y su
// registro `role.update` van en la misma transacción con el guard anti-bloqueo
// (FR-008): quitar `admin_usuarios_roles` al único rol que lo tiene → 409.
func (s *roleService) UpdateRole(ctx context.Context, actorID uuid.UUID, id uuid.UUID, in RoleUpdateInput) (RoleItem, error) {
	if err := validate.Struct(in); err != nil {
		s.recordFailure(ctx, actorID, audit.ActionRoleUpdate, &id, "")
		return RoleItem{}, err
	}
	nameProvided := strings.TrimSpace(in.Name) != ""
	permissionsProvided := in.Permissions != nil
	if !nameProvided && !permissionsProvided {
		s.recordFailure(ctx, actorID, audit.ActionRoleUpdate, &id, "")
		return RoleItem{}, apperr.Invalid(messageNoChanges)
	}

	current, err := s.repository.GetRoleByID(ctx, id)
	if err != nil {
		// El rol no existe: no se referencia como objetivo (la FK de
		// admin_actions lo rechazaría); el fallo se registra igual.
		s.recordFailure(ctx, actorID, audit.ActionRoleUpdate, nil, "")
		return RoleItem{}, err
	}

	name := current.Name
	if nameProvided {
		name = normalizeRoleName(in.Name)
	}

	permissions := current.Permissions
	var ids map[string]uuid.UUID
	if permissionsProvided {
		codes, err := rolePermissionCodes(in.Permissions)
		if err != nil {
			s.recordFailure(ctx, actorID, audit.ActionRoleUpdate, &id, current.Name)
			return RoleItem{}, err
		}
		ids, err = s.resolvePermissionIDs(ctx, codes)
		if err != nil {
			s.recordFailure(ctx, actorID, audit.ActionRoleUpdate, &id, current.Name)
			return RoleItem{}, err
		}
		permissions = codes
	}

	if nameProvided && !strings.EqualFold(name, current.Name) {
		if err := s.ensureRoleNameAvailable(ctx, name, current.ID); err != nil {
			s.recordFailure(ctx, actorID, audit.ActionRoleUpdate, &id, current.Name)
			return RoleItem{}, err
		}
	}

	err = s.repository.WithAdminGuard(ctx, func(tx GuardTx) error {
		if nameProvided {
			if _, err := tx.UpdateRoleName(ctx, id, name); err != nil {
				return err
			}
		}
		if permissionsProvided {
			if err := tx.DeleteRolePermissions(ctx, id); err != nil {
				return err
			}
			for _, code := range permissions {
				if err := tx.InsertRolePermission(ctx, id, ids[code]); err != nil {
					return fmt.Errorf("asociar el permiso %q: %w", code, err)
				}
			}
		}
		if _, err := tx.InsertAdminAction(ctx, audit.Action{
			ActorUserID:  &actorID,
			Code:         audit.ActionRoleUpdate,
			TargetKind:   audit.TargetRole,
			TargetRoleID: &current.ID,
			TargetLabel:  name,
			Result:       audit.ResultSuccess,
		}); err != nil {
			return fmt.Errorf("registrar la edición del rol: %w", err)
		}
		return nil
	})
	if err != nil {
		err = normalizeRoleConflict(err)
		s.recordFailure(ctx, actorID, audit.ActionRoleUpdate, &id, current.Name)
		return RoleItem{}, err
	}

	current.Name = name
	current.Permissions = permissions
	return RoleItemFrom(current), nil
}

// DeleteRole elimina un rol solo cuando ninguna cuenta lo tiene asignado
// (FR-017/US6 esc. 4–5). Si hay cuentas → 409 explicando que primero deben
// reasignarse; la FK ON DELETE RESTRICT es la red de seguridad ante carreras. La
// eliminación y su registro `role.delete` van en la misma transacción con el
// guard anti-bloqueo (FR-008). En el registro `target_role_id` queda NULL (la FK
// lo anula al borrar el rol) y `target_label` conserva el nombre (FR-025).
func (s *roleService) DeleteRole(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, err := s.repository.GetRoleByID(ctx, id)
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionRoleDelete, nil, "")
		return err
	}

	count, err := s.repository.CountRoleUsers(ctx, id)
	if err != nil {
		s.recordFailure(ctx, actorID, audit.ActionRoleDelete, &id, current.Name)
		return fmt.Errorf("contar las cuentas del rol: %w", err)
	}
	if count > 0 {
		s.recordFailure(ctx, actorID, audit.ActionRoleDelete, &id, current.Name)
		return apperr.Conflict(
			messageRoleInUse,
			apperr.WithDetails(map[string]any{"userCount": count}),
		)
	}

	err = s.repository.WithAdminGuard(ctx, func(tx GuardTx) error {
		// La acción se registra ANTES de borrar el rol (la FK lo exige: no se
		// puede referenciar un rol que ya no existe). El `ON DELETE SET NULL`
		// deja su target_role_id en NULL y conserva target_label (FR-025).
		if _, err := tx.InsertAdminAction(ctx, audit.Action{
			ActorUserID:  &actorID,
			Code:         audit.ActionRoleDelete,
			TargetKind:   audit.TargetRole,
			TargetRoleID: &current.ID,
			TargetLabel:  current.Name,
			Result:       audit.ResultSuccess,
		}); err != nil {
			return fmt.Errorf("registrar la eliminación del rol: %w", err)
		}
		if _, err := tx.DeleteRole(ctx, id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		// Se conserva el mensaje del conflicto de FK (rol en uso por una
		// carrera), no el del nombre duplicado.
		s.recordFailure(ctx, actorID, audit.ActionRoleDelete, &id, current.Name)
		return err
	}
	return nil
}

// ListPermissions devuelve el catálogo fijo de permisos por módulo (FR-015) con
// su etiqueta. Los módulos de F3–F9 aparecen reservados: existen en el catálogo
// aunque aún no den acceso a ninguna pantalla.
func (s *roleService) ListPermissions(ctx context.Context) (PermissionList, error) {
	permissions, err := s.repository.ListPermissions(ctx)
	if err != nil {
		return PermissionList{}, fmt.Errorf("listar permisos: %w", err)
	}
	items := make([]PermissionItem, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, PermissionItemFrom(permission))
	}
	return PermissionList{Items: items}, nil
}

// resolvePermissionIDs traduce los códigos a ids y falla con 400 + details si
// alguno no pertenece al catálogo (FR-015).
func (s *roleService) resolvePermissionIDs(ctx context.Context, codes []string) (map[string]uuid.UUID, error) {
	ids, err := s.repository.GetPermissionIDsByCodes(ctx, codes)
	if err != nil {
		return nil, fmt.Errorf("resolver los permisos: %w", err)
	}
	var unknown []string
	for _, code := range codes {
		if _, ok := ids[code]; !ok {
			unknown = append(unknown, code)
		}
	}
	if len(unknown) > 0 {
		return nil, apperr.Invalid(
			messageUnknownPermission,
			apperr.WithDetails(map[string]any{"permissions": unknown}),
		)
	}
	return ids, nil
}

// ensureRoleNameAvailable comprueba que el nombre normalizado no esté en uso por
// OTRO rol (exceptID). Devuelve apperr.Conflict con el mensaje del contrato y
// nil si está libre.
func (s *roleService) ensureRoleNameAvailable(ctx context.Context, name string, exceptID uuid.UUID) error {
	existing, err := s.repository.GetRoleByNameLower(ctx, name)
	if err != nil {
		if isNotFound(err) {
			return nil // libre
		}
		return fmt.Errorf("comprobar el nombre del rol: %w", err)
	}
	if existing.ID != exceptID {
		return apperr.Conflict(messageRoleNameInUse)
	}
	return nil
}

// recordFailure deja la fila `result='failure'` de una operación de roles que no
// se completó (P20). Best-effort: nunca cambia la respuesta.
func (s *roleService) recordFailure(ctx context.Context, actorID uuid.UUID, code string, roleID *uuid.UUID, label string) {
	if s.audit == nil {
		return
	}
	action := audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetRole,
		TargetLabel: label,
		Result:      audit.ResultFailure,
	}
	if roleID != nil {
		action.TargetRoleID = roleID
	}
	s.audit.RecordActionBestEffort(ctx, action)
}

// normalizeRoleName aplica la normalización de Q5 al nombre de un rol: recorta
// los espacios de los extremos y colapsa los interiores a uno solo, conservando
// las mayúsculas de presentación. La unicidad la comprueba la base con
// lower(name).
func normalizeRoleName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

// rolePermissionCodes prepara la lista de códigos de permiso de un rol: recorta
// cada uno, descarta repetidos y exige al menos uno (FR-014). Un conjunto vacío
// → 400 con el mensaje de la spec.
func rolePermissionCodes(raw []string) ([]string, error) {
	seen := make(map[string]struct{}, len(raw))
	codes := make([]string, 0, len(raw))
	for _, value := range raw {
		code := strings.TrimSpace(value)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, apperr.Invalid(messageRoleNeedsPermission)
	}
	return codes, nil
}

// normalizeRoleConflict traduce el 409 de un nombre duplicado al mensaje del
// contrato, SIN tocar el 409 del guard anti-bloqueo (FR-008), que lleva
// details.reason="admin_required" y debe propagarse tal cual.
func normalizeRoleConflict(err error) error {
	if !isConflict(err) || isAdminRequired(err) {
		return err
	}
	return apperr.Conflict(messageRoleNameInUse, apperr.WithCause(err))
}

// roleService implementa el puerto que publican los handlers de roles.
var _ RolesService = (*roleService)(nil)
