package internal

import (
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/dependabot"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/make"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/tools"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/vscode"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/workflow"
)

func Generate(exampleVar context.ExampleVar, isModule bool) error {
	ctx, err := context.GetRootContext()
	if err != nil {
		return fmt.Errorf(">> could not get the root dir: %w", err)
	}

	example := context.Example{
		Image:     exampleVar.Image,
		IsModule:  isModule,
		Name:      exampleVar.Name,
		TitleName: exampleVar.NameTitle,
	}

	err = GenerateFiles(ctx, example)
	if err != nil {
		return fmt.Errorf(">> error generating the example: %w", err)
	}

	cmdDir := filepath.Join(ctx.RootDir, example.ParentDir(), example.Lower())
	err = tools.GoModTidy(cmdDir)
	if err != nil {
		return fmt.Errorf(">> error synchronizing the dependencies: %w", err)
	}
	err = tools.GoVet(cmdDir)
	if err != nil {
		return fmt.Errorf(">> error checking generated code: %w", err)
	}

	fmt.Println("Please go to", cmdDir, "directory to check the results, where 'go mod tidy' and 'go vet' was executed to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
	return nil
}

func GenerateFiles(ctx *context.Context, example context.Example) error {
	if err := example.Validate(); err != nil {
		return err
	}
	// creates Makefile for example
	err := make.GenerateMakefile(ctx, example)
	if err != nil {
		return err
	}

	err = module.GenerateGoModule(ctx, example)
	if err != nil {
		return err
	}

	// update github ci workflow
	err = workflow.GenerateWorkflow(ctx)
	if err != nil {
		return err
	}
	// update examples in mkdocs
	err = mkdocs.GenerateMkdocs(ctx, example)
	if err != nil {
		return err
	}
	// update examples in dependabot
	err = dependabot.GenerateDependabotUpdates(ctx, example)
	if err != nil {
		return err
	}
	// generate vscode workspace
	err = vscode.GenerateVSCodeWorkspace(ctx)
	if err != nil {
		return err
	}
	return nil
}
