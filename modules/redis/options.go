package redis

import (
	"crypto/tls"
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

		if len(req.Cmd) == 0 {
			req.Cmd = []string{redisServerProcess, defaultConfigFile}
			return nil
		}

		// prepend the command to run the redis server with the config file, which must be the first argument of the redis server process
		if req.Cmd[0] == redisServerProcess {
			// just insert the config file, then the rest of the args
			req.Cmd = append([]string{redisServerProcess, defaultConfigFile}, req.Cmd[1:]...)
		} else if req.Cmd[0] != redisServerProcess {
			// prepend the redis server and the config file, then the rest of the args
			req.Cmd = append([]string{redisServerProcess, defaultConfigFile}, req.Cmd...)
		}

		return nil
	}
}

// WithLogLevel sets the log level for the redis server process
// See https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log for more information.
func WithLogLevel(level LogLevel) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		processRedisServerArgs(req, []string{"--loglevel", string(level)})

		return nil
	}
}

// WithSnapshotting sets the snapshotting configuration for the redis server process. You can configure Redis to have it
// save the dataset every N seconds if there are at least M changes in the dataset.
// This method allows Redis to benefit from copy-on-write semantics.
// See https://redis.io/docs/management/persistence/#snapshotting for more information.
func WithSnapshotting(seconds int, changedKeys int) testcontainers.CustomizeRequestOption {
	if changedKeys < 1 {
		changedKeys = 1
	}
	if seconds < 1 {
		seconds = 1
	}

	return func(req *testcontainers.GenericContainerRequest) error {
		processRedisServerArgs(req, []string{"--save", strconv.Itoa(seconds), strconv.Itoa(changedKeys)})
		return nil
	}
}

// createTLSCerts creates a CA certificate, a client certificate and a Redis certificate,
// storing them in the given temporary directory.
func createTLSCerts(tmpDir string) (*tlscert.Certificate, *tlscert.Certificate, *tlscert.Certificate) {
	// ips is the extra list of IPs to include in the certificates.
	// It's used to allow the client and Redis certificates to be used in the same host
	// when the tests are run using a remote docker daemon.
	ips := []net.IP{net.ParseIP("127.0.0.1")}

	// Generate CA certificate
	caCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Host:              "localhost",
		IPAddresses:       ips,
		Name:              "ca",
		SubjectCommonName: "ca",
		IsCA:              true,
		ParentDir:         tmpDir,
	})

	// Generate client certificate
	clientCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Host:              "localhost",
		Name:              "Redis Client",
		SubjectCommonName: "localhost",
		IPAddresses:       ips,
		Parent:            caCert,
		ParentDir:         tmpDir,
	})

	// Generate Redis certificate
	redisCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Host:        "localhost",
		IPAddresses: ips,
		Name:        "Redis Server",
		Parent:      caCert,
		ParentDir:   tmpDir,
	})

	return caCert, clientCert, redisCert
}
