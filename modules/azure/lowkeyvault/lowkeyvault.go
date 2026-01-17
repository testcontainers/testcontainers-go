package lowkeyvalt

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/docker/go-connections/nat"
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
	// defaultApiPort is the default port used by for the Lowkey Vault Key Vault API endpoints
	defaultApiPort nat.Port = "8443/tcp"
	// defaultMetadataPort is the default port used for the Lowkey Vault Metadata endpoints
	defaultMetadataPort nat.Port = "8080/tcp"
)

// LowkeyVaultContainer represents the Lowkey Vault container type used in the module
type LowkeyVaultContainer struct {
	testcontainers.Container
}

// WithNetworkAlias sets the alias of the container for the provided network and adds the specified name as a key vault alias for the default, "localhost", vault.
func WithNetworkAlias(alias string, forNetwork *testcontainers.DockerNetwork) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		err := network.WithNetwork([]string{alias}, forNetwork).Customize(req)
		if err != nil {
			return err
		}
		envs := map[string]string{"LOWKEY_VAULT_ALIASES": "localhost=" + alias + ":<port>"}
		err = testcontainers.WithEnv(envs).Customize(req)
		if err != nil {
			return err
		}
		return nil
	}
}

// Run creates an instance of the Lowkey Vault container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*LowkeyVaultContainer, error) {
	// Initialize with module defaults
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultApiPort.Port(), defaultMetadataPort.Port()),
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
	var c *LowkeyVaultContainer
	if ctr != nil {
		c = &LowkeyVaultContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run lowkeyvault: %w", err)
	}

	return c, nil
}

// PrepareClientForSelfSignedCert prepares a client which will accept insecure (self-signed) certificates.
func (c *LowkeyVaultContainer) PrepareClientForSelfSignedCert() http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return http.Client{Transport: customTransport}
}

// ConnectionUrl returns the connection URL for the Lowkey Vault API based on the provided access mode.
func (c *LowkeyVaultContainer) ConnectionUrl(ctx context.Context, accessMode int) (string, error) {
	hostAuthority, err := c._MappedHostAuthority(ctx, defaultApiPort, accessMode)
	if err != nil {
		return "", fmt.Errorf("host authority: %w", err)
	}

	return fmt.Sprintf("https://%s", hostAuthority), nil
}

// TokenUrl returns the connection URL for the Lowkey Vault token endpoint based on the provided access mode.
func (c *LowkeyVaultContainer) TokenUrl(ctx context.Context, accessMode int) (string, error) {
	hostAuthority, err := c._MappedHostAuthority(ctx, defaultMetadataPort, accessMode)
	if err != nil {
		return "", fmt.Errorf("host authority: %w", err)
	}
	return fmt.Sprintf("http://%s/metadata/identity/oauth2/token", hostAuthority), nil
}

// SetManagedIdentityEnvVariables sets the environment variables required for managed identity authentication.
// This works only with local access mode
func (c *LowkeyVaultContainer) SetManagedIdentityEnvVariables(ctx context.Context) error {
	tokenUrl, err := c.TokenUrl(ctx, Local)
	if err != nil {
		return fmt.Errorf("token url: %w", err)
	}
	err = os.Setenv("IDENTITY_ENDPOINT", tokenUrl)
	if err != nil {
		return fmt.Errorf("set env IDENTITY_ENDPOINT: %w", err)
	}
	err = os.Setenv("IDENTITY_HEADER", "header")
	if err != nil {
		return fmt.Errorf("set env IDENTITY_HEADER: %w", err)
	}
	return nil
}

func (c *LowkeyVaultContainer) _MappedHostAuthority(ctx context.Context, exposedPort nat.Port, accessMode int) (string, error) {
	if accessMode == Local {
		host, err := c.Host(ctx)
		if err != nil {
			return "", fmt.Errorf("host: %w", err)
		}
		port, err := c.MappedPort(ctx, exposedPort)
		if err != nil {
			return "", fmt.Errorf("api port: %w", err)
		}
		return fmt.Sprintf("%s:%d", host, port.Int()), nil
	}
	networks, err := c.Networks(ctx)
	if err != nil {
		return "", fmt.Errorf("networks: %w", err)
	}
	hosts, err := c.NetworkAliases(ctx)
	if err != nil {
		return "", fmt.Errorf("network aliases: %w", err)
	}
	host := hosts[networks[0]][0]
	return fmt.Sprintf("%s:%d", host, exposedPort.Int()), nil
}
