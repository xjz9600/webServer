//go:build e2e

package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	web "routing"
	"testing"
	"time"
)

func TestPrometheusBuilderE2E(t *testing.T) {
	builder := NewPrometheusBuilder("http_response", "web", "geekbang", "firstPrometheus")
	h := web.NewHttpServer(web.ServerWithMiddleware(builder.Build()))
	h.AddRoute(http.MethodGet, "/user/:id", func(context *web.Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Duration(val) * time.Millisecond)
		context.RespJson(200, User{
			Name: "Tom",
		})
	})
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8085", nil)
	}()
	h.Start(":8083")
}

type User struct {
	Name string
}
