package portada

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// T321: cada ruta del módulo resuelve a su acción `home.*`.
func TestAuditActionFromRoute(t *testing.T) {
	id := uuid.New()
	cases := []struct {
		method string
		path   string
		code   string
		want   bool
	}{
		{http.MethodPut, "/api/v1/admin/portada/identidad", audit.ActionHomeIdentityUpdate, true},
		{http.MethodPut, "/api/v1/admin/portada/quienes-somos", audit.ActionHomeAboutUpdate, true},
		{http.MethodPut, "/api/v1/admin/portada/contacto", audit.ActionHomeContactUpdate, true},
		{http.MethodPost, "/api/v1/admin/portada/horario", audit.ActionHomeScheduleCreate, true},
		{http.MethodPatch, "/api/v1/admin/portada/horario/" + id.String(), audit.ActionHomeScheduleUpdate, true},
		{http.MethodDelete, "/api/v1/admin/portada/horario/" + id.String(), audit.ActionHomeScheduleDelete, true},
		{http.MethodPost, "/api/v1/admin/portada/whatsapp", audit.ActionHomeWhatsappCreate, true},
		{http.MethodPatch, "/api/v1/admin/portada/whatsapp/" + id.String(), audit.ActionHomeWhatsappUpdate, true},
		{http.MethodDelete, "/api/v1/admin/portada/whatsapp/" + id.String(), audit.ActionHomeWhatsappDelete, true},
		{http.MethodPost, "/api/v1/admin/portada/redes", audit.ActionHomeSocialCreate, true},
		{http.MethodPatch, "/api/v1/admin/portada/redes/" + id.String(), audit.ActionHomeSocialUpdate, true},
		{http.MethodDelete, "/api/v1/admin/portada/redes/" + id.String(), audit.ActionHomeSocialDelete, true},
		{http.MethodPost, "/api/v1/admin/portada/imagenes", audit.ActionHomeImageUpload, true},
		{http.MethodGet, "/api/v1/admin/portada", "", false},
		{http.MethodGet, "/api/v1/admin/portada/identidad", "", false},
		{http.MethodPost, "/api/v1/admin/usuarios", "", false},
		{http.MethodPatch, "/api/v1/portada", "", false},
	}
	for _, tc := range cases {
		action, ok := auditActionFromRoute(tc.method, tc.path)
		if ok != tc.want {
			t.Errorf("%s %s: ok = %v, se esperaba %v", tc.method, tc.path, ok, tc.want)
			continue
		}
		if !ok {
			continue
		}
		if action.Code != tc.code {
			t.Errorf("%s %s: code = %q, se esperaba %q", tc.method, tc.path, action.Code, tc.code)
		}
		if action.TargetKind != audit.TargetContent || action.TargetLabel == "" {
			t.Errorf("%s %s: objetivo mal formado: %+v", tc.method, tc.path, action)
		}
	}
}

// T321: el objetivo de una ruta con id incluye el UUID ("sin nombre, el id").
func TestAuditActionFromRouteLabelWithID(t *testing.T) {
	id := uuid.New()
	action, ok := auditActionFromRoute(http.MethodPatch, "/api/v1/admin/portada/horario/"+id.String())
	if !ok {
		t.Fatal("se esperaba acción")
	}
	want := "Portada · Horario · " + id.String()
	if action.TargetLabel != want {
		t.Fatalf("targetLabel = %q, se esperaba %q", action.TargetLabel, want)
	}
}

// T321: la denegación de una acción sensible deja su fila `denied`.
func TestRecordDenied(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})
	actor := uuid.New()

	err := service.RecordDenied(context.Background(), audit.Denial{
		ActorUserID: &actor,
		Method:      http.MethodPatch,
		Path:        "/api/v1/admin/portada/redes/" + uuid.New().String(),
	})
	if err != nil {
		t.Fatalf("RecordDenied = %v", err)
	}
	actions := repo.recorded()
	if len(actions) != 1 || actions[0].Result != audit.ResultDenied {
		t.Fatalf("acciones = %+v", actions)
	}
	if actions[0].ActorUserID == nil || *actions[0].ActorUserID != actor {
		t.Fatalf("actor no registrado: %+v", actions[0])
	}
}

// T321: una ruta sin acción asociada no deja fila.
func TestRecordDeniedUnknownRoute(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	if err := service.RecordDenied(context.Background(), audit.Denial{Method: http.MethodGet, Path: "/api/v1/admin/portada"}); err != nil {
		t.Fatalf("RecordDenied = %v", err)
	}
	if len(repo.recorded()) != 0 {
		t.Fatalf("no debía dejar fila: %+v", repo.recorded())
	}
}

// T321: el rechazo por DTO inválido se registra como `failure` best-effort.
func TestRecordRejectedBestEffort(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})
	actor := uuid.New()

	service.RecordRejectedBestEffort(context.Background(), Rejection{
		ActorUserID: &actor,
		Method:      http.MethodPost,
		Path:        "/api/v1/admin/portada/whatsapp",
	})
	actions := repo.recorded()
	if len(actions) != 1 || actions[0].Result != audit.ResultFailure || actions[0].Code != audit.ActionHomeWhatsappCreate {
		t.Fatalf("acciones = %+v", actions)
	}
}

// T321: un rechazo de ruta desconocida no deja fila.
func TestRecordRejectedUnknownRoute(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(ServiceDeps{Repository: repo})

	service.RecordRejectedBestEffort(context.Background(), Rejection{Method: http.MethodPost, Path: "/api/v1/otra"})
	if len(repo.recorded()) != 0 {
		t.Fatalf("no debía dejar fila: %+v", repo.recorded())
	}
}

// T321: un fallo de registro no cambia la respuesta (best-effort).
func TestAuditBestEffortSwallowsErrors(t *testing.T) {
	repo := newFakeRepository()
	repo.recordErr = errors.New("boom")
	service := NewService(ServiceDeps{Repository: repo})

	if err := service.RecordDenied(context.Background(), audit.Denial{Method: http.MethodPost, Path: "/api/v1/admin/portada/imagenes"}); err == nil {
		t.Fatal("RecordDenied debe propagar el error para que el middleware lo registre")
	}
	// El rechazo best-effort no debe entrar en pánico ni propagar.
	service.RecordRejectedBestEffort(context.Background(), Rejection{Method: http.MethodPost, Path: "/api/v1/admin/portada/imagenes"})
}
