package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"slices"
)

//go:embed **.html **/**.html
var fs embed.FS

type Templates struct {
	tmpl *template.Template
}

func New() (*Templates, error) {
	tmpl, err := template.
		New("").
		Funcs(template.FuncMap{
			"slicecontains": func(ss []string, s string) bool { return slices.Contains(ss, s) },
		}).
		ParseFS(fs, "**.html", "**/**.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return &Templates{tmpl: tmpl}, nil
}

func (t *Templates) Render(w io.Writer, name string, data any) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}
