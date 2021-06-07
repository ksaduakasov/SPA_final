package mid

import (
	"aitu/foundation/web"
	"context"
	"log"
	"net/http"
	"time"
)


func Logger(log *log.Logger) web.Middleware{

	m:=func(handler web.Handler) web.Handler{
		h :=  func(ctx context.Context, w http.ResponseWriter, r *http.Request) error{

			//BOILERPLATE-Logging

			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok{
				return web.NewShutdownError("web value missing from content")
			}

			log.Printf("%s : started   : %s %s -> %s", v.TraceID, r.Method, r.URL.Path, r.RemoteAddr)

			err := handler(ctx, w, r)

			log.Printf("%s : completed : %s %s -> %s (%d) (%s)", v.TraceID, r.Method, r.URL.Path, r.RemoteAddr, v.StatusCode, time.Since(v.Now))



			//BOILERPLATE-Logging


			return err
		}
		return h
	}
	return m
}
