package proteccionzz

import "testing"

// Prueba deliberadamente roja: solo existe para verificar que la proteccion de
// rama de `main` bloquea el merge cuando un check requerido falla (T033).
// Esta rama y este archivo se descartan.
func TestFalloDeliberadoDeProteccion(t *testing.T) {
	t.Fatal("fallo deliberado para verificar la proteccion de rama (T033)")
}
