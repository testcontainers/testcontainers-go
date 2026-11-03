// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"archive/tar"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDockerIgnore(t *testing.T) {
	testCases := []struct {
		filePath         string
		exists           bool
		expectedErr      error
		expectedExcluded []string
	}{
		{
			filePath:         "./testdata/dockerignore",
			expectedErr:      nil,
			exists:           true,
			expectedExcluded: []string{"vendor", "foo", "bar"},
		},
		{
			filePath:         "./testdata",
			expectedErr:      nil,
			exists:           true,
			expectedExcluded: []string{"Dockerfile", "echo.Dockerfile"},
		},
		{
			filePath:         "./testdata/data",
			expectedErr:      nil,
			expectedExcluded: nil, // it's nil because the parseDockerIgnore function uses the zero value of a slice
		},
	}

	for _, testCase := range testCases {
		exists, excluded, err := parseDockerIgnore(testCase.filePath)
		assert.Equal(t, testCase.exists, exists)
		require.ErrorIs(t, testCase.expectedErr, err)
		assert.Equal(t, testCase.expectedExcluded, excluded)
	}
}

func tarNames(t *testing.T, dir string, files map[string]string, dirs ...string) []string {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0o755))
	}

	r, err := (&ContainerRequest{FromDockerfile: FromDockerfile{Context: dir}}).GetContext()
	require.NoError(t, err)

	var got []string
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		got = append(got, hdr.Name)
	}
	return got
}

func TestGetContext_dockerignoreReincludesNestedFiles(t *testing.T) {
	got := tarNames(t, t.TempDir(), map[string]string{
		".dockerignore":                     "*\n!lambda/**/*.go\n!go.mod\n",
		"Dockerfile":                        "FROM scratch\n",
		"go.mod":                            "module x\n",
		"lambda/describedocument/cmd/m.go":  "package main\n",
		"lambda/describedocument/cmd/m.txt": "ignored\n",
		"junk.txt":                          "ignored\n",
	})
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "go.mod", "lambda/describedocument/cmd/m.go"}, got)
}

func TestGetContext_dockerignoreKeepsDirectories(t *testing.T) {
	got := tarNames(t, t.TempDir(), map[string]string{
		".dockerignore": "nothing-matches\n!nothing\n",
		"Dockerfile":    "FROM scratch\n",
		"real/a.txt":    "a\n",
	}, "empty")
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "real/", "real/a.txt", "empty/"}, got)
}

func TestGetContext_dockerignoreKeepsEmptyDirUnderExcludedParent(t *testing.T) {
	got := tarNames(t, t.TempDir(), map[string]string{
		".dockerignore": "*\n!lambda/**/empty\n!lambda/**/a.go\n",
		"Dockerfile":    "FROM scratch\n",
		"lambda/x/a.go": "package x\n",
	}, "lambda/x/empty")
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "lambda/x/a.go", "lambda/x/empty/"}, got)
}

func TestGetContext_dockerignoreKeepsReincludedDirWithIgnoredChildren(t *testing.T) {
	got := tarNames(t, t.TempDir(), map[string]string{
		".dockerignore":       "*\n!lambda/**/keep\nlambda/**/keep/*\n",
		"Dockerfile":          "FROM scratch\n",
		"lambda/x/keep/a.txt": "ignored\n",
	})
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "lambda/x/keep/"}, got)
}

func TestGetContext_dockerignoreKeepsReincludedDirWithIncludedChild(t *testing.T) {
	got := tarNames(t, t.TempDir(), map[string]string{
		".dockerignore":       "*\n!lambda/**/keep\nlambda/**/keep/*\n!lambda/**/keep/b.go\n",
		"Dockerfile":          "FROM scratch\n",
		"lambda/x/keep/a.txt": "ignored\n",
		"lambda/x/keep/b.go":  "package x\n",
	})
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "lambda/x/keep/", "lambda/x/keep/b.go"}, got)
}

func TestGetContext_dockerignoreSkipsUnreadableExcludedDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read any directory")
	}
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret")
	require.NoError(t, os.MkdirAll(secret, 0o755))
	require.NoError(t, os.Chmod(secret, 0o000))
	t.Cleanup(func() { _ = os.Chmod(secret, 0o755) })

	got := tarNames(t, dir, map[string]string{
		".dockerignore": "secret\n!keep\n",
		"Dockerfile":    "FROM scratch\n",
		"keep":          "k\n",
	})
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "keep"}, got)
}

func TestGetContext_dockerignorePrunesExcludedDirWithoutReinclude(t *testing.T) {
	got := tarNames(t, t.TempDir(), map[string]string{
		".dockerignore":     "build\n!keep.txt\n",
		"Dockerfile":        "FROM scratch\n",
		"keep.txt":          "keep\n",
		"build/a/b/deep.go": "package b\n",
		"build/other.txt":   "ignored\n",
	})
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "keep.txt"}, got)
}

func TestGetContext_dockerignoreDoesNotPruneDirWithNestedReinclude(t *testing.T) {
	got := tarNames(t, t.TempDir(), map[string]string{
		".dockerignore":     "build\n!build/**/keep.go\n",
		"Dockerfile":        "FROM scratch\n",
		"build/a/b/keep.go": "package b\n",
		"build/a/drop.go":   "package a\n",
		"build/b/deep/x.go": "package b\n",
	})
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "build/a/b/keep.go"}, got)
}

func TestGetContext_dockerignoreInvalidNegationReturnsError(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte("![\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0o600))

	_, err := (&ContainerRequest{FromDockerfile: FromDockerfile{Context: dir}}).GetContext()
	require.ErrorIs(t, err, filepath.ErrBadPattern)
	require.EqualError(t, err, "syntax error in pattern")
}

func TestGetContext_dockerignoreWalkNeverEntersExcludedDirWithoutReinclude(t *testing.T) {
	orig := walkDir
	t.Cleanup(func() { walkDir = orig })

	dir := t.TempDir()
	var visited []string
	walkDir = func(root string, fn fs.WalkDirFunc) error {
		return orig(root, func(p string, d fs.DirEntry, err error) error {
			rel, relErr := filepath.Rel(dir, p)
			require.NoError(t, relErr)
			visited = append(visited, filepath.ToSlash(rel))
			return fn(p, d, err)
		})
	}

	got := tarNames(t, dir, map[string]string{
		".dockerignore":     "build\n!keep.txt\n",
		"Dockerfile":        "FROM scratch\n",
		"keep.txt":          "keep\n",
		"build/a/b/deep.go": "package b\n",
	})
	require.ElementsMatch(t, []string{".dockerignore", "Dockerfile", "keep.txt"}, got)

	for _, p := range visited {
		assert.False(t, strings.HasPrefix(p, "build/"), "walk entered pruned directory: %s", p)
	}
}
