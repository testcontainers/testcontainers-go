package bigtable_test

import (
	"context"
	"testing"

	"cloud.google.com/go/bigtable"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcbigtable "github.com/testcontainers/testcontainers-go/modules/gcloud/bigtable"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	bigTableContainer, err := tcbigtable.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		tcbigtable.WithProjectID("bigtable-project"),
	)
	testcontainers.CleanupContainer(t, bigTableContainer)
	require.NoError(t, err)

	projectID := bigTableContainer.ProjectID()

	const (
		instanceID = "test-instance"
		tableName  = "test-table"
	)

	options := []option.ClientOption{
		option.WithEndpoint(bigTableContainer.URI()),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	adminClient, err := bigtable.NewAdminClient(ctx, projectID, instanceID, options...)
	require.NoError(t, err)
	defer adminClient.Close()

	err = adminClient.CreateTable(ctx, tableName)
	require.NoError(t, err)

	err = adminClient.CreateColumnFamily(ctx, tableName, "name")
	require.NoError(t, err)

	client, err := bigtable.NewClient(ctx, projectID, instanceID, options...)
	require.NoError(t, err)
	defer client.Close()

	tbl := client.Open(tableName)

	mut := bigtable.NewMutation()
	mut.Set("name", "firstName", bigtable.Now(), []byte("Gopher"))

	err = tbl.Apply(ctx, "1", mut)
	require.NoError(t, err)

	row, err := tbl.ReadRow(ctx, "1", bigtable.RowFilter(bigtable.FamilyFilter("name")))
	require.NoError(t, err)

	require.Equal(t, "Gopher", string(row["name"][0].Value))
}
