package mockserver_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	client "github.com/BraspagDevelopers/mock-server-client"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mockserver"
)

func ExampleRun() {
	// runMockServerContainer {
	ctx := context.Background()

	mockserverContainer, err := mockserver.Run(ctx, "mockserver/mockserver:5.15.0")
	defer func() {
		if err := testcontainers.TerminateContainer(mockserverContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := mockserverContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	// connectToMockServer {
	ctx := context.Background()

	mockserverContainer, err := mockserver.Run(ctx, "mockserver/mockserver:5.15.0")
	defer func() {
		if err := testcontainers.TerminateContainer(mockserverContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	url, err := mockserverContainer.URL(ctx)
	if err != nil {
		log.Printf("failed to get container URL: %s", err)
		return
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
		log.Printf("failed to register expectation: %s", err)
		return
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Post(url+"/api/categories", "application/json", strings.NewReader(`{"name": "Tools"}`))
	if err != nil {
		log.Printf("failed to send request: %s", err)
		return
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response: %s", err)
		return
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
