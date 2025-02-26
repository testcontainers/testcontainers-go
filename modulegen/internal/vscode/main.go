package vscode

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

type Generator struct{}

// Generate updates the workspace for vscode
func (g Generator) Generate(ctx context.Context) error {
	examples, err := ctx.GetExamples()
	if err != nil {
		return err
	}
	modules, err := ctx.GetModules()
	if err != nil {
		return err
	}

	return writeConfig(ctx.VSCodeWorkspaceFile(), newConfig(examples, modules))
}

// Refresh refresh the vscode workspace
func (g Generator) Refresh(ctx context.Context, _ []context.TestcontainersModule) error {
	return g.Generate(ctx)
}
