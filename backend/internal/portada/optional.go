package portada

import "encoding/json"

// Optional es un campo anulable de un PATCH. Distingue los tres estados que un
// `*T` no puede separar (encoding/json deja el puntero en nil tanto si la clave
// llegó como `null` como si no llegó):
//
//   - ausente: la clave no venía en el JSON → no se toca el valor actual;
//   - con valor: la clave venía con un valor → se aplica (normalizado);
//   - `null` explícito: la clave venía como `null` → se vacía el campo.
//
// Es exactamente la semántica que promete el contrato (ScheduleItemPatch y
// WhatsappChannelPatch): «un campo `*En` con `null` o `""`/espacios vacía su
// traducción al inglés; `endTime: null` quita la hora de fin» (analyze I6).
//
// El valor cero del tipo (Set=false) significa "ausente": encoding/json solo
// invoca UnmarshalJSON cuando la clave existe en el objeto, así que el campo no
// se toca si no se envió. Los campos no anulables del contrato siguen como `*T`
// porque `null` no es un valor válido para ellos.
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
