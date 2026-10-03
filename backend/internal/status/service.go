package status

import (
	"context"
	"errors"
	"fmt"

	"simiente-santa/backend/internal/platform/apperr"
)

// Mensaje y clave de detalle del 503 previsto por el contrato (FR-004, D7). Son
// seguros para el cliente: nunca contienen el error interno, que viaja envuelto
// solo para el log.
const (
	// MessageDatabaseUnavailable es el message del 503 database_unavailable.
	MessageDatabaseUnavailable = "La base de datos no está conectada"
	// DetailDatabaseKey es la clave de details que acompaña ese 503.
	DetailDatabaseKey = "database"
)

// Repository es la interfaz que este servicio necesita del almacén de datos.
// La define quien la consume (arq. R3) y la implementa repository.go sobre el
// *pgxpool.Pool compartido. El timeout de la comprobación ya viene aplicado en
// platform/database, así que el servicio no conoce pgx ni SQL (arq. R4).
type Repository interface {
	// Ping comprueba la conexión con la base de datos en cada llamada, para que
	// el estado sea el real y no uno memorizado (FR-003).
	Ping(ctx context.Context) error
}

// service implementa las reglas de negocio del estado del sistema. No conoce
// net/http ni pgx (arq. R4).
type service struct {
	repo Repository
}

// NewService construye el servicio del dominio status. La interfaz Service que
// declara handler.go (quien la consume) la cumple el valor devuelto.
func NewService(repo Repository) *service {
	return &service{repo: repo}
}

// Status consulta el estado real de la conexión (FR-003) y lo traduce al DTO
// del contrato. Un fallo del ping es el caso previsible "base de datos no
// conectada" (apperr.DatabaseUnavailable → 503 con details); una cancelación
// del contexto no es evidencia de que la base de datos esté caída, así que se
// propaga envuelta y el handler la trata como error inesperado (500 internal).
func (s *service) Status(ctx context.Context) (SystemStatus, error) {
	if err := ctx.Err(); err != nil {
		return SystemStatus{}, fmt.Errorf("check system status: %w", err)
	}

	if err := s.repo.Ping(ctx); err != nil {
		if isContextError(err) {
			return SystemStatus{}, fmt.Errorf("check system status: %w", err)
		}
		return SystemStatus{}, apperr.DatabaseUnavailable(
			MessageDatabaseUnavailable,
			apperr.WithDetails(map[string]any{DetailDatabaseKey: DatabaseDisconnected}),
			apperr.WithCause(err),
		)
	}

	return SystemStatus{Status: StatusOK, Database: DatabaseConnected}, nil
}

// isContextError distingue un fallo por cancelación o vencimiento del contexto
// de un fallo de conexión con la base de datos.
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
