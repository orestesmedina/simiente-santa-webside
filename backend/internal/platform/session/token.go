// Package session implementa el plumbing de la sesión del panel: el token
// opaco y su hash, las cookies httpOnly de sesión y CSRF, los tipos de
// identidad que implementa el dominio usuarios, el Store de sesiones sobre
// Redis y el mecanismo de los contadores de intentos de FR-006.
//
// platform no conoce dominios (arq. R1): Resolver y Store son puertos que el
// dominio usuarios implementa/consume. Este paquete nunca ve una contraseña ni
// un hash de contraseña.
package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// TokenBytes es la entropía del token de sesión (R1): 32 bytes de crypto/rand.
const TokenBytes = 32

// NewToken genera el token opaco que viaja en la cookie ss_session: 32 bytes
// de crypto/rand en base64url sin relleno. El token en claro NUNCA se guarda
// (P1): Redis almacena solo su SHA-256.
func NewToken() (string, error) {
	buf := make([]byte, TokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generar token de sesión: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashToken devuelve el SHA-256 del token en hexadecimal (64 caracteres): la
// forma que se usa como clave Redis (sess:<hash>). Es estable, de longitud fija
// y no reversible (P1).
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
