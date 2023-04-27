//go:build e2e

package file

import (
	"github.com/stretchr/testify/require"
	web "routing"
	"testing"
)

func TestStaticResourceBuilder(t *testing.T) {
	h := web.NewHttpServer()
	s, err := NewstaticResourceBuilder()
	require.NoError(t, err)
	h.Get("/static/:file", s.Handle)
	h.Start(":8082")
}
