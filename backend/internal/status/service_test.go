package status

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"simiente-santa/backend/internal/platform/apperr"
)

// ctxKey es una clave privada para comprobar que el contexto se propaga hasta
// el repositorio.
type ctxKey struct{}

// fakeRepository guioniza la respuesta de Ping y registra el contexto con el
// que se le llamó, para verificar la propagación y la cancelación.
type fakeRepository struct {
	err      error
	called   bool
	ctxErr   error
	ctxValue any
}

func (f *fakeRepository) Ping(ctx context.Context) error {
	f.called = true
	f.ctxErr = ctx.Err()
	f.ctxValue = ctx.Value(ctxKey{})
	return f.err
}

func TestServiceStatus(t *testing.T) {
	sentinel := errors.New("dial tcp 127.0.0.1:5432: connection refused")

	tests := []struct {
		name         string
		repoErr      error
		want         SystemStatus
		wantCode     string
		wantInternal bool
		wantCause    error
	}{
		{
			name:    "conexión viva",
			repoErr: nil,
			want:    SystemStatus{Status: StatusOK, Database: DatabaseConnected},
		},
		{
			name:      "base de datos caída",
			repoErr:   sentinel,
			wantCode:  "database_unavailable",
			wantCause: sentinel,
		},
		{
			name:         "error inesperado del repositorio",
			repoErr:      context.DeadlineExceeded,
			wantInternal: true,
			wantCause:    context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{err: tt.repoErr}
			svc := NewService(repo)

			got, err := svc.Status(context.Background())

			if tt.repoErr == nil {
				if err != nil {
					t.Fatalf("Status() error inesperado: %v", err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Status() = %+v, se esperaba %+v", got, tt.want)
				}
				return
			}

			if err == nil {
				t.Fatal("Status() no devolvió error")
			}
			if !errors.Is(err, tt.wantCause) {
				t.Errorf("errors.Is no reconoce la causa envuelta con %%w: %v", err)
			}

			var domainErr *apperr.Error
			isDomain := errors.As(err, &domainErr)

			if tt.wantInternal {
				if isDomain {
					t.Errorf("Status() devolvió un error de dominio donde se esperaba uno inesperado: %v", err)
				}
				return
			}

			if !isDomain {
				t.Fatalf("Status() no devolvió *apperr.Error: %v", err)
			}
			if domainErr.Code() != tt.wantCode {
				t.Errorf("Code() = %q, se esperaba %q", domainErr.Code(), tt.wantCode)
			}
			if got := domainErr.Details[DetailDatabaseKey]; got != DatabaseDisconnected {
				t.Errorf("details.database = %v, se esperaba %q", got, DatabaseDisconnected)
			}
		})
	}
}

func TestServiceStatusPropagatesContext(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	ctx := context.WithValue(context.Background(), ctxKey{}, "valor-padre")

	if _, err := svc.Status(ctx); err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if !repo.called {
		t.Fatal("no se llamó a Ping del repositorio")
	}
	if repo.ctxValue != "valor-padre" {
		t.Errorf("el valor del contexto no llegó al repositorio: %v", repo.ctxValue)
	}
}

func TestServiceStatusRespectsCancelledContext(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Status(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Status() error = %v, se esperaba context.Canceled envuelto", err)
	}
	var domainErr *apperr.Error
	if errors.As(err, &domainErr) {
		t.Errorf("una cancelación no debe traducirse a error de dominio: %v", err)
	}
	if repo.called {
		t.Error("no debe consultarse la base de datos con el contexto ya cancelado")
	}
}
