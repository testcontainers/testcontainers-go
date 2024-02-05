package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestPostgres(t *testing.T) {
	ctx := context.Background()

	// 1. Start the postgres container and run any migrations on it
	container, err := runContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Create a snapshot of the database to restore later
	err = snapshotDB(container)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	dbURL, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test inserting a user", func(t *testing.T) {
		t.Cleanup(func() {
			// 3. In each test, reset the DB to its snapshot state.
			err = resetDB(container)
			if err != nil {
				t.Fatal(err)
			}
		})

		conn, err := pgx.Connect(context.Background(), dbURL)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close(context.Background())

		_, err = conn.Exec(ctx, "INSERT INTO users(name, age) VALUES ($1, $2)", "test", 42)
		if err != nil {
			t.Fatal(err)
		}

		var name string
		var age int64
		err = conn.QueryRow(context.Background(), "SELECT name, age FROM users LIMIT 1").Scan(&name, &age)
		if err != nil {
			t.Fatal(err)
		}

		if name != "test" {
			t.Fatalf("Expected %s to equal `test`", name)
		}
		if age != 42 {
			t.Fatalf("Expected %d to equal `42`", age)
		}
	})

	// 4. Run as many tests as you need, they will each get a clean database
	t.Run("Test querying empty DB", func(t *testing.T) {
		t.Cleanup(func() {
			err = resetDB(container)
			if err != nil {
				t.Fatal(err)
			}
		})

		conn, err := pgx.Connect(context.Background(), dbURL)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close(context.Background())

		var name string
		var age int64
		err = conn.QueryRow(context.Background(), "SELECT name, age FROM users LIMIT 1").Scan(&name, &age)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("Expected error to be a NoRows error, since the DB should be empty on every test. For %s instead", err)
		}
	})
}
