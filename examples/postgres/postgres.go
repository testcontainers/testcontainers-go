package postgres

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// runContainer {
const (
	dbName     = "test"
	dbUser     = "test"
	dbPassword = "test"
)

// runContainer creates an instance of the Postgres container type
func runContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	// Run any migrations on the database
	_, _, err = container.Exec(ctx, []string{"psql", "-U", dbUser, "-c", "CREATE TABLE users (id SERIAL, name TEXT NOT NULL, age INT NOT NULL)"})
	if err != nil {
		return nil, err
	}

	return container, nil
}

// }

// snapshotAndReset {
// snapshotDB takes a snapshot of the current state of the database as a template, which can then be restored using
// resetDB.
func snapshotDB(container *postgres.PostgresContainer) error {
	ctx := context.Background()

	// Create a copy of the database to another database to use as a template now that it was fully migrated
	_, _, err := container.Exec(ctx, []string{"psql", "-U", dbUser, "-c", "CREATE DATABASE migrated_template WITH TEMPLATE " + dbName + " OWNER " + dbUser})
	if err != nil {
		return err
	}

	// Snapshot the template database so we can restore it onto our original database going forward
	_, _, err = container.Exec(ctx, []string{"psql", "-U", dbUser, "-c", "ALTER DATABASE migrated_template WITH is_template = TRUE"})
	if err != nil {
		return err
	}

	return nil
}

// resetDB will reset the DB to its original migrated state, cleaning all the data.
func resetDB(container *postgres.PostgresContainer) error {
	ctx := context.Background()

	// Drop the entire database by connecting to the postgres global database
	_, _, err := container.Exec(ctx, []string{"psql", "-U", dbUser, "-d", "postgres", "-c", "DROP DATABASE " + dbName + " with (FORCE)"})
	if err != nil {
		return err
	}

	// Then restore the previous snapshot
	_, _, err = container.Exec(ctx, []string{"psql", "-U", dbUser, "-d", "postgres", "-c", "CREATE DATABASE " + dbName + " WITH TEMPLATE migrated_template OWNER " + dbUser})
	if err != nil {
		return err
	}

	return nil
}

// }
