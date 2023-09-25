package gcloud_test

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	instancepb "cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func ExampleRunSpannerContainer() {
	// runSpannerContainer {
	ctx := context.Background()

	spannerContainer, err := gcloud.RunSpannerContainer(
		ctx,
		testcontainers.WithImage("gcr.io/cloud-spanner-emulator/emulator:1.4.0"),
		gcloud.WithProjectID("spanner-project"),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := spannerContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	// spannerAdminClient {
	projectId := spannerContainer.Settings.ProjectID

	const (
		instanceId   = "test-instance"
		databaseName = "test-db"
	)

	options := []option.ClientOption{
		option.WithEndpoint(spannerContainer.URI),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}

	instanceAdmin, err := instance.NewInstanceAdminClient(ctx, options...)
	if err != nil {
		panic(err)
	}
	defer instanceAdmin.Close()
	// }

	instanceOp, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectId),
		InstanceId: instanceId,
		Instance: &instancepb.Instance{
			DisplayName: instanceId,
		},
	})
	if err != nil {
		panic(err)
	}

	_, err = instanceOp.Wait(ctx)
	if err != nil {
		panic(err)
	}

	// spannerDBAdminClient {
	c, err := database.NewDatabaseAdminClient(ctx, options...)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// }

	databaseOp, err := c.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectId, instanceId),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseName),
		ExtraStatements: []string{
			"CREATE TABLE Languages (Language STRING(MAX), Mascot STRING(MAX)) PRIMARY KEY (Language)",
		},
	})
	if err != nil {
		panic(err)
	}
	_, err = databaseOp.Wait(ctx)
	if err != nil {
		panic(err)
	}

	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectId, instanceId, databaseName)
	client, err := spanner.NewClient(ctx, db, options...)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_, err = client.Apply(ctx, []*spanner.Mutation{
		spanner.Insert("Languages",
			[]string{"language", "mascot"},
			[]interface{}{"Go", "Gopher"}),
	})
	if err != nil {
		panic(err)
	}
	row, err := client.Single().ReadRow(ctx, "Languages",
		spanner.Key{"Go"}, []string{"mascot"})
	if err != nil {
		panic(err)
	}

	var mascot string
	err = row.ColumnByName("Mascot", &mascot)
	if err != nil {
		panic(err)
	}

	fmt.Println(mascot)

	// Output:
	// Gopher
}
