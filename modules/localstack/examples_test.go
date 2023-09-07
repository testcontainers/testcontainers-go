package localstack_test

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRunContainer() {
	// runLocalstackContainer {
	ctx := context.Background()

	localstackContainer, err := localstack.RunContainer(ctx,
		testcontainers.WithImage("localstack/localstack:1.4.0"),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := localstackContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_withNetwork() {
	// localstackWithNetwork {
	ctx := context.Background()

	nwName := "localstack-network"

	_, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: nwName,
		},
	})
	if err != nil {
		panic(err)
	}

	localstackContainer, err := localstack.RunContainer(
		ctx,
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:          "localstack/localstack:0.13.0",
				Env:            map[string]string{"SERVICES": "s3,sqs"},
				Networks:       []string{nwName},
				NetworkAliases: map[string][]string{nwName: {"localstack"}},
			},
		}),
	)
	if err != nil {
		panic(err)
	}
	// }

	// Clean up the container
	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	networks, err := localstackContainer.Networks(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(networks))
	fmt.Println(networks[0])

	// Output:
	// 1
	// localstack-network
}

func ExampleRunContainer_legacyMode() {
	ctx := context.Background()

	_, err := localstack.RunContainer(
		ctx,
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:      "localstack/localstack:0.10.0",
				Env:        map[string]string{"SERVICES": "s3,sqs"},
				WaitingFor: wait.ForLog("Ready.").WithStartupTimeout(5 * time.Minute).WithOccurrence(1),
			},
		}),
	)
	if err == nil {
		panic(err)
	}

	fmt.Println(err)

	// Output:
	// version=localstack/localstack:0.10.0. Testcontainers for Go does not support running LocalStack in legacy mode. Please use a version >= 0.11.0
}
