package pubsub_test

import (
	"context"
	"log"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcpubsub "github.com/testcontainers/testcontainers-go/modules/gcloud/pubsub"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	pubsubContainer, err := tcpubsub.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		tcpubsub.WithProjectID("pubsub-project"),
	)
	testcontainers.CleanupContainer(t, pubsubContainer)
	require.NoError(t, err)

	projectID := pubsubContainer.ProjectID()

	conn, err := grpc.NewClient(pubsubContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to dial: %v", err)
		return
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := pubsub.NewClient(ctx, projectID, options...)
	require.NoError(t, err)
	defer client.Close()

	topic, err := client.CreateTopic(ctx, "greetings")
	require.NoError(t, err)

	subscription, err := client.CreateSubscription(ctx, "subscription",
		pubsub.SubscriptionConfig{Topic: topic})
	require.NoError(t, err)

	result := topic.Publish(ctx, &pubsub.Message{Data: []byte("Hello World")})
	_, err = result.Get(ctx)
	require.NoError(t, err)

	var data []byte
	cctx, cancel := context.WithCancel(ctx)
	err = subscription.Receive(cctx, func(_ context.Context, m *pubsub.Message) {
		data = m.Data
		m.Ack()
		defer cancel()
	})
	require.NoError(t, err)

	require.Equal(t, "Hello World", string(data))
}
