package testcontainers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// writeNestedContext builds the same 10,000-file tree (100 directories of 100
// files) plus a Dockerfile and the given .dockerignore content.
func writeNestedContext(b *testing.B, dockerignore string) string {
	b.Helper()
	dir := b.TempDir()
	must := func(err error) {
		b.Helper()
		if err != nil {
			b.Fatal(err)
		}
	}
	must(os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0o600))
	must(os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte(dockerignore), 0o600))
	for d := range 100 {
		sub := filepath.Join(dir, "pkg", fmt.Sprintf("d%03d", d), "inner")
		must(os.MkdirAll(sub, 0o755))
		for f := range 100 {
			must(os.WriteFile(filepath.Join(sub, fmt.Sprintf("f%03d.go", f)), []byte("package x\n"), 0o600))
		}
	}
	return dir
}

// BenchmarkGetContextDockerignorePreScan compares GetContext on identical trees
// whose .dockerignore is empty (no pre-scan) or holds only "!nothing" (the
// pre-scan walks the whole tree but lists no path, since no directory is
// excluded). The trees are built before the timer starts. It asserts no timing.
func BenchmarkGetContextDockerignorePreScan(b *testing.B) {
	for _, tc := range []struct{ name, ignore string }{
		{"empty", ""},
		{"nothing", "!nothing\n"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			dir := writeNestedContext(b, tc.ignore)
			b.ReportAllocs()
			for b.Loop() {
				r, err := (&ContainerRequest{FromDockerfile: FromDockerfile{Context: dir}}).GetContext()
				if err != nil {
					b.Fatal(err)
				}
				if _, err := io.Copy(io.Discard, r); err != nil {
					b.Fatal(err)
				}
				if c, ok := r.(io.Closer); ok {
					if err := c.Close(); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
