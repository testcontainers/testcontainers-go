package mockserver_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	client "github.com/BraspagDevelopers/mock-server-client"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mockserver"
)

func ExampleRunContainer() {
	// runMockServerContainer {
	ctx := context.Background()

	mockserverContainer, err := mockserver.RunContainer(ctx, testcontainers.WithImage("mockserver/mockserver"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := mockserverContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := mockserverContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connect() {
	// connectToMockServer {
	ctx := context.Background()

	mockserverContainer, err := mockserver.RunContainer(ctx, testcontainers.WithImage("mockserver/mockserver"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := mockserverContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	host, err := mockserverContainer.GetHost(ctx)
	if err != nil {
		panic(err)
	}
	port, err := mockserverContainer.GetPort(ctx)
	if err != nil {
		panic(err)
	}
	ms := client.NewClient(host, port)

	requestMatcher := client.RequestMatcher{
		Method: http.MethodPost,
		Path:   "/api/categories",
	}
	requestMatcher = requestMatcher.WithJSONFields(map[string]interface{}{"name": "Tools"})
	err = ms.RegisterExpectation(client.NewExpectation(requestMatcher).WithResponse(client.NewResponseOK().WithJSONBody(map[string]any{"test": "value"})))
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Post("http://"+host+":"+fmt.Sprintf("%d", port)+"/api/categories", "application/json", strings.NewReader(`{"name": "Tools"}`))
	if err != nil {
		panic(err)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(string(buf))

	// Output:
	// 200
	// {
	//   "test" : "value"
	// }
}
