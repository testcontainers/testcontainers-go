package gcloud_test

import (
	"context"
	"fmt"
	"log"

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
		log.Fatalf("failed to run container: %v", err)
	}

	// Clean up the container
	defer func() {
		if err := spannerContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
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
		log.Fatalf("failed to create instance admin client: %v", err) // nolint:gocritic
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
		log.Fatalf("failed to create instance: %v", err)
	}

	_, err = instanceOp.Wait(ctx)
	if err != nil {
		log.Fatalf("failed to wait for instance creation: %v", err)
	}

	// spannerDBAdminClient {
	c, err := database.NewDatabaseAdminClient(ctx, options...)
	if err != nil {
		log.Fatalf("failed to create admin client: %v", err)
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
		log.Fatalf("failed to create database: %v", err)
	}
	_, err = databaseOp.Wait(ctx)
	if err != nil {
		log.Fatalf("failed to wait for database creation: %v", err)
	}

	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectId, instanceId, databaseName)
	client, err := spanner.NewClient(ctx, db, options...)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	_, err = client.Apply(ctx, []*spanner.Mutation{
		spanner.Insert("Languages",
			[]string{"language", "mascot"},
			[]interface{}{"Go", "Gopher"}),
	})
	if err != nil {
		log.Fatalf("failed to apply mutation: %v", err)
	}
	row, err := client.Single().ReadRow(ctx, "Languages",
		spanner.Key{"Go"}, []string{"mascot"})
	if err != nil {
		log.Fatalf("failed to read row: %v", err)
	}

	var mascot string
	err = row.ColumnByName("Mascot", &mascot)
	if err != nil {
		log.Fatalf("failed to read column: %v", err)
	}

	fmt.Println(mascot)

	// Output:
	// Gopher
}
