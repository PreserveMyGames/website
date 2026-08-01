package render

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"sync"
)

const defaultBufferSize = 16 << 10

//go:embed templates/*.html
var templateFS embed.FS

type Engine struct {
	templates *template.Template
	pool      sync.Pool
}

func New(funcs template.FuncMap) (*Engine, error) {
	tmpl, err := template.New("root").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	e := &Engine{templates: tmpl}
	e.pool.New = func() any {
		return bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	}
	return e, nil
}

func (e *Engine) RenderPage(w io.Writer, contentName string, view any, setBody func([]byte)) error {
	buf := e.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer e.put(buf)

	if err := e.templates.ExecuteTemplate(buf, contentName, view); err != nil {
		return err
	}
	if setBody != nil {
		setBody(buf.Bytes())
	}
	buf.Reset()
	return e.templates.ExecuteTemplate(w, "layout", view)
}

func (e *Engine) put(buf *bytes.Buffer) {
	if buf.Cap() > 1<<20 {
		return
	}
	e.pool.Put(buf)
}
