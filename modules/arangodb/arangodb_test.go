package arangodb_test

import (
	"context"
	"testing"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcarangodb "github.com/testcontainers/testcontainers-go/modules/arangodb"
)

func TestArangoDB(t *testing.T) {
	ctx := context.Background()

	const password = "t3stc0ntain3rs!"

	ctr, err := tcarangodb.Run(ctx, "arangodb:3.11.5", tcarangodb.WithRootPassword(password))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	httpAddress, err := ctr.HTTPEndpoint(ctx)
	require.NoError(t, err)

	endpoint := connection.NewRoundRobinEndpoints([]string{httpAddress})
	conn := connection.NewHttp2Connection(connection.DefaultHTTP2ConfigurationWrapper(endpoint, true))

	auth := connection.NewBasicAuth(ctr.Credentials())
	err = conn.SetAuthentication(auth)
	require.NoError(t, err)

	client := arangodb.NewClient(conn)

	versionInfo, err := client.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, "arango", versionInfo.Server)
}
