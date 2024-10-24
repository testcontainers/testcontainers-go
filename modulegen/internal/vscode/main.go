package vscode

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
)

type Generator struct{}

// Generate updates the workspace for vscode
func (g Generator) Generate(ctx context.Context) error {
	examples, modules, err := module.ListExamplesAndModules(ctx)
	if err != nil {
		return err
	}

	return writeConfig(ctx.VSCodeWorkspaceFile(), newConfig(examples, modules))
}
