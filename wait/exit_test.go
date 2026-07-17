package wait_test

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWaitForExit(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.EXPECT().State(anyContext).
		Return(&container.State{Running: false}, nil)

	wg := wait.NewExitStrategy().WithExitTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}
