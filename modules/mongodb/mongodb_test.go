package mongodb_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/mongodb/internal/mongotest"
)

func TestMongoDB(t *testing.T) {
	type tests struct {
		name string
		img  string
		opts []testcontainers.ContainerCustomizer
	}
	testCases := []tests{
		{
			name: "From Docker Hub",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "Community Server",
			img:  "mongodb/mongodb-community-server:7.0.2-ubi8",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "Enterprise Server",
			img:  "mongodb/mongodb-enterprise-server:7.0.0-ubi8",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "with-replica/mongo:4",
			img:  "mongo:4",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
		{
			name: "with-replica/mongo:6",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
		{
			name: "with-replica/mongo:7",
			img:  "mongo:7",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
		{
			name: "with-auth/replica/mongo:7",
			img:  "mongo:7",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongo:6",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/mongo:6",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongodb-enterprise-server:7.0.0-ubi8",
			img:  "mongodb/mongodb-enterprise-server:7.0.0-ubi8",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongodb-community-server:7.0.2-ubi8",
			img:  "mongodb/mongodb-community-server:7.0.2-ubi8",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongo:4",
			img:  "mongo:4",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			ctx := context.Background()

			mongodbContainer, err := mongodb.Run(ctx, tc.img, tc.opts...)
			testcontainers.CleanupContainer(tt, mongodbContainer)
			require.NoError(tt, err)

			endpoint, err := mongodbContainer.ConnectionString(ctx)
			require.NoError(tt, err)

			shell := mongotest.New(tt, ctx, endpoint)
			var basic struct {
				Ping         int
				Database     string
				Acknowledged bool
			}
			err = shell.Run(ctx, `
const ping = commandWorked(db.getSiblingDB('admin').runCommand({ping:1}));
const database = db.getSiblingDB('test').getName();
const inserted = acknowledged(db.getSiblingDB('testcontainer').getCollection('test').insertOne({}));
return {ping:ping.ok, database, acknowledged:inserted.acknowledged};
`, &basic)
			require.NoError(tt, err)
			require.Equal(tt, 1, basic.Ping)
			require.Equal(tt, "test", basic.Database)
			require.True(tt, basic.Acknowledged)

			// Establish the stream before inserting, through the same mapped endpoint.
			hasReplica, err := hasReplicaSet(endpoint)
			require.NoError(tt, err)
			if hasReplica {
				var change struct {
					Acknowledged  bool
					OperationType string
					Message       string
				}
				err = shell.Run(ctx, `
const coll = db.getSiblingDB('test').getCollection('changes');
const stream = coll.watch([], {fullDocument:'updateLookup', maxAwaitTimeMS:500});
try {
  // watch() is lazy in mongosh; this call opens the cursor before the write.
  if (stream.tryNext() !== null) throw new Error('unexpected pre-insert change');
  const inserted = acknowledged(coll.insertOne({message:'hello change streams'}));
  const deadline = Date.now() + 30000;
  let event = null;
  while (event === null && Date.now() < deadline) event = stream.tryNext();
  if (event === null) throw new Error('no change stream event');
  return {acknowledged:inserted.acknowledged, operationType:event.operationType, message:event.fullDocument.message};
} finally { stream.close(); }
`, &change)
				require.NoError(tt, err)
				require.True(tt, change.Acknowledged)
				require.Equal(tt, "insert", change.OperationType, "Expected operationType to be 'insert'")
				require.Equal(tt, "hello change streams", change.Message)
			}
		})
	}
}

// hasReplicaSet checks if the connection string includes a replicaSet query parameter.
func hasReplicaSet(connStr string) (bool, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return false, fmt.Errorf("parse connection string: %w", err)
	}
	q := u.Query()
	return q.Get("replicaSet") != "", nil
}
