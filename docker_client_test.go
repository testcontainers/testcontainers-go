package testcontainers

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetDockerInfo(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		ctx := context.Background()
		c, err := NewDockerClientWithOpts(ctx)
		assert.NilError(t, err)

		info, err := c.Info(ctx)
		assert.NilError(t, err)
		assert.Check(t, !reflect.ValueOf(info).IsZero())
	})

	t.Run("is goroutine safe", func(t *testing.T) {
		ctx := context.Background()
		c, err := NewDockerClientWithOpts(ctx)
		assert.NilError(t, err)

		count := 1024
		wg := sync.WaitGroup{}
		wg.Add(count)

		for i := 0; i < count; i++ {
			go func() {
				defer wg.Done()
				info, err := c.Info(ctx)
				assert.NilError(t, err)
				assert.Check(t, !reflect.ValueOf(info).IsZero())
			}()
		}
		wg.Wait()
	})
}
