package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
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
	flag.StringVar(&nameVar, "name", "", "Name of the example, use camel-case when needed. Only alphabetical characters are allowed.")
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

	mkdocsConfig, err := readMkdocsConfig(rootDir)
	if err != nil {
		fmt.Printf(">> could not read MkDocs config: %v\n", err)
		os.Exit(1)
	}

	err = generate(Example{Name: nameVar, Image: imageVar, TCVersion: mkdocsConfig.Extra.LatestVersion}, rootDir)
	if err != nil {
		fmt.Printf(">> error generating the example: %v\n", err)
		os.Exit(1)
	}
}

func generate(example Example, rootDir string) error {
	if !regexp.MustCompile(`^[A-Za-z]+$`).MatchString(example.Name) {
		return fmt.Errorf("invalid name: %s. Only alphabetical characters are allowed", example.Name)
	}

	githubWorkflowsDir := filepath.Join(rootDir, ".github", "workflows")
	examplesDir := filepath.Join(rootDir, "examples")
	docsDir := filepath.Join(rootDir, "docs", "examples")

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

	// update examples in mkdocs
	err = generateMkdocs(rootDir, exampleLower)
	if err != nil {
		return err
	}

	// update examples in dependabot
	err = generateDependabotUpdates(rootDir, exampleLower)
	if err != nil {
		return err
	}

	fmt.Println("Please go to", example.Lower(), "directory and execute 'go mod tidy' to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
	return nil
}

func generateDependabotUpdates(rootDir string, exampleLower string) error {
	// update examples in dependabot
	dependabotConfig, err := readDependabotConfig(rootDir)
	if err != nil {
		return err
	}

	dependabotExampleUpdates := dependabotConfig.Updates

	// make sure the main module is the first element in the list of examples
	// and the e2e module is the second element
	exampleUpdates := make(Updates, len(dependabotExampleUpdates)-2)
	j := 0

	for _, exampleUpdate := range dependabotExampleUpdates {
		// filter out the index.md file
		if exampleUpdate.Directory != "/" && exampleUpdate.Directory != "/e2e" {
			exampleUpdates[j] = exampleUpdate
			j++
		}
	}

	exampleUpdates = append(exampleUpdates, NewUpdate(exampleLower))
	sort.Sort(exampleUpdates)

	// prepend the main and e2e modules
	exampleUpdates = append([]Update{dependabotExampleUpdates[0], dependabotExampleUpdates[1]}, exampleUpdates...)

	dependabotConfig.Updates = exampleUpdates

	return writeDependabotConfig(rootDir, dependabotConfig)
}

func generateMkdocs(rootDir string, exampleLower string) error {
	// update examples in mkdocs
	mkdocsConfig, err := readMkdocsConfig(rootDir)
	if err != nil {
		return err
	}

	mkdocsExamplesNav := mkdocsConfig.Nav[3].Examples

	// make sure the index.md is the first element in the list of examples in the nav
	examplesNav := make([]string, len(mkdocsExamplesNav)-1)
	j := 0

	for _, exampleNav := range mkdocsExamplesNav {
		// filter out the index.md file
		if !strings.HasSuffix(exampleNav, "index.md") {
			examplesNav[j] = exampleNav
			j++
		}
	}

	examplesNav = append(examplesNav, "examples/"+exampleLower+".md")
	sort.Strings(examplesNav)

	// prepend the index.md file
	examplesNav = append([]string{"examples/index.md"}, examplesNav...)

	mkdocsConfig.Nav[3].Examples = examplesNav

	return writeMkdocsConfig(rootDir, mkdocsConfig)
}
