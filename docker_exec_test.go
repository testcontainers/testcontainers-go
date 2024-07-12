package testcontainers

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

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
				ContainerRequest: req,
				Started:          true,
			})

			assert.NilError(t, err)
			terminateContainerOnEnd(t, ctx, container)

			// always append the multiplexed option for having the output
			// in a readable format
			tt.opts = append(tt.opts, tcexec.Multiplexed())

			code, reader, err := container.Exec(ctx, tt.cmds, tt.opts...)
			assert.NilError(t, err)
			assert.Equal(t, code, 0)
			assert.Assert(t, reader != nil)

			b, err := io.ReadAll(reader)
			assert.NilError(t, err)
			assert.Assert(t, b != nil)

			str := string(b)
			assert.Assert(t, is.Contains(str, tt.want))
		})
	}
}

func TestExecWithMultiplexedResponse(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image: nginxAlpineImage,
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	assert.NilError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	code, reader, err := container.Exec(ctx, []string{"sh", "-c", "echo stdout; echo stderr >&2"}, tcexec.Multiplexed())
	assert.NilError(t, err)
	assert.Equal(t, code, 0)
	assert.Assert(t, reader != nil)

	b, err := io.ReadAll(reader)
	assert.NilError(t, err)
	assert.Assert(t, b != nil)

	str := string(b)
	assert.Assert(t, is.Contains(str, "stdout"))
	assert.Assert(t, is.Contains(str, "stderr"))
}

func TestExecWithNonMultiplexedResponse(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image: nginxAlpineImage,
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	assert.NilError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	code, reader, err := container.Exec(ctx, []string{"sh", "-c", "echo stdout; echo stderr >&2"})
	assert.NilError(t, err)
	assert.Equal(t, code, 0)
	assert.Assert(t, reader != nil)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	written, err := stdcopy.StdCopy(&stdout, &stderr, reader)
	assert.NilError(t, err)
	assert.Assert(t, written != 0)
	assert.Assert(t, stdout.String() != "")
	assert.Assert(t, stderr.String() != "")

	assert.Equal(t, "stdout\n", stdout.String())
	assert.Equal(t, "stderr\n", stderr.String())
}
