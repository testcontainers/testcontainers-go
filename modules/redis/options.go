package redis

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"

	"github.com/mdelapenya/tlscert"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	tlsEnabled bool
	tlsConfig  *tls.Config
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithTLS sets the TLS configuration for the redis container, setting
// the 6380/tcp port to listen on for TLS connections and using a secure URL (rediss://).
func WithTLS() Option {
	return func(o *options) error {
		o.tlsEnabled = true
		return nil
	}
}

// WithConfigFile sets the config file to be used for the redis container, and sets the command to run the redis server
// using the passed config file
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	const defaultConfigFile = "/usr/local/redis.conf"

	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: defaultConfigFile,
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		// prepend the command to run the redis server with the config file, which must be the first argument of the redis server process
		req.Cmd = append([]string{defaultConfigFile}, req.Cmd...)

		return nil
	}
}

// WithLogLevel sets the log level for the redis server process
// See "[RedisModule_Log]" for more information.
//
// [RedisModule_Log]: https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log
func WithLogLevel(level LogLevel) testcontainers.CustomizeRequestOption {
	return testcontainers.WithCmdArgs("--loglevel", string(level))
}

// WithSnapshotting sets the snapshotting configuration for the redis server process. You can configure Redis to have it
// save the dataset every N seconds if there are at least M changes in the dataset.
// This method allows Redis to benefit from copy-on-write semantics.
// See [Snapshotting] for more information.
//
// [Snapshotting]: https://redis.io/docs/management/persistence/#snapshotting
func WithSnapshotting(seconds int, changedKeys int) testcontainers.CustomizeRequestOption {
	if changedKeys < 1 {
		changedKeys = 1
	}
	if seconds < 1 {
		seconds = 1
	}

	return testcontainers.WithCmdArgs("--save", strconv.Itoa(seconds), strconv.Itoa(changedKeys))
}

// createTLSCerts creates a CA certificate, a client certificate and a Redis certificate.
func createTLSCerts() (caCert *tlscert.Certificate, clientCert *tlscert.Certificate, redisCert *tlscert.Certificate, err error) {
	// ips is the extra list of IPs to include in the certificates.
	// It's used to allow the client and Redis certificates to be used in the same host
	// when the tests are run using a remote docker daemon.
	ips := []net.IP{net.ParseIP("127.0.0.1")}

	// Generate CA certificate
	caCert, err = tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              "localhost",
		IPAddresses:       ips,
		Name:              "ca",
		SubjectCommonName: "ca",
		IsCA:              true,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate CA certificate: %w", err)
	}

	// Generate client certificate
	clientCert, err = tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              "localhost",
		Name:              "Redis Client",
		SubjectCommonName: "localhost",
		IPAddresses:       ips,
		Parent:            caCert,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate client certificate: %w", err)
	}

	// Generate Redis certificate
	redisCert, err = tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:        "localhost",
		IPAddresses: ips,
		Name:        "Redis Server",
		Parent:      caCert,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate Redis certificate: %w", err)
	}

	return caCert, clientCert, redisCert, nil
}
