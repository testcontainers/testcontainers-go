package testcontainer

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainer-go/wait"
)

func TestTwoContainersExposingTheSamePort(t *testing.T) {
	ctx := context.Background()
	nginxA, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nginxA.Terminate(ctx, t)

	nginxB, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nginxB.Terminate(ctx, t)

	ipA, err := nginxA.GetIPAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}

	ipB, err := nginxB.GetIPAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s", ipA))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	resp, err = http.Get(fmt.Sprintf("http://%s", ipB))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreation(t *testing.T) {
	ctx := context.Background()
	nginxC, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nginxC.Terminate(ctx, t)
	ip, err := nginxC.GetIPAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s", ip))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationAndWaitForListeningPortLongEnough(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := RunContainer(ctx, "menedev/delayed-nginx:1.15.2", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
		WaitingFor: wait.ForListeningPort(), // default startupTimeout is 60s
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nginxC.Terminate(ctx, t)
	ip, err := nginxC.GetIPAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s", ip))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationTimesOut(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := RunContainer(ctx, "menedev/delayed-nginx:1.15.2", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
		WaitingFor: wait.ForListeningPort().WithStartupTimeout(1 * time.Second),
	})
	if err == nil {
		t.Error("Expected timeout")
		nginxC.Terminate(ctx, t)
	}
}

func TestContainerRespondsWithHttp200ForIndex(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
		WaitingFor: wait.ForHttp("/"),
	})
	defer nginxC.Terminate(ctx, t)

	ip, err := nginxC.GetIPAddress(ctx)
	resp, err := http.Get(fmt.Sprintf("http://%s", ip))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerRespondsWithHttp404ForNonExistingPage(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
		WaitingFor: wait.ForHttp("/nonExistingPage").WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusNotFound
		}),
	})
	defer nginxC.Terminate(ctx, t)

	ip, err := nginxC.GetIPAddress(ctx)
	resp, err := http.Get(fmt.Sprintf("http://%s/nonExistingPage", ip))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d. Got %d.", http.StatusNotFound, resp.StatusCode)
	}
}


func TestContainerCreationTimesOutWithHttp(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := RunContainer(ctx, "menedev/delayed-nginx:1.15.2", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
		WaitingFor: wait.ForHttp("/").WithStartupTimeout(1 * time.Second),
	})
	defer nginxC.Terminate(ctx, t)
	if err == nil {
		t.Error("Expected timeout")
	}
}
