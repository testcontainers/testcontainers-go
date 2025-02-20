package scylladb_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
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
		host, err := ctr.ConnectionHost(ctx, 19042)
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
		host, err := ctr.ConnectionHost(ctx, 19042)
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
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ctr)

	t.Run("test-without-shard-awareness", func(t *testing.T) {
		host, err := ctr.ConnectionHost(ctx, 9042)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var cluster_name string
		err = session.Query("SELECT cluster_name FROM system.local").Scan(&cluster_name)
		require.NoError(t, err)

		require.Equal(t, "Amazing ScyllaDB Test", cluster_name)
	})

	t.Run("test-with-shard-awareness", func(t *testing.T) {
		host, err := ctr.ConnectionHost(ctx, 19042)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var cluster_name string
		err = session.Query("SELECT cluster_name FROM system.local").Scan(&cluster_name)
		require.NoError(t, err)

		require.Equal(t, "Amazing ScyllaDB Test", cluster_name)
	})
}

func TestScyllaAlternator(t *testing.T) {
	ctx := context.Background()

	alternatorPort := uint16(8000)

	t.Run("test-with-alternator", func(t *testing.T) {
		ctr, err := scylladb.Run(ctx,
			"scylladb/scylla:6.2.2",
			scylladb.WithAlternator(alternatorPort),
		)
		require.NoError(t, err)
		testcontainers.CleanupContainer(t, ctr)

		client, err := getDynamoAlternatorClient(t, ctr, alternatorPort)
		require.NoError(t, err)
		err = createTable(client)
		require.NoError(t, err)
	})

	t.Run("test-without-alternator", func(t *testing.T) {
		ctr, err := scylladb.Run(ctx,
			"scylladb/scylla:6.2",
		)
		require.NoError(t, err)
		testcontainers.CleanupContainer(t, ctr)

		client, err := getDynamoAlternatorClient(t, ctr, alternatorPort)
		require.Error(t, err)
		err = createTable(client)
		require.Error(t, err)
	})
}

type scyllaAlternatorResolver struct {
	HostPort string
}

func (r *scyllaAlternatorResolver) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	return smithyendpoints.Endpoint{
		URI: url.URL{Host: r.HostPort, Scheme: "http"},
	}, nil
}

func getDynamoAlternatorClient(t *testing.T, c *scylladb.Container, port uint16) (*dynamodb.Client, error) {
	t.Helper()
	var errs []error

	hostPort, err := c.ConnectionHost(context.Background(), int(port))
	if err != nil {
		errs = append(errs, fmt.Errorf("get connection string: %w", err))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "SCYLLA_ALTERNATOR_ACCESS_KEY_ID",
			SecretAccessKey: "SCYLLA_ALTERNATOR_SECRET_ACCESS",
		},
	}))
	if err != nil {
		errs = append(errs, fmt.Errorf("load default config: %w", err))
	}

	return dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolverV2(&scyllaAlternatorResolver{HostPort: hostPort})), errors.Join(errs...)
}

func createTable(client *dynamodb.Client) error {
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
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	return nil
}
