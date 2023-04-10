//go:build e2e

package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := &HTTPServer{}
	h.AddRoute(http.MethodGet, "/user", func(context Context) {
		fmt.Println("处理事件")
	})
	h.Start(":8081")
}
