package clickhouse

import (
	"context"
	"path/filepath"
	"testing"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
)

const dbname = "testdb"
const user = "clickhouse"
const password = "password"

const createTableQuery = "create table if not exists test_table (id UInt64) engine = MergeTree PRIMARY KEY (id) ORDER BY (id) SETTINGS index_granularity = 8192;"
const insertQuery = "INSERT INTO test_table (id) VALUES (1);"
const selectQuery = "SELECT * FROM test_table;"

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

func TestClickHouseIpPort(t *testing.T) {
	ctx := context.Background()

	// customInitialization {
	container, err := RunContainer(ctx,
		WithUsername(user),
		WithPassword(password),
		WithDatabase(dbname),
	)
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	// connectionHost {
	connectionHost, err := container.ConnectionHost(ctx)
	assert.NoError(t, err)
	// }

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
	err = conn.Exec(context.Background(), createTableQuery)
	assert.NoError(t, err)

	err = conn.Exec(context.Background(), insertQuery)
	assert.NoError(t, err)

	rows, err := conn.Query(context.Background(), selectQuery)
	assert.NoError(t, err)
	assert.NotNil(t, rows)

	var data []Test
	for rows.Next() {
		var r Test
		err := rows.Scan(&r.Id)

		assert.NoError(t, err)

		data = append(data, r)
	}

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
	assert.NoError(t, err)
	// }

	opts, err := ch.ParseDSN(connectionString)
	assert.NoError(t, err)

	conn, err := ch.Open(opts)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	err = conn.Exec(context.Background(), createTableQuery)
	assert.NoError(t, err)

	err = conn.Exec(context.Background(), insertQuery)
	assert.NoError(t, err)

	rows, err := conn.Query(context.Background(), selectQuery)
	assert.NoError(t, err)
	assert.NotNil(t, rows)

	var data []Test
	for rows.Next() {
		var r Test
		err := rows.Scan(&r.Id)

		assert.NoError(t, err)

		data = append(data, r)
	}

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
	rows, err := conn.Query(context.Background(), selectQuery)
	assert.NoError(t, err)
	assert.NotNil(t, rows)

	var data []Test
	for rows.Next() {
		var r Test
		err := rows.Scan(&r.Id)

		assert.NoError(t, err)

		data = append(data, r)
	}

	assert.Len(t, data, 1)
}

func TestClickHouseWithConfigFile(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx,
		WithUsername(user),
		WithPassword(""),
		WithDatabase(dbname),
		WithConfigFile(filepath.Join("testdata", "config.xml")), // allow_no_password = 1
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
			// Password: password,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// perform assertions
	err = conn.Exec(context.Background(), createTableQuery)
	assert.NoError(t, err)

	err = conn.Exec(context.Background(), insertQuery)
	assert.NoError(t, err)

	rows, err := conn.Query(context.Background(), selectQuery)
	assert.NoError(t, err)
	assert.NotNil(t, rows)

	var data []Test
	for rows.Next() {
		var r Test
		err := rows.Scan(&r.Id)

		assert.NoError(t, err)

		data = append(data, r)
	}

	assert.Len(t, data, 1)
}
