package valkey

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
			expectedCmds: []string{"/usr/local/valkey.conf"},
		},
		{
			name:         "existing commands",
			cmds:         []string{"a", "b", "c"},
			expectedCmds: []string{"/usr/local/valkey.conf", "a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: tt.cmds,
				},
			}

			err := WithConfigFile("valkey.conf")(req)
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
			expectedCmds: []string{"--loglevel", "debug"},
		},
		{
			name:         "existing commands",
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
			name:         "no existing command",
			cmds:         []string{},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{"--save", "60", "100"},
		},
		{
			name:         "existing commands",
			cmds:         []string{"a", "b", "c"},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{"a", "b", "c", "--save", "60", "100"},
		},
		{
			name:         "zero values get normalized",
			cmds:         []string{"a", "b", "c"},
			seconds:      0,
			changedKeys:  0,
			expectedCmds: []string{"a", "b", "c", "--save", "1", "1"},
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
