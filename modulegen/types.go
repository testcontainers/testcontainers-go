package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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
