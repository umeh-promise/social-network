package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string
	aud    string
	issuer string
}

func NewJWTAuthenticator(secret, aud, issuer string) *JWTAuthenticator {
	return &JWTAuthenticator{secret: secret, aud: aud, issuer: issuer}
}

func (auth *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(auth.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (auth *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		return []byte(auth.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(auth.aud),
		jwt.WithIssuer(auth.aud),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
