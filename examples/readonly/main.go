package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func main() {
	ctx := context.Background()

	// Example 1: Container with read-only root filesystem
	fmt.Println("=== Example 1: Read-only root filesystem ===")
	
	container, err := testcontainers.Run(ctx, "alpine:latest",
		testcontainers.WithReadOnlyRootFilesystem(),
		testcontainers.WithCmd("sh", "-c", "echo 'Attempting to write to root filesystem...' && echo 'test' > /test.txt && echo 'Write succeeded' || echo 'Write failed (expected)'"),
		testcontainers.WithWaitStrategy(wait.ForExit()),
	)
	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}()

	// Get the logs
	logs, err := container.Logs(ctx)
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	if err != nil {
		log.Fatalf("Failed to read logs: %v", err)
	}

	fmt.Printf("Container output:\n%s\n", string(logBytes))

	// Example 2: Read-only root filesystem with tmpfs for writable areas
	fmt.Println("=== Example 2: Read-only root filesystem with tmpfs ===")
	
	container2, err := testcontainers.Run(ctx, "alpine:latest",
		testcontainers.WithReadOnlyRootFilesystem(),
		testcontainers.WithTmpfs(map[string]string{"/tmp": "rw,noexec,nosuid,size=100m"}),
		testcontainers.WithCmd("sh", "-c", "echo 'Attempting to write to /tmp (tmpfs)...' && echo 'test' > /tmp/test.txt && echo 'Write to tmpfs succeeded' || echo 'Write to tmpfs failed'"),
		testcontainers.WithWaitStrategy(wait.ForExit()),
	)
	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container2); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}()

	// Get the logs
	logs2, err := container2.Logs(ctx)
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}
	defer logs2.Close()

	logBytes2, err := io.ReadAll(logs2)
	if err != nil {
		log.Fatalf("Failed to read logs: %v", err)
	}

	fmt.Printf("Container output:\n%s\n", string(logBytes2))

	// Verify the containers were configured correctly
	inspect1, err := container.Inspect(ctx)
	if err != nil {
		log.Fatalf("Failed to inspect container: %v", err)
	}

	inspect2, err := container2.Inspect(ctx)
	if err != nil {
		log.Fatalf("Failed to inspect container: %v", err)
	}

	fmt.Printf("Container 1 ReadonlyRootfs: %t\n", inspect1.HostConfig.ReadonlyRootfs)
	fmt.Printf("Container 2 ReadonlyRootfs: %t\n", inspect2.HostConfig.ReadonlyRootfs)
	fmt.Printf("Container 2 Tmpfs mounts: %v\n", inspect2.HostConfig.Tmpfs)
}