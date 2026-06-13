package eventhubs_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
	"github.com/testcontainers/testcontainers-go/modules/azure/eventhubs"
	"github.com/testcontainers/testcontainers-go/network"
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
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

// TestEventHubs_withAzuriteContainer verifies that when a user-supplied
// Azurite container is provided, Terminate() does NOT tear it down.
func TestEventHubs_withAzuriteContainer(t *testing.T) {
	ctx := context.Background()

	// Create an independent network and Azurite container that the test owns.
	nw, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, nw)
	require.NoError(t, err)

	azuriteCtr, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.33.0",
		network.WithNetwork([]string{"azurite"}, nw),
	)
	testcontainers.CleanupContainer(t, azuriteCtr)
	require.NoError(t, err)

	// Run the EventHubs container using the pre-built Azurite.
	ehCtr, err := eventhubs.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithAzuriteContainer(azuriteCtr, nw, "azurite"),
	)
	// Note: no CleanupContainer here — we call Terminate explicitly below to
	// test that it does NOT stop azurite, then CleanupContainer handles the rest.
	testcontainers.CleanupContainer(t, ehCtr)
	require.NoError(t, err)

	// Both containers must be on the same network.
	ehNetworks, err := ehCtr.Networks(ctx)
	require.NoError(t, err)
	require.Len(t, ehNetworks, 1)

	azNetworks, err := azuriteCtr.Networks(ctx)
	require.NoError(t, err)
	require.Len(t, azNetworks, 1)
	require.Equal(t, azNetworks[0], ehNetworks[0])

	// Terminate the EventHubs container — must NOT touch azurite.
	require.NoError(t, ehCtr.Terminate(ctx))

	// Azurite should still be running.
	state, err := azuriteCtr.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running, "azurite container must still be running after eventhubs Terminate()")
}

// TestEventHubs_withAzuriteContainer_nilGuards verifies that nil container
// and nil network both result in an error from Run.
func TestEventHubs_withAzuriteContainer_nilGuards(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, nw)
	require.NoError(t, err)

	azuriteCtr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.33.0",
		network.WithNetwork([]string{"azurite"}, nw),
	)
	testcontainers.CleanupContainer(t, azuriteCtr)
	require.NoError(t, err)

	t.Run("nil container", func(t *testing.T) {
		_, err := eventhubs.Run(
			ctx,
			"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
			eventhubs.WithAcceptEULA(),
			eventhubs.WithAzuriteContainer(nil, nw, "azurite"),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "azurite container is nil")
	})

	t.Run("nil network", func(t *testing.T) {
		_, err := eventhubs.Run(
			ctx,
			"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
			eventhubs.WithAcceptEULA(),
			eventhubs.WithAzuriteContainer(azuriteCtr, nil, "azurite"),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "docker network is nil")
	})
}

// TestEventHubs_withAzuriteContainer_conflict verifies that using both
// WithAzurite and WithAzuriteContainer returns a clear error.
func TestEventHubs_withAzuriteContainer_conflict(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, nw)
	require.NoError(t, err)

	azuriteCtr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.33.0",
		network.WithNetwork([]string{"azurite"}, nw),
	)
	testcontainers.CleanupContainer(t, azuriteCtr)
	require.NoError(t, err)

	t.Run("WithAzurite then WithAzuriteContainer", func(t *testing.T) {
		_, err := eventhubs.Run(
			ctx,
			"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
			eventhubs.WithAcceptEULA(),
			eventhubs.WithAzurite("mcr.microsoft.com/azure-storage/azurite:3.33.0"),
			eventhubs.WithAzuriteContainer(azuriteCtr, nw, "azurite"),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("WithAzuriteContainer then WithAzurite", func(t *testing.T) {
		_, err := eventhubs.Run(
			ctx,
			"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
			eventhubs.WithAcceptEULA(),
			eventhubs.WithAzuriteContainer(azuriteCtr, nw, "azurite"),
			eventhubs.WithAzurite("mcr.microsoft.com/azure-storage/azurite:3.33.0"),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "mutually exclusive")
	})
}

// TestEventHubs_withConfigObject builds a *Config via functional options, runs
// the container, copies the config file back out and checks it round-trips
// to the expected JSON structure.
func TestEventHubs_withConfigObject(t *testing.T) {
	ctx := context.Background()

	cfg, err := eventhubs.NewConfig(
		eventhubs.WithNamespace("emulatorNs1",
			eventhubs.WithEntity("eh1", 1,
				eventhubs.WithConsumerGroup("cg1"),
			),
		),
	)
	require.NoError(t, err)

	ctr, err := eventhubs.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithConfigObject(cfg),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Read the config file back from the container.
	rc, err := ctr.CopyFileFromContainer(ctx, "/Eventhubs_Emulator/ConfigFiles/Config.json")
	require.NoError(t, err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)

	// Unmarshal and compare as map[string]any (key-order insensitive).
	var gotMap, wantMap map[string]any
	require.NoError(t, json.Unmarshal(content, &gotMap))

	wantJSON, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(wantJSON, &wantMap))

	require.Equal(t, wantMap, gotMap)
}

// TestEventHubs_withConfigObject_nil verifies that WithConfigObject(nil) errors.
func TestEventHubs_withConfigObject_nil(t *testing.T) {
	ctx := context.Background()

	ctr, err := eventhubs.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithConfigObject(nil),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config is nil")
}
