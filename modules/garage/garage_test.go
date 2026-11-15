package garage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/garage"
)

// TestGarage exercises bootstrap and endpoint configuration across supported versions and custom settings.
func TestGarage(t *testing.T) {
	for _, tt := range []struct {
		name                                 string
		image                                string
		opts                                 []testcontainers.ContainerCustomizer
		accessKey, secretKey, bucket, region string
	}{
		{name: "minimum version", image: "dxflrs/garage:v2.3.0", accessKey: "testcontainers", secretKey: "testcontainers-secret", bucket: "testcontainers", region: "garage"},
		{name: "defaults", image: "dxflrs/garage:v2.4.1", accessKey: "testcontainers", secretKey: "testcontainers-secret", bucket: "testcontainers", region: "garage"},
		{name: "custom", image: "dxflrs/garage:v2.4.1", opts: []testcontainers.ContainerCustomizer{garage.WithCredentials("custom-access", "custom-secret-key"), garage.WithBucket("custom-bucket"), garage.WithRegion("eu-west-1")}, accessKey: "custom-access", secretKey: "custom-secret-key", bucket: "custom-bucket", region: "eu-west-1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctr, err := garage.Run(ctx, tt.image, tt.opts...)
			testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
			require.NoError(t, err)
			require.Equal(t, tt.accessKey, ctr.AccessKey())
			require.Equal(t, tt.secretKey, ctr.SecretKey())
			require.Equal(t, tt.bucket, ctr.Bucket())
			require.Equal(t, tt.region, ctr.Region())
			endpoint, err := ctr.ConnectionString(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, endpoint)
		})
	}
}

// TestInvalidOptions verifies invalid settings fail before a container is created.
func TestInvalidOptions(t *testing.T) {
	for _, tt := range []struct {
		name string
		opt  testcontainers.ContainerCustomizer
		want string
	}{
		{"access key", garage.WithCredentials("short", "valid-secret-key"), "access key must be at least 8 bytes"},
		{"secret key", garage.WithCredentials("valid-access", "short"), "secret key must be at least 16 bytes"},
		{"bucket", garage.WithBucket(""), "bucket must not be empty"},
		{"region", garage.WithRegion(""), "region must not be empty"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctr, err := garage.Run(context.Background(), "dxflrs/garage:v2.4.1", tt.opt)
			testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
			require.ErrorContains(t, err, tt.want)
			require.Nil(t, ctr)
		})
	}
}
