package aws

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

const testImage = "ghcr.io/xsavikx/aws-emulator:1.0.509.1-20261118.43"

func TestRun(t *testing.T) {
	ctx := context.Background()
	c, err := Run(ctx, testImage, WithDisableTelemetry())
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	endpoint, err := c.EndpointURL(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(endpoint, "https://"))
	keys, err := c.Credentials(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, keys.AccessKeyID)
	require.NotEmpty(t, keys.SecretAccessKey)

	tlsConfig, err := c.TLSConfig(ctx)
	require.NoError(t, err)
	require.False(t, tlsConfig.InsecureSkipVerify)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/__admin/health", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRunError(t *testing.T) {
	want := errors.New("invalid custom option")
	c, err := Run(context.Background(), testImage, testcontainers.CustomizeRequestOption(func(_ *testcontainers.GenericContainerRequest) error {
		return want
	}))
	testcontainers.CleanupContainer(t, c)
	require.ErrorIs(t, err, want)
}

func TestParseCredentials(t *testing.T) {
	for _, tc := range []struct {
		name  string
		data  string
		valid bool
	}{
		{"generated", "[default]\naws_access_key_id = emulator-key\naws_secret_access_key = secret==\n", true},
		{"other profile ignored", "[other]\naws_access_key_id = host-key\naws_secret_access_key = host-secret\n[default]\naws_access_key_id = emulator-key\naws_secret_access_key = secret==\n[other]\naws_access_key_id = wrong\n", true},
		{"empty", "", false},
		{"missing secret", "[default]\naws_access_key_id = emulator-key\n", false},
		{"other profile only", "[other]\naws_access_key_id = host-key\naws_secret_access_key = host-secret\n", false},
		{"oversized line", "[default]\n" + strings.Repeat("x", 70*1024), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCredentials(tc.data)
			if !tc.valid {
				require.Error(t, err)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, Credentials{AccessKeyID: "emulator-key", SecretAccessKey: "secret=="}, got)
		})
	}
}
