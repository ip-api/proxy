package handlers_test

import (
	"testing"

	"github.com/ip-api/proxy/handlers"
	"github.com/valyala/fasthttp"
)

func Test404(t *testing.T) {
	h := handlers.Handler{}

	var ctx fasthttp.RequestCtx

	for _, path := range []string{
		"/",
		"/bla",
		"/jsons",
		"/jsons/",
		"/batch/",
		"/batchasd",
	} {
		t.Run(path, func(t *testing.T) {
			ctx.Request.SetRequestURI(path)
			h.Index(&ctx)
			if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
				t.Errorf("expected 404 got %d", ctx.Response.StatusCode())
			}
		})
	}
}
