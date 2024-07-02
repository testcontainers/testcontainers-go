package postgres_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun() {
	// runPostgresContainer {
	ctx := context.Background()

	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := postgresContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
