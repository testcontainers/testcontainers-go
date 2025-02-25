package scylladb_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net/url"
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

func TestWithCustomCommands(t *testing.T) {
	t.Run("invalid-flag", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: []string{"--memory=1G", "--smp=2"},
			},
		}

		// Same commands as in the Cmd, overriding the values.
		opt := scylladb.WithCustomCommands("--memory=2G", "--smp=4", "invalid-flag")

		err := opt.Customize(&req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid flag")

		// The invalid flag is not present in the Cmd, so the original values are kept.
		require.Len(t, req.Cmd, 2)
		require.Equal(t, "--memory=1G", req.Cmd[0])
		require.Equal(t, "--smp=2", req.Cmd[1])
	})

	t.Run("equals-override", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: []string{"--memory=1G", "--smp=2"},
			},
		}

		// Same commands as in the Cmd, overriding the values.
		opt := scylladb.WithCustomCommands("--memory=2G", "--smp=4")

		err := opt.Customize(&req)
		require.NoError(t, err)

		require.Len(t, req.Cmd, 2)
		require.Equal(t, "--memory=2G", req.Cmd[0])
		require.Equal(t, "--smp=4", req.Cmd[1])
	})

	t.Run("equals-override/no-equals", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: []string{"--memory=1G", "--flag1=true", "--flag2"},
			},
		}

		// Same commands as in the Cmd, overriding the values, and adding new flags
		// of several types: with and without equals.
		opt := scylladb.WithCustomCommands("--memory=2G", "--smp=4", "--flag1=false", "--flag2", "--flag3")

		err := opt.Customize(&req)
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
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "scylladb/scylla:6.2",
				Cmd:   []string{"--memory=1G", "--smp=2"},
			},
		}

		opt := scylladb.WithCustomCommands("--smp=4", "--memory=2G")

		err := opt.Customize(&req)
		require.NoError(t, err)

		require.Len(t, req.Cmd, 2)
		require.Equal(t, "--memory=2G", req.Cmd[0])
		require.Equal(t, "--smp=4", req.Cmd[1])
	})
}
