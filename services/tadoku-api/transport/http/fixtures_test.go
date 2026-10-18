package http

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
)

func rsaJWK(kid string, publicKey *rsa.PublicKey) string {
	return fmt.Sprintf(`{"kty":"RSA","kid":%q,"alg":"RS256","use":"sig","n":%q,"e":"AQAB"}`,
		kid, base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()))
}
