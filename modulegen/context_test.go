package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

func getTestRootContext(t *testing.T) context.Context {
	current, err := os.Getwd()
	require.NoError(t, err)
	return context.New(filepath.Dir(current))
}
