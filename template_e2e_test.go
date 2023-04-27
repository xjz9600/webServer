//go:build e2e

package web

import (
	"github.com/stretchr/testify/require"
	temp "html/template"
	"net/http"
	"routing/template"
	"testing"
)

func TestLogin(t *testing.T) {
	tpl, err := temp.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)
	h := NewHttpServer(ServerWithTemplateEngine(template.NewGoTemplateEngine(tpl)))
	h.AddRoute(http.MethodGet, "/retest/:id(re.+)", func(context *Context) {
		str, _ := context.PathValue("id").AsString()
		context.Render("login.gohtml", str)
	})
	h.Start(":8081")
}
