package testcontainers

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDockerInfo(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		ctx := context.Background()
		c, err := NewDockerClientWithOpts(ctx)
		require.NoError(t, err)

		info, err := c.Info(ctx)
		require.NoError(t, err)
		require.NotNil(t, info)
	})

	t.Run("is goroutine safe", func(t *testing.T) {
		ctx := context.Background()
		c, err := NewDockerClientWithOpts(ctx)
		require.NoError(t, err)

		count := 1024
		wg := sync.WaitGroup{}
		wg.Add(count)

		for i := 0; i < count; i++ {
			go func() {
				defer wg.Done()
				info, err := c.Info(ctx)
				require.NoError(t, err)
				require.NotNil(t, info)
			}()
		}
		wg.Wait()
	})
}
