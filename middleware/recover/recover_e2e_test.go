//go:build e2e

package recover

import (
	"fmt"
	web "routing"
	"testing"
)

func TestRecoverBuilderE2E(t *testing.T) {
	builder := NewRecoverBuilder(500, []byte(`发生 painc 了`), func(ctx *web.Context) {
		fmt.Printf("panic 路径: %s", ctx.Req.URL.String())
	})
	server := web.NewHttpServer(web.ServerWithMiddleware(builder.Build()))
	server.Get("/user", func(context *web.Context) {
		panic("我挂了")
	})
	server.Start(":8088")
}
