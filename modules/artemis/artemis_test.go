package artemis_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func TestArtemis(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		opts       []testcontainers.ContainerCustomizer
		user, pass string
		hook       func(*testing.T, *artemis.Container)
	}{
		{
			name: "Default",
			user: "artemis",
			pass: "artemis",
		},
		{
			name: "WithCredentials",
			opts: []testcontainers.ContainerCustomizer{
				// withCredentials {
				artemis.WithCredentials("test", "test"),
				// }
			},
			user: "test",
			pass: "test",
		},
		{
			name: "WithAnonymous",
			opts: []testcontainers.ContainerCustomizer{
				// withAnonymousLogin {
				artemis.WithAnonymousLogin(),
				// }
			},
		},
		{
			name: "WithExtraArgs",
			opts: []testcontainers.ContainerCustomizer{
				// withExtraArgs {
				artemis.WithExtraArgs("--http-host 0.0.0.0 --relax-jolokia --queues ArgsTestQueue"),
				// }
			},
			user: "artemis", // default user
			pass: "artemis", // default password
			hook: func(t *testing.T, container *artemis.Container) {
				t.Helper()
				expectQueue(t, container, "ArgsTestQueue")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctr, err := artemis.Run(ctx, "apache/activemq-artemis:2.30.0-alpine", test.opts...)
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			// consoleURL {
			u, err := ctr.ConsoleURL(ctx)
			// }
			require.NoError(t, err)

			res, err := http.Get(u)
			require.NoError(t, err, "failed to access console")
			res.Body.Close()
			require.Equal(t, http.StatusOK, res.StatusCode, "failed to access console")

			if test.user != "" {
				require.Equal(t, test.user, ctr.User(), "unexpected user")
			}

			if test.pass != "" {
				require.Equal(t, test.pass, ctr.Password(), "unexpected password")
			}

			// brokerEndpoint {
			host, err := ctr.BrokerEndpoint(ctx)
			// }
			require.NoError(t, err)

			endpointHost, endpointPort, err := net.SplitHostPort(host)
			require.NoError(t, err)
			require.NotEmpty(t, endpointHost)
			require.NotEmpty(t, endpointPort)
			// The four-case STOMP auth/pubsub matrix lives in examples/artemis_test.go.

			if test.hook != nil {
				test.hook(t, ctr)
			}
		})
	}
}

func expectQueue(t *testing.T, container *artemis.Container, queueName string) {
	t.Helper()

	u, err := container.ConsoleURL(context.Background())
	require.NoError(t, err)

	r, err := http.Get(u + `/jolokia/read/org.apache.activemq.artemis:broker="0.0.0.0"/QueueNames`)
	require.NoError(t, err, "failed to request QueueNames")
	defer r.Body.Close()

	var res struct{ Value []string }
	err = json.NewDecoder(r.Body).Decode(&res)
	require.NoError(t, err, "failed to decode QueueNames response")

	require.Containsf(t, res.Value, queueName, "should contain queue")
}
