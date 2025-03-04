package modfile

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

func readModFile(modFilePath string) (*modfile.File, error) {
	file, err := os.ReadFile(modFilePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return modfile.Parse(modFilePath, file, nil)
}
