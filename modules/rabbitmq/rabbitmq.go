package rabbitmq

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

const (
	DefaultAMQPSPort      = "5671/tcp"
	DefaultAMQPPort       = "5672/tcp"
	DefaultHTTPSPort      = "15671/tcp"
	DefaultHTTPPort       = "15672/tcp"
	defaultPassword       = "guest"
	defaultUser           = "guest"
	defaultCustomConfPath = "/etc/rabbitmq/rabbitmq-testcontainers.conf"
)

//go:embed mounts/rabbitmq-testcontainers.conf.tpl
var customConfigTpl string

// RabbitMQContainer represents the RabbitMQ container type used in the module
type RabbitMQContainer struct {
	testcontainers.Container
	AdminPassword string
	AdminUsername string
}

// AmqpURL returns the URL for AMQP clients.
func (c *RabbitMQContainer) AmqpURL(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, nat.Port(DefaultAMQPPort), "")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("amqp://%s:%s@%s", c.AdminUsername, c.AdminPassword, endpoint), nil
}

// AmqpURL returns the URL for AMQPS clients.
func (c *RabbitMQContainer) AmqpsURL(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, nat.Port(DefaultAMQPSPort), "")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("amqps://%s:%s@%s", c.AdminUsername, c.AdminPassword, endpoint), nil
}

// HttpURL returns the URL for HTTP management.
func (c *RabbitMQContainer) HttpURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, nat.Port(DefaultHTTPPort), "http")
}

// HttpsURL returns the URL for HTTPS management.
func (c *RabbitMQContainer) HttpsURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, nat.Port(DefaultHTTPSPort), "https")
}

// Deprecated: use Run instead
// RunContainer creates an instance of the RabbitMQ container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RabbitMQContainer, error) {
	return Run(ctx, "rabbitmq:3.12.11-management-alpine", opts...)
}

// Run creates an instance of the RabbitMQ container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RabbitMQContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": defaultUser,
			"RABBITMQ_DEFAULT_PASS": defaultPassword,
		},
		ExposedPorts: []string{
			DefaultAMQPPort,
			DefaultAMQPSPort,
			DefaultHTTPSPort,
			DefaultHTTPPort,
		},
		WaitingFor: wait.ForLog(".*Server startup complete.*").AsRegexp().WithStartupTimeout(60 * time.Second),
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			{
				PostStarts: []testcontainers.ContainerHook{},
			},
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Logger:           testcontainers.Logger,
		Started:          true,
	}

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	if settings.SSLSettings != nil {
		if err := applySSLSettings(settings.SSLSettings)(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	nodeConfig, err := renderRabbitMQConfig(settings)
	if err != nil {
		return nil, err
	}

	tmpConfigFile := filepath.Join(os.TempDir(), "rabbitmq-testcontainers.conf")
	err = os.WriteFile(tmpConfigFile, nodeConfig, 0o600)
	if err != nil {
		return nil, err
	}

	if err := withConfig(tmpConfigFile)(&genericContainerReq); err != nil {
		return nil, err
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *RabbitMQContainer
	if container != nil {
		c = &RabbitMQContainer{
			Container:     container,
			AdminUsername: settings.AdminUsername,
			AdminPassword: settings.AdminPassword,
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func withConfig(hostPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["RABBITMQ_CONFIG_FILE"] = defaultCustomConfPath

		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      hostPath,
			ContainerFilePath: defaultCustomConfPath,
			FileMode:          0o644,
		})

		return nil
	}
}

// applySSLSettings transfers the SSL settings to the container request.
func applySSLSettings(sslSettings *SSLSettings) testcontainers.CustomizeRequestOption {
	const rabbitCaCertPath = "/etc/rabbitmq/ca_cert.pem"
	const rabbitCertPath = "/etc/rabbitmq/rabbitmq_cert.pem"
	const rabbitKeyPath = "/etc/rabbitmq/rabbitmq_key.pem"

	const defaultPermission = 0o644

	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      sslSettings.CACertFile,
			ContainerFilePath: rabbitCaCertPath,
			FileMode:          defaultPermission,
		})
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      sslSettings.CertFile,
			ContainerFilePath: rabbitCertPath,
			FileMode:          defaultPermission,
		})
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      sslSettings.KeyFile,
			ContainerFilePath: rabbitKeyPath,
			FileMode:          defaultPermission,
		})

		// To verify that TLS has been enabled on the node, container logs should contain an entry about a TLS listener being enabled
		// See https://www.rabbitmq.com/ssl.html#enabling-tls-verify-configuration
		req.WaitingFor = wait.ForAll(req.WaitingFor, wait.ForLog("started TLS (SSL) listener on [::]:5671"))

		return nil
	}
}

func renderRabbitMQConfig(opts options) ([]byte, error) {
	rabbitCustomConfigTpl, err := template.New("rabbitmq-testcontainers.conf").Parse(customConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RabbitMQ config file template: %w", err)
	}

	var rabbitMQConfig bytes.Buffer
	if err := rabbitCustomConfigTpl.Execute(&rabbitMQConfig, opts); err != nil {
		return nil, fmt.Errorf("failed to render RabbitMQ config template: %w", err)
	}

	return rabbitMQConfig.Bytes(), nil
}
