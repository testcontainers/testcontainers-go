package trinogateway_test

import (
	"context"
	_ "embed"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/trino"
	"github.com/testcontainers/testcontainers-go/modules/trinogateway"
	"github.com/testcontainers/testcontainers-go/network"
)

const (
	gatewayImage  = "trinodb/trino-gateway:22"
	trinoImage    = "trinodb/trino:435"
	postgresImage = "postgres:16-alpine"
)

//go:embed testdata/config.yaml
var customConfig string

func TestTrinoGateway(t *testing.T) {
	ctx := context.Background()

	ctr, err := trinogateway.Run(ctx, gatewayImage,
		trinogateway.WithPostgres(postgresImage, testcontainers.WithEnv(map[string]string{"TESTCONTAINERS_TEST_VAR": "test"})),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(connStr, "http://"), "unexpected connection string %q", connStr)

	// the gateway and the postgres containers belong to the module network
	require.NotNil(t, ctr.Network())

	gatewayNetworks, err := ctr.Networks(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{ctr.Network().Name}, gatewayNetworks)

	pg := ctr.PostgresContainer()
	require.NotNil(t, pg)

	postgresNetworks, err := pg.Networks(ctx)
	require.NoError(t, err)
	require.Equal(t, gatewayNetworks, postgresNetworks)

	// postgres image version and custom options
	inspect, err := pg.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, postgresImage, inspect.Config.Image)
	require.Contains(t, inspect.Config.Env, "TESTCONTAINERS_TEST_VAR=test")

	// the data store is reachable: the gateway starts with no backends
	backends, err := ctr.Backends(ctx)
	require.NoError(t, err)
	require.Empty(t, backends)

	// terminating the gateway removes the postgres container and the network too
	require.NoError(t, ctr.Terminate(ctx))

	_, err = pg.State(ctx)
	require.Error(t, err)
}

func TestTrinoGateway_backends(t *testing.T) {
	ctx := context.Background()

	ctr, err := trinogateway.Run(ctx, gatewayImage)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Trino must process the forwarded headers set by the gateway, so that the URLs
	// for the next results of a query point to the gateway instead of the Trino container.
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

	// queries sent to the gateway are routed to the backend
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

func TestTrinoGateway_withDataStore(t *testing.T) {
	ctx := context.Background()

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
		trinogateway.WithDataStore("jdbc:postgresql://db:5432/gateway", "gateway", "secret", "org.postgresql.Driver"),
		network.WithNetwork([]string{"gateway"}, nw),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// no postgres container is started by the module
	require.Nil(t, ctr.PostgresContainer())

	// the gateway is attached to the module network and to the external one
	networks, err := ctr.Networks(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{ctr.Network().Name, nw.Name}, networks)

	backends, err := ctr.Backends(ctx)
	require.NoError(t, err)
	require.Empty(t, backends)
}

func TestTrinoGateway_withConfig(t *testing.T) {
	ctx := context.Background()

	ctr, err := trinogateway.Run(ctx, gatewayImage, trinogateway.WithConfig(strings.NewReader(customConfig)))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// the custom config file was copied to the right location
	rc, err := ctr.CopyFileFromContainer(ctx, "/etc/trino-gateway/config.yaml")
	require.NoError(t, err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, customConfig, string(content))

	// the data store placeholders were resolved from the environment variables set by the module
	backends, err := ctr.Backends(ctx)
	require.NoError(t, err)
	require.Empty(t, backends)
}

func TestTrinoGateway_invalidOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("empty-data-store-url", func(t *testing.T) {
		ctr, err := trinogateway.Run(ctx, gatewayImage, trinogateway.WithDataStore("", "user", "password", "org.postgresql.Driver"))
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
		require.Nil(t, ctr)
	})

	t.Run("nil-config", func(t *testing.T) {
		ctr, err := trinogateway.Run(ctx, gatewayImage, trinogateway.WithConfig(nil))
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
		require.Nil(t, ctr)
	})
}
