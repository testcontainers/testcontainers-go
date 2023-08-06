package mkdocs

import (
	html_template "html/template"
	"path/filepath"
	text_template "text/template"

	templates "github.com/testcontainers/testcontainers-go/cmd/pkg/templates"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Example struct {
	Image     string
	Name      string
	ParentDir string
	Title     string
}

func GenerateDocFromContext(ctx *Context) error {
	example := newExample(ctx)
	funcMap := text_template.FuncMap{
		"codeinclude": func(s string) html_template.HTML { return html_template.HTML(s) }, // escape HTML comments for codeinclude
	}
	name := "example.md.tmpl"
	t, err := text_template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_templates", "mkdocs", name))
	if err != nil {
		return err
	}
	// docs example file will go into the docs directory
	exampleFilePath := filepath.Join(ctx.docDir(), example.Name+".md")
	return templates.Generate(t, exampleFilePath, name, example)
}

func newExample(ctx *Context) *Example {
	example := &Example{
		Image:     ctx.RootContext.Image,
		Name:      ctx.RootContext.Lower(),
		ParentDir: ctx.RootContext.ParentDir(),
	}
	example.setTitle(ctx)
	return example
}

func (example *Example) setTitle(ctx *Context) {
	if ctx.RootContext.Title != "" {
		example.Title = ctx.RootContext.Title
	} else {
		example.Title = cases.Title(language.Und, cases.NoLower).String(ctx.RootContext.Lower())
	}
}
