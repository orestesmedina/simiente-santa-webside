package database

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// ctxKey es una clave privada para comprobar que el contexto se propaga.
type ctxKey struct{}

// fakePinger registra el contexto con el que se le llamó para inspeccionarlo en
// las pruebas.
type fakePinger struct {
	err       error
	called    bool
	deadline  time.Time
	hasValue  any
	valueSeen bool
}

func (f *fakePinger) Ping(ctx context.Context) error {
	f.called = true
	f.deadline, _ = ctx.Deadline()
	f.hasValue = ctx.Value(ctxKey{})
	f.valueSeen = f.hasValue != nil
	return f.err
}

func TestPingTimeoutIsTwoSeconds(t *testing.T) {
	if PingTimeout != 2*time.Second {
		t.Errorf("PingTimeout = %v, se esperaba 2s", PingTimeout)
	}
}

func TestPingAppliesDeadlineAndPropagatesContext(t *testing.T) {
	fake := &fakePinger{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "valor-padre")

	before := time.Now()
	err := ping(ctx, fake)
	after := time.Now()

	if err != nil {
		t.Fatalf("ping() error: %v", err)
	}
	if !fake.called {
		t.Fatal("no se llamó a Ping del pinger")
	}
	if fake.deadline.IsZero() {
		t.Fatal("el contexto que recibió el pinger no tiene deadline")
	}
	wantMin := before.Add(PingTimeout)
	wantMax := after.Add(PingTimeout)
	if fake.deadline.Before(wantMin) || fake.deadline.After(wantMax) {
		t.Errorf("deadline = %v, se esperaba entre %v y %v", fake.deadline, wantMin, wantMax)
	}
	if !fake.valueSeen || fake.hasValue != "valor-padre" {
		t.Errorf("el valor del contexto no se propagó: %v", fake.hasValue)
	}
}

func TestPingWrapsErrorWithContext(t *testing.T) {
	sentinel := errors.New("connection refused")
	fake := &fakePinger{err: sentinel}

	err := ping(context.Background(), fake)
	if err == nil {
		t.Fatal("ping() no devolvió error")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is no reconoce la causa envuelta con %%w: %v", err)
	}
	if !strings.Contains(err.Error(), "ping database") {
		t.Errorf("el error no aporta contexto: %q", err.Error())
	}
}

func TestNewPoolRejectsInvalidURL(t *testing.T) {
	if _, err := NewPool(context.Background(), "no es una url"); err == nil {
		t.Fatal("NewPool() no devolvió error con una URL inválida")
	}
}

// TestNewPoolIsLazy comprueba que construir el pool no abre conexión: con un
// host inalcanzable, NewPool devuelve el pool sin bloquear ni fallar.
func TestNewPoolIsLazy(t *testing.T) {
	pool, err := NewPool(context.Background(), "postgres://app:app@127.0.0.1:1/app?sslmode=disable")
	if err != nil {
		t.Fatalf("NewPool() error: %v", err)
	}
	if pool == nil {
		t.Fatal("NewPool() devolvió un pool nil")
	}
	pool.Close()
}
