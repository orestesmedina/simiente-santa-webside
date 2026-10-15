// Package donaciones implementa el dominio de la sección «Ofrendas y diezmos»
// de F7: los métodos de donación (cuentas IBAN y números de SINPE Móvil) y los
// dos textos libres de la sección.
//
// Este archivo (numero.go) contiene las funciones PURAS de validación y
// normalización de números (research R7-2/R7-3, P7-2/P7-3/P7-4): no dependen de
// la base de datos ni de HTTP y son la ÚNICA implementación de las reglas de
// número del dominio. El dato normalizado es el que se guarda (FR-015/FR-016).
package donaciones

import (
	"errors"
	"regexp"
	"strings"
)

// ErrNumeroInvalido es el error centinela de un número de donación que no cumple
// su formato. El service lo traduce a `400 invalid` con `details.number`.
var ErrNumeroInvalido = errors.New("donaciones: número de donación inválido")

// ibanEstructura valida la forma general de un IBAN: 2 letras de país, 2 dígitos
// de control y un BBAN alfanumérico. La longitud exacta por país se comprueba
// aparte contra el registro embebido (R7-2).
var ibanEstructura = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]+$`)

// sinpeDigitos valida que el número SINPE, ya sin prefijo ni separadores, sea
// solo dígitos.
var sinpeDigitos = regexp.MustCompile(`^[0-9]+$`)

// ibanLengths es el registro IBAN embebido (longitud exacta por código de país).
// Se usa la estructura y la longitud, NUNCA mod-97 (default revisable de
// research R7-2, pregunta #3 al humano). Fuente: registro IBAN ISO 13616.
var ibanLengths = map[string]int{
	"AD": 24, "AE": 23, "AL": 28, "AT": 20, "AZ": 28,
	"BA": 20, "BE": 16, "BG": 22, "BH": 22, "BR": 29,
	"CH": 21, "CR": 22, "CY": 28, "CZ": 24,
	"DE": 22, "DK": 18, "DO": 28,
	"EE": 20, "ES": 24,
	"FI": 18, "FO": 18, "FR": 27,
	"GB": 22, "GE": 22, "GI": 23, "GL": 18, "GR": 27, "GT": 28,
	"HR": 21, "HU": 28,
	"IE": 22, "IL": 23, "IS": 26, "IT": 27,
	"JO": 30,
	"KZ": 20, "KW": 30, "XK": 20,
	"LB": 28, "LC": 32, "LI": 21, "LT": 20, "LU": 20, "LV": 21,
	"MC": 27, "MD": 24, "ME": 22, "MK": 19, "MR": 27, "MT": 31, "MU": 30,
	"NL": 18, "NO": 15,
	"PK": 24, "PL": 28, "PS": 29, "PT": 25,
	"QA": 29,
	"RO": 24, "RS": 22,
	"SA": 24, "SC": 31, "SE": 24, "SI": 19, "SK": 24, "SM": 27, "ST": 25,
	"TL": 23, "TN": 24, "TR": 26,
	"UA": 29, "VA": 22, "VG": 24,
}

// NormalizarIban devuelve el IBAN normalizado —solo los caracteres del número,
// en mayúsculas, sin espacios ni guiones— o ErrNumeroInvalido si su estructura
// o su longitud por país no son válidas (FR-015/SC-009, R7-2).
func NormalizarIban(raw string) (string, error) {
	normalizado := strings.ToUpper(quitarSeparadores(raw))
	if !ibanEstructura.MatchString(normalizado) {
		return "", ErrNumeroInvalido
	}
	longitud, ok := ibanLengths[normalizado[:2]]
	if !ok || len(normalizado) != longitud {
		return "", ErrNumeroInvalido
	}
	return normalizado, nil
}

// ValidarIban comprueba que raw es un IBAN válido (misma regla que
// NormalizarIban; una sola implementación pura).
func ValidarIban(raw string) error {
	_, err := NormalizarIban(raw)
	return err
}

// NormalizarSinpe devuelve el número SINPE Móvil normalizado `+506XXXXXXXX` o
// ErrNumeroInvalido si no es un celular costarricense de 8 dígitos con primer
// dígito 6/7/8 (FR-016/SC-009, R7-3). Acepta el número con o sin `+506` y con
// espacios o guiones.
func NormalizarSinpe(raw string) (string, error) {
	normalizado := quitarSeparadores(raw)
	digitos := ""
	switch {
	case strings.HasPrefix(normalizado, "+506"):
		digitos = normalizado[len("+506"):]
	case strings.HasPrefix(normalizado, "506") && len(normalizado) == 11:
		digitos = normalizado[len("506"):]
	default:
		digitos = normalizado
	}
	if len(digitos) != 8 || !sinpeDigitos.MatchString(digitos) {
		return "", ErrNumeroInvalido
	}
	if inicio := digitos[0]; inicio < '6' || inicio > '8' {
		return "", ErrNumeroInvalido
	}
	return "+506" + digitos, nil
}

// ValidarSinpe comprueba que raw es un SINPE Móvil válido (misma regla que
// NormalizarSinpe).
func ValidarSinpe(raw string) error {
	_, err := NormalizarSinpe(raw)
	return err
}

// quitarSeparadores elimina los separadores de lectura admitidos —espacios
// (incluido el no separable) y guiones— y recorta los extremos. No elimina
// ningún otro carácter: un emoji o una letra donde van dígitos se rechazan
// después por estructura.
func quitarSeparadores(raw string) string {
	replacer := strings.NewReplacer(
		" ", "",
		"\t", "",
		"\u00a0", "",
		"-", "",
		"\u2013", "",
		"\u2014", "",
	)
	return strings.TrimSpace(replacer.Replace(raw))
}
