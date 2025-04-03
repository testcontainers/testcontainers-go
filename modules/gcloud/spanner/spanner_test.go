package spanner_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	instancepb "cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcspanner "github.com/testcontainers/testcontainers-go/modules/gcloud/spanner"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	spannerContainer, err := tcspanner.Run(
		ctx,
		"gcr.io/cloud-spanner-emulator/emulator:1.4.0",
		tcspanner.WithProjectID("spanner-project"),
	)
	testcontainers.CleanupContainer(t, spannerContainer)
	require.NoError(t, err)

	projectID := spannerContainer.ProjectID()

	const (
		instanceID   = "test-instance"
		databaseName = "test-db"
	)

	options := []option.ClientOption{
		option.WithEndpoint(spannerContainer.URI()),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}

	instanceAdmin, err := instance.NewInstanceAdminClient(ctx, options...)
	if err != nil {
		log.Printf("failed to create instance admin client: %v", err)
		return
	}
	defer instanceAdmin.Close()

	instanceOp, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     "projects/" + projectID,
		InstanceId: instanceID,
		Instance: &instancepb.Instance{
			DisplayName: instanceID,
		},
	})
	require.NoError(t, err)

	_, err = instanceOp.Wait(ctx)
	require.NoError(t, err)

	c, err := database.NewDatabaseAdminClient(ctx, options...)
	require.NoError(t, err)
	defer c.Close()

	databaseOp, err := c.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseName),
		ExtraStatements: []string{
			"CREATE TABLE Languages (Language STRING(MAX), Mascot STRING(MAX)) PRIMARY KEY (Language)",
		},
	})
	require.NoError(t, err)

	_, err = databaseOp.Wait(ctx)
	require.NoError(t, err)

	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseName)
	client, err := spanner.NewClient(ctx, db, options...)
	require.NoError(t, err)
	defer client.Close()

	_, err = client.Apply(ctx, []*spanner.Mutation{
		spanner.Insert("Languages",
			[]string{"language", "mascot"},
			[]any{"Go", "Gopher"}),
	})
	require.NoError(t, err)

	row, err := client.Single().ReadRow(ctx, "Languages",
		spanner.Key{"Go"}, []string{"mascot"})
	require.NoError(t, err)

	var mascot string
	err = row.ColumnByName("Mascot", &mascot)
	require.NoError(t, err)

	require.Equal(t, "Gopher", mascot)
}
