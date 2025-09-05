package dockermcpgateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithTools(t *testing.T) {
	t.Run("one-server", func(t *testing.T) {
		t.Run("single-tool", func(t *testing.T) {
			settings := defaultOptions()

			err := WithTools("server", []string{"tool1"})(&settings)
			require.NoError(t, err)
			require.Len(t, settings.tools, 1)
			require.Equal(t, []string{"tool1"}, settings.tools["server"])
		})

		t.Run("multiple-tools", func(t *testing.T) {
			settings := defaultOptions()

			err := WithTools("server", []string{"tool1", "tool2"})(&settings)
			require.NoError(t, err)
			require.Len(t, settings.tools, 1)
			require.Contains(t, settings.tools["server"], "tool1")
			require.Contains(t, settings.tools["server"], "tool2")
		})
	})

	t.Run("empty-server", func(t *testing.T) {
		settings := defaultOptions()

		err := WithTools("", []string{"tool"})(&settings)
		require.ErrorContains(t, err, "server cannot be empty")
	})

	t.Run("empty-tools", func(t *testing.T) {
		settings := defaultOptions()

		err := WithTools("server", nil)(&settings)
		require.ErrorContains(t, err, "tools cannot be empty")
	})

	t.Run("empty-tool-in-slice", func(t *testing.T) {
		settings := defaultOptions()

		err := WithTools("server", []string{"tool1", "", "tool3"})(&settings)
		require.ErrorContains(t, err, "tool cannot be empty")
	})

	t.Run("duplicated-tools", func(t *testing.T) {
		settings := defaultOptions()

		err := WithTools("server", []string{"tool1"})(&settings)
		require.NoError(t, err)

		err = WithTools("server", []string{"tool1"})(&settings)
		require.NoError(t, err)

		require.Len(t, settings.tools, 1)
		require.Equal(t, []string{"tool1"}, settings.tools["server"])
	})

	t.Run("duplicated-server", func(t *testing.T) {
		settings := defaultOptions()

		err := WithTools("server", []string{"tool1"})(&settings)
		require.NoError(t, err)

		err = WithTools("server", []string{"tool2"})(&settings)
		require.NoError(t, err)

		require.Len(t, settings.tools, 1)
		require.Equal(t, []string{"tool1", "tool2"}, settings.tools["server"])
	})

	t.Run("multiple-servers", func(t *testing.T) {
		settings := defaultOptions()

		err := WithTools("server1", []string{"tool1.1"})(&settings)
		require.NoError(t, err)
		err = WithTools("server2", []string{"tool2.1", "tool2.2"})(&settings)
		require.NoError(t, err)

		require.Len(t, settings.tools, 2)
		require.Contains(t, settings.tools["server1"], "tool1.1")
		require.Contains(t, settings.tools["server2"], "tool2.1")
		require.Contains(t, settings.tools["server2"], "tool2.2")
	})
}

func TestWithSecrets(t *testing.T) {
	t.Run("single-secret", func(t *testing.T) {
		settings := defaultOptions()

		err := WithSecret("key", "value")(&settings)
		require.NoError(t, err)
		require.Contains(t, settings.secrets, "key")
		require.Equal(t, "value", settings.secrets["key"])
	})

	t.Run("multiple-secrets", func(t *testing.T) {
		settings := defaultOptions()

		err := WithSecrets(map[string]string{
			"key1": "value1",
			"key2": "value2",
		})(&settings)
		require.NoError(t, err)
		require.Len(t, settings.secrets, 2)
		require.Equal(t, "value1", settings.secrets["key1"])
		require.Equal(t, "value2", settings.secrets["key2"])
	})

	t.Run("empty-key", func(t *testing.T) {
		settings := defaultOptions()

		err := WithSecret("", "value")(&settings)
		require.ErrorContains(t, err, "secret key cannot be empty")
	})

	t.Run("empty-value", func(t *testing.T) {
		settings := defaultOptions()

		err := WithSecret("key", "")(&settings)
		require.NoError(t, err)
		require.Empty(t, settings.secrets["key"])
	})
}
