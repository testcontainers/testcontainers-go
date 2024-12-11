package gcloud_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/require"
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

	testDataPath, err := filepath.Abs(filepath.Join(".", "testdata"))
	require.NoError(t, err)

	r, err := os.Open(filepath.Join(testDataPath, "data.yaml"))
	require.NoError(t, err)

	bigQueryContainer, err := gcloud.RunBigQuery(
		ctx,
		"ghcr.io/goccy/bigquery-emulator:0.6.1",
		gcloud.WithProjectID("test"),
		gcloud.WithDataYAML(r),
	)
	testcontainers.CleanupContainer(t, bigQueryContainer)
	require.NoError(t, err)

	projectID := bigQueryContainer.Settings.ProjectID

	opts := []option.ClientOption{
		option.WithEndpoint(bigQueryContainer.URI),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}

	client, err := bigquery.NewClient(ctx, projectID, opts...)
	require.NoError(t, err)
	defer client.Close()

	selectQuery := client.Query("SELECT * FROM dataset1.table_a where name = @name")
	selectQuery.QueryConfig.Parameters = []bigquery.QueryParameter{
		{Name: "name", Value: "bob"},
	}
	it, err := selectQuery.Read(ctx)
	require.NoError(t, err)

	var val []bigquery.Value
	for {
		err := it.Next(&val)
		if errors.Is(err, iterator.Done) {
			break
		}
		require.NoError(t, err)
	}

	require.Equal(t, int64(30), val[0])
}

func TestBigQueryWithDataYamlFile_multiple(t *testing.T) {
	ctx := context.Background()

	testDataPath, err := filepath.Abs(filepath.Join(".", "testdata"))
	require.NoError(t, err)

	r1, err := os.Open(filepath.Join(testDataPath, "data.yaml"))
	require.NoError(t, err)

	r2, err := os.Open(filepath.Join(testDataPath, "data2.yaml"))
	require.NoError(t, err)

	bigQueryContainer, err := gcloud.RunBigQuery(
		ctx,
		"ghcr.io/goccy/bigquery-emulator:0.6.1",
		gcloud.WithProjectID("test"),
		gcloud.WithDataYAML(r1),
		gcloud.WithDataYAML(r2), // last file will be used
	)
	testcontainers.CleanupContainer(t, bigQueryContainer)
	require.NoError(t, err)

	projectID := bigQueryContainer.Settings.ProjectID

	opts := []option.ClientOption{
		option.WithEndpoint(bigQueryContainer.URI),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}

	client, err := bigquery.NewClient(ctx, projectID, opts...)
	require.NoError(t, err)
	defer client.Close()

	t.Run("select/dataset1-not-found", func(t *testing.T) {
		selectQuery := client.Query("SELECT * FROM dataset1.table_a where name = @name")
		selectQuery.QueryConfig.Parameters = []bigquery.QueryParameter{
			{Name: "name", Value: "bob"},
		}
		_, err := selectQuery.Read(ctx)
		require.Error(t, err)
	})

	t.Run("select/dataset2", func(t *testing.T) {
		selectQuery := client.Query("SELECT * FROM dataset2.table_b where name = @name")
		selectQuery.QueryConfig.Parameters = []bigquery.QueryParameter{
			{Name: "name", Value: "naomi"},
		}
		it, err := selectQuery.Read(ctx)
		require.NoError(t, err)

		var val []bigquery.Value
		for {
			err := it.Next(&val)
			if errors.Is(err, iterator.Done) {
				break
			}
			require.NoError(t, err)
		}

		require.Equal(t, int64(60), val[0])
	})
}
