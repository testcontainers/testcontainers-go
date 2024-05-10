package modfile

import (
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func GenerateModFile(exampleDir string, rootGoModFilePath string, directory string, tcVersion string) error {
	rootGoMod, err := readModFile(rootGoModFilePath)
	if err != nil {
		return err
	}
	moduleStmt := rootGoMod.Module.Mod.Path + directory
	goStmt := rootGoMod.Go.Version
	tcPath := rootGoMod.Module.Mod.Path
	file, err := newModFile(moduleStmt, goStmt, tcPath, tcVersion)
	if err != nil {
		return err
	}
	return writeModFile(filepath.Join(exampleDir, "go.mod"), file)
}

func newModFile(moduleStmt string, goStmt string, tcPath string, tcVersion string) (*modfile.File, error) {
	file := &modfile.File{}
	err := file.AddModuleStmt(moduleStmt)
	if err != nil {
		return nil, err
	}
	err = file.AddGoStmt(goStmt)
	if err != nil {
		return nil, err
	}
	err = file.AddRequire(tcPath, tcVersion)
	if err != nil {
		return nil, err
	}
	err = file.AddReplace(tcPath, "", "../..", "")
	if err != nil {
		return nil, err
	}
	return file, nil
}
