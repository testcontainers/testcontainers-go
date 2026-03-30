package wait_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWaitForExit(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).Return(&container.State{Running: false}, nil)
	wg := wait.NewExitStrategy().WithExitTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}
