package nginx

import (
	"context"
	"net/http"
	"testing"
)

func TestIntegrationNginxLatestReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nginxC, err := setupNginx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		t.Log("terminating container")
		if err := nginxC.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: :%v", err)
		}
	})

	resp, err := http.Get(nginxC.URI)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}
