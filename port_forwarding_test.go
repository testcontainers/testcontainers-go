package testcontainers_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	expectedResponse = "Hello, World!"
)

func TestExposeHostPorts(t *testing.T) {
	tests := []struct {
		name             string
		numberOfPorts    int
		hasNetwork       bool
		bindOnPostStarts bool
	}{
		{
			name:          "single port",
			numberOfPorts: 1,
		},
		{
			name:          "single port using a network",
			numberOfPorts: 1,
			hasNetwork:    true,
		},
		{
			name:          "multiple ports",
			numberOfPorts: 3,
		},
		{
			name:             "multiple ports bound on PostStarts",
			numberOfPorts:    3,
			bindOnPostStarts: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			servers := make([]*httptest.Server, tc.numberOfPorts)
			freePorts := make([]int, tc.numberOfPorts)
			waitStrategies := make([]wait.Strategy, tc.numberOfPorts)
			for i := range freePorts {
				server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, expectedResponse)
				}))

				if !tc.bindOnPostStarts {
					server.Start()
				}

				servers[i] = server
				freePort := server.Listener.Addr().(*net.TCPAddr).Port
				freePorts[i] = freePort
				waitStrategies[i] = wait.
					ForExec([]string{"wget", "-q", "-O", "-", fmt.Sprintf("http://%s:%d", testcontainers.HostInternal, freePort)}).
					WithExitCodeMatcher(func(code int) bool {
						return code == 0
					}).
					WithResponseMatcher(func(body io.Reader) bool {
						bs, err := io.ReadAll(body)
						require.NoError(tt, err)
						return string(bs) == expectedResponse
					})

				tt.Cleanup(func() {
					server.Close()
				})
			}

			req := testcontainers.GenericContainerRequest{
				// hostAccessPorts {
				ContainerRequest: testcontainers.ContainerRequest{
					Image:           "alpine:3.17",
					HostAccessPorts: freePorts,
					WaitingFor:      wait.ForAll(waitStrategies...),
					LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
						{
							PostStarts: []testcontainers.ContainerHook{
								func(ctx context.Context, c testcontainers.Container) error {
									if tc.bindOnPostStarts {
										for _, server := range servers {
											server.Start()
										}
									}

									return nil
								},
								func(ctx context.Context, c testcontainers.Container) error {
									return waitStrategies[0].WaitUntilReady(ctx, c)
								},
							},
						},
					},
					Cmd: []string{"top"},
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
			c, err := testcontainers.GenericContainer(ctx, req)
			require.NoError(tt, err)
			_ = c.Terminate(ctx)
		})
	}
}
