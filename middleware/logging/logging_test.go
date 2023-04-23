package logging

import (
	"fmt"
	"net/http"
	web "routing"
	"testing"
)

func TestLogBuilder(t *testing.T) {
	logging := NewLog(func(log string) {
		fmt.Println(log)
	})
	server := web.NewHttpServer(web.ServerWithMiddleware(logging.Build()))
	server.Get("/a/b/*", func(context *web.Context) {
		fmt.Println("hello, it's me")
	})
	req, err := http.NewRequest(http.MethodGet, "/a/b/c", nil)
	if err != nil {
		t.Fatal(err)
	}
	server.ServeHTTP(nil, req)
}
