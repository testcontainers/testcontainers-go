package testcontainers

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

func TestExecWithMultiplexedResponse(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image: nginxAlpineImage,
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	code, reader, err := container.Exec(ctx, []string{"ls", "/usr/share/nginx"}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Zero(t, code)
	require.NotNil(t, reader)

	b, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotNil(t, b)

	str := string(b)
	require.Equal(t, "html\n", str)
}

func TestExecWithOptions(t *testing.T) {
	tests := []struct {
		name string
		cmds []string
		opts []tcexec.ProcessOption
		want string
	}{
		{
			name: "with user",
			cmds: []string{"whoami"},
			opts: []tcexec.ProcessOption{
				tcexec.WithUser("nginx"),
			},
			want: "nginx\n",
		},
		{
			name: "with working dir",
			cmds: []string{"pwd"},
			opts: []tcexec.ProcessOption{
				tcexec.WithWorkingDir("/var/log/nginx"),
			},
			want: "/var/log/nginx\n",
		},
		{
			name: "with env",
			cmds: []string{"env"},
			opts: []tcexec.ProcessOption{
				tcexec.WithEnv([]string{"TEST_ENV=test"}),
			},
			want: "TEST_ENV=test\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := ContainerRequest{
				Image: nginxAlpineImage,
			}

			container, err := GenericContainer(ctx, GenericContainerRequest{
				ProviderType:     providerType,
				ContainerRequest: req,
				Started:          true,
			})

			require.NoError(t, err)
			terminateContainerOnEnd(t, ctx, container)

			// always append the multiplexed option for having the output
			// in a readable format
			tt.opts = append(tt.opts, tcexec.Multiplexed())

			code, reader, err := container.Exec(ctx, tt.cmds, tt.opts...)
			require.NoError(t, err)
			require.Zero(t, code)
			require.NotNil(t, reader)

			b, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.NotNil(t, b)

			str := string(b)
			require.Contains(t, str, tt.want)
		})
	}
}

func TestExecWithNonMultiplexedResponse(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image: nginxAlpineImage,
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	code, reader, err := container.Exec(ctx, []string{"ls", "/usr/share/nginx"})
	require.NoError(t, err)
	require.Zero(t, code)
	require.NotNil(t, reader)

	b, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotNil(t, b)

	str := string(b)
	require.True(t, strings.HasSuffix(str, "html\n"))
}
