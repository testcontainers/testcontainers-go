package mosquitto

import (
	"bytes"
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultMQTTPort   = "1883/tcp"
	defaultConfigPath = "/mosquitto/config/mosquitto.conf"
)

// defaultConfig is a minimal Mosquitto configuration that allows anonymous
// connections on the default MQTT port. The eclipse-mosquitto image requires
// a listener directive; without it the broker exits immediately.
var defaultConfig = []byte(`listener 1883
allow_anonymous true
`)

// Container represents the Mosquitto container type used in the module.
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Mosquitto container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultMQTTPort),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            bytes.NewReader(defaultConfig),
			ContainerFilePath: defaultConfigPath,
			FileMode:          0o644,
		}),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(defaultMQTTPort)),
	)

	// User-provided options are appended last so they can override defaults.
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run mosquitto: %w", err)
	}

	return c, nil
}

// BrokerURL returns the MQTT broker URL (mqtt://host:1883).
func (c *Container) BrokerURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultMQTTPort, "mqtt")
}

// WithConfigFile mounts a custom mosquitto.conf, replacing the default
// embedded configuration. The provided file must contain at least a
// "listener 1883" directive.
func WithConfigFile(cfgHostPath string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath:      cfgHostPath,
		ContainerFilePath: defaultConfigPath,
		FileMode:          0o644,
	})
}
