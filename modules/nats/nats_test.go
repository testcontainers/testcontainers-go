package nats_test

import (
	"bufio"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcnats "github.com/testcontainers/testcontainers-go/modules/nats"
)

func TestNATS(t *testing.T) {
	ctx := context.Background()

	//  createNATSContainer {
	ctr, err := tcnats.Run(ctx, "nats:2.9")
	//  }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	uri, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	mustURI := ctr.MustConnectionString(ctx)
	require.Equal(t, mustURI, uri)

	// perform assertions
	nc, err := nats.Connect(uri)
	require.NoError(t, err)
	defer nc.Close()

	js, err := nc.JetStream()
	require.NoError(t, err)

	// add stream to nats
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "hello",
		Subjects: []string{"hello"},
	})
	require.NoError(t, err)

	// add subscriber to nats
	sub, err := js.SubscribeSync("hello", nats.Durable("worker"))
	require.NoError(t, err)

	// publish a message to nats
	_, err = js.Publish("hello", []byte("hello"))
	require.NoError(t, err)

	// wait for the message to be received
	msg, err := sub.NextMsgWithContext(ctx)
	require.NoError(t, err)

	require.Equal(t, "hello", string(msg.Data))
}

func TestNATSWithConfigFile(t *testing.T) {
	const natsConf = `
listen: 0.0.0.0:4222
authorization {
    token: "s3cr3t"
}
`
	ctx := context.Background()

	ctr, err := tcnats.Run(ctx, "nats:2.9", tcnats.WithConfigFile(strings.NewReader(natsConf)))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	uri, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	// connect without a correct token must fail
	mallory, err := nats.Connect(uri, nats.Name("Mallory"), nats.Token("secret"))
	t.Cleanup(mallory.Close)
	require.EqualError(t, err, "nats: Authorization Violation")

	// connect with a correct token must succeed
	nc, err := nats.Connect(uri, nats.Name("API Token Test"), nats.Token("s3cr3t"))
	t.Cleanup(nc.Close)
	require.NoError(t, err)

	// validate /etc/nats.conf mentioned in logs
	const expected = "Using configuration file: /etc/nats.conf"
	logs, err := ctr.Logs(ctx)
	require.NoError(t, err)
	sc := bufio.NewScanner(logs)
	found := false
	time.AfterFunc(5*time.Second, func() {
		require.Truef(t, found, "expected log line not found after 5 seconds: %s", expected)
	})
	for sc.Scan() {
		if strings.Contains(sc.Text(), expected) {
			found = true
			break
		}
	}
	require.Truef(t, found, "expected log line not found: %s", expected)
}
