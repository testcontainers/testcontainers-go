package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/modfile"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/workfile"
)

func copyModFile(t *testing.T, ctx context.Context) {
	t.Helper()

	rootCtx := getTestRootContext(t)
	f, err := modfile.Read(filepath.Join(rootCtx.RootDir, "go.mod"))
	require.NoError(t, err)

	err = modfile.Write(filepath.Join(ctx.RootDir, "go.mod"), f)
	require.NoError(t, err)
}

func copyWorkFile(t *testing.T, ctx context.Context) {
	t.Helper()

	rootCtx := getTestRootContext(t)

	f, err := workfile.Read(filepath.Join(rootCtx.RootDir, "go.work"))
	require.NoError(t, err)

	err = workfile.Write(filepath.Join(ctx.RootDir, "go.work"), f)
	require.NoError(t, err)
}
