package weaviate_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/grpc"

	"github.com/testcontainers/testcontainers-go"
	tcweaviate "github.com/testcontainers/testcontainers-go/modules/weaviate"
)

func ExampleRunContainer() {
	// runWeaviateContainer {
	ctx := context.Background()

	weaviateContainer, err := tcweaviate.RunContainer(ctx, testcontainers.WithImage("semitechnologies/weaviate:1.24.5"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := weaviateContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := weaviateContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connectWithClient() {
	// createClientNoModules {
	ctx := context.Background()

	weaviateContainer, err := tcweaviate.RunContainer(ctx, testcontainers.WithImage("semitechnologies/weaviate:1.23.9"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := weaviateContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	scheme, host, err := weaviateContainer.HttpHostAddress(ctx)
	if err != nil {
		log.Fatalf("failed to get http schema and host: %s", err) // nolint:gocritic
	}

	grpcHost, err := weaviateContainer.GrpcHostAddress(ctx)
	if err != nil {
		log.Fatalf("failed to get gRPC host: %s", err) // nolint:gocritic
	}

	connectionClient := &http.Client{}
	headers := map[string]string{
		// put here the custom API key, e.g. for OpenAPI
		"Authorization": fmt.Sprintf("Bearer %s", "custom-api-key"),
	}

	cli := weaviate.New(weaviate.Config{
		Scheme: scheme,
		Host:   host,
		GrpcConfig: &grpc.Config{
			Secured: false, // set true if gRPC connection is secured
			Host:    grpcHost,
		},
		Headers:          headers,
		AuthConfig:       nil, // put here the weaviate auth.Config, if you need it
		ConnectionClient: connectionClient,
	})

	err = cli.WaitForWeavaite(5 * time.Second)
	// }
	fmt.Println(err)

	// Output:
	// <nil>
}

func ExampleRunContainer_connectWithClientWithModules() {
	// createClientAndModules {
	ctx := context.Background()

	enableModules := []string{
		"backup-filesystem",
		"text2vec-openai",
		"text2vec-cohere",
		"text2vec-huggingface",
		"generative-openai",
	}
	envs := map[string]string{
		"ENABLE_MODULES":         strings.Join(enableModules, ","),
		"BACKUP_FILESYSTEM_PATH": "/tmp/backups",
	}

	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithImage("semitechnologies/weaviate:1.25.5"),
		testcontainers.WithEnv(envs),
	}

	weaviateContainer, err := tcweaviate.RunContainer(ctx, opts...)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := weaviateContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	scheme, host, err := weaviateContainer.HttpHostAddress(ctx)
	if err != nil {
		log.Fatalf("failed to get http schema and host: %s", err) // nolint:gocritic
	}

	grpcHost, err := weaviateContainer.GrpcHostAddress(ctx)
	if err != nil {
		log.Fatalf("failed to get gRPC host: %s", err) // nolint:gocritic
	}

	connectionClient := &http.Client{}
	headers := map[string]string{
		// put here the custom API key, e.g. for OpenAPI
		"Authorization": fmt.Sprintf("Bearer %s", "custom-api-key"),
	}

	cli := weaviate.New(weaviate.Config{
		Scheme: scheme,
		Host:   host,
		GrpcConfig: &grpc.Config{
			Secured: false, // set true if gRPC connection is secured
			Host:    grpcHost,
		},
		Headers:          headers,
		AuthConfig:       nil, // put here the weaviate auth.Config, if you need it
		ConnectionClient: connectionClient,
	})

	err = cli.WaitForWeavaite(5 * time.Second)
	// }
	fmt.Println(err)

	// Output:
	// <nil>
}
