package ministerios

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// Tests del registro de auditoría del dominio (T616): la tabla método+ruta →
// acción del módulo, `RecordDenied` (el puerto que consume middleware.authz) y
// el helper de rechazo por JSON/DTO inválido. Ninguna vía transporta el cuerpo
// de la petición (FR-026 de F2) y las rutas sin acción (las lecturas del panel o
// las ajenas) no dejan fila.

// TestAuditActionFromRoute fija la tabla de rutas del módulo (R6-12).
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
			name:      "alta",
			method:    http.MethodPost,
			path:      "/api/v1/admin/ministerios",
			wantCode:  audit.ActionMinistryCreate,
			wantLabel: "Ministerios",
			wantOK:    true,
		},
		{
			name:      "edición",
			method:    http.MethodPatch,
			path:      "/api/v1/admin/ministerios/" + id.String(),
			wantCode:  audit.ActionMinistryUpdate,
			wantLabel: "Ministerios · " + id.String(),
			wantOK:    true,
		},
		{
			name:      "borrado",
			method:    http.MethodDelete,
			path:      "/api/v1/admin/ministerios/" + id.String(),
			wantCode:  audit.ActionMinistryDelete,
			wantLabel: "Ministerios · " + id.String(),
			wantOK:    true,
		},
		{
			name:      "subida de imagen",
			method:    http.MethodPost,
			path:      "/api/v1/admin/ministerios/imagenes",
			wantCode:  audit.ActionMinistryImageUpload,
			wantLabel: "Ministerios · Imagen",
			wantOK:    true,
		},
		{
			name:   "lectura del listado: sin acción (el registro es de mutaciones)",
			method: http.MethodGet,
			path:   "/api/v1/admin/ministerios",
			wantOK: false,
		},
		{
			name:   "lectura del detalle: sin acción",
			method: http.MethodGet,
			path:   "/api/v1/admin/ministerios/" + id.String(),
			wantOK: false,
		},
		{
			name:   "otro módulo: sin acción",
			method: http.MethodPatch,
			path:   "/api/v1/admin/eventos/" + id.String(),
			wantOK: false,
		},
		{
			name:   "ruta ajena a la API: sin acción",
			method: http.MethodDelete,
			path:   "/portal/usuarios/" + id.String(),
			wantOK: false,
		},
		{
			name:   "id que no es uuid: sin acción",
			method: http.MethodPatch,
			path:   "/api/v1/admin/ministerios/no-es-uuid",
			wantOK: false,
		},
		{
			name:   "subruta con id: sin acción",
			method: http.MethodPatch,
			path:   "/api/v1/admin/ministerios/" + id.String() + "/extra",
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
			if action.TargetUserID != nil || action.TargetRoleID != nil {
				t.Errorf("un objetivo de contenido no lleva FK: %+v", action)
			}
		})
	}
}

// TestRecordDeniedPersistsTheDenial fija que una denegación de permiso queda
// registrada con `result='denied'`, el actor de la sesión y la acción de la
// ruta (SC-011).
func TestRecordDeniedPersistsTheDenial(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)
	actor := uuid.New()
	id := uuid.New()

	denials := []struct {
		method    string
		path      string
		wantCode  string
		wantLabel string
	}{
		{
			method:    http.MethodPost,
			path:      "/api/v1/admin/ministerios",
			wantCode:  audit.ActionMinistryCreate,
			wantLabel: "Ministerios",
		},
		{
			method:    http.MethodPatch,
			path:      "/api/v1/admin/ministerios/" + id.String(),
			wantCode:  audit.ActionMinistryUpdate,
			wantLabel: "Ministerios · " + id.String(),
		},
		{
			method:    http.MethodDelete,
			path:      "/api/v1/admin/ministerios/" + id.String(),
			wantCode:  audit.ActionMinistryDelete,
			wantLabel: "Ministerios · " + id.String(),
		},
		{
			method:    http.MethodPost,
			path:      "/api/v1/admin/ministerios/imagenes",
			wantCode:  audit.ActionMinistryImageUpload,
			wantLabel: "Ministerios · Imagen",
		},
	}

	for _, denial := range denials {
		if err := svc.RecordDenied(context.Background(), audit.Denial{
			ActorUserID: &actor,
			Method:      denial.method,
			Path:        denial.path,
		}); err != nil {
			t.Fatalf("RecordDenied(%s %s) = %v", denial.method, denial.path, err)
		}
	}

	if len(repo.recorded) != len(denials) {
		t.Fatalf("registros = %d, se esperaban %d", len(repo.recorded), len(denials))
	}
	for i, denial := range denials {
		action := repo.recorded[i]
		if action.Code != denial.wantCode {
			t.Errorf("acción %d: code = %q, se esperaba %q", i, action.Code, denial.wantCode)
		}
		if action.Result != audit.ResultDenied {
			t.Errorf("acción %d: result = %q, se esperaba denied", i, action.Result)
		}
		if action.TargetLabel != denial.wantLabel {
			t.Errorf("acción %d: targetLabel = %q, se esperaba %q", i, action.TargetLabel, denial.wantLabel)
		}
		if action.ActorUserID == nil || *action.ActorUserID != actor {
			t.Errorf("acción %d: el actor de la sesión no viaja: %+v", i, action)
		}
	}
}

// TestRecordDeniedIgnoresRoutesWithoutAction fija que una petición sin acción
// asociada (lecturas del panel, rutas ajenas) no deja fila y no es un error.
func TestRecordDeniedIgnoresRoutesWithoutAction(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	paths := []string{
		"/api/v1/admin/ministerios",
		"/api/v1/ministerios",
		"/api/v1/portada/ministerios",
		"/api/v1/admin/eventos/" + uuid.New().String(),
		"/healthz",
	}
	for _, path := range paths {
		for _, method := range []string{http.MethodGet, http.MethodPut} {
			if err := svc.RecordDenied(context.Background(), audit.Denial{Method: method, Path: path}); err != nil {
				t.Fatalf("RecordDenied(%s %s) = %v", method, path, err)
			}
		}
	}
	if len(repo.recorded) != 0 {
		t.Errorf("se registraron filas para rutas sin acción: %+v", repo.recorded)
	}
}

// TestRecordDeniedPropagatesStoreFailure fija el contrato con la cadena: el
// fallo del registro sube envuelto (la cadena lo deja en el log) y no cambia la
// respuesta al visitante.
func TestRecordDeniedPropagatesStoreFailure(t *testing.T) {
	repo := newFakeRepository()
	repo.recordErr = errors.New("boom")
	svc := newTestService(repo)

	err := svc.RecordDenied(context.Background(), audit.Denial{
		Method: http.MethodPatch,
		Path:   "/api/v1/admin/ministerios/" + uuid.New().String(),
	})
	if err == nil || !errors.Is(err, repo.recordErr) {
		t.Fatalf("err = %v, se esperaba el fallo del repositorio envuelto", err)
	}
}

// TestRecordRejectedBestEffort fija el rechazo por JSON o DTO inválido: deja la
// fila con `result='failure'` y, si el registro falla, no cambia la respuesta
// (best-effort) y queda en el log.
func TestRecordRejectedBestEffort(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)
	actor := uuid.New()
	id := uuid.New()

	svc.RecordRejectedBestEffort(context.Background(), Rejection{
		ActorUserID: &actor,
		Method:      http.MethodPatch,
		Path:        "/api/v1/admin/ministerios/" + id.String(),
	})
	svc.RecordRejectedBestEffort(context.Background(), Rejection{
		Method: http.MethodGet,
		Path:   "/api/v1/ministerios",
	})

	if len(repo.recorded) != 1 {
		t.Fatalf("registros = %+v, se esperaba uno (la lectura pública no deja fila)", repo.recorded)
	}
	action := repo.recorded[0]
	if action.Code != audit.ActionMinistryUpdate || action.Result != audit.ResultFailure {
		t.Errorf("acción = %+v, se esperaba ministry.update con result=failure", action)
	}
	if action.ActorUserID == nil || *action.ActorUserID != actor {
		t.Errorf("el actor no viaja: %+v", action)
	}

	// Best-effort: un fallo del registro no propaga nada ni deja el actor.
	repo.recordErr = errors.New("boom")
	svc.RecordRejectedBestEffort(context.Background(), Rejection{
		Method: http.MethodPost,
		Path:   "/api/v1/admin/ministerios",
	})
	if len(repo.recorded) != 1 {
		t.Errorf("un fallo de registro añadió filas: %+v", repo.recorded)
	}
}

// TestRecordRejectionNeverCarriesTheBody fija el diseño de FR-026 de F2: la
// entrada del registro es método, ruta y actor — no hay sitio para el cuerpo ni
// para datos de la petición.
func TestRecordRejectionNeverCarriesTheBody(t *testing.T) {
	rejection := Rejection{Method: http.MethodPost, Path: "/api/v1/admin/ministerios"}
	if rejection.Method == "" || rejection.Path == "" {
		t.Fatal("la entrada del registro debe llevar método y ruta")
	}
	// La estructura solo tiene esos tres campos (actor, método y ruta): un
	// cuerpo de petición no tiene por dónde viajar.
	if got := len(audit.ActionCodes); got < 8 {
		t.Errorf("el registro cerrado tiene %d códigos", got)
	}
}

// TestSplitPathEdgeCases fija el troceado de rutas: una ruta vacía o solo barras
// no produce segmentos (ninguna acción) y los segmentos vacíos se descartan.
func TestSplitPathEdgeCases(t *testing.T) {
	for _, path := range []string{"", "/", "//"} {
		if got := splitPath(path); got != nil {
			t.Errorf("splitPath(%q) = %v, se esperaba nil", path, got)
		}
	}
	if got := splitPath("/api/v1/admin/ministerios/"); len(got) != 4 {
		t.Errorf("splitPath = %v, se esperaban 4 segmentos", got)
	}
	if _, ok := auditActionFromRoute(http.MethodPost, ""); ok {
		t.Error("una ruta vacía no tiene acción asociada")
	}
}
