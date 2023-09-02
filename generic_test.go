package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	reusableContainerName = "my_test_reusable_container"
)

func TestGenericReusableContainer(t *testing.T) {
	ctx := context.Background()

	n1, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
			Name:         reusableContainerName,
		},
		Started: true,
	})
	require.NoError(t, err)
	require.True(t, n1.IsRunning())
	terminateContainerOnEnd(t, ctx, n1)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testdata/hello.sh", "/"+copiedFileName, 700)
	require.NoError(t, err)

	tests := []struct {
		name          string
		containerName string
		errorMatcher  func(err error) error
		reuseOption   bool
	}{
		{
			name: "reuse option with empty name",
			errorMatcher: func(err error) error {
				if errors.Is(err, ErrReuseEmptyName) {
					return nil
				}
				return err
			},
			reuseOption: true,
		},
		{
			name:          "container already exists (reuse=false)",
			containerName: reusableContainerName,
			errorMatcher: func(err error) error {
				if err == nil {
					return errors.New("expected error but got none")
				}
				return nil
			},
			reuseOption: false,
		},
		{
			name:          "success reusing",
			containerName: reusableContainerName,
			reuseOption:   true,
			errorMatcher: func(err error) error {
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n2, err := GenericContainer(ctx, GenericContainerRequest{
				ProviderType: providerType,
				ContainerRequest: ContainerRequest{
					Image:        nginxAlpineImage,
					ExposedPorts: []string{nginxDefaultPort},
					WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
					Name:         tc.containerName,
				},
				Started: true,
				Reuse:   tc.reuseOption,
			})

			require.NoError(t, tc.errorMatcher(err))

			if err == nil {
				c, _, err := n2.Exec(ctx, []string{"/bin/ash", copiedFileName})
				require.NoError(t, err)
				require.Zero(t, c)
			}
		})
	}
}

type testExecutable struct {
	cmds []string
}

func (t testExecutable) AsCommand() []string {
	return t.cmds
}

func TestWithStartupCommand(t *testing.T) {
	req := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:      "alpine",
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
		Started: true,
	}

	testExec := testExecutable{
		cmds: []string{"touch", "/tmp/.testcontainers"},
	}

	WithStartupCommand(testExec)(&req)

	c, err := GenericContainer(context.Background(), req)
	require.NoError(t, err)
	defer func() {
		err = c.Terminate(context.Background())
		require.NoError(t, err)
	}()

	_, reader, err := c.Exec(context.Background(), []string{"ls", "/tmp/.testcontainers"}, exec.Multiplexed())
	require.NoError(t, err)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/.testcontainers\n", string(content))
}

func TestGenericReusableContainerInSubprocess(t *testing.T) {
	containerIDOnce := sync.Once{}
	containerID := ""

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()

			// create containers in subprocesses, as "go test ./..." does.
			output := createReuseContainerInSubprocess(t)

			// check is container reused.
			re := regexp.MustCompile(fmt.Sprintf("%s(.*)%s",
				"🚧 Waiting for container id ",
				regexp.QuoteMeta(fmt.Sprintf(" image: %s", nginxDelayedImage)),
			))
			match := re.FindStringSubmatch(output)

			containerIDOnce.Do(func() {
				containerID = match[1]
			})
			require.Equal(t, containerID, match[1])
		}()
	}

	wg.Wait()
}

func createReuseContainerInSubprocess(t *testing.T) string {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperContainerStarterProcess")
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	return string(output)
}

func TestHelperContainerStarterProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxDelayedImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort), // default startupTimeout is 60s
			Name:         reusableContainerName,
		},
		Started: true,
		Reuse:   true,
	})
	require.NoError(t, err)
	require.True(t, nginxC.IsRunning())

	origin, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	// check is reuse container with WaitingFor work correctly.
	resp, err := http.Get(origin)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	os.Exit(0)
}
