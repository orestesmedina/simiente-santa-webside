package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testSecret = "secreto-de-prueba-para-el-hmac"

func TestNewSessionCookieAttributes(t *testing.T) {
	cfg := NewCookieConfig(true, time.Hour)
	cookie := NewSessionCookie(cfg, "token-abc")

	if cookie.Name != CookieSession {
		t.Errorf("Name = %q, se esperaba %q", cookie.Name, CookieSession)
	}
	if cookie.Value != "token-abc" {
		t.Errorf("Value = %q, se esperaba token-abc", cookie.Value)
	}
	if cookie.Path != "/" {
		t.Errorf("Path = %q, se esperaba /", cookie.Path)
	}
	if cookie.MaxAge != 3600 {
		t.Errorf("MaxAge = %d, se esperaba 3600", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Error("HttpOnly = false, se esperaba true (P1)")
	}
	if !cookie.Secure {
		t.Error("Secure = false, se esperaba true")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, se esperaba Lax", cookie.SameSite)
	}
	if cookie.Domain != "" {
		t.Errorf("Domain = %q, se esperaba host-only (vacío)", cookie.Domain)
	}
}

func TestNewCSRFCookieAttributes(t *testing.T) {
	cfg := NewCookieConfig(false, 30*time.Minute)
	cookie := NewCSRFCookie(cfg, "nonce.firma")

	if cookie.Name != CookieCSRF {
		t.Errorf("Name = %q, se esperaba %q", cookie.Name, CookieCSRF)
	}
	if cookie.Value != "nonce.firma" {
		t.Errorf("Value = %q, se esperaba nonce.firma", cookie.Value)
	}
	if cookie.HttpOnly {
		t.Error("HttpOnly = true, se esperaba false: la SPA debe leer la cookie (P10)")
	}
	if cookie.Secure {
		t.Error("Secure = true, se esperaba false con la configuración de desarrollo")
	}
	if cookie.MaxAge != 1800 {
		t.Errorf("MaxAge = %d, se esperaba 1800", cookie.MaxAge)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, se esperaba Lax", cookie.SameSite)
	}
}

func TestClearCookiesDelete(t *testing.T) {
	cfg := NewCookieConfig(true, time.Hour)

	for name, cookie := range map[string]*http.Cookie{
		"sesión": ClearSessionCookie(cfg),
		"csrf":   ClearCSRFCookie(cfg),
	} {
		if cookie.MaxAge >= 0 {
			t.Errorf("%s: MaxAge = %d, se esperaba negativo", name, cookie.MaxAge)
		}
		if cookie.Value != "" {
			t.Errorf("%s: Value = %q, se esperaba vacío", name, cookie.Value)
		}
		if cookie.Path != "/" {
			t.Errorf("%s: Path = %q, se esperaba /", name, cookie.Path)
		}
		if cookie.Expires.IsZero() {
			t.Errorf("%s: Expires no debería ser cero", name)
		}
	}
}

func TestNewCSRFTokenAndVerify(t *testing.T) {
	token, err := NewCSRFToken(testSecret)
	if err != nil {
		t.Fatalf("NewCSRFToken() error: %v", err)
	}
	nonce, signature, found := strings.Cut(token, ".")
	if !found || nonce == "" || signature == "" {
		t.Fatalf("NewCSRFToken() = %q, se esperaba nonce.firma", token)
	}
	if !VerifyCSRFToken(testSecret, token) {
		t.Error("VerifyCSRFToken() = false para un token recién generado")
	}

	other, err := NewCSRFToken(testSecret)
	if err != nil {
		t.Fatalf("NewCSRFToken() error: %v", err)
	}
	if other == token {
		t.Error("NewCSRFToken() repitió el nonce")
	}

	tests := []struct {
		name   string
		secret string
		token  string
	}{
		{name: "secreto distinto", secret: "otro-secreto", token: token},
		{name: "firma alterada", secret: testSecret, token: nonce + ".firma-falsa"},
		{name: "nonce alterado", secret: testSecret, token: nonce + "x." + signature},
		{name: "sin separador", secret: testSecret, token: nonce + signature},
		{name: "vacío", secret: testSecret, token: ""},
		{name: "solo nonce", secret: testSecret, token: nonce + "."},
		{name: "solo firma", secret: testSecret, token: "." + signature},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if VerifyCSRFToken(tt.secret, tt.token) {
				t.Errorf("VerifyCSRFToken() = true para %q", tt.token)
			}
		})
	}
}

func TestValidateCSRF(t *testing.T) {
	token, err := NewCSRFToken(testSecret)
	if err != nil {
		t.Fatalf("NewCSRFToken() error: %v", err)
	}

	tests := []struct {
		name         string
		cookie, head string
		want         bool
	}{
		{name: "par válido", cookie: token, head: token, want: true},
		{name: "cabecera distinta", cookie: token, head: "otro", want: false},
		{name: "cookie inválida", cookie: "nonce.firma-falsa", head: "nonce.firma-falsa", want: false},
		{name: "cookie vacía", cookie: "", head: token, want: false},
		{name: "cabecera vacía", cookie: token, head: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCSRF(testSecret, tt.cookie, tt.head); got != tt.want {
				t.Errorf("ValidateCSRF() = %v, se esperaba %v", got, tt.want)
			}
		})
	}
}

func TestCookiesFromRequest(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: CookieSession, Value: "tok-sesion"})
	request.AddCookie(&http.Cookie{Name: CookieCSRF, Value: "nonce.firma"})

	if got, ok := SessionTokenFromRequest(request); !ok || got != "tok-sesion" {
		t.Errorf("SessionTokenFromRequest() = %q, %v; se esperaba tok-sesion, true", got, ok)
	}
	if got, ok := CSRFFromRequest(request); !ok || got != "nonce.firma" {
		t.Errorf("CSRFFromRequest() = %q, %v; se esperaba nonce.firma, true", got, ok)
	}

	empty := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, ok := SessionTokenFromRequest(empty); ok {
		t.Error("SessionTokenFromRequest() sin cookie = true, se esperaba false")
	}
	if _, ok := CSRFFromRequest(empty); ok {
		t.Error("CSRFFromRequest() sin cookie = true, se esperaba false")
	}
	if _, ok := SessionTokenFromRequest(nil); ok {
		t.Error("SessionTokenFromRequest(nil) = true, se esperaba false")
	}

	blank := httptest.NewRequest(http.MethodGet, "/", nil)
	blank.AddCookie(&http.Cookie{Name: CookieSession, Value: ""})
	if _, ok := SessionTokenFromRequest(blank); ok {
		t.Error("SessionTokenFromRequest() con cookie vacía = true, se esperaba false")
	}
}
