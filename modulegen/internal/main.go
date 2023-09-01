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

func Generate(moduleVar context.TestcontainersModuleVar, isModule bool) error {
	ctx, err := context.GetRootContext()
	if err != nil {
		return fmt.Errorf(">> could not get the root dir: %w", err)
	}

	m := context.TestcontainersModule{
		Image:     moduleVar.Image,
		IsModule:  isModule,
		Name:      moduleVar.Name,
		TitleName: moduleVar.NameTitle,
	}

	err = GenerateFiles(ctx, m)
	if err != nil {
		return fmt.Errorf(">> error generating the module: %w", err)
	}

	cmdDir := filepath.Join(ctx.RootDir, m.ParentDir(), m.Lower())
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

type ProjectGenerator interface {
	Generate(*context.Context) error
}
type FileGenerator interface {
	AddModule(*context.Context, context.TestcontainersModule) error
}

func GenerateFiles(ctx *context.Context, m context.TestcontainersModule) error {
	if err := m.Validate(); err != nil {
		return err
	}

	err := module.GenerateGoModule(ctx, m)
	if err != nil {
		return err
	}

	// update github ci workflow
	err = workflow.GenerateWorkflow(ctx)
	if err != nil {
		return err
	}
	// update examples in mkdocs
	err = mkdocs.GenerateMkdocs(ctx, m)
	if err != nil {
		return err
	}

	generators := []FileGenerator{
		make.Generator{},       // creates Makefile for module
		dependabot.Generator{}, // update examples in dependabot
	}

	for _, generator := range generators {
		err := generator.AddModule(ctx, m)
		if err != nil {
			return err
		}
	}

	// generate vscode workspace
	err = vscode.GenerateVSCodeWorkspace(ctx)
	if err != nil {
		return err
	}
	return nil
}
