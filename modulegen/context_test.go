package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

func getTestRootContext(t *testing.T) context.Context {
	t.Helper()
	current, err := os.Getwd()
	require.NoError(t, err)
	return context.New(filepath.Dir(current))
}

func copyModulesAndExamples(t *testing.T, ctx context.Context) {
	t.Helper()
	rootCtx := getTestRootContext(t)

	// copy examples and modules dir to the test context
	err := os.MkdirAll(filepath.Join(ctx.RootDir, "examples"), 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(ctx.RootDir, "modules"), 0o777)
	require.NoError(t, err)

	examples, err := rootCtx.GetExamples()
	require.NoError(t, err)

	for _, example := range examples {
		// create dirs for each example, as we do not need to create the rest of the files
		err = os.MkdirAll(filepath.Join(ctx.RootDir, "examples", example), 0o777)
		require.NoError(t, err)
	}

	modules, err := rootCtx.GetModules()
	require.NoError(t, err)

	for _, module := range modules {
		// create dirs for each module, as we do not need to create the rest of the files
		err = os.MkdirAll(filepath.Join(ctx.RootDir, "modules", module), 0o777)
		require.NoError(t, err)
	}
}
