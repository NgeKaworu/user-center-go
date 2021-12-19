package app

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/NgeKaworu/user-center/src/model"
	"github.com/NgeKaworu/user-center/src/util/responser"
	"github.com/hetiansu5/urlquery"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PermCreate 新增权限
func (app *App) PermCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	var p model.Perm
	err = json.Unmarshal(body, &p)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if err := app.validate.Struct(&p); err != nil {
		responser.RetFail(w, err)
		return
	}

	tt := time.Now().Local()
	p.CreateAt = &tt

	t := app.mongoClient.GetColl(model.TPerm)

	res, err := t.InsertOne(context.Background(), p)

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "dup key") {
			errMsg = "该key已经使用"
		}

		responser.RetFail(w, errors.New(errMsg))
		return

	}

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOk(w, res.InsertedID)
}

// PermRemove 删除权限
func (app *App) PermRemove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	if id == "" {
		responser.RetFail(w, errors.New("ID不能为空"))
		return
	}

	t := app.mongoClient.GetColl(model.TPerm)

	c, err := t.CountDocuments(context.Background(), bson.M{
		"pID": id,
	})

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if c != 0 {
		responser.RetFail(w, errors.New("删除失败：该菜单下存在子菜单。"))
		return
	}

	res := t.FindOneAndDelete(context.Background(), bson.M{
		"_id": id,
	})

	if res.Err() != nil {
		responser.RetFail(w, res.Err())
		return
	}

	responser.RetOk(w, "删除成功")
}

// PermUpdate 修改权限
func (app *App) PermUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if len(body) == 0 {
		responser.RetFail(w, errors.New("not has body"))
	}

	var u model.Perm

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

	u.IsMenu = nil
	tt := time.Now().Local()
	u.UpdateAt = &tt
	updater := bson.M{"$set": &u}
	if u.PID == nil {
		updater["$unset"] = bson.M{"pID": ""}
	}
	res := app.mongoClient.GetColl(model.TPerm).FindOneAndUpdate(context.Background(), bson.M{"_id": *u.ID}, updater)

	if res.Err() != nil {
		errMsg := res.Err().Error()
		if strings.Contains(errMsg, "dup key") {
			errMsg = "该key已经使用"
		}

		responser.RetFail(w, errors.New(errMsg))
		return
	}

	responser.RetOk(w, u.ID)
}

// PermList 查找权限
func (app *App) PermList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p := struct {
		IsMenu  *bool   `query:"isMenu,omitempty" validate:"omitempty"`
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

	params := bson.M{}

	if p.Keyword != nil {
		params["$or"] = []bson.M{
			{"name": bson.M{"$regex": p.Keyword}},
			{"_id": bson.M{"$regex": p.Keyword}},
		}
	}

	if p.IsMenu != nil {
		params["isMenu"] = p.IsMenu
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
	t := app.mongoClient.GetColl(model.TPerm)

	cur, err := t.Find(context.Background(), params, opt)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	var perms []model.Perm
	err = cur.All(context.Background(), &perms)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	total, err := t.CountDocuments(context.Background(), &params)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOkWithTotal(w, perms, total)
}

// PermValidateKey key 校验
func (app *App) PermValidateKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p := struct {
		ID *string `query:"id,omitempty" validate:"omitempty,required"`
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

	total, err := app.mongoClient.GetColl(model.TPerm).CountDocuments(context.Background(), bson.M{"_id": *p.ID})

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if total != 0 {
		responser.RetFail(w, errors.New("key 重复"))
		return
	}

	responser.RetOk(w, "validate key")
}
