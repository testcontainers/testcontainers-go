package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/tools"
)

var (
	asModuleVar        bool
	nameVar            string
	nameTitleVar       string
	imageVar           string
	vsCodeWorkspaceVar bool
)

var templates = []string{"docs_example.md", "example_test.go", "example.go", "go.mod"}

func init() {
	flag.StringVar(&nameVar, "name", "", "Name of the example. Only alphabetical characters are allowed.")
	flag.StringVar(&nameTitleVar, "title", "", "(Optional) Title of the example name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	flag.StringVar(&imageVar, "image", "", "Fully-qualified name of the Docker image to be used by the example")
	flag.BoolVar(&asModuleVar, "as-module", false, "If set, the example will be generated as a Go module, under the modules directory. Otherwise, it will be generated as a subdirectory of the examples directory.")
	flag.BoolVar(&vsCodeWorkspaceVar, "vscode-workspace", false, "If set, the string representation of the VSCode workspace file will be printed out to the stdout.")
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

	currentDir, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		fmt.Printf(">> could not get the root dir: %v\n", err)
		os.Exit(1)
	}

	ctx := NewContext(filepath.Dir(currentDir))

	mkdocsConfig, err := mkdocs.ReadConfig(ctx.MkdocsConfigFile())
	if err != nil {
		fmt.Printf(">> could not read MkDocs config: %v\n", err)
		os.Exit(1)
	}

	example := Example{
		Image:     imageVar,
		IsModule:  asModuleVar,
		Name:      nameVar,
		TitleName: nameTitleVar,
		TCVersion: mkdocsConfig.Extra.LatestVersion,
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

	if vsCodeWorkspaceVar {
		err = generateVSCodeWorkspace(ctx)
		if err != nil {
			os.Exit(1)
		}
	}

	fmt.Println("Please go to", cmdDir, "directory to check the results, where 'go mod tidy' and 'go vet' was executed to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
}

func generate(example Example, ctx *Context) error {
	if err := example.Validate(); err != nil {
		return err
	}

	outputDir := filepath.Join(ctx.RootDir, example.ParentDir())
	docsOuputDir := filepath.Join(ctx.DocsDir(), example.ParentDir())

	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return example.Entrypoint() },
		"ContainerName": func() string { return example.ContainerName() },
		"ExampleType":   func() string { return example.Type() },
		"ParentDir":     func() string { return example.ParentDir() },
		"ToLower":       func() string { return example.Lower() },
		"Title":         func() string { return example.Title() },
		"codeinclude":   func(s string) template.HTML { return template.HTML(s) }, // escape HTML comments for codeinclude
	}

	exampleLower := example.Lower()

	// create the example dir
	err := os.MkdirAll(filepath.Join(outputDir, exampleLower), 0o700)
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		name := tmpl + ".tmpl"
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
		if err != nil {
			return err
		}

		// initialize the data using the example struct, which is the default data to be used while
		// doing the interpolation of the data and the template
		var data any

		syncDataFn := func() any {
			return example
		}

		// create a new file
		var exampleFilePath string

		if strings.EqualFold(tmpl, "docs_example.md") {
			// docs example file will go into the docs directory
			exampleFilePath = filepath.Join(docsOuputDir, exampleLower+".md")
		} else {
			exampleFilePath = filepath.Join(outputDir, exampleLower, strings.ReplaceAll(tmpl, "example", exampleLower))
		}

		err = os.MkdirAll(filepath.Dir(exampleFilePath), 0o777)
		if err != nil {
			return err
		}

		exampleFile, _ := os.Create(exampleFilePath)
		defer exampleFile.Close()

		data = syncDataFn()

		err = t.ExecuteTemplate(exampleFile, name, data)
		if err != nil {
			return err
		}
	}
	// creates Makefile for example
	err = generateMakefile(ctx, example)
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
	return nil
}

func getRootContext() (*Context, error) {
	current, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return NewContext(filepath.Dir(current)), nil
}
