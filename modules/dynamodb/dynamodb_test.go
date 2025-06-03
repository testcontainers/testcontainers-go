package dynamodb_test

import (
	"context"
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
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

const (
	tableName    string = "demo_table"
	pkColumnName string = "demo_pk"
	baseImage    string = "amazon/dynamodb-local:"
)

var image2_2_1 = baseImage + "2.2.1"

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	cli := getDynamoDBClient(t, ctr)
	require.NoError(t, err, "failed to get dynamodb client handle")

	requireTableExists(t, cli, tableName)

	value := "test_value"
	addDataToTable(t, cli, value)

	queryResult := queryItem(t, cli, value)
	require.Equal(t, value, queryResult)
}

func TestRun_withCustomImageVersion(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:2.2.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestRun_withInvalidCustomImageVersion(t *testing.T) {
	ctx := context.Background()

	_, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:0.0.7")
	require.Error(t, err)
}

func TestRun_withoutEndpointResolver(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err, "container should start successfully")

	cli := dynamodb.New(dynamodb.Options{})

	err = createTable(cli)
	require.Error(t, err)
}

func TestRun_withSharedDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithSharedDB())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	cli1 := getDynamoDBClient(t, ctr)

	requireTableExists(t, cli1, tableName)

	// create a second container: it should have the table created in the first container
	ctr2, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithSharedDB())
	testcontainers.CleanupContainer(t, ctr2)
	require.NoError(t, err)

	// fetch client handle again
	cli2 := getDynamoDBClient(t, ctr2)
	require.NoError(t, err, "failed to get dynamodb client handle")

	// list tables and verify

	result, err := cli2.ListTables(context.Background(), nil)
	require.NoError(t, err, "dynamodb list tables operation failed")

	actualTableName := result.TableNames[0]
	require.Equal(t, tableName, actualTableName)

	// add and query data from the second container
	value := "test_value"
	addDataToTable(t, cli2, value)

	// read data from the first container
	queryResult := queryItem(t, cli1, value)
	require.NoError(t, err)
	require.Equal(t, value, queryResult)
}

func TestRun_withoutSharedDB(t *testing.T) {
	ctx := context.Background()

	ctr1, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr1)
	require.NoError(t, err)

	cli := getDynamoDBClient(t, ctr1)
	require.NoError(t, err, "failed to get dynamodb client handle")

	requireTableExists(t, cli, tableName)

	// create a second container: it should not have the table created in the first container
	ctr2, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr2)
	require.NoError(t, err)

	// fetch client handle again
	cli = getDynamoDBClient(t, ctr2)
	require.NoError(t, err, "failed to get dynamodb client handle")

	// list tables and verify

	result, err := cli.ListTables(context.Background(), nil)
	require.NoError(t, err, "dynamodb list tables operation failed")
	require.Empty(t, result.TableNames, "table should not exist after restarting container")
}

func TestRun_shouldStartWithTelemetryDisabled(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithDisableTelemetry())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestRun_shouldStartWithSharedDBEnabledAndTelemetryDisabled(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithSharedDB(), tcdynamodb.WithDisableTelemetry())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func createTable(client *dynamodb.Client) error {
	_, err := client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(pkColumnName),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(pkColumnName),
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

func addDataToTable(t *testing.T, client *dynamodb.Client, val string) {
	t.Helper()

	_, err := client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			pkColumnName: &types.AttributeValueMemberS{Value: val},
		},
	})
	require.NoError(t, err)
}

func queryItem(t *testing.T, client *dynamodb.Client, val string) string {
	t.Helper()

	output, err := client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			pkColumnName: &types.AttributeValueMemberS{Value: val},
		},
	})
	require.NoError(t, err)

	result := output.Item[pkColumnName].(*types.AttributeValueMemberS)

	return result.Value
}

type dynamoDBResolver struct {
	HostPort string
}

func (r *dynamoDBResolver) ResolveEndpoint(_ context.Context, _ dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	return smithyendpoints.Endpoint{
		URI: url.URL{Host: r.HostPort, Scheme: "http"},
	}, nil
}

// getDynamoDBClient returns a new DynamoDB client with the endpoint resolver set to the DynamoDB container's host and port
func getDynamoDBClient(t *testing.T, c *tcdynamodb.DynamoDBContainer) *dynamodb.Client {
	t.Helper()

	// createClient {
	var errs []error

	hostPort, err := c.ConnectionString(context.Background())
	if err != nil {
		errs = append(errs, fmt.Errorf("get connection string: %w", err))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "DUMMYIDEXAMPLE",
			SecretAccessKey: "DUMMYEXAMPLEKEY",
		},
	}))
	if err != nil {
		errs = append(errs, fmt.Errorf("load default config: %w", err))
	}

	require.NoError(t, errors.Join(errs...))

	return dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolverV2(&dynamoDBResolver{HostPort: hostPort}))
	// }
}

func requireTableExists(t *testing.T, cli *dynamodb.Client, tableName string) {
	t.Helper()

	err := createTable(cli)
	require.NoError(t, err)

	result, err := cli.ListTables(context.Background(), nil)
	require.NoError(t, err, "dynamodb list tables operation failed")

	actualTableName := result.TableNames[0]
	require.Equal(t, tableName, actualTableName)
}
