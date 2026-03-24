package wait_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleExecStrategy() {
	ctx := context.Background()
	ctr, err := testcontainers.Run(
		ctx, "alpine:latest",
		testcontainers.WithEntrypoint("tail", "-f", "/dev/null"),
		testcontainers.WithWaitStrategy(wait.ForExec([]string{"ls", "/"}).WithStartupTimeout(1*time.Second)),
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

	// Output:
	// true
}

func TestExecStrategyWaitUntilReady(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(0, nil, nil)
	wg := wait.NewExecStrategy([]string{"true"}).
		WithStartupTimeout(30 * time.Second)
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestExecStrategyWaitUntilReadyForExec(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(0, nil, nil)
	wg := wait.ForExec([]string{"true"})
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestExecStrategyWaitUntilReady_MultipleChecks(t *testing.T) {
	successAfter := time.Now().Add(2 * time.Second)
	target := newMockStrategyTarget(t)
	target.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, cmd []string, opts ...tcexec.ProcessOption) (int, io.Reader, error) {
			if time.Now().After(successAfter) {
				return 0, nil, nil
			}
			return 10, nil, nil
		},
	)
	wg := wait.NewExecStrategy([]string{"true"}).
		WithPollInterval(500 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestExecStrategyWaitUntilReady_DeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	target := newMockStrategyTarget(t)
	target.On("Exec", mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, cmd []string, opts ...tcexec.ProcessOption) (int, io.Reader, error) {
			time.Sleep(1 * time.Second)
			if err := ctx.Err(); err != nil {
				return 0, nil, err
			}
			return 0, nil, nil
		},
	)
	wg := wait.NewExecStrategy([]string{"true"})
	err := wg.WaitUntilReady(ctx, target)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestExecStrategyWaitUntilReady_CustomExitCode(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(10, nil, nil)
	wg := wait.NewExecStrategy([]string{"true"}).WithExitCodeMatcher(func(exitCode int) bool {
		return exitCode == 10
	})
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestExecStrategyWaitUntilReady_withExitCode(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(10, nil, nil)

	wg := wait.NewExecStrategy([]string{"true"}).WithExitCode(10)
	// Default is 60. Let's shorten that
	wg.WithStartupTimeout(time.Second * 2)
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)

	// Ensure we aren't spuriously returning on any code
	wg = wait.NewExecStrategy([]string{"true"}).WithExitCode(0)
	wg.WithStartupTimeout(time.Second * 2)
	err = wg.WaitUntilReady(context.Background(), target)
	require.Errorf(t, err, "Expected strategy to timeout out")
}

func TestExecStrategyWaitUntilReady_CustomResponseMatcher(t *testing.T) {
	// waitForExecExitCodeResponse {
	ctx := context.Background()
	ctr, err := testcontainers.Run(
		ctx, "nginx:latest",
		testcontainers.WithWaitStrategy(wait.ForExec([]string{"echo", "hello world!"}).
			WithStartupTimeout(time.Second*10).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}).
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Equal(data, []byte("hello world!\n"))
			}),
		),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}
