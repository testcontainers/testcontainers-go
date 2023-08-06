package make

import (
	"path/filepath"
	"text/template"

	templates "github.com/testcontainers/testcontainers-go/cmd/pkg/templates"
)

func GenerateFromContext(ctx *Context) error {
	name := "Makefile.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_templates", "make", name))
	if err != nil {
		return err
	}
	return templates.Generate(t, ctx.makeFile(), name, ctx.Lower())
}
