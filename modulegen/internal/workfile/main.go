package workfile

import (
	"fmt"
	"path/filepath"

	"golang.org/x/mod/modfile"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
)

type Generator struct{}

// Generate updates github ci workflow
func (g Generator) Generate(ctx context.Context) error {
	examples, modules, err := module.ListExamplesAndModules(ctx)
	if err != nil {
		return err
	}

	rootDir := ctx.RootDir

	rootGoWorkFilePath := filepath.Join(rootDir, "go.work")

	rootGoWork, err := Read(rootGoWorkFilePath)
	if err != nil {
		return fmt.Errorf("read go.mod file: %w", err)
	}

	err = newWorkFile(rootGoWork, examples, modules)
	if err != nil {
		return fmt.Errorf("create go.mod file: %w", err)
	}

	return Write(rootGoWorkFilePath, rootGoWork)
}

func newWorkFile(goWork *modfile.WorkFile, examples []string, modules []string) error {
	defer goWork.Cleanup()

	if err := goWork.AddUse(".", ""); err != nil {
		return fmt.Errorf("add use .: %w", err)
	}

	for _, example := range examples {
		if err := goWork.AddUse("./examples/"+example, ""); err != nil {
			return fmt.Errorf("add use examples/%s: %w", example, err)
		}
	}

	if err := goWork.AddUse("./modulegen", ""); err != nil {
		return fmt.Errorf("add use modulegen: %w", err)
	}

	for _, module := range modules {
		if err := goWork.AddUse("./modules/"+module, ""); err != nil {
			return fmt.Errorf("add use modules/%s: %w", module, err)
		}
	}

	goWork.SortBlocks()

	return nil
}