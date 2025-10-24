package dockermcpgateway_test

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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

func ExampleRun_connectMCPClient() {
	// run_mcp_gateway {
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
	// }

	// get_gateway {
	gatewayEndpoint, err := ctr.GatewayEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get gateway endpoint: %s", err)
		return
	}
	// }

	// connect_mcp_client {
	transport := mcp.NewSSEClientTransport(gatewayEndpoint, nil)

	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)

	cs, err := client.Connect(context.Background(), transport)
	if err != nil {
		log.Printf("Failed to connect to MCP gateway: %v", err)
		return
	}
	// }

	// list_tools {
	mcpTools, err := cs.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		log.Printf("Failed to list tools: %v", err)
		return
	}
	// }

	fmt.Println(len(mcpTools.Tools))
	fmt.Println(len(ctr.Tools()))
}
