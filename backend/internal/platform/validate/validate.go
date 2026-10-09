// Package validate valida los DTOs de entrada por reflexión sobre su etiqueta
// `validate` (P3 de F2).
//
// Es una implementación propia mínima, sin dependencias (D-A8): cubre las
// etiquetas que usan los DTOs —`required`, `omitempty`, `min`, `max`, `email`,
// `oneof` y `phone`— y devuelve un apperr.Invalid con un mensaje por campo en
// `details` (el formato que documenta el sobre de error de §5.11). La validación
// de negocio (política de contraseñas, normalización, existencia de un rol)
// vive en el service; aquí solo se comprueba la forma (CWE-20: toda entrada se
// valida en el backend aunque el frontend ya valide).
//
// Las reglas se separan por comas y los valores por espacios:
//
//	validate:"required,min=1,max=120"
//	validate:"omitempty,email,max=254"
//	validate:"required,phone,max=32"
//	validate:"required,oneof=success failure"
//	validate:"omitempty,url,max=500"
package validate

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"simiente-santa/backend/internal/platform/apperr"
)

// MessageInvalid es el mensaje del sobre para un DTO que no pasa la validación;
// el detalle de cada campo viaja en `details`.
const MessageInvalid = "Revisa los datos del formulario"

// minPhoneDigits es el mínimo de dígitos que exige FR-009 para un teléfono.
const minPhoneDigits = 7

// maxURLLength acota un enlace válido (R3-14): 500 caracteres, el límite del
// contrato para las URLs de WhatsApp, redes y (F4–F9) eventos.
const maxURLLength = 500

var (
	// emailRe exige algo@algo.algo: sin espacios y con un punto en el dominio.
	emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	// phoneRe admite dígitos y los separadores habituales (espacios, guiones y
	// paréntesis) con un prefijo internacional opcional al principio.
	phoneRe = regexp.MustCompile(`^\+?[0-9() -]+$`)
)

// Struct valida el DTO indicado (un struct o un puntero a struct). Devuelve un
// *apperr.Error de kind Invalid con un mensaje por campo en `details`, o nil si
// el DTO es válido. Un valor que no es struct no produce error.
func Struct(dto any) error {
	value, ok := derefStruct(dto)
	if !ok {
		return nil
	}

	typ := value.Type()
	details := make(map[string]any)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" { // campo no exportado
			continue
		}
		tag := field.Tag.Get("validate")
		if tag == "" || tag == "-" {
			continue
		}
		if msg := validateField(value.Field(i), parseRules(tag)); msg != "" {
			details[jsonName(field)] = msg
		}
	}

	if len(details) == 0 {
		return nil
	}
	return apperr.Invalid(MessageInvalid, apperr.WithDetails(details))
}

// rule es una etiqueta `validate` ya separada en nombre y valor.
type rule struct {
	name  string
	value string
}

func parseRules(tag string) []rule {
	parts := strings.Split(tag, ",")
	rules := make([]rule, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, value, found := strings.Cut(part, "=")
		if !found {
			rules = append(rules, rule{name: name})
			continue
		}
		rules = append(rules, rule{name: name, value: value})
	}
	return rules
}

// validateField aplica las reglas del campo y devuelve el primer mensaje de
// error (uno por campo, para no abrumar con varios avisos de la misma casilla).
func validateField(value reflect.Value, rules []rule) string {
	if hasRule(rules, "omitempty") && isZero(value) {
		return ""
	}

	for _, r := range rules {
		switch r.name {
		case "omitempty":
			// Modificador: ya aplicado arriba.
		case "required":
			if isZero(value) {
				return "Este campo es obligatorio."
			}
		case "min":
			if n, err := strconv.Atoi(r.value); err == nil {
				if msg := checkMin(value, n); msg != "" {
					return msg
				}
			}
		case "max":
			if n, err := strconv.Atoi(r.value); err == nil {
				if msg := checkMax(value, n); msg != "" {
					return msg
				}
			}
		case "email":
			if value.Kind() == reflect.String {
				email := strings.TrimSpace(value.String())
				if email != "" && !emailRe.MatchString(email) {
					return "Escribe un correo con este formato: nombre@dominio.com."
				}
			}
		case "oneof":
			if !matchesOneOf(value, r.value) {
				return fmt.Sprintf("Elige una de las opciones válidas: %s.", strings.Join(strings.Fields(r.value), ", "))
			}
		case "phone":
			if value.Kind() == reflect.String {
				phone := strings.TrimSpace(value.String())
				if phone != "" && !validPhone(phone) {
					return "Escribe un número de teléfono válido (al menos 7 dígitos, con espacios, guiones o paréntesis)."
				}
			}
		case "url":
			if value.Kind() == reflect.String {
				link := strings.TrimSpace(value.String())
				if link != "" {
					if len(link) > maxURLLength {
						return fmt.Sprintf("No puede tener más de %d caracteres.", maxURLLength)
					}
					if !validURL(link) {
						return "Escribe un enlace válido que empiece por https://."
					}
				}
			}
		}
	}
	return ""
}

func hasRule(rules []rule, name string) bool {
	for _, r := range rules {
		if r.name == name {
			return true
		}
	}
	return false
}

// derefStruct normaliza el DTO a un reflect.Value de struct. Devuelve false si
// es nil, un puntero nil o algo que no es struct.
func derefStruct(dto any) (reflect.Value, bool) {
	if dto == nil {
		return reflect.Value{}, false
	}
	value := reflect.ValueOf(dto)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	return value, true
}

// isZero decide si un campo cuenta como ausente. Los textos en blanco cuentan
// como vacíos (un nombre con solo espacios no es un nombre).
func isZero(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return strings.TrimSpace(value.String()) == ""
	case reflect.Pointer, reflect.Interface:
		return value.IsNil()
	case reflect.Slice, reflect.Map:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	default:
		return value.IsZero()
	}
}

func checkMin(value reflect.Value, n int) string {
	switch {
	case value.Kind() == reflect.String:
		if utf8.RuneCountInString(value.String()) < n {
			return fmt.Sprintf("Debe tener al menos %d caracteres.", n)
		}
	case isInt(value.Kind()):
		if value.Int() < int64(n) {
			return fmt.Sprintf("El valor debe ser %d o mayor.", n)
		}
	}
	return ""
}

func checkMax(value reflect.Value, n int) string {
	switch {
	case value.Kind() == reflect.String:
		if utf8.RuneCountInString(value.String()) > n {
			return fmt.Sprintf("No puede tener más de %d caracteres.", n)
		}
	case isInt(value.Kind()):
		if value.Int() > int64(n) {
			return fmt.Sprintf("El valor debe ser %d o menor.", n)
		}
	}
	return ""
}

func matchesOneOf(value reflect.Value, options string) bool {
	var candidate string
	switch {
	case value.Kind() == reflect.String:
		candidate = value.String()
	case isInt(value.Kind()):
		candidate = strconv.FormatInt(value.Int(), 10)
	default:
		return true
	}
	for _, option := range strings.Fields(options) {
		if candidate == option {
			return true
		}
	}
	return false
}

// validPhone comprueba FR-009: separadores habituales, prefijo internacional
// opcional y al menos 7 dígitos. No se normaliza la forma de presentación.
func validPhone(phone string) bool {
	if !phoneRe.MatchString(phone) {
		return false
	}
	digits := 0
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits++
		}
	}
	return digits >= minPhoneDigits
}

// validURL comprueba R3-14: solo esquema `https` con host no vacío (vía
// net/url). Rechaza `http://`, `javascript:`, `data:` y cadenas sin host. El
// host concreto (WhatsApp, red social) lo comprueba el service del dominio, que
// es quien conoce el negocio (R3-6/R3-7).
func validURL(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Scheme == "https" && u.Host != ""
}

func isInt(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

// jsonName devuelve el nombre del campo tal y como viaja en el JSON, que es la
// clave que el frontend espera en `details`.
func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return field.Name
	}
	return name
}
