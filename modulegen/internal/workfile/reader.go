package workfile

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

func Read(workFilePath string) (*modfile.WorkFile, error) {
	file, err := os.ReadFile(workFilePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return modfile.ParseWork(workFilePath, file, nil)
}
