package wait_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleExecStrategy() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:      "alpine:latest",
		Entrypoint: []string{"tail", "-f", "/dev/null"}, // needed for the container to stay alive
		WaitingFor: wait.ForExec([]string{"ls", "/"}).WithStartupTimeout(1 * time.Second),
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
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

	// Output:
	// true
}

type mockExecTarget struct {
	waitDuration time.Duration
	successAfter time.Time
	exitCode     int
	response     string
	failure      error
}

func (st mockExecTarget) Host(_ context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (st mockExecTarget) Inspect(ctx context.Context) (*types.ContainerJSON, error) {
	return nil, errors.New("not implemented")
}

// Deprecated: use Inspect instead
func (st mockExecTarget) Ports(ctx context.Context) (nat.PortMap, error) {
	return nil, errors.New("not implemented")
}

func (st mockExecTarget) MappedPort(_ context.Context, n nat.Port) (nat.Port, error) {
	return n, errors.New("not implemented")
}

func (st mockExecTarget) Logs(_ context.Context) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (st mockExecTarget) Exec(ctx context.Context, _ []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	time.Sleep(st.waitDuration)

	var reader io.Reader
	if st.response != "" {
		reader = bytes.NewReader([]byte(st.response))
	}

	if err := ctx.Err(); err != nil {
		return st.exitCode, reader, err
	}

	if !st.successAfter.IsZero() && time.Now().After(st.successAfter) {
		return 0, reader, st.failure
	}

	return st.exitCode, reader, st.failure
}

func (st mockExecTarget) State(_ context.Context) (*types.ContainerState, error) {
	return nil, errors.New("not implemented")
}

func (st mockExecTarget) CopyFileFromContainer(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func TestExecStrategyWaitUntilReady(t *testing.T) {
	target := mockExecTarget{}
	wg := wait.NewExecStrategy([]string{"true"}).
		WithStartupTimeout(30 * time.Second)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecStrategyWaitUntilReadyForExec(t *testing.T) {
	target := mockExecTarget{}
	wg := wait.ForExec([]string{"true"})
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecStrategyWaitUntilReady_MultipleChecks(t *testing.T) {
	target := mockExecTarget{
		exitCode:     10,
		successAfter: time.Now().Add(2 * time.Second),
	}
	wg := wait.NewExecStrategy([]string{"true"}).
		WithPollInterval(500 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecStrategyWaitUntilReady_DeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	target := mockExecTarget{
		waitDuration: 1 * time.Second,
	}
	wg := wait.NewExecStrategy([]string{"true"})
	err := wg.WaitUntilReady(ctx, target)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestExecStrategyWaitUntilReady_CustomExitCode(t *testing.T) {
	target := mockExecTarget{
		exitCode: 10,
	}
	wg := wait.NewExecStrategy([]string{"true"}).WithExitCodeMatcher(func(exitCode int) bool {
		return exitCode == 10
	})
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecStrategyWaitUntilReady_withExitCode(t *testing.T) {
	target := mockExecTarget{
		exitCode: 10,
	}
	wg := wait.NewExecStrategy([]string{"true"}).WithExitCode(10)
	// Default is 60. Let's shorten that
	wg.WithStartupTimeout(time.Second * 2)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure we aren't spuriously returning on any code
	wg = wait.NewExecStrategy([]string{"true"}).WithExitCode(0)
	wg.WithStartupTimeout(time.Second * 2)
	err = wg.WaitUntilReady(context.Background(), target)
	if err == nil {
		t.Fatalf("Expected strategy to timeout out")
	}
}

func TestExecStrategyWaitUntilReady_CustomResponseMatcher(t *testing.T) {
	// waitForExecExitCodeResponse {
	dockerReq := testcontainers.ContainerRequest{
		Image: "docker.io/nginx:latest",
		WaitingFor: wait.ForExec([]string{"echo", "hello world!"}).
			WithStartupTimeout(time.Second * 10).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}).
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("hello world!\n"))
			}),
	}
	// }

	ctx := context.Background()
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: dockerReq, Started: true})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	// }
}
