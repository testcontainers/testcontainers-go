package pubsub_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	pubsubpb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
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
		"gcr.io/google.com/cloudsdktool/cloud-sdk:emulators",
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

	topicProto, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: fmt.Sprintf("projects/%s/topics/%s", projectID, "greetings"),
	})
	require.NoError(t, err)

	subProto, err := client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:  fmt.Sprintf("projects/%s/subscriptions/%s", projectID, "subscription"),
		Topic: topicProto.GetName(),
	})
	require.NoError(t, err)

	publisher := client.Publisher(topicProto.GetName())
	defer publisher.Stop()
	result := publisher.Publish(ctx, &pubsub.Message{Data: []byte("Hello World")})
	_, err = result.Get(ctx)
	require.NoError(t, err)

	sub := client.Subscriber(subProto.GetName())
	var data []byte
	cctx, cancel := context.WithCancel(ctx)
	err = sub.Receive(cctx, func(_ context.Context, m *pubsub.Message) {
		data = m.Data
		m.Ack()
		defer cancel()
	})
	require.NoError(t, err)

	require.Equal(t, "Hello World", string(data))
}
