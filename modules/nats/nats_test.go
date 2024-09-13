package nats_test

import (
	"context"
	"testing"

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

	mustUri := ctr.MustConnectionString(ctx)
	require.Equal(t, mustUri, uri)

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
