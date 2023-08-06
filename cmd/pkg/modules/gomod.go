package modules

import (
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func GenerateGomod(ctx *Context) error {
	rootData, err := os.ReadFile(ctx.RootGoModFile())
	if err != nil {
		return err
	}
	rootGoModFile, err := modfile.Parse(ctx.RootGoModFile(), rootData, nil)
	if err != nil {
		return err
	}
	tcPath := rootGoModFile.Module.Mod.Path

	file := &modfile.File{}
	file.AddModuleStmt(ctx.ModulePath(tcPath))
	file.AddGoStmt(rootGoModFile.Go.Version)

	file.AddRequire(tcPath, ctx.TCVersion)
	file.AddReplace(tcPath, "", "../..", "")

	data, err := file.Format()
	if err != nil {
		return err
	}

	exampleFilePath := ctx.GoModFile()
	err = os.MkdirAll(filepath.Dir(exampleFilePath), 0o777)
	if err != nil {
		return err
	}
	exampleFile, _ := os.Create(exampleFilePath)
	defer exampleFile.Close()
	_, err = exampleFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}
