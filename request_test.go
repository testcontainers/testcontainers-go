package testcontainers

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func TestMergePortBindings(t *testing.T) {
	type arg struct {
		configPortMap nat.PortMap
		parsedPortMap nat.PortMap
		exposedPorts  []string
	}
	cases := []struct {
		name     string
		arg      arg
		expected nat.PortMap
	}{
		{
			name: "empty ports",
			arg: arg{
				configPortMap: nil,
				parsedPortMap: nil,
				exposedPorts:  nil,
			},
			expected: map[nat.Port][]nat.PortBinding{},
		},
		{
			name: "config port map but not exposed",
			arg: arg{
				configPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "1", HostPort: "2"}},
				},
				parsedPortMap: nil,
				exposedPorts:  nil,
			},
			expected: map[nat.Port][]nat.PortBinding{},
		},
		{
			name: "parsed port map without config",
			arg: arg{
				configPortMap: nil,
				parsedPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "", HostPort: ""}},
				},
				exposedPorts: nil,
			},
			expected: map[nat.Port][]nat.PortBinding{
				"80/tcp": {{HostIP: "", HostPort: ""}},
			},
		},
		{
			name: "parsed and configured but not exposed",
			arg: arg{
				configPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "1", HostPort: "2"}},
				},
				parsedPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "", HostPort: ""}},
				},
				exposedPorts: nil,
			},
			expected: map[nat.Port][]nat.PortBinding{
				"80/tcp": {{HostIP: "", HostPort: ""}},
			},
		},
		{
			name: "merge both parsed and config",
			arg: arg{
				configPortMap: map[nat.Port][]nat.PortBinding{
					"60/tcp": {{HostIP: "1", HostPort: "2"}},
					"70/tcp": {{HostIP: "1", HostPort: "2"}},
					"80/tcp": {{HostIP: "1", HostPort: "2"}},
				},
				parsedPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "", HostPort: ""}},
					"90/tcp": {{HostIP: "", HostPort: ""}},
				},
				exposedPorts: []string{"70", "80/tcp"},
			},
			expected: map[nat.Port][]nat.PortBinding{
				"70/tcp": {{HostIP: "1", HostPort: "2"}},
				"80/tcp": {{HostIP: "1", HostPort: "2"}},
				"90/tcp": {{HostIP: "", HostPort: ""}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := mergePortBindings(c.arg.configPortMap, c.arg.parsedPortMap, c.arg.exposedPorts)
			require.Equal(t, c.expected, res)
		})
	}
}

func TestCustomLabelsImage(t *testing.T) {
	const (
		myLabelName  = "org.my.label"
		myLabelValue = "my-label-value"
	)

	ctx := context.Background()
	req := Request{
		Image:  "alpine:latest",
		Labels: map[string]string{myLabelName: myLabelValue},
	}

	ctr, err := Run(ctx, req)

	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ctr.Terminate(ctx)) })

	ctrJSON, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, myLabelValue, ctrJSON.Config.Labels[myLabelName])
}

func TestCustomLabelsBuildOptionsModifier(t *testing.T) {
	const (
		myLabelName        = "org.my.label"
		myLabelValue       = "my-label-value"
		myBuildOptionLabel = "org.my.bo.label"
		myBuildOptionValue = "my-bo-label-value"
	)

	ctx := context.Background()
	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata",
			Dockerfile: "Dockerfile",
			BuildOptionsModifier: func(opts *types.ImageBuildOptions) {
				opts.Labels = map[string]string{
					myBuildOptionLabel: myBuildOptionValue,
				}
			},
		},
		Labels: map[string]string{myLabelName: myLabelValue},
	}

	ctr, err := Run(ctx, req)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	ctrJSON, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, myLabelValue, ctrJSON.Config.Labels[myLabelName])
	require.Equal(t, myBuildOptionValue, ctrJSON.Config.Labels[myBuildOptionLabel])
}
