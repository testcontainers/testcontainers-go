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
	t.Run("single-port", func(t *testing.T) {
		testExposeHostPorts(t, 1, false, false)
	})

	t.Run("single-port-network", func(t *testing.T) {
		testExposeHostPorts(t, 1, true, false)
	})

	t.Run("single-port-host-access", func(t *testing.T) {
		testExposeHostPorts(t, 1, false, true)
	})

	t.Run("single-port-network-host-access", func(t *testing.T) {
		testExposeHostPorts(t, 1, true, true)
	})

	t.Run("multi-port", func(t *testing.T) {
		testExposeHostPorts(t, 3, false, false)
	})

	t.Run("multi-port-network", func(t *testing.T) {
		testExposeHostPorts(t, 3, true, false)
	})

	t.Run("multi-port-host-access", func(t *testing.T) {
		testExposeHostPorts(t, 3, false, true)
	})

	t.Run("multi-port-network-host-access", func(t *testing.T) {
		testExposeHostPorts(t, 3, true, true)
	})
}

func testExposeHostPorts(t *testing.T, numberOfPorts int, hasNetwork, hasHostAccess bool) {
	t.Helper()

	freePorts := make([]int, numberOfPorts)
	for i := range freePorts {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, expectedResponse)
		}))
		freePorts[i] = server.Listener.Addr().(*net.TCPAddr).Port
		t.Cleanup(func() {
			server.Close()
		})
	}

	req := testcontainers.GenericContainerRequest{
		// hostAccessPorts {
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "alpine:3.17",
			HostAccessPorts: freePorts,
			Cmd:             []string{"top"},
		},
		// }
		Started: true,
	}

	var nw *testcontainers.DockerNetwork
	if hasNetwork {
		var err error
		nw, err = network.New(context.Background())
		require.NoError(t, err)
		testcontainers.CleanupNetwork(t, nw)

		req.Networks = []string{nw.Name}
		req.NetworkAliases = map[string][]string{nw.Name: {"myalpine"}}
	}

	ctx := context.Background()
	if !hasHostAccess {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
	}

	c, err := testcontainers.GenericContainer(ctx, req)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	if hasHostAccess {
		// Create a container that has host access, which will
		// automatically forward the port to the container.
		assertContainerHasHostAccess(t, c, freePorts...)
	} else {
		// Force cancellation because of timeout.
		time.Sleep(4 * time.Second)

		assertContainerHasNoHostAccess(t, c, freePorts...)
	}
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
