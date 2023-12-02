package mssql

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestMSSQLServer(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("mcr.microsoft.com/mssql/server:2022-latest"))
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
