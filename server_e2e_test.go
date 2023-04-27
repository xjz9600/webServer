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
	h.AddRoute(http.MethodGet, "/retest/:id(re.+)", func(context *Context) {
		http.SetCookie(context.Resp, &http.Cookie{Name: "aaa", Value: "bbb"})
		context.Resp.Write([]byte("HELLO WORD"))
	}, func(next HandleFunc) HandleFunc {
		return func(context *Context) {
			fmt.Println("second begin")
			next(context)
			fmt.Println("second end")
		}
	})
	h.Start(":8081")
}
