package gcloud_test

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigtable"
	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func ExampleRunBigTableContainer() {
	// runBigTableContainer {
	ctx := context.Background()

	bigTableContainer, err := gcloud.RunBigTableContainer(ctx, testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := bigTableContainer.Terminate(ctx); err != nil {
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
		option.WithEndpoint(bigTableContainer.URI),
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

func ExampleRunDatastoreContainer() {
	// runDatastoreContainer {
	ctx := context.Background()

	datastoreContainer, err := gcloud.RunDatastoreContainer(ctx, testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := datastoreContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	options := []option.ClientOption{
		option.WithEndpoint(datastoreContainer.URI),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	dsClient, err := datastore.NewClient(ctx, "test-project", options...)
	if err != nil {
		panic(err)
	}
	defer dsClient.Close()

	type Task struct {
		Description string
	}

	k := datastore.NameKey("Task", "sample", nil)
	data := Task{
		Description: "my description",
	}
	_, err = dsClient.Put(ctx, k, &data)
	if err != nil {
		panic(err)
	}

	saved := Task{}
	err = dsClient.Get(ctx, k, &saved)
	if err != nil {
		panic(err)
	}

	fmt.Println(saved.Description)

	// Output:
	// my description
}
