package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TestAuthenticator struct{}

var secret = "test"

var testClaims = jwt.MapClaims{
	"aud":    "test-aud",
	"issuer": "test-aud",
	"sub":    int64(1),
	"exp":    time.Now().Add(time.Hour).Unix(),
}

func (auth *TestAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)
	tokenString, _ := token.SignedString([]byte(secret))

	return tokenString, nil
}

func (auth *TestAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
