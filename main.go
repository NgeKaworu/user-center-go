package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/NgeKaworu/user-center/src/app"
	mongoClient "github.com/NgeKaworu/user-center/src/db/mongo"
	"github.com/NgeKaworu/user-center/src/middleware/cors"
	"github.com/NgeKaworu/user-center/src/service/auth"
	"github.com/NgeKaworu/user-center/src/util/validator"
	"github.com/go-redis/redis/v8"
	"gopkg.in/gomail.v2"

	"github.com/julienschmidt/httprouter"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		addr   = flag.String("l", ":8020", "绑定Host地址")
		dbInit = flag.Bool("i", true, "init database flag")
		m      = flag.String("m", "mongodb://localhost:27017", "mongod addr flag")
		db     = flag.String("db", "uc", "database name")
		k      = flag.String("k", "f3fa39nui89Wi707", "iv key")
		r      = flag.String("r", "localhost:6379", "rdb addr")
		ePwd   = flag.String("d", "", "email pwd")
	)
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	mongoClient := mongoClient.New()
	err := mongoClient.Open(*m, *db, *dbInit)
	if err != nil {
		log.Println(err.Error())
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     *r,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	validate := validator.NewValidator()
	trans := validator.NewValidatorTranslator(validate)

	rand.Seed(time.Now().Unix())
	d := gomail.NewDialer("smtp.gmail.com", 587, "ngekaworu@gmail.com", *ePwd)

	auth := auth.New(*k)
	app := app.New(mongoClient, rdb, validate, trans, auth, d)

	router := httprouter.New()
	// user ctrl
	router.POST("/login", app.Login)
	router.POST("/register", app.Regsiter)
	router.GET("/profile", app.JWT(app.Profile))
	// user mgt
	router.POST("/user/create", app.JWT(app.CheckPerm("admin")(app.CreateUser)))
	router.DELETE("/user/remove/:uid", app.JWT(app.CheckPerm("admin")(app.RemoveUser)))
	router.PUT("/user/update", app.JWT(app.CheckPerm("admin")(app.UpdateUser)))
	router.GET("/user/list", app.JWT(app.CheckPerm("admin")(app.UserList)))
	router.GET("/user/validate", app.UserValidateEmail)

	// role mgt
	router.POST("/role/create", app.JWT(app.CheckPerm("admin")(app.RoleCreate)))
	router.DELETE("/role/remove/:id", app.JWT(app.CheckPerm("admin")(app.RoleRemove)))
	router.PUT("/role/update", app.JWT(app.CheckPerm("admin")(app.RoleUpdate)))
	router.GET("/role/list", app.JWT(app.RoleList))
	router.GET("/role/validate", app.JWT(app.RoleValidateKey))

	// perm mgt
	router.POST("/perm/create", app.JWT(app.CheckPerm("admin")(app.PermCreate)))
	router.DELETE("/perm/remove/:id", app.JWT(app.CheckPerm("admin")(app.PermRemove)))
	router.PUT("/perm/update", app.JWT(app.CheckPerm("admin")(app.PermUpdate)))
	router.GET("/perm/list", app.JWT(app.PermList))
	router.GET("/perm/validate", app.JWT(app.PermValidateKey))
	router.GET("/menu", app.JWT(app.Menu))
	router.GET("/micro-app", app.MicroApp)

	// rpc
	router.HEAD("/check-perm-rpc/:perm", app.JWT(app.CheckPermRPC))

	// captcha
	router.GET("/captcha/fetch", app.FetchCaptcha)
	router.GET("/captcha/check", app.CheckCaptcha)

	srv := &http.Server{Handler: cors.CORS(router), ErrorLog: nil}
	srv.Addr = *addr

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("server on http port", srv.Addr)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	cleanup := make(chan bool)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range signalChan {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			go func() {
				_ = srv.Shutdown(ctx)
				cleanup <- true
			}()
			<-cleanup
			mongoClient.Close()
			rdb.Close()
			fmt.Println("safe exit")
			cleanupDone <- true

		}
	}()
	<-cleanupDone

}
