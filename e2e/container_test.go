package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	. "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
)

func TestContainerWithWaitForSQL(t *testing.T) {
	const dbname = "postgres"

	ctx := context.Background()

	var env = map[string]string{
		"POSTGRES_PASSWORD": "password",
		"POSTGRES_USER":     "postgres",
		"POSTGRES_DB":       dbname,
	}
	var port = "5432/tcp"
	dbURL := func(host string, port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:password@%s:%s/%s?sslmode=disable", host, port.Port(), dbname)
	}

	t.Run("default query", func(t *testing.T) {
		req := ContainerRequest{
			Image:        "postgres:14.1-alpine",
			ExposedPorts: []string{port},
			Cmd:          []string{"postgres", "-c", "fsync=off"},
			Env:          env,
			WaitingFor: wait.ForSQL(nat.Port(port), "postgres", dbURL).
				WithStartupTimeout(time.Second * 5),
		}
		container, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			t.Fatal(err)
		}

		terminateContainerOnEnd(t, ctx, container)
	})
	t.Run("custom query", func(t *testing.T) {
		req := ContainerRequest{
			Image:        "postgres:14.1-alpine",
			ExposedPorts: []string{port},
			Cmd:          []string{"postgres", "-c", "fsync=off"},
			Env:          env,
			WaitingFor: wait.ForSQL(nat.Port(port), "postgres", dbURL).
				WithStartupTimeout(time.Second * 5).
				WithQuery("SELECT 10"),
		}
		container, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			t.Fatal(err)
		}

		terminateContainerOnEnd(t, ctx, container)
	})
	t.Run("custom bad query", func(t *testing.T) {
		req := ContainerRequest{
			Image:        "postgres:14.1-alpine",
			ExposedPorts: []string{port},
			Cmd:          []string{"postgres", "-c", "fsync=off"},
			Env:          env,
			WaitingFor: wait.ForSQL(nat.Port(port), "postgres", dbURL).
				WithStartupTimeout(time.Second * 5).
				WithQuery("SELECT 'a' from b"),
		}
		container, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err == nil {
			t.Fatal("expected error, but got a nil")
		}

		terminateContainerOnEnd(t, ctx, container)
	})
}

func terminateContainerOnEnd(tb testing.TB, ctx context.Context, ctr Container) {
	tb.Helper()
	if ctr == nil {
		return
	}
	tb.Cleanup(func() {
		tb.Log("terminating container")
		require.NoError(tb, ctr.Terminate(ctx))
	})
}
