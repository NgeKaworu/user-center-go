package auth

import (
	"net/http"
	"time"

	"github.com/NgeKaworu/time-mgt-go/src/resultor"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

// JWT json web token
func (a *Auth) JWT(next httprouter.Handle) httprouter.Handle {
	//权限验证
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			token, err := jwt.ParseWithClaims(auth, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				return a.Key, nil
			})
			if err == nil {
				if tk, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
					r.Header.Set("uid", tk.Audience)
					next(w, r, ps)
					return
				}
			}
		}

		// Request Basic Authentication otherwise
		w.Header().Set("WWW-Authenticate", "Bearer realm=Restricted")
		w.WriteHeader(http.StatusUnauthorized)
		resultor.RetFail(w, "身份认证失败，请重新登录")
	}
}

// GenJWT generate jwt
func (a *Auth) GenJWT(aud string) (string, error) {
	time.Now()
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 24 * 15).Unix(),
		Issuer:    "fuRan",
		Audience:  aud,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.Key)
}
