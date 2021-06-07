package main

import (
	"context"
	"crypto/rsa"
	"expvar"
	"fmt"
	"github.com/ardanlabs/conf"
	"github.com/dgrijalva/jwt-go"
	"gitlab.com/altercolt/alproject/business/auth"
	"gitlab.com/altercolt/alproject/foundation/database"
	"io/ioutil"

	"github.com/pkg/errors"
	"gitlab.com/altercolt/alproject/app/sales-api/handlers"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var build = "SNAPSHOT 0.0.1"

func main() {
	log := log.New(os.Stdout, "SALES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if err := run(log); err != nil {
		log.Println("main: error: ", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	//Configuration
	var cfg struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s,noprint"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Auth struct {
			KeyID          string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
			PrivateKeyFile string `conf:"default:./private.pem"`
			Algorithm      string `conf:"default:RS256"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:1952,noprint"`
			Host       string `conf:"default:localhost"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:true"`
		}
	}

	cfg.Version.Desc = "copyright information here"
	cfg.Version.Desc = build

	if err := conf.Parse(os.Args[1:], "SALES", &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage("SALES", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString("SALES", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")

			}
			fmt.Println(version)
			return nil
		}
	}
	// ================================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	log.Printf("main : Started : Application initializing :version %q", build)
	defer log.Println("main: Completed")
	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: Config :\n%v\n", out)


	privatePEM, err := ioutil.ReadFile(cfg.Auth.PrivateKeyFile)
	if err != nil{
		return errors.Wrap(err, "reading auth private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return errors.Wrap(err, "parsing auth private key")
	}
	lookup := func(kid string) (*rsa.PublicKey, error) {
		switch kid {
		case cfg.Auth.KeyID:
			return &privateKey.PublicKey, nil
		}
		return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
	}

	auth, err := auth.New(cfg.Auth.Algorithm, lookup, auth.Keys{cfg.Auth.KeyID: privateKey})
	if err != nil {
		return errors.Wrap(err, "constructing auth")
	}

	log.Println("main: Initializing database support")

	db, err := database.Open(database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer func() {
		log.Printf("main: Database Stopping : %s", cfg.DB.Host)
		db.Close()
	}()
	// ============================================================================
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package
	//
	// Not concerned with shutting this down when teh application is shutdown.

	log.Println("main: Initializing debugging support")

	go func() {

		log.Printf("main: Debug Listening %s", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			log.Printf("main: Debug Listener closed: %v", err)
		}
	}()

	// Start API Service

	log.Println("main: Initializing API support")

	// Make a channel to listen for interrupt or terminate signal from OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handlers.API(build, shutdown, log, auth, db),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error
	serverErrors := make(chan error, 1)

	go func(){
		log.Printf("main: API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()


	//Blocking main and waiting for shutdown

	select{
	case err:= <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig:= <-shutdown:
		log.Printf("main: %v : Start shutdown", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err!=nil{
			api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}


	}




	return nil

}
