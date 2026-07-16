package mosquitto_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mosquitto"
)

func TestMosquitto(t *testing.T) {
	ctx := context.Background()

	ctr, err := mosquitto.Run(ctx, "eclipse-mosquitto:2")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Verify broker URL is returned correctly.
	brokerURL, err := ctr.BrokerURL(ctx)
	require.NoError(t, err)
	require.Contains(t, brokerURL, "mqtt://")
}

func TestMosquittoWithConfigFile(t *testing.T) {
	ctx := context.Background()

	// testdata/mosquitto.conf differs from the embedded default: it adds
	// max_inflight_messages 20. Starting successfully with this config confirms
	// that WithConfigFile can provide a working configuration to the broker.
	ctr, err := mosquitto.Run(ctx, "eclipse-mosquitto:2",
		mosquitto.WithConfigFile("testdata/mosquitto.conf"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	brokerURL, err := ctr.BrokerURL(ctx)
	require.NoError(t, err)
	require.Contains(t, brokerURL, "mqtt://")
}
