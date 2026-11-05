package clickhouse_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbname   = "testdb"
	user     = "clickhouse"
	password = "password"
)

func TestClickHouseDefaultConfig(t *testing.T) {
	ctx := context.Background()

	ctr, err := clickhouse.Run(ctx, "clickhouse/clickhouse-server:23.3.8.21-alpine")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionHost, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionHost)
}

func TestClickHouseConnectionHost(t *testing.T) {
	ctx := context.Background()

	ctr, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:23.3.8.21-alpine",
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(dbname),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionHost {
	connectionHost, err := ctr.ConnectionHost(ctx)
	// }
	require.NoError(t, err)

	require.NotEmpty(t, connectionHost)
	require.Equal(t, dbname, ctr.DbName)
	require.Equal(t, user, ctr.User)
	require.Equal(t, password, ctr.Password)
}

func TestClickHouseDSN(t *testing.T) {
	ctx := context.Background()

	ctr, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:23.3.8.21-alpine",
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(dbname),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx, "debug=true")
	// }
	require.NoError(t, err)

	require.Contains(t, connectionString, "debug=true")
	require.Equal(t, dbname, ctr.DbName)
	require.Equal(t, user, ctr.User)
	require.Equal(t, password, ctr.Password)
}

func TestClickHouseWithInitScripts(t *testing.T) {
	ctx := context.Background()

	// withInitScripts {
	ctr, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:23.3.8.21-alpine",
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(dbname),
		clickhouse.WithInitScripts(filepath.Join("testdata", "init-db.sh")),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	// }

	connectionHost, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionHost)
	require.Equal(t, dbname, ctr.DbName)
	require.Equal(t, user, ctr.User)
	require.Equal(t, password, ctr.Password)
}

func TestClickHouseWithConfigFile(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		desc         string
		configOption testcontainers.CustomizeRequestOption
	}{
		{"XML_Config", clickhouse.WithConfigFile(filepath.Join("testdata", "config.xml"))},       // <allow_no_password>1</allow_no_password>
		{"YAML_Config", clickhouse.WithYamlConfigFile(filepath.Join("testdata", "config.yaml"))}, // allow_no_password: true
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctr, err := clickhouse.Run(ctx,
				"clickhouse/clickhouse-server:23.3.8.21-alpine",
				clickhouse.WithUsername(user),
				clickhouse.WithPassword(""),
				clickhouse.WithDatabase(dbname),
				tC.configOption,
			)
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			connectionHost, err := ctr.ConnectionHost(ctx)
			require.NoError(t, err)

			require.NotEmpty(t, connectionHost)
			require.Equal(t, dbname, ctr.DbName)
			require.Equal(t, user, ctr.User)
			require.Empty(t, ctr.Password)
		})
	}
}

func TestClickHouseWithZookeeper(t *testing.T) {
	ctx := context.Background()

	// withZookeeper {
	zkPort := "2181/tcp"

	zkcontainer, err := testcontainers.Run(ctx, "zookeeper:3.7", testcontainers.WithExposedPorts(zkPort), testcontainers.WithWaitStrategy(wait.ForListeningPort(zkPort)))
	testcontainers.CleanupContainer(t, zkcontainer)
	require.NoError(t, err)

	ipaddr, err := zkcontainer.ContainerIP(ctx)
	require.NoError(t, err)

	ctr, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:23.3.8.21-alpine",
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(dbname),
		clickhouse.WithZookeeper(ipaddr, "2181"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	// }

	connectionHost, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionHost)
	require.Equal(t, dbname, ctr.DbName)
	require.Equal(t, user, ctr.User)
	require.Equal(t, password, ctr.Password)
}
