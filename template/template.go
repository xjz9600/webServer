package template

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateEngine interface {
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}

type goTemplateEngine struct {
	temp *template.Template
}

func NewGoTemplateEngine(temp *template.Template) *goTemplateEngine {
	return &goTemplateEngine{temp: temp}
}
func (g *goTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.temp.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}
