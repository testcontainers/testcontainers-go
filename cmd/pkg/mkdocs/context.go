package mkdocs

import (
	"path/filepath"

	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

type Context struct {
	*pkg_cmd.RootContext
}

func (ctx *Context) ExampleMd() string {
	return ctx.RootContext.ParentDir() + "/" + ctx.RootContext.Lower() + ".md"
}

func (ctx *Context) IndexMd() string {
	return ctx.RootContext.ParentDir() + "/index.md"
}

func (ctx *Context) docDir() string {
	return filepath.Join(ctx.RootContext.RootDir, "docs", ctx.RootContext.ParentDir())
}
