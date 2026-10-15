package garage

import (
	"errors"

	"github.com/testcontainers/testcontainers-go"
)

type options struct{ region string }

// Option configures Garage-specific settings.
type Option func(*options) error

var _ testcontainers.ContainerCustomizer = Option(nil)

// Customize implements testcontainers.ContainerCustomizer. Options are applied by Run.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error { return nil }

// WithRegion sets the S3 region. The default is "garage".
func WithRegion(region string) Option {
	return func(o *options) error {
		if region == "" {
			return errors.New("region must not be empty")
		}
		o.region = region
		return nil
	}
}

// WithCredentials sets the initial S3 access key ID and secret access key.
// Garage requires at least 8 bytes for the access key and 16 bytes for the secret key.
func WithCredentials(accessKey, secretKey string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"GARAGE_DEFAULT_ACCESS_KEY": accessKey,
		"GARAGE_DEFAULT_SECRET_KEY": secretKey,
	})
}

// WithBucket sets the bucket created at startup. The default is "testcontainers".
func WithBucket(bucket string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{"GARAGE_DEFAULT_BUCKET": bucket})
}

func validateCredentials(req *testcontainers.GenericContainerRequest) error {
	if len(req.Env["GARAGE_DEFAULT_ACCESS_KEY"]) < 8 {
		return errors.New("access key must be at least 8 bytes")
	}
	if len(req.Env["GARAGE_DEFAULT_SECRET_KEY"]) < 16 {
		return errors.New("secret key must be at least 16 bytes")
	}
	if req.Env["GARAGE_DEFAULT_BUCKET"] == "" {
		return errors.New("bucket must not be empty")
	}
	return nil
}
