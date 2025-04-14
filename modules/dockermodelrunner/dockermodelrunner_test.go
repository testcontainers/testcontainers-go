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
