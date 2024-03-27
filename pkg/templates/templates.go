package templates

import (
	"bytes"
	"embed"
	"encoding/json"
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
			"add": func(a, b int) int { return a + b },
			"sub": func(a, b int) int { return a - b },
			"interval": func(n int) []int {
				if n <= 0 {
					return []int{}
				}

				interval := make([]int, n)

				for i := range interval {
					interval[i] = i
				}

				return interval
			},
			"prettyjson": func(s string) string {
				buf := &bytes.Buffer{}

				if err := json.Indent(buf, []byte(s), "", "  "); err != nil {
					return s
				}

				return buf.String()
			},
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
