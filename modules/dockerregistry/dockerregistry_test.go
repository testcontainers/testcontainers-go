package dockerregistry

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestDockerRegistry(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	port, err := container.MappedPort(ctx, "5000")
	if err != nil {
		t.Fatal(err)
	}

	ipAddress, err := container.Host(ctx)

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	// perform assertions
	resp, err := http.Get("http://" + ipAddress + ":" + port.Port() + "/v2/_catalog")
	if err != nil {
		// handle err
		t.Fatal(err)
	}
	defer resp.Body.Close()
}
