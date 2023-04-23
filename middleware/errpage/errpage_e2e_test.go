//go:build e2e

package errpage

import (
	"net/http"
	web "routing"
	"testing"
)

func TestErrPageE2E(t *testing.T) {
	builder := NewErrPageBuilder()
	builder.AddErrPage(http.StatusNotFound, []byte(`
<html>
	<body>
		<h1>哈哈哈，走失了</h1>
	</body>
</html>`))
	server := web.NewHttpServer(web.ServerWithMiddleware(builder.Build()))
	server.Start(":8086")
}
