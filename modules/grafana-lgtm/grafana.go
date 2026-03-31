package grafanalgtm

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/api/types/network"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	GrafanaPort    = "3000/tcp"
	LokiPort       = "3100/tcp"
	TempoPort      = "3200/tcp"
	OtlpGrpcPort   = "4317/tcp"
	OtlpHttpPort   = "4318/tcp" //nolint:revive,staticcheck //FIXME
	PrometheusPort = "9090/tcp"
)

// GrafanaLGTMContainer represents the Grafana LGTM container type used in the module
type GrafanaLGTMContainer struct {
	testcontainers.Container
}

// Run creates an instance of the Grafana LGTM container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GrafanaLGTMContainer, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(GrafanaPort, LokiPort, TempoPort, OtlpGrpcPort, OtlpHttpPort, PrometheusPort),
		testcontainers.WithWaitStrategyAndDeadline(2*time.Minute,
			wait.ForLog(".*The OpenTelemetry collector and the Grafana LGTM stack are up and running.*\\s").AsRegexp().WithOccurrence(1),
			wait.ForListeningPort(network.MustParsePort(GrafanaPort)),
			wait.ForListeningPort(network.MustParsePort(LokiPort)),
			wait.ForListeningPort(network.MustParsePort(TempoPort)),
			wait.ForListeningPort(network.MustParsePort(OtlpGrpcPort)),
			wait.ForListeningPort(network.MustParsePort(OtlpHttpPort)),
			wait.ForListeningPort(network.MustParsePort(PrometheusPort)),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	var c *GrafanaLGTMContainer
	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	if ctr != nil {
		c = &GrafanaLGTMContainer{Container: ctr}
	}
	if err != nil {
		return nil, fmt.Errorf("run grafana lgtm: %w", err)
	}

	url, err := c.OtlpHttpEndpoint(ctx)
	if err != nil {
		// return the container instance to allow the caller to clean up
		return c, fmt.Errorf("otlp http endpoint: %w", err)
	}

	log.Printf("Access to the Grafana dashboard: %s", url)

	return c, nil
}

// WithAdminCredentials sets the admin credentials for the Grafana LGTM container
func WithAdminCredentials(user, password string) testcontainers.ContainerCustomizer {
	return testcontainers.WithEnv(map[string]string{
		"GF_SECURITY_ADMIN_USER":     user,
		"GF_SECURITY_ADMIN_PASSWORD": password,
	})
}

// LokiEndpoint returns the Loki endpoint
func (c *GrafanaLGTMContainer) LokiEndpoint(ctx context.Context) (string, error) {
	url, err := c.PortEndpoint(ctx, network.MustParsePort(LokiPort), "")
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustLokiEndpoint returns the Loki endpoint or panics if an error occurs
func (c *GrafanaLGTMContainer) MustLokiEndpoint(ctx context.Context) string {
	url, err := c.LokiEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// TempoEndpoint returns the Tempo endpoint
func (c *GrafanaLGTMContainer) TempoEndpoint(ctx context.Context) (string, error) {
	url, err := c.PortEndpoint(ctx, network.MustParsePort(TempoPort), "")
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustTempoEndpoint returns the Tempo endpoint or panics if an error occurs
func (c *GrafanaLGTMContainer) MustTempoEndpoint(ctx context.Context) string {
	url, err := c.TempoEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// HttpEndpoint returns the HTTP URL
//
//nolint:revive,staticcheck //FIXME
func (c *GrafanaLGTMContainer) HttpEndpoint(ctx context.Context) (string, error) {
	url, err := c.PortEndpoint(ctx, network.MustParsePort(GrafanaPort), "")
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustHttpEndpoint returns the HTTP endpoint or panics if an error occurs
//
//nolint:revive,staticcheck //FIXME
func (c *GrafanaLGTMContainer) MustHttpEndpoint(ctx context.Context) string {
	url, err := c.HttpEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// OtlpHttpEndpoint returns the OTLP HTTP endpoint
//
//nolint:revive,staticcheck //FIXME
func (c *GrafanaLGTMContainer) OtlpHttpEndpoint(ctx context.Context) (string, error) {
	url, err := c.PortEndpoint(ctx, network.MustParsePort(OtlpHttpPort), "")
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustOtlpHttpEndpoint returns the OTLP HTTP endpoint or panics if an error occurs
//
//nolint:revive,staticcheck //FIXME
func (c *GrafanaLGTMContainer) MustOtlpHttpEndpoint(ctx context.Context) string {
	url, err := c.OtlpHttpEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// OtlpGrpcEndpoint returns the OTLP gRPC endpoint
func (c *GrafanaLGTMContainer) OtlpGrpcEndpoint(ctx context.Context) (string, error) {
	url, err := c.PortEndpoint(ctx, network.MustParsePort(OtlpGrpcPort), "")
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustOtlpGrpcEndpoint returns the OTLP gRPC endpoint or panics if an error occurs
func (c *GrafanaLGTMContainer) MustOtlpGrpcEndpoint(ctx context.Context) string {
	url, err := c.OtlpGrpcEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// PrometheusHttpEndpoint returns the Prometheus HTTP endpoint
//
//nolint:revive,staticcheck //FIXME
func (c *GrafanaLGTMContainer) PrometheusHttpEndpoint(ctx context.Context) (string, error) {
	url, err := c.PortEndpoint(ctx, network.MustParsePort(PrometheusPort), "")
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustPrometheusHttpEndpoint returns the Prometheus HTTP endpoint or panics if an error occurs
//
//nolint:revive,staticcheck //FIXME
func (c *GrafanaLGTMContainer) MustPrometheusHttpEndpoint(ctx context.Context) string {
	url, err := c.PrometheusHttpEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}
