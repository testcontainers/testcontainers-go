package gcloud

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestRunBigTableContainer(t *testing.T) {
	ctx := context.Background()

	container, err := RunBigTableContainer(ctx, testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
}
