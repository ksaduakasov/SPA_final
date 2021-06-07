package mid

import (
	"aitu/foundation/web"
	"context"
	"log"
	"net/http"
)

func Errors(log *log.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler{
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}
			if err := handler(ctx, w, r); err != nil {
				log.Printf("%s : ERROR     : %v",v.TraceID,err)
				if err := web.RespondError(ctx, w, err); err != nil {
					return err
				}
				if ok := web.IsShutdown(err); ok {
					return err
				}
			}
			return nil
		}
		return h
	}
	return m
}