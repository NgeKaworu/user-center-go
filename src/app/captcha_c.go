package app

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/NgeKaworu/user-center/src/util/responser"
	"github.com/hetiansu5/urlquery"
	"github.com/julienschmidt/httprouter"
)

func (app *App) FetchCaptcha(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	if app.getSetSessionLocked(w, r) {
		w.WriteHeader(http.StatusNotModified)
		responser.RetOk(w, "验证码已经发送")
		return
	}

	p := struct {
		Email *string `query:"email,omitempty" validate:"required,email"`
	}{}

	err := urlquery.Unmarshal([]byte(r.URL.RawQuery), &p)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	err = app.validate.Struct(&p)
	if err != nil {
		responser.RetFailWithTrans(w, err, app.trans)
		return
	}

	lock, err := app.getRedisLocked(p.Email)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if lock {
		w.WriteHeader(http.StatusNotModified)
		responser.RetOk(w, "验证码已经发送")
		return
	}

	captcha := padStartZero(rand.Intn(10000))
	app.setRedisCaptcha(p.Email, &captcha)

	w.Header().Set("Cache-Control", "max-age="+strconv.FormatInt(int64(CAPTCHA_MAX_AGE), 10))
	go app.sendCaptcha(p.Email, &captcha)

	responser.RetOk(w, "验证码已经发送")

}

func (app *App) CheckCaptcha(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	if !app.getSetSessionLocked(w, r) {
		responser.RetFail(w, errors.New("验证码已过期, code: 001"))
		return
	}

	p := struct {
		Captcha *string `query:"captcha,omitempty" validate:"required"`
		Email   *string `query:"email,omitempty" validate:"required,email"`
	}{}

	err := urlquery.Unmarshal([]byte(r.URL.RawQuery), &p)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	err = app.validate.Struct(&p)
	if err != nil {
		responser.RetFailWithTrans(w, err, app.trans)
		return
	}

	captcha, err := app.getCacheCaptcha(p.Email)

	if err != nil {
		log.Println(err)
		responser.RetFail(w, errors.New("验证码已过期, code: 002"))
		return
	}

	if captcha != *p.Captcha {
		responser.RetFail(w, errors.New("验证码错误"))
		return
	}

	responser.RetOk(w, "验证通过")

}

func padStartZero(i int) string {
	s := strconv.FormatInt(int64(i), 10)
	l := 4 - len(s)
	for l > 0 {
		l--
		s = "0" + s
	}
	return s
}
