package localstack

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestLegacyMode(t *testing.T) {
	ctx := context.Background()

	// withoutNetwork {
	container, err := StartContainer(
		ctx,
		OverrideContainerRequest(testcontainers.ContainerRequest{
			Image:      "localstack/localstack:0.11.0",
			WaitingFor: wait.ForLog("Ready.").WithStartupTimeout(5 * time.Minute).WithOccurrence(1),
		}),
		WithLegacyMode,
		WithServices(S3, SQS),
	)
	require.Nil(t, err)
	assert.NotNil(t, container)
	// }

	t.Run("multiple services should be exposed using their own port", func(t *testing.T) {

		ports, err := container.Ports(ctx)
		require.Nil(t, err)
		assert.Greater(t, len(ports), 1) // multiple ports are exposed

		s3Endpoint, err := container.ServicePort(ctx, S3)
		require.Nil(t, err)
		sqsEndpoint, err := container.ServicePort(ctx, SQS)
		require.Nil(t, err)

		assert.NotEqual(t, sqsEndpoint, s3Endpoint)
	})
}

func TestLegacyModeForVersionGreaterThan_0_11(t *testing.T) {
	// forceLegacyMode {
	ctx := context.Background()

	container, err := StartContainer(
		ctx,
		OverrideContainerRequest(testcontainers.ContainerRequest{
			Image: "localstack/localstack:0.12.0",
		}),
		WithServices(S3, SQS),
		WithLegacyMode,
	)
	// }

	t.Run("multiple services should be exposed using the same port", func(t *testing.T) {
		require.Nil(t, err)
		assert.NotNil(t, container)

		rawPorts, err := container.Ports(ctx)
		require.Nil(t, err)

		ports := 0
		// only one port is exposed among all the ports in the container
		for _, v := range rawPorts {
			if len(v) > 0 {
				ports++
			}
		}

		assert.Equal(t, 1, ports) // a single port is exposed

		s3Endpoint, err := container.ServicePort(ctx, S3)
		require.Nil(t, err)

		sqsEndpoint, err := container.ServicePort(ctx, SQS)
		require.Nil(t, err)

		assert.Equal(t, sqsEndpoint, s3Endpoint)
	})
}
