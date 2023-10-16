package cassandra

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

type Test struct {
	Id   uint64
	Name string
}

func TestCassandra(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	// connectionString {
	connectionHost, err := container.ConnectionHost(ctx)
	// }
	assert.NoError(t, err)

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	// perform assertions
	err = session.Query("CREATE KEYSPACE test_keyspace WITH REPLICATION = {'class' : 'SimpleStrategy', 'replication_factor' : 1}").Exec()
	assert.NoError(t, err)
	err = session.Query("CREATE TABLE test_keyspace.test_table (id int PRIMARY KEY, name text)").Exec()
	assert.NoError(t, err)

	err = session.Query("INSERT INTO test_keyspace.test_table (id, name) VALUES (1, 'NAME')").Exec()
	assert.NoError(t, err)

	var test Test
	err = session.Query("SELECT id, name FROM test_keyspace.test_table WHERE id=1").Scan(&test.Id, &test.Name)
	assert.NoError(t, err)
	assert.Equal(t, Test{Id: 1, Name: "NAME"}, test)
}

func TestCassandraWithConfigFile(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, WithConfigFile(filepath.Join("testdata", "config.yaml")))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	connectionHost, err := container.ConnectionHost(ctx)
	assert.NoError(t, err)

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	var result string
	err = session.Query("SELECT cluster_name FROM system.local").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, "My Cluster", result)
}

func TestCassandraWithInitScripts(t *testing.T) {
	t.Run("with init cql script", func(t *testing.T) {
		ctx := context.Background()

		// withInitScripts {
		container, err := RunContainer(ctx, WithInitScripts(filepath.Join("testdata", "init.cql")))
		// }
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

		cluster := gocql.NewCluster(connectionHost)
		session, err := cluster.CreateSession()
		if err != nil {
			t.Fatal(err)
		}
		defer session.Close()

		var test Test
		err = session.Query("SELECT id, name FROM test_keyspace.test_table WHERE id=1").Scan(&test.Id, &test.Name)
		assert.NoError(t, err)
		assert.Equal(t, Test{Id: 1, Name: "NAME"}, test)
	})

	t.Run("with init bash script", func(t *testing.T) {
		ctx := context.Background()

		container, err := RunContainer(ctx, WithInitScripts(filepath.Join("testdata", "init.sh")))
		if err != nil {
			t.Fatal(err)
		}

		// Clean up the container after the test is complete
		t.Cleanup(func() {
			assert.NoError(t, container.Terminate(ctx))
		})

		connectionHost, err := container.ConnectionHost(ctx)
		assert.NoError(t, err)

		cluster := gocql.NewCluster(connectionHost)
		session, err := cluster.CreateSession()
		if err != nil {
			t.Fatal(err)
		}
		defer session.Close()

		var test Test
		err = session.Query("SELECT id, name FROM init_sh_keyspace.test_table WHERE id=1").Scan(&test.Id, &test.Name)
		assert.NoError(t, err)
		assert.Equal(t, Test{Id: 1, Name: "NAME"}, test)
	})
}
