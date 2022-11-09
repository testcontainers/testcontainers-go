package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

var nameVar string
var imageVar string

var templates = []string{
	"docs_example.md", "example_test.go", "example.go", "go.mod", "go.sum", "Makefile", "tools.go",
}

func init() {
	flag.StringVar(&nameVar, "name", "", "Name of the example, use camel-case when needed")
	flag.StringVar(&imageVar, "image", "", "Fully-qualified name of the Docker image to be used by the example")
}

type Example struct {
	Image string // fully qualified name of the Docker image
	Name  string
}

func (e *Example) Lower() string {
	return strings.ToLower(e.Name)
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

	examplesDir, err := filepath.Abs(filepath.Dir(nameVar))
	if err != nil {
		fmt.Printf(">> could not get the examples dir: %v\n", err)
		os.Exit(1)
	}

	examplesDocsPath := filepath.Join(filepath.Dir(examplesDir), "docs", "examples")

	err = generate(nameVar, imageVar, examplesDir, examplesDocsPath)
	if err != nil {
		fmt.Printf(">> error generating the example: %v\n", err)
		os.Exit(1)
	}
}

func generate(name string, image string, examplesDir string, docsDir string) error {
	example := Example{Name: name, Image: image}

	funcMap := template.FuncMap{
		"ToLower":     strings.ToLower,
		"codeinclude": func(s string) template.HTML { return template.HTML(s) }, // escape HTML comments for codeinclude
	}

	// create the example dir
	err := os.MkdirAll(examplesDir, 0700)
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		name := tmpl + ".tmpl"
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
		if err != nil {
			return err
		}

		// create a new file
		exampleFilePath := filepath.Join(examplesDir, example.Lower(), tmpl)
		exampleFilePath = strings.ReplaceAll(exampleFilePath, "example", example.Lower())

		if strings.EqualFold(tmpl, "docs_example.md") {
			// docs example file will go into the docs directory
			exampleFilePath = filepath.Join(docsDir, example.Lower()+".md")
		} else if strings.EqualFold(tmpl, "tools.go") {
			// tools.go example file will go into the tools package
			exampleFilePath = filepath.Join(examplesDir, example.Lower(), "tools", tmpl)
		}

		err = os.MkdirAll(filepath.Dir(exampleFilePath), 0777)
		if err != nil {
			return err
		}

		exampleFile, _ := os.Create(exampleFilePath)
		defer exampleFile.Close()

		err = t.ExecuteTemplate(exampleFile, name, example)
		if err != nil {
			return err
		}
	}

	fmt.Println("Please go to", example.Lower(), "directory and execute 'go mod tidy' to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
	return nil
}
