package nats_test

import (
	"bufio"
	"context"
	"strings"
	"testing"
	"time"

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

	require.NotEmpty(t, uri)
	require.Equal(t, uri, ctr.MustConnectionString(ctx))

	// validate /etc/nats.conf mentioned in logs
	const expected = "Using configuration file: /etc/nats.conf"
	logs, err := ctr.Logs(ctx)
	require.NoError(t, err)
	sc := bufio.NewScanner(logs)
	found := false
	timer := time.AfterFunc(5*time.Second, func() {
		require.Truef(t, found, "expected log line not found after 5 seconds: %s", expected)
	})
	defer timer.Stop()
	defer logs.Close()
	for sc.Scan() {
		if strings.Contains(sc.Text(), expected) {
			found = true
			break
		}
	}
	require.Truef(t, found, "expected log line not found: %s", expected)
}
