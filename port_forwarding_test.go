package testcontainers_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			freePorts := make([]int, tt.numberOfPorts)
			for i := range freePorts {
				freePort, err := getFreePort()
				if err != nil {
					t.Fatal(err)
				}

				freePorts[i] = freePort

				// create an http server for each port
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

			if tt.hasNetwork {
				nw, err := network.New(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := nw.Remove(context.Background()); err != nil {
						t.Fatal(err)
					}
				})

				req.Networks = []string{nw.ID}
				req.NetworkAliases = map[string][]string{nw.ID: {"myalpine"}}
			}

			ctx := context.Background()
			if !tt.hasHostAccess {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
				defer cancel()
			}

			c, err := testcontainers.GenericContainer(ctx, req)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := c.Terminate(context.Background()); err != nil {
					t.Fatal(err)
				}
			})

			if tt.hasHostAccess {
				// create a container that has host access, which will
				// automatically forward the port to the container
				assertContainerHasHostAccess(t, c, freePorts...)
			} else {
				// force cancellation because of timeout
				time.Sleep(11 * time.Second)

				assertContainerHasNoHostAccess(t, c, freePorts...)
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
	if err != nil {
		t.Fatal(err)
	}

	// read the response
	bs, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

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
