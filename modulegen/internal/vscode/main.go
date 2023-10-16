package vscode

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

type Generator struct{}

// Generate updates the workspace for vscode
func (g Generator) Generate(ctx context.Context) error {
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

	return writeConfig(ctx.VSCodeWorkspaceFile(), newConfig(examples, modules))
}
