package modfile

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func writeModFile(modFilePath string, file *modfile.File) error {
	err := os.MkdirAll(filepath.Dir(modFilePath), 0o755)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	data, err := file.Format()
	if err != nil {
		return fmt.Errorf("format file: %w", err)
	}
	return os.WriteFile(modFilePath, data, 0o644)
}
