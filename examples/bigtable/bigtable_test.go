package bigtable

import (
	"cloud.google.com/go/bigtable"
	"context"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

const (
	projectId  = "test-project"
	instanceId = "test-instance"
	tableName  = "test-table"
)

func TestBigtable(t *testing.T) {
	ctx := context.Background()

	container, err := startContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	options := []option.ClientOption{
		option.WithEndpoint(container.URI),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	adminClient, err := bigtable.NewAdminClient(ctx, projectId, instanceId, options...)
	if err != nil {
		t.Fatal(err)
	}
	err = adminClient.CreateTable(ctx, tableName)
	if err != nil {
		t.Fatal(err)
	}
	err = adminClient.CreateColumnFamily(ctx, tableName, "name")
	if err != nil {
		t.Fatal(err)
	}

	client, err := bigtable.NewClient(ctx, projectId, instanceId, options...)
	if err != nil {
		t.Fatal(err)
	}
	tbl := client.Open(tableName)

	mut := bigtable.NewMutation()
	mut.Set("name", "firstName", bigtable.Now(), []byte("Gopher"))
	err = tbl.Apply(ctx, "1", mut)
	if err != nil {
		t.Fatal(err)
	}

	row, err := tbl.ReadRow(ctx, "1", bigtable.RowFilter(bigtable.FamilyFilter("name")))
	if err != nil {
		t.Fatal(err)
	}
	// perform assertions
	name := string(row["name"][0].Value)
	if name != "Gopher" {
		t.Fatalf("expected row key to be 'Gopher', got '%s'", name)
	}
}
