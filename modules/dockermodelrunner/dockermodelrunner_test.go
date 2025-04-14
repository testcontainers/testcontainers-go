package dockermodelrunner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner"
)

func TestDockerModelRunner(t *testing.T) {
	ctx := context.Background()

	ctr, err := dockermodelrunner.Run(ctx, "alpine/socat:1.8.0.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}

func TestRun_client(t *testing.T) {
	ctx := context.Background()

	ctr, err := dockermodelrunner.Run(ctx, "alpine/socat:1.8.0.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("model-pull", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := ctr.PullModel(ctx, "ai/llama3.2:latest")
			require.NoError(t, err)
		})

		t.Run("failure", func(t *testing.T) {
			err := ctr.PullModel(ctx, "ai/non-existent:latest")
			require.Error(t, err)
		})
	})
}
