package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
}

type header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

func Generate(userID int64, email string, secret string, expiration time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Exp:    time.Now().Add(expiration).Unix(),
	}

	return sign(claims, secret)
}

func Validate(token string, secret string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	if !validSignature(signingInput, parts[2], secret) {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: decode payload", ErrInvalidToken)
	}

	var claims Claims
	if err = json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("%w: parse claims", ErrInvalidToken)
	}

	if claims.UserID <= 0 || claims.Email == "" || claims.Exp <= 0 {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() >= claims.Exp {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func sign(claims Claims, secret string) (string, error) {
	encodedHeader, err := encodeJSON(header{
		Algorithm: "HS256",
		Type:      "JWT",
	})
	if err != nil {
		return "", err
	}

	encodedClaims, err := encodeJSON(claims)
	if err != nil {
		return "", err
	}

	signingInput := encodedHeader + "." + encodedClaims
	signature := createSignature(signingInput, secret)

	return signingInput + "." + signature, nil
}

func encodeJSON(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func createSignature(signingInput string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validSignature(signingInput string, signature string, secret string) bool {
	expectedSignature := createSignature(signingInput, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (claims Claims) Subject() string {
	return strconv.FormatInt(claims.UserID, 10)
}
