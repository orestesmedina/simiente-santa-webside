package session

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// Identity es la identidad resuelta de la cuenta autenticada para la petición
// en curso (R3). La construye el dominio usuarios con los permisos del rol
// vigentes (FR-018/SC-009) y solo si la cuenta sigue activa (FR-012); el
// middleware authn la deja en el contexto de la petición y authz la consulta.
type Identity struct {
	UserID             uuid.UUID
	Email              string
	FirstName          string
	LastName           string
	RoleID             uuid.UUID
	RoleName           string
	Permissions        []string
	MustChangePassword bool
}

// HasPermission indica si el rol de la identidad incluye el módulo code.
func (i Identity) HasPermission(code string) bool {
	for _, permission := range i.Permissions {
		if permission == code {
			return true
		}
	}
	return false
}

// FullName devuelve nombre y apellidos separados por un espacio, sin espacios
// sobrantes cuando alguno falta.
func (i Identity) FullName() string {
	return strings.TrimSpace(strings.TrimSpace(i.FirstName) + " " + strings.TrimSpace(i.LastName))
}

// Resolver resuelve la identidad de una cuenta a partir de su id, por petición,
// para que los cambios de rol, permisos y estado se reflejen de inmediato
// (FR-018/SC-009). La implementa el dominio usuarios (T225); platform no conoce
// dominios (R1).
type Resolver interface {
	Resolve(ctx context.Context, userID uuid.UUID) (Identity, error)
}

// identityContextKey es la clave privada del contexto de petición donde authn
// deja la Identity resuelta.
type identityContextKey struct{}

// ContextWithIdentity guarda la identidad resuelta en el contexto de la
// petición.
func ContextWithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, identity)
}

// IdentityFromContext devuelve la identidad resuelta, o false si la petición no
// está autenticada.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(Identity)
	return identity, ok
}
