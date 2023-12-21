package cassandra

import (
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type initScript struct {
	testcontainers.ExecOptions
	File string
}

func (i initScript) AsCommand() []string {
	if strings.HasSuffix(i.File, ".cql") {
		return []string{"cqlsh", "-f", i.File}
	} else if strings.HasSuffix(i.File, ".sh") {
		return []string{"/bin/sh", i.File}
	}
	return nil
}
