package app

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

const (
	CAPTCHA_MAX_AGE = 60 * 10 // seconds
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

func (app *App) getRedisLocked(email *string) (bool, error) {
	exists := app.rdb.Exists(context.Background(), "test")
	if exists.Err() != nil {
		return true, exists.Err()
	}

	if exists.Val() == 1 {
		return true, nil
	}

	return false, nil
}

func (app *App) setRedisCaptcha(email, captcha *string) error {

	cmd := app.rdb.Set(context.Background(), *email, *captcha, time.Duration(CAPTCHA_MAX_AGE)*time.Second)
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
			MaxAge:   600,
		}
		w.Header().Set("Set-Cookie", c.String())

		return false
	}

	return true

}
