package dependabot

import (
	"path/filepath"

	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

type Context struct {
	*pkg_cmd.RootContext
}

func (ctx *Context) Directory() string {
	return "/" + ctx.RootContext.ParentDir() + "/" + ctx.RootContext.Lower()
}

func (ctx *Context) ConfigFile() string {
	return filepath.Join(ctx.RootContext.RootDir, ".github", "dependabot.yml")
}
