package vscode

import (
	"path/filepath"
)

func Generate(rootDir string, examples []string, modules []string) error {
	config := newConfig(examples, modules)

	exampleFilePath := filepath.Join(rootDir, ".vscode", ".testcontainers-go.code-workspace")

	return writeConfig(exampleFilePath, config)
}
