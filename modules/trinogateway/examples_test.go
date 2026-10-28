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
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/trino"
	"github.com/testcontainers/testcontainers-go/modules/trinogateway"
	"github.com/testcontainers/testcontainers-go/network"
)

func ExampleRun() {
	// runTrinoGatewayContainer {
	ctx := context.Background()

	gatewayContainer, err := trinogateway.Run(ctx, "trinodb/trino-gateway:22",
		trinogateway.WithDataStore(trinogateway.NewManagedPostgresProvider("postgres:17-alpine")),
	)
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

	gatewayContainer, err := trinogateway.Run(ctx, "trinodb/trino-gateway:22",
		trinogateway.WithDataStore(trinogateway.NewManagedPostgresProvider("postgres:17-alpine")),
	)
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
	// Trino must process the forwarded headers, so that nextUri points to the gateway.
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
	// Any Trino client can use the connection string, runQuery calls the REST API directly.
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

func ExampleRun_externalDataStore() {
	ctx := context.Background()

	// runWithExternalDatabase {
	// The database is owned by the caller, and the gateway joins its network.
	nw, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %s", err)
		return
	}
	defer func() {
		if err := nw.Remove(context.Background()); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()

	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("gateway"),
		postgres.WithUsername("gateway"),
		postgres.WithPassword("secret"),
		postgres.BasicWaitStrategies(),
		network.WithNetwork([]string{"db"}, nw),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	gatewayContainer, err := trinogateway.Run(ctx, "trinodb/trino-gateway:22",
		trinogateway.WithDataStore(trinogateway.NewExternalProvider(
			"jdbc:postgresql://db:5432/gateway", "gateway", "secret", "org.postgresql.Driver",
		)),
		network.WithNetwork([]string{"gateway"}, nw),
	)
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

	// Terminating the gateway leaves the database running.
	if err := gatewayContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %s", err)
		return
	}

	state, err := pgContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

// runQuery {

// trinoResult is part of a Trino client protocol response, see
// https://trino.io/docs/current/develop/client-protocol.html
type trinoResult struct {
	NextURI string  `json:"nextUri"`
	Data    [][]any `json:"data"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// runQuery runs a query and follows the nextUri links until all the rows are fetched.
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

// fetchResult sends a request and decodes the response.
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
