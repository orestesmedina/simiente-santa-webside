package password

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"simiente-santa/backend/internal/platform/apperr"
)

func TestValidate(t *testing.T) {
	long := strings.Repeat("Abc1!", 13)          // 65 caracteres, cumple las clases.
	multibyte := strings.Repeat("𐐀", 18) + "a1!" // 21 caracteres, 75 bytes.

	tests := []struct {
		name        string
		value       string
		ctx         Context
		wantKeyword string // cadena vacía = válida
	}{
		{name: "válida sin contexto", value: "Abcdef1!", wantKeyword: ""},
		{name: "válida contiene el nombre", value: "Anais123!", ctx: Context{FirstName: "Ana"}, wantKeyword: ""},
		{name: "válida con longitud máxima", value: strings.Repeat("Abc1!", 12) + "A!", wantKeyword: ""},
		{name: "demasiado corta", value: "Ab1!a", wantKeyword: "al menos 8 caracteres"},
		{name: "demasiado larga", value: long, wantKeyword: "más de 64 caracteres"},
		{name: "demasiados bytes para bcrypt", value: multibyte, wantKeyword: "almacenamiento seguro"},
		{name: "sin mayúscula", value: "abcdefg1!", wantKeyword: "mayúscula"},
		{name: "sin minúscula", value: "ABCDEFG1!", wantKeyword: "minúscula"},
		{name: "sin número", value: "Abcdefgh!", wantKeyword: "un número"},
		{name: "sin carácter especial", value: "Abcdefg1", wantKeyword: "carácter especial"},
		{name: "igual al nombre", value: "Anais123!", ctx: Context{FirstName: "Anais123!"}, wantKeyword: "tu nombre"},
		{name: "igual a los apellidos", value: "Pérez1!", ctx: Context{LastName: "Pérez1!"}, wantKeyword: "tus apellidos"},
		{name: "igual al correo", value: "Ana1@x.com", ctx: Context{Email: "Ana1@x.com"}, wantKeyword: "tu correo"},
		{name: "igual al nombre normalizado", value: "  ANAIS123!  ", ctx: Context{FirstName: "anais123!"}, wantKeyword: "tu nombre"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.value, tt.ctx)
			if tt.wantKeyword == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, se esperaba nil", err)
				}
				return
			}
			assertPolicyError(t, err, tt.wantKeyword)
		})
	}
}

func TestHashUsesCost12(t *testing.T) {
	hash, err := Hash("Abcdef1!", Context{})
	if err != nil {
		t.Fatalf("Hash() error inesperado: %v", err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt.Cost() error: %v", err)
	}
	if cost != bcryptCost {
		t.Errorf("cost = %d, se esperaba %d", cost, bcryptCost)
	}
	if hash == "Abcdef1!" {
		t.Error("Hash() devolvió la contraseña sin hashear")
	}
}

func TestHashUsesFreshSalt(t *testing.T) {
	first, err := Hash("Abcdef1!", Context{})
	if err != nil {
		t.Fatalf("primer Hash() error: %v", err)
	}
	second, err := Hash("Abcdef1!", Context{})
	if err != nil {
		t.Fatalf("segundo Hash() error: %v", err)
	}
	if first == second {
		t.Error("dos hashes de la misma contraseña son iguales: falta la sal por llamada")
	}
}

func TestHashRejectsPolicyViolation(t *testing.T) {
	_, err := Hash("corta", Context{})
	if err == nil {
		t.Fatal("Hash() = nil, se esperaba que aplicara la política")
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) || domainErr.Kind != apperr.KindInvalid {
		t.Fatalf("error = %v, se esperaba apperr.Invalid", err)
	}
}

func TestVerify(t *testing.T) {
	hash, err := Hash("Abcdef1!", Context{})
	if err != nil {
		t.Fatalf("Hash() error inesperado: %v", err)
	}

	if err := Verify(hash, "Abcdef1!"); err != nil {
		t.Errorf("Verify() con la contraseña correcta = %v, se esperaba nil", err)
	}
	if err := Verify(hash, "Abcdef2!"); err == nil {
		t.Error("Verify() con la contraseña incorrecta = nil, se esperaba error")
	}
}

// assertPolicyError comprueba que err es un apperr.Invalid con el requisito
// incumplido en Details["newPassword"].
func assertPolicyError(t *testing.T, err error, wantKeyword string) {
	t.Helper()
	if err == nil {
		t.Fatal("Validate() = nil, se esperaba un error de política")
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("error = %T, se esperaba *apperr.Error", err)
	}
	if domainErr.Kind != apperr.KindInvalid {
		t.Errorf("Kind = %v, se esperaba KindInvalid", domainErr.Kind)
	}
	if domainErr.Code() != "invalid" {
		t.Errorf("Code() = %q, se esperaba invalid", domainErr.Code())
	}
	detail, ok := domainErr.Details[detailKey]
	if !ok {
		t.Fatalf("Details = %v, falta la clave %q", domainErr.Details, detailKey)
	}
	if !strings.Contains(detail.(string), wantKeyword) {
		t.Errorf("Details[%q] = %q, se esperaba que contuviera %q", detailKey, detail, wantKeyword)
	}
}
