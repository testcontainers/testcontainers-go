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
	OtlpHttpPort   = "4318/tcp"
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
		ExposedPorts: []string{GrafanaPort, LokiPort, TempoPort, OtlpGrpcPort, OtlpHttpPort, PrometheusPort},
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
		return nil, err
	}

	c := &GrafanaLGTMContainer{Container: container}

	url, err := c.OtlpHttpURL(ctx)
	if err != nil {
		// return the container instance to allow the caller to clean up
		return c, err
	}

	testcontainers.Logger.Printf("Access to the Grafana dashboard: %s", url)

	return c, nil
}

// LokiURL returns the Loki URL
func (c *GrafanaLGTMContainer) LokiURL(ctx context.Context) (string, error) {
	url, err := baseURL(ctx, c, LokiPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustLokiURL returns the Loki URL or panics if an error occurs
func (c *GrafanaLGTMContainer) MustLokiURL(ctx context.Context) string {
	url, err := c.LokiURL(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// TempoURL returns the Tempo URL
func (c *GrafanaLGTMContainer) TempoURL(ctx context.Context) (string, error) {
	url, err := baseURL(ctx, c, TempoPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustTempoURL returns the Tenmpo URL or panics if an error occurs
func (c *GrafanaLGTMContainer) MustTempoURL(ctx context.Context) string {
	url, err := c.TempoURL(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// HttpURL returns the HTTP URL
func (c *GrafanaLGTMContainer) HttpURL(ctx context.Context) (string, error) {
	url, err := baseURL(ctx, c, GrafanaPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustHttpURL returns the HTTP URL or panics if an error occurs
func (c *GrafanaLGTMContainer) MustHttpURL(ctx context.Context) string {
	url, err := c.HttpURL(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// OtlpHttpURL returns the OTLP HTTP URL
func (c *GrafanaLGTMContainer) OtlpHttpURL(ctx context.Context) (string, error) {
	url, err := baseURL(ctx, c, OtlpHttpPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustOtlpHttpURL returns the OTLP HTTP URL or panics if an error occurs
func (c *GrafanaLGTMContainer) MustOtlpHttpURL(ctx context.Context) string {
	url, err := c.OtlpHttpURL(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// OtlpGrpcURL returns the OTLP gRPC URL
func (c *GrafanaLGTMContainer) OtlpGrpcURL(ctx context.Context) (string, error) {
	url, err := baseURL(ctx, c, OtlpGrpcPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustOtlpGrpcURL returns the OTLP gRPC URL or panics if an error occurs
func (c *GrafanaLGTMContainer) MustOtlpGrpcURL(ctx context.Context) string {
	url, err := c.OtlpGrpcURL(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

// PrometheusHttpURL returns the Prometheus HTTP URL
func (c *GrafanaLGTMContainer) PrometheusHttpURL(ctx context.Context) (string, error) {
	url, err := baseURL(ctx, c, PrometheusPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustPrometheusHttp returns the Prometheus HTTP URL or panics if an error occurs
func (c *GrafanaLGTMContainer) MustPrometheusHttp(ctx context.Context) string {
	url, err := c.PrometheusHttpURL(ctx)
	if err != nil {
		panic(err)
	}

	return url
}

func baseURL(ctx context.Context, c *GrafanaLGTMContainer, port nat.Port) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, mappedPort.Port()), nil
}
