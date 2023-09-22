package gcloud_test

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func ExampleRunPubsubContainer() {
	// runPubsubContainer {
	ctx := context.Background()

	pubsubContainer, err := gcloud.RunPubsubContainer(
		ctx,
		testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"),
		gcloud.WithProjectID("pubsub-project"),
	)
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
	projectID := pubsubContainer.Settings.ProjectID

	conn, err := grpc.Dial(pubsubContainer.URI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := pubsub.NewClient(ctx, projectID, options...)
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
