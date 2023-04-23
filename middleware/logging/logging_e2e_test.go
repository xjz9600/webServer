//go:build e2e

package logging

import (
	"fmt"
	"net/http"
	web "routing"
	"testing"
)

func TestLogBuilderE2E(t *testing.T) {
	h := web.NewHttpServer(web.ServerWithMiddleware(NewLog(func(log string) {
		fmt.Println(log)
	}).Build()))
	h.AddRoute(http.MethodPost, "/user/:id", func(context *web.Context) {
		http.SetCookie(context.Resp, &http.Cookie{Name: "aaa", Value: "bbb"})
		context.Resp.Write([]byte("HELLO WORD"))
	})
	h.AddRoute(http.MethodGet, "/retest/(re.+)", func(context *web.Context) {
		context.Resp.Write([]byte("HELLO RE tree"))
	})
	h.Start(":8082")
}
