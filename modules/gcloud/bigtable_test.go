package gcloud_test

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/bigtable"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func ExampleRunBigTableContainer() {
	// runBigTableContainer {
	ctx := context.Background()

	bigTableContainer, err := gcloud.RunBigTableContainer(
		ctx,
		testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"),
		gcloud.WithProjectID("bigtable-project"),
	)
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	// Clean up the container
	defer func() {
		if err := bigTableContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()
	// }

	// bigTableAdminClient {
	projectId := bigTableContainer.Settings.ProjectID

	const (
		instanceId = "test-instance"
		tableName  = "test-table"
	)

	options := []option.ClientOption{
		option.WithEndpoint(bigTableContainer.URI),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	adminClient, err := bigtable.NewAdminClient(ctx, projectId, instanceId, options...)
	if err != nil {
		log.Fatalf("failed to create admin client: %v", err) // nolint:gocritic
	}
	defer adminClient.Close()
	// }

	err = adminClient.CreateTable(ctx, tableName)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}
	err = adminClient.CreateColumnFamily(ctx, tableName, "name")
	if err != nil {
		log.Fatalf("failed to create column family: %v", err)
	}

	// bigTableClient {
	client, err := bigtable.NewClient(ctx, projectId, instanceId, options...)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()
	// }

	tbl := client.Open(tableName)

	mut := bigtable.NewMutation()
	mut.Set("name", "firstName", bigtable.Now(), []byte("Gopher"))
	err = tbl.Apply(ctx, "1", mut)
	if err != nil {
		log.Fatalf("failed to apply mutation: %v", err)
	}

	row, err := tbl.ReadRow(ctx, "1", bigtable.RowFilter(bigtable.FamilyFilter("name")))
	if err != nil {
		log.Fatalf("failed to read row: %v", err)
	}

	fmt.Println(string(row["name"][0].Value))

	// Output:
	// Gopher
}
