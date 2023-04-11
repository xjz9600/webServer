//go:build e2e

package web

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := NewHttpServer()
	h.AddRoute(http.MethodGet, "/user", func(context *Context) {
		context.Resp.Write([]byte("HELLO WORD"))
	})
	h.Start(":8081")
}
