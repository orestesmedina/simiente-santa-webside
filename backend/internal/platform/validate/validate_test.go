package validate

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"simiente-santa/backend/internal/platform/apperr"
)

// invalidDetails comprueba que el error es un *apperr.Error de kind Invalid y
// devuelve sus details para inspeccionarlos.
func invalidDetails(t *testing.T, err error) map[string]any {
	t.Helper()
	if err == nil {
		t.Fatal("se esperaba un error de validación, se obtuvo nil")
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("el error no es *apperr.Error: %T", err)
	}
	if domainErr.Code() != "invalid" {
		t.Errorf("code = %q, se esperaba invalid", domainErr.Code())
	}
	if got := domainErr.HTTPStatus(); got != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba 400", got)
	}
	if domainErr.Message != MessageInvalid {
		t.Errorf("message = %q, se esperaba %q", domainErr.Message, MessageInvalid)
	}
	return domainErr.Details
}

func TestRequired(t *testing.T) {
	type dto struct {
		Name string `json:"name" validate:"required"`
	}

	valid := []dto{{Name: "Ana"}, {Name: " Ana "}}
	for _, in := range valid {
		if err := Struct(in); err != nil {
			t.Errorf("Struct(%+v) = %v, se esperaba nil", in, err)
		}
	}

	invalid := []dto{{Name: ""}, {Name: "   "}}
	for _, in := range invalid {
		details := invalidDetails(t, Struct(in))
		msg, ok := details["name"].(string)
		if !ok || !strings.Contains(msg, "obligatorio") {
			t.Errorf("details[name] = %v, se esperaba un mensaje de campo obligatorio", details["name"])
		}
	}
}

func TestOmitempty(t *testing.T) {
	type dto struct {
		Phone string `json:"phone" validate:"omitempty,max=5"`
	}

	if err := Struct(dto{Phone: ""}); err != nil {
		t.Errorf("omitempty con campo vacío = %v, se esperaba nil", err)
	}
	if err := Struct(dto{Phone: "12345"}); err != nil {
		t.Errorf("omitempty con valor válido = %v, se esperaba nil", err)
	}

	details := invalidDetails(t, Struct(dto{Phone: "123456"}))
	if _, ok := details["phone"]; !ok {
		t.Errorf("details = %v, se esperaba la clave phone", details)
	}
}

func TestMinAndMaxOnText(t *testing.T) {
	type dto struct {
		FirstName string `json:"firstName" validate:"required,min=2,max=5"`
	}

	if err := Struct(dto{FirstName: "Ana"}); err != nil {
		t.Errorf("valor dentro de rango = %v, se esperaba nil", err)
	}
	if err := Struct(dto{FirstName: "An"}); err != nil {
		t.Errorf("valor en el mínimo = %v, se esperaba nil", err)
	}

	details := invalidDetails(t, Struct(dto{FirstName: "A"}))
	if msg, _ := details["firstName"].(string); !strings.Contains(msg, "al menos 2") {
		t.Errorf("details[firstName] = %v, se esperaba el aviso de mínimo", details["firstName"])
	}

	details = invalidDetails(t, Struct(dto{FirstName: "Anastasia"}))
	if msg, _ := details["firstName"].(string); !strings.Contains(msg, "más de 5") {
		t.Errorf("details[firstName] = %v, se esperaba el aviso de máximo", details["firstName"])
	}
}

func TestMinAndMaxOnNumbers(t *testing.T) {
	type dto struct {
		Level int `json:"level" validate:"min=1,max=10"`
	}

	if err := Struct(dto{Level: 5}); err != nil {
		t.Errorf("valor dentro de rango = %v, se esperaba nil", err)
	}

	if details := invalidDetails(t, Struct(dto{Level: 0})); details["level"] == nil {
		t.Errorf("details = %v, se esperaba la clave level", details)
	}
	if details := invalidDetails(t, Struct(dto{Level: 11})); details["level"] == nil {
		t.Errorf("details = %v, se esperaba la clave level", details)
	}
}

func TestEmail(t *testing.T) {
	type dto struct {
		Email string `json:"email" validate:"required,email,max=254"`
	}

	valid := []string{"ana@ejemplo.com", "carlos.ayudante@sub.ejemplo.org"}
	for _, email := range valid {
		if err := Struct(dto{Email: email}); err != nil {
			t.Errorf("Struct(email=%q) = %v, se esperaba nil", email, err)
		}
	}

	invalid := []string{"no-es-correo", "ana@ejemplo", "ana @ejemplo.com", "@ejemplo.com", "ana@.com"}
	for _, email := range invalid {
		details := invalidDetails(t, Struct(dto{Email: email}))
		if msg, _ := details["email"].(string); !strings.Contains(msg, "correo") {
			t.Errorf("details[email] para %q = %v, se esperaba un aviso de correo", email, details["email"])
		}
	}
}

func TestOneOf(t *testing.T) {
	type dto struct {
		Result string `json:"result" validate:"required,oneof=success failure denied"`
	}

	for _, result := range []string{"success", "failure", "denied"} {
		if err := Struct(dto{Result: result}); err != nil {
			t.Errorf("Struct(result=%q) = %v, se esperaba nil", result, err)
		}
	}

	details := invalidDetails(t, Struct(dto{Result: "other"}))
	msg, _ := details["result"].(string)
	if !strings.Contains(msg, "success") || !strings.Contains(msg, "denied") {
		t.Errorf("details[result] = %q, se esperaban las opciones válidas", msg)
	}
}

func TestPhone(t *testing.T) {
	type dto struct {
		Phone string `json:"phone" validate:"required,phone,max=32"`
	}

	valid := []string{
		"+34 612 345 678",
		"612 345 678",
		"(555) 123-4567",
		"+1 (555) 123-4567",
		"1234567",
	}
	for _, phone := range valid {
		if err := Struct(dto{Phone: phone}); err != nil {
			t.Errorf("Struct(phone=%q) = %v, se esperaba nil", phone, err)
		}
	}

	invalid := []string{"12", "no es un teléfono", "123456", "() -", "612.345.678"}
	for _, phone := range invalid {
		details := invalidDetails(t, Struct(dto{Phone: phone}))
		msg, ok := details["phone"].(string)
		if !ok || !strings.Contains(msg, "teléfono") {
			t.Errorf("details[phone] para %q = %v, se esperaba un aviso de teléfono", phone, details["phone"])
		}
	}
}

func TestValidDTOProducesNil(t *testing.T) {
	type dto struct {
		FirstName string `json:"firstName" validate:"required,min=1,max=120"`
		LastName  string `json:"lastName" validate:"required,min=1,max=120"`
		Email     string `json:"email" validate:"required,email,max=254"`
		Phone     string `json:"phone" validate:"required,phone,max=32"`
		RoleID    string `json:"roleId" validate:"required"`
		Password  string `json:"password" validate:"required,min=8,max=64"`
	}

	in := dto{
		FirstName: "Ana",
		LastName:  "Responsable",
		Email:     "ana@ejemplo.com",
		Phone:     "+34 612 345 678",
		RoleID:    "b0f0e4c2-0000-0000-0000-000000000000",
		Password:  "Semilla.2026",
	}
	if err := Struct(in); err != nil {
		t.Errorf("DTO válido = %v, se esperaba nil", err)
	}
	if err := Struct(&in); err != nil {
		t.Errorf("puntero a DTO válido = %v, se esperaba nil", err)
	}
}

func TestMultipleInvalidFieldsReportedTogether(t *testing.T) {
	type dto struct {
		FirstName string `json:"firstName" validate:"required"`
		Email     string `json:"email" validate:"required,email"`
		Phone     string `json:"phone" validate:"required,phone"`
	}

	details := invalidDetails(t, Struct(dto{FirstName: "", Email: "malo", Phone: "12"}))
	for _, key := range []string{"firstName", "email", "phone"} {
		if _, ok := details[key]; !ok {
			t.Errorf("details = %v, se esperaba la clave %q", details, key)
		}
	}
}

func TestDetailsUseJSONFieldName(t *testing.T) {
	type dto struct {
		RoleID string `json:"roleId" validate:"required"`
	}

	details := invalidDetails(t, Struct(dto{RoleID: ""}))
	if _, ok := details["roleId"]; !ok {
		t.Errorf("details = %v, se esperaba la clave roleId (nombre JSON)", details)
	}
	if _, ok := details["RoleID"]; ok {
		t.Errorf("details = %v, no debería usar el nombre del campo Go", details)
	}
}

func TestOmitemptyEmailSkipsEmptyValue(t *testing.T) {
	type dto struct {
		Phone string `json:"phone" validate:"omitempty,email"`
	}
	if err := Struct(dto{Phone: ""}); err != nil {
		t.Errorf("omitempty,email con vacío = %v, se esperaba nil", err)
	}
}

func TestNonStructValuesProduceNil(t *testing.T) {
	type dto struct {
		Name string `json:"name" validate:"required"`
	}

	if err := Struct(nil); err != nil {
		t.Errorf("Struct(nil) = %v, se esperaba nil", err)
	}
	if err := Struct("texto"); err != nil {
		t.Errorf("Struct(string) = %v, se esperaba nil", err)
	}
	if err := Struct((*dto)(nil)); err != nil {
		t.Errorf("Struct(puntero nil) = %v, se esperaba nil", err)
	}
}

func TestFieldWithoutTagIsIgnored(t *testing.T) {
	type dto struct {
		Internal string `json:"internal"`
	}
	if err := Struct(dto{}); err != nil {
		t.Errorf("campo sin etiqueta = %v, se esperaba nil", err)
	}
}
