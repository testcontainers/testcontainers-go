package internal

import (
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/make"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/sonar"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/tools"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/vscode"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/workflow"
)

func Generate(moduleVar context.TestcontainersModuleVar, isModule bool) error {
	ctx, err := context.GetRootContext()
	if err != nil {
		return fmt.Errorf(">> could not get the root dir: %w", err)
	}

	tcModule := context.TestcontainersModule{
		Image:     moduleVar.Image,
		IsModule:  isModule,
		Name:      moduleVar.Name,
		TitleName: moduleVar.NameTitle,
	}

	err = GenerateFiles(ctx, tcModule)
	if err != nil {
		return fmt.Errorf(">> error generating the module: %w", err)
	}

	cmdDir := filepath.Join(ctx.RootDir, tcModule.ParentDir(), tcModule.Lower())
	lintCmds := []func(string) error{
		tools.GoModTidy,
		tools.GoVet,
		tools.MakeLint,
	}

	for _, lintCmd := range lintCmds {
		err = lintCmd(cmdDir)
		if err != nil {
			return err
		}
	}

	fmt.Println("Please go to", cmdDir, "directory to check the results, where 'go mod tidy', 'go vet' and 'make lint' were executed.")
	fmt.Println("🙏 Commit the modified files and submit a pull request to include them into the project.")
	fmt.Println("Remember to run 'make lint' before submitting the pull request.")
	fmt.Println("Thanks!")
	return nil
}

type ProjectGenerator interface {
	Generate(context.Context) error
}
type FileGenerator interface {
	AddModule(context.Context, context.TestcontainersModule) error
}

func GenerateFiles(ctx context.Context, tcModule context.TestcontainersModule) error {
	if err := tcModule.Validate(); err != nil {
		return err
	}

	fileGenerators := []FileGenerator{
		make.Generator{},   // creates Makefile for module
		module.Generator{}, // creates go.mod for module
		mkdocs.Generator{}, // update examples in mkdocs
	}

	for _, generator := range fileGenerators {
		err := generator.AddModule(ctx, tcModule)
		if err != nil {
			return err
		}
	}

	// they are based on the content of the modules in the project workspace,
	// not in the new module to be added, that's why they happen after the actual
	// module generation
	projectGenerators := []ProjectGenerator{
		workflow.Generator{}, // update github ci workflow
		vscode.Generator{},   // update vscode workspace
		sonar.Generator{},    // update sonar-project.properties
	}

	for _, generator := range projectGenerators {
		err := generator.Generate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
