package redpanda

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
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
)

const (
	defaultKafkaAPIPort       = "9092/tcp"
	defaultAdminAPIPort       = "9644/tcp"
	defaultSchemaRegistryPort = "8081/tcp"

	redpandaDir         = "/etc/redpanda"
	entrypointFile      = "/entrypoint-tc.sh"
	bootstrapConfigFile = ".bootstrap.yaml"
	certFile            = "cert.pem"
	keyFile             = "key.pem"
)

// Container represents the Redpanda container type used in the module.
type Container struct {
	testcontainers.Container
	urlScheme string
}

// RunContainer creates an instance of the Redpanda container type.
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	tmpDir, err := os.MkdirTemp("", "redpanda")
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

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
				entrypointFile,
				"redpanda",
				"start",
				"--mode=dev-container",
				"--smp=1",
				"--memory=1G",
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
	entrypointPath := filepath.Join(tmpDir, entrypointFile)
	if err := os.WriteFile(entrypointPath, entrypoint, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create entrypoint file: %w", err)
	}

	// Bootstrap config file contains cluster configurations which will only be considered
	// the very first time you start a cluster.
	bootstrapConfigPath := filepath.Join(tmpDir, bootstrapConfigFile)
	bootstrapConfig, err := renderBootstrapConfig(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap config file: %w", err)
	}
	if err := os.WriteFile(bootstrapConfigPath, bootstrapConfig, 0o600); err != nil {
		return nil, fmt.Errorf("failed to create bootstrap config file: %w", err)
	}

	req.Files = append(req.Files,
		testcontainers.ContainerFile{
			HostFilePath:      entrypointPath,
			ContainerFilePath: entrypointFile,
			FileMode:          700,
		},
		testcontainers.ContainerFile{
			HostFilePath:      bootstrapConfigPath,
			ContainerFilePath: filepath.Join(redpandaDir, bootstrapConfigFile),
			FileMode:          600,
		},
	)

	// 4. Create certificate and key for TLS connections.
	if settings.EnableTLS {
		certPath := filepath.Join(tmpDir, certFile)
		if err := os.WriteFile(certPath, settings.cert, 0o600); err != nil {
			return nil, fmt.Errorf("failed to create certificate file: %w", err)
		}
		keyPath := filepath.Join(tmpDir, keyFile)
		if err := os.WriteFile(keyPath, settings.key, 0o600); err != nil {
			return nil, fmt.Errorf("failed to create key file: %w", err)
		}

		req.Files = append(req.Files,
			testcontainers.ContainerFile{
				HostFilePath:      certPath,
				ContainerFilePath: filepath.Join(redpandaDir, certFile),
				FileMode:          600,
			},
			testcontainers.ContainerFile{
				HostFilePath:      keyPath,
				ContainerFilePath: filepath.Join(redpandaDir, keyFile),
				FileMode:          600,
			},
		)
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	// 5. Get mapped port for the Kafka API, so that we can render and then mount
	// the Redpanda config with the advertised Kafka address.
	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	kafkaPort, err := container.MappedPort(ctx, nat.Port(defaultKafkaAPIPort))
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped Kafka port: %w", err)
	}

	// 6. Render redpanda.yaml config and mount it.
	nodeConfig, err := renderNodeConfig(settings, hostIP, kafkaPort.Int())
	if err != nil {
		return nil, fmt.Errorf("failed to render node config: %w", err)
	}

	err = container.CopyToContainer(ctx, nodeConfig, filepath.Join(redpandaDir, "redpanda.yaml"), 600)
	if err != nil {
		return nil, fmt.Errorf("failed to copy redpanda.yaml into container: %w", err)
	}

	// 6. Wait until Redpanda is ready to serve requests
	err = wait.ForAll(
		wait.ForListeningPort(defaultKafkaAPIPort),
		wait.ForLog("Successfully started Redpanda!").WithPollInterval(100*time.Millisecond)).
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

	scheme := "http"
	if settings.EnableTLS {
		scheme += "s"
	}

	return &Container{Container: container, urlScheme: scheme}, nil
}

// KafkaSeedBroker returns the seed broker that should be used for connecting
// to the Kafka API with your Kafka client. It'll be returned in the format:
// "host:port" - for example: "localhost:55687".
func (c *Container) KafkaSeedBroker(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, nat.Port(defaultKafkaAPIPort), "")
}

// AdminAPIAddress returns the address to the Redpanda Admin API. This
// is an HTTP-based API and thus the returned format will be: http://host:port.
func (c *Container) AdminAPIAddress(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, nat.Port(defaultAdminAPIPort), c.urlScheme)
}

// SchemaRegistryAddress returns the address to the schema registry API. This
// is an HTTP-based API and thus the returned format will be: http://host:port.
func (c *Container) SchemaRegistryAddress(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, nat.Port(defaultSchemaRegistryPort), c.urlScheme)
}

// renderBootstrapConfig renders the config template for the .bootstrap.yaml config,
// which configures Redpanda's cluster properties.
// Reference: https://docs.redpanda.com/docs/reference/cluster-properties/
func renderBootstrapConfig(settings options) ([]byte, error) {
	bootstrapTplParams := redpandaBootstrapConfigTplParams{
		Superusers:                  settings.Superusers,
		KafkaAPIEnableAuthorization: settings.KafkaEnableAuthorization,
		AutoCreateTopics:            settings.AutoCreateTopics,
	}

	tpl, err := template.New("bootstrap.yaml").Parse(bootstrapConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redpanda config file template: %w", err)
	}

	var bootstrapConfig bytes.Buffer
	if err := tpl.Execute(&bootstrapConfig, bootstrapTplParams); err != nil {
		return nil, fmt.Errorf("failed to render redpanda bootstrap config template: %w", err)
	}

	return bootstrapConfig.Bytes(), nil
}

// renderNodeConfig renders the redpanda.yaml node config and returns it as
// byte array.
func renderNodeConfig(settings options, hostIP string, advertisedKafkaPort int) ([]byte, error) {
	tplParams := redpandaConfigTplParams{
		AutoCreateTopics: settings.AutoCreateTopics,
		KafkaAPI: redpandaConfigTplParamsKafkaAPI{
			AdvertisedHost:       hostIP,
			AdvertisedPort:       advertisedKafkaPort,
			AuthenticationMethod: settings.KafkaAuthenticationMethod,
			EnableAuthorization:  settings.KafkaEnableAuthorization,
		},
		SchemaRegistry: redpandaConfigTplParamsSchemaRegistry{
			AuthenticationMethod: settings.SchemaRegistryAuthenticationMethod,
		},
		EnableTLS: settings.EnableTLS,
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
	AutoCreateTopics            bool
}

type redpandaConfigTplParams struct {
	KafkaAPI         redpandaConfigTplParamsKafkaAPI
	SchemaRegistry   redpandaConfigTplParamsSchemaRegistry
	AutoCreateTopics bool
	EnableTLS        bool
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
