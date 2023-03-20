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

	container, err := StartContainer(
		ctx,
		OverrideContainerRequest(testcontainers.ContainerRequest{
			Image:      "localstack/localstack:0.10.0",
			Env:        map[string]string{"SERVICES": "s3,sqs"},
			WaitingFor: wait.ForLog("Ready.").WithStartupTimeout(5 * time.Minute).WithOccurrence(1),
		}),
	)
	require.NotNil(t, err)
	assert.Nil(t, container)
}
