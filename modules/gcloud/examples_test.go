package gcloud_test

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigtable"
	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	instancepb "cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func ExampleRunBigQueryContainer() {
	// runBigQueryContainer {
	ctx := context.Background()

	bigQueryContainer, err := gcloud.RunBigQueryContainer(ctx, testcontainers.WithImage("ghcr.io/goccy/bigquery-emulator:0.4.3"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := bigQueryContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	const (
		projectID = "test-project"
	)

	// bigQueryClient {
	opts := []option.ClientOption{
		option.WithEndpoint(bigQueryContainer.URI),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		internaloption.SkipDialSettingsValidation(),
	}

	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	// }

	createFnQuery := client.Query("CREATE FUNCTION testr(arr ARRAY<STRUCT<name STRING, val INT64>>) AS ((SELECT SUM(IF(elem.name = \"foo\",elem.val,null)) FROM UNNEST(arr) AS elem))")
	_, err = createFnQuery.Read(ctx)
	if err != nil {
		panic(err)
	}

	selectQuery := client.Query("SELECT testr([STRUCT<name STRING, val INT64>(\"foo\", 10), STRUCT<name STRING, val INT64>(\"bar\", 40), STRUCT<name STRING, val INT64>(\"foo\", 20)])")
	it, err := selectQuery.Read(ctx)
	if err != nil {
		panic(err)
	}

	var val []bigquery.Value
	for {
		err := it.Next(&val)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(val)

	// Output:
	// [30]
}

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

	// bigTableAdminClient {
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
	defer adminClient.Close()
	// }

	err = adminClient.CreateTable(ctx, tableName)
	if err != nil {
		panic(err)
	}
	err = adminClient.CreateColumnFamily(ctx, tableName, "name")
	if err != nil {
		panic(err)
	}

	// bigTableClient {
	client, err := bigtable.NewClient(ctx, projectId, instanceId, options...)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	// }

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

	// datastoreClient {
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
	// }

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

type emulatorCreds struct{}

func (ec emulatorCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer owner"}, nil
}

func (ec emulatorCreds) RequireTransportSecurity() bool {
	return false
}

func ExampleRunFirestoreContainer() {
	// runFirestoreContainer {
	ctx := context.Background()

	firestoreContainer, err := gcloud.RunFirestoreContainer(ctx, testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := firestoreContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	// firestoreClient {
	conn, err := grpc.Dial(firestoreContainer.URI, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(emulatorCreds{}))
	if err != nil {
		panic(err)
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := firestore.NewClient(ctx, "test-project", options...)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	// }

	users := client.Collection("users")
	docRef := users.Doc("alovelace")

	type Person struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	data := Person{
		Firstname: "Ada",
		Lastname:  "Lovelace",
	}
	_, err = docRef.Create(ctx, data)
	if err != nil {
		panic(err)
	}

	docsnap, err := docRef.Get(ctx)
	if err != nil {
		panic(err)
	}

	var saved Person
	if err := docsnap.DataTo(&saved); err != nil {
		panic(err)
	}

	fmt.Println(saved.Firstname, saved.Lastname)

	// Output:
	// Ada Lovelace
}

func ExampleRunPubsubContainer() {
	// runPubsubContainer {
	ctx := context.Background()

	pubsubContainer, err := gcloud.RunPubsubContainer(ctx, testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := pubsubContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	// pubsubClient {
	conn, err := grpc.Dial(pubsubContainer.URI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := pubsub.NewClient(ctx, "my-project-id", options...)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	// }

	topic, err := client.CreateTopic(ctx, "greetings")
	if err != nil {
		panic(err)
	}
	subscription, err := client.CreateSubscription(ctx, "subscription",
		pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		panic(err)
	}
	result := topic.Publish(ctx, &pubsub.Message{Data: []byte("Hello World")})
	_, err = result.Get(ctx)
	if err != nil {
		panic(err)
	}

	var data []byte
	cctx, cancel := context.WithCancel(ctx)
	err = subscription.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
		data = m.Data
		m.Ack()
		defer cancel()
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	// Output:
	// Hello World
}

func ExampleRunSpannerContainer() {
	// runSpannerContainer {
	ctx := context.Background()

	spannerContainer, err := gcloud.RunSpannerContainer(ctx, testcontainers.WithImage("gcr.io/cloud-spanner-emulator/emulator:1.4.0"))
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
	const (
		projectId    = "test-project"
		instanceId   = "test-instance"
		databaseName = "test-db"
	)

	options := []option.ClientOption{
		option.WithEndpoint(spannerContainer.GRPCEndpoint),
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
