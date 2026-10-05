//go:build integration

package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Imágenes de los servicios que levantan las pruebas de integración (R19). Se
// fijan las versiones para que el arranque sea reproducible.
const (
	redisImage    = "redis:7-alpine"
	postgresImage = "postgres:16-alpine"
)

// startupTimeout acota el arranque de un contenedor para que una prueba no se
// quede colgada si la imagen no llega.
const startupTimeout = 3 * time.Minute

// requireDocker omite la prueba con un mensaje claro si no hay un demonio de
// Docker accesible. Los runners de GitHub Actions y `make up` ya lo exigen; en
// un entorno sin Docker la integración se salta, nunca falla.
func requireDocker(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		t.Skipf("Docker no disponible (%v): se omite la prueba de integración", err)
	}
	// Health cierra el provider internamente, así que no se cierra aquí.
	if err := provider.Health(ctx); err != nil {
		t.Skipf("Docker no disponible (%v): se omite la prueba de integración", err)
	}
}

// startContainer levanta un contenedor genérico esperando su estrategia de
// arranque y lo destruye al terminar la prueba (t.Cleanup).
func startContainer(t *testing.T, request testcontainers.ContainerRequest) testcontainers.Container {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("levantar contenedor %s: %v", request.Image, err)
	}
	t.Cleanup(func() {
		terminateCtx, terminateCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer terminateCancel()
		if err := container.Terminate(terminateCtx); err != nil {
			t.Logf("terminar contenedor %s: %v", request.Image, err)
		}
	})
	return container
}

// RedisURL levanta un Redis 7 efímero (redis:7-alpine) y devuelve su URL
// (redis://host:puerto/0). Si no hay Docker, omite la prueba.
func RedisURL(t *testing.T) string {
	t.Helper()
	requireDocker(t)

	container := startContainer(t, testcontainers.ContainerRequest{
		Image:        redisImage,
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	})

	ctx := context.Background()
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host del contenedor Redis: %v", err)
	}
	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("puerto mapeado del contenedor Redis: %v", err)
	}
	return fmt.Sprintf("redis://%s:%s/0", host, port.Port())
}

// PostgresURL devuelve la cadena de conexión de PostgreSQL para las pruebas: si
// DATABASE_URL_TEST está definida la reutiliza (es lo que aporta el CI del kit);
// si no, levanta un PostgreSQL 16 efímero. Si no hay Docker ni la variable,
// omite la prueba.
func PostgresURL(t *testing.T) string {
	t.Helper()

	if url := strings.TrimSpace(os.Getenv(DatabaseURLEnv)); url != "" {
		return url
	}
	requireDocker(t)

	const (
		user     = "simiente"
		password = "simiente"
		database = "simiente_test"
	)
	container := startContainer(t, testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     user,
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       database,
		},
		// PostgreSQL anuncia este mensaje dos veces: al inicializar y cuando ya
		// acepta conexiones. Se espera la segunda.
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	})

	ctx := context.Background()
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host del contenedor PostgreSQL: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("puerto mapeado del contenedor PostgreSQL: %v", err)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port.Port(), database)
}
