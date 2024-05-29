package testcontainers

import (
	"github.com/testcontainers/testcontainers-go/internal/core"
)

// GenericLabels returns a map of labels that can be used to identify containers created by this library
func GenericLabels() map[string]string {
	return core.DefaultLabels(core.SessionID())
}
