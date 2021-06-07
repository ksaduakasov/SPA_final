package handlers

import (
	"aitu/business/auth"
	"aitu/business/mid"
	"aitu/foundation/web"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"os"
)

func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth, db *sqlx.DB) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log),mid.Metrics(), mid.Panics(log))

	check := checkGroup{
		db: db,
	}

	app.Handle(http.MethodGet, "/readiness", check.readiness, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))

	return app
}
