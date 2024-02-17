package smocker_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/smocker"
)

func TestSmocker(t *testing.T) {
	ctx := context.Background()

	container, err := smocker.RunContainer(ctx, testcontainers.WithImage("thiht/smocker:0.18.5"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	// apiURL {
	apiURL, err := container.ApiURL(ctx)
	// }
	if err != nil {
		t.Fatal(err)
	}

	testMock := `
- request:
    method: GET
    path: /test
  response:
    status: 200
    body: this is the reply from the mock
`
	_, err = http.Post(apiURL+"/mocks", "application/x-yaml", strings.NewReader(testMock))
	if err != nil {
		t.Fatal(err)
	}

	// mockURL {
	mockUrl, err := container.MockURL(ctx)
	// }
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(mockUrl + "/test")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected response status code to be 200, got %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(respBody) != "this is the reply from the mock" {
		t.Fatalf("expected response body to be 'this is the reply from the mock', got '%s'", respBody)
	}

	_, err = http.Post(apiURL+"/reset", "application/x-yaml", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Get(mockUrl + "/test")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode == 200 {
		t.Fatalf("expected response status code to be 404, got %d", resp.StatusCode)
	}
}
