package render

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
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
		buf := bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
		return buf
	}
	return e, nil
}

func (e *Engine) getBuffer() *bytes.Buffer {
	buf := e.pool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func (e *Engine) putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 1<<20 {
		return
	}
	e.pool.Put(buf)
}

func (e *Engine) RenderPage(w io.Writer, contentName, layoutName string, view any, setBody func([]byte)) error {
	buf := e.getBuffer()
	if err := e.templates.ExecuteTemplate(buf, contentName, view); err != nil {
		e.putBuffer(buf)
		return err
	}
	if setBody != nil {
		setBody(buf.Bytes())
	}
	buf.Reset()
	if err := e.templates.ExecuteTemplate(w, layoutName, view); err != nil {
		e.putBuffer(buf)
		return err
	}
	e.putBuffer(buf)
	return nil
}

func (e *Engine) ExecuteBytes(name string, data any) ([]byte, error) {
	buf := e.getBuffer()
	defer e.putBuffer(buf)

	if err := e.templates.ExecuteTemplate(buf, name, data); err != nil {
		return nil, err
	}
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func StaticFS() fs.FS {
	return templateFS
}
