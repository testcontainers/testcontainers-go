package image

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/patternmatcher/ignorefile"
)

// ParseDockerIgnore returns if the file exists, the excluded files and an error if any
func ParseDockerIgnore(targetDir string) (bool, []string, error) {
	// based on https://github.com/docker/cli/blob/master/cli/command/image/build/dockerignore.go#L14
	fileLocation := filepath.Join(targetDir, ".dockerignore")
	var excluded []string
	exists := false
	if f, openErr := os.Open(fileLocation); openErr == nil {
		defer f.Close()

		exists = true

		var err error
		excluded, err = ignorefile.ReadAll(f)
		if err != nil {
			return true, excluded, fmt.Errorf("error reading .dockerignore: %w", err)
		}
	}
	return exists, excluded, nil
}
