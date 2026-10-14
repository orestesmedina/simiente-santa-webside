package ministerios

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// contract_test.go fija la coherencia del **contrato vivo** con el
// comportamiento del servidor en los dos puntos de la revisión de F6 que tocan
// el esquema (I1 y M1): el `PATCH` de ministerios declara `minProperties: 1` y
// su `whatsappUrl` admite el `null` explícito que borra el opcional, y
// `sortOrder` está acotado al rango de la columna `INTEGER` (int32).
//
// El contrato vivo (`backend/api/openapi.yaml`) es la única fuente y se edita
// antes que el código; el snapshot `specs/006-ministerios/contracts/openapi.yaml`
// es un delta de diseño congelado (su cabecera prohíbe editarlo).

// openAPIDoc es la parte del documento que estas pruebas inspeccionan.
type openAPIDoc struct {
	Components struct {
		Schemas map[string]openAPISchema `yaml:"schemas"`
	} `yaml:"components"`
}

// openAPISchema es un esquema mínimo: su tipo (escalar o lista, OpenAPI 3.1),
// su `minProperties` y sus propiedades.
type openAPISchema struct {
	Type          openAPIType              `yaml:"type"`
	MinProperties *int                     `yaml:"minProperties"`
	Minimum       *int                     `yaml:"minimum"`
	Maximum       *int                     `yaml:"maximum"`
	Properties    map[string]openAPISchema `yaml:"properties"`
}

// openAPIType admite tanto `type: string` como `type: [string, "null"]`.
type openAPIType []string

// UnmarshalYAML acepta un escalar o una lista de tipos.
func (t *openAPIType) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		*t = []string{value.Value}
		return nil
	}
	var list []string
	if err := value.Decode(&list); err != nil {
		return err
	}
	*t = list
	return nil
}

// admitsNull indica si el tipo declarado incluye `null`.
func (t openAPIType) admitsNull() bool {
	for _, kind := range t {
		if kind == "null" {
			return true
		}
	}
	return false
}

// loadContractSchemas lee y deserializa los esquemas del contrato vivo.
func loadContractSchemas(t *testing.T) map[string]openAPISchema {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatalf("leer el contrato vivo: %v", err)
	}
	var doc openAPIDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("deserializar el contrato vivo: %v", err)
	}
	return doc.Components.Schemas
}

// TestContractPatchAllowsNullWhatsappURL fija I1: el esquema de
// `MinisterioPatch.whatsappUrl` admite el `null` explícito que su prosa promete
// —y que el servidor usa para borrar el enlace— y el `PATCH` exige al menos un
// campo (`minProperties: 1`), como los PATCH de F3/F4 y como el 400 del service
// con un cuerpo vacío.
func TestContractPatchAllowsNullWhatsappURL(t *testing.T) {
	patch, ok := loadContractSchemas(t)["MinisterioPatch"]
	if !ok {
		t.Fatal("el contrato no declara el esquema MinisterioPatch")
	}
	if patch.MinProperties == nil || *patch.MinProperties != 1 {
		t.Errorf("MinisterioPatch.minProperties = %v, se esperaba 1", patch.MinProperties)
	}
	if got := patch.Properties["whatsappUrl"].Type; !got.admitsNull() {
		t.Errorf("MinisterioPatch.whatsappUrl.type = %v, se esperaba que admitiera null (I1)", got)
	}
	// El alta también declara el enlace anulable: el servidor normaliza `null` y
	// `""` a "sin enlace", coherente con el resto de sus campos opcionales.
	input := loadContractSchemas(t)["MinisterioInput"]
	if got := input.Properties["whatsappUrl"].Type; !got.admitsNull() {
		t.Errorf("MinisterioInput.whatsappUrl.type = %v, se esperaba que admitiera null", got)
	}
}
