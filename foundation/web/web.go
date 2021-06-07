//package web contains a small web framework (extending mux)
package web

import (
	"context"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"syscall"
	"time"
)

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

//App...
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

//NewApp...
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	app := App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw: mw,
	}
	return &app
}

type ctxKey int

const KeyValues ctxKey = 1

type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {
	handler  = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)


	h := func(w http.ResponseWriter, r *http.Request) {

		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}
		ctx := context.WithValue(r.Context(), KeyValues, &v)

		//BOILERPLATE
		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}

		//BOILERPLATE
	}
	a.ContextMux.Handle(method, path, h)
}

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}
