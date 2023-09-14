package rabbitmq

import (
	"context"
	"fmt"
	"strings"
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

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

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

// WithStartupCommand will execute the command representation of each Executable into the container.
// It will leverage the container lifecycle hooks to call the command right after the container
// is started.
func WithStartupCommand(execs ...Executable) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		for _, exec := range execs {
			execFn := func(ctx context.Context, c testcontainers.Container) error {
				_, _, err := c.Exec(ctx, exec.AsCommand())
				return err
			}

			req.LifecycleHooks[0].PostStarts = append(req.LifecycleHooks[0].PostStarts, execFn)
		}
	}
}

// WithSSL enables SSL on the RabbitMQ container, adding the necessary environment variables,
// files and waiting conditions.
// From https://hub.docker.com/_/rabbitmq: "As of RabbitMQ 3.9, all of the docker-specific variables
// listed below are deprecated and no longer used. Please use a configuration file instead;
// visit https://rabbitmq.com/configure to learn more about the configuration file. For a starting point,
// the 3.8 images will print out the config file it generated from supplied environment variables.
func WithSSL(sslSettings SSLSettings) testcontainers.CustomizeRequestOption {
	const rabbitCaCertPath = "/etc/rabbitmq/ca_cert.pem"
	const rabbitCertPath = "/etc/rabbitmq/rabbitmq_cert.pem"
	const rabbitKeyPath = "/etc/rabbitmq/rabbitmq_key.pem"

	const defaultPermission = 0o644

	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["RABBITMQ_SSL_VERIFY"] = string(sslSettings.VerificationMode)

		req.Env["RABBITMQ_SSL_CACERTFILE"] = rabbitCaCertPath
		req.Env["RABBITMQ_SSL_CERTFILE"] = rabbitCertPath
		req.Env["RABBITMQ_SSL_KEYFILE"] = rabbitKeyPath

		if sslSettings.VerificationDepth > 0 {
			req.Env["RABBITMQ_SSL_DEPTH"] = fmt.Sprintf("%d", sslSettings.VerificationDepth)
		}

		req.Env["RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT"] = fmt.Sprintf("%t", sslSettings.FailIfNoCert)

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
