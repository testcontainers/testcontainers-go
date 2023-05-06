package redpanda

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	//go:embed mounts/redpanda.yaml.tpl
	nodeConfigTpl string

	//go:embed mounts/bootstrap.yaml.tpl
	bootstrapConfigTpl string

	//go:embed mounts/entrypoint-tc.sh
	entrypoint []byte

	defaultKafkaAPIPort       = "9092/tcp"
	defaultAdminAPIPort       = "9644/tcp"
	defaultSchemaRegistryPort = "8081/tcp"
)

// Container represents the Redpanda container type used in the module
type Container struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Redpanda container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// 1. Create container request.
	// Some (e.g. Image) may be overridden by providing an option argument to this function.
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.redpanda.com/redpandadata/redpanda:v23.1.7",
			User:  "root:root",
			// Files: Will be added later after we've rendered our YAML templates.
			ExposedPorts: []string{
				defaultKafkaAPIPort,
				defaultAdminAPIPort,
				defaultSchemaRegistryPort,
			},
			Entrypoint: []string{},
			Cmd: []string{
				"/entrypoint-tc.sh",
				"redpanda",
				"start",
				"--mode=dev-container",
			},
		},
		Started: true,
	}

	// 2. Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		opt.Customize(&req)
	}

	// 3. Create temporary entrypoint file. We need a custom entrypoint that waits
	// until the actual Redpanda node config is mounted. Once the redpanda config is
	// mounted we will call the original entrypoint with the same parameters.
	// We have to do this kind of two-step process, because we need to know the mapped
	// port, so that we can use this in Redpanda's advertised listeners configuration for
	// the Kafka API.
	entrypointFile, err := createEntrypointTmpFile()
	if err != nil {
		return nil, fmt.Errorf("failed to create entrypoint file: %w", err)
	}

	// Bootstrap config file contains cluster configurations which will only be considered
	// the very first time you start a cluster.
	bootstrapConfigFile, err := createBootstrapConfigFile(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap config file: %w", err)
	}

	toBeMountedFiles := []testcontainers.ContainerFile{
		{
			HostFilePath:      entrypointFile.Name(),
			ContainerFilePath: "/entrypoint-tc.sh",
			FileMode:          700,
		},
		{
			HostFilePath:      bootstrapConfigFile.Name(),
			ContainerFilePath: "/etc/redpanda/.bootstrap.yaml",
			FileMode:          700,
		},
	}
	req.Files = append(req.Files, toBeMountedFiles...)

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. Get mapped port for the Kafka API, so that we can render and then mount
	// the Redpanda config with the advertised Kafka address.
	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	kafkaPort, err := container.MappedPort(ctx, nat.Port(defaultKafkaAPIPort))
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped Kafka port: %w", err)
	}

	// 5. Render redpanda.yaml config and mount it.
	nodeConfig, err := renderNodeConfig(settings, hostIP, kafkaPort.Int())
	if err != nil {
		return nil, fmt.Errorf("failed to render node config: %w", err)
	}

	err = container.CopyToContainer(
		ctx,
		nodeConfig,
		"/etc/redpanda/redpanda.yaml",
		700,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to copy redpanda.yaml into container: %w", err)
	}

	// 6. Wait until Redpanda is ready to serve requests
	err = wait.ForLog("Successfully started Redpanda!").
		WithPollInterval(100*time.Millisecond).
		WaitUntilReady(ctx, container)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for Redpanda readiness: %w", err)
	}

	// 7. Create Redpanda Service Accounts if configured to do so.
	if len(settings.ServiceAccounts) > 0 {
		adminAPIPort, err := container.MappedPort(ctx, nat.Port(defaultAdminAPIPort))
		if err != nil {
			return nil, fmt.Errorf("failed to get mapped Admin API port: %w", err)
		}

		adminAPIUrl := fmt.Sprintf("http://%v:%d", hostIP, adminAPIPort.Int())
		adminCl := NewAdminAPIClient(adminAPIUrl)

		for username, password := range settings.ServiceAccounts {
			if err := adminCl.CreateUser(ctx, username, password); err != nil {
				return nil, fmt.Errorf("failed to create service account with username %q: %w", username, err)
			}
		}
	}

	return &Container{Container: container}, nil
}

// KafkaSeedBroker returns the seed broker that should be used for connecting
// to the Kafka API with your Kafka client. It'll be returned in the format:
// "host:port" - for example: "localhost:55687".
func (c *Container) KafkaSeedBroker(ctx context.Context) (string, error) {
	return c.getMappedHostPort(ctx, nat.Port(defaultKafkaAPIPort))
}

// AdminAPIAddress returns the address to the Redpanda Admin API. This
// is an HTTP-based API and thus the returned format will be: http://host:port.
func (c *Container) AdminAPIAddress(ctx context.Context) (string, error) {
	hostPort, err := c.getMappedHostPort(ctx, nat.Port(defaultAdminAPIPort))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%v", hostPort), nil
}

// SchemaRegistryAddress returns the address to the schema registry API. This
// is an HTTP-based API and thus the returned format will be: http://host:port.
func (c *Container) SchemaRegistryAddress(ctx context.Context) (string, error) {
	hostPort, err := c.getMappedHostPort(ctx, nat.Port(defaultSchemaRegistryPort))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%v", hostPort), nil
}

// getMappedHostPort returns the mapped host and port a given nat.Port following
// this format: "host:port". The mapped port is the port that is accessible from
// the host system and is remapped to the given container port.
func (c *Container) getMappedHostPort(ctx context.Context, port nat.Port) (string, error) {
	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get hostIP: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	return fmt.Sprintf("%v:%d", hostIP, mappedPort.Int()), nil
}

// createEntrypointTmpFile returns a temporary file with the custom entrypoint
// that awaits the actual Redpanda config after the container has been started,
// before it's going to start the Redpanda process.
func createEntrypointTmpFile() (*os.File, error) {
	entrypointTmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(entrypointTmpFile.Name(), entrypoint, 0o700); err != nil {
		return nil, err
	}

	return entrypointTmpFile, nil
}

// createBootstrapConfigFile renders the config template for the .bootstrap.yaml config,
// which configures Redpanda's cluster properties.
// Reference: https://docs.redpanda.com/docs/reference/cluster-properties/
func createBootstrapConfigFile(settings options) (*os.File, error) {
	bootstrapTplParams := redpandaBootstrapConfigTplParams{
		Superusers:                  settings.Superusers,
		KafkaAPIEnableAuthorization: settings.KafkaEnableAuthorization,
	}

	tpl, err := template.New("bootstrap.yaml").Parse(bootstrapConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redpanda config file template: %w", err)
	}

	var bootstrapConfig bytes.Buffer
	if err := tpl.Execute(&bootstrapConfig, bootstrapTplParams); err != nil {
		return nil, fmt.Errorf("failed to render redpanda bootstrap config template: %w", err)
	}

	bootstrapTmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(bootstrapTmpFile.Name(), bootstrapConfig.Bytes(), 0o700); err != nil {
		return nil, err
	}

	return bootstrapTmpFile, nil
}

// renderNodeConfig renders the redpanda.yaml node config and retuns it as
// byte array.
func renderNodeConfig(settings options, hostIP string, advertisedKafkaPort int) ([]byte, error) {
	tplParams := redpandaConfigTplParams{
		KafkaAPI: redpandaConfigTplParamsKafkaAPI{
			AdvertisedHost:       hostIP,
			AdvertisedPort:       advertisedKafkaPort,
			AuthenticationMethod: settings.KafkaAuthenticationMethod,
			EnableAuthorization:  settings.KafkaEnableAuthorization,
		},
		SchemaRegistry: redpandaConfigTplParamsSchemaRegistry{
			AuthenticationMethod: settings.SchemaRegistryAuthenticationMethod,
		},
	}

	ncTpl, err := template.New("redpanda.yaml").Parse(nodeConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redpanda config file template: %w", err)
	}

	var redpandaYaml bytes.Buffer
	if err := ncTpl.Execute(&redpandaYaml, tplParams); err != nil {
		return nil, fmt.Errorf("failed to render redpanda node config template: %w", err)
	}

	return redpandaYaml.Bytes(), nil
}

type redpandaBootstrapConfigTplParams struct {
	Superusers                  []string
	KafkaAPIEnableAuthorization bool
}

type redpandaConfigTplParams struct {
	KafkaAPI       redpandaConfigTplParamsKafkaAPI
	SchemaRegistry redpandaConfigTplParamsSchemaRegistry
}

type redpandaConfigTplParamsKafkaAPI struct {
	AdvertisedHost       string
	AdvertisedPort       int
	AuthenticationMethod string
	EnableAuthorization  bool
}

type redpandaConfigTplParamsSchemaRegistry struct {
	AuthenticationMethod string
}
