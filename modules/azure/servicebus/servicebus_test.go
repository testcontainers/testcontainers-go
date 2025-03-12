package servicebus_test

import (
	"context"
	_ "embed"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/servicebus"
)

//go:embed testdata/servicebus_config.json
var servicebusConfig string

func TestServiceBus_topology(t *testing.T) {
	ctx := context.Background()

	const mssqlImage = "mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04"

	ctr, err := servicebus.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/servicebus-emulator:1.1.2",
		servicebus.WithAcceptEULA(),
		servicebus.WithMSSQL(mssqlImage, testcontainers.WithEnv(map[string]string{"TESTCONTAINERS_TEST_VAR": "test"})),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// assert that both containers belong to the same network
	serviceBusNetworks, err := ctr.Networks(ctx)
	require.NoError(t, err)
	require.Len(t, serviceBusNetworks, 1)

	mssqlContainer := ctr.MSSQLContainer()
	mssqlNetworks, err := mssqlContainer.Networks(ctx)
	require.NoError(t, err)
	require.Len(t, mssqlNetworks, 1)

	require.Equal(t, mssqlNetworks[0], serviceBusNetworks[0])

	// mssql image version and custom options
	inspect, err := mssqlContainer.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, mssqlImage, inspect.Config.Image)
	require.Contains(t, inspect.Config.Env, "TESTCONTAINERS_TEST_VAR=test")
}

func TestServiceBus_withConfig(t *testing.T) {
	ctx := context.Background()

	const mssqlImage = "mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04"

	ctr, err := servicebus.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/servicebus-emulator:1.1.2",
		servicebus.WithAcceptEULA(),
		servicebus.WithMSSQL(mssqlImage),
		servicebus.WithConfig(strings.NewReader(servicebusConfig)),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// assert that the config file was created in the right location.
	rc, err := ctr.CopyFileFromContainer(ctx, "/ServiceBus_Emulator/ConfigFiles/Config.json")
	require.NoError(t, err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, servicebusConfig, string(content))
}

func TestServiceBus_noEULA(t *testing.T) {
	ctx := context.Background()

	ctr, err := servicebus.Run(ctx, "mcr.microsoft.com/azure-messaging/servicebus-emulator:1.1.2")
	require.Error(t, err)
	require.Nil(t, ctr)
}
