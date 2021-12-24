package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/NgeKaworu/user-center/src/model"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JWT json web token
func (app *App) JWT(next httprouter.Handle) httprouter.Handle {
	//权限验证
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		audience, err := app.checkUser(r)
		if err != nil {
			log.Println(err)
			// Request Basic Authentication otherwise
			w.Header().Set("WWW-Authenticate", "Bearer realm=Restricted")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Set("uid", *audience)
		next(w, r, ps)
	}
}

// perm mid
func (app *App) CheckPerm(perm string) func(httprouter.Handle) httprouter.Handle {
	return func(next httprouter.Handle) httprouter.Handle {
		//权限验证
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			u := r.Header.Get("uid")
			k := u + ":perm"

			const (
				NONEXIST = 0
				EXIST    = 1
			)

			e, _ := app.rdb.Exists(context.Background(), k).Result()

			if e == EXIST {
				if b, _ := app.rdb.SIsMember(context.Background(), k, perm).Result(); b {
					next(w, r, ps)
					return
				}
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if e == NONEXIST {
				p, err := app.getSetPerm(u)

				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				for _, role := range p {
					if role == perm {
						next(w, r, ps)
						return
					}
				}

				w.WriteHeader(http.StatusForbidden)
			}
		}
	}
}

// perm rpc
func (app *App) CheckPermRPC(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p, err := app.getSetPerm(r.Header.Get("uid"))

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	for _, role := range p {
		if role == ps.ByName("perm") {
			return
		}
	}

	w.WriteHeader(http.StatusForbidden)
}

// service

// checkUser
func (app *App) checkUser(r *http.Request) (*string, error) {
	bear, err := app.getBearer(r)
	if err != nil {
		return nil, err
	}

	s, err := app.rdb.Get(context.Background(), *bear).Result()

	if err != nil {
		return nil, err
	}

	return &s, nil
}

// getBearer
func (app *App) getBearer(r *http.Request) (*string, error) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, errors.New("unknown authorization type")
	}
	auth = auth[7:]
	return &auth, nil
}

// getSetPerm
func (app *App) getSetPerm(u string) ([]string, error) {

	uid, err := primitive.ObjectIDFromHex(u)
	if err != nil {
		return nil, err
	}

	cur, err := app.mongoClient.GetColl(model.TUser).Aggregate(context.Background(), []bson.M{
		{
			"$match": bson.M{
				"_id": uid,
			},
		},
		{"$lookup": bson.M{
			"from":         "t_role",
			"localField":   "roles",
			"foreignField": "_id",
			"as":           "roles",
		},
		},
		{"$group": bson.M{
			"_id":   "$_id",
			"roles": bson.M{"$first": "$roles.perms"},
		},
		},
		{"$unwind": "$roles"},
		{"$unwind": "$roles"},
		{"$group": bson.M{"_id": "$_id",
			"roles": bson.M{"$addToSet": "$roles"},
		}},
	})

	if err != nil {
		return nil, err
	}

	res := make([]struct {
		Roles []string `bson:"roles"`
	}, 0)

	if err = cur.All(context.Background(), &res); err != nil {
		return nil, err
	}
	k := u + ":perm"
	app.rdb.SAdd(context.Background(), k, res[0].Roles)
	app.rdb.Expire(context.Background(), k, time.Hour*12)

	return res[0].Roles, nil

}
