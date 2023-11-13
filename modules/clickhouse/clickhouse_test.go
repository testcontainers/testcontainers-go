package clickhouse

import (
	"context"
	"path/filepath"
	"testing"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go"
)

const (
	dbname   = "testdb"
	user     = "clickhouse"
	password = "password"
)

type Test struct {
	Id uint64
}

func TestClickHouseDefaultConfig(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	connectionHost, err := container.ConnectionHost(ctx)
	assert.NoError(t, err)

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionHost},
		Auth: ch.Auth{
			Database: container.dbName,
			Username: container.user,
			Password: container.password,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	err = conn.Ping(context.Background())
	assert.NoError(t, err)
}

func TestClickHouseConnectionHost(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx,
		WithUsername(user),
		WithPassword(password),
		WithDatabase(dbname),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	// connectionHost {
	connectionHost, err := container.ConnectionHost(ctx)
	// }
	assert.NoError(t, err)

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionHost},
		Auth: ch.Auth{
			Database: dbname,
			Username: user,
			Password: password,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	data, err := performCRUD(conn)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
}

func TestClickHouseDSN(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, WithUsername(user), WithPassword(password), WithDatabase(dbname))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	// connectionString {
	connectionString, err := container.ConnectionString(ctx, "debug=true")
	// }
	assert.NoError(t, err)

	opts, err := ch.ParseDSN(connectionString)
	assert.NoError(t, err)

	conn, err := ch.Open(opts)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	data, err := performCRUD(conn)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
}

func TestClickHouseWithInitScripts(t *testing.T) {
	ctx := context.Background()

	// withInitScripts {
	container, err := RunContainer(ctx,
		WithUsername(user),
		WithPassword(password),
		WithDatabase(dbname),
		WithInitScripts(filepath.Join("testdata", "init-db.sh")),
	)
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	connectionHost, err := container.ConnectionHost(ctx)
	assert.NoError(t, err)

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionHost},
		Auth: ch.Auth{
			Database: dbname,
			Username: user,
			Password: password,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	data, err := getAllRows(conn)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
}

func TestClickHouseWithConfigFile(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		desc         string
		configOption testcontainers.CustomizeRequestOption
	}{
		{"XML_Config", WithConfigFile(filepath.Join("testdata", "config.xml"))},       // <allow_no_password>1</allow_no_password>
		{"YAML_Config", WithYamlConfigFile(filepath.Join("testdata", "config.yaml"))}, // allow_no_password: true
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			container, err := RunContainer(ctx,
				WithUsername(user),
				WithPassword(""),
				WithDatabase(dbname),
				tC.configOption,
			)
			if err != nil {
				t.Fatal(err)
			}

			// Clean up the container after the test is complete
			t.Cleanup(func() {
				assert.NoError(t, container.Terminate(ctx))
			})

			connectionHost, err := container.ConnectionHost(ctx)
			assert.NoError(t, err)

			conn, err := ch.Open(&ch.Options{
				Addr: []string{connectionHost},
				Auth: ch.Auth{
					Database: dbname,
					Username: user,
					// Password: password, // --> password is not required
				},
			})
			assert.NoError(t, err)
			assert.NotNil(t, conn)
			defer conn.Close()

			// perform assertions
			data, err := performCRUD(conn)
			assert.NoError(t, err)
			assert.Len(t, data, 1)
		})
	}
}

func performCRUD(conn driver.Conn) ([]Test, error) {
	var (
		err  error
		rows []Test
	)

	err = backoff.Retry(func() error {
		err = conn.Exec(context.Background(), "create table if not exists test_table (id UInt64) engine = MergeTree PRIMARY KEY (id) ORDER BY (id) SETTINGS index_granularity = 8192;")
		if err != nil {
			return err
		}

		err = conn.Exec(context.Background(), "INSERT INTO test_table (id) VALUES (1);")
		if err != nil {
			return err
		}

		rows, err = getAllRows(conn)
		if err != nil {
			return err
		}

		return nil
	}, backoff.NewExponentialBackOff())

	return rows, err
}

func getAllRows(conn driver.Conn) ([]Test, error) {
	rows, err := conn.Query(context.Background(), "SELECT * FROM test_table;")
	if err != nil {
		return nil, err
	}

	var data []Test
	for rows.Next() {
		var r Test

		err := rows.Scan(&r.Id)
		if err != nil {
			return nil, err
		}

		data = append(data, r)
	}

	return data, nil
}
