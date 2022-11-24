package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var nameVar string
var imageVar string

var templates = []string{
	"ci.yml", "docs_example.md", "example_test.go", "example.go", "go.mod", "go.sum", "Makefile", "tools.go",
}

func init() {
	flag.StringVar(&nameVar, "name", "", "Name of the example, use camel-case when needed")
	flag.StringVar(&imageVar, "image", "", "Fully-qualified name of the Docker image to be used by the example")
}

type Example struct {
	Image     string // fully qualified name of the Docker image
	Name      string
	TCVersion string // Testcontainers for Go version
}

func (e *Example) Lower() string {
	return strings.ToLower(e.Name)
}

func (e *Example) Title() string {
	return cases.Title(language.Und, cases.NoLower).String(e.Lower())
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

	rootDir := filepath.Dir(examplesDir)
	githubWorkflowsPath := filepath.Join(rootDir, ".github", "workflows")
	examplesDocsPath := filepath.Join(rootDir, "docs", "examples")

	mkdocsConfig, err := readMkdocsConfig(rootDir)
	if err != nil {
		fmt.Printf(">> could not read MkDocs config: %v\n", err)
		os.Exit(1)
	}

	err = generate(Example{Name: nameVar, Image: imageVar, TCVersion: mkdocsConfig.Extra.LatestVersion}, examplesDir, examplesDocsPath, githubWorkflowsPath)
	if err != nil {
		fmt.Printf(">> error generating the example: %v\n", err)
		os.Exit(1)
	}
}

func generate(example Example, examplesDir string, docsDir string, githubWorkflowsDir string) error {
	funcMap := template.FuncMap{
		"ToLower":     strings.ToLower,
		"Title":       cases.Title(language.Und, cases.NoLower).String,
		"codeinclude": func(s string) template.HTML { return template.HTML(s) }, // escape HTML comments for codeinclude
	}

	// create the example dir
	err := os.MkdirAll(examplesDir, 0700)
	if err != nil {
		return err
	}

	exampleLower := example.Lower()

	for _, tmpl := range templates {
		name := tmpl + ".tmpl"
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
		if err != nil {
			return err
		}

		// create a new file
		var exampleFilePath string

		if strings.EqualFold(tmpl, "docs_example.md") {
			// docs example file will go into the docs directory
			exampleFilePath = filepath.Join(docsDir, exampleLower+".md")
		} else if strings.EqualFold(tmpl, "ci.yml") {
			// GitHub workflow example file will go into the .github/workflows directory
			exampleFilePath = filepath.Join(githubWorkflowsDir, exampleLower+"-example.yml")
		} else if strings.EqualFold(tmpl, "tools.go") {
			// tools.go example file will go into the tools package
			exampleFilePath = filepath.Join(examplesDir, exampleLower, "tools", tmpl)
		} else {
			exampleFilePath = filepath.Join(examplesDir, exampleLower, strings.ReplaceAll(tmpl, "example", exampleLower))
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

	rootDir := filepath.Dir(examplesDir)

	// update examples in mkdocs
	mkdocsConfig, err := readMkdocsConfig(rootDir)
	if err != nil {
		return err
	}

	mkdocsExamplesNav := mkdocsConfig.Nav[3].Examples

	mkdocsExamplesNav = append(mkdocsExamplesNav, "examples/"+exampleLower+".md")
	sort.Strings(mkdocsExamplesNav)
	mkdocsConfig.Nav[3].Examples = mkdocsExamplesNav

	err = writeMkdocsConfig(rootDir, mkdocsConfig)
	if err != nil {
		return err
	}

	fmt.Println("Please go to", example.Lower(), "directory and execute 'go mod tidy' to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
	return nil
}
