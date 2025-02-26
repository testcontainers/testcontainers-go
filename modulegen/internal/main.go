package internal

import (
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/dependabot"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/make"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/sonar"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/tools"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/vscode"
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

func Refresh(ctx context.Context) error {
	var modulesAndExamples []context.TestcontainersModule

	modules, err := ctx.GetModules()
	if err != nil {
		return fmt.Errorf("get modules: %w", err)
	}
	for _, module := range modules {
		tcModule := context.TestcontainersModule{
			Image:     "",
			IsModule:  true,
			Name:      module,
			TitleName: "",
		}
		modulesAndExamples = append(modulesAndExamples, tcModule)
	}

	examples, err := ctx.GetExamples()
	if err != nil {
		return fmt.Errorf("get examples: %w", err)
	}

	for _, example := range examples {
		tcModule := context.TestcontainersModule{
			Image:     "",
			IsModule:  false,
			Name:      example,
			TitleName: "",
		}
		modulesAndExamples = append(modulesAndExamples, tcModule)
	}

	refreshers := []ModuleRefresher{
		mkdocs.Generator{},     // update examples in mkdocs
		dependabot.Generator{}, // update examples in dependabot
		vscode.Generator{},     // update vscode workspace
		sonar.Generator{},      // update sonar-project.properties
	}

	for _, refresher := range refreshers {
		err := refresher.Refresh(ctx, modulesAndExamples)
		if err != nil {
			return fmt.Errorf("refresh modules: %w", err)
		}
	}

	fmt.Println("Modules and examples refreshed.")
	fmt.Println("🙏 Commit the modified files and submit a pull request to include them into the project.")
	fmt.Println("Thanks!")
	return nil
}

type ProjectGenerator interface {
	Generate(context.Context) error
}
type FileGenerator interface {
	AddModule(context.Context, context.TestcontainersModule) error
}

type ModuleRefresher interface {
	Refresh(context.Context, []context.TestcontainersModule) error
}

func GenerateFiles(ctx context.Context, tcModule context.TestcontainersModule) error {
	if err := tcModule.Validate(); err != nil {
		return err
	}

	fileGenerators := []FileGenerator{
		make.Generator{},       // creates Makefile for module
		module.Generator{},     // creates go.mod for module
		mkdocs.Generator{},     // update examples in mkdocs
		dependabot.Generator{}, // update examples in dependabot
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
		vscode.Generator{}, // update vscode workspace
		sonar.Generator{},  // update sonar-project.properties
	}

	for _, generator := range projectGenerators {
		err := generator.Generate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
