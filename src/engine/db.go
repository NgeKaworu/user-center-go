package engine

import (
	"context"
	"log"
	"time"

	"github.com/NgeKaworu/user-center/src/auth"
	"github.com/NgeKaworu/user-center/src/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// DbEngine 关系型数据库引擎
type DbEngine struct {
	MgEngine *mongo.Client //关系型数据库引擎
	Mdb      string
	Auth     *auth.Auth // 加解密客户端
}

// NewDbEngine 实例工厂
func NewDbEngine() *DbEngine {
	return &DbEngine{}
}

// Open 开启连接池
func (d *DbEngine) Open(mg, mdb string, initdb bool, a *auth.Auth) error {
	d.Mdb = mdb
	d.Auth = a
	ops := options.Client().ApplyURI(mg)
	p := uint64(39000)
	ops.MaxPoolSize = &p
	ops.WriteConcern = writeconcern.New(writeconcern.J(true), writeconcern.W(1))
	ops.ReadPreference = readpref.PrimaryPreferred()
	db, err := mongo.NewClient(ops)

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Connect(ctx)
	if err != nil {
		panic(err)
	}

	//err = db.Ping(ctx, readpref.PrimaryPreferred())
	//if err != nil {
	//	log.Println("ping err", err)
	//}

	d.MgEngine = db

	//初始化数据库
	if initdb {
		var session *mongo.Client
		session, err = mongo.NewClient(ops)
		if err != nil {
			panic(err)
		}
		err = session.Connect(context.Background())
		if err != nil {
			panic(err)
		}
		defer session.Disconnect(context.Background())
		// 用户表
		t := session.Database(mdb).Collection(models.TUser)

		// 初始化超管
		res, err := d.insertOneUser(map[string]interface{}{
			"email":   "NgeKaworu@163.com",
			"name":    "furan",
			"pwd":     "12345678",
			"isAdmin": true,
		})

		if err != nil {
			log.Println(err)
		} else {
			log.Println(res.InsertedID)
		}

		indexView := t.Indexes()
		_, err = indexView.CreateMany(context.Background(), []mongo.IndexModel{
			{Keys: bsonx.Doc{bsonx.Elem{Key: "email", Value: bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)},
			{Keys: bsonx.Doc{bsonx.Elem{Key: "name", Value: bsonx.Int32(1)}}},
		})
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

// GetColl 获取表名
func (d *DbEngine) GetColl(coll string) *mongo.Collection {
	col, _ := d.MgEngine.Database(d.Mdb).Collection(coll).Clone()
	return col
}

// Close 关闭连接池
func (d *DbEngine) Close() {
	d.MgEngine.Disconnect(context.Background())
}
