package pubsub_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/testrest"
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

	requestCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 30 * time.Second}
	defer client.CloseIdleConnections()
	endpoint := "http://" + pubsubContainer.URI() + "/v1/"
	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, "greetings")
	var topic struct {
		Name string `json:"name"`
	}
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPut, endpoint+topicName, map[string]any{}, &topic))
	require.Equal(t, topicName, topic.Name)

	subName := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, "subscription")
	var subscription struct {
		Name string `json:"name"`
	}
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPut, endpoint+subName, map[string]string{"topic": topic.Name}, &subscription))
	require.Equal(t, subName, subscription.Name)

	var published struct {
		MessageIDs []string `json:"messageIds"`
	}
	// encoding/json encodes []byte as the API's base64 message data.
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+topic.Name+":publish", map[string]any{
		"messages": []any{map[string]any{"data": []byte("Hello World")}},
	}, &published))
	require.Len(t, published.MessageIDs, 1)
	require.NotEmpty(t, published.MessageIDs[0])

	type pullResult struct {
		ReceivedMessages []struct {
			AckID   string `json:"ackId"`
			Message struct {
				Data []byte `json:"data"`
			} `json:"message"`
		} `json:"receivedMessages"`
	}
	var received pullResult
	for len(received.ReceivedMessages) == 0 {
		require.NoError(t, requestCtx.Err())
		require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+subscription.Name+":pull", map[string]any{"maxMessages": 1}, &received))
	}
	require.Len(t, received.ReceivedMessages, 1)
	message := received.ReceivedMessages[0]
	require.NotEmpty(t, message.AckID)
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+subscription.Name+":acknowledge", map[string]any{"ackIds": []string{message.AckID}}, nil))
	require.Equal(t, "Hello World", string(message.Message.Data))
}
