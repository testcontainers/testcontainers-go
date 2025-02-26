package vscode

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

// Generator is a struct that contains the logic to generate the vscode workspace file.
type Generator struct{}

// Generate updates the workspace for vscode
func (g Generator) Generate(ctx context.Context, examples []string, modules []string) error {
	return writeConfig(ctx.VSCodeWorkspaceFile(), newConfig(examples, modules))
}
