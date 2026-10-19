package medios

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// Tests del registro de auditoría del dominio (T922/P9-14): la tabla
// método+ruta → acción de las rutas del módulo, `RecordDenied` (el puerto que
// consume middleware.authz) y el helper de rechazo por JSON/DTO inválido.
// Ninguna vía transporta el cuerpo de la petición (regla de F2) y las rutas sin
// acción (las lecturas del panel o las ajenas) no dejan fila.

// TestAuditActionFromRoute fija la tabla de rutas del módulo: las cuatro
// mutaciones del panel resuelven a su código `media.*` con su objetivo, y todo lo
// demás (lecturas, otros módulos, ids inválidos) no deja fila.
func TestAuditActionFromRoute(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name      string
		method    string
		path      string
		wantCode  string
		wantLabel string
		wantOK    bool
	}{
		{
			name:      "búsqueda de episodios nuevos",
			method:    http.MethodPost,
			path:      "/api/v1/admin/medios/buscar-nuevos",
			wantCode:  audit.ActionMediaSync,
			wantLabel: busquedaLabel,
			wantOK:    true,
		},
		{
			name:      "edición de la descripción y la visibilidad",
			method:    http.MethodPatch,
			path:      "/api/v1/admin/medios/" + id.String(),
			wantCode:  audit.ActionMediaDescriptionUpdate,
			wantLabel: "Medios · " + id.String(),
			wantOK:    true,
		},
		{
			name:      "unión de duplicados",
			method:    http.MethodPost,
			path:      "/api/v1/admin/medios/" + id.String() + "/unir",
			wantCode:  audit.ActionMediaMerge,
			wantLabel: "Medios · " + id.String(),
			wantOK:    true,
		},
		{
			name:      "resolución de una marca",
			method:    http.MethodPost,
			path:      "/api/v1/admin/medios/" + id.String() + "/marcas/resolver",
			wantCode:  audit.ActionMediaMarkResolve,
			wantLabel: "Medios · " + id.String(),
			wantOK:    true,
		},
		{
			name:   "lectura del listado: sin acción",
			method: http.MethodGet,
			path:   "/api/v1/admin/medios",
			wantOK: false,
		},
		{
			name:   "la búsqueda por GET: sin acción",
			method: http.MethodGet,
			path:   "/api/v1/admin/medios/buscar-nuevos",
			wantOK: false,
		},
		{
			name:   "unir por PATCH: sin acción",
			method: http.MethodPatch,
			path:   "/api/v1/admin/medios/" + id.String() + "/unir",
			wantOK: false,
		},
		{
			name:   "resolver marcas sin la subruta: sin acción",
			method: http.MethodPost,
			path:   "/api/v1/admin/medios/" + id.String() + "/marcas",
			wantOK: false,
		},
		{
			name:   "la colección sin recurso: sin acción",
			method: http.MethodPost,
			path:   "/api/v1/admin/medios",
			wantOK: false,
		},
		{
			name:   "otro módulo: sin acción",
			method: http.MethodPatch,
			path:   "/api/v1/admin/donaciones/" + id.String(),
			wantOK: false,
		},
		{
			name:   "la superficie pública: sin acción",
			method: http.MethodGet,
			path:   "/api/v1/medios",
			wantOK: false,
		},
		{
			name:   "ruta ajena a la API: sin acción",
			method: http.MethodPost,
			path:   "/portal/medios/" + id.String() + "/unir",
			wantOK: false,
		},
		{
			name:   "id que no es uuid: sin acción",
			method: http.MethodPatch,
			path:   "/api/v1/admin/medios/no-es-uuid",
			wantOK: false,
		},
		{
			name:   "subruta desconocida con id: sin acción",
			method: http.MethodPatch,
			path:   "/api/v1/admin/medios/" + id.String() + "/extra",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, ok := auditActionFromRoute(tt.method, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, se esperaba %v (acción %+v)", ok, tt.wantOK, action)
			}
			if !tt.wantOK {
				return
			}
			if action.Code != tt.wantCode {
				t.Errorf("code = %q, se esperaba %q", action.Code, tt.wantCode)
			}
			if action.TargetKind != audit.TargetContent {
				t.Errorf("targetKind = %q, se esperaba content", action.TargetKind)
			}
			if action.TargetLabel != tt.wantLabel {
				t.Errorf("targetLabel = %q, se esperaba %q", action.TargetLabel, tt.wantLabel)
			}
			// El registro de `admin_actions` admite 254 caracteres: ninguna
			// etiqueta de ruta puede romperlo.
			if runes := []rune(action.TargetLabel); len(runes) > maxTargetLabel {
				t.Errorf("targetLabel de %d caracteres, el máximo es %d", len(runes), maxTargetLabel)
			}
		})
	}
}

// TestRecordDenied fija la denegación de permiso (FR-021/SC-009): resuelve la
// acción de la ruta y la persiste con `result='denied'` y el actor de la sesión.
func TestRecordDenied(t *testing.T) {
	actor := uuid.New()
	id := uuid.New()

	tests := []struct {
		name      string
		method    string
		path      string
		wantCode  string
		wantLabel string
		wantRow   bool
	}{
		{
			name:      "búsqueda denegada",
			method:    http.MethodPost,
			path:      "/api/v1/admin/medios/buscar-nuevos",
			wantCode:  audit.ActionMediaSync,
			wantLabel: busquedaLabel,
			wantRow:   true,
		},
		{
			name:      "edición denegada",
			method:    http.MethodPatch,
			path:      "/api/v1/admin/medios/" + id.String(),
			wantCode:  audit.ActionMediaDescriptionUpdate,
			wantLabel: "Medios · " + id.String(),
			wantRow:   true,
		},
		{
			name:      "unión denegada",
			method:    http.MethodPost,
			path:      "/api/v1/admin/medios/" + id.String() + "/unir",
			wantCode:  audit.ActionMediaMerge,
			wantLabel: "Medios · " + id.String(),
			wantRow:   true,
		},
		{
			name:      "resolución denegada",
			method:    http.MethodPost,
			path:      "/api/v1/admin/medios/" + id.String() + "/marcas/resolver",
			wantCode:  audit.ActionMediaMarkResolve,
			wantLabel: "Medios · " + id.String(),
			wantRow:   true,
		},
		{
			name:    "una lectura denegada no deja fila",
			method:  http.MethodGet,
			path:    "/api/v1/admin/medios",
			wantRow: false,
		},
		{
			name:    "una ruta ajena no deja fila",
			method:  http.MethodPost,
			path:    "/api/v1/admin/ministerios",
			wantRow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			svc := NewService(ServiceDeps{Repository: repo})

			err := svc.RecordDenied(context.Background(), audit.Denial{
				ActorUserID: &actor,
				Method:      tt.method,
				Path:        tt.path,
			})
			if err != nil {
				t.Fatalf("RecordDenied: %v", err)
			}
			if !tt.wantRow {
				if len(repo.recorded) != 0 {
					t.Fatalf("no debía dejar fila: %+v", repo.recorded)
				}
				return
			}
			if len(repo.recorded) != 1 {
				t.Fatalf("filas = %+v, se esperaba 1", repo.recorded)
			}
			action := repo.recorded[0]
			if action.Code != tt.wantCode || action.Result != audit.ResultDenied {
				t.Errorf("acción = %+v, se esperaba %s con result=denied", action, tt.wantCode)
			}
			if action.TargetLabel != tt.wantLabel {
				t.Errorf("targetLabel = %q, se esperaba %q", action.TargetLabel, tt.wantLabel)
			}
			if action.TargetKind != audit.TargetContent {
				t.Errorf("targetKind = %q, se esperaba content", action.TargetKind)
			}
			if action.ActorUserID == nil || *action.ActorUserID != actor {
				t.Errorf("la denegación no lleva el actor de la sesión: %+v", action)
			}
		})
	}
}

// TestRecordDeniedError fija que un fallo de registro no se traga: `RecordDenied`
// devuelve el error (el middleware lo registra en el log) y nunca cambia la
// respuesta, que ya salió con su 403.
func TestRecordDeniedError(t *testing.T) {
	repo := newFakeRepository()
	repo.recordErr = errors.New("boom")
	svc := NewService(ServiceDeps{Repository: repo})

	err := svc.RecordDenied(context.Background(), audit.Denial{
		Method: http.MethodPatch,
		Path:   "/api/v1/admin/medios/" + uuid.NewString(),
	})
	if err == nil {
		t.Fatal("un fallo de registro debía devolver error")
	}
	if len(repo.recorded) != 0 {
		t.Errorf("no debía quedar fila: %+v", repo.recorded)
	}
}

// TestRecordRejectedBestEffort fija el rechazo por JSON o DTO inválido
// (`result='failure'`, R7-10): deja la fila con el actor cuando lo hay, no deja
// fila en las rutas sin acción y **nunca** cambia la respuesta (un fallo de
// registro se queda en el log).
func TestRecordRejectedBestEffort(t *testing.T) {
	actor := uuid.New()
	id := uuid.New()

	t.Run("rechazo de una mutación del módulo", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewService(ServiceDeps{Repository: repo})

		svc.RecordRejectedBestEffort(context.Background(), Rejection{
			ActorUserID: &actor,
			Method:      http.MethodPatch,
			Path:        "/api/v1/admin/medios/" + id.String(),
		})

		if len(repo.recorded) != 1 {
			t.Fatalf("filas = %+v, se esperaba 1", repo.recorded)
		}
		action := repo.recorded[0]
		if action.Code != audit.ActionMediaDescriptionUpdate || action.Result != audit.ResultFailure {
			t.Errorf("acción = %+v, se esperaba media.description.update con result=failure", action)
		}
		if action.ActorUserID == nil || *action.ActorUserID != actor {
			t.Errorf("el rechazo no lleva el actor: %+v", action)
		}
	})

	t.Run("sin actor (petición sin sesión)", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewService(ServiceDeps{Repository: repo})

		svc.RecordRejectedBestEffort(context.Background(), Rejection{
			Method: http.MethodPost,
			Path:   "/api/v1/admin/medios/buscar-nuevos",
		})

		if len(repo.recorded) != 1 {
			t.Fatalf("filas = %+v, se esperaba 1", repo.recorded)
		}
		if repo.recorded[0].ActorUserID != nil {
			t.Errorf("sin identidad el actor debe ser nil: %+v", repo.recorded[0])
		}
		if repo.recorded[0].Code != audit.ActionMediaSync {
			t.Errorf("code = %q, se esperaba %q", repo.recorded[0].Code, audit.ActionMediaSync)
		}
	})

	t.Run("ruta sin acción", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewService(ServiceDeps{Repository: repo})

		svc.RecordRejectedBestEffort(context.Background(), Rejection{
			Method: http.MethodGet,
			Path:   "/api/v1/admin/medios",
		})

		if len(repo.recorded) != 0 {
			t.Errorf("una lectura no deja fila: %+v", repo.recorded)
		}
	})

	t.Run("un fallo de registro no cambia la respuesta", func(t *testing.T) {
		repo := newFakeRepository()
		repo.recordErr = errors.New("boom")
		svc := NewService(ServiceDeps{Repository: repo})

		// No debe entrar en pánico ni propagar nada: es best-effort.
		svc.RecordRejectedBestEffort(context.Background(), Rejection{
			Method: http.MethodPatch,
			Path:   "/api/v1/admin/medios/" + id.String(),
		})
		if len(repo.recorded) != 0 {
			t.Errorf("el fallo no debía dejar fila: %+v", repo.recorded)
		}
	})
}

// TestServiceImplementsAuditRecorder fija el contrato que consume
// middleware.authz: el service del dominio es el `audit.Recorder` de su cadena.
func TestServiceImplementsAuditRecorder(t *testing.T) {
	var _ audit.Recorder = NewService(ServiceDeps{Repository: newFakeRepository()})
}
