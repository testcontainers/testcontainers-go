package modules

import (
	"path/filepath"
	text_template "text/template"
	"unicode"
	"unicode/utf8"

	templates "github.com/testcontainers/testcontainers-go/cmd/pkg/templates"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Example struct {
	Image         string
	Name          string
	Title         string
	ContainerName string
	ParentDir     string
}

func GenerateFilesFromContext(ctx *Context) error {
	example := newExample(ctx)
	for _, tmpl := range []string{"example_test.go", "example.go"} {
		name := tmpl + ".tmpl"
		t, err := text_template.New(name).ParseFiles(filepath.Join("_templates", "modules", name))
		if err != nil {
			return err
		}
		// docs example file will go into the docs directory
		exampleFilePath := ctx.templateFile(tmpl)

		err = templates.Generate(t, exampleFilePath, name, example)
		if err != nil {
			return err
		}
	}
	return nil
}

func newExample(ctx *Context) *Example {
	example := &Example{
		Image:     ctx.RootContext.Image,
		Name:      ctx.RootContext.Lower(),
		ParentDir: ctx.RootContext.ParentDir(),
	}
	example.setTitle(ctx)
	example.setContainerName(ctx)
	return example
}

func (example *Example) setTitle(ctx *Context) {
	if ctx.RootContext.Title != "" {
		example.Title = ctx.RootContext.Title
	} else {
		example.Title = cases.Title(language.Und, cases.NoLower).String(ctx.RootContext.Lower())
	}
}

// setContainerName sets the name of the container, which is the lower-cased title of the example
// If the title is set, it will be used instead of the name
func (example *Example) setContainerName(ctx *Context) {
	name := ctx.RootContext.Lower()

	if ctx.RootContext.IsModule {
		name = example.Title
	} else {
		if ctx.RootContext.Title != "" {
			r, n := utf8.DecodeRuneInString(ctx.RootContext.Title)
			name = string(unicode.ToUpper(r)) + ctx.RootContext.Title[n:]
		}
	}

	example.ContainerName = name + "Container"
}
