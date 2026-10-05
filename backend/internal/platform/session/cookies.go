package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Nombres de las cookies y de la cabecera del double-submit firmado (P10).
const (
	// CookieSession es la cookie de sesión (HttpOnly, SameSite=Lax, Secure
	// configurable, Path=/). Su valor es el token opaco de NewToken.
	CookieSession = "ss_session"
	// CookieCSRF es la cookie del token anti-CSRF (no HttpOnly: la SPA debe
	// poder leerla para copiarla en la cabecera HeaderCSRF).
	CookieCSRF = "csrf_token"
	// HeaderCSRF es la cabecera que el cliente envía con el mismo valor que la
	// cookie CookieCSRF en todo método no seguro.
	HeaderCSRF = "X-CSRF-Token"

	cookiePath     = "/"
	csrfSeparator  = "."
	csrfNonceBytes = 32
)

// CookieConfig agrupa los atributos configurables de las cookies de sesión:
// `Secure` (SESSION_COOKIE_SECURE) y el Max-Age, fijado a la vida absoluta de
// la sesión (R2/R15), de modo que el navegador descarte la cookie justo cuando
// el servidor la daría por vencida.
type CookieConfig struct {
	// Secure marca las cookies como Secure; true cuando hay TLS.
	Secure bool
	// AbsoluteTTL es la vida absoluta de la sesión y, por tanto, el Max-Age de
	// las cookies.
	AbsoluteTTL time.Duration
}

// NewCookieConfig construye la configuración de cookies a partir de la
// configuración de la aplicación (platform/config).
func NewCookieConfig(secure bool, absoluteTTL time.Duration) CookieConfig {
	return CookieConfig{Secure: secure, AbsoluteTTL: absoluteTTL}
}

// maxAgeSeconds traduce la vida absoluta a segundos de Max-Age (0 = sesión de
// navegador; solo ocurre con una configuración inválida).
func (c CookieConfig) maxAgeSeconds() int {
	if c.AbsoluteTTL <= 0 {
		return 0
	}
	return int(c.AbsoluteTTL.Seconds())
}

// NewSessionCookie construye la cookie ss_session con el token indicado:
// HttpOnly, SameSite=Lax, Path=/ y Secure según la configuración (P1/R2). No
// lleva Domain: es host-only.
func NewSessionCookie(cfg CookieConfig, token string) *http.Cookie {
	return &http.Cookie{
		Name:     CookieSession,
		Value:    token,
		Path:     cookiePath,
		MaxAge:   cfg.maxAgeSeconds(),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// NewCSRFCookie construye la cookie csrf_token con el token firmado indicado.
// A diferencia de la de sesión, NO es HttpOnly (la SPA la lee para el
// double-submit) y comparte el resto de atributos (P10).
func NewCSRFCookie(cfg CookieConfig, token string) *http.Cookie {
	return &http.Cookie{
		Name:     CookieCSRF,
		Value:    token,
		Path:     cookiePath,
		MaxAge:   cfg.maxAgeSeconds(),
		HttpOnly: false,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearSessionCookie construye la cookie ss_session de borrado (Max-Age
// negativo): la usan logout y toda revocación (FR-004).
func ClearSessionCookie(cfg CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     CookieSession,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearCSRFCookie construye la cookie csrf_token de borrado.
func ClearCSRFCookie(cfg CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     CookieCSRF,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
		HttpOnly: false,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// SessionTokenFromRequest devuelve el token de la cookie ss_session, o false si
// no está presente o viene vacía.
func SessionTokenFromRequest(r *http.Request) (string, bool) {
	return cookieValue(r, CookieSession)
}

// CSRFFromRequest devuelve el token de la cookie csrf_token, o false si no está
// presente o viene vacía.
func CSRFFromRequest(r *http.Request) (string, bool) {
	return cookieValue(r, CookieCSRF)
}

func cookieValue(r *http.Request, name string) (string, bool) {
	if r == nil {
		return "", false
	}
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

// NewCSRFToken genera el token del double-submit firmado (P10): un nonce
// aleatorio de 32 bytes en base64url y su firma HMAC-SHA256(SESSION_SECRET,
// nonce), unidos por ".". El mismo valor viaja en la cookie csrf_token y en la
// cabecera HeaderCSRF.
func NewCSRFToken(secret string) (string, error) {
	nonce := make([]byte, csrfNonceBytes)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generar nonce CSRF: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(nonce)
	return encoded + csrfSeparator + signCSRF(secret, encoded), nil
}

// signCSRF calcula HMAC-SHA256(secret, nonce) en base64url.
func signCSRF(secret, nonce string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(nonce))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// VerifyCSRFToken comprueba que token tiene la forma nonce.firma y que la firma
// es HMAC-SHA256(secret, nonce). La comparación es en tiempo constante.
func VerifyCSRFToken(secret, token string) bool {
	nonce, signature, ok := strings.Cut(token, csrfSeparator)
	if !ok || nonce == "" || signature == "" {
		return false
	}
	return hmac.Equal([]byte(signature), []byte(signCSRF(secret, nonce)))
}

// ValidateCSRF comprueba el double-submit completo: que la cookie y la cabecera
// coinciden y que la cookie lleva una firma válida. Es el punto único donde
// middleware.CSRF (T227) decide si un método no seguro pasa.
func ValidateCSRF(secret, cookieToken, headerToken string) bool {
	if cookieToken == "" || headerToken == "" {
		return false
	}
	if !hmac.Equal([]byte(cookieToken), []byte(headerToken)) {
		return false
	}
	return VerifyCSRFToken(secret, cookieToken)
}
