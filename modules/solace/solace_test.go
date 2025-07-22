package solace_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	sc "github.com/testcontainers/testcontainers-go/modules/solace"
)

func TestSolace(t *testing.T) {
	ctx := context.Background()

	queueName := "TestQueue"
	topicName := "Topic/ActualTopic"

	ctr, err := sc.Run(ctx, "solace-pubsub-standard:latest",
		sc.WithCredentials("admin", "admin"),
		sc.WithVpn("test-vpn"),
		sc.WithExposedPorts("5672/tcp", "8080/tcp"),
		sc.WithEnv(map[string]string{
			"username_admin_globalaccesslevel": "admin",
			"username_admin_password":          "admin",
		}),
		sc.WithShmSize(1<<30),
		sc.WithQueue(queueName, topicName),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	fmt.Println("Executing")
	require.NoError(t, err)

	// Assert container is running
	state, err := ctr.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	// Assert service origin URL is accessible (format check)
	origin, err := ctr.BrokerURLFor(ctx, sc.ServiceAMQP)
	require.NoError(t, err)
	require.Contains(t, origin, "amqp://")

	// Wait a bit for the broker to be fully ready
	time.Sleep(2 * time.Second)

	// Test message publishing and consumption using Solace SDK
	err = testMessagePublishAndConsume(ctr, queueName, topicName)
	require.NoError(t, err, "Message publish and consume test should pass")
}
