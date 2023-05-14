package mirror

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateEngine interface {
	// Render 渲染页面
	Render(ctx context.Context, templateName string, data any) ([]byte, error)
}

type GoTemplateEngine struct {
	T *template.Template
}

func (e *GoTemplateEngine) Render(ctx context.Context, templateName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := e.T.ExecuteTemplate(bs, templateName, data)
	return bs.Bytes(), err
}
