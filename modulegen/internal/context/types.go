package context

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TestcontainersModule struct {
	Image     string // fully qualified name of the Docker image
	IsModule  bool   // if true, the module will be generated as a Go module, otherwise an example
	Name      string
	TitleName string // title of the name: m.g. "mongodb" -> "MongoDB"
	TCVersion string // Testcontainers for Go version
}

// ContainerName returns the name of the container, which is the lower-cased title of the example
// If the title is set, it will be used instead of the name
func (m *TestcontainersModule) ContainerName() string {
	name := m.Lower()

	if m.IsModule {
		name = m.Title()
	} else {
		if m.TitleName != "" {
			r, n := utf8.DecodeRuneInString(m.TitleName)
			name = string(unicode.ToLower(r)) + m.TitleName[n:]
		}
	}

	return name + "Container"
}

// Entrypoint returns the name of the entrypoint function, which is the lower-cased title of the example
// If the example is a module, the entrypoint will be "RunContainer"
func (m *TestcontainersModule) Entrypoint() string {
	if m.IsModule {
		return "RunContainer"
	}

	return "runContainer"
}

func (m *TestcontainersModule) Lower() string {
	return strings.ToLower(m.Name)
}

func (m *TestcontainersModule) ParentDir() string {
	if m.IsModule {
		return "modules"
	}

	return "examples"
}

func (m *TestcontainersModule) Title() string {
	if m.TitleName != "" {
		return m.TitleName
	}

	return cases.Title(language.Und, cases.NoLower).String(m.Lower())
}

func (m *TestcontainersModule) Type() string {
	if m.IsModule {
		return "module"
	}
	return "example"
}

func (m *TestcontainersModule) Validate() error {
	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(m.Name) {
		return fmt.Errorf("invalid name: %s. Only alphanumerical characters are allowed (leading character must be a letter)", m.Name)
	}

	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(m.TitleName) {
		return fmt.Errorf("invalid title: %s. Only alphanumerical characters are allowed (leading character must be a letter)", m.TitleName)
	}

	return nil
}
