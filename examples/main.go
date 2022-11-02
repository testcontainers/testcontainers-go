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

func init() {
	flag.StringVar(&nameVar, "name", "", "Name of the example, use camel-case when needed")
}

type Example struct {
	Name string
}

func (e *Example) Lower() string {
	return strings.ToLower(e.Name)
}

func main() {
	required := []string{"name"}
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

	example := Example{Name: nameVar}

	funcMap := template.FuncMap{
		"ToLower":     strings.ToLower,
		"codeinclude": func(s string) template.HTML { return template.HTML(s) }, // escape HTML comments for codeinclude
	}

	tmpls := []string{
		"docs_example.md", "example_test.go", "example.go", "go.mod", "go.sum", "Makefile", "tools.go",
	}

	// create the example dir
	err := os.Mkdir(example.Lower(), 0700)
	if err != nil {
		panic(err)
	}

	for _, tmpl := range tmpls {
		name := tmpl + ".tmpl"
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
		if err != nil {
			panic(err)
		}

		// create a new file
		exampleFilePath := filepath.Join(example.Lower(), tmpl)
		exampleFilePath = strings.ReplaceAll(exampleFilePath, "example", example.Lower())

		// docs example file will go into the docs directory
		if strings.EqualFold(tmpl, "docs_example.md") {
			abs, err := filepath.Abs(exampleFilePath)
			if err != nil {
				panic(err)
			}

			examplesDocsPath := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(abs))), "docs", "examples")

			exampleFilePath = filepath.Join(examplesDocsPath, example.Lower()+".md")
		}

		err = os.MkdirAll(filepath.Dir(exampleFilePath), 0777)
		if err != nil {
			panic(err)
		}

		exampleFile, _ := os.Create(exampleFilePath)
		defer exampleFile.Close()

		err = t.ExecuteTemplate(exampleFile, name, example)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Please go to", example.Lower(), "directory and execute 'go mod tidy' to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
}
