// Package password aplica la política de contraseñas FR-010 y el hash bcrypt
// con cost 12 (§IV, CWE-256; P5/R4).
//
// La política vive aquí y solo aquí: los tres flujos del dominio (contraseña
// inicial al crear la cuenta, restablecimiento por un administrador y cambio
// propio) comparten Validate. Hash la aplica antes de hashear, de modo que no
// se puede guardar una contraseña que no la cumpla. La contraseña nunca se
// registra ni se devuelve: de este paquete solo salen el hash y los errores.
package password

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"simiente-santa/backend/internal/platform/apperr"
)

const (
	// MinLength es la longitud mínima en caracteres de la política FR-010.
	MinLength = 8
	// MaxLength es la longitud máxima en caracteres (maxLength: 64 del contrato,
	// en todos los campos de contraseña).
	MaxLength = 64
	// bcryptCost es el coste acordado en P5/R4.
	bcryptCost = 12
	// bcryptMaxBytes es el límite duro de bcrypt (RG6): 72 bytes. Solo puede
	// alcanzarlo una contraseña de pocos caracteres y varios bytes por carácter.
	bcryptMaxBytes = 72
	// detailKey es la clave de Details con el mensaje de la política (el campo
	// del contrato al que corresponde).
	detailKey = "newPassword"
)

// Context son los datos de la cuenta con los que la contraseña no puede
// coincidir (FR-010). La comparación es de igualdad normalizada (trim +
// minúsculas): la contraseña puede contenerlos, lo que no puede es ser igual a
// ellos.
type Context struct {
	FirstName string
	LastName  string
	Email     string
}

// Validate aplica la política FR-010 y devuelve apperr.Invalid con
// Details["newPassword"] nombrando el requisito incumplido.
func Validate(value string, ctx Context) error {
	if equalsNormalized(value, ctx.FirstName) {
		return policyError("La contraseña no puede ser igual a tu nombre")
	}
	if equalsNormalized(value, ctx.LastName) {
		return policyError("La contraseña no puede ser igual a tus apellidos")
	}
	if equalsNormalized(value, ctx.Email) {
		return policyError("La contraseña no puede ser igual a tu correo")
	}

	length := utf8.RuneCountInString(value)
	if length < MinLength {
		return policyError(fmt.Sprintf("La contraseña debe tener al menos %d caracteres", MinLength))
	}
	if length > MaxLength {
		return policyError(fmt.Sprintf("La contraseña no puede tener más de %d caracteres", MaxLength))
	}
	if len(value) > bcryptMaxBytes {
		return policyError("La contraseña es demasiado larga para el almacenamiento seguro")
	}

	if !hasUpper(value) {
		return policyError("La contraseña debe incluir al menos una letra mayúscula")
	}
	if !hasLower(value) {
		return policyError("La contraseña debe incluir al menos una letra minúscula")
	}
	if !hasDigit(value) {
		return policyError("La contraseña debe incluir al menos un número")
	}
	if !hasSpecial(value) {
		return policyError("La contraseña debe incluir al menos un carácter especial")
	}
	return nil
}

// Hash valida la contraseña contra la política y devuelve su hash bcrypt con
// cost 12. Un hash distinto en cada llamada (sal aleatoria de bcrypt).
func Hash(value string, ctx Context) (string, error) {
	if err := Validate(value, ctx); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// Verify compara una contraseña con su hash bcrypt. Devuelve nil si coinciden y
// un error (que envuelve el de bcrypt) si no.
func Verify(hash, value string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(value)); err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}

// policyError construye el apperr.Invalid uniforme de la política, con el
// requisito incumplido en Details["newPassword"].
func policyError(message string) error {
	return apperr.Invalid(
		"La contraseña no cumple la política de seguridad",
		apperr.WithDetails(map[string]any{detailKey: message}),
	)
}

// equalsNormalized compara la contraseña (normalizada) con un dato de la
// cuenta. Un dato vacío nunca coincide.
func equalsNormalized(value, accountData string) bool {
	normalized := normalize(accountData)
	if normalized == "" {
		return false
	}
	return normalize(value) == normalized
}

// normalize aplica la comparación acordada (Q5/R4): trim + minúsculas.
func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func hasUpper(value string) bool { return hasFunc(value, unicode.IsUpper) }
func hasLower(value string) bool { return hasFunc(value, unicode.IsLower) }
func hasDigit(value string) bool { return hasFunc(value, unicode.IsDigit) }

// hasSpecial es true si hay algún carácter que no sea letra, dígito ni espacio.
func hasSpecial(value string) bool {
	return hasFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r)
	})
}

func hasFunc(value string, predicate func(rune) bool) bool {
	for _, r := range value {
		if predicate(r) {
			return true
		}
	}
	return false
}
