package app

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NgeKaworu/user-center/src/model"
	"github.com/NgeKaworu/user-center/src/parsup"
	"github.com/NgeKaworu/user-center/src/util/responser"
	"github.com/NgeKaworu/user-center/src/utils"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Login 登录
func (app *App) Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	type user struct {
		ID    *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" `                         // id
		Pwd   *string             `json:"pwd,omitempty" bson:"pwd,omitempty" validate:"required"`     // 账号
		Email *string             `json:"email,omitempty" bson:"email,omitempty" validate:"required"` // 密码
	}

	inputUser := new(user)

	if err := json.Unmarshal(body, &inputUser); err != nil {
		responser.RetFail(w, err)
		return
	}

	if err := app.validate.Struct(inputUser); err != nil {
		responser.RetFailWithTrans(w, err, app.trans)
		return
	}

	t := app.mongoClient.GetColl(model.TUser)

	email := strings.ToLower(strings.Replace(*inputUser.Email, " ", "", -1))
	res := t.FindOne(context.Background(), bson.M{
		"email": email,
	})

	if res.Err() != nil {
		responser.RetFail(w, errors.New("用户名或密码不正确"))
		return
	}

	outputUser := new(user)

	err = res.Decode(&outputUser)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	dec, err := app.auth.CFBDecrypter(*outputUser.Pwd)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if string(dec) != *inputUser.Pwd {
		responser.RetFail(w, errors.New("用户名或密码不正确"))
		return
	}

	tk, err := app.auth.GenJWT(outputUser.ID.Hex())

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOk(w, tk)
	return
}

// Regsiter 注册
func (app *App) Regsiter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	var u model.User

	if err := json.Unmarshal(body, u); err != nil {
		responser.RetFail(w, err)
		return
	}

	res, err := app.insertOneUser(&u)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	tk, err := app.auth.GenJWT(res.InsertedID.(primitive.ObjectID).Hex())
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOk(w, tk)

}

// Profile 获取用户档案
func (app *App) Profile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid, err := primitive.ObjectIDFromHex(r.Header.Get("uid"))
	if err != nil {
		responser.RetFail(w, err)
		return
	}
	t := d.GetColl(model.TUser)

	res := t.FindOne(context.Background(), bson.M{"_id": uid}, options.FindOne().SetProjection(bson.M{
		"pwd": 0,
	}))

	if res.Err() != nil {
		w.WriteHeader(http.StatusUnauthorized)
		responser.RetFail(w, res.Err())
		return
	}

	var u model.User

	res.Decode(&u)

	responser.RetOk(w, u)
}

// CreateUser 新增用户
func (app *App) CreateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	p, err := parsup.ParSup().ConvJSON(body)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	res, err := d.insertOneUser(p)
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOk(w, res.InsertedID.(primitive.ObjectID).Hex())
}

// RemoveUser 删除用户
func (app *App) RemoveUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid, err := primitive.ObjectIDFromHex(ps.ByName("uid"))

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	res := d.GetColl(model.TUser).FindOneAndDelete(context.Background(), bson.M{
		"_id": uid,
	})

	if res.Err() != nil {
		responser.RetFail(w, res.Err())
		return
	}

	responser.RetOk(w, "删除成功")
}

// UpdateUser 修改用户
func (app *App) UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if len(body) == 0 {
		responser.RetFail(w, errors.New("not has body"))
	}

	p, err := parsup.ParSup().ConvJSON(body)
	if err != nil {
		responser.RetFail(w, err)
	}

	err = utils.Required(p, map[string]string{
		"uid": "用户id不能为空",
	})

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	if pwd, ok := p["pwd"]; ok {
		enc, err := d.Auth.CFBEncrypter(pwd.(string))

		if err != nil {
			responser.RetFail(w, err)
		}
		p["pwd"] = string(enc)
	}

	if _, ok := p["email"]; ok {
		p["email"] = strings.ToLower(strings.Replace(p["email"].(string), " ", "", -1))
	}

	p["updateAt"] = time.Now().Local()

	uid := p["uid"]

	delete(p, "uid")

	t := d.GetColl(model.TUser)

	res := t.FindOneAndUpdate(context.Background(), bson.M{"_id": uid}, bson.M{"$set": p})

	if res.Err() != nil {
		errMsg := res.Err().Error()
		if strings.Contains(errMsg, "dup key") {
			errMsg = "该邮箱已经被注册"
		}

		responser.RetFail(w, errors.New(errMsg))
		return
	}

	responser.RetOk(w, "操作成功")
}

// UserList 查找用户
func (app *App) UserList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	q := r.URL.Query()

	var skip, limit int64 = 0, 10

	if value := q.Get("skip"); value != "" {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			responser.RetFail(w, errors.New("skip not number"))
			return
		}
		skip = i
	}

	if value := q.Get("limit"); value != "" {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			responser.RetFail(w, errors.New("limit not number"))
			return
		}
		limit = i
	}

	t := d.GetColl(model.TUser)

	params := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": q.Get("keyword")}},
			{"email": bson.M{"$regex": q.Get("keyword")}},
		},
	}

	cur, err := t.Find(context.Background(), params,
		options.Find().
			SetSkip(skip).
			SetLimit(limit).
			SetProjection(bson.M{
				"pwd": 0,
			}),
	)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	var users []model.User
	err = cur.All(context.Background(), &users)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	total, err := t.CountDocuments(context.Background(), params)

	if err != nil {
		responser.RetFail(w, err)
		return
	}

	responser.RetOkWithTotal(w, users, total)
}

func (app *App) insertOneUser(u *model.User) (*mongo.InsertOneResult, error) {
	err := utils.Required(u, map[string]string{
		"pwd":   "密码不能为空",
		"email": "邮箱不能为空",
		"name":  "昵称不能为空",
	})

	if err := app.validate.Struct(u); err != nil {
		return nil, err
	}

	enc, err := app.auth.CFBEncrypter(*u.Pwd)

	email := strings.ToLower(strings.Replace(*u.Email, " ", "", -1))
	pwd := string(enc)
	now := time.Now().Local()
	if err != nil {
		return nil, err
	}

	u.Email = &email
	u.Pwd = &pwd
	u.CreateAt = &now

	t := app.mongoClient.GetColl(model.TUser)

	res, err := t.InsertOne(context.Background(), u)

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "dup key") {
			errMsg = "该邮箱已经被注册"
		}

		return nil, errors.New(errMsg)

	}
	return res, nil
}
