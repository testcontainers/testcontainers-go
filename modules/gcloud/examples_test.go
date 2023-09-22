package gcloud_test

import (
	"context"
	"fmt"

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

	gcloudContainer, err := gcloud.RunBigTableContainer(ctx, testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := gcloudContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	const (
		projectId  = "test-project"
		instanceId = "test-instance"
		tableName  = "test-table"
	)

	options := []option.ClientOption{
		option.WithEndpoint(gcloudContainer.URI),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	adminClient, err := bigtable.NewAdminClient(ctx, projectId, instanceId, options...)
	if err != nil {
		panic(err)
	}
	err = adminClient.CreateTable(ctx, tableName)
	if err != nil {
		panic(err)
	}
	err = adminClient.CreateColumnFamily(ctx, tableName, "name")
	if err != nil {
		panic(err)
	}

	client, err := bigtable.NewClient(ctx, projectId, instanceId, options...)
	if err != nil {
		panic(err)
	}
	tbl := client.Open(tableName)

	mut := bigtable.NewMutation()
	mut.Set("name", "firstName", bigtable.Now(), []byte("Gopher"))
	err = tbl.Apply(ctx, "1", mut)
	if err != nil {
		panic(err)
	}

	row, err := tbl.ReadRow(ctx, "1", bigtable.RowFilter(bigtable.FamilyFilter("name")))
	if err != nil {
		panic(err)
	}

	fmt.Println(string(row["name"][0].Value))

	// Output:
	// Gopher
}
