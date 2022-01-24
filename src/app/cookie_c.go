package app

import (
	"log"
	"net/http"

	"github.com/NgeKaworu/user-center/src/util/responser"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func (app *App) Cookie(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	uid, err := r.Cookie("cookie")
	if uid != nil {
		log.Println(uid)
	}

	if err != nil {

		c := &http.Cookie{
			Name:     "cookie",
			Value:    uuid.NewString(),
			HttpOnly: true,
		}

		w.Header().Set("Set-Cookie", c.String())
	}

	responser.RetOk(w, "")
}
