package gcloud_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func TestBigQueryContainer(t *testing.T) {
	// runBigQueryContainer {
	ctx := context.Background()

	bigQueryContainer, err := gcloud.RunBigQueryContainer(
		ctx,
		testcontainers.WithImage("ghcr.io/goccy/bigquery-emulator:0.6.1"),
		gcloud.WithProjectID("bigquery-project"),
	)
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	// Clean up the container
	defer func() {
		if err := bigQueryContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()
	// }

	// bigQueryClient {
	projectID := bigQueryContainer.Settings.ProjectID

	opts := []option.ClientOption{
		option.WithEndpoint(bigQueryContainer.URI),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}

	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		log.Fatalf("failed to create bigquery client: %v", err) // nolint:gocritic
	}
	defer client.Close()
	// }

	createFnQuery := client.Query("CREATE FUNCTION testr(arr ARRAY<STRUCT<name STRING, val INT64>>) AS ((SELECT SUM(IF(elem.name = \"foo\",elem.val,null)) FROM UNNEST(arr) AS elem))")
	_, err = createFnQuery.Read(ctx)
	if err != nil {
		log.Fatalf("failed to create function: %v", err)
	}

	selectQuery := client.Query("SELECT testr([STRUCT<name STRING, val INT64>(\"foo\", 10), STRUCT<name STRING, val INT64>(\"bar\", 40), STRUCT<name STRING, val INT64>(\"foo\", 20)])")
	it, err := selectQuery.Read(ctx)
	if err != nil {
		log.Fatalf("failed to read query: %v", err)
	}

	var val []bigquery.Value
	for {
		err := it.Next(&val)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Fatalf("failed to iterate: %v", err)
		}
	}

	// Output:
	// [30]
	expectedValue := int64(30)
	actualValue := val[0]
	fmt.Println(val[0])

	require.NoError(t, err)
	if assert.NotNil(t, val) {
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestBigQueryWithDataYamlFile(t *testing.T) {
	// runBigQueryContainer {
	ctx := context.Background()

	absPath, err := filepath.Abs(filepath.Join(".", "testdata", "data.yaml"))
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	bigQueryContainer, err := gcloud.RunBigQueryContainer(
		ctx,
		testcontainers.WithImage("ghcr.io/goccy/bigquery-emulator:0.6.1"),
		gcloud.WithProjectID("test"),
		gcloud.WithDataYamlFile(absPath),
	)
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	// Clean up the container
	defer func() {
		if err := bigQueryContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()
	// }

	// bigQueryClient {
	projectID := bigQueryContainer.Settings.ProjectID

	opts := []option.ClientOption{
		option.WithEndpoint(bigQueryContainer.URI),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}

	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		log.Fatalf("failed to create bigquery client: %v", err) // nolint:gocritic
	}
	defer client.Close()
	// }

	selectQuery := client.Query("SELECT * FROM dataset1.table_a where name = @name")
	selectQuery.QueryConfig.Parameters = []bigquery.QueryParameter{
		{Name: "name", Value: "bob"},
	}
	it, err := selectQuery.Read(ctx)
	if err != nil {
		log.Fatalf("failed to read query: %v", err)
	}

	var val []bigquery.Value
	for {
		err := it.Next(&val)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Fatalf("failed to iterate: %v", err)
		}
	}

	// Output:
	// [30]
	expectedValue := int64(30)
	actualValue := val[0]
	assert.Equal(t, expectedValue, actualValue)
}
