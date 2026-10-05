package session

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestNewTokenDistinctAndShape(t *testing.T) {
	const samples = 200
	seen := make(map[string]struct{}, samples)

	for i := 0; i < samples; i++ {
		token, err := NewToken()
		if err != nil {
			t.Fatalf("NewToken() error: %v", err)
		}
		if token == "" {
			t.Fatal("NewToken() devolvió un token vacío")
		}
		if _, dup := seen[token]; dup {
			t.Fatalf("NewToken() repitió un token: %q", token)
		}
		seen[token] = struct{}{}

		raw, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			t.Fatalf("el token %q no es base64url sin relleno: %v", token, err)
		}
		if len(raw) != TokenBytes {
			t.Fatalf("el token decodificado mide %d bytes, se esperaban %d", len(raw), TokenBytes)
		}
	}
}

func TestHashTokenStableAndKnown(t *testing.T) {
	token := "token-de-prueba"
	sum := sha256.Sum256([]byte(token))
	want := hex.EncodeToString(sum[:])

	got := HashToken(token)
	if got != want {
		t.Errorf("HashToken() = %q, se esperaba %q", got, want)
	}
	if again := HashToken(token); again != got {
		t.Errorf("HashToken() no es estable: %q != %q", again, got)
	}
	if len(got) != 64 {
		t.Errorf("HashToken() mide %d caracteres, se esperaban 64", len(got))
	}
	if got == token {
		t.Error("HashToken() devolvió el token en claro")
	}
	if HashToken("otro-token") == got {
		t.Error("HashToken() colisionó para tokens distintos")
	}
}
