package engine

import (
	"context"
	"net/http"

	"github.com/NgeKaworu/user-center/src/models"
	"github.com/NgeKaworu/user-center/src/returnee"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Permission 校验用户是否超管
func (d *DbEngine) Permission(next httprouter.Handle) httprouter.Handle {
	//权限验证
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		uid, err := primitive.ObjectIDFromHex(r.Header.Get("uid"))
		if err == nil {
			if access, err := d.GetColl(models.TUser).CountDocuments(context.Background(), bson.M{
				"_id":     uid,
				"isAdmin": true,
			}); access > 0 && err == nil {
				next(w, r, ps)
				return
			}
		}

		// Request Basic Authentication otherwise
		w.WriteHeader(http.StatusUnauthorized)
		returnee.RetFail(w, errors.New("无权访问"))
	}
}
