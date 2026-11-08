package testcontainers

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

func sharedDaemonChildEnv(t *testing.T, extra ...string) []string {
	t.Helper()

	runtimeConfig, err := core.Runtime(t.Context())
	require.NoError(t, err)

	env := os.Environ()
	if runtimeConfig.Name == "sbx" {
		env = replaceChildEnv(env,
			"TESTCONTAINERS_PROVIDER=sbx-inherited",
			"DOCKER_HOST="+runtimeConfig.DockerHost,
			"TCGO_SBX_INHERITED_DOCKER_HOST="+runtimeConfig.DockerHost,
			"TCGO_SBX_INHERITED_SOCKET_PATH="+runtimeConfig.DockerSocketPath,
			"TCGO_SBX_INHERITED_PUBLISHED_HOST="+runtimeConfig.PublishedHost,
			"TCGO_SBX_INHERITED_BRIDGE="+runtimeConfig.DefaultBridgeNetwork,
		)
	}

	return replaceChildEnv(env, extra...)
}

func replaceChildEnv(env []string, replacements ...string) []string {
	for _, replacement := range replacements {
		key, _, _ := strings.Cut(replacement, "=")
		next := make([]string, 0, len(env)+1)
		replaced := false
		for _, value := range env {
			valueKey, _, _ := strings.Cut(value, "=")
			if valueKey != key {
				next = append(next, value)
				continue
			}
			if !replaced {
				next = append(next, replacement)
				replaced = true
			}
		}
		if !replaced {
			next = append(next, replacement)
		}
		env = next
	}
	return env
}
