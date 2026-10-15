// Package garage provides a Testcontainers module for Garage, an S3-compatible object store.
package garage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents a single-node Garage container.
type Container struct {
	testcontainers.Container
	accessKey string
	secretKey string
	bucket    string
	region    string
}

// AccessKey returns the initial S3 access key ID.
func (c *Container) AccessKey() string { return c.accessKey }

// SecretKey returns the initial S3 secret access key.
func (c *Container) SecretKey() string { return c.secretKey }

// Bucket returns the bucket created during startup.
func (c *Container) Bucket() string { return c.bucket }

// Region returns the S3 region.
func (c *Container) Region() string { return c.region }

// ConnectionString returns the HTTP endpoint for the S3 API.
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "3900/tcp", "http")
}

// Run starts a single-node Garage server with an initial access key and bucket.
// The image must support the bootstrap flags introduced in Garage v2.3.0.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	settings := options{region: "garage"}
	for _, opt := range opts {
		if opt, ok := opt.(Option); ok {
			if err := opt(&settings); err != nil {
				return nil, fmt.Errorf("garage option: %w", err)
			}
		}
	}
	// JSON string escaping is also valid for TOML basic strings.
	region, err := json.Marshal(settings.region)
	if err != nil {
		return nil, fmt.Errorf("encode region: %w", err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate RPC secret: %w", err)
	}
	config := fmt.Sprintf(`metadata_dir = "/var/lib/garage/meta"
data_dir = "/var/lib/garage/data"
db_engine = "sqlite"
replication_factor = 1
rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "%s"
[s3_api]
s3_region = %s
api_bind_addr = "[::]:3900"
`, hex.EncodeToString(secret), region)
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 6+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts("3900/tcp"),
		testcontainers.WithCmd("/garage", "server", "--single-node", "--default-bucket"),
		testcontainers.WithEnv(map[string]string{
			"GARAGE_DEFAULT_ACCESS_KEY": "testcontainers",
			"GARAGE_DEFAULT_SECRET_KEY": "testcontainers-secret",
			"GARAGE_DEFAULT_BUCKET":     "testcontainers",
		}),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            strings.NewReader(config),
			ContainerFilePath: "/etc/garage.toml",
			FileMode:          0o600,
		}),
		// The Garage image has no shell for an internal port check.
		testcontainers.WithWaitStrategy(wait.ForLog("S3 API server listening"), wait.ForListeningPort("3900/tcp").SkipInternalCheck()),
	)
	moduleOpts = append(moduleOpts, opts...)
	moduleOpts = append(moduleOpts, testcontainers.CustomizeRequestOption(validateCredentials))
	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, region: settings.region}
	}
	if err != nil {
		return c, fmt.Errorf("run garage: %w", err)
	}
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect garage: %w", err)
	}
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "GARAGE_DEFAULT_ACCESS_KEY="); ok {
			c.accessKey = v
		}
		if v, ok := strings.CutPrefix(env, "GARAGE_DEFAULT_SECRET_KEY="); ok {
			c.secretKey = v
		}
		if v, ok := strings.CutPrefix(env, "GARAGE_DEFAULT_BUCKET="); ok {
			c.bucket = v
		}
	}
	return c, nil
}
