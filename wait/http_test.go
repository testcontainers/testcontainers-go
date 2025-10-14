package wait_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed testdata/root.pem
var caBytes []byte

// https://github.com/testcontainers/testcontainers-go/issues/183
func ExampleHTTPStrategy() {
	// waitForHTTPWithDefaultPort {
	ctx := context.Background()
	c, err := testcontainers.Run(ctx, "nginx:latest",
		testcontainers.WithExposedPorts("80/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
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

	c, err := testcontainers.Run(
		ctx, "",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context: filepath.Join("testdata", "http"),
		}),
		testcontainers.WithExposedPorts("6443/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/headers").
				WithTLS(true, tlsconfig).
				WithPort("6443/tcp").
				WithHeaders(map[string]string{"X-request-header": "value"}).
				WithResponseHeadersMatcher(func(headers http.Header) bool {
					return headers.Get("X-response-header") == "value"
				}),
		),
	)
	// }
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
	c, err := testcontainers.Run(
		ctx, "nginx:latest",
		testcontainers.WithExposedPorts("8080/tcp", "80/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/").WithPort("80/tcp")),
	)
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
	c, err := testcontainers.Run(
		ctx, "nginx:latest",
		testcontainers.WithExposedPorts("8080/tcp", "80/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/").WithForcedIPv4LocalHost()),
	)
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
	gogs, err := testcontainers.Run(ctx, "gogs/gogs:0.11.91",
		testcontainers.WithExposedPorts("3000/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/").WithBasicAuth("username", "password")),
	)
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
	certpool := x509.NewCertPool()
	require.Truef(t, certpool.AppendCertsFromPEM(caBytes), "the ca file isn't valid")

	tlsconfig := &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}
	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context: filepath.Join("testdata", "http"),
		}),
		testcontainers.WithExposedPorts("6443/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/auth-ping").WithTLS(true, tlsconfig).
			WithStartupTimeout(time.Second*10).WithPort("6443/tcp").
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			}).
			WithBasicAuth("admin", "admin").
			WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
		),
	}
	ctx := t.Context()

	ctr, err := testcontainers.Run(ctx, "", opts...)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s:%s", host, port.Port()), http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPStrategyWaitUntilReadyWithQueryString(t *testing.T) {
	certpool := x509.NewCertPool()
	require.Truef(t, certpool.AppendCertsFromPEM(caBytes), "the ca file isn't valid")

	tlsconfig := &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}
	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context: filepath.Join("testdata", "http"),
		}),
		testcontainers.WithExposedPorts("6443/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/query-params-ping?v=pong").WithTLS(true, tlsconfig).
			WithStartupTimeout(time.Second * 10).WithPort("6443/tcp").
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			})),
	}
	ctx := t.Context()

	ctr, err := testcontainers.Run(ctx, "", opts...)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s:%s", host, port.Port()), http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPStrategyWaitUntilReadyNoBasicAuth(t *testing.T) {
	certpool := x509.NewCertPool()
	require.Truef(t, certpool.AppendCertsFromPEM(caBytes), "the ca file isn't valid")

	// waitForHTTPStatusCode {
	tlsconfig := &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}
	var i int
	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context: filepath.Join("testdata", "http"),
		}),
		testcontainers.WithExposedPorts("6443/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/ping").WithTLS(true, tlsconfig).
			WithStartupTimeout(time.Second * 10).WithPort("6443/tcp").
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			}).
			WithStatusCodeMatcher(func(status int) bool {
				i++ // always fail the first try in order to force the polling loop to be re-run
				return i > 1 && status == 200
			}).
			WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
		),
	}
	// }

	ctx := t.Context()
	ctr, err := testcontainers.Run(ctx, "", opts...)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s:%s", host, port.Port()), http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				OOMKilled: true,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status: "dead",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				OOMKilled: true,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status: "dead",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:  "running",
				Running: true,
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
	expected := "no exposed tcp ports or mapped ports - cannot wait for status"
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
				Status:  "running",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
	expected := "no exposed tcp ports or mapped ports - cannot wait for status"
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
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
				Status:  "running",
			}, nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
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
	expected := "no exposed tcp ports or mapped ports - cannot wait for status"
	require.EqualError(t, err, expected)
}
