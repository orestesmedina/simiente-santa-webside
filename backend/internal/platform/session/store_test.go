package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNextTTLUsesIdleWhileAbsoluteRemains(t *testing.T) {
	now := time.Now().UTC()
	session := Session{
		UserID:            uuid.New(),
		CreatedAt:         now,
		LastSeenAt:        now,
		AbsoluteExpiresAt: now.Add(time.Hour),
	}

	if got := nextTTL(now, session, 30*time.Minute); got != 30*time.Minute {
		t.Errorf("nextTTL() = %v, se esperaba 30m", got)
	}
}

func TestNextTTLCapsAtAbsoluteExpiry(t *testing.T) {
	now := time.Now().UTC()
	session := Session{
		UserID:            uuid.New(),
		CreatedAt:         now.Add(-59 * time.Minute),
		LastSeenAt:        now,
		AbsoluteExpiresAt: now.Add(10 * time.Second),
	}

	if got := nextTTL(now, session, 30*time.Minute); got != 10*time.Second {
		t.Errorf("nextTTL() = %v, se esperaba 10s (acotado a la vida absoluta)", got)
	}
}

func TestSessionExpiredAt(t *testing.T) {
	now := time.Now().UTC()
	session := Session{AbsoluteExpiresAt: now.Add(time.Minute)}

	if session.ExpiredAt(now) {
		t.Error("ExpiredAt(now) = true, se esperaba false")
	}
	if !session.ExpiredAt(now.Add(time.Minute)) {
		t.Error("ExpiredAt(absolute) = false, se esperaba true")
	}
	if !session.ExpiredAt(now.Add(2 * time.Minute)) {
		t.Error("ExpiredAt(pasado) = false, se esperaba true")
	}
}

func TestStoreConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     StoreConfig
		wantErr bool
	}{
		{name: "válida", cfg: StoreConfig{IdleTTL: 30 * time.Minute, AbsoluteTTL: time.Hour}},
		{name: "idle cero", cfg: StoreConfig{AbsoluteTTL: time.Hour}, wantErr: true},
		{name: "idle negativa", cfg: StoreConfig{IdleTTL: -time.Minute, AbsoluteTTL: time.Hour}, wantErr: true},
		{name: "absolute cero", cfg: StoreConfig{IdleTTL: 30 * time.Minute}, wantErr: true},
		{name: "absolute negativa", cfg: StoreConfig{IdleTTL: 30 * time.Minute, AbsoluteTTL: -time.Hour}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() = nil, se esperaba error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() = %v, se esperaba nil", err)
			}
		})
	}
}

func TestStoreConfigLastSeenThrottle(t *testing.T) {
	if got := (StoreConfig{}).lastSeenThrottle(); got != DefaultLastSeenThrottle {
		t.Errorf("lastSeenThrottle() por defecto = %v, se esperaba %v", got, DefaultLastSeenThrottle)
	}
	if got := (StoreConfig{LastSeenThrottle: 5 * time.Second}).lastSeenThrottle(); got != 5*time.Second {
		t.Errorf("lastSeenThrottle() = %v, se esperaba 5s", got)
	}
}
