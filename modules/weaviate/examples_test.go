package weaviate_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/grpc"

	"github.com/testcontainers/testcontainers-go"
	tcweaviate "github.com/testcontainers/testcontainers-go/modules/weaviate"
)

func ExampleRun() {
	// runWeaviateContainer {
	ctx := context.Background()

	weaviateContainer, err := tcweaviate.Run(ctx, "semitechnologies/weaviate:1.29.0")
	defer func() {
		if err := testcontainers.TerminateContainer(weaviateContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := weaviateContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connectWithClient() {
	// createClientNoModules {
	ctx := context.Background()

	weaviateContainer, err := tcweaviate.Run(ctx, "semitechnologies/weaviate:1.28.7")
	defer func() {
		if err := testcontainers.TerminateContainer(weaviateContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	scheme, host, err := weaviateContainer.HttpHostAddress(ctx)
	if err != nil {
		log.Printf("failed to get http schema and host: %s", err)
		return
	}

	grpcHost, err := weaviateContainer.GrpcHostAddress(ctx)
	if err != nil {
		log.Printf("failed to get gRPC host: %s", err)
		return
	}

	connectionClient := &http.Client{}
	headers := map[string]string{
		// put here the custom API key, e.g. for OpenAPI
		"Authorization": "Bearer custom-api-key",
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

func ExampleRun_connectWithClientWithModules() {
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
		testcontainers.WithEnv(envs),
	}

	weaviateContainer, err := tcweaviate.Run(ctx, "semitechnologies/weaviate:1.25.5", opts...)
	defer func() {
		if err := testcontainers.TerminateContainer(weaviateContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	scheme, host, err := weaviateContainer.HttpHostAddress(ctx)
	if err != nil {
		log.Printf("failed to get http schema and host: %s", err)
		return
	}

	grpcHost, err := weaviateContainer.GrpcHostAddress(ctx)
	if err != nil {
		log.Printf("failed to get gRPC host: %s", err)
		return
	}

	connectionClient := &http.Client{}
	headers := map[string]string{
		// put here the custom API key, e.g. for OpenAPI
		"Authorization": "Bearer custom-api-key",
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
