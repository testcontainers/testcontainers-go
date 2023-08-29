package main

import (
	"os"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/vscode"
)

// print out the workspace for vscode
func generateVSCodeWorkspace(ctx *Context) error {
	rootCtx, err := getRootContext()
	if err != nil {
		return err
	}
	examples, err := rootCtx.GetExamples()
	if err != nil {
		return err
	}
	modules, err := rootCtx.GetModules()
	if err != nil {
		return err
	}

	return vscode.Generate(os.Stdout, examples, modules)
}
