package valkey_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valkey-io/valkey-go"

	"github.com/testcontainers/testcontainers-go"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
)

func TestIntegrationSetGet(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5")
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 1)
}

func TestValkeyWithConfigFile(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5", tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")))
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 1)
}

func TestValkeyWithImage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		image string
	}{
		// There is only one release of Valkey at the time of writing
		{
			name:  "Valkey7.2.5",
			image: "docker.io/valkey/valkey:7.2.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valkeyContainer, err := tcvalkey.Run(ctx, tt.image, tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")))
			testcontainers.CleanupContainer(t, valkeyContainer)
			require.NoError(t, err)

			assertSetsGets(t, ctx, valkeyContainer, 1)
		})
	}
}

func TestValkeyWithLogLevel(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5", tcvalkey.WithLogLevel(tcvalkey.LogLevelVerbose))
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 10)
}

func TestValkeyWithSnapshotting(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5", tcvalkey.WithSnapshotting(10, 1))
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 10)
}

func assertSetsGets(t *testing.T, ctx context.Context, valkeyContainer *tcvalkey.ValkeyContainer, keyCount int) {
	// connectionString {
	uri, err := valkeyContainer.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	// You will likely want to wrap your Valkey package of choice in an
	// interface to aid in unit testing and limit lock-in throughout your
	// codebase but that's out of scope for this example
	options, err := valkey.ParseURL(uri)
	require.NoError(t, err)

	client, err := valkey.NewClient(options)
	require.NoError(t, err)
	defer func(t *testing.T, ctx context.Context, client *valkey.Client) {
		require.NoError(t, flushValkey(ctx, *client))
	}(t, ctx, &client)

	t.Log("pinging valkey")
	res := client.Do(ctx, client.B().Ping().Build())
	require.NoError(t, res.Error())

	t.Log("received response from valkey")

	msg, err := res.ToString()
	require.NoError(t, err)

	if msg != "PONG" {
		t.Fatalf("received unexpected response from valkey: %s", res.String())
	}

	for i := 0; i < keyCount; i++ {
		// Set data
		key := fmt.Sprintf("{user.%s}.favoritefood.%d", uuid.NewString(), i)
		value := fmt.Sprintf("Cabbage Biscuits %d", i)

		ttl, _ := time.ParseDuration("2h")

		err = client.Do(ctx, client.B().Set().Key(key).Value(value).Exat(time.Now().Add(ttl)).Build()).Error()
		require.NoError(t, err)

		err = client.Do(ctx, client.B().Expire().Key(key).Seconds(int64(ttl.Seconds())).Build()).Error()
		require.NoError(t, err)

		// Get data
		resp := client.Do(ctx, client.B().Get().Key(key).Build())
		require.NoError(t, resp.Error())

		retVal, err := resp.ToString()
		require.NoError(t, err)
		if retVal != value {
			t.Fatalf("Expected value %s. Got %s.", value, retVal)
		}
	}
}

func flushValkey(ctx context.Context, client valkey.Client) error {
	return client.Do(ctx, client.B().Flushall().Build()).Error()
}
