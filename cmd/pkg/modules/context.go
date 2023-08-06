package modules

import (
	"path/filepath"
	"strings"

	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

type Context struct {
	*pkg_cmd.RootContext
	TCVersion string
}

func (ctx *Context) ModulePath(tcPath string) string {
	return tcPath + "/" + ctx.RootContext.ParentDir() + "/" + ctx.RootContext.Lower()
}

func (ctx *Context) GoModFile() string {
	return filepath.Join(ctx.RootContext.ExampleDirectory(), "go.mod")
}

func (ctx *Context) RootGoModFile() string {
	return filepath.Join(ctx.RootContext.RootDir, "go.mod")
}

func (ctx *Context) templateFile(tmpl string) string {
	return filepath.Join(ctx.RootContext.ExampleDirectory(), strings.ReplaceAll(tmpl, "example", ctx.RootContext.Lower()))
}
