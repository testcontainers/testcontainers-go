package testcontainers

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/stretchr/testify/require"

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

			ctr, err := Run(ctx, nginxAlpineImage)
			CleanupContainer(t, ctr)
			require.NoError(t, err)

			// always append the multiplexed option for having the output
			// in a readable format
			tt.opts = append(tt.opts, tcexec.Multiplexed())

			code, reader, err := ctr.Exec(ctx, tt.cmds, tt.opts...)
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

func TestExecReturnsNonZeroExitCodeAndOutput(t *testing.T) {
	// Regression for exit codes read before the output stream drained: a
	// created-but-not-yet-started exec reports {Running:false, ExitCode:null},
	// which the SDK flattens to 0, so an inspect landing in the pre-start window
	// returned 0 for a command that ultimately failed. Draining the stream to EOF
	// first makes the subsequent inspect a real terminal state. The race needs
	// daemon load to reproduce (~1 hit per few thousand concurrent execs); this
	// asserts the correct code and full output for a short-lived failing command.
	ctx := context.Background()

	ctr, err := Run(ctx, nginxAlpineImage)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	code, reader, err := ctr.Exec(ctx, []string{"sh", "-c", "echo MARKER; exit 7"}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Equal(t, 7, code)

	b, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Contains(t, string(b), "MARKER")
}

func TestExecWithMultiplexedResponse(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, nginxAlpineImage)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	code, reader, err := ctr.Exec(ctx, []string{"sh", "-c", "echo stdout; echo stderr >&2"}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Zero(t, code)
	require.NotNil(t, reader)

	b, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotNil(t, b)

	str := string(b)
	require.Contains(t, str, "stdout")
	require.Contains(t, str, "stderr")
}

func TestExecWithNonMultiplexedResponse(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, nginxAlpineImage)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	code, reader, err := ctr.Exec(ctx, []string{"sh", "-c", "echo stdout; echo stderr >&2"})
	require.NoError(t, err)
	require.Zero(t, code)
	require.NotNil(t, reader)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	written, err := stdcopy.StdCopy(&stdout, &stderr, reader)
	require.NoError(t, err)
	require.NotZero(t, written)
	require.NotNil(t, stdout)
	require.NotNil(t, stderr)

	require.Equal(t, "stdout\n", stdout.String())
	require.Equal(t, "stderr\n", stderr.String())
}
