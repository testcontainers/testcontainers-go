package artemis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/stretchr/testify/assert"
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
			user: "artemis",
			pass: "artemis",
			hook: func(t *testing.T, container *artemis.Container) {
				expectQueue(t, container, "ArgsTestQueue")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctr, err := artemis.Run(ctx, "docker.io/apache/activemq-artemis:2.30.0-alpine", test.opts...)
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			// consoleURL {
			u, err := ctr.ConsoleURL(ctx)
			// }
			require.NoError(t, err)

			res, err := http.Get(u)
			require.NoError(t, err, "failed to access console")
			res.Body.Close()
			assert.Equal(t, http.StatusOK, res.StatusCode, "failed to access console")

			if test.user != "" {
				assert.Equal(t, test.user, ctr.User(), "unexpected user")
			}

			if test.pass != "" {
				assert.Equal(t, test.pass, ctr.Password(), "unexpected password")
			}

			// brokerEndpoint {
			host, err := ctr.BrokerEndpoint(ctx)
			// }
			require.NoError(t, err)

			var opt []func(*stomp.Conn) error
			if test.user != "" || test.pass != "" {
				opt = append(opt, stomp.ConnOpt.Login(test.user, test.pass))
			}

			conn, err := stomp.Dial("tcp", host, opt...)
			require.NoError(t, err, "failed to connect")
			t.Cleanup(func() { require.NoError(t, conn.Disconnect()) })

			sub, err := conn.Subscribe("test", stomp.AckAuto)
			require.NoError(t, err, "failed to subscribe")
			t.Cleanup(func() { require.NoError(t, sub.Unsubscribe()) })

			err = conn.Send("test", "", []byte("test"))
			require.NoError(t, err, "failed to send")

			ticker := time.NewTicker(10 * time.Second)
			select {
			case <-ticker.C:
				t.Fatal("timed out waiting for message")
			case msg := <-sub.C:
				require.Equal(t, "test", string(msg.Body), "received unexpected message")
			}

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
