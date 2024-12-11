package k6_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k6"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun() {
	// runHTTPBin {
	ctx := context.Background()

	// create a container with the httpbin application that will be the target
	// for the test script that runs in the k6 container
	gcr := testcontainers.GenericContainerRequest{
		ProviderType: testcontainers.ProviderDocker,
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "kennethreitz/httpbin",
			ExposedPorts: []string{
				"80",
			},
			WaitingFor: wait.ForExposedPort(),
		},
		Started: true,
	}
	httpbin, err := testcontainers.GenericContainer(ctx, gcr)
	defer func() {
		if err := testcontainers.TerminateContainer(httpbin); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// getHTTPBinIP {
	httpbinIP, err := httpbin.ContainerIP(ctx)
	if err != nil {
		log.Printf("failed to get container IP: %s", err)
		return
	}
	// }

	absPath, err := filepath.Abs(filepath.Join("scripts", "httpbin.js"))
	if err != nil {
		log.Printf("failed to get absolute path to test script: %s", err)
		return
	}

	// runK6Container {
	// run the httpbin.js test scripts passing the IP address the httpbin container
	k6, err := k6.Run(
		ctx,
		"szkiba/k6x:v0.3.1",
		k6.WithCache(),
		k6.WithTestScript(absPath),
		k6.SetEnvVar("HTTPBIN", httpbinIP),
	)
	defer func() {
		cacheMount, err := k6.CacheMount(ctx)
		if err != nil {
			log.Printf("failed to determine cache mount: %s", err)
		}

		var options []testcontainers.TerminateOption
		if cacheMount != "" {
			options = append(options, testcontainers.RemoveVolumes(cacheMount))
		}

		if err = testcontainers.TerminateContainer(k6, options...); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	//}

	// assert the result of the test
	state, err := k6.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.ExitCode)
	// Output: 0
}
