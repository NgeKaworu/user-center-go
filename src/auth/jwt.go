package auth

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/NgeKaworu/user-center/src/returnee"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

// JWT json web token
func (a *Auth) JWT(next httprouter.Handle) httprouter.Handle {
	//权限验证
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		audience, err := a.checkTokenAudience(r.Header.Get("Authorization"))
		if err != nil {
			// Request Basic Authentication otherwise
			w.Header().Set("WWW-Authenticate", "Bearer realm=Restricted")
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(err)
			returnee.RetFail(w, errors.New("身份认证失败，请重新登录"))
			return
		}

		r.Header.Set("uid", *audience)
		next(w, r, ps)
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

// IsLogin 校验用户己登录
func (a *Auth) IsLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	audience, err := a.checkTokenAudience(r.Header.Get("Authorization"))
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	returnee.RetOk(w, audience)
}

func (a *Auth) checkTokenAudience(auth string) (audience *string, err error) {
	if auth == "" {
		err = errors.New("auth is empty")
		return
	}

	token, err := jwt.ParseWithClaims(auth, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.Key, nil
	})

	if err != nil {
		err = errors.New("token is invalid")
		return
	}

	if tk, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		audience = &tk.Audience
	}

	return
}
