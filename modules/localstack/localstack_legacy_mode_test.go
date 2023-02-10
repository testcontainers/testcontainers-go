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
			Image:      "localstack/localstack:0.10.0",
			WaitingFor: wait.ForLog("Ready.").WithStartupTimeout(5 * time.Minute).WithOccurrence(1),
		}),
		WithServices(S3, SQS),
	)
	require.NotNil(t, err)
	assert.Nil(t, container)
	// }
}
