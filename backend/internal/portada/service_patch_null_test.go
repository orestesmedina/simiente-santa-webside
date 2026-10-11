package portada

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
)

/**
 * Pruebas de la semántica `null` del contrato (B1, `revision-2026-10-10-codigo.md`).
 *
 * El contrato (`ScheduleItemPatch`, `WhatsappChannelPatch`) promete: «Un campo
 * `*En` con `null` o `""`/espacios vacía su traducción al inglés (analyze I6);
 * `endTime: null` quita la hora de fin». El panel cumple el contrato y envía
 * `null` para limpiar (`ServiceForm.tsx` con `nullable()`, `WhatsAppForm.tsx`).
 *
 * Estas pruebas decodifican el CUERPO JSON REAL del panel con `encoding/json`
 * (no construyen el struct a mano) para ejercitar el camino exacto de la UI.
 *
 * Estado: FALLAN mientras el defecto B1 no se corrija (JSON `null` queda
 * indistinguible de un campo ausente y la limpieza se pierde en silencio, con
 * es «Guardado» del panel que devuelve el valor anterior tras el refetch).
 * Deben quedar en verde con la corrección de B1 y ser su guardia de regresión.
 */

// decodePatch deserializa un cuerpo JSON del panel en el tipo del contrato.
func decodePatch[T any](t *testing.T, body string) T {
	t.Helper()
	var patch T
	if err := json.Unmarshal([]byte(body), &patch); err != nil {
		t.Fatalf("el cuerpo del panel (%s) debe decodificarse: %v", body, err)
	}
	return patch
}

// TestPatchScheduleJSONNullClearsEndTime es el escenario exacto del panel: el
// usuario quita la hora de fin de un servicio y pulsa Guardar.
func TestPatchScheduleJSONNullClearsEndTime(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedService(Service{
		NameEs: "Culto", PlaceEs: "Templo", StartTime: "10:00",
		EndTime: strptr("12:00"), PublicationState: StatePublished,
	})
	service := NewService(ServiceDeps{Repository: repo})

	// Payload tal cual lo envía ServiceForm.tsx cuando el campo queda vacío.
	patch := decodePatch[ScheduleItemPatch](t, `{"endTime": null}`)
	if !patch.hasChanges() {
		t.Fatal(`{"endTime": null} es un cambio real (minProperties: 1): el PATCH no debe tratarse como vacío`)
	}

	saved, err := service.UpdateService(context.Background(), uuid.New(), current.ID, patch)
	if err != nil {
		t.Fatalf("UpdateService = %v", err)
	}
	if saved.EndTime != nil {
		t.Fatalf(`endTime: null del contrato debe QUITAR la hora de fin; quedó %v`, *saved.EndTime)
	}

	// La limpieza de una traducción al inglés es el mismo mecanismo.
	patchNameEn := decodePatch[ScheduleItemPatch](t, `{"nameEn": null}`)
	if !patchNameEn.hasChanges() {
		t.Fatal(`{"nameEn": null} es un cambio real`)
	}
	withEn := repo.seedService(Service{
		NameEs: "Servicio con inglés", NameEn: strptr("English name"), PlaceEs: "Templo",
		StartTime: "10:00", PublicationState: StatePublished,
	})
	savedEn, err := service.UpdateService(context.Background(), uuid.New(), withEn.ID, patchNameEn)
	if err != nil {
		t.Fatalf("UpdateService = %v", err)
	}
	if savedEn.NameEn != nil {
		t.Fatalf(`nameEn: null debe vaciar la traducción; quedó %v`, *savedEn.NameEn)
	}
}

// TestPatchWhatsappJSONNullClearsNameEn es el escenario exacto de
// WhatsAppForm.tsx: vacía el nombre en inglés y espera que se quite.
func TestPatchWhatsappJSONNullClearsNameEn(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedChannel(WhatsappChannel{
		NameEs: "Escríbenos", NameEn: strptr("Write to us"),
		Kind: KindDirect, Destination: "04121234567", PublicationState: StatePublished,
	})
	service := NewService(ServiceDeps{Repository: repo})

	patch := decodePatch[WhatsappChannelPatch](t, `{"nameEn": null}`)
	if !patch.hasChanges() {
		t.Fatal(`{"nameEn": null} es un cambio real`)
	}
	saved, err := service.UpdateWhatsappChannel(context.Background(), uuid.New(), current.ID, patch)
	if err != nil {
		t.Fatalf("UpdateWhatsappChannel = %v", err)
	}
	if saved.NameEn != nil {
		t.Fatalf(`nameEn: null debe vaciar la traducción; quedó %v`, *saved.NameEn)
	}
}

// TestPatchScheduleNullOnlyIsAccepted documenta que un PATCH cuyo único campo
// es `null` no puede responder 400 «No hay cambios que guardar»: limpiar ES un
// cambio (el contrato admite el objeto con `minProperties: 1`).
func TestPatchScheduleNullOnlyIsAccepted(t *testing.T) {
	repo := newFakeRepository()
	current := repo.seedService(Service{
		NameEs: "Culto", PlaceEs: "Templo", StartTime: "10:00",
		EndTime: strptr("12:00"), PublicationState: StatePublished,
	})
	service := NewService(ServiceDeps{Repository: repo})

	_, err := service.UpdateService(
		context.Background(), uuid.New(), current.ID,
		decodePatch[ScheduleItemPatch](t, `{"endTime": null}`),
	)
	if err != nil && errorKind(err) == apperr.KindInvalid && strings.Contains(err.Error(), "No hay cambios") {
		t.Fatalf(`limpiar la hora de fin es un cambio y no debe responder 400 «No hay cambios que guardar»: %v`, err)
	}
}
