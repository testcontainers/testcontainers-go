package socat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTarget(t *testing.T) {
	t.Run("exposed-port", func(t *testing.T) {
		target := NewTarget(8080, "helloworld")
		require.Equal(t, 8080, target.exposedPort)
		require.Equal(t, 8080, target.internalPort)
		require.Equal(t, "helloworld", target.host)
	})

	t.Run("exposed-port-zero", func(t *testing.T) {
		target := NewTarget(0, "helloworld")
		require.Equal(t, 0, target.exposedPort)
		require.Equal(t, 0, target.internalPort)

		opts := options{}

		err := WithTarget(target)(&opts)
		require.Error(t, err)
	})

	t.Run("with-internal-port", func(t *testing.T) {
		target := NewTargetWithInternalPort(8080, 8081, "helloworld")
		require.Equal(t, 8080, target.exposedPort)
		require.Equal(t, 8081, target.internalPort)
	})

	t.Run("with-internal-port-zero", func(t *testing.T) {
		target := NewTargetWithInternalPort(8080, 0, "helloworld")
		require.Equal(t, 8080, target.exposedPort)
		require.Equal(t, 8080, target.internalPort)
	})
}
