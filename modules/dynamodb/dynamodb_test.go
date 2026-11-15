package dynamodb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

const (
	baseImage string = "amazon/dynamodb-local:"
)

var image2_2_1 = baseImage + "2.2.1"

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Client operations and errors are covered in the examples module.
	endpoint, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, endpoint)
}

func TestRun_withCustomImageVersion(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:2.2.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestRun_withInvalidCustomImageVersion(t *testing.T) {
	ctx := context.Background()

	_, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:0.0.7")
	require.Error(t, err)
}

func TestRun_withoutEndpointResolver(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err, "container should start successfully")

	// Client operations and errors are covered in the examples module.
	endpoint, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, endpoint)
}

func TestRun_withSharedDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithSharedDB())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// create a second container: it should have the table created in the first container
	ctr2, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithSharedDB())
	testcontainers.CleanupContainer(t, ctr2)
	require.NoError(t, err)

	require.Equal(t, ctr.GetContainerID(), ctr2.GetContainerID())
	for _, c := range []*tcdynamodb.DynamoDBContainer{ctr, ctr2} {
		endpoint, err := c.ConnectionString(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, endpoint)
	}
}

func TestRun_withoutSharedDB(t *testing.T) {
	ctx := context.Background()

	ctr1, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr1)
	require.NoError(t, err)

	// create a second container: it should not have the table created in the first container
	ctr2, err := tcdynamodb.Run(ctx, image2_2_1)
	testcontainers.CleanupContainer(t, ctr2)
	require.NoError(t, err)

	require.NotEqual(t, ctr1.GetContainerID(), ctr2.GetContainerID())
	for _, c := range []*tcdynamodb.DynamoDBContainer{ctr1, ctr2} {
		endpoint, err := c.ConnectionString(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, endpoint)
	}
}

func TestRun_shouldStartWithTelemetryDisabled(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithDisableTelemetry())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestWithInMemory(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Cmd: []string{"-jar", "DynamoDBLocal.jar"},
		},
	}

	require.NoError(t, tcdynamodb.WithInMemory().Customize(&req))
	require.Equal(t, []string{"-jar", "DynamoDBLocal.jar", "-inMemory"}, req.Cmd)
}

func TestRun_shouldStartWithInMemory(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithInMemory())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	inspect, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	require.Contains(t, inspect.Config.Cmd, "-inMemory")
}

func TestRun_shouldStartWithSharedDBEnabledAndTelemetryDisabled(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, image2_2_1, tcdynamodb.WithSharedDB(), tcdynamodb.WithDisableTelemetry())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}
