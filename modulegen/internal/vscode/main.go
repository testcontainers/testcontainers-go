package vscode

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
)

type Generator struct{}

// Generate updates the workspace for vscode
func (g Generator) Generate(ctx context.Context) error {
	examples, modules, err := module.ListExamplesAndModules(ctx)
	if err != nil {
		return fmt.Errorf("list examples and modules: %w", err)
	}

	return writeConfig(ctx.VSCodeWorkspaceFile(), newConfig(examples, modules))
}
