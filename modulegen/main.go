package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var asModuleVar bool
var nameVar string
var nameTitleVar string
var imageVar string

var templates = []string{
	"ci.yml", "docs_example.md", "example_test.go", "example.go", "go.mod", "Makefile", "tools.go",
}

func init() {
	flag.StringVar(&nameVar, "name", "", "Name of the example. Only alphabetical characters are allowed.")
	flag.StringVar(&nameTitleVar, "title", "", "(Optional) Title of the example name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	flag.StringVar(&imageVar, "image", "", "Fully-qualified name of the Docker image to be used by the example")
	flag.BoolVar(&asModuleVar, "as-module", false, "If set, the example will be generated as a Go module, under the modules directory. Otherwise, it will be generated as a subdirectory of the examples directory.")
}

type Example struct {
	Image     string // fully qualified name of the Docker image
	IsModule  bool   // if true, the example will be generated as a Go module
	Name      string
	TitleName string // title of the name: e.g. "mongodb" -> "MongoDB"
	TCVersion string // Testcontainers for Go version
}

// ContainerName returns the name of the container, which is the lower-cased title of the example
// If the title is set, it will be used instead of the name
func (e *Example) ContainerName() string {
	name := e.Lower()

	if e.IsModule {
		name = e.Title()
	} else {
		if e.TitleName != "" {
			r, n := utf8.DecodeRuneInString(e.TitleName)
			name = string(unicode.ToLower(r)) + e.TitleName[n:]
		}
	}

	return name + "Container"
}

// Entrypoint returns the name of the entrypoint function, which is the lower-cased title of the example
// If the example is a module, the entrypoint will be "RunContainer"
func (e *Example) Entrypoint() string {
	if e.IsModule {
		return "RunContainer"
	}

	return "runContainer"
}

func (e *Example) Lower() string {
	return strings.ToLower(e.Name)
}

func (e *Example) ParentDir() string {
	if e.IsModule {
		return "modules"
	}

	return "examples"
}

func (e *Example) Title() string {
	if e.TitleName != "" {
		return e.TitleName
	}

	return cases.Title(language.Und, cases.NoLower).String(e.Lower())
}

func (e *Example) Type() string {
	if e.IsModule {
		return "module"
	}
	return "example"
}

func (e *Example) Validate() error {
	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(e.Name) {
		return fmt.Errorf("invalid name: %s. Only alphanumerical characters are allowed (leading character must be a letter)", e.Name)
	}

	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(e.TitleName) {
		return fmt.Errorf("invalid title: %s. Only alphanumerical characters are allowed (leading character must be a letter)", e.TitleName)
	}

	return nil
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

	rootDir := filepath.Dir(currentDir)

	mkdocsConfig, err := readMkdocsConfig(rootDir)
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

	err = generate(example, rootDir)
	if err != nil {
		fmt.Printf(">> error generating the example: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = filepath.Join(rootDir, example.ParentDir(), example.Lower())
	err = cmd.Run()
	if err != nil {
		fmt.Printf(">> error synchronizing the dependencies: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Please go to", cmd.Dir, "directory to check the results, where 'go mod tidy' was executed to synchronize the dependencies")
	fmt.Println("Commit the modified files and submit a pull request to include them into the project")
	fmt.Println("Thanks!")
}

func generate(example Example, rootDir string) error {
	if err := example.Validate(); err != nil {
		return err
	}

	githubWorkflowsDir := filepath.Join(rootDir, ".github", "workflows")
	outputDir := filepath.Join(rootDir, example.ParentDir())
	docsOuputDir := filepath.Join(rootDir, "docs", example.ParentDir())

	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return example.Entrypoint() },
		"ContainerName": func() string { return example.ContainerName() },
		"ExampleType":   func() string { return example.Type() },
		"ParentDir":     func() string { return example.ParentDir() },
		"ToLower":       func() string { return example.Lower() },
		"Title":         func() string { return example.Title() },
		"codeinclude":   func(s string) template.HTML { return template.HTML(s) }, // escape HTML comments for codeinclude
	}

	// create the example dir
	err := os.MkdirAll(outputDir, 0700)
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
			exampleFilePath = filepath.Join(docsOuputDir, exampleLower+".md")
		} else if strings.EqualFold(tmpl, "ci.yml") {
			// GitHub workflow example file will go into the .github/workflows directory
			fileName := exampleLower + "-example.yml"
			if example.IsModule {
				fileName = "module-" + exampleLower + ".yml"
			}

			exampleFilePath = filepath.Join(githubWorkflowsDir, fileName)
		} else if strings.EqualFold(tmpl, "tools.go") {
			// tools.go example file will go into the tools package
			exampleFilePath = filepath.Join(outputDir, exampleLower, "tools", tmpl)
		} else {
			exampleFilePath = filepath.Join(outputDir, exampleLower, strings.ReplaceAll(tmpl, "example", exampleLower))
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
	err = generateMkdocs(rootDir, example)
	if err != nil {
		return err
	}

	// update examples in dependabot
	err = generateDependabotUpdates(rootDir, example)
	if err != nil {
		return err
	}

	return nil
}

func generateDependabotUpdates(rootDir string, example Example) error {
	// update examples in dependabot
	dependabotConfig, err := readDependabotConfig(rootDir)
	if err != nil {
		return err
	}

	dependabotExampleUpdates := dependabotConfig.Updates

	// make sure the main module is the first element in the list of examples
	exampleUpdates := make(Updates, len(dependabotExampleUpdates)-1)
	j := 0

	for _, exampleUpdate := range dependabotExampleUpdates {
		// filter out the root module
		if exampleUpdate.Directory != "/" {
			exampleUpdates[j] = exampleUpdate
			j++
		}
	}

	exampleUpdates = append(exampleUpdates, NewUpdate(example))
	sort.Sort(exampleUpdates)

	// prepend the main and compose modules
	exampleUpdates = append([]Update{dependabotExampleUpdates[0]}, exampleUpdates...)

	dependabotConfig.Updates = exampleUpdates

	return writeDependabotConfig(rootDir, dependabotConfig)
}

func generateMkdocs(rootDir string, example Example) error {
	// update examples in mkdocs
	mkdocsConfig, err := readMkdocsConfig(rootDir)
	if err != nil {
		return err
	}

	mkdocsExamplesNav := mkdocsConfig.Nav[4].Examples
	if example.IsModule {
		mkdocsExamplesNav = mkdocsConfig.Nav[3].Modules
	}

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

	examplesNav = append(examplesNav, example.ParentDir()+"/"+example.Lower()+".md")
	sort.Strings(examplesNav)

	// prepend the index.md file
	examplesNav = append([]string{example.ParentDir() + "/index.md"}, examplesNav...)

	if example.IsModule {
		mkdocsConfig.Nav[3].Modules = examplesNav
	} else {
		mkdocsConfig.Nav[4].Examples = examplesNav
	}

	return writeMkdocsConfig(rootDir, mkdocsConfig)
}
