package context

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TestcontainersModule represents a module or an example of the Testcontainers project.
// It defines the common properties for both modules and examples,
// such as the image name, if it's a module or an example, the name and the title of the module.
// It provides methods to be used during the code generation of the module or example.
type TestcontainersModule struct {
	// Image fully qualified name of the Docker image
	Image string

	// IsModule if true, the module will be generated as a Go module, otherwise an example
	IsModule bool

	// Name name of the module
	Name string

	// TitleName title of the name: m.g. "mongodb" -> "MongoDB"
	TitleName string

	// TCVersion Testcontainers for Go version
	TCVersion string
}

// ContainerName returns the name of the container, which is the lower-cased title of the example
// If the title is set, it will be used instead of the name
func (m *TestcontainersModule) ContainerName() string {
	return "Container"
}

// Entrypoint returns the name of the entrypoint function, which is the lower-cased title of the example
// If the example is a module, the entrypoint will be "Run"
func (m *TestcontainersModule) Entrypoint() string {
	if m.IsModule {
		return "Run"
	}

	return "run"
}

// Lower returns the lower-cased name of the module
func (m *TestcontainersModule) Lower() string {
	return strings.ToLower(m.Name)
}

// ParentDir returns the parent directory of the module: "modules" or "examples"
func (m *TestcontainersModule) ParentDir() string {
	if m.IsModule {
		return "modules"
	}

	return "examples"
}

// Title returns the title of the module. If it's not set,
// it will be the title of the lower-cased name.
func (m *TestcontainersModule) Title() string {
	if m.TitleName != "" {
		return m.TitleName
	}

	return cases.Title(language.Und, cases.NoLower).String(m.Lower())
}

// Type returns the type of the module: "module" or "example"
func (m *TestcontainersModule) Type() string {
	if m.IsModule {
		return "module"
	}
	return "example"
}

// Validate validates the module name and title
func (m *TestcontainersModule) Validate() error {
	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(m.Name) {
		return fmt.Errorf("invalid name: %s. Only alphanumerical characters are allowed (leading character must be a letter)", m.Name)
	}

	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(m.TitleName) {
		return fmt.Errorf("invalid title: %s. Only alphanumerical characters are allowed (leading character must be a letter)", m.TitleName)
	}

	return nil
}
