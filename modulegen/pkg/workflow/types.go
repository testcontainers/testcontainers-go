package workflow

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Directories struct {
	Examples string
	Modules  string
}

func GetDirectories(rootDir string) *Directories {
	exampleList, err := GetProjectsAsString(rootDir, "examples")
	if err != nil {
		exampleList = ""
	}
	moduleList, err := GetProjectsAsString(rootDir, "modules")
	if err != nil {
		moduleList = ""
	}

	return &Directories{
		Examples: exampleList,
		Modules:  moduleList,
	}
}

func GetProjectsAsString(rootDir string, baseDir string) (string, error) {
	dirs, err := GetProjects(rootDir, baseDir)
	if err != nil {
		return "", err
	}

	// sort the dir names by name
	names := make([]string, len(dirs))
	for i, f := range dirs {
		names[i] = f.Name()
	}

	sort.Strings(names)

	return strings.Join(names, ", "), nil
}

func GetProjects(rootDir string, baseDir string) ([]os.DirEntry, error) {
	dirs := make([]os.DirEntry, 0)
	dir := filepath.Join(rootDir, baseDir)

	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range allFiles {
		if f.IsDir() {
			dirs = append(dirs, f)
		}
	}
	return dirs, nil
}
