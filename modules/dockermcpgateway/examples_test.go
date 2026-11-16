package dockermcpgateway_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	dmcpg "github.com/testcontainers/testcontainers-go/modules/dockermcpgateway"
)

func ExampleRun() {
	ctx := context.Background()

	ctr, err := dmcpg.Run(
		ctx, "docker/mcp-gateway:latest",
		dmcpg.WithTools("curl", []string{"curl"}),
		dmcpg.WithTools("duckduckgo", []string{"fetch_content", "search"}),
		dmcpg.WithTools("github-official", []string{"add_issue_comment"}),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := ctr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)
	fmt.Println(len(ctr.Tools()))

	// Output:
	// true
	// 3
}
