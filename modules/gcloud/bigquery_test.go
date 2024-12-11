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

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func ExampleRunBigQueryContainer() {
	// runBigQueryContainer {
	ctx := context.Background()

	bigQueryContainer, err := gcloud.RunBigQuery(
		ctx,
		"ghcr.io/goccy/bigquery-emulator:0.6.1",
		gcloud.WithProjectID("bigquery-project"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(bigQueryContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run container: %v", err)
		return
	}
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
		log.Printf("failed to create bigquery client: %v", err)
		return
	}
	defer client.Close()
	// }

	createFnQuery := client.Query("CREATE FUNCTION testr(arr ARRAY<STRUCT<name STRING, val INT64>>) AS ((SELECT SUM(IF(elem.name = \"foo\",elem.val,null)) FROM UNNEST(arr) AS elem))")
	_, err = createFnQuery.Read(ctx)
	if err != nil {
		log.Printf("failed to create function: %v", err)
		return
	}

	selectQuery := client.Query("SELECT testr([STRUCT<name STRING, val INT64>(\"foo\", 10), STRUCT<name STRING, val INT64>(\"bar\", 40), STRUCT<name STRING, val INT64>(\"foo\", 20)])")
	it, err := selectQuery.Read(ctx)
	if err != nil {
		log.Printf("failed to read query: %v", err)
		return
	}

	var val []bigquery.Value
	for {
		err := it.Next(&val)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Printf("failed to iterate: %v", err)
			return
		}
	}

	fmt.Println(val[0])
	// Output:
	// 30
}

func TestBigQueryWithDataYamlFile(t *testing.T) {
	ctx := context.Background()

	absPath, err := filepath.Abs(filepath.Join(".", "testdata", "data.yaml"))
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	bigQueryContainer, err := gcloud.RunBigQuery(
		ctx,
		"ghcr.io/goccy/bigquery-emulator:0.6.1",
		gcloud.WithProjectID("test"),
		gcloud.WithDataYamlFile(absPath),
	)
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	defer func() {
		if err := bigQueryContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()

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
	if expectedValue != actualValue {
		t.Errorf("BigQuery value didn't match. \nExpected %v, \nbut got: %v", expectedValue, actualValue)
	}
}
