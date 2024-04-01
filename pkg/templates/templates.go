package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"slices"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

//go:embed **.html **/**.html
var fs embed.FS

type Templates struct {
	m    *minify.M
	tmpl *template.Template
}

func New() (*Templates, error) {
	m := minify.New()
	m.Add("text/html", &html.Minifier{})

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

	return &Templates{
		m:    m,
		tmpl: tmpl,
	}, nil
}

func (t *Templates) Render(w io.Writer, name string, data any) error {
	wc := t.m.Writer("text/html", w)
	defer wc.Close()
	return t.tmpl.ExecuteTemplate(wc, name, data)
}

func (t *Templates) RenderSSR(w io.Writer, event string, name string, data any) error {
	if _, err := w.Write([]byte(fmt.Sprintf("event: %s\ndata: ", event))); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	if err := t.Render(w, name, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}
	return nil
}
