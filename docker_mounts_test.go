package testcontainers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/log"
)

type captureLogger struct {
	b *strings.Builder
}

func (c captureLogger) Printf(format string, v ...any) {
	fmt.Fprintf(c.b, format, v...)
}

func TestMapToDockerMountsBindDiagnosticFormat(t *testing.T) {
	// serial: mutates package log.Default()

	// MountType is a defined uint with no String method; the string verb yields a fmt error marker.
	badVerb := "%" + "s"
	require.Contains(t, fmt.Sprintf(badVerb, MountTypeBind), "%!s(")
	require.Equal(t, "0", fmt.Sprintf("%v", MountTypeBind))

	var buf strings.Builder
	prev := log.Default()
	log.SetDefault(captureLogger{b: &buf})
	t.Cleanup(func() { log.SetDefault(prev) })

	mounts := ContainerMounts{
		{
			Source: DockerBindMountSource{HostPath: "/host/path"},
			Target: "/container/path",
		},
	}
	_ = mounts.PrepareMounts()

	msg := buf.String()
	require.NotContains(t, msg, "%!s(")
	require.Contains(t, msg, "Mount type 0 is not supported by Testcontainers for Go")
}
