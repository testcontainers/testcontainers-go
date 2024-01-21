package cockroachdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	certsDir = "/tmp"

	defaultSQLPort   = "26257"
	defaultAdminPort = "8080"

	defaultImage     = "cockroachdb/cockroach"
	defaultImageTag  = "latest-v23.1"
	defaultDatabase  = "defaultdb"
	defaultStoreSize = "100%"
)

// CockroachDBContainer represents the CockroachDB container type used in the module
type CockroachDBContainer struct {
	testcontainers.Container
	opts options
}

// MustConnectionString panics if the address cannot be determined.
func (c *CockroachDBContainer) MustConnectionString(ctx context.Context) string {
	addr, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	return addr
}

// ConnectionString returns the dial address to open a new connection to CockroachDB.
func (c *CockroachDBContainer) ConnectionString(ctx context.Context) (string, error) {
	mappedport, err := c.MappedPort(ctx, defaultSQLPort+"/tcp")
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	sslMode := "disable"
	if c.opts.TLS != nil {
		sslMode = "verify-full"
	}
	params := url.Values{
		"sslmode": []string{sslMode},
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.User("root"),
		Host:     net.JoinHostPort(hostIP, mappedport.Port()),
		Path:     c.opts.Database,
		RawQuery: params.Encode(),
	}

	return u.String(), nil
}

// TLSConfig returns config neccessary to connect to CockroachDB over TLS.
func (c *CockroachDBContainer) TLSConfig() (*tls.Config, error) {
	if c.opts.TLS == nil {
		return nil, fmt.Errorf("tls not enabled")
	}

	keyPair, err := tls.X509KeyPair(c.opts.TLS.ClientCert, c.opts.TLS.ClientKey)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(c.opts.TLS.CACert)

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{keyPair},
		ServerName:   "localhost",
	}, nil
}

// RunContainer creates an instance of the CockroachDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			ExposedPorts: []string{
				defaultSQLPort + "/tcp",
				defaultAdminPort + "/tcp",
			},
			WaitingFor: wait.ForHTTP("/health").WithPort(defaultAdminPort),
		},
	}

	// apply options
	o := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&o)
		}
		opt.Customize(&req)
	}

	req.Image = image(req, o)
	req.Cmd = cmd(o)

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	if o.TLS != nil {
		addTLS(ctx, container, o)
	}

	// start
	if err := container.Start(ctx); err != nil {
		return nil, err
	}
	return &CockroachDBContainer{Container: container, opts: o}, nil
}

func image(req testcontainers.GenericContainerRequest, opts options) string {
	if req.Image != "" {
		return req.Image
	}
	return fmt.Sprintf("%s:%s", defaultImage, opts.ImageTag)
}

func cmd(opts options) []string {
	cmd := []string{
		"start-single-node",
		"--store=type=mem,size=" + opts.StoreSize,
	}

	if opts.TLS != nil {
		cmd = append(cmd, "--certs-dir="+certsDir)
	} else {
		cmd = append(cmd, "--insecure")
	}

	return cmd
}

func addTLS(ctx context.Context, container testcontainers.Container, opts options) error {
	caBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: opts.TLS.CACert.Raw,
	})
	files := map[string][]byte{
		"ca.crt":          caBytes,
		"node.crt":        opts.TLS.NodeCert,
		"node.key":        opts.TLS.NodeKey,
		"client.root.crt": opts.TLS.ClientCert,
		"client.root.key": opts.TLS.ClientKey,
	}
	for filename, contents := range files {
		if err := container.CopyToContainer(ctx, contents, filepath.Join(certsDir, filename), 0o600); err != nil {
			return err
		}
	}
	return nil
}
