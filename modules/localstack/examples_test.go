package localstack_test

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
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

func ExampleRunContainer_usingLambdas() {
	ctx := context.Background()

	flagsFn := func() string {
		labels := testcontainers.GenericLabels()

		flags := ""
		for k, v := range labels {
			flags = fmt.Sprintf("%s -l %s=%s", flags, k, v)
		}

		return flags
	}

	lambdaName := "localstack-lambda-url-example"

	container, err := localstack.RunContainer(ctx,
		testcontainers.WithImage("localstack/localstack:2.3.0"),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Env: map[string]string{
					"SERVICES":            "lambda",
					"LAMBDA_DOCKER_FLAGS": flagsFn(),
				},
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      filepath.Join("testdata", "function.zip"),
						ContainerFilePath: "/tmp/function.zip",
					},
				},
			},
		}),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			panic(err)
		}
	}()

	// the three commands below are doing the following:
	// 1. create a lambda function
	// 2. wait for the lambda function to be active
	// 3. invoke the lambda function with a payload, writing the result to the output.txt file
	lambdaCommands := [][]string{
		{
			"awslocal", "lambda",
			"create-function", "--function-name", lambdaName,
			"--runtime", "nodejs18.x",
			"--zip-file",
			"fileb:///tmp/function.zip",
			"--handler", "index.handler",
			"--role", "arn:aws:iam::000000000000:role/lambda-role",
		},
		{"awslocal", "lambda", "wait", "function-active-v2", "--function-name", lambdaName},
		{"awslocal", "lambda", "invoke", "--function-name", lambdaName, "--payload", `{"body": "{\"num1\": \"10\", \"num2\": \"10\"}" }`, "output.txt"},
	}
	for _, cmd := range lambdaCommands {
		_, _, err = container.Exec(ctx, cmd)
		if err != nil {
			panic(err)
		}
	}

	// the output.txt file lives in the WORKDIR of the localstack container
	_, reader, err := container.Exec(ctx, []string{"cat", "output.txt"}, exec.Multiplexed())
	if err != nil {
		panic(err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))

	// Output:
	// {"statusCode":200,"body":"The product of 10 and 10 is 100"}
}
