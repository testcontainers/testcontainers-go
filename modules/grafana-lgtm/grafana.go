package grafanalgtm

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	GrafanaPort    = "3000/tcp"
	LokiPort       = "3100/tcp"
	TempoPort      = "3200/tcp"
	OtlpGrpcPort   = "4317/tcp"
	OtlpHTTPPort   = "4318/tcp"
	PrometheusPort = "9090/tcp"
)

// GrafanaLGTMContainer represents the Grafana LGTM container type used in the module
type GrafanaLGTMContainer struct {
	testcontainers.Container
}

// Run creates an instance of the Grafana LGTM container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GrafanaLGTMContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{GrafanaPort, LokiPort, TempoPort, OtlpGrpcPort, OtlpHTTPPort, PrometheusPort},
		WaitingFor:   wait.ForLog(".*The OpenTelemetry collector and the Grafana LGTM stack are up and running.*\\s").AsRegexp().WithOccurrence(1),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, fmt.Errorf("generic container: %w", err)
	}

	c := &GrafanaLGTMContainer{Container: container}

	url, err := c.OtlpHTTPEndpoint(ctx)
	if err != nil {
		// return the container instance to allow the caller to clean up
		return c, fmt.Errorf("otlp http endpoint: %w", err)
	}

	testcontainers.Logger.Printf("Access to the Grafana dashboard: %s", url)

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
	url, err := baseEndpoint(ctx, c, LokiPort)
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
	url, err := baseEndpoint(ctx, c, TempoPort)
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

// HTTPEndpoint returns the HTTP URL
func (c *GrafanaLGTMContainer) HTTPEndpoint(ctx context.Context) (string, error) {
	url, err := baseEndpoint(ctx, c, GrafanaPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustHTTPEndpoint returns the HTTP endpoint or panics if an error occurs
func (c *GrafanaLGTMContainer) MustHTTPEndpoint(ctx context.Context) string {
	url, err := c.HTTPEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// OtlpHTTPEndpoint returns the OTLP HTTP endpoint
func (c *GrafanaLGTMContainer) OtlpHTTPEndpoint(ctx context.Context) (string, error) {
	url, err := baseEndpoint(ctx, c, OtlpHTTPPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustOtlpHTTPEndpoint returns the OTLP HTTP endpoint or panics if an error occurs
func (c *GrafanaLGTMContainer) MustOtlpHTTPEndpoint(ctx context.Context) string {
	url, err := c.OtlpHTTPEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// OtlpGrpcEndpoint returns the OTLP gRPC endpoint
func (c *GrafanaLGTMContainer) OtlpGrpcEndpoint(ctx context.Context) (string, error) {
	url, err := baseEndpoint(ctx, c, OtlpGrpcPort)
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

// PrometheusHTTPEndpoint returns the Prometheus HTTP endpoint
func (c *GrafanaLGTMContainer) PrometheusHTTPEndpoint(ctx context.Context) (string, error) {
	url, err := baseEndpoint(ctx, c, PrometheusPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustPrometheusHTTPEndpoint returns the Prometheus HTTP endpoint or panics if an error occurs
func (c *GrafanaLGTMContainer) MustPrometheusHTTPEndpoint(ctx context.Context) string {
	url, err := c.PrometheusHTTPEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

func baseEndpoint(ctx context.Context, c *GrafanaLGTMContainer, port nat.Port) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", host, mappedPort.Port()), nil
}
