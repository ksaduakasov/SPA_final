package handlers

import (
	"aitu/foundation/database"
	"aitu/foundation/web"
	"context"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type checkGroup struct {
	db *sqlx.DB
}

func (c checkGroup) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := "ok"
	statusCode := http.StatusOK
	if err := database.StatusCheck(ctx, c.db); err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
	}

	health := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	return web.Respond(ctx, w, health, statusCode)

}
