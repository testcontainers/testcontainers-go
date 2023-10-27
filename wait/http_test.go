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

	gogs, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	// }

	defer func() {
		if err := gogs.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := gogs.State(ctx)
	if err != nil {
		panic(err)
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

	gogs, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	// }

	defer func() {
		if err := gogs.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := gogs.State(ctx)
	if err != nil {
		panic(err)
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
	if err != nil {
		panic(err)
	}
	// }

	defer func() {
		if err := gogs.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := gogs.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func TestHTTPStrategyWaitUntilReady(t *testing.T) {
	workdir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	capath := filepath.Join(workdir, "testdata", "root.pem")
	cafile, err := os.ReadFile(capath)
	if err != nil {
		t.Errorf("can't load ca file: %v", err)
		return
	}

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

	container, err := testcontainers.GenericContainer(context.Background(),
		testcontainers.GenericContainerRequest{ContainerRequest: dockerReq, Started: true})
	if err != nil {
		t.Error(err)
		return
	}
	defer container.Terminate(context.Background()) // nolint: errcheck

	host, err := container.Host(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	port, err := container.MappedPort(context.Background(), "6443/tcp")
	if err != nil {
		t.Error(err)
		return
	}
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
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code isn't ok: %s", resp.Status)
		return
	}
}

func TestHTTPStrategyWaitUntilReadyNoBasicAuth(t *testing.T) {
	workdir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	capath := filepath.Join(workdir, "testdata", "root.pem")
	cafile, err := os.ReadFile(capath)
	if err != nil {
		t.Errorf("can't load ca file: %v", err)
		return
	}

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
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: dockerReq, Started: true})
	if err != nil {
		t.Error(err)
		return
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	port, err := container.MappedPort(ctx, "6443/tcp")
	if err != nil {
		t.Error(err)
		return
	}
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
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code isn't ok: %s", resp.Status)
		return
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "49152",
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container crashed with out-of-memory (OOMKilled)"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "49152",
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container exited with code 1"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "49152",
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "unexpected container status \"dead\""
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "49152",
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container crashed with out-of-memory (OOMKilled)"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "49152",
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container exited with code 1"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "49152",
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "unexpected container status \"dead\""
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "No exposed tcp ports or mapped ports - cannot wait for status"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/udp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "49152",
					},
				},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "No exposed tcp ports or mapped ports - cannot wait for status"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
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
		PortsImpl: func(ctx context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"8080/tcp": []nat.PortBinding{},
			}, nil
		},
	}

	wg := wait.ForHTTP("/").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "No exposed tcp ports or mapped ports - cannot wait for status"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}
