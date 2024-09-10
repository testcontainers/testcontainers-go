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
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	// }

	defer func() {
		if err := c.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := c.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleHTTPStrategy_WithHeaders() {
	capath := filepath.Join("testdata", "root.pem")
	cafile, err := os.ReadFile(capath)
	if err != nil {
		log.Fatalf("can't load ca file: %v", err)
	}

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(cafile) {
		log.Fatalf("the ca file isn't valid")
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
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := c.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := c.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	// }

	defer func() {
		if err := c.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := c.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := c.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := c.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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
		log.Fatalf("failed to start container: %s", err)
	}
	// }

	defer func() {
		if err := gogs.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := gogs.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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

func TestHTTPStrategyWaitUntilReadyWithQueryString(t *testing.T) {
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
		WaitingFor: wait.NewHTTPStrategy("/query-params-ping?v=pong").WithTLS(true, tlsconfig).
			WithStartupTimeout(time.Second * 10).WithPort("6443/tcp").
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			}),
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

// testForHTTP is a helper function to test the HTTP strategy.
func testForHTTP(t *testing.T, strategy wait.Strategy) {
	udpPorts := nat.PortMap{
		"80/udp": []nat.PortBinding{{
			HostIP:   hostAddress,
			HostPort: defaultHostPort,
		}},
	}

	testPortScenarios(t, strategy, func(t *testing.T, b *waitBuilder) *waitBuilder {
		if b.NoTCP() {
			// No TCP ports so no HTTP server needed.
			return b
		}

		// Start a HTTP server and update the mapped ports.
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(svr.Close)

		_, port, err := net.SplitHostPort(svr.Listener.Addr().String())
		require.NoError(t, err)

		// Maintain SendingRequest state.
		if b.Ports() == 2 {
			return b.MappedPorts("", nat.Port(port))
		}

		return b.MappedPorts(nat.Port(port))
	})

	t.Run("getting-port/only-udp-ports", func(t *testing.T) {
		var portErr wait.PortNotFoundErr
		newWaitBuilder().
			InternalPort("80/udp").
			InspectPortMap(udpPorts).
			ErrorAs(&portErr).Run(t, strategy)
	})

	t.Run("sending-request/only-udp-ports", func(t *testing.T) {
		var portErr wait.PortNotFoundErr
		newWaitBuilder().
			InternalPort("80/udp").
			InspectPortMap(udpPorts).
			SendingRequest(true).
			ErrorAs(&portErr).Run(t, strategy)
	})
}

func TestForHTTP(t *testing.T) {
	t.Run("no-port", func(t *testing.T) {
		strategy := wait.ForHTTP("/").
			WithStartupTimeout(200 * time.Millisecond).
			WithPollInterval(10 * time.Millisecond)
		testForHTTP(t, strategy)
	})

	t.Run("with-port", func(t *testing.T) {
		strategy := wait.ForHTTP("/").
			WithPort("80/tcp").
			WithStartupTimeout(200 * time.Millisecond).
			WithPollInterval(10 * time.Millisecond)
		testForHTTP(t, strategy)
	})
}
