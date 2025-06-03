package modfile

import (
	"fmt"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// GenerateModFile generates a go.mod file for a module or example.
func GenerateModFile(exampleDir string, rootGoModFilePath string, directory string, tcVersion string) error {
	rootGoMod, err := readModFile(rootGoModFilePath)
	if err != nil {
		return fmt.Errorf("read mod file: %w", err)
	}
	moduleStmt := rootGoMod.Module.Mod.Path + directory
	goStmt := rootGoMod.Go.Version
	tcPath := rootGoMod.Module.Mod.Path
	file, err := newModFile(moduleStmt, goStmt, tcPath, tcVersion)
	if err != nil {
		return fmt.Errorf("new mod file: %w", err)
	}
	return writeModFile(filepath.Join(exampleDir, "go.mod"), file)
}

func newModFile(moduleStmt string, goStmt string, tcPath string, tcVersion string) (*modfile.File, error) {
	file := &modfile.File{}
	err := file.AddModuleStmt(moduleStmt)
	if err != nil {
		return nil, fmt.Errorf("add module stmt: %w", err)
	}
	err = file.AddGoStmt(goStmt)
	if err != nil {
		return nil, fmt.Errorf("add go stmt: %w", err)
	}
	err = file.AddRequire(tcPath, tcVersion)
	if err != nil {
		return nil, fmt.Errorf("add require: %w", err)
	}
	err = file.AddReplace(tcPath, "", "../..", "")
	if err != nil {
		return nil, fmt.Errorf("add replace: %w", err)
	}
	return file, nil
}
