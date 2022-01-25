package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/NgeKaworu/user-center/src/model"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

const (
	MAX_AGE         = 60           // seconds
	CAPTCHA_MAX_AGE = MAX_AGE * 10 // seconds
	CAPTCHA_KEY     = "session"
)

func (app *App) sendCaptcha(mail, captcha *string) {

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress("ngekaworu@gmail.com", "盈虚"))
	m.SetHeader("To", *mail)
	m.SetHeader("Subject", "验证码")
	m.SetBody("text/html", "你的验证码是："+*captcha+", 10分钟内有效")

	if err := app.d.DialAndSend(m); err != nil {
		panic(err)
	}

}

func (app *App) getCacheCaptcha(key *string) (string, error) {

	cmd := app.rdb.Get(context.Background(), *key)
	if cmd.Err() != nil {
		return "", cmd.Err()
	}

	return cmd.Val(), nil

}

func (app *App) setRedisCaptcha(email, captcha *string) error {

	cmd := app.rdb.Set(context.Background(), *email, *captcha, time.Duration(CAPTCHA_MAX_AGE)*time.Second)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}

func (app *App) removeRedisCaptcha(email *string) error {
	cmd := app.rdb.Del(context.Background(), *email)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (app *App) getSetSessionLocked(w http.ResponseWriter, r *http.Request) bool {
	_, err := r.Cookie(CAPTCHA_KEY)
	if err != nil {

		c := &http.Cookie{
			Name:     CAPTCHA_KEY,
			Value:    uuid.NewString(),
			HttpOnly: true,
			MaxAge:   MAX_AGE,
			Path:     "/",
		}
		w.Header().Set("Set-Cookie", c.String())

		return false
	}

	return true

}

func (app *App) checkCaptcha(w http.ResponseWriter, r *http.Request, capcha *model.Captcha) error {
	err := app.validate.Struct(capcha)

	if err != nil {
		var errMsg string
		for _, v := range err.(validator.ValidationErrors).Translate(*app.trans) {
			errMsg += v + ","
		}
		return errors.New(errMsg)
	}

	captcha, err := app.getCacheCaptcha(capcha.Email)

	if err != nil {
		return err

	}

	if captcha != *capcha.Captcha {
		return errors.New("验证码错误")
	}

	return nil

}
