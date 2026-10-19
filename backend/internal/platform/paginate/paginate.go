// Package paginate normaliza los parámetros de paginación de los listados del
// panel (limit/offset) con el defecto y el tope acordados en P14 y §8.1.2/§8.1.3
// de arquitectura.md.
//
// No conoce ningún dominio (arq. R1): solo traduce los query params de entrada a
// unos límites seguros y devuelve apperr.Invalid con el detalle por campo cuando
// la entrada no sirve. El sobre de salida {items, total, limit, offset} lo
// construye cada handler; este paquete no lo arma.
package paginate

import (
	"math"
	"net/url"
	"strconv"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
)

const (
	// DefaultLimit es el tamaño de página por defecto (P14).
	DefaultLimit = 20
	// MaxLimit es el tope duro de filas por página (P14, §8.1.3).
	MaxLimit = 100
	// MaxOffset es el tope duro de `offset` (§8.1.3): más allá la página está
	// vacía de todos modos y el valor no cabe en el `int32` que esperan las
	// consultas (un `OFFSET` negativo es un error de PostgreSQL, no una página
	// vacía). Un offset mayor se **acota**, no es un error: la misma regla que
	// `limit` (el tope nunca es un error).
	MaxOffset = math.MaxInt32
)

// Params son los límites normalizados de un listado.
type Params struct {
	Limit  int
	Offset int
}

// FromValues lee limit y offset de unos query params ya parseados y los
// normaliza:
//
//   - ausentes o vacíos → limit DefaultLimit, offset 0;
//   - limit por encima de MaxLimit → MaxLimit (tope, no error);
//   - offset por encima de MaxOffset → MaxOffset (tope, no error: la página está
//     vacía igual y el valor no cabe en el `int32` de las consultas);
//   - limit <= 0, offset < 0 o valores no numéricos → apperr.Invalid con el
//     campo en Details.
func FromValues(values url.Values) (Params, error) {
	limit, err := parseLimit(values.Get("limit"))
	if err != nil {
		return Params{}, err
	}
	offset, err := parseOffset(values.Get("offset"))
	if err != nil {
		return Params{}, err
	}
	return Params{Limit: limit, Offset: offset}, nil
}

// parseLimit valida y acota limit.
func parseLimit(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultLimit, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, invalidParam("limit", "debe ser un número entero")
	}
	if limit <= 0 {
		return 0, invalidParam("limit", "debe ser mayor que 0")
	}
	if limit > MaxLimit {
		return MaxLimit, nil
	}
	return limit, nil
}

// parseOffset valida y acota offset: un valor por encima del tope se recorta
// (la página está vacía igual y así nunca desborda el `int32` de las consultas).
func parseOffset(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(raw)
	if err != nil {
		return 0, invalidParam("offset", "debe ser un número entero")
	}
	if offset < 0 {
		return 0, invalidParam("offset", "debe ser mayor o igual a 0")
	}
	if offset > MaxOffset {
		return MaxOffset, nil
	}
	return offset, nil
}

// invalidParam construye el apperr.Invalid uniforme de los parámetros de
// paginación, con el nombre del campo y el motivo que no se cumplió.
func invalidParam(field, reason string) error {
	return apperr.Invalid(
		"Parámetros de paginación inválidos",
		apperr.WithDetails(map[string]any{field: reason}),
	)
}
