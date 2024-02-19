package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbname   = "test-db"
	user     = "postgres"
	password = "password"
)

func TestPostgres(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		image string
		wait  wait.Strategy
	}{
		{
			name:  "Postgres",
			image: "docker.io/postgres:15.2-alpine",
			wait:  wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
		},
		{
			name: "Timescale",
			// timescale {
			image: "docker.io/timescale/timescaledb:2.1.0-pg11",
			wait:  wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
			// }
		},
		{
			name: "Postgis",
			// postgis {
			image: "docker.io/postgis/postgis:12-3.0",
			wait:  wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(30 * time.Second),
			// }
		},
		{
			name: "Pgvector",
			// pgvector {
			image: "docker.io/pgvector/pgvector:pg16",
			wait:  wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(30 * time.Second),
			// }
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := postgres.RunContainer(ctx,
				testcontainers.WithImage(tt.image),
				postgres.WithDatabase(dbname),
				postgres.WithUsername(user),
				postgres.WithPassword(password),
				testcontainers.WithWaitStrategy(tt.wait),
			)
			if err != nil {
				t.Fatal(err)
			}

			// Clean up the container after the test is complete
			t.Cleanup(func() {
				if err := container.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			// connectionString {
			// explicitly set sslmode=disable because the container is not configured to use TLS
			connStr, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")
			// }
			require.NoError(t, err)

			// Ensure connection string is using generic format
			id, err := container.MappedPort(ctx, "5432/tcp")
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&application_name=test", user, password, "localhost", id.Port(), dbname), connStr)

			// perform assertions
			db, err := sql.Open("postgres", connStr)
			require.NoError(t, err)
			assert.NotNil(t, db)
			defer db.Close()

			result, err := db.Exec("CREATE TABLE IF NOT EXISTS test (id int, name varchar(255));")
			require.NoError(t, err)
			assert.NotNil(t, result)

			result, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'test');")
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestContainerWithWaitForSQL(t *testing.T) {
	ctx := context.Background()

	port := "5432/tcp"
	dbURL := func(host string, port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:password@%s:%s/%s?sslmode=disable", host, port.Port(), dbname)
	}

	t.Run("default query", func(t *testing.T) {
		container, err := postgres.RunContainer(
			ctx,
			postgres.WithDatabase(dbname),
			postgres.WithUsername(user),
			postgres.WithPassword(password),
			testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port(port), "postgres", dbURL)),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
	})
	t.Run("custom query", func(t *testing.T) {
		container, err := postgres.RunContainer(
			ctx,
			postgres.WithDatabase(dbname),
			postgres.WithUsername(user),
			postgres.WithPassword(password),
			testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port(port), "postgres", dbURL).WithStartupTimeout(time.Second*5).WithQuery("SELECT 10")),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
	})
	t.Run("custom bad query", func(t *testing.T) {
		container, err := postgres.RunContainer(
			ctx,
			postgres.WithDatabase(dbname),
			postgres.WithUsername(user),
			postgres.WithPassword(password),
			testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port(port), "postgres", dbURL).WithStartupTimeout(time.Second*5).WithQuery("SELECT 'a' from b")),
		)
		require.Error(t, err)
		require.Nil(t, container)
	})
}

func TestWithConfigFile(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// explicitly set sslmode=disable because the container is not configured to use TLS
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestWithInitScript(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
		postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// explicitly set sslmode=disable because the container is not configured to use TLS
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// database created in init script. See testdata/init-user-db.sh
	result, err := db.Exec("SELECT * FROM testdb;")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSnapshot(t *testing.T) {
	// snapshotAndReset {
	ctx := context.Background()

	// 1. Start the postgres container and run any migrations on it
	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Run any migrations on the database
	_, _, err = container.Exec(ctx, []string{"psql", "-U", user, "-d", dbname, "-c", "CREATE TABLE users (id SERIAL, name TEXT NOT NULL, age INT NOT NULL)"})
	if err != nil {
		t.Fatal(err)
	}

	// 2. Create a snapshot of the database to restore later
	err = container.Snapshot(ctx, postgres.WithSnapshotName("test-snapshot"))
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
			err = container.Restore(ctx)
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
			err = container.Restore(ctx)
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
			t.Fatalf("Expected error to be a NoRows error, since the DB should be empty on every test. Got %s instead", err)
		}
	})
	// }
}
