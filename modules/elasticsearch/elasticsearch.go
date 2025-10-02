package elasticsearch

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultHTTPPort   = "9200"
	defaultTCPPort    = "9300"
	defaultPassword   = "changeme"
	defaultUsername   = "elastic"
	defaultCaCertPath = "/usr/share/elasticsearch/config/certs/http_ca.crt"
	envPassword       = "ELASTIC_PASSWORD"
)

const (
	// Deprecated: it will be removed in the next major version
	DefaultBaseImage = "docker.elastic.co/elasticsearch/elasticsearch"
	// Deprecated: it will be removed in the next major version
	DefaultBaseImageOSS = "docker.elastic.co/elasticsearch/elasticsearch-oss"
)

// ElasticsearchContainer represents the Elasticsearch container type used in the module
type ElasticsearchContainer struct {
	testcontainers.Container
	Settings Options
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Elasticsearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ElasticsearchContainer, error) {
	return Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:7.9.2", opts...)
}

// Run creates an instance of the Elasticsearch container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ElasticsearchContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultHTTPPort+"/tcp", defaultTCPPort+"/tcp"),
		testcontainers.WithEnv(map[string]string{
			"discovery.type": "single-node",
			"cluster.routing.allocation.disk.threshold_enabled": "false",
		}),
	}

	// Gather all config options (defaults and then apply provided options)
	options := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(options)
		}
	}

	moduleOpts = append(moduleOpts, opts...)

	// Transfer the password settings to the container request
	moduleOpts = append(moduleOpts, configurePassword(options))
	moduleOpts = append(moduleOpts, configureJvmOpts())

	// Set the default waiting strategy if not already set.
	moduleOpts = append(moduleOpts, configureWaitFor(options))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var esContainer *ElasticsearchContainer
	if ctr != nil {
		esContainer = &ElasticsearchContainer{Container: ctr, Settings: *options}
	}
	if err != nil {
		return esContainer, fmt.Errorf("generic container: %w", err)
	}

	if err := esContainer.configureAddress(ctx); err != nil {
		return esContainer, fmt.Errorf("configure address: %w", err)
	}

	return esContainer, nil
}

// certWriter is a helper that writes the details of a CA cert to options.
type certWriter struct {
	options  *Options
	certPool *x509.CertPool
}

// Read reads the CA cert from the reader and appends it to the options.
func (w *certWriter) Read(r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read CA cert: %w", err)
	}

	w.options.CACert = buf
	w.certPool.AppendCertsFromPEM(w.options.CACert)

	return nil
}

// configureWaitFor sets the req.WaitingFor strategy based on settings.
func configureWaitFor(options *Options) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		var strategies []wait.Strategy

		waitHTTP := wait.ForHTTP("/").WithPort(defaultHTTPPort)
		if sslRequired(req) {
			cw := &certWriter{
				options:  options,
				certPool: x509.NewCertPool(),
			}

			waitHTTP = waitHTTP.
				WithTLS(true, &tls.Config{RootCAs: cw.certPool}).WithAllowInsecure(true)

			strategies = append(strategies, wait.ForFile(defaultCaCertPath).WithMatcher(cw.Read))
		}

		if options.Password != "" || options.Username != "" {
			waitHTTP = waitHTTP.WithBasicAuth(options.Username, options.Password)
		}

		strategies = append(strategies, waitHTTP)

		return testcontainers.WithAdditionalWaitStrategy(strategies...)(req)
	}
}

// configureAddress sets the address of the Elasticsearch container.
// If the certificate is set, it will use https as protocol, otherwise http.
func (c *ElasticsearchContainer) configureAddress(ctx context.Context) error {
	proto := "http"
	if c.Settings.CACert != nil {
		proto = "https"
	}

	endpoint, err := c.PortEndpoint(ctx, defaultHTTPPort+"/tcp", proto)
	if err != nil {
		return fmt.Errorf("port endpoint: %w", err)
	}

	c.Settings.Address = endpoint

	return nil
}

// sslRequired returns true if the SSL is required, otherwise false.
func sslRequired(req *testcontainers.GenericContainerRequest) bool {
	if !isAtLeastVersion(req.Image, 8) {
		return false
	}

	// These configuration keys explicitly disable CA generation.
	// If any are set we skip the file retrieval.
	configKeys := []string{
		"xpack.security.enabled",
		"xpack.security.http.ssl.enabled",
		"xpack.security.transport.ssl.enabled",
	}
	for _, configKey := range configKeys {
		if value, ok := req.Env[configKey]; ok {
			if value == "false" {
				return false
			}
		}
	}

	return true
}

// configurePassword transfers the password settings to the container request.
// If the password is not set, it will be set to "changeme" for Elasticsearch 8
func configurePassword(settings *Options) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		// set "changeme" as default password for Elasticsearch 8
		if isAtLeastVersion(req.Image, 8) && settings.Password == "" {
			WithPassword(defaultPassword)(settings)
		}

		if settings.Password != "" {
			if isOSS(req.Image) {
				return errors.New("it's not possible to activate security on Elastic OSS Image. Please switch to the default distribution")
			}

			if _, ok := req.Env[envPassword]; !ok {
				req.Env[envPassword] = settings.Password
			}

			// major version 8 is secure by default and does not need this to enable authentication
			if !isAtLeastVersion(req.Image, 8) {
				req.Env["xpack.security.enabled"] = "true"
			}
		}

		return nil
	}
}

// configureJvmOpts sets the default memory of the Elasticsearch instance to 2GB.
// This functions, which is only available since version 7, is called as a post create hook
// for the container request.
func configureJvmOpts() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if !isAtLeastVersion(req.Image, 7) {
			return nil
		}

		return testcontainers.WithAdditionalLifecycleHooks(
			testcontainers.ContainerLifecycleHooks{
				PostCreates: []testcontainers.ContainerHook{
					func(ctx context.Context, container testcontainers.Container) error {
						// Sets default memory of elasticsearch instance to 2GB
						defaultJVMOpts := `-Xms2G
-Xmx2G
-Dingest.geoip.downloader.enabled.default=false
`

						tmpDir := os.TempDir()

						tmpFile, err := os.CreateTemp(tmpDir, "elasticsearch-default-memory-vm.options")
						if err != nil {
							return err
						}
						defer os.Remove(tmpFile.Name()) // clean up

						if _, err := tmpFile.WriteString(defaultJVMOpts); err != nil {
							return err
						}

						// Spaces are deliberate to allow user to define additional jvm options as elasticsearch resolves option files lexicographically
						if err := container.CopyFileToContainer(
							ctx, tmpFile.Name(),
							"/usr/share/elasticsearch/config/jvm.options.d/ elasticsearch-default-memory-vm.options", 0o644); err != nil {
							return err
						}

						return nil
					},
				},
			},
		)(req)
	}
}
