package status

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/database"
)

// repository implementa la interfaz Repository (definida en service.go) sobre
// el *pgxpool.Pool compartido. No conoce HTTP ni escribe SQL de negocio: F1
// solo comprueba la conexión (arq. R4).
type repository struct {
	pool *pgxpool.Pool
}

// NewRepository construye el repositorio del dominio status a partir del pool
// compartido que compone cmd/api.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// Ping delega en el helper de salud de platform/database: aplica un timeout de
// 2 s por petición (D8) y envuelve el fallo con %w (arq. R7). Se ejecuta en
// cada consulta, de modo que el estado refleja la conexión real y no uno
// memorizado del arranque (FR-003).
func (r *repository) Ping(ctx context.Context) error {
	return database.Ping(ctx, r.pool)
}
