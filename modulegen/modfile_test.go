package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/modfile"
)

func copyModFile(t *testing.T, ctx context.Context) {
	t.Helper()

	rootCtx := getTestRootContext(t)
	f, err := modfile.Read(filepath.Join(rootCtx.RootDir, "go.mod"))
	require.NoError(t, err)

	err = modfile.Write(filepath.Join(ctx.RootDir, "go.mod"), f)
	require.NoError(t, err)
}
