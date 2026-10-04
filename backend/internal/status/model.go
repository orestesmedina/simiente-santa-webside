// Package status implementa el dominio "estado del sistema": el informe de
// salud que expone GET /healthz y la regla que lo decide (conexión viva → ok;
// conexión caída → 503 database_unavailable).
//
// Es la primera área construida con la receta de docs/tecnico/arquitectura.md
// §8 y el patrón que copiarán los dominios de F2–F9: capas handler → service →
// repository, interfaces definidas por quien las consume (arq. R3) y sin
// estado global (arq. R6).
package status

// Valores del contrato backend/api/openapi.yaml (esquema SystemStatus). El DTO
// es espejo exacto del contrato: un 200 siempre lleva status "ok" y database
// "connected"; el caso no conectado se comunica con el sobre de error 503
// (error.code = database_unavailable, details.database = "disconnected").
const (
	// StatusOK es el único valor de status en F1.
	StatusOK = "ok"
	// DatabaseConnected indica que la conexión con PostgreSQL está establecida
	// en el momento de la consulta (FR-003).
	DatabaseConnected = "connected"
	// DatabaseDisconnected acompaña el 503 como details.database.
	DatabaseDisconnected = "disconnected"
)

// SystemStatus es el sobre de éxito de GET /healthz: el DTO directo de la
// operación (sin wrapper), en camelCase y con additionalProperties: false en el
// contrato. Los nombres de los campos y sus valores son espejo del esquema
// SystemStatus de backend/api/openapi.yaml.
type SystemStatus struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
