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
)

// Generate generates all the files for a module or example,
// running the `go mod tidy`, `go vet` and `make lint` commands
// in the given directory.
func Generate(moduleVar context.TestcontainersModuleVar, isModule bool) error {
	ctx, err := context.GetRootContext()
	if err != nil {
		return fmt.Errorf("get root context: %w", err)
	}

	tcModule := context.TestcontainersModule{
		Image:     moduleVar.Image,
		IsModule:  isModule,
		Name:      moduleVar.Name,
		TitleName: moduleVar.NameTitle,
	}

	err = GenerateFiles(ctx, tcModule)
	if err != nil {
		return fmt.Errorf("generate files for module %s: %w", tcModule.Name, err)
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

// Refresh refreshes the modules and examples, returning an error if something goes wrong.
func Refresh(ctx context.Context) error {
	modules, err := ctx.GetModules()
	if err != nil {
		return fmt.Errorf("get modules: %w", err)
	}

	examples, err := ctx.GetExamples()
	if err != nil {
		return fmt.Errorf("get examples: %w", err)
	}

	generators := []ProjectGenerator{
		mkdocs.Generator{},     // update examples in mkdocs
		dependabot.Generator{}, // update examples in dependabot
		vscode.Generator{},     // update vscode workspace
	}

	for _, generator := range generators {
		err := generator.Generate(ctx, examples, modules)
		if err != nil {
			return fmt.Errorf("refresh modules: %w", err)
		}
	}

	return nil
}

// ProjectGenerator is the interface for the project generators, which takes
// a context and for each module in the context, adds it to the project files,
// returning an error if something goes wrong.
type ProjectGenerator interface {
	Generate(ctx context.Context, examples []string, modules []string) error
}

// FileGenerator is the interface for the file generators, which takes
// a module and generate a file for it, returning an error if something goes wrong.
type FileGenerator interface {
	AddModule(context.Context, context.TestcontainersModule) error
}

func GenerateFiles(ctx context.Context, tcModule context.TestcontainersModule) error {
	if err := tcModule.Validate(); err != nil {
		return fmt.Errorf("validate module %s: %w", tcModule.Name, err)
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
			return fmt.Errorf("add module %s: %w", tcModule.Name, err)
		}
	}

	// they are based on the content of the modules in the project workspace,
	// not in the new module to be added, that's why they happen after the actual
	// module generation
	projectGenerators := []ProjectGenerator{
		vscode.Generator{}, // update vscode workspace
	}

	examples, err := ctx.GetExamples()
	if err != nil {
		return fmt.Errorf("get examples: %w", err)
	}
	modules, err := ctx.GetModules()
	if err != nil {
		return fmt.Errorf("get modules: %w", err)
	}

	for _, generator := range projectGenerators {
		err := generator.Generate(ctx, examples, modules)
		if err != nil {
			return fmt.Errorf("generate project: %w", err)
		}
	}

	return nil
}
