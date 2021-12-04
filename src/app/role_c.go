package app

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NgeKaworu/user-center/src/model"
	"github.com/NgeKaworu/user-center/src/util/responser"
	"github.com/hetiansu5/urlquery"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RoleCreate 新增角色
func (app *App) RoleCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if len(body) == 0 {
		responser.RetFail(w, errors.New("not has body"))
		return
	}

	var u model.Role
	err = json.Unmarshal(body, &u)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if err := app.validate.Struct(u); err != nil {
		responser.RetFail(w, err)
		return
	}

	t := app.mongoClient.GetColl(model.TRole)

	res, err := t.InsertOne(context.Background(), u)

	if err != nil {

		responser.RetFail(w, err)
		return

	}

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOk(w, res.InsertedID.(primitive.ObjectID).Hex())
}

// RoleRemove 删除角色
func (app *App) RoleRemove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := primitive.ObjectIDFromHex(ps.ByName("id"))

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	res := app.mongoClient.GetColl(model.TRole).FindOneAndDelete(context.Background(), bson.M{
		"_id": id,
	})

	if res.Err() != nil {
		responser.RetFail(w, res.Err())
		return
	}

	responser.RetOk(w, "删除成功")
}

// RoleUpdate 修改角色
func (app *App) RoleUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if len(body) == 0 {
		responser.RetFail(w, errors.New("not has body"))
	}

	var u model.Role

	err = json.Unmarshal(body, &u)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if u.ID == nil {
		responser.RetFail(w, errors.New("id不能为空"))
		return
	}

	if err := app.validate.Struct(u); err != nil {
		responser.RetFail(w, err)
		return
	}

	updateAt := time.Now().Local()
	u.UpdateAt = &updateAt

	res := app.mongoClient.GetColl(model.TRole).FindOneAndUpdate(context.Background(), bson.M{"_id": *u.ID}, bson.M{"$set": &u})

	if res.Err() != nil {

		responser.RetFail(w, res.Err())
		return
	}

	responser.RetOk(w, "操作成功")
}

// RoleList 查找角色
func (app *App) RoleList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p := struct {
		Keyword *string `query:"keyword,omitempty" validate:"omitempty"`
		Skip    *int64  `query:"skip,omitempty" validate:"omitempty,min=0"`
		Limit   *int64  `query:"limit,omitempty" validate:"omitempty,min=0"`
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

	params := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": p.Keyword}},
		},
	}

	opt := options.Find()

	if p.Limit != nil {
		opt.SetLimit(*p.Limit)
	} else {
		opt.SetLimit(10)
	}

	if p.Skip != nil {
		opt.SetSkip(*p.Skip)
	}
	t := app.mongoClient.GetColl(model.TRole)

	cur, err := t.Find(context.Background(), params, opt)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	var perms []model.Role
	err = cur.All(context.Background(), &perms)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	total, err := t.CountDocuments(context.Background(), params)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOkWithTotal(w, perms, total)
}
