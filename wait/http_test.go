package wait_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// https://github.com/testcontainers/testcontainers-go/issues/183
func ExampleHTTPStrategy() {
	// waitForHTTPWithDefaultPort {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nginx:latest",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(c); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := c.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleHTTPStrategy_WithHeaders() {
	capath := filepath.Join("testdata", "root.pem")
	cafile, err := os.ReadFile(capath)
	if err != nil {
		log.Printf("can't load ca file: %v", err)
		return
	}

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(cafile) {
		log.Printf("the ca file isn't valid")
		return
	}

	ctx := context.Background()

	// waitForHTTPHeaders {
	tlsconfig := &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: "testdata",
		},
		ExposedPorts: []string{"6443/tcp"},
		WaitingFor: wait.ForHTTP("/headers").
			WithTLS(true, tlsconfig).
			WithPort("6443/tcp").
			WithHeaders(map[string]string{"X-request-header": "value"}).
			WithResponseHeadersMatcher(func(headers http.Header) bool {
				return headers.Get("X-response-header") == "value"
			},
			),
	}
	// }

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(c); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := c.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleHTTPStrategy_WithPort() {
	// waitForHTTPWithPort {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nginx:latest",
		ExposedPorts: []string{"8080/tcp", "80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("80/tcp"),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(c); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := c.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleHTTPStrategy_WithForcedIPv4LocalHost() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nginx:latest",
		ExposedPorts: []string{"8080/tcp", "80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithForcedIPv4LocalHost(),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(c); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := c.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleHTTPStrategy_WithBasicAuth() {
	// waitForBasicAuth {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "gogs/gogs:0.11.91",
		ExposedPorts: []string{"3000/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithBasicAuth("username", "password"),
	}

	gogs, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(gogs); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := gogs.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func TestHTTPStrategyWaitUntilReady(t *testing.T) {
	workdir, err := os.Getwd()
	require.NoError(t, err)

	capath := filepath.Join(workdir, "testdata", "root.pem")
	cafile, err := os.ReadFile(capath)
	require.NoError(t, err)

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(cafile) {
		t.Errorf("the ca file isn't valid")
		return
	}

	tlsconfig := &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}
	dockerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: filepath.Join(workdir, "testdata"),
		},
		ExposedPorts: []string{"6443/tcp"},
		WaitingFor: wait.NewHTTPStrategy("/auth-ping").WithTLS(true, tlsconfig).
			WithStartupTimeout(time.Second*10).WithPort("6443/tcp").
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			}).
			WithBasicAuth("admin", "admin").
			WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
	}

	ctr, err := testcontainers.GenericContainer(context.Background(),
		testcontainers.GenericContainerRequest{ContainerRequest: dockerReq, Started: true})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.Host(context.Background())
	require.NoError(t, err)

	port, err := ctr.MappedPort(context.Background(), "6443/tcp")
	require.NoError(t, err)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsconfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	resp, err := client.Get(fmt.Sprintf("https://%s:%s", host, port.Port()))
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPStrategyWaitUntilReadyWithQueryString(t *testing.T) {
	workdir, err := os.Getwd()
	require.NoError(t, err)

	capath := filepath.Join(workdir, "testdata", "root.pem")
	cafile, err := os.ReadFile(capath)
	require.NoError(t, err)

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(cafile) {
		t.Errorf("the ca file isn't valid")
		return
	}

	tlsconfig := &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}
	dockerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: filepath.Join(workdir, "testdata"),
		},

		ExposedPorts: []string{"6443/tcp"},
		WaitingFor: wait.NewHTTPStrategy("/query-params-ping?v=pong").WithTLS(true, tlsconfig).
			WithStartupTimeout(time.Second * 10).WithPort("6443/tcp").
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			}),
	}

	ctr, err := testcontainers.GenericContainer(context.Background(),
		testcontainers.GenericContainerRequest{ContainerRequest: dockerReq, Started: true})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.Host(context.Background())
	require.NoError(t, err)

	port, err := ctr.MappedPort(context.Background(), "6443/tcp")
	require.NoError(t, err)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsconfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	resp, err := client.Get(fmt.Sprintf("https://%s:%s", host, port.Port()))
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPStrategyWaitUntilReadyNoBasicAuth(t *testing.T) {
	workdir, err := os.Getwd()
	require.NoError(t, err)

	capath := filepath.Join(workdir, "testdata", "root.pem")
	cafile, err := os.ReadFile(capath)
	require.NoError(t, err)

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(cafile) {
		t.Errorf("the ca file isn't valid")
		return
	}

	// waitForHTTPStatusCode {
	tlsconfig := &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}
	var i int
	dockerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: filepath.Join(workdir, "testdata"),
		},
		ExposedPorts: []string{"6443/tcp"},
		WaitingFor: wait.NewHTTPStrategy("/ping").WithTLS(true, tlsconfig).
			WithStartupTimeout(time.Second * 10).
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			}).
			WithStatusCodeMatcher(func(status int) bool {
				i++ // always fail the first try in order to force the polling loop to be re-run
				return i > 1 && status == 200
			}).
			WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
	}
	// }

	ctx := context.Background()
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: dockerReq, Started: true})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.Host(ctx)
	require.NoError(t, err)

	port, err := ctr.MappedPort(ctx, "6443/tcp")
	require.NoError(t, err)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsconfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	resp, err := client.Get(fmt.Sprintf("https://%s:%s", host, port.Port()))
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHttpStrategyFailsWhileGettingPortDueToOOMKilledContainer(t *testing.T) {
	var mappedPortCount int
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", wait.ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				OOMKilled: true,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": []nat.PortBinding{
								{
									HostIP:   "127.0.0.1",
									HostPort: "49152",
								},
							},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "container crashed with out-of-memory (OOMKilled)"
	require.EqualError(t, err, expected)
}

func TestHttpStrategyFailsWhileGettingPortDueToExitedContainer(t *testing.T) {
	var mappedPortCount int
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", wait.ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": []nat.PortBinding{
								{
									HostIP:   "127.0.0.1",
									HostPort: "49152",
								},
							},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "container exited with code 1"
	require.EqualError(t, err, expected)
}

func TestHttpStrategyFailsWhileGettingPortDueToUnexpectedContainerStatus(t *testing.T) {
	var mappedPortCount int
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", wait.ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status: "dead",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": []nat.PortBinding{
								{
									HostIP:   "127.0.0.1",
									HostPort: "49152",
								},
							},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "unexpected container status \"dead\""
	require.EqualError(t, err, expected)
}

func TestHTTPStrategyFailsWhileRequestSendingDueToOOMKilledContainer(t *testing.T) {
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				OOMKilled: true,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": []nat.PortBinding{
								{
									HostIP:   "127.0.0.1",
									HostPort: "49152",
								},
							},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "container crashed with out-of-memory (OOMKilled)"
	require.EqualError(t, err, expected)
}

func TestHttpStrategyFailsWhileRequestSendingDueToExitedContainer(t *testing.T) {
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": []nat.PortBinding{
								{
									HostIP:   "127.0.0.1",
									HostPort: "49152",
								},
							},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "container exited with code 1"
	require.EqualError(t, err, expected)
}

func TestHttpStrategyFailsWhileRequestSendingDueToUnexpectedContainerStatus(t *testing.T) {
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status: "dead",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": []nat.PortBinding{
								{
									HostIP:   "127.0.0.1",
									HostPort: "49152",
								},
							},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "unexpected container status \"dead\""
	require.EqualError(t, err, expected)
}

func TestHttpStrategyFailsWhileGettingPortDueToNoExposedPorts(t *testing.T) {
	var mappedPortCount int
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", wait.ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status:  "running",
				Running: true,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "No exposed tcp ports or mapped ports - cannot wait for status"
	require.EqualError(t, err, expected)
}

func TestHttpStrategyFailsWhileGettingPortDueToOnlyUDPPorts(t *testing.T) {
	var mappedPortCount int
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", wait.ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Running: true,
				Status:  "running",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/udp": []nat.PortBinding{
								{
									HostIP:   "127.0.0.1",
									HostPort: "49152",
								},
							},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "No exposed tcp ports or mapped ports - cannot wait for status"
	require.EqualError(t, err, expected)
}

func TestHttpStrategyFailsWhileGettingPortDueToExposedPortNoBindings(t *testing.T) {
	var mappedPortCount int
	target := &wait.MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", wait.ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Running: true,
				Status:  "running",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": []nat.PortBinding{},
						},
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	expected := "No exposed tcp ports or mapped ports - cannot wait for status"
	require.EqualError(t, err, expected)
}
