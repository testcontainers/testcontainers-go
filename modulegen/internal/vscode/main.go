package vscode

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

// print out the workspace for vscode
func GenerateVSCodeWorkspace(ctx *context.Context) error {
	rootCtx, err := context.GetRootContext()
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

	return Generate(ctx.VSCodeWorkspaceFile(), examples, modules)
}

func Generate(exampleFilePath string, examples []string, modules []string) error {
	config := newConfig(examples, modules)

	return writeConfig(exampleFilePath, config)
}
