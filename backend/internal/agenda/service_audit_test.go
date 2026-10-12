package agenda

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// Pruebas del registro de denegaciones y rechazos del módulo (T422): cada ruta
// de F4 resuelve a su acción y su etiqueta, las lecturas y las rutas ajenas no
// dejan fila y un fallo de registro no cambia la respuesta.

// TestAuditActionFromRoute fija la tabla método+ruta → acción (R4-10) de la
// superficie de F4.
func TestAuditActionFromRoute(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name      string
		method    string
		path      string
		wantOK    bool
		wantCode  string
		wantLabel string
	}{
		{name: "alta de evento", method: http.MethodPost, path: "/api/v1/admin/eventos", wantOK: true, wantCode: audit.ActionEventCreate, wantLabel: "Eventos"},
		{name: "edición de evento", method: http.MethodPatch, path: "/api/v1/admin/eventos/" + id.String(), wantOK: true, wantCode: audit.ActionEventUpdate, wantLabel: "Eventos · " + id.String()},
		{name: "borrado de evento", method: http.MethodDelete, path: "/api/v1/admin/eventos/" + id.String(), wantOK: true, wantCode: audit.ActionEventDelete, wantLabel: "Eventos · " + id.String()},
		{name: "imagen de evento", method: http.MethodPost, path: "/api/v1/admin/eventos/imagenes", wantOK: true, wantCode: audit.ActionEventImageUpload, wantLabel: "Eventos · Imagen"},
		{name: "alta de actividad", method: http.MethodPost, path: "/api/v1/admin/actividades", wantOK: true, wantCode: audit.ActionActivityCreate, wantLabel: "Actividades"},
		{name: "edición de actividad", method: http.MethodPatch, path: "/api/v1/admin/actividades/" + id.String(), wantOK: true, wantCode: audit.ActionActivityUpdate, wantLabel: "Actividades · " + id.String()},
		{name: "borrado de actividad", method: http.MethodDelete, path: "/api/v1/admin/actividades/" + id.String(), wantOK: true, wantCode: audit.ActionActivityDelete, wantLabel: "Actividades · " + id.String()},
		{name: "imagen de actividad", method: http.MethodPost, path: "/api/v1/admin/actividades/imagenes", wantOK: true, wantCode: audit.ActionActivityImageUpload, wantLabel: "Actividades · Imagen"},
		{name: "alta de categoría", method: http.MethodPost, path: "/api/v1/admin/categorias", wantOK: true, wantCode: audit.ActionCategoryCreate, wantLabel: "Categorías"},
		{name: "edición de categoría", method: http.MethodPatch, path: "/api/v1/admin/categorias/" + id.String(), wantOK: true, wantCode: audit.ActionCategoryUpdate, wantLabel: "Categorías · " + id.String()},
		{name: "borrado de categoría", method: http.MethodDelete, path: "/api/v1/admin/categorias/" + id.String(), wantOK: true, wantCode: audit.ActionCategoryDelete, wantLabel: "Categorías · " + id.String()},

		// Lecturas del panel: no dejan fila (el registro es de mutaciones).
		{name: "lectura de eventos", method: http.MethodGet, path: "/api/v1/admin/eventos", wantOK: false},
		{name: "ficha de evento", method: http.MethodGet, path: "/api/v1/admin/eventos/" + id.String(), wantOK: false},
		{name: "lectura de actividades", method: http.MethodGet, path: "/api/v1/admin/actividades", wantOK: false},
		{name: "lectura de categorías", method: http.MethodGet, path: "/api/v1/admin/categorias", wantOK: false},

		// Rutas ajenas al módulo, rutas inventadas y el catálogo sin imágenes.
		{name: "otro módulo", method: http.MethodPatch, path: "/api/v1/admin/usuarios/" + id.String(), wantOK: false},
		{name: "portada", method: http.MethodPut, path: "/api/v1/admin/portada/identidad", wantOK: false},
		{name: "recurso desconocido", method: http.MethodPost, path: "/api/v1/admin/eventos/desconocido", wantOK: false},
		{name: "id no uuid", method: http.MethodPatch, path: "/api/v1/admin/eventos/no-es-uuid", wantOK: false},
		{name: "método no documentado", method: http.MethodPut, path: "/api/v1/admin/eventos/" + id.String(), wantOK: false},
		{name: "imágenes de categorías", method: http.MethodPost, path: "/api/v1/admin/categorias/imagenes", wantOK: false},
		{name: "ruta pública", method: http.MethodGet, path: "/api/v1/proximamente", wantOK: false},
		{name: "ruta vacía", method: http.MethodPost, path: "/", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, ok := auditActionFromRoute(tt.method, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, se esperaba %v (%+v)", ok, tt.wantOK, action)
			}
			if !tt.wantOK {
				return
			}
			if action.Code != tt.wantCode {
				t.Errorf("código = %q, se esperaba %q", action.Code, tt.wantCode)
			}
			if action.TargetKind != audit.TargetContent {
				t.Errorf("targetKind = %q, se esperaba content", action.TargetKind)
			}
			if action.TargetLabel != tt.wantLabel {
				t.Errorf("targetLabel = %q, se esperaba %q", action.TargetLabel, tt.wantLabel)
			}
			// Todo código resuelto pertenece al registro cerrado de platform/audit.
			if !audit.ValidActionCode(action.Code) {
				t.Errorf("el código %q no está en el registro cerrado", action.Code)
			}
			// El actor y el resultado los fija quien registra, no la tabla.
			if action.ActorUserID != nil || action.Result != "" {
				t.Errorf("la tabla no debe fijar actor ni resultado: %+v", action)
			}
		})
	}
}

// TestRecordDeniedPersistsTheDenial fija el contrato de audit.Recorder: la
// denegación se persiste con `result='denied'`, su actor y la etiqueta de la
// ruta; sin acción asociada no deja fila y sin cuerpo de petición (FR-026).
func TestRecordDeniedPersistsTheDenial(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	actor := uuid.New()
	id := uuid.New()

	err := svc.RecordDenied(context.Background(), audit.Denial{
		ActorUserID: &actor,
		Method:      http.MethodPatch,
		Path:        "/api/v1/admin/eventos/" + id.String(),
	})
	if err != nil {
		t.Fatalf("RecordDenied: %v", err)
	}
	if len(repo.recorded) != 1 {
		t.Fatalf("filas = %+v, se esperaba 1", repo.recorded)
	}
	action := repo.recorded[0]
	if action.Code != audit.ActionEventUpdate || action.Result != audit.ResultDenied {
		t.Errorf("acción = %+v", action)
	}
	if action.ActorUserID == nil || *action.ActorUserID != actor {
		t.Errorf("actor = %v", action.ActorUserID)
	}
	if action.TargetLabel != "Eventos · "+id.String() {
		t.Errorf("targetLabel = %q", action.TargetLabel)
	}

	// Una lectura no deja fila.
	if err := svc.RecordDenied(context.Background(), audit.Denial{
		ActorUserID: &actor,
		Method:      http.MethodGet,
		Path:        "/api/v1/admin/eventos",
	}); err != nil {
		t.Fatalf("RecordDenied(lectura): %v", err)
	}
	if len(repo.recorded) != 1 {
		t.Errorf("una lectura no debe dejar fila: %+v", repo.recorded)
	}
}

// TestRecordDeniedReportsRepositoryFailure fija que el error del registro se
// devuelve (el middleware lo registra) sin romper la respuesta al cliente.
func TestRecordDeniedReportsRepositoryFailure(t *testing.T) {
	repo := newFakeRepository()
	repo.recordErr = errors.New("auditoría caída")
	svc := newAdminService(t, repo)

	err := svc.RecordDenied(context.Background(), audit.Denial{
		Method: http.MethodDelete,
		Path:   "/api/v1/admin/categorias/" + uuid.New().String(),
	})
	if err == nil {
		t.Fatal("se esperaba el error del repositorio")
	}
}

// TestRecordRejectedBestEffort fija el helper de rechazos: `result='failure'`,
// el actor de la petición y ningún dato del cuerpo (FR-026), sin romper nada si
// el registro falla.
func TestRecordRejectedBestEffort(t *testing.T) {
	t.Run("registra el rechazo", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)
		actor := uuid.New()

		svc.RecordRejectedBestEffort(context.Background(), Rejection{
			ActorUserID: &actor,
			Method:      http.MethodPost,
			Path:        "/api/v1/admin/actividades",
		})
		if len(repo.recorded) != 1 {
			t.Fatalf("filas = %+v, se esperaba 1", repo.recorded)
		}
		action := repo.recorded[0]
		if action.Code != audit.ActionActivityCreate || action.Result != audit.ResultFailure {
			t.Errorf("acción = %+v", action)
		}
		if action.ActorUserID == nil || *action.ActorUserID != actor {
			t.Errorf("actor = %v", action.ActorUserID)
		}
		if action.TargetLabel != "Actividades" {
			t.Errorf("targetLabel = %q", action.TargetLabel)
		}
	})

	t.Run("una ruta sin acción no deja fila", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)

		svc.RecordRejectedBestEffort(context.Background(), Rejection{
			Method: http.MethodGet,
			Path:   "/api/v1/admin/categorias",
		})
		if len(repo.recorded) != 0 {
			t.Errorf("filas = %+v, se esperaba ninguna", repo.recorded)
		}
	})

	t.Run("un fallo del registro no rompe", func(t *testing.T) {
		repo := newFakeRepository()
		repo.recordErr = errors.New("auditoría caída")
		svc := newAdminService(t, repo)

		// No debe entrar en pánico ni propagar nada: es best-effort.
		svc.RecordRejectedBestEffort(context.Background(), Rejection{
			Method: http.MethodPost,
			Path:   "/api/v1/admin/eventos",
		})
	})
}
