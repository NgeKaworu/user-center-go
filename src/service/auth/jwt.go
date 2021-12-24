package auth

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

// GenJWT generate jwt
func (a *Auth) GenJWT(claims *jwt.StandardClaims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.key)
}

func (a *Auth) CheckTokenAudience(auth string) (audience *string, err error) {
	if auth == "" {
		err = errors.New("auth is empty")
		return
	}

	token, err := jwt.ParseWithClaims(auth, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.key, nil
	})

	if err != nil {
		return
	}

	if tk, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		audience = &tk.Audience
	}

	return
}
