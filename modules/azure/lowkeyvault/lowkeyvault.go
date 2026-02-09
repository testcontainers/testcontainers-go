package lowkeyvault

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/go-connections/nat"
	"golang.org/x/crypto/pkcs12"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// possible access modes
const (
	Local = iota
	Network
)

const (
	// defaultAPIPort is the default port used by for the Lowkey Vault Key Vault API endpoints
	defaultAPIPort nat.Port = "8443/tcp"
	// defaultMetadataPort is the default port used for the Lowkey Vault Metadata endpoints
	defaultMetadataPort nat.Port = "8080/tcp"
)

// Container represents the Lowkey Vault container type used in the module
type Container struct {
	testcontainers.Container
	localHostName  string
	remoteHostName string
}

// WithNetworkAlias sets the alias of the container for the provided network and adds the specified name as a key vault alias for the default, "localhost", vault.
func WithNetworkAlias(alias string, forNetwork *testcontainers.DockerNetwork) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		err := network.WithNetwork([]string{alias}, forNetwork).Customize(req)
		if err != nil {
			return err
		}
		// The <port> placeholder will be replaced by the container automatically just in time
		envs := map[string]string{"LOWKEY_VAULT_ALIASES": "localhost=" + alias + ":<port>"}
		err = testcontainers.WithEnv(envs).Customize(req)
		if err != nil {
			return err
		}
		return nil
	}
}

// Run creates an instance of the Lowkey Vault container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Initialize with module defaults
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultAPIPort.Port(), defaultMetadataPort.Port()),
		testcontainers.WithEnv(map[string]string{
			"LOWKEY_VAULT_RELAXED_PORTS": "true",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Started LowkeyVaultApp in "),
		),
	}

	// Add user-provided options
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, localHostName: "", remoteHostName: ""}
	}

	if err != nil {
		return c, fmt.Errorf("run lowkeyvault: %w", err)
	}

	return c, nil
}

// Client prepares a client which will accept insecure (self-signed) certificates.
func (c *Container) Client(ctx context.Context) (http.Client, error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	authority, err := c.mappedHostAuthority(ctx, defaultMetadataPort, Local)
	if err != nil {
		return http.Client{}, fmt.Errorf("host authority: %w", err)
	}
	pass, err := c.fetchDefaultCertPassword(authority)
	if err != nil {
		return http.Client{}, fmt.Errorf("default cert password: %w", err)
	}
	p12Data, err := c.fetchDefaultCertContent(authority)
	if err != nil {
		return http.Client{}, fmt.Errorf("default cert content: %w", err)
	}

	_, cert, err := pkcs12.Decode(p12Data, pass)
	if err != nil {
		return http.Client{}, fmt.Errorf("decode cert content: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	customTransport.TLSClientConfig = &tls.Config{RootCAs: certPool}
	return http.Client{Transport: customTransport}, nil
}

func (c *Container) fetchDefaultCertPassword(authority string) (string, error) {
	bytes, err := c.fetchContent("http://" + authority + "/metadata/default-cert/password")
	if err != nil {
		return "", fmt.Errorf("default cert password: %w", err)
	}
	return string(bytes), nil
}

func (c *Container) fetchDefaultCertContent(authority string) ([]byte, error) {
	return c.fetchContent("http://" + authority + "/metadata/default-cert/lowkey-vault.p12")
}

func (c *Container) fetchContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("get content: %w", err)
	}

	if resp != nil && resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		return bodyBytes, nil
	}

	return nil, errors.New("content not found")
}

// ConnectionURL returns the connection URL for the Lowkey Vault API based on the provided access mode.
func (c *Container) ConnectionURL(ctx context.Context, accessMode int) (string, error) {
	hostAuthority, err := c.mappedHostAuthority(ctx, defaultAPIPort, accessMode)
	if err != nil {
		return "", fmt.Errorf("host authority: %w", err)
	}

	return "https://" + hostAuthority, nil
}

// IdentityEndpoint returns the URL value of the IDENTITY_ENDPOINT environment variable for the managed identity simulation. This will be used to obtain an access token for the Lowkey Vault API.
func (c *Container) IdentityEndpoint(ctx context.Context, accessMode int) (string, error) {
	hostAuthority, err := c.mappedHostAuthority(ctx, defaultMetadataPort, accessMode)
	if err != nil {
		return "", fmt.Errorf("host authority: %w", err)
	}
	return fmt.Sprintf("http://%s/metadata/identity/oauth2/token", hostAuthority), nil
}

// IdentityHeader returns the value of the IDENTITY_HEADER environment variable for the managed identity simulation.
func (c *Container) IdentityHeader() string {
	return "header"
}

func (c *Container) mappedHostAuthority(ctx context.Context, exposedPort nat.Port, accessMode int) (string, error) {
	switch accessMode {
	case Local:
		host, err := c.resolveLocalHostName(ctx)
		if err != nil {
			return "", fmt.Errorf("host: %w", err)
		}
		port, err := c.MappedPort(ctx, exposedPort)
		if err != nil {
			return "", fmt.Errorf("port: %w", err)
		}
		return fmt.Sprintf("%s:%d", host, port.Int()), nil
	case Network:
		host, err := c.resolveNetworkHostName(ctx)
		if err != nil {
			return "", fmt.Errorf("host: %w", err)
		}
		return fmt.Sprintf("%s:%d", host, exposedPort.Int()), nil
	default:
		return "", fmt.Errorf("unsupported access mode: %d", accessMode)
	}
}

func (c *Container) resolveLocalHostName(ctx context.Context) (string, error) {
	if c.localHostName != "" {
		return c.localHostName, nil
	}
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}
	c.localHostName = host
	return host, nil
}

func (c *Container) resolveNetworkHostName(ctx context.Context) (string, error) {
	if c.remoteHostName != "" {
		return c.remoteHostName, nil
	}
	networks, err := c.Networks(ctx)
	if err != nil {
		return "", fmt.Errorf("networks: %w", err)
	}
	if len(networks) != 1 {
		return "", fmt.Errorf("the container must have exactly one network, but it has %d", len(networks))
	}
	hosts, err := c.NetworkAliases(ctx)
	if err != nil {
		return "", fmt.Errorf("network aliases: %w", err)
	}
	if len(hosts) == 0 {
		return "", errors.New("no network aliases found in the Lowkey Vault container")
	}
	aliases := hosts[networks[0]]
	if len(aliases) != 1 {
		return "", fmt.Errorf("the container must have exactly one network alias, but it has %d", len(aliases))
	}
	host := aliases[0]
	c.remoteHostName = host
	return host, nil
}
