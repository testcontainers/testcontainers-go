package testcontainers_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/network"
)

const (
	expectedResponse = "Hello, World!"
)

func TestExposeHostPorts(t *testing.T) {
	hostPorts := make([]int, 3)
	for i := range hostPorts {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, expectedResponse)
		}))
		hostPorts[i] = server.Listener.Addr().(*net.TCPAddr).Port
		t.Cleanup(server.Close)
	}

	singlePort := hostPorts[0:1]

	t.Run("single-port", func(t *testing.T) {
		testExposeHostPorts(t, singlePort, false, false)
	})

	t.Run("single-port-network", func(t *testing.T) {
		testExposeHostPorts(t, singlePort, true, false)
	})

	t.Run("single-port-host-access", func(t *testing.T) {
		testExposeHostPorts(t, singlePort, false, true)
	})

	t.Run("single-port-network-host-access", func(t *testing.T) {
		testExposeHostPorts(t, singlePort, true, true)
	})

	t.Run("multi-port", func(t *testing.T) {
		testExposeHostPorts(t, hostPorts, false, false)
	})

	t.Run("multi-port-network", func(t *testing.T) {
		testExposeHostPorts(t, hostPorts, true, false)
	})

	t.Run("multi-port-host-access", func(t *testing.T) {
		testExposeHostPorts(t, hostPorts, false, true)
	})

	t.Run("multi-port-network-host-access", func(t *testing.T) {
		testExposeHostPorts(t, hostPorts, true, true)
	})
}

func testExposeHostPorts(t *testing.T, hostPorts []int, hasNetwork, hasHostAccess bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var hostAccessPorts []int
	if hasHostAccess {
		hostAccessPorts = hostPorts
	}
	req := testcontainers.GenericContainerRequest{
		// hostAccessPorts {
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "alpine:3.17",
			HostAccessPorts: hostAccessPorts,
			Cmd:             []string{"top"},
		},
		// }
		Started: true,
	}

	if hasNetwork {
		nw, err := network.New(ctx)
		require.NoError(t, err)
		testcontainers.CleanupNetwork(t, nw)

		req.Networks = []string{nw.Name}
		req.NetworkAliases = map[string][]string{nw.Name: {"myalpine"}}
	}

	c, err := testcontainers.GenericContainer(ctx, req)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	if hasHostAccess {
		// Verify that the container can access the host ports.
		containerHasHostAccess(t, c, hostPorts...)
		return
	}

	// Verify that the container cannot access the host ports.
	containerHasNoHostAccess(t, c, hostPorts...)
}

// httpRequest sends an HTTP request from the container to the host port via
// [testcontainers.HostInternal] address.
func httpRequest(t *testing.T, c testcontainers.Container, port int) (int, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// wgetHostInternal {
	code, reader, err := c.Exec(
		ctx,
		[]string{"wget", "-q", "-O", "-", "-T", "2", fmt.Sprintf("http://%s:%d", testcontainers.HostInternal, port)},
		tcexec.Multiplexed(),
	)
	// }
	require.NoError(t, err)

	// read the response
	bs, err := io.ReadAll(reader)
	require.NoError(t, err)

	return code, string(bs)
}

// containerHasHostAccess verifies that the container can access the host ports
// via [testcontainers.HostInternal] address.
func containerHasHostAccess(t *testing.T, c testcontainers.Container, ports ...int) {
	t.Helper()
	for _, port := range ports {
		code, response := httpRequest(t, c, port)
		require.Zero(t, code)
		require.Equal(t, expectedResponse, response)
	}
}

// containerHasNoHostAccess verifies that the container cannot access the host ports
// via [testcontainers.HostInternal] address.
func containerHasNoHostAccess(t *testing.T, c testcontainers.Container, ports ...int) {
	t.Helper()
	for _, port := range ports {
		code, response := httpRequest(t, c, port)
		require.NotZero(t, code)
		require.Contains(t, response, "bad address")
	}
}
