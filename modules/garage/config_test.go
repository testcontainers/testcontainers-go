package garage

import (
	"encoding/hex"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

// TestGarageConfig verifies that arbitrary region values cannot inject TOML fields or tables.
func TestGarageConfig(t *testing.T) {
	for _, region := range []string{"garage", "eu-west-1", "quote\"and\\slash", "region\"\n[admin]\napi_bind_addr = \"0.0.0.0:3903", "control\x7f\t\n日本語"} {
		t.Run(region, func(t *testing.T) {
			data, err := garageConfig(region)
			require.NoError(t, err)
			var config map[string]any
			require.NoError(t, toml.Unmarshal(data, &config))
			require.Len(t, config, 8)
			require.NotContains(t, config, "admin")
			require.Equal(t, int64(1), config["replication_factor"])
			require.Equal(t, map[string]any{"s3_region": region, "api_bind_addr": "[::]:3900"}, config["s3_api"])
			secret, ok := config["rpc_secret"].(string)
			require.True(t, ok)
			decoded, err := hex.DecodeString(secret)
			require.NoError(t, err)
			require.Len(t, decoded, 32)
		})
	}
}
