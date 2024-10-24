package module

import (
	"sort"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

func ListExamplesAndModules(ctx context.Context) ([]string, []string, error) {
	var examples []string
	var modules []string

	rootCtx, err := context.GetRootContext()
	if err != nil {
		return examples, modules, err
	}

	examples, err = rootCtx.GetExamples()
	if err != nil {
		return examples, modules, err
	}
	modules, err = rootCtx.GetModules()
	if err != nil {
		return examples, modules, err
	}

	newExamples, err := ctx.GetExamples()
	if err != nil {
		return examples, modules, err
	}
	newModules, err := ctx.GetModules()
	if err != nil {
		return examples, modules, err
	}

	// merge the new examples and modules with the existing ones, if they are not already there
	for _, example := range newExamples {
		if !contains(examples, example) {
			examples = append(examples, example)
		}
	}
	for _, module := range newModules {
		if !contains(modules, module) {
			modules = append(modules, module)
		}
	}

	sort.Strings(examples)
	sort.Strings(modules)

	return examples, modules, nil
}

func contains(slice []string, element string) bool {
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}
