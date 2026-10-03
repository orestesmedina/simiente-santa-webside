// Package testutil reúne los helpers compartidos de las pruebas del backend:
// conexión a la base de datos de integración con skip automático, utilidades de
// httptest con la cadena completa y captura de logs en memoria.
//
// Se llama testutil (y no testing) para no chocar con el paquete testing de la
// stdlib en cada _test.go (D11). Es infraestructura de pruebas: no conoce
// dominios.
package testutil

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/database"
)

// DatabaseURLEnv es la variable de entorno con la cadena de conexión de la base
// de datos de pruebas.
const DatabaseURLEnv = "DATABASE_URL_TEST"

// DatabaseURL devuelve DATABASE_URL_TEST. Si no está definida, omite la prueba
// con t.Skip: `go test ./...` corre sin base de datos y las pruebas de
// integración se saltan solas.
func DatabaseURL(t *testing.T) string {
	t.Helper()
	url := strings.TrimSpace(os.Getenv(DatabaseURLEnv))
	if url == "" {
		t.Skipf("%s no está definida: se omite la prueba de integración", DatabaseURLEnv)
	}
	return url
}

// Pool construye un *pgxpool.Pool contra DATABASE_URL_TEST, omitiendo la prueba
// si no hay base de datos, y lo cierra al terminar (t.Cleanup). El pool es el
// perezoso de platform/database: no bloquea si la BD tarda.
func Pool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	url := DatabaseURL(t)
	pool, err := database.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("crear pool de pruebas: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
