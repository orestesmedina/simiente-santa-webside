package ministerios

import "encoding/json"

// Optional es un campo anulable de un PATCH. Distingue los tres estados que un
// `*T` no puede separar (encoding/json deja el puntero en nil tanto si la clave
// llegó como `null` como si no llegó):
//
//   - ausente: la clave no venía en el JSON → no se toca el valor actual;
//   - con valor: la clave venía con un valor → se aplica (normalizado);
//   - `null` explícito: la clave venía como `null` → se vacía el campo.
//
// Es exactamente la semántica que promete el contrato (`MinisterioPatch`: «un
// campo con `null` explícito borra su valor opcional, p. ej. `whatsappUrl`,
// `imageFile`, `leaderPhone`, `scheduleEs`») y la que exige A6-3/P6-20: el
// `PATCH` fusiona los campos **presentes** sobre la fila actual y los ausentes
// no cambian. Es la misma pieza que F3 (portada) y F4 (agenda) usan en sus
// dominios; se repite aquí (candidata a promover a `platform/` cuando aparezca
// un cuarto consumidor).
//
// El valor cero del tipo (Set=false) significa "ausente": encoding/json solo
// invoca UnmarshalJSON cuando la clave existe en el objeto. Los campos no
// anulables del contrato (`leaderVisible`, `sortOrder`, `publicationState`)
// siguen como `*T` porque `null` no es un valor válido para ellos.
type Optional[T any] struct {
	value T
	set   bool
	null  bool
}

// OptionalOf envuelve un valor como campo presente, para construir PATCHes en
// código (pruebas, llamadas internas) sin pasar por JSON.
func OptionalOf[T any](v T) Optional[T] {
	return Optional[T]{value: v, set: true}
}

// OptionalNull construye un campo presente con `null` explícito (el valor se
// vacía), para construir PATCHes en código sin pasar por JSON.
func OptionalNull[T any]() Optional[T] {
	return Optional[T]{set: true, null: true}
}

// UnmarshalJSON marca el campo como presente y recuerda si llegó como `null`.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.set = true
	if string(data) == "null" {
		o.null = true
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.value = v
	return nil
}

// Set indica si la clave estaba presente en el JSON (con valor o con `null`).
func (o Optional[T]) Set() bool { return o.set }

// Null indica si la clave llegó explícitamente como `null`.
func (o Optional[T]) Null() bool { return o.null }

// Value devuelve el valor transportado; el cero del tipo cuando la clave llegó
// como `null` (los normalizadores del service lo convierten en "vacío").
func (o Optional[T]) Value() T { return o.value }
