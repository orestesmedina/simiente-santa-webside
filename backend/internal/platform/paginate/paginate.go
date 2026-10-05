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

// parseOffset valida offset.
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
