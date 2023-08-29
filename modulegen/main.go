package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/tools"
)

var (
	asModuleVar  bool
	nameVar      string
	nameTitleVar string
	imageVar     string
)

func init() {
	flag.StringVar(&nameVar, "name", "", "Name of the example. Only alphabetical characters are allowed.")
	flag.StringVar(&nameTitleVar, "title", "", "(Optional) Title of the example name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	flag.StringVar(&imageVar, "image", "", "Fully-qualified name of the Docker image to be used by the example")
	flag.BoolVar(&asModuleVar, "as-module", false, "If set, the example will be generated as a Go module, under the modules directory. Otherwise, it will be generated as a subdirectory of the examples directory.")
}

func main() {
	required := []string{"name", "image"}
	flag.Parse()

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			// or possibly use `log.Fatalf` instead of:
			fmt.Fprintf(os.Stderr, "missing required -%s argument/flag\n", req)
			os.Exit(2) // the same exit code flag.Parse uses
		}
	}

	ctx, err := getRootContext()
	if err != nil {
		fmt.Printf(">> could not get the root dir: %v\n", err)
		os.Exit(1)
	}

	example := Example{
		Image:     imageVar,
		IsModule:  asModuleVar,
		Name:      nameVar,
		TitleName: nameTitleVar,
	}

	err = generate(example, ctx)
	if err != nil {
		fmt.Printf(">> error generating the example: %v\n", err)
		os.Exit(1)
	}

	cmdDir := filepath.Join(ctx.RootDir, example.ParentDir(), example.Lower())
	err = tools.GoModTidy(cmdDir)
	if err != nil {
		fmt.Printf(">> error synchronizing the dependencies: %v\n", err)
		os.Exit(1)
	}
	err = tools.GoVet(cmdDir)
	if err != nil {
		fmt.Printf(">> error checking generated code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Please go to", cmdDir, "directory to check the results, where 'go mod tidy' and 'go vet' was executed to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
}

func generate(example Example, ctx *Context) error {
	if err := example.Validate(); err != nil {
		return err
	}
	// creates Makefile for example
	err := generateMakefile(ctx, example)
	if err != nil {
		return err
	}

	err = generateGoModule(ctx, example)
	if err != nil {
		return err
	}

	// update github ci workflow
	err = generateWorkFlow(ctx)
	if err != nil {
		return err
	}
	// update examples in mkdocs
	err = generateMkdocs(ctx, example)
	if err != nil {
		return err
	}
	// update examples in dependabot
	err = generateDependabotUpdates(ctx, example)
	if err != nil {
		return err
	}
	// generate vscode workspace
	err = generateVSCodeWorkspace(ctx)
	if err != nil {
		return err
	}
	return nil
}

func getRootContext() (*Context, error) {
	current, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return NewContext(filepath.Dir(current)), nil
}
