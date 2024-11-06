package workfile

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// Write writes the given go.work to the given path
func Write(workFilePath string, file *modfile.WorkFile) error {
	err := os.MkdirAll(filepath.Dir(workFilePath), 0o755)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	data := modfile.Format(file.Syntax)

	return os.WriteFile(workFilePath, data, 0o644)
}
