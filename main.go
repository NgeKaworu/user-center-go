package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/NgeKaworu/user-center/src/auth"
	"github.com/NgeKaworu/user-center/src/cors"
	"github.com/NgeKaworu/user-center/src/engine"
	"github.com/julienschmidt/httprouter"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		addr   = flag.String("l", ":8011", "绑定Host地址")
		dbInit = flag.Bool("i", false, "init database flag")
		mongo  = flag.String("m", "mongodb://localhost:27017", "mongod addr flag")
		db     = flag.String("db", "uc", "database name")
		k      = flag.String("k", "f3fa39nui89Wi707", "iv key")
	)
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	a := auth.NewAuth(*k)
	eng := engine.NewDbEngine()
	err := eng.Open(*mongo, *db, *dbInit, a)

	if err != nil {
		log.Println(err.Error())
	}

	router := httprouter.New()
	// user ctrl
	router.POST("/login", eng.Login)
	router.POST("/register", eng.Regsiter)
	router.GET("/profile", a.JWT(eng.Profile))
	// user mgt
	router.POST("/user/create", a.JWT(eng.Permission(eng.CreateUser)))
	router.DELETE("/user/remove/:uid", a.JWT(eng.Permission(eng.RemoveUser)))
	router.PUT("/user/update", a.JWT(eng.Permission(eng.UpdateUser)))
	router.GET("/user/list", a.JWT(eng.Permission(eng.UserList)))
	// jwt check rpc
	router.GET("/isLogin", a.IsLogin)

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
			eng.Close()
			fmt.Println("safe exit")
			cleanupDone <- true

		}
	}()
	<-cleanupDone

}
