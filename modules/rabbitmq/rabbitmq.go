package rabbitmq

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DefaultAMQPSPort              = "5671/tcp"
	DefaultAMQPPort               = "5672/tcp"
	DefaultHTTPSPort              = "15671/tcp"
	DefaultHTTPPort               = "15672/tcp"
	defaultPassword               = "guest"
	defaultUser                   = "guest"
	defaultCustomConfPath         = "/etc/rabbitmq/rabbitmq-custom.conf"
	defaultCustomConfigErlangPath = "/etc/rabbitmq/rabbitmq-custom.config"
)

var (
	//go:embed mounts/rabbitmq-custom.config.tpl
	customConfigTpl string
)

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
	endpoint, err := c.PortEndpoint(ctx, nat.Port(DefaultAMQPPort), "")
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

// RunContainer creates an instance of the RabbitMQ container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RabbitMQContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "rabbitmq:3.7.25-management-alpine",
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
		opt.Customize(&genericContainerReq)
	}

	if settings.SSLSettings != nil {
		applySSLSettings(settings.SSLSettings)(&genericContainerReq)
	}

	nodeConfig, err := renderRabbitMQConfig(settings)
	if err != nil {
		return nil, err
	}

	tmpConfigFile := filepath.Join(os.TempDir(), "rabbitmq-custom.config")
	err = os.WriteFile(tmpConfigFile, nodeConfig, 0o600)
	if err != nil {
		return nil, err
	}

	WithConfigErlang(tmpConfigFile)(&genericContainerReq)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	user := genericContainerReq.Env["RABBITMQ_DEFAULT_USER"]
	password := genericContainerReq.Env["RABBITMQ_DEFAULT_PASS"]

	c := &RabbitMQContainer{Container: container, AdminUsername: user, AdminPassword: password}

	return c, nil
}

// WithAdminPassword sets the password for the default admin user
func WithAdminPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["RABBITMQ_DEFAULT_PASS"] = password
	}
}

// WithAdminUsername sets the default admin username
func WithAdminUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["RABBITMQ_DEFAULT_USER"] = username
	}
}

// WithConfig overwrites the default RabbitMQ configuration file with the supplied one.
// This function (which uses the Sysctl format) is recommended for RabbitMQ >= 3.7
// It's important to notice that this function does not work with RabbitMQ < 3.7.
// The "configFilePath" parameter holds the path to the file to use (in sysctl format, don't forget empty line in the end of file)
// and it will check if the file has the ".conf" extension.
func WithConfig(confFilePath string) testcontainers.CustomizeRequestOption {
	return withConfig(confFilePath, defaultCustomConfPath, func(s string) bool {
		return strings.HasSuffix(s, ".conf")
	})
}

// WithConfigErlang overwrites the default RabbitMQ configuration file with the supplied one.
// The "configFilePath" parameter holds the path to the file to use (in sysctl format, don't forget empty line in the end of file)
// and it will check if the file has the ".config" extension.
func WithConfigErlang(confFilePath string) testcontainers.CustomizeRequestOption {
	return withConfig(confFilePath, defaultCustomConfigErlangPath, func(s string) bool {
		return strings.HasSuffix(s, ".config")
	})
}

func withConfig(hostPath string, containerPath string, validateFn func(string) bool) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if !validateFn(hostPath) {
			return
		}

		req.Env["RABBITMQ_CONFIG_FILE"] = containerPath

		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      hostPath,
			ContainerFilePath: containerPath,
			FileMode:          0o644,
		})
	}
}

// applySSLSettings transfers the SSL settings to the container request.
func applySSLSettings(sslSettings *SSLSettings) testcontainers.CustomizeRequestOption {
	const rabbitCaCertPath = "/etc/rabbitmq/ca_cert.pem"
	const rabbitCertPath = "/etc/rabbitmq/rabbitmq_cert.pem"
	const rabbitKeyPath = "/etc/rabbitmq/rabbitmq_key.pem"

	const defaultPermission = 0o644

	return func(req *testcontainers.GenericContainerRequest) {
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
	}
}

func renderRabbitMQConfig(opts options) ([]byte, error) {
	rabbitCustomConfigTpl, err := template.New("rabbitmq-custom.config").Parse(customConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RabbitMQ config file template: %w", err)
	}

	var rabbitMQConfig bytes.Buffer
	if err := rabbitCustomConfigTpl.Execute(&rabbitMQConfig, opts); err != nil {
		return nil, fmt.Errorf("failed to render RabbitMQ config template: %w", err)
	}

	return rabbitMQConfig.Bytes(), nil
}
