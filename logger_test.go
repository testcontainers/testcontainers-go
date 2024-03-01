package testcontainers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithLogger(t *testing.T) {
	logger := TestLogger(t)
	logOpt := WithLogger(logger)
	t.Run("container", func(t *testing.T) {
		var req GenericContainerRequest
		logOpt.Customize(&req)
		require.Equal(t, logger, req.Logger)
	})

	t.Run("provider", func(t *testing.T) {
		var opts GenericProviderOptions
		logOpt.ApplyGenericTo(&opts)
		require.Equal(t, logger, opts.Logger)
	})

	t.Run("docker", func(t *testing.T) {
		opts := &DockerProviderOptions{
			GenericProviderOptions: &GenericProviderOptions{},
		}
		logOpt.ApplyDockerTo(opts)
		require.Equal(t, logger, opts.Logger)
	})
}
