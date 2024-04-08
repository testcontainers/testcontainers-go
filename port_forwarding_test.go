package testcontainers_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/network"
)

const (
	expectedResponse = "Hello, World!"
)

func TestExposeHostPorts(t *testing.T) {
	freePort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Free port:", freePort)

	// create a simple http server running in the host
	// this server will be accessed by the container
	// to check if the port forwarding is working
	server, err := createHttpServer(freePort)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	t.Cleanup(func() {
		server.Close()
	})

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "alpine:3.17",
			HostAccessPorts: []int{freePort},
			Cmd:             []string{"top"},
		},
		Started: true,
	}

	c, err := testcontainers.GenericContainer(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	// create a container that has host access, which will
	// automatically forward the port to the container
	assertContainerHasHostAccess(t, c, freePort)
}

func TestExposeHostPortsInNetwork(t *testing.T) {
	freePort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Free port:", freePort)

	// create a simple http server running in the host
	// this server will be accessed by the container
	// to check if the port forwarding is working
	server, err := createHttpServer(freePort)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	t.Cleanup(func() {
		server.Close()
	})

	nw, err := network.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := nw.Remove(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "alpine:3.17",
			HostAccessPorts: []int{freePort},
			Cmd:             []string{"top"},
			Networks:        []string{nw.ID},
			NetworkAliases:  map[string][]string{nw.ID: {"myalpine"}},
		},
		Started: true,
	}

	c, err := testcontainers.GenericContainer(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	// create a container that has host access, which will
	// automatically forward the port to the container
	assertContainerHasHostAccess(t, c, freePort)
}

func assertContainerHasHostAccess(t *testing.T, c testcontainers.Container, port int) {
	_, reader, err := c.Exec(
		context.Background(),
		[]string{"wget", "-O", "-", fmt.Sprintf("http://host.testcontainers.internal:%d", port)},
		tcexec.Multiplexed(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// read the response
	bs, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	// assert the response
	if string(bs) != expectedResponse {
		t.Fatalf("expected '%s' but got %s", expectedResponse, string(bs))
	}
}

func createHttpServer(port int) (*http.Server, error) {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedResponse)
	})

	return server, nil
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
