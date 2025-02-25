package client

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/docker/docker/client"

	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

const (
	// Headers used for docker client requests.
	headerProjectPath = "x-tc-pp"
	headerSessionID   = "x-tc-sid"
	headerUserAgent   = "User-Agent"

	// TLS certificate files.
	tlsCACertFile = "ca.pem"
	tlsCertFile   = "cert.pem"
	tlsKeyFile    = "key.pem"
)

// DefaultClient is the default client for interacting with containers.
var DefaultClient = &Client{}

// Client is a type that represents a client for interacting with containers.
type Client struct {
	log slog.Logger

	// mtx is a mutex for synchronizing access to the fields below.
	mtx    sync.RWMutex
	client *client.Client
	cfg    *config
	err    error
}

// ClientOption is a type that represents an option for configuring a client.
type ClientOption func(*Client) error

// Logger returns a client option that sets the logger for the client.
func Logger(log slog.Logger) ClientOption {
	return func(c *Client) error {
		c.log = log
		return nil
	}
}

// NewClient returns a new client for interacting with containers.
func NewClient(ctx context.Context, options ...ClientOption) (*Client, error) {
	client := &Client{}
	for _, opt := range options {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	if err := client.initOnce(ctx); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return client, nil
}

// initOnce initializes the client once.
// This method is safe for concurrent use by multiple goroutines.
func (c *Client) initOnce(ctx context.Context) error {
	c.mtx.RLock()
	if c.client != nil || c.err != nil {
		err := c.err
		c.mtx.RUnlock()
		return err
	}
	c.mtx.RUnlock()

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.cfg, c.err = newConfig(); c.err != nil {
		return c.err
	}

	opts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}

	// TODO: handle internally / replace with context related code.
	if dockerHost := core.MustExtractDockerHost(ctx); dockerHost != "" {
		opts = append(opts, client.WithHost(dockerHost))
	}

	if c.cfg.TLSVerify {
		// For further information see:
		// https://docs.docker.com/engine/security/protect-access/#use-tls-https-to-protect-the-docker-daemon-socket
		opts = append(opts, client.WithTLSClientConfig(
			filepath.Join(c.cfg.CertPath, tlsCACertFile),
			filepath.Join(c.cfg.CertPath, tlsCertFile),
			filepath.Join(c.cfg.CertPath, tlsKeyFile),
		))
	}

	opts = append(opts, client.WithHTTPHeaders(
		map[string]string{
			headerProjectPath: core.ProjectPath(),
			headerSessionID:   core.SessionID(),
			headerUserAgent:   "tc-go/" + internal.Version,
		}),
	)

	if c.client, c.err = client.NewClientWithOpts(opts...); c.err != nil {
		c.err = fmt.Errorf("new client: %w", c.err)
		return c.err
	}

	return nil
}
