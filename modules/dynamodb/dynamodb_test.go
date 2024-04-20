package dynamodb_test

import (
	"context"
	"testing"

	awsDynamoDB "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

func TestDynamoDBIsWorking(t *testing.T) {
	ctx := context.Background()

	container, err := dynamodb.RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// assert DynamoDB resources
	cl, err := container.DynamoDBClient(ctx)
	require.NoError(t, err, "It should generate DynamoDB SDK client")

	listTablesOut, err := cl.ListTables(ctx, &awsDynamoDB.ListTablesInput{})
	require.NoError(t, err, "It should list tables")
	assert.Empty(t, listTablesOut.TableNames, "It should not return any tables")
}

func TestDynamoDBWithCreateTables(t *testing.T) {
	ctx := context.Background()

	container, err := dynamodb.RunContainer(
		ctx,
		// create table1
		dynamodb.WithCreateTable(awsDynamoDB.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{{
				AttributeName: aws.String("year"),
				AttributeType: types.ScalarAttributeTypeN,
			}, {
				AttributeName: aws.String("title"),
				AttributeType: types.ScalarAttributeTypeS,
			}},
			KeySchema: []types.KeySchemaElement{{
				AttributeName: aws.String("year"),
				KeyType:       types.KeyTypeHash,
			}, {
				AttributeName: aws.String("title"),
				KeyType:       types.KeyTypeRange,
			}},
			TableName: aws.String("table1"),
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10),
				WriteCapacityUnits: aws.Int64(10),
			},
		}),
		// create table2
		dynamodb.WithCreateTable(awsDynamoDB.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{{
				AttributeName: aws.String("year"),
				AttributeType: types.ScalarAttributeTypeN,
			}, {
				AttributeName: aws.String("title"),
				AttributeType: types.ScalarAttributeTypeS,
			}},
			KeySchema: []types.KeySchemaElement{{
				AttributeName: aws.String("year"),
				KeyType:       types.KeyTypeHash,
			}, {
				AttributeName: aws.String("title"),
				KeyType:       types.KeyTypeRange,
			}},
			TableName: aws.String("table2"),
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10),
				WriteCapacityUnits: aws.Int64(10),
			},
		}),
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

	// assert DynamoDB resources
	cl, err := container.DynamoDBClient(ctx)
	require.NoError(t, err, "It should generate DynamoDB SDK client")

	listTablesOut, err := cl.ListTables(ctx, &awsDynamoDB.ListTablesInput{})
	require.NoError(t, err, "It should list tables")
	assert.Len(t, listTablesOut.TableNames, 2, "It should return 2 tables")
	assert.Contains(t, listTablesOut.TableNames, "table1", "It should contain a created table")
	assert.Contains(t, listTablesOut.TableNames, "table2", "It should contain a created table")
}
