package context

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TestcontainersModule represents a technology module of the Testcontainers project.
type TestcontainersModule struct {
	// Image fully qualified name of the Docker image
	Image string

	// Name name of the module
	Name string

	// TitleName title of the name: m.g. "mongodb" -> "MongoDB"
	TitleName string

	// TCVersion Testcontainers for Go version
	TCVersion string
}

// ContainerName returns the name of the container type.
func (m *TestcontainersModule) ContainerName() string {
	return "Container"
}

// Entrypoint returns the name of the module entrypoint function.
func (m *TestcontainersModule) Entrypoint() string {
	return "Run"
}

// Lower returns the lower-cased name of the module
func (m *TestcontainersModule) Lower() string {
	return strings.ToLower(m.Name)
}

// ParentDir returns the parent directory of the module.
func (m *TestcontainersModule) ParentDir() string {
	return "modules"
}

// Title returns the title of the module. If it's not set,
// it will be the title of the lower-cased name.
func (m *TestcontainersModule) Title() string {
	if m.TitleName != "" {
		return m.TitleName
	}

	return cases.Title(language.Und, cases.NoLower).String(m.Lower())
}

// Type returns the type of the module.
func (m *TestcontainersModule) Type() string {
	return "module"
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
