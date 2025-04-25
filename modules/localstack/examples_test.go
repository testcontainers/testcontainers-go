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

func ExampleRun() {
	// runLocalstackContainer {
	ctx := context.Background()

	localstackContainer, err := localstack.Run(ctx, "localstack/localstack:1.4.0")
	defer func() {
		if err := testcontainers.TerminateContainer(localstackContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := localstackContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withNetwork() {
	// localstackWithNetwork {
	ctx := context.Background()

	newNetwork, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %s", err)
		return
	}

	defer func() {
		if err := newNetwork.Remove(context.Background()); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()

	nwName := newNetwork.Name

	localstackContainer, err := localstack.Run(
		ctx,
		"localstack/localstack:0.13.0",
		testcontainers.WithEnv(map[string]string{"SERVICES": "s3,sqs"}),
		network.WithNetwork([]string{nwName}, newNetwork),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(localstackContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	networks, err := localstackContainer.Networks(ctx)
	if err != nil {
		log.Printf("failed to get container networks: %s", err)
		return
	}

	fmt.Println(len(networks))

	// Output:
	// 1
}

func ExampleRun_legacyMode() {
	ctx := context.Background()

	ctr, err := localstack.Run(
		ctx,
		"localstack/localstack:0.10.0",
		testcontainers.WithEnv(map[string]string{"SERVICES": "s3,sqs"}),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready.").WithStartupTimeout(5*time.Minute).WithOccurrence(1)),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err == nil {
		log.Printf("expected an error, got nil")
		return
	}

	fmt.Println(err)

	// Output:
	// version=localstack/localstack:0.10.0. Testcontainers for Go does not support running LocalStack in legacy mode. Please use a version >= 0.11.0
}

func ExampleRun_usingLambdas() {
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
	ctr, err := localstack.Run(ctx,
		"localstack/localstack:2.3.0",
		testcontainers.WithEnv(map[string]string{
			"SERVICES":            "lambda",
			"LAMBDA_DOCKER_FLAGS": flagsFn(),
		}),
		testcontainers.WithFiles(
			testcontainers.ContainerFile{
				HostFilePath:      filepath.Join("testdata", "function.zip"),
				ContainerFilePath: "/tmp/function.zip",
			},
		),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

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
		_, _, err := ctr.Exec(ctx, cmd)
		if err != nil {
			log.Printf("failed to execute command %v: %s", cmd, err)
			return
		}
	}

	// 4. get the URL for the lambda function
	cmd := []string{
		"awslocal", "lambda", "list-function-url-configs", "--function-name", lambdaName,
	}
	_, reader, err := ctr.Exec(ctx, cmd, exec.Multiplexed())
	if err != nil {
		log.Printf("failed to execute command %v: %s", cmd, err)
		return
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	if err != nil {
		log.Printf("failed to read from reader: %s", err)
		return
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
		log.Printf("failed to unmarshal content: %s", err)
		return
	}

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	functionURL := v.FunctionURLConfigs[0].FunctionURL
	// replace the port with the one exposed by the container

	mappedPort, err := ctr.MappedPort(ctx, "4566/tcp")
	if err != nil {
		log.Printf("failed to get mapped port: %s", err)
		return
	}

	functionURL = strings.ReplaceAll(functionURL, "4566", mappedPort.Port())

	resp, err := httpClient.Post(functionURL, "application/json", bytes.NewBufferString(`{"num1": "10", "num2": "10"}`))
	if err != nil {
		log.Printf("failed to send request to lambda function: %s", err)
		return
	}

	jsonResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response body: %s", err)
		return
	}

	fmt.Println(string(jsonResponse))

	// Output:
	// The product of 10 and 10 is 100
}
