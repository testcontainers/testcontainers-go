package testcontainer

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestContainerCreation(t *testing.T) {
	ctx := context.Background()
	nginxC, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			"80/tpc",
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
