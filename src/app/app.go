package app

import (
	mongoClient "github.com/NgeKaworu/user-center/src/db/mongo"
	"github.com/NgeKaworu/user-center/src/service/auth"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type App struct {
	mongoClient *mongoClient.MongoClient
	validate    *validator.Validate
	trans       *ut.Translator
	auth        *auth.Auth
}

func New(
	mongoClient *mongoClient.MongoClient,
	validate *validator.Validate,
	trans *ut.Translator,
	auth *auth.Auth,
) *App {
	return &App{
		mongoClient,
		validate,
		trans,
		auth,
	}
}
