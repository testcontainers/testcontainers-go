package main

import (
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/make"
)

// creates Makefile for example
func generateMakefile(ctx *Context, example Example) error {
	exampleDir := filepath.Join(ctx.RootDir, example.ParentDir(), example.Lower())
	exampleName := example.Lower()
	return make.Generate(exampleDir, exampleName)
}
