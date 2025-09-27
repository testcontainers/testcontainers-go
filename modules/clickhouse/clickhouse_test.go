package clickhouse_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/cenkalti/backoff/v4"
	"github.com/docker/go-connections/nat"
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

type Test struct {
	ID uint64
}

func TestClickHouseDefaultConfig(t *testing.T) {
	ctx := t.Context()

	ctr, err := clickhouse.Run(ctx, "clickhouse/clickhouse-server:23.3.8.21-alpine")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionHost, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionHost},
		Auth: ch.Auth{
			Database: ctr.DbName,
			Username: ctr.User,
			Password: ctr.Password,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	err = conn.Ping(t.Context())
	require.NoError(t, err)
}

func TestClickHouseConnectionHost(t *testing.T) {
	ctx := t.Context()

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

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionHost},
		Auth: ch.Auth{
			Database: dbname,
			Username: user,
			Password: password,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	data, err := performCRUD(t, conn)
	require.NoError(t, err)
	require.Len(t, data, 1)
}

func TestClickHouseDSN(t *testing.T) {
	ctx := t.Context()

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

	opts, err := ch.ParseDSN(connectionString)
	require.NoError(t, err)

	opts.Debugf = t.Logf
	conn, err := ch.Open(opts)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	data, err := performCRUD(t, conn)
	require.NoError(t, err)
	require.Len(t, data, 1)
}

func TestClickHouseWithInitScripts(t *testing.T) {
	ctx := t.Context()

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

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionHost},
		Auth: ch.Auth{
			Database: dbname,
			Username: user,
			Password: password,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	data, err := getAllRows(ctx, conn)
	require.NoError(t, err)
	require.Len(t, data, 1)
}

func TestClickHouseWithConfigFile(t *testing.T) {
	ctx := t.Context()

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

			conn, err := ch.Open(&ch.Options{
				Addr: []string{connectionHost},
				Auth: ch.Auth{
					Database: dbname,
					Username: user,
					// Password: password, // --> password is not required
				},
			})
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer conn.Close()

			// perform assertions
			data, err := performCRUD(t, conn)
			require.NoError(t, err)
			require.Len(t, data, 1)
		})
	}
}

func TestClickHouseWithZookeeper(t *testing.T) {
	ctx := t.Context()

	// withZookeeper {
	zkPort := nat.Port("2181/tcp")

	zkcontainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			ExposedPorts: []string{zkPort.Port()},
			Image:        "zookeeper:3.7",
			WaitingFor:   wait.ForListeningPort(zkPort),
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, zkcontainer)
	require.NoError(t, err)

	ipaddr, err := zkcontainer.ContainerIP(ctx)
	require.NoError(t, err)

	ctr, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:23.3.8.21-alpine",
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(dbname),
		clickhouse.WithZookeeper(ipaddr, zkPort.Port()),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	// }

	connectionHost, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionHost},
		Auth: ch.Auth{
			Database: dbname,
			Username: user,
			Password: password, // --> password is not required
		},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	data, err := performReplicatedCRUD(t, conn)
	require.NoError(t, err)
	require.Len(t, data, 1)
}

func performReplicatedCRUD(t *testing.T, conn driver.Conn) ([]Test, error) {
	t.Helper()
	return backoff.RetryNotifyWithData(
		func() ([]Test, error) {
			err := conn.Exec(t.Context(), "CREATE TABLE replicated_test_table (id UInt64) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/mdb.data_transfer_cp_cdc', '{replica}') PRIMARY KEY (id) ORDER BY (id) SETTINGS index_granularity = 8192;")
			if err != nil {
				return nil, err
			}

			err = conn.Exec(t.Context(), "INSERT INTO replicated_test_table (id) VALUES (1);")
			if err != nil {
				return nil, err
			}

			rows, err := conn.Query(t.Context(), "SELECT * FROM replicated_test_table;")
			if err != nil {
				return nil, err
			}

			var res []Test
			for rows.Next() {
				var r Test

				err := rows.Scan(&r.ID)
				if err != nil {
					return nil, err
				}

				res = append(res, r)
			}
			return res, nil
		},
		backoff.NewExponentialBackOff(),
		func(err error, _ time.Duration) {
			t.Log(err)
		},
	)
}

func performCRUD(t *testing.T, conn driver.Conn) ([]Test, error) {
	t.Helper()
	return backoff.RetryNotifyWithData(
		func() ([]Test, error) {
			ctx := t.Context()
			err := conn.Exec(ctx, "create table if not exists test_table (id UInt64) engine = MergeTree PRIMARY KEY (id) ORDER BY (id) SETTINGS index_granularity = 8192;")
			if err != nil {
				return nil, err
			}

			err = conn.Exec(ctx, "INSERT INTO test_table (id) VALUES (1);")
			if err != nil {
				return nil, err
			}

			return getAllRows(ctx, conn)
		},
		backoff.NewExponentialBackOff(),
		func(err error, _ time.Duration) {
			t.Log(err)
		},
	)
}

func getAllRows(ctx context.Context, conn driver.Conn) ([]Test, error) {
	rows, err := conn.Query(ctx, "SELECT * FROM test_table;")
	if err != nil {
		return nil, err
	}

	var data []Test
	for rows.Next() {
		var r Test

		err := rows.Scan(&r.ID)
		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
