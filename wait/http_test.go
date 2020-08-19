package wait_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
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

	// Here you have a running container

	_ = gogs.Terminate(ctx)
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

	dockerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    workdir + "/testdata",
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"6443/tcp"},
		WaitingFor: wait.NewHTTPStrategy("/").
			WithTLS(true, &tls.Config{RootCAs: certpool, ServerName: "testcontainer.go.test"}).
			WithStartupTimeout(time.Second * 10).WithPort("6443/tcp").
			WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{ContainerRequest: dockerReq, Started: false})

	if err != nil {
		t.Error(err)
		return
	}
	_ = container.Terminate(context.Background())
}
