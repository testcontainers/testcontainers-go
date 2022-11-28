package spanner

import (
	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"context"
	"fmt"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	"testing"
)

const (
	projectId    = "test-project"
	instanceId   = "test-instance"
	databaseName = "test-db"
)

func TestSpanner(t *testing.T) {
	ctx := context.Background()

	container, err := setupSpanner(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Setenv("SPANNER_EMULATOR_HOST", container.GRPCEndpoint)

	instanceAdmin, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer instanceAdmin.Close()

	instanceOp, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectId),
		InstanceId: instanceId,
		Instance: &instancepb.Instance{
			DisplayName: instanceId,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = instanceOp.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}

	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	databaseOp, err := c.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectId, instanceId),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseName),
		ExtraStatements: []string{
			"CREATE TABLE Languages (Language STRING(MAX), Mascot STRING(MAX)) PRIMARY KEY (Language)",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = databaseOp.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}

	client, err := spanner.NewClient(ctx, fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectId, instanceId, databaseName))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	_, err = client.Apply(ctx, []*spanner.Mutation{
		spanner.Insert("Languages",
			[]string{"language", "mascot"},
			[]interface{}{"Go", "Gopher"})})
	if err != nil {
		t.Fatal(err)
	}
	row, err := client.Single().ReadRow(ctx, "Languages",
		spanner.Key{"Go"}, []string{"mascot"})
	if err != nil {
		t.Fatal(err)
	}

	var mascot string
	err = row.ColumnByName("Mascot", &mascot)
	if err != nil {
		t.Fatal(err)
	}
	// perform assertions
	if mascot != "Gopher" {
		t.Fatalf("Expected value %s. Got %s.", "Gopher", mascot)
	}
}
