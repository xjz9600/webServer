//go:build e2e

package web

import (
	"net/http"
	"testing"
)

var TestMap = make(map[string]string)

func TestServer(t *testing.T) {
	h := NewHttpServer()
	h.AddRoute(http.MethodPost, "/user/:id", func(context *Context) {
		http.SetCookie(context.Resp, &http.Cookie{Name: "aaa", Value: "bbb"})
		context.Resp.Write([]byte("HELLO WORD"))
	})
	h.AddRoute(http.MethodGet, "/retest/(re.+)", func(context *Context) {
		context.Resp.Write([]byte("HELLO RE tree"))
	})
	h.Start(":8081")
}
