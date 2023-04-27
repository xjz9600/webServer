//go:build e2e

package file

import (
	"github.com/stretchr/testify/require"
	te "html/template"
	"log"
	web "routing"
	"routing/template"
	"testing"
)

func TestUpload(t *testing.T) {
	tpl, err := te.ParseGlob("../testdata/tpls/*.gohtml")
	require.NoError(t, err)
	engine := template.NewGoTemplateEngine(tpl)
	h := web.NewHttpServer(web.ServerWithTemplateEngine(engine))
	h.Get("/upload", func(ctx *web.Context) {
		err := ctx.Render("upload.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})
	f := NewFileUploader()
	h.Post("/upload", f.Handle)
	h.Start(":8081")
}
