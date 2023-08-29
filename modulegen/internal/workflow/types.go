package workflow

import (
	"strings"
)

type ProjectDirectories struct {
	Examples string
	Modules  string
}

func newProjectDirectories(examples []string, modules []string) *ProjectDirectories {
	return &ProjectDirectories{
		Examples: strings.Join(examples, ", "),
		Modules:  strings.Join(modules, ", "),
	}
}
