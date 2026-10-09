//go:build integration

package testutil

import (
	"testing"

	"simiente-santa/backend/internal/platform/testutil/containers"
)

// RedisURL delega en el subpaquete containers. Se mantiene aquí para conservar
// la API pública de testutil (F1); el subpaquete existe para que las pruebas de
// paquetes que testutil importa (session) no creen un ciclo de imports.
func RedisURL(t *testing.T) string { return containers.RedisURL(t) }

// PostgresURL delega en el subpaquete containers (ver RedisURL).
func PostgresURL(t *testing.T) string { return containers.PostgresURL(t) }
