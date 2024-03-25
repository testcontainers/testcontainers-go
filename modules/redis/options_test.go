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
			name:         "no existing command",
			cmds:         []string{},
			expectedCmds: []string{redisServerProcess, "/usr/local/redis.conf"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "/usr/local/redis.conf", "a", "b", "c"},
		},
		{
			name:         "non existing redis-server command",
			cmds:         []string{"a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "/usr/local/redis.conf", "a", "b", "c"},
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
			name:         "no existing command",
			cmds:         []string{},
			expectedCmds: []string{redisServerProcess, "--loglevel", "debug"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--loglevel", "debug"},
		},
		{
			name:         "non existing redis-server command",
			cmds:         []string{"a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--loglevel", "debug"},
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
			name:         "no existing command",
			cmds:         []string{},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{redisServerProcess, "--save", "60", "100"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--save", "60", "100"},
		},
		{
			name:         "non existing redis-server command",
			cmds:         []string{"a", "b", "c"},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--save", "60", "100"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			seconds:      0,
			changedKeys:  0,
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--save", "1", "1"},
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
