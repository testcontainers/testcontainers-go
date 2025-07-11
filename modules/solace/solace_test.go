package solace_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	solacecontainer "github.com/testcontainers/testcontainers-go/modules/solace"
)

func TestSolace(t *testing.T) {
	ctx := context.Background()

	queueName := "TestQueue"
	topicName := "Topic/ActualTopic"

	solaceC := solacecontainer.NewSolaceContainer(ctx, "solace-pubsub-standard:latest").
		WithCredentials("admin", "admin").
		WithVpn("test-vpn").
		WithExposedPorts("5672/tcp", "8080/tcp").
		WithEnv(map[string]string{
			"username_admin_globalaccesslevel": "admin",
			"username_admin_password":          "admin",
		}).
		WithShmSize(1<<30).
		WithQueue(queueName, topicName)

	fmt.Println("Executing")
	err := solaceC.Run(ctx)
	testcontainers.CleanupContainer(t, solaceC.Container)
	require.NoError(t, err)

	// Assert container is running
	state, err := solaceC.Container.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	// Assert service origin URL is accessible (format check)
	origin, err := solaceC.BrokerURLFor(solacecontainer.ServiceAMQP)
	require.NoError(t, err)
	require.Contains(t, origin, "amqp://")

	// Wait a bit for the broker to be fully ready
	time.Sleep(2 * time.Second)

	// Test message publishing and consumption using Solace SDK
	err = testMessagePublishAndConsume(solaceC, queueName, topicName)
	require.NoError(t, err, "Message publish and consume test should pass")
}
