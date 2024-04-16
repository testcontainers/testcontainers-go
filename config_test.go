package testcontainers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/config"
)

// unset environment variables to avoid side effects
// execute this function before each test
func resetTestEnv(t *testing.T) {
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "")
	t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "")
}

func TestReadConfig(t *testing.T) {
	resetTestEnv(t)
	t.Cleanup(func() {
		config.Reset()
	})

	t.Run("Config is read just once", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support

		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		cfg1 := testcontainers.ReadConfig()

		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "false")
		cfg2 := testcontainers.ReadConfig()

		assert.Equal(t, cfg1, cfg2)
	})
}
