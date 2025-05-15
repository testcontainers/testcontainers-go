package redpanda

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"golang.org/x/mod/semver"

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
	defaultDockerKafkaAPIPort = "29092"

	redpandaDir         = "/etc/redpanda"
	entrypointFile      = "/entrypoint-tc.sh"
	bootstrapConfigFile = ".bootstrap.yaml"
	certFile            = "cert.pem"
	keyFile             = "key.pem"

	bootstrapAdminAPIUser     = "redpanda_bootstrap_admin_user"
	bootstrapAdminAPIPassword = "redpanda_bootstrap_admin_password"
)

// Container represents the Redpanda container type used in the module.
type Container struct {
	testcontainers.Container
	urlScheme string
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Redpanda container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	return Run(ctx, "docker.redpanda.com/redpandadata/redpanda:v23.3.3", opts...)
}

// Run creates an instance of the Redpanda container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// 1. Create container request.
	// Some (e.g. Image) may be overridden by providing an option argument to this function.
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: img,
			ConfigModifier: func(c *container.Config) {
				c.User = "root:root"
			},
			// Files: Will be added later after we've rendered our YAML templates.
			ExposedPorts: []string{
				defaultKafkaAPIPort,
				defaultAdminAPIPort,
				defaultSchemaRegistryPort,
			},
			Entrypoint: []string{entrypointFile},
			Cmd: []string{
				"redpanda",
				"start",
				"--mode=dev-container",
				"--smp=1",
				"--memory=1G",
			},
			WaitingFor: wait.ForAll(
				// Wait for the ports to be mapped without accessing them,
				// because container needs Redpanda configuration before Redpanda is started
				// and the mapped ports are part of that configuration.
				wait.ForMappedPort(defaultKafkaAPIPort),
				wait.ForMappedPort(defaultAdminAPIPort),
				wait.ForMappedPort(defaultSchemaRegistryPort),
			),
		},
		Started: true,
	}

	// 2. Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	// 2.1. If the image is not at least v23.3, disable wasm transform
	if !isAtLeastVersion(req.Image, "23.3") {
		settings.EnableWasmTransform = false
	}

	// 2.2. If enabled, bootstrap user account
	if settings.enableAdminAPIAuthentication {
		// set the RP_BOOTSTRAP_USER env var
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["RP_BOOTSTRAP_USER"] = bootstrapAdminAPIUser + ":" + bootstrapAdminAPIPassword

		// add our internal bootstrap admin user to superusers
		settings.Superusers = append(settings.Superusers, bootstrapAdminAPIUser)

		// enable admin_api_require_auth
		if settings.ExtraBootstrapConfig == nil {
			settings.ExtraBootstrapConfig = map[string]any{}
		}

		settings.ExtraBootstrapConfig["admin_api_require_auth"] = true
	}

	// 3. Register extra kafka listeners if provided, network aliases will be
	// set
	if err := registerListeners(settings, req); err != nil {
		return nil, fmt.Errorf("register listeners: %w", err)
	}

	// Bootstrap config file contains cluster configurations which will only be considered
	// the very first time you start a cluster.
	bootstrapConfig, err := renderBootstrapConfig(settings)
	if err != nil {
		return nil, err
	}

	// We need a custom entrypoint that waits until the actual Redpanda node config is mounted.
	// Once the redpanda config is mounted we will call the original entrypoint with the same parameters.
	// We have to do this kind of two-step process, because we need to know the mapped
	// port, so that we can use this in Redpanda's advertised listeners configuration for
	// the Kafka API.
	req.Files = append(req.Files,
		testcontainers.ContainerFile{
			Reader:            bytes.NewReader(entrypoint),
			ContainerFilePath: entrypointFile,
			FileMode:          700,
		},
		testcontainers.ContainerFile{
			Reader:            bytes.NewReader(bootstrapConfig),
			ContainerFilePath: filepath.Join(redpandaDir, bootstrapConfigFile),
			FileMode:          600,
		},
	)

	// 4. Create certificate and key for TLS connections.
	if settings.EnableTLS {
		req.Files = append(req.Files,
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(settings.cert),
				ContainerFilePath: filepath.Join(redpandaDir, certFile),
				FileMode:          600,
			},
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(settings.key),
				ContainerFilePath: filepath.Join(redpandaDir, keyFile),
				FileMode:          600,
			},
		)
	}

	ctr, err := testcontainers.GenericContainer(ctx, req)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	// 5. Get mapped port for the Kafka API, so that we can render and then mount
	// the Redpanda config with the advertised Kafka address.
	hostIP, err := ctr.Host(ctx)
	if err != nil {
		return c, fmt.Errorf("host: %w", err)
	}

	kafkaPort, err := ctr.MappedPort(ctx, nat.Port(defaultKafkaAPIPort))
	if err != nil {
		return c, fmt.Errorf("mapped kafka port: %w", err)
	}

	// 6. Render redpanda.yaml config and mount it.
	nodeConfig, err := renderNodeConfig(settings, hostIP, kafkaPort.Int())
	if err != nil {
		return c, err
	}

	err = ctr.CopyToContainer(ctx, nodeConfig, filepath.Join(redpandaDir, "redpanda.yaml"), 0o600)
	if err != nil {
		return c, fmt.Errorf("copy to container: %w", err)
	}

	// 7. Wait until Redpanda is ready to serve requests.
	waitHTTP := wait.ForHTTP(defaultAdminAPIPort).
		WithStatusCodeMatcher(func(status int) bool {
			// Redpanda's admin API returns 404 for requests to "/".
			return status == http.StatusNotFound
		})

	var tlsConfig *tls.Config
	if settings.EnableTLS {
		cert, err := tls.X509KeyPair(settings.cert, settings.key)
		if err != nil {
			return c, fmt.Errorf("create admin cert: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(settings.cert)
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		waitHTTP = waitHTTP.WithTLS(true, tlsConfig)
	}
	err = wait.ForAll(
		wait.ForListeningPort(defaultKafkaAPIPort),
		waitHTTP,
		wait.ForListeningPort(defaultSchemaRegistryPort),
		wait.ForLog("Successfully started Redpanda!"),
	).WaitUntilReady(ctx, ctr)
	if err != nil {
		return c, fmt.Errorf("wait for readiness: %w", err)
	}

	c.urlScheme = "http"
	if settings.EnableTLS {
		c.urlScheme += "s"
	}

	// 8. Create Redpanda Service Accounts if configured to do so.
	if len(settings.ServiceAccounts) > 0 {
		adminAPIPort, err := ctr.MappedPort(ctx, nat.Port(defaultAdminAPIPort))
		if err != nil {
			return c, fmt.Errorf("mapped admin port: %w", err)
		}

		adminAPIUrl := fmt.Sprintf("%s://%v:%d", c.urlScheme, hostIP, adminAPIPort.Int())
		adminCl := NewAdminAPIClient(adminAPIUrl)
		if settings.enableAdminAPIAuthentication {
			adminCl = adminCl.WithAuthentication(bootstrapAdminAPIUser, bootstrapAdminAPIPassword)
		}

		if settings.EnableTLS {
			adminCl = adminCl.WithHTTPClient(&http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					ForceAttemptHTTP2:   true,
					TLSHandshakeTimeout: 10 * time.Second,
					TLSClientConfig:     tlsConfig,
				},
			})
		}

		for username, password := range settings.ServiceAccounts {
			if err := adminCl.CreateUser(ctx, username, password); err != nil {
				return c, fmt.Errorf("create user %q: %w", username, err)
			}
		}
	}

	return c, nil
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
		EnableWasmTransform:         settings.EnableWasmTransform,
		ExtraBootstrapConfig:        settings.ExtraBootstrapConfig,
	}

	tpl, err := template.New("bootstrap.yaml").Parse(bootstrapConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("parse bootstrap template: %w", err)
	}

	var bootstrapConfig bytes.Buffer
	if err := tpl.Execute(&bootstrapConfig, bootstrapTplParams); err != nil {
		return nil, fmt.Errorf("render bootstrap template: %w", err)
	}

	return bootstrapConfig.Bytes(), nil
}

// registerListeners validates that the provided listeners are valid and set network aliases for the provided addresses.
// The container must be attached to at least one network.
func registerListeners(settings options, req testcontainers.GenericContainerRequest) error {
	if len(settings.Listeners) == 0 {
		return nil
	}

	if len(req.Networks) == 0 {
		return errors.New("container must be attached to at least one network")
	}

	for _, listener := range settings.Listeners {
		if listener.Port < 0 || listener.Port > math.MaxUint16 {
			return fmt.Errorf("invalid port on listener %s:%d (must be between 0 and 65535)", listener.Address, listener.Port)
		}

		for _, network := range req.Networks {
			req.NetworkAliases[network] = append(req.NetworkAliases[network], listener.Address)
		}
	}
	return nil
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
			Listeners:            settings.Listeners,
		},
		SchemaRegistry: redpandaConfigTplParamsSchemaRegistry{
			AuthenticationMethod: settings.SchemaRegistryAuthenticationMethod,
		},
		EnableTLS: settings.EnableTLS,
	}

	ncTpl, err := template.New("redpanda.yaml").Parse(nodeConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("parse node config template: %w", err)
	}

	var redpandaYaml bytes.Buffer
	if err := ncTpl.Execute(&redpandaYaml, tplParams); err != nil {
		return nil, fmt.Errorf("render node config template: %w", err)
	}

	return redpandaYaml.Bytes(), nil
}

type redpandaBootstrapConfigTplParams struct {
	Superusers                  []string
	KafkaAPIEnableAuthorization bool
	AutoCreateTopics            bool
	EnableWasmTransform         bool
	ExtraBootstrapConfig        map[string]any
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
	Listeners            []listener
}

type redpandaConfigTplParamsSchemaRegistry struct {
	AuthenticationMethod string
}

type listener struct {
	Address              string
	Port                 int
	AuthenticationMethod string
}

// isAtLeastVersion returns true if the base image (without tag) is in a version or above
func isAtLeastVersion(image, major string) bool {
	parts := strings.Split(image, ":")
	version := parts[len(parts)-1]

	if version == "latest" {
		return true
	}

	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v"+major) >= 0 // version >= v8.x
	}

	return false
}
