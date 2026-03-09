package redis

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestWithConfigFile(t *testing.T) {
	tests := []struct {
		name         string
		cmds         []string
		expectedCmds []string
	}{
		{
			name:         "empty-cmd",
			cmds:         []string{},
			expectedCmds: []string{"/usr/local/redis.conf"},
		},
		{
			name:         "existing-cmd",
			cmds:         []string{"a", "b", "c"},
			expectedCmds: []string{"/usr/local/redis.conf", "a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: tt.cmds,
				},
			}

			err := WithConfigFile("redis.conf")(req)
			require.NoError(t, err)

			require.Equal(t, tt.expectedCmds, req.Cmd)
		})
	}
}

func TestWithLogLevel(t *testing.T) {
	tests := []struct {
		name         string
		cmds         []string
		expectedCmds []string
	}{
		{
			name:         "empty-cmd",
			cmds:         []string{},
			expectedCmds: []string{"--loglevel", "debug"},
		},
		{
			name:         "existing-cmd",
			cmds:         []string{"a", "b", "c"},
			expectedCmds: []string{"a", "b", "c", "--loglevel", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: tt.cmds,
				},
			}

			err := WithLogLevel(LogLevelDebug)(req)
			require.NoError(t, err)

			require.Equal(t, tt.expectedCmds, req.Cmd)
		})
	}
}

func TestWithSnapshotting(t *testing.T) {
	tests := []struct {
		name         string
		cmds         []string
		expectedCmds []string
		seconds      int
		changedKeys  int
	}{
		{
			name:         "empty-cmd",
			cmds:         []string{},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{"--save", "60", "100"},
		},
		{
			name:         "existing-cmd",
			cmds:         []string{"a", "b", "c"},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{"a", "b", "c", "--save", "60", "100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: tt.cmds,
				},
			}

			err := WithSnapshotting(tt.seconds, tt.changedKeys)(req)
			require.NoError(t, err)

			require.Equal(t, tt.expectedCmds, req.Cmd)
		})
	}
}
