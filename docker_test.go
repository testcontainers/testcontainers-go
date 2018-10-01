package testcontainer

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestTwoContainersExposingTheSamePort(t *testing.T) {
	ctx := context.Background()
	nginxA, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
	})
	if err != nil {
		t.Error(err)
	}
	defer nginxA.Terminate(ctx, t)

	nginxB, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
	})
	if err != nil {
		t.Error(err)
	}
	defer nginxB.Terminate(ctx, t)

	ipA, err := nginxA.GetIPAddress(ctx)
	if err != nil {
		t.Error(err)
	}

	ipB, err := nginxB.GetIPAddress(ctx)
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s", ipA))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	resp, err = http.Get(fmt.Sprintf("http://%s", ipB))
	if err != nil {
		t.Error(err)
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
		t.Error(err)
	}
	defer nginxC.Terminate(ctx, t)
	ip, err := nginxC.GetIPAddress(ctx)
	if err != nil {
		t.Error(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s", ip))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}
