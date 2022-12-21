package wait_test

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleExecStrategy() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:      "localstack/localstack:latest",
		WaitingFor: wait.ForExec([]string{"awslocal", "dynamodb", "list-tables"}),
	}

	localstack, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := localstack.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Here you have a running container
}

type mockExecTarget struct {
	waitDuration time.Duration
	successAfter time.Time
	exitCode     int
	failure      error
}

func (st mockExecTarget) Host(_ context.Context) (string, error) {
	return "", errors.New("not implemented")
}

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

	if err := ctx.Err(); err != nil {
		return st.exitCode, nil, err
	}

	if !st.successAfter.IsZero() && time.Now().After(st.successAfter) {
		return 0, nil, st.failure
	}

	return st.exitCode, nil, st.failure
}

func (st mockExecTarget) State(_ context.Context) (*types.ContainerState, error) {
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
	if err != context.DeadlineExceeded {
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
