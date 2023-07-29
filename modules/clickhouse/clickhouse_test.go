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

func TestClickHouse(t *testing.T) {
	t.Run("TestClickHouse_DefaultConfig", func(t *testing.T) {
		ctx := context.Background()

		container, err := RunContainer(ctx, WithDatabase(""), WithUsername(""), WithPassword(""))
		if err != nil {
			t.Fatal(err)
		}

		// Clean up the container after the test is complete
		t.Cleanup(func() {
			if err := container.Terminate(ctx); err != nil {
				t.Fatalf("failed to terminate container: %s", err)
			}
		})

		connectionString, err := container.ConnectionHost(ctx)
		assert.NoError(t, err)

		conn, err := ch.Open(&ch.Options{
			Addr: []string{connectionString},
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
	})

	t.Run("TestClickHouse_IpPort", func(t *testing.T) {
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
			if err := container.Terminate(ctx); err != nil {
				t.Fatalf("failed to terminate container: %s", err)
			}
		})

		// connectionHost {
		connectionString, err := container.ConnectionHost(ctx)
		assert.NoError(t, err)
		// }

		conn, err := ch.Open(&ch.Options{
			Addr: []string{connectionString},
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
		err = conn.Exec(context.Background(), "create table if not exists test_table (id UInt64) engine = MergeTree PRIMARY KEY (id) ORDER BY (id) SETTINGS index_granularity = 8192;")
		assert.NoError(t, err)

		err = conn.Exec(context.Background(), "INSERT INTO test_table (id) VALUES (1);")
		assert.NoError(t, err)

		rows, err := conn.Query(context.Background(), "SELECT * FROM test_table")
		assert.NoError(t, err)
		assert.NotNil(t, rows)

		type Test struct {
			Id uint64
		}

		var data []Test
		for rows.Next() {
			var r Test
			err := rows.Scan(&r.Id)

			assert.NoError(t, err)

			data = append(data, r)
		}

		assert.Len(t, data, 1)
	})

	t.Run("TestClickHouse_DSN", func(t *testing.T) {
		ctx := context.Background()

		container, err := RunContainer(ctx, WithUsername(user), WithPassword(password), WithDatabase(dbname))
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
		err = conn.Exec(context.Background(), "create table if not exists test_table (id UInt64) engine = MergeTree PRIMARY KEY (id) ORDER BY (id) SETTINGS index_granularity = 8192;")
		assert.NoError(t, err)

		err = conn.Exec(context.Background(), "INSERT INTO test_table (id) VALUES (1);")
		assert.NoError(t, err)

		rows, err := conn.Query(context.Background(), "SELECT * FROM test_table")
		assert.NoError(t, err)
		assert.NotNil(t, rows)

		type Test struct {
			Id uint64
		}

		var data []Test
		for rows.Next() {
			var r Test
			err := rows.Scan(&r.Id)

			assert.NoError(t, err)

			data = append(data, r)
		}

		assert.Len(t, data, 1)
	})
}

func TestClickHouse_WithInitScripts(t *testing.T) {
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
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connectionString, err := container.ConnectionHost(ctx)
	assert.NoError(t, err)

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionString},
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
	rows, err := conn.Query(context.Background(), "SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, rows)

	type Test struct {
		Id uint64
	}

	var data []Test
	for rows.Next() {
		var r Test
		err := rows.Scan(&r.Id)

		assert.NoError(t, err)

		data = append(data, r)
	}

	assert.Len(t, data, 1)
}

func TestClickHouse_WithConfigFile(t *testing.T) {
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
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connectionString, err := container.ConnectionHost(ctx)
	assert.NoError(t, err)

	conn, err := ch.Open(&ch.Options{
		Addr: []string{connectionString},
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
	err = conn.Exec(context.Background(), "create table if not exists test_table (id UInt64) engine = MergeTree PRIMARY KEY (id) ORDER BY (id) SETTINGS index_granularity = 8192;")
	assert.NoError(t, err)

	err = conn.Exec(context.Background(), "INSERT INTO test_table (id) VALUES (1);")
	assert.NoError(t, err)

	rows, err := conn.Query(context.Background(), "SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, rows)

	type Test struct {
		Id uint64
	}

	var data []Test
	for rows.Next() {
		var r Test
		err := rows.Scan(&r.Id)

		assert.NoError(t, err)

		data = append(data, r)
	}

	assert.Len(t, data, 1)
}
