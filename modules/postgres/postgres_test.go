package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := RunContainer(ctx,
				testcontainers.WithImage(tt.image),
				WithDatabase(dbname),
				WithUsername(user),
				WithPassword(password),
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
			assert.NoError(t, err)

			// Ensure connection string is using generic format
			id, err := container.MappedPort(ctx, "5432/tcp")
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&application_name=test", user, password, "localhost", id.Port(), dbname), connStr)

			// perform assertions
			db, err := sql.Open("postgres", connStr)
			assert.NoError(t, err)
			assert.NotNil(t, db)
			defer db.Close()

			result, err := db.Exec("CREATE TABLE IF NOT EXISTS test (id int, name varchar(255));")
			assert.NoError(t, err)
			assert.NotNil(t, result)

			result, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'test');")
			assert.NoError(t, err)
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
		container, err := RunContainer(
			ctx,
			WithDatabase(dbname),
			WithUsername(user),
			WithPassword(password),
			testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port(port), "postgres", dbURL)),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
	})
	t.Run("custom query", func(t *testing.T) {
		container, err := RunContainer(
			ctx,
			WithDatabase(dbname),
			WithUsername(user),
			WithPassword(password),
			testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port(port), "postgres", dbURL).WithStartupTimeout(time.Second*5).WithQuery("SELECT 10")),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
	})
	t.Run("custom bad query", func(t *testing.T) {
		container, err := RunContainer(
			ctx,
			WithDatabase(dbname),
			WithUsername(user),
			WithPassword(password),
			testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port(port), "postgres", dbURL).WithStartupTimeout(time.Second*5).WithQuery("SELECT 'a' from b")),
		)
		require.Error(t, err)
		require.Nil(t, container)
	})
}

func TestWithConfigFile(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx,
		WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		WithDatabase(dbname),
		WithUsername(user),
		WithPassword(password),
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
	assert.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestWithInitScript(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
		WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		WithDatabase(dbname),
		WithUsername(user),
		WithPassword(password),
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
	assert.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// database created in init script. See testdata/init-user-db.sh
	result, err := db.Exec("SELECT * FROM testdb;")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
