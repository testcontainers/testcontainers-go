package vscode

import "path/filepath"

type ProjectDirectories struct {
	RootDir            string
	ModuleGeneratorDir string
	Examples           []string
	Modules            []string
}

func newProjectDirectories(rootDir string, examples []string, modules []string) *ProjectDirectories {
	rootDirAbs, err := filepath.Abs(rootDir)
	if err != nil {
		rootDirAbs = rootDir
	}

	moduleGeneratorDirAbs := filepath.Join(rootDirAbs, "modulegen")

	for i, example := range examples {
		examples[i] = filepath.Join(rootDirAbs, "examples", example)
	}

	for i, module := range modules {
		modules[i] = filepath.Join(rootDirAbs, "modules", module)
	}

	return &ProjectDirectories{
		RootDir:            rootDirAbs,
		ModuleGeneratorDir: moduleGeneratorDirAbs,
		Examples:           examples,
		Modules:            modules,
	}
}
