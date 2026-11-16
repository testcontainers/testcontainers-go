package examples_test

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/scylladb"
)

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

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var address string
		err = session.Query("SELECT address FROM system.clients").Scan(&address)
		require.NoError(t, err)
	})

	t.Run("test-with-shard-awareness", func(t *testing.T) {
		host, err := ctr.ShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var address string
		err = session.Query("SELECT address FROM system.clients").Scan(&address)
		require.NoError(t, err)
	})
}

func TestScyllaWithConfig(t *testing.T) {
	scyllaYaml, err := os.ReadFile("../testdata/scylla.yaml")
	require.NoError(t, err)
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

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var clusterName string
		err = session.Query("SELECT cluster_name FROM system.local").Scan(&clusterName)
		require.NoError(t, err)

		// the "Amazing ScyllaDB Test" cluster name is set in the scylla.yaml file
		require.Equal(t, "Amazing ScyllaDB Test", clusterName)
	})

	t.Run("test-with-shard-awareness", func(t *testing.T) {
		host, err := ctr.ShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var clusterName string
		err = session.Query("SELECT cluster_name FROM system.local").Scan(&clusterName)
		require.NoError(t, err)

		// the "Amazing ScyllaDB Test" cluster name is set in the scylla.yaml file
		require.Equal(t, "Amazing ScyllaDB Test", clusterName)
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

		cli, err := getDynamoAlternatorClient(t, ctr)
		require.NoError(t, err)
		requireCreateTable(t, cli)
	})

	t.Run("test-without-alternator", func(t *testing.T) {
		ctr, err := scylladb.Run(ctx,
			"scylladb/scylla:6.2",
		)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		cli, err := getDynamoAlternatorClient(t, ctr)
		require.Error(t, err)
		require.Nil(t, cli)
	})
}

// scyllaAlternatorResolver is a custom endpoint resolver for the ScyllaDB Alternator.
type scyllaAlternatorResolver struct {
	HostPort string
}

// ResolveEndpoint resolves the endpoint for the ScyllaDB Alternator.
func (r *scyllaAlternatorResolver) ResolveEndpoint(_ context.Context, _ dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	return smithyendpoints.Endpoint{
		URI: url.URL{Host: r.HostPort, Scheme: "http"},
	}, nil
}

// getDynamoAlternatorClient returns a new DynamoDB client for the ScyllaDB Alternator.
func getDynamoAlternatorClient(t *testing.T, c *scylladb.Container) (*dynamodb.Client, error) {
	t.Helper()

	hostPort, err := c.AlternatorConnectionHost(context.Background())
	if err != nil {
		return nil, fmt.Errorf("connection host: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "SCYLLA_ALTERNATOR_ACCESS_KEY_ID",
			SecretAccessKey: "SCYLLA_ALTERNATOR_SECRET_ACCESS",
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("load default config: %w", err)
	}

	return dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolverV2(&scyllaAlternatorResolver{HostPort: hostPort})), nil
}

// requireCreateTable creates a table in the ScyllaDB Alternator, failing the test if an error occurs.
func requireCreateTable(t *testing.T, client *dynamodb.Client) {
	t.Helper()

	_, err := client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String("demo_table"),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	require.NoError(t, err)
}
