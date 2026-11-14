package testcontainers

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// The pre-scan walk records src/config.go. After it returns, the hook replaces
// src with a symlink to a directory outside the context. A regular-file payload
// that carries the outside marker is a leak; an error or an omitted entry is not.
func TestGetContext_dockerignoreSymlinkSwapAfterPrescan(t *testing.T) {
	const marker = "OUTSIDE-MARKER-9d1e7a"
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "config.go"), []byte("package x // "+marker+"\n"), 0o600))
	src := filepath.Join(root, "src")
	require.NoError(t, os.MkdirAll(src, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "config.go"), []byte("package src\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".dockerignore"), []byte("!nothing\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "Dockerfile"), []byte("FROM scratch\n"), 0o600))

	orig := walkDir
	t.Cleanup(func() { walkDir = orig })
	fired := false
	var setupErr error
	walkDir = func(r string, fn fs.WalkDirFunc) error {
		if err := orig(r, fn); err != nil {
			return err
		}
		if setupErr = os.Rename(src, filepath.Join(root, "src-backup")); setupErr != nil {
			return setupErr
		}
		if setupErr = os.Symlink(outside, src); setupErr != nil {
			return setupErr
		}
		fired = true
		return nil
	}

	r, err := (&ContainerRequest{FromDockerfile: FromDockerfile{Context: root}}).GetContext()
	require.NoError(t, setupErr, "symlink swap setup failed; the test is not exercising the race")
	require.True(t, fired, "swap hook never ran; the test is not exercising the race")
	if err != nil {
		t.Logf("outcome: error: %v", err)
		return
	}
	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}
	tr := tar.NewReader(r)
	leaked := false
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		body, err := io.ReadAll(tr)
		require.NoError(t, err)
		t.Logf("entry %s type=%c", hdr.Name, hdr.Typeflag)
		if hdr.Typeflag == tar.TypeReg && bytes.Contains(body, []byte(marker)) {
			leaked = true
			t.Logf("leak: regular entry %s carries the outside marker", hdr.Name)
		}
	}
	require.False(t, leaked, "outside file leaked into the context")
}

// The pre-scan walk records src/config.go under the excluded src, and the
// re-include keeps it. After the walk returns, the hook replaces src with a
// symlink to a directory outside the context. The listed path must not reach
// the outside file; an error or an omitted entry is not a leak.
func TestGetContext_dockerignoreExcludedDirSwapAfterPrescan(t *testing.T) {
	const marker = "OUTSIDE-MARKER-4b2f09"
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "config.go"), []byte("package x // "+marker+"\n"), 0o600))
	src := filepath.Join(root, "src")
	require.NoError(t, os.MkdirAll(src, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "config.go"), []byte("package src\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".dockerignore"), []byte("src\n!src/**/config.go\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "Dockerfile"), []byte("FROM scratch\n"), 0o600))

	orig := walkDir
	t.Cleanup(func() { walkDir = orig })
	fired := false
	var setupErr error
	walkDir = func(r string, fn fs.WalkDirFunc) error {
		if err := orig(r, fn); err != nil {
			return err
		}
		if setupErr = os.Rename(src, filepath.Join(root, "src-backup")); setupErr != nil {
			return setupErr
		}
		if setupErr = os.Symlink(outside, src); setupErr != nil {
			return setupErr
		}
		fired = true
		return nil
	}

	r, err := (&ContainerRequest{FromDockerfile: FromDockerfile{Context: root}}).GetContext()
	require.NoError(t, setupErr, "symlink swap setup failed; the test is not exercising the race")
	require.True(t, fired, "swap hook never ran; the test is not exercising the race")
	if err != nil {
		t.Logf("outcome: error: %v", err)
		return
	}
	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}
	tr := tar.NewReader(r)
	leaked := false
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		body, err := io.ReadAll(tr)
		require.NoError(t, err)
		t.Logf("entry %s type=%c", hdr.Name, hdr.Typeflag)
		if bytes.Contains(body, []byte(marker)) {
			leaked = true
			t.Logf("leak: entry %s carries the outside marker", hdr.Name)
		}
	}
	require.False(t, leaked, "outside file leaked into the context through an excluded directory")
}
