package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
)

func getTestRootContext(t *testing.T) context.Context {
	current, err := os.Getwd()
	require.NoError(t, err)

	// we are in the cmd/devtools directory, so we need to go up
	// two directories to get to the root of the project
	return context.New(filepath.Dir(filepath.Dir(current)))
}
