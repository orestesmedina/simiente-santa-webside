//go:build integration

package portada

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// TestIntegrationContentMutationAndAuditAreAtomic verifica el edge case de
// FR-017: la mutación de contenido y su fila de auditoría comparten transacción,
// de modo que si el registro falla, la mutación NO se aplica (nada sin
// registro).
//
// El fallo del registro se fuerza con un actor que no existe en `users` (UUID
// aleatorio): el INSERT en admin_actions viola su FK y revierte toda la
// transacción. No se crea un actor real para no acoplar la prueba a la tabla
// `users`, que las pruebas de integración de otros dominios truncan en paralelo.
func TestIntegrationContentMutationAndAuditAreAtomic(t *testing.T) {
	repo, _ := newHomeRepo(t)
	ctx := context.Background()

	// Línea base: una escritura sin auditoría que debe sobrevivir al intento
	// fallido posterior.
	if _, err := repo.UpsertHomeAbout(ctx, About{
		TextEs:           "Base",
		PublicationState: StatePublished,
	}); err != nil {
		t.Fatalf("UpsertHomeAbout(base): %v", err)
	}

	ghost := uuid.New()
	action := audit.Action{
		ActorUserID: &ghost,
		Code:        audit.ActionHomeAboutUpdate,
		TargetKind:  audit.TargetContent,
		TargetLabel: "Portada · Quiénes somos",
		Result:      audit.ResultSuccess,
	}
	if _, err := repo.UpsertHomeAbout(ctx, About{
		TextEs:           "No debe aplicarse",
		PublicationState: StateDraft,
	}, action); err == nil {
		t.Fatal("se esperaba un error al fallar el registro de auditoría")
	}

	// La mutación no se aplicó: la fila sigue con los valores de la línea base.
	got, ok, err := repo.GetHomeAbout(ctx)
	if err != nil {
		t.Fatalf("GetHomeAbout tras el fallo: %v", err)
	}
	if !ok {
		t.Fatal("la fila base no debería haber desaparecido")
	}
	if got.TextEs != "Base" || got.PublicationState != StatePublished {
		t.Fatalf("la mutación sin registro sí se aplicó: %+v", got)
	}
}
