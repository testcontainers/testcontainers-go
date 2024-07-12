package main

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

func getTestRootContext(t *testing.T) context.Context {
	current, err := os.Getwd()
	assert.NoError(t, err)
	return context.New(filepath.Dir(current))
}
