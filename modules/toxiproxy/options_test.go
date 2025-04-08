package toxiproxy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithPortRange(t *testing.T) {
	t.Run("positive-port", func(t *testing.T) {
		portsCount := 1

		opt := WithPortRange(portsCount)

		var opts options
		err := opt(&opts)
		require.NoError(t, err)

		require.Equal(t, portsCount, opts.portRange)
	})

	t.Run("negative-port", func(t *testing.T) {
		portsCount := -1

		opt := WithPortRange(portsCount)

		var opts options
		err := opt(&opts)
		require.Error(t, err)
	})

	t.Run("zero-port", func(t *testing.T) {
		portsCount := 0

		opt := WithPortRange(portsCount)

		var opts options
		err := opt(&opts)
		require.Error(t, err)
	})
}
