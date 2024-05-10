package modfile

import (
	"os"

	"golang.org/x/mod/modfile"
)

func readModFile(modFilePath string) (*modfile.File, error) {
	file, err := os.ReadFile(modFilePath)
	if err != nil {
		return nil, err
	}
	return modfile.Parse(modFilePath, file, nil)
}
