package activemq_test

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/activemq"
)

func TestActiveMQ(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		opts          []testcontainers.ContainerCustomizer
		adminUser     string
		adminPassword string
	}{
		{
			name:          "Default",
			adminUser:     "admin",
			adminPassword: "admin",
		},
		{
			name: "WithAdminCredentials",
			opts: []testcontainers.ContainerCustomizer{
				// withAdminCredentials {
				activemq.WithAdminCredentials("testuser", "testpass"),
				// }
			},
			adminUser:     "testuser",
			adminPassword: "testpass",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctr, err := activemq.Run(ctx, "apache/activemq-classic:5.18.7", tc.opts...)
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			require.Equal(t, tc.adminUser, ctr.AdminUser())
			require.Equal(t, tc.adminPassword, ctr.AdminPassword())

			// brokerURL {
			brokerURL, err := ctr.BrokerURL(ctx)
			// }
			require.NoError(t, err)
			require.Contains(t, brokerURL, "tcp://")

			// Verify the broker port (OpenWire) is reachable.
			u, err := url.Parse(brokerURL)
			require.NoError(t, err)
			conn, err := net.Dial("tcp", u.Host)
			require.NoError(t, err)
			conn.Close()

			// webConsoleURL {
			consoleURL, err := ctr.WebConsoleURL(ctx)
			// }
			require.NoError(t, err)
			require.Contains(t, consoleURL, "http://")

			// Verify the web console is accessible with the expected credentials.
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, consoleURL+"/admin", nil)
			require.NoError(t, err)
			req.SetBasicAuth(tc.adminUser, tc.adminPassword)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
