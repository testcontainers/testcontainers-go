package context

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type RootContext struct {
	Image    string
	IsModule bool
	Name     string
	RootDir  string
	Title    string
}

func (ctx *RootContext) Lower() string {
	return strings.ToLower(ctx.Name)
}

func (ctx *RootContext) ParentDir() string {
	if ctx.IsModule {
		return "modules"
	}
	return "examples"
}

func (ctx *RootContext) ExampleDirectory() string {
	return filepath.Join(ctx.RootDir, ctx.ParentDir(), ctx.Lower())
}

func (ctx *RootContext) MkDocsConfigFile() string {
	return filepath.Join(ctx.RootDir, "mkdocs.yml")
}

func (rc *RootContext) Validate() error {
	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(rc.Name) {
		return fmt.Errorf("invalid name: %s. Only alphanumerical characters are allowed (leading character must be a letter)", rc.Name)
	}

	if len(rc.Title) > 0 {
		if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(rc.Title) {
			return fmt.Errorf("invalid title: %s. Only alphanumerical characters are allowed (leading character must be a letter)", rc.Title)
		}
	}
	return nil
}
