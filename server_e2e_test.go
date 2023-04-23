//go:build e2e

package web

import (
	"fmt"
	"net/http"
	"testing"
)

var TestMap = make(map[string]string)

func TestServer(t *testing.T) {
	h := NewHttpServer()
	h.AddRoute(http.MethodGet, "/retest/(re.+)", func(context *Context) {
		http.SetCookie(context.Resp, &http.Cookie{Name: "aaa", Value: "bbb"})
		context.Resp.Write([]byte("HELLO WORD"))
	}, func(next HandleFunc) HandleFunc {
		return func(context *Context) {
			fmt.Println("second begin")
			next(context)
			fmt.Println("second end")
		}
	})
	h.AddRoute(http.MethodGet, "/retest/r/aa", func(context *Context) {
		context.Resp.Write([]byte("HELLO haha tree"))
	}, func(next HandleFunc) HandleFunc {
		return func(context *Context) {
			fmt.Println("third begin")
			next(context)
			fmt.Println("third end")
		}
	})
	h.AddRoute(http.MethodGet, "/*", func(context *Context) {
		context.Resp.Write([]byte("HELLO start"))
	}, func(next HandleFunc) HandleFunc {
		return func(context *Context) {
			fmt.Println("first begin")
			next(context)
			fmt.Println("first end")
		}
	})
	h.Start(":8081")
}
