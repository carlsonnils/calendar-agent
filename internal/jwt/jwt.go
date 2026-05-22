package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
)

var (
	HmacKey []byte
)

type Header struct {
    Alg string `json:"alg"`
    Typ string `json:"typ"`
}

type Body struct {
    Username string `json:"username"`
}

// create a new jwt from a header and body
func Make(h Header, b Body) (string, error) {
	// encode header
    headerBytes, err := json.Marshal(h)
    if err != nil {
        return "", err
    }
    header := base64.URLEncoding.EncodeToString(headerBytes)

	// encode body
    bodyBytes, err := json.Marshal(b)
    if err != nil {
        return "", err
    }
    body := base64.URLEncoding.EncodeToString(bodyBytes)

	// combine header and body for payload
    payload := header + "." + body

	// create the signature
    mac := hmac.New(sha256.New, HmacKey)
    mac.Write([]byte(payload))

    signatureBytes := mac.Sum(nil)
    signature := base64.URLEncoding.EncodeToString(signatureBytes)
	
	// combine payload with signature for jwt
    return payload + "." + signature, nil
}

// check body is what was originally sent with the jwt
func CheckSignature(j string) bool {
	// break jwt into head, body, signature
    parts := strings.Split(j, ".")
    if len(parts) != 3 {
        return false
    }

	// create new signature from current head and body
    mac := hmac.New(sha256.New, HmacKey)
    mac.Write([]byte(parts[0] + "." + parts[1]))
    exSignature := mac.Sum(nil)

	// extract signature from jwt
    signature, err := base64.URLEncoding.DecodeString(parts[2])
    if err != nil {
        log.Println(err)
        return false
    }

	// check signature is the same as newly created signature
    return hmac.Equal(signature, exSignature)
}

// decode the jwt body information
func DecodeBody(j string) (Body, error) {
	// split jwt into parts
    var b Body
    jwtParts := strings.Split(j, ".")
    if len(jwtParts) != 3 {
        return b, errors.New("jwt token not 3 parts header, body, and signature")
    }

	// decode the base64 url encoding
    bodyBytes, err := base64.URLEncoding.DecodeString(jwtParts[1])
    if err != nil {
        return b, err
    }

	// marshal body to golang type
    err = json.Unmarshal(bodyBytes, &b)
    if err != nil {
        return b, err
    }

    return b, nil
}
