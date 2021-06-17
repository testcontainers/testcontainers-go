package wait_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//
// https://github.com/testcontainers/testcontainers-go/issues/183
func ExampleHTTPStrategy() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "gogs/gogs:0.11.91",
		ExposedPorts: []string{"3000/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("3000/tcp"),
	}

	gogs, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	defer gogs.Terminate(ctx) // nolint: errcheck
	// Here you have a running container
}

func TestHTTPStrategyWaitUntilReady(t *testing.T) {
	workdir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	capath := workdir + "/testdata/root.pem"
	cafile, err := ioutil.ReadFile(capath)
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
			Context: workdir + "/testdata",
		},
		ExposedPorts: []string{"6443/tcp"},
		WaitingFor: wait.NewHTTPStrategy("/ping").WithTLS(true, tlsconfig).
			WithTimeout(time.Second * 10).WithPort("6443/tcp").
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := ioutil.ReadAll(body)
				return bytes.Equal(data, []byte("pong"))
			}).
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
