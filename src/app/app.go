package app

import (
	mongoClient "github.com/NgeKaworu/user-center/src/db/mongo"
	"github.com/NgeKaworu/user-center/src/service/auth"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"gopkg.in/gomail.v2"
)

type App struct {
	mongoClient *mongoClient.MongoClient
	rdb         *redis.Client
	validate    *validator.Validate
	trans       *ut.Translator
	auth        *auth.Auth
	d           *gomail.Dialer
}

func New(
	mongoClient *mongoClient.MongoClient,
	rdb *redis.Client,
	validate *validator.Validate,
	trans *ut.Translator,
	auth *auth.Auth,
	d *gomail.Dialer,
) *App {
	return &App{
		mongoClient,
		rdb,
		validate,
		trans,
		auth,
		d,
	}
}
