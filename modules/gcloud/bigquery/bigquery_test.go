package bigquery_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcbigquery "github.com/testcontainers/testcontainers-go/modules/gcloud/bigquery"
)

//go:embed testdata/data.yaml
var dataYaml []byte

func TestBigQueryWithDataYAML(t *testing.T) {
	ctx := context.Background()

	t.Run("valid", func(t *testing.T) {
		bigQueryContainer, err := tcbigquery.Run(
			ctx,
			"ghcr.io/goccy/bigquery-emulator:0.6.1",
			tcbigquery.WithProjectID("test"),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
		)
		testcontainers.CleanupContainer(t, bigQueryContainer)
		require.NoError(t, err)

		projectID := bigQueryContainer.ProjectID()

		opts := []option.ClientOption{
			option.WithEndpoint(bigQueryContainer.URI()),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			option.WithoutAuthentication(),
			internaloption.SkipDialSettingsValidation(),
		}

		client, err := bigquery.NewClient(ctx, projectID, opts...)
		require.NoError(t, err)
		defer client.Close()

		selectQuery := client.Query("SELECT * FROM dataset1.table_a where name = @name")
		selectQuery.Parameters = []bigquery.QueryParameter{
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
	})

	t.Run("multi-value-set", func(t *testing.T) {
		bigQueryContainer, err := tcbigquery.Run(
			ctx,
			"ghcr.io/goccy/bigquery-emulator:0.6.1",
			tcbigquery.WithProjectID("test"),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
		)
		testcontainers.CleanupContainer(t, bigQueryContainer)
		require.EqualError(t, err, `data yaml already exists`)
	})

	t.Run("multi-value-not-set", func(t *testing.T) {
		noValueOption := func() testcontainers.CustomizeRequestOption {
			return func(req *testcontainers.GenericContainerRequest) error {
				req.Cmd = append(req.Cmd, "--data-from-yaml")
				return nil
			}
		}

		bigQueryContainer, err := tcbigquery.Run(
			ctx,
			"ghcr.io/goccy/bigquery-emulator:0.6.1",
			noValueOption(), // because --project is always added last, this option will receive `--project` as value, which results in an error
			tcbigquery.WithProjectID("test"),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
		)
		testcontainers.CleanupContainer(t, bigQueryContainer)
		require.Error(t, err)
	})
}
