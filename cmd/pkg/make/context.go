package make

import (
	"path/filepath"

	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

type Context struct {
	*pkg_cmd.RootContext
}

func (ctx *Context) makeFile() string {
	return filepath.Join(ctx.RootContext.ExampleDirectory(), "Makefile")
}
