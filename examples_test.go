package testcontainers_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun() {
	ctx := context.Background()

	nw, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %s", err)
		return
	}
	defer func() {
		if err := nw.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()

	testFileContent := "Hello from file!"

	ctr, err := testcontainers.Run(
		ctx,
		"nginx:alpine",
		network.WithNetwork([]string{"nginx-alias"}, nw),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            strings.NewReader(testFileContent),
			ContainerFilePath: "/tmp/file.txt",
			FileMode:          0o644,
		}),
		testcontainers.WithTmpfs(map[string]string{
			"/tmp": "rw",
		}),
		testcontainers.WithLabels(map[string]string{
			"testcontainers.label": "true",
		}),
		testcontainers.WithEnv(map[string]string{
			"TEST": "true",
		}),
		testcontainers.WithExposedPorts("80/tcp"),
		testcontainers.WithAfterReadyCommand(testcontainers.NewRawCommand([]string{"echo", "hello", "world"})),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("80/tcp").WithStartupTimeout(time.Second*5)),
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

	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		log.Printf("failed to create docker client: %s", err)
		return
	}

	ctrResp, err := cli.ContainerInspect(ctx, ctr.GetContainerID())
	if err != nil {
		log.Printf("failed to inspect container: %s", err)
		return
	}

	// networks
	respNw, ok := ctrResp.NetworkSettings.Networks[nw.Name]
	if !ok {
		log.Printf("network not found")
		return
	}
	fmt.Println(respNw.Aliases)

	// env
	fmt.Println(ctrResp.Config.Env[0])

	// tmpfs
	tmpfs, ok := ctrResp.HostConfig.Tmpfs["/tmp"]
	if !ok {
		log.Printf("tmpfs not found")
		return
	}
	fmt.Println(tmpfs)

	// labels
	fmt.Println(ctrResp.Config.Labels["testcontainers.label"])

	// files
	f, err := ctr.CopyFileFromContainer(ctx, "/tmp/file.txt")
	if err != nil {
		log.Printf("failed to copy file from container: %s", err)
		return
	}

	content, err := io.ReadAll(f)
	if err != nil {
		log.Printf("failed to read file: %s", err)
		return
	}
	fmt.Println(string(content))

	// Output:
	// true
	// [nginx-alias]
	// TEST=true
	// rw
	// true
	// Hello from file!
}
