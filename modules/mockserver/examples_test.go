package mockserver_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	client "github.com/BraspagDevelopers/mock-server-client"

	"github.com/testcontainers/testcontainers-go/modules/mockserver"
)

func ExampleRun() {
	// runMockServerContainer {
	ctx := context.Background()

	mockserverContainer, err := mockserver.Run(ctx, "mockserver/mockserver:5.15.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := mockserverContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := mockserverContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	// connectToMockServer {
	ctx := context.Background()

	mockserverContainer, err := mockserver.Run(ctx, "mockserver/mockserver:5.15.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := mockserverContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	url, err := mockserverContainer.URL(ctx)
	if err != nil {
		log.Fatalf("failed to get container URL: %s", err) // nolint:gocritic
	}
	ms := client.NewClientURL(url)
	// }

	requestMatcher := client.RequestMatcher{
		Method: http.MethodPost,
		Path:   "/api/categories",
	}
	requestMatcher = requestMatcher.WithJSONFields(map[string]interface{}{"name": "Tools"})
	err = ms.RegisterExpectation(client.NewExpectation(requestMatcher).WithResponse(client.NewResponseOK().WithJSONBody(map[string]any{"test": "value"})))
	if err != nil {
		log.Fatalf("failed to register expectation: %s", err)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Post(url+"/api/categories", "application/json", strings.NewReader(`{"name": "Tools"}`))
	if err != nil {
		log.Fatalf("failed to send request: %s", err)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response: %s", err)
	}
	resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(string(buf))

	err = ms.Verify(requestMatcher, client.Once())
	fmt.Println(err == nil)

	// Output:
	// 200
	// {
	//   "test" : "value"
	// }
	// true
}
