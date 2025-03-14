package eventhubs_test

import (
	"context"
	_ "embed"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/eventhubs"
)

//go:embed testdata/eventhubs_config.json
var eventhubsConfig string

func TestEventHubs_topology(t *testing.T) {
	ctx := context.Background()

	const azuriteImage = "mcr.microsoft.com/azure-storage/azurite:3.33.0"

	ctr, err := eventhubs.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithAzurite(azuriteImage, testcontainers.WithEnv(map[string]string{"TESTCONTAINERS_TEST_VAR": "test"})),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// assert that both containers belong to the same network
	eventHubsNetworks, err := ctr.Networks(ctx)
	require.NoError(t, err)
	require.Len(t, eventHubsNetworks, 1)

	azuriteContainer := ctr.AzuriteContainer()
	azuriteNetworks, err := azuriteContainer.Networks(ctx)
	require.NoError(t, err)
	require.Len(t, azuriteNetworks, 1)

	require.Equal(t, azuriteNetworks[0], eventHubsNetworks[0])

	// azurite image version and custom options
	inspect, err := azuriteContainer.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, azuriteImage, inspect.Config.Image)
	require.Contains(t, inspect.Config.Env, "TESTCONTAINERS_TEST_VAR=test")
}

func TestEventHubs_withConfig(t *testing.T) {
	ctx := context.Background()

	const azuriteImage = "mcr.microsoft.com/azure-storage/azurite:3.33.0"

	ctr, err := eventhubs.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithAzurite(azuriteImage),
		eventhubs.WithConfig(strings.NewReader(eventhubsConfig)),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// assert that the config file was created in the right location.
	rc, err := ctr.CopyFileFromContainer(ctx, "/Eventhubs_Emulator/ConfigFiles/Config.json")
	require.NoError(t, err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, eventhubsConfig, string(content))
}

func TestEventHubs_noEULA(t *testing.T) {
	ctx := context.Background()

	ctr, err := eventhubs.Run(ctx, "mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0")
	require.Error(t, err)
	require.Nil(t, ctr)
}
