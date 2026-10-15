package donaciones

import (
	"errors"
	"testing"
)

// TestNormalizarIban fija la validación y normalización de cuentas IBAN
// (FR-015/SC-009, R7-2/P7-2): quita espacios y guiones, pasa a mayúsculas y
// valida estructura (`^[A-Z]{2}[0-9]{2}[A-Z0-9]+$`) y longitud exacta por país
// del registro IBAN embebido (CR = 22, ES = 24, DE = 22, GB = 22…). No hay
// mod-97 (default revisable). El valor devuelto es el número normalizado que se
// guarda y no cambia al persistirlo.
func TestNormalizarIban(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"CR con espacios y guiones", "CR05 0152-0200-1234-5678-90", "CR05015202001234567890"},
		{"CR en minúsculas", "cr05 0152 0200 1234 5678 90", "CR05015202001234567890"},
		{"CR sin separadores", "CR05015202001234567890", "CR05015202001234567890"},
		{"ES (24)", "ES91 2100 0418 4502 0005 1332", "ES9121000418450200051332"},
		{"DE (22)", "DE89 3704 0044 0532 0130 00", "DE89370400440532013000"},
		{"GB (22) con letras en el BBAN", "GB29 NWBK 6016 1331 9268 19", "GB29NWBK60161331926819"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizarIban(tt.raw)
			if err != nil {
				t.Fatalf("NormalizarIban(%q) = %v, se esperaba nil", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizarIban(%q) = %q, se esperaba %q", tt.raw, got, tt.want)
			}
		})
	}

	invalid := []struct {
		name string
		raw  string
	}{
		{"CR de 21 (le falta uno)", "CR05 0152 0200 1234 5678 9"},
		{"CR de 23 (le sobra uno)", "CR05 0152 0200 1234 5678 901"},
		{"sufijo tras un guion (normaliza a 23)", "CR05 0152-0200-1234-5678-90-1"},
		{"país fuera del registro", "ZZ05 0152 0200 1234 5678 90"},
		{"letras en los dígitos de control", "CRAB 0152 0200 1234 5678 90"},
		{"emoji fuera de la estructura", "CR05 0152-0200-1234-5678-9😀"},
		{"texto libre", "hola"},
		{"vacío", ""},
		{"solo espacios", "   "},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizarIban(tt.raw)
			if !errors.Is(err, ErrNumeroInvalido) {
				t.Fatalf("NormalizarIban(%q) = (%q, %v), se esperaba %v", tt.raw, got, err, ErrNumeroInvalido)
			}
			if got != "" {
				t.Fatalf("NormalizarIban(%q) devolvió %q con error, se esperaba \"\"", tt.raw, got)
			}
		})
	}
}

// TestNormalizarSinpe fija la validación y normalización de SINPE Móvil
// (FR-016/SC-009, R7-3/P7-4): acepta con o sin `+506` y con espacios/guiones,
// exige 8 dígitos con primer dígito 6/7/8 (celular, no línea fija) y devuelve
// `+506XXXXXXXX`.
func TestNormalizarSinpe(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"8 dígitos", "88888888", "+50688888888"},
		{"con guion", "8888-8888", "+50688888888"},
		{"con +506 y espacios", "+506 8888 8888", "+50688888888"},
		{"con +506 pegado", "+50688888888", "+50688888888"},
		{"prefijo 7", "70000000", "+50670000000"},
		{"prefijo 6", "60000000", "+50660000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizarSinpe(tt.raw)
			if err != nil {
				t.Fatalf("NormalizarSinpe(%q) = %v, se esperaba nil", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizarSinpe(%q) = %q, se esperaba %q", tt.raw, got, tt.want)
			}
		})
	}

	invalid := []struct {
		name string
		raw  string
	}{
		{"7 dígitos", "8888888"},
		{"9 dígitos", "888888888"},
		{"línea fija (2xxx)", "22222222"},
		{"letras", "8888-888A"},
		{"otro país (+507)", "+507 8888 8888"},
		{"sufijo extra", "88888888 8"},
		{"vacío", ""},
		{"solo prefijo", "+506"},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizarSinpe(tt.raw)
			if !errors.Is(err, ErrNumeroInvalido) {
				t.Fatalf("NormalizarSinpe(%q) = (%q, %v), se esperaba %v", tt.raw, got, err, ErrNumeroInvalido)
			}
			if got != "" {
				t.Fatalf("NormalizarSinpe(%q) devolvió %q con error, se esperaba \"\"", tt.raw, got)
			}
		})
	}
}

// TestValidarIbanYSinpe comprueba que las variantes `Validar*` delegan en la
// misma regla (una sola implementación pura).
func TestValidarIbanYSinpe(t *testing.T) {
	if err := ValidarIban("CR05015202001234567890"); err != nil {
		t.Fatalf("ValidarIban válido = %v, se esperaba nil", err)
	}
	if err := ValidarIban("CR0501520200123456789"); !errors.Is(err, ErrNumeroInvalido) {
		t.Fatalf("ValidarIban inválido = %v, se esperaba %v", err, ErrNumeroInvalido)
	}
	if err := ValidarSinpe("+50688888888"); err != nil {
		t.Fatalf("ValidarSinpe válido = %v, se esperaba nil", err)
	}
	if err := ValidarSinpe("22222222"); !errors.Is(err, ErrNumeroInvalido) {
		t.Fatalf("ValidarSinpe inválido = %v, se esperaba %v", err, ErrNumeroInvalido)
	}
}
