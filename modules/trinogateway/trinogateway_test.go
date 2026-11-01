package trinogateway_test

import (
	"context"
	_ "embed"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/trino"
	"github.com/testcontainers/testcontainers-go/modules/trinogateway"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	gatewayImage  = "trinodb/trino-gateway:22"
	trinoImage    = "trinodb/trino:435"
	postgresImage = "postgres:17-alpine"
)

//go:embed testdata/config.yaml
var customConfig string

func TestTrinoGateway(t *testing.T) {
	ctx := context.Background()

	provider := trinogateway.NewManagedPostgresProvider(postgresImage,
		testcontainers.WithEnv(map[string]string{"TESTCONTAINERS_TEST_VAR": "test"}),
	)

	ctr, err := trinogateway.Run(ctx, gatewayImage, trinogateway.WithDataStore(provider))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(connStr, "http://"), "unexpected connection string %q", connStr)

	// gateway and postgres share the gateway network
	nw := ctr.Network()
	require.NotNil(t, nw)

	gatewayNetworks, err := ctr.Networks(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{nw.Name}, gatewayNetworks)

	pg := provider.Container()
	require.NotNil(t, pg)

	postgresNetworks, err := pg.Networks(ctx)
	require.NoError(t, err)
	require.Equal(t, gatewayNetworks, postgresNetworks)

	// postgres image and custom options
	inspect, err := pg.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, postgresImage, inspect.Config.Image)
	require.Contains(t, inspect.Config.Env, "TESTCONTAINERS_TEST_VAR=test")

	// the data store is reachable
	backends, err := ctr.Backends(ctx)
	require.NoError(t, err)
	require.Empty(t, backends)

	// a provider cannot be used by two gateways
	other, err := trinogateway.Run(ctx, gatewayImage, trinogateway.WithDataStore(provider))
	testcontainers.CleanupContainer(t, other)
	require.Error(t, err)
	require.Nil(t, other)

	// terminate removes the gateway, postgres and the network
	require.NoError(t, ctr.Terminate(ctx))

	requireContainerRemoved(t, ctr.GetContainerID())
	requireContainerRemoved(t, pg.GetContainerID())
	requireNetworkRemoved(t, nw.ID)
	require.Nil(t, provider.Container())
}

func TestTrinoGateway_backends(t *testing.T) {
	ctx := context.Background()

	ctr, err := trinogateway.Run(ctx, gatewayImage,
		trinogateway.WithDataStore(trinogateway.NewManagedPostgresProvider(postgresImage)),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Trino must process the forwarded headers, so that nextUri points to the gateway
	trinoCtr, err := trino.Run(ctx, trinoImage,
		network.WithNetwork([]string{"trino"}, ctr.Network()),
		testcontainers.WithEnv(map[string]string{"JAVA_TOOL_OPTIONS": "-Dhttp-server.process-forwarded=true"}),
	)
	testcontainers.CleanupContainer(t, trinoCtr)
	require.NoError(t, err)

	err = ctr.AddBackend(ctx, trinogateway.Backend{Name: "trino", ProxyTo: "http://trino:8080"})
	require.NoError(t, err)

	waitCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	require.NoError(t, ctr.WaitForBackendHealthy(waitCtx, "trino"))

	state, err := ctr.BackendState(ctx, "trino")
	require.NoError(t, err)
	require.True(t, state.Healthy())
	require.Equal(t, "http://trino:8080", state.ProxyTo)
	require.Equal(t, "adhoc", state.RoutingGroup)

	backends, err := ctr.Backends(ctx)
	require.NoError(t, err)
	require.Len(t, backends, 1)
	require.Equal(t, "trino", backends[0].Name)
	require.Equal(t, "http://trino:8080", backends[0].ProxyTo)
	require.Equal(t, "adhoc", backends[0].RoutingGroup)
	require.True(t, backends[0].Active)

	// queries are routed to the backend
	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	rows, err := runQuery(ctx, connStr, "SELECT 1")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.EqualValues(t, 1, rows[0][0])

	// no queries are routed to a deactivated backend
	require.NoError(t, ctr.DeactivateBackend(ctx, "trino"))

	active, err := ctr.ActiveBackends(ctx)
	require.NoError(t, err)
	require.Empty(t, active)

	_, err = runQuery(ctx, connStr, "SELECT 1")
	require.Error(t, err)

	require.NoError(t, ctr.ActivateBackend(ctx, "trino"))

	active, err = ctr.ActiveBackends(ctx)
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.Equal(t, "trino", active[0].Name)

	require.NoError(t, ctr.DeleteBackend(ctx, "trino"))

	backends, err = ctr.Backends(ctx)
	require.NoError(t, err)
	require.Empty(t, backends)

	_, err = ctr.BackendState(ctx, "trino")
	require.Error(t, err)
}

func TestTrinoGateway_externalDataStore(t *testing.T) {
	ctx := context.Background()

	// the database is owned by the test
	nw, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, nw)
	require.NoError(t, err)

	pg, err := postgres.Run(ctx, postgresImage,
		postgres.WithDatabase("gateway"),
		postgres.WithUsername("gateway"),
		postgres.WithPassword("secret"),
		postgres.BasicWaitStrategies(),
		network.WithNetwork([]string{"db"}, nw),
	)
	testcontainers.CleanupContainer(t, pg)
	require.NoError(t, err)

	ctr, err := trinogateway.Run(ctx, gatewayImage,
		trinogateway.WithDataStore(trinogateway.NewExternalProvider("jdbc:postgresql://db:5432/gateway", "gateway", "secret", "org.postgresql.Driver")),
		network.WithNetwork([]string{"gateway"}, nw),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// the gateway joins the database network
	networks, err := ctr.Networks(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{ctr.Network().Name, nw.Name}, networks)

	// the gateway writes the backend to the database
	require.NoError(t, ctr.AddBackend(ctx, trinogateway.Backend{Name: "trino", ProxyTo: "http://trino:8080"}))

	// terminate removes the gateway and its network
	require.NoError(t, ctr.Terminate(ctx))

	requireContainerRemoved(t, ctr.GetContainerID())
	requireNetworkRemoved(t, ctr.Network().ID)

	// the database, its network and its data are preserved
	state, err := pg.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	postgresNetworks, err := pg.Networks(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{nw.Name}, postgresNetworks)

	code, out, err := pg.Exec(ctx, []string{"psql", "-U", "gateway", "-d", "gateway", "-tAc", "SELECT name FROM gateway_backend"}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Zero(t, code)

	names, err := io.ReadAll(out)
	require.NoError(t, err)
	require.Equal(t, "trino", strings.TrimSpace(string(names)))
}

func TestTrinoGateway_dataStoreStartFailure(t *testing.T) {
	ctx := context.Background()

	// the postgres container is created, but never becomes ready
	var pgID string
	provider := trinogateway.NewManagedPostgresProvider(postgresImage,
		testcontainers.WithAdditionalLifecycleHooks(testcontainers.ContainerLifecycleHooks{
			PostCreates: []testcontainers.ContainerHook{
				func(_ context.Context, c testcontainers.Container) error {
					pgID = c.GetContainerID()
					return nil
				},
			},
		}),
		testcontainers.WithWaitStrategyAndDeadline(5*time.Second, wait.ForLog("this log line is never written")),
	)

	ctr, err := trinogateway.Run(ctx, gatewayImage, trinogateway.WithDataStore(provider))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)

	// the provider removed postgres, and no gateway container was created
	require.NotEmpty(t, pgID)
	requireContainerRemoved(t, pgID)
	require.Nil(t, provider.Container())

	require.NotNil(t, ctr)
	require.Nil(t, ctr.Container)

	// terminate removes the network
	require.NoError(t, ctr.Terminate(ctx))
	requireNetworkRemoved(t, ctr.Network().ID)
}

func TestTrinoGateway_gatewayStartFailure(t *testing.T) {
	ctx := context.Background()

	provider := trinogateway.NewManagedPostgresProvider(postgresImage)

	// the data store starts, but the gateway never becomes ready
	ctr, err := trinogateway.Run(ctx, gatewayImage,
		trinogateway.WithDataStore(provider),
		testcontainers.WithWaitStrategyAndDeadline(5*time.Second, wait.ForLog("this log line is never written")),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)

	require.NotNil(t, ctr)
	require.NotNil(t, ctr.Container)

	pg := provider.Container()
	require.NotNil(t, pg)

	// terminate removes the gateway, postgres and the network
	require.NoError(t, ctr.Terminate(ctx))

	requireContainerRemoved(t, ctr.GetContainerID())
	requireContainerRemoved(t, pg.GetContainerID())
	requireNetworkRemoved(t, ctr.Network().ID)
}

func TestTrinoGateway_withConfig(t *testing.T) {
	ctx := context.Background()

	ctr, err := trinogateway.Run(ctx, gatewayImage,
		trinogateway.WithDataStore(trinogateway.NewManagedPostgresProvider(postgresImage)),
		trinogateway.WithConfig(strings.NewReader(customConfig)),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// the custom config file was copied to the right location
	rc, err := ctr.CopyFileFromContainer(ctx, "/etc/trino-gateway/config.yaml")
	require.NoError(t, err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, customConfig, string(content))

	// the data store placeholders were resolved
	backends, err := ctr.Backends(ctx)
	require.NoError(t, err)
	require.Empty(t, backends)
}

func TestTrinoGateway_invalidOptions(t *testing.T) {
	ctx := context.Background()

	tests := map[string][]testcontainers.ContainerCustomizer{
		"no-data-store":        {},
		"nil-data-store":       {trinogateway.WithDataStore(nil)},
		"nil-external":         {trinogateway.WithDataStore((*trinogateway.ExternalProvider)(nil))},
		"external-no-jdbc-url": {trinogateway.WithDataStore(trinogateway.NewExternalProvider("", "gateway", "secret", "org.postgresql.Driver"))},
		"external-no-driver":   {trinogateway.WithDataStore(trinogateway.NewExternalProvider("jdbc:postgresql://db:5432/gateway", "gateway", "secret", ""))},
		"nil-managed":          {trinogateway.WithDataStore((*trinogateway.ManagedPostgresProvider)(nil))},
		"managed-no-image":     {trinogateway.WithDataStore(trinogateway.NewManagedPostgresProvider(""))},
		"nil-config": {
			trinogateway.WithDataStore(trinogateway.NewManagedPostgresProvider(postgresImage)),
			trinogateway.WithConfig(nil),
		},
	}

	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			// nothing is created for invalid options
			ctr, err := trinogateway.Run(ctx, gatewayImage, opts...)
			testcontainers.CleanupContainer(t, ctr)
			require.Error(t, err)
			require.Nil(t, ctr)
		})
	}
}

// requireContainerRemoved checks that the container no longer exists.
func requireContainerRemoved(t *testing.T, id string) {
	t.Helper()

	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)
	defer cli.Close()

	_, err = cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
	require.True(t, errdefs.IsNotFound(err), "container %s still exists: %v", id, err)
}

// requireNetworkRemoved checks that the network no longer exists.
func requireNetworkRemoved(t *testing.T, id string) {
	t.Helper()

	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)
	defer cli.Close()

	_, err = cli.NetworkInspect(context.Background(), id, client.NetworkInspectOptions{})
	require.True(t, errdefs.IsNotFound(err), "network %s still exists: %v", id, err)
}
