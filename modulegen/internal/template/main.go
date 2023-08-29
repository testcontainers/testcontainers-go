package template

import (
	"io"
	"text/template"
)

func Generate(t *template.Template, wr io.Writer, name string, data any) error {
	err := t.ExecuteTemplate(wr, name, data)
	if err != nil {
		return err
	}
	return nil
}
