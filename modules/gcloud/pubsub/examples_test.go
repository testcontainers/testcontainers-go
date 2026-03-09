package pubsub_test

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcpubsub "github.com/testcontainers/testcontainers-go/modules/gcloud/pubsub"
)

func ExampleRun() {
	// runPubsubContainer {
	ctx := context.Background()

	pubsubContainer, err := tcpubsub.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		tcpubsub.WithProjectID("pubsub-project"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(pubsubContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run container: %v", err)
		return
	}
	// }

	// pubsubClient {
	projectID := pubsubContainer.ProjectID()

	conn, err := grpc.NewClient(pubsubContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to dial: %v", err)
		return
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := pubsub.NewClient(ctx, projectID, options...)
	if err != nil {
		log.Printf("failed to create client: %v", err)
		return
	}
	defer client.Close()
	// }

	topic, err := client.CreateTopic(ctx, "greetings")
	if err != nil {
		log.Printf("failed to create topic: %v", err)
		return
	}
	subscription, err := client.CreateSubscription(ctx, "subscription",
		pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		log.Printf("failed to create subscription: %v", err)
		return
	}
	result := topic.Publish(ctx, &pubsub.Message{Data: []byte("Hello World")})
	_, err = result.Get(ctx)
	if err != nil {
		log.Printf("failed to publish message: %v", err)
		return
	}

	var data []byte
	cctx, cancel := context.WithCancel(ctx)
	err = subscription.Receive(cctx, func(_ context.Context, m *pubsub.Message) {
		data = m.Data
		m.Ack()
		defer cancel()
	})
	if err != nil {
		log.Printf("failed to receive message: %v", err)
		return
	}

	fmt.Println(string(data))

	// Output:
	// Hello World
}
