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
	tests := []struct {
		name          string
		numberOfPorts int
		hasNetwork    bool
		hasHostAccess bool
	}{
		{
			name:          "single port",
			numberOfPorts: 1,
			hasHostAccess: true,
		},
		{
			name:          "single port using a network",
			numberOfPorts: 1,
			hasNetwork:    true,
			hasHostAccess: true,
		},
		{
			name:          "multiple ports",
			numberOfPorts: 3,
			hasHostAccess: true,
		},
		{
			name:          "single port with cancellation",
			numberOfPorts: 1,
			hasHostAccess: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			freePorts := make([]int, tc.numberOfPorts)
			for i := range freePorts {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, expectedResponse)
				}))
				freePorts[i] = server.Listener.Addr().(*net.TCPAddr).Port
				tt.Cleanup(func() {
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
			if tc.hasNetwork {
				var err error
				nw, err = network.New(context.Background())
				require.NoError(tt, err)
				testcontainers.CleanupNetwork(t, nw)

				req.Networks = []string{nw.Name}
				req.NetworkAliases = map[string][]string{nw.Name: {"myalpine"}}
			}

			ctx := context.Background()
			if !tc.hasHostAccess {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
				defer cancel()
			}

			c, err := testcontainers.GenericContainer(ctx, req)
			testcontainers.CleanupContainer(t, c)
			require.NoError(tt, err)

			if tc.hasHostAccess {
				// create a container that has host access, which will
				// automatically forward the port to the container
				assertContainerHasHostAccess(tt, c, freePorts...)
			} else {
				// force cancellation because of timeout
				time.Sleep(11 * time.Second)

				assertContainerHasNoHostAccess(tt, c, freePorts...)
			}
		})
	}
}

func httpRequest(t *testing.T, c testcontainers.Container, port int) (int, string) {
	// wgetHostInternal {
	code, reader, err := c.Exec(
		context.Background(),
		[]string{"wget", "-q", "-O", "-", fmt.Sprintf("http://%s:%d", testcontainers.HostInternal, port)},
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
	for _, port := range ports {
		code, response := httpRequest(t, c, port)
		if code != 0 {
			t.Fatalf("expected status code [%d] but got [%d]", 0, code)
		}

		if response != expectedResponse {
			t.Fatalf("expected [%s] but got [%s]", expectedResponse, response)
		}
	}
}

func assertContainerHasNoHostAccess(t *testing.T, c testcontainers.Container, ports ...int) {
	for _, port := range ports {
		_, response := httpRequest(t, c, port)

		if response == expectedResponse {
			t.Fatalf("expected not to get [%s] but got [%s]", expectedResponse, response)
		}
	}
}
