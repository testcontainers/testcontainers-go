package cockroachdb_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func TestCockroach(t *testing.T) {
	ctx := context.Background()

	container, err := cockroachdb.RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	conn, err := pgx.Connect(ctx, container.MustConnectionString(ctx))
	if err != nil {
		t.Fatal(err)
	}

	if err := conn.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}
