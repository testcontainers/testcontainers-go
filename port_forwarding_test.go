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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, expectedResponse)
		}))
		hostPorts[i] = server.Listener.Addr().(*net.TCPAddr).Port
		t.Cleanup(func() {
			server.Close()
		})
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

	req := testcontainers.GenericContainerRequest{
		// hostAccessPorts {
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "alpine:3.17",
			HostAccessPorts: hostPorts,
			Cmd:             []string{"top"},
		},
		// }
		Started: true,
	}

	if hasNetwork {
		nw, err := network.New(context.Background())
		require.NoError(t, err)
		testcontainers.CleanupNetwork(t, nw)

		req.Networks = []string{nw.Name}
		req.NetworkAliases = map[string][]string{nw.Name: {"myalpine"}}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := testcontainers.GenericContainer(ctx, req)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	if hasHostAccess {
		// Create a container that has host access, which will
		// automatically forward the port to the container.
		assertContainerHasHostAccess(t, c, hostPorts...)
		return
	}

	// Force cancellation.
	cancel()

	assertContainerHasNoHostAccess(t, c, hostPorts...)
}

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

func assertContainerHasHostAccess(t *testing.T, c testcontainers.Container, ports ...int) {
	t.Helper()
	for _, port := range ports {
		code, response := httpRequest(t, c, port)
		require.Zerof(t, code, "expected status code [%d] but got [%d]", 0, code)

		require.Equalf(t, expectedResponse, response, "expected [%s] but got [%s]", expectedResponse, response)
	}
}

func assertContainerHasNoHostAccess(t *testing.T, c testcontainers.Container, ports ...int) {
	t.Helper()
	for _, port := range ports {
		_, response := httpRequest(t, c, port)

		require.NotEqualf(t, expectedResponse, response, "expected not to get [%s] but got [%s]", expectedResponse, response)
	}
}
