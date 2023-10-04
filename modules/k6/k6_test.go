package k6

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestK6(t *testing.T) {
	ctx := context.Background()

	absPath, err := filepath.Abs(filepath.Join("scripts", "test.js"))
	if err != nil {
		t.Fatal(err)
	}

	container, err := RunContainer(ctx, WithTestScript(absPath))
	if err != nil {
		t.Fatal(err)
	}
	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// assert the test script was executed
	logs, err := container.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}

	buffer := bytes.Buffer{}
	_, err = buffer.ReadFrom(logs)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buffer.String(), "Test executed") {
		t.Fatalf("expected 'Test executed'. got %q", buffer.String())
	}
}
