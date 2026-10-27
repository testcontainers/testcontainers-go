package trinogateway_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/trino"
	"github.com/testcontainers/testcontainers-go/modules/trinogateway"
	"github.com/testcontainers/testcontainers-go/network"
)

func ExampleRun() {
	// runTrinoGatewayContainer {
	ctx := context.Background()

	gatewayContainer, err := trinogateway.Run(ctx, "trinodb/trino-gateway:22")
	defer func() {
		if err := testcontainers.TerminateContainer(gatewayContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := gatewayContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_backends() {
	ctx := context.Background()

	gatewayContainer, err := trinogateway.Run(ctx, "trinodb/trino-gateway:22")
	defer func() {
		if err := testcontainers.TerminateContainer(gatewayContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// runTrinoBackend {
	// The Trino container is attached to the gateway network, and configured to process
	// the forwarded headers set by the gateway, so that the URLs for the next results of
	// a query point to the gateway instead of the Trino container.
	trinoContainer, err := trino.Run(ctx, "trinodb/trino:435",
		network.WithNetwork([]string{"trino"}, gatewayContainer.Network()),
		testcontainers.WithEnv(map[string]string{"JAVA_TOOL_OPTIONS": "-Dhttp-server.process-forwarded=true"}),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(trinoContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// addBackend {
	err = gatewayContainer.AddBackend(ctx, trinogateway.Backend{
		Name:    "trino",
		ProxyTo: "http://trino:8080",
	})
	if err != nil {
		log.Printf("failed to add backend: %s", err)
		return
	}

	waitCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	if err := gatewayContainer.WaitForBackendHealthy(waitCtx, "trino"); err != nil {
		log.Printf("failed to wait for backend: %s", err)
		return
	}
	// }

	// queryThroughGateway {
	// Any Trino client can be pointed to the gateway's connection string. Here the
	// Trino REST API is used directly: submit the query, then follow the nextUri links
	// until the results are complete. Those links point to the gateway, which routes
	// them to the backend running the query.
	connectionString, err := gatewayContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	rows, err := runQuery(ctx, connectionString, "SELECT 1")
	if err != nil {
		log.Printf("failed to run query: %s", err)
		return
	}
	// }

	backends, err := gatewayContainer.ActiveBackends(ctx)
	if err != nil {
		log.Printf("failed to list active backends: %s", err)
		return
	}

	fmt.Println(rows[0][0])
	fmt.Println(len(backends))

	// Output:
	// 1
	// 1
}

// runQuery {

// trinoResult is the subset of the Trino statement response used to fetch the query results.
// See https://trino.io/docs/current/develop/client-protocol.html
type trinoResult struct {
	NextURI string  `json:"nextUri"`
	Data    [][]any `json:"data"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// runQuery submits a query to a Trino endpoint and follows the nextUri links
// until the results are complete, returning the rows.
func runQuery(ctx context.Context, baseURL, query string) ([][]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/statement", strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Trino-User", "test")

	var rows [][]any
	for {
		result, err := fetchResult(req)
		if err != nil {
			return nil, err
		}

		rows = append(rows, result.Data...)

		if result.NextURI == "" {
			return rows, nil
		}

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, result.NextURI, nil)
		if err != nil {
			return nil, err
		}
	}
}

// fetchResult sends a request of the Trino client protocol and decodes the response.
func fetchResult(req *http.Request) (trinoResult, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return trinoResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return trinoResult{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return trinoResult{}, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result trinoResult
	if err := json.Unmarshal(body, &result); err != nil {
		return trinoResult{}, err
	}

	if result.Error != nil {
		return trinoResult{}, errors.New(result.Error.Message)
	}

	return result, nil
}

// }
