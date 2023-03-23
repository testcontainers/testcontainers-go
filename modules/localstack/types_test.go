package localstack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOverrideContainerRequest(t *testing.T) {
	req := testcontainers.ContainerRequest{
		Env:          map[string]string{},
		Image:        "foo",
		ExposedPorts: []string{},
		WaitingFor: wait.ForNop(
			func(ctx context.Context, target wait.StrategyTarget) error {
				return nil
			},
		),
		Networks: []string{"foo", "bar", "baaz"},
		NetworkAliases: map[string][]string{
			"foo": {"foo0", "foo1", "foo2", "foo3"},
		},
	}

	merged := OverrideContainerRequest(testcontainers.ContainerRequest{
		Env: map[string]string{
			"FOO": "BAR",
		},
		Image:        "bar",
		ExposedPorts: []string{"12345/tcp"},
		Networks:     []string{"foo1", "bar1"},
		NetworkAliases: map[string][]string{
			"foo1": {"bar"},
		},
		WaitingFor: wait.ForLog("foo"),
	})(req)

	assert.Equal(t, "BAR", merged.Env["FOO"])
	assert.Equal(t, "bar", merged.Image)
	assert.Equal(t, []string{"12345/tcp"}, merged.ExposedPorts)
	assert.Equal(t, []string{"foo1", "bar1"}, merged.Networks)
	assert.Equal(t, []string{"foo0", "foo1", "foo2", "foo3"}, merged.NetworkAliases["foo"])
	assert.Equal(t, []string{"bar"}, merged.NetworkAliases["foo1"])
	assert.Equal(t, wait.ForLog("foo"), merged.WaitingFor)
}
