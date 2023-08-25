package main

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
)

// update examples in mkdocs
func generateMkdocs(ctx *Context, example Example) error {
	exampleMd := example.ParentDir() + "/" + example.Lower() + ".md"
	indexMd := example.ParentDir() + "/index.md"
	return mkdocs.UpdateConfig(ctx.MkdocsConfigFile(), example.IsModule, exampleMd, indexMd)
}
