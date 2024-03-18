package localstack_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRunContainer() {
	// runLocalstackContainer {
	ctx := context.Background()

	localstackContainer, err := localstack.RunContainer(ctx,
		testcontainers.WithImage("localstack/localstack:1.4.0"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := localstackContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_withNetwork() {
	// localstackWithNetwork {
	ctx := context.Background()

	newNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		log.Fatalf("failed to create network: %s", err)
	}

	nwName := newNetwork.Name

	localstackContainer, err := localstack.RunContainer(
		ctx,
		testcontainers.WithImage("localstack/localstack:0.13.0"),
		testcontainers.WithEnv(map[string]string{"SERVICES": "s3,sqs"}),
		network.WithNetwork([]string{nwName}, newNetwork),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	// }

	// Clean up the container
	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	networks, err := localstackContainer.Networks(ctx)
	if err != nil {
		log.Fatalf("failed to get container networks: %s", err) // nolint:gocritic
	}

	fmt.Println(len(networks))

	// Output:
	// 1
}

func ExampleRunContainer_legacyMode() {
	ctx := context.Background()

	_, err := localstack.RunContainer(
		ctx,
		testcontainers.WithImage("localstack/localstack:0.10.0"),
		testcontainers.WithEnv(map[string]string{"SERVICES": "s3,sqs"}),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready.").WithStartupTimeout(5*time.Minute).WithOccurrence(1)),
	)
	if err == nil {
		log.Fatalf("expected an error, got nil")
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

	// withCustomContainerRequest {
	container, err := localstack.RunContainer(ctx,
		testcontainers.WithImage("localstack/localstack:2.3.0"),
		testcontainers.WithEnv(map[string]string{
			"SERVICES":            "lambda",
			"LAMBDA_DOCKER_FLAGS": flagsFn(),
		}),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      filepath.Join("testdata", "function.zip"),
						ContainerFilePath: "/tmp/function.zip",
					},
				},
			},
		}),
		// }
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// the three commands below are doing the following:
	// 1. create a lambda function
	// 2. create the URL function configuration for the lambda function
	// 3. wait for the lambda function to be active
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
		{"awslocal", "lambda", "create-function-url-config", "--function-name", lambdaName, "--auth-type", "NONE"},
		{"awslocal", "lambda", "wait", "function-active-v2", "--function-name", lambdaName},
	}
	for _, cmd := range lambdaCommands {
		_, _, err := container.Exec(ctx, cmd)
		if err != nil {
			log.Fatalf("failed to execute command %v: %s", cmd, err) // nolint:gocritic
		}
	}

	// 4. get the URL for the lambda function
	cmd := []string{
		"awslocal", "lambda", "list-function-url-configs", "--function-name", lambdaName,
	}
	_, reader, err := container.Exec(ctx, cmd, exec.Multiplexed())
	if err != nil {
		log.Fatalf("failed to execute command %v: %s", cmd, err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	if err != nil {
		log.Fatalf("failed to read from reader: %s", err)
	}

	content := buf.Bytes()

	type FunctionURLConfig struct {
		FunctionURLConfigs []struct {
			FunctionURL      string `json:"FunctionUrl"`
			FunctionArn      string `json:"FunctionArn"`
			CreationTime     string `json:"CreationTime"`
			LastModifiedTime string `json:"LastModifiedTime"`
			AuthType         string `json:"AuthType"`
		} `json:"FunctionUrlConfigs"`
	}

	v := &FunctionURLConfig{}
	err = json.Unmarshal(content, v)
	if err != nil {
		log.Fatalf("failed to unmarshal content: %s", err)
	}

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	functionURL := v.FunctionURLConfigs[0].FunctionURL
	// replace the port with the one exposed by the container

	mappedPort, err := container.MappedPort(ctx, "4566/tcp")
	if err != nil {
		log.Fatalf("failed to get mapped port: %s", err)
	}

	functionURL = strings.ReplaceAll(functionURL, "4566", mappedPort.Port())

	resp, err := httpClient.Post(functionURL, "application/json", bytes.NewBufferString(`{"num1": "10", "num2": "10"}`))
	if err != nil {
		log.Fatalf("failed to send request to lambda function: %s", err)
	}

	jsonResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %s", err)
	}

	fmt.Println(string(jsonResponse))

	// Output:
	// The product of 10 and 10 is 100
}
