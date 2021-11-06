package engine

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NgeKaworu/user-center/src/models"
	"github.com/NgeKaworu/user-center/src/parsup"
	"github.com/NgeKaworu/user-center/src/returnee"
	"github.com/NgeKaworu/user-center/src/utils"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Login 登录
func (d *DbEngine) Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		returnee.RetFail(w, err)
		return
	}
	if len(body) == 0 {
		returnee.RetFail(w, errors.New("not has body"))
		return
	}

	p, err := parsup.ParSup().ConvJSON(body)
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	err = utils.Required(p, map[string]string{
		"pwd":   "密码不能为空",
		"email": "邮箱不能为空",
	})

	t := d.GetColl(models.TUser)

	email := strings.ToLower(strings.Replace(p["email"].(string), " ", "", -1))
	res := t.FindOne(context.Background(), bson.M{
		"email": email,
	})

	if res.Err() != nil {
		returnee.RetFail(w, errors.New("没有此用户"))
		return
	}

	var u models.User

	err = res.Decode(&u)
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	dec, err := d.Auth.CFBDecrypter(*u.Pwd)
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	if string(dec) != p["pwd"] {
		returnee.RetFail(w, errors.New("用户名密码不匹配，请注意大小写。"))
		return
	}

	tk, err := d.Auth.GenJWT(u.ID.Hex())

	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	returnee.RetOk(w, tk)
	return
}

// Regsiter 注册
func (d *DbEngine) Regsiter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	if len(body) == 0 {
		returnee.RetFail(w, errors.New("not has body"))
		return
	}

	p, err := parsup.ParSup().ConvJSON(body)
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	// 注册用户默认无权限
	p["isAdmin"] = false

	res, err := d.insertOneUser(p)
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	tk, err := d.Auth.GenJWT(res.InsertedID.(primitive.ObjectID).Hex())
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	returnee.RetOk(w, tk)

}

// Profile 获取用户档案
func (d *DbEngine) Profile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid, err := primitive.ObjectIDFromHex(r.Header.Get("uid"))
	if err != nil {
		returnee.RetFail(w, err)
		return
	}
	t := d.GetColl(models.TUser)

	res := t.FindOne(context.Background(), bson.M{"_id": uid}, options.FindOne().SetProjection(bson.M{
		"pwd": 0,
	}))

	if res.Err() != nil {
		w.WriteHeader(http.StatusUnauthorized)
		returnee.RetFail(w, res.Err())
		return
	}

	var u models.User

	res.Decode(&u)

	returnee.RetOk(w, u)
}

// CreateUser 新增用户
func (d *DbEngine) CreateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	if len(body) == 0 {
		returnee.RetFail(w, errors.New("not has body"))
		return
	}

	p, err := parsup.ParSup().ConvJSON(body)
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	res, err := d.insertOneUser(p)
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	returnee.RetOk(w, res.InsertedID.(primitive.ObjectID).Hex())
}

// RemoveUser 删除用户
func (d *DbEngine) RemoveUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid, err := primitive.ObjectIDFromHex(ps.ByName("uid"))

	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	res := d.GetColl(models.TUser).FindOneAndDelete(context.Background(), bson.M{
		"_id": uid,
	})

	if res.Err() != nil {
		returnee.RetFail(w, res.Err())
		return
	}

	returnee.RetOk(w, "删除成功")
}

// UpdateUser 修改用户
func (d *DbEngine) UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	if len(body) == 0 {
		returnee.RetFail(w, errors.New("not has body"))
	}

	p, err := parsup.ParSup().ConvJSON(body)
	if err != nil {
		returnee.RetFail(w, err)
	}

	err = utils.Required(p, map[string]string{
		"uid": "用户id不能为空",
	})

	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	if pwd, ok := p["pwd"]; ok {
		enc, err := d.Auth.CFBEncrypter(pwd.(string))

		if err != nil {
			returnee.RetFail(w, err)
		}
		p["pwd"] = string(enc)
	}

	if _, ok := p["email"]; ok {
		p["email"] = strings.ToLower(strings.Replace(p["email"].(string), " ", "", -1))
	}

	p["updateAt"] = time.Now().Local()

	uid := p["uid"]

	delete(p, "uid")

	t := d.GetColl(models.TUser)

	res := t.FindOneAndUpdate(context.Background(), bson.M{"_id": uid}, bson.M{"$set": p})

	if res.Err() != nil {
		errMsg := res.Err().Error()
		if strings.Contains(errMsg, "dup key") {
			errMsg = "该邮箱已经被注册"
		}

		returnee.RetFail(w, errors.New(errMsg))
		return
	}

	returnee.RetOk(w, "操作成功")
}

// UserList 查找用户
func (d *DbEngine) UserList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	q := r.URL.Query()

	var skip, limit int64 = 0, 10

	if value := q.Get("skip"); value != "" {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			returnee.RetFail(w, errors.New("skip not number"))
			return
		}
		skip = i
	}

	if value := q.Get("limit"); value != "" {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			returnee.RetFail(w, errors.New("limit not number"))
			return
		}
		limit = i
	}

	t := d.GetColl(models.TUser)

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
		returnee.RetFail(w, err)
		return
	}

	var users []models.User
	err = cur.All(context.Background(), &users)

	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	total, err := t.CountDocuments(context.Background(), params)

	if err != nil {
		returnee.RetFail(w, err)
		return
	}

	returnee.RetOkWithTotal(w, users, total)
}

func (d *DbEngine) insertOneUser(user map[string]interface{}) (*mongo.InsertOneResult, error) {
	err := utils.Required(user, map[string]string{
		"pwd":   "密码不能为空",
		"email": "邮箱不能为空",
		"name":  "昵称不能为空",
	})

	if err != nil {
		return nil, err
	}

	enc, err := d.Auth.CFBEncrypter(user["pwd"].(string))

	if err != nil {
		return nil, err
	}

	user["email"] = strings.ToLower(strings.Replace(user["email"].(string), " ", "", -1))
	user["pwd"] = string(enc)
	user["createAt"] = time.Now().Local()

	t := d.GetColl(models.TUser)

	res, err := t.InsertOne(context.Background(), user)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "dup key") {
			errMsg = "该邮箱已经被注册"
		}

		return nil, errors.New(errMsg)

	}
	return res, nil
}
