//go:build e2e

package file

import (
	web "routing"
	"testing"
)

func TestDownload(t *testing.T) {
	h := web.NewHttpServer()
	f := NewDownLoader()
	h.Get("/download", f.Handle)
	h.Start(":8081")
}
