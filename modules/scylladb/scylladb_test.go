package scylladb_test

import (
	"bytes"
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/scylladb"
)

//go:embed testdata/scylla.yaml
var scyllaYaml []byte

func TestScyllaDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithShardAwareness(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("test-without-shard-awareness", func(t *testing.T) {
		host, err := ctr.NonShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		require.NotEmpty(t, host)
	})

	t.Run("test-with-shard-awareness", func(t *testing.T) {
		host, err := ctr.ShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		require.NotEmpty(t, host)
	})
}

func TestScyllaWithConfig(t *testing.T) {
	ctx := context.Background()

	ctr, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithConfig(bytes.NewReader(scyllaYaml)),
		scylladb.WithShardAwareness(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("test-without-shard-awareness", func(t *testing.T) {
		host, err := ctr.NonShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		require.NotEmpty(t, host)
	})

	t.Run("test-with-shard-awareness", func(t *testing.T) {
		host, err := ctr.ShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		require.NotEmpty(t, host)
	})
}

func TestScyllaAlternator(t *testing.T) {
	ctx := context.Background()

	t.Run("test-with-alternator", func(t *testing.T) {
		ctr, err := scylladb.Run(ctx,
			"scylladb/scylla:6.2.2",
			scylladb.WithAlternator(),
		)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		host, err := ctr.AlternatorConnectionHost(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, host)
	})

	t.Run("test-without-alternator", func(t *testing.T) {
		ctr, err := scylladb.Run(ctx,
			"scylladb/scylla:6.2",
		)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		_, err = ctr.AlternatorConnectionHost(ctx)
		require.Error(t, err)
	})
}

func TestWithCustomCommands(t *testing.T) {
	t.Run("invalid-flag", func(t *testing.T) {
		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: []string{"--memory=1G", "--smp=2"},
			},
		}

		// Same commands as in the Cmd, overriding the values.
		opt := scylladb.WithCustomCommands("--memory=2G", "--smp=4", "invalid-flag")

		err := opt.Customize(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid flag")

		// The invalid flag is not present in the Cmd, so the original values are kept.
		require.Len(t, req.Cmd, 2)
		require.Equal(t, "--memory=1G", req.Cmd[0])
		require.Equal(t, "--smp=2", req.Cmd[1])
	})

	t.Run("equals-override", func(t *testing.T) {
		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: []string{"--memory=1G", "--smp=2"},
			},
		}

		// Same commands as in the Cmd, overriding the values.
		opt := scylladb.WithCustomCommands("--memory=2G", "--smp=4")

		err := opt.Customize(req)
		require.NoError(t, err)

		require.Len(t, req.Cmd, 2)
		require.Equal(t, "--memory=2G", req.Cmd[0])
		require.Equal(t, "--smp=4", req.Cmd[1])
	})

	t.Run("equals-override/no-equals", func(t *testing.T) {
		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: []string{"--memory=1G", "--flag1=true", "--flag2"},
			},
		}

		// Same commands as in the Cmd, overriding the values, and adding new flags
		// of several types: with and without equals.
		opt := scylladb.WithCustomCommands("--memory=2G", "--smp=4", "--flag1=false", "--flag2", "--flag3")

		err := opt.Customize(req)
		require.NoError(t, err)

		require.Len(t, req.Cmd, 5)
		require.Equal(t, "--memory=2G", req.Cmd[0])
		require.Equal(t, "--flag1=false", req.Cmd[1])
		require.Equal(t, "--flag2", req.Cmd[2])

		// because the added flags not present in the Cmd come from a map, they could come in any order
		require.Contains(t, req.Cmd, "--smp=4")
		require.Contains(t, req.Cmd, "--flag3")
	})

	t.Run("equals-override/different-order", func(t *testing.T) {
		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "scylladb/scylla:6.2",
				Cmd:   []string{"--memory=1G", "--smp=2"},
			},
		}

		opt := scylladb.WithCustomCommands("--smp=4", "--memory=2G")

		err := opt.Customize(req)
		require.NoError(t, err)

		require.Len(t, req.Cmd, 2)
		require.Equal(t, "--memory=2G", req.Cmd[0])
		require.Equal(t, "--smp=4", req.Cmd[1])
	})
}
