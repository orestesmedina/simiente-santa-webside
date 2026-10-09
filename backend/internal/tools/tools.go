//go:build tools

// Package tools fija las dependencias de F2 que todavía no importa ningún
// paquete de producción o prueba (llegan en T215 `platform/password`,
// T218 `platform/session`, T219 `platform/testutil` y T220 dominio `usuarios`).
//
// `go mod tidy` elimina cualquier módulo que ningún archivo importe, pero
// considera todos los build tags al calcular el conjunto de imports. Con la
// restricción `//go:build tools` este archivo nunca entra en un build real
// (`go build ./...` lo ignora) y, a la vez, `go mod tidy` conserva los cuatro
// módulos justificados en research.md R16 / plan.md D-A8. Cuando las tareas
// posteriores importen estos paquetes de verdad, el archivo puede retirarse sin
// ningún otro cambio.
package tools

import (
	_ "github.com/google/uuid"
	_ "github.com/redis/go-redis/v9"
	_ "github.com/testcontainers/testcontainers-go"
	_ "golang.org/x/crypto/bcrypt"
)
