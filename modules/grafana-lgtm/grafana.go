package grafanalgtm

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	GrafanaPort = "3000/tcp"

	OtlpGrpcPort = "4317/tcp"

	OtlpHttpPort = "4318/tcp"

	PrometheusPort = "9090/tcp"
)

// GrafanaContainer represents the Grafana container type used in the module
type GrafanaContainer struct {
	testcontainers.Container
}

// Run creates an instance of the Grafana container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GrafanaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{GrafanaPort, OtlpGrpcPort, OtlpHttpPort, PrometheusPort},
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

	return &GrafanaContainer{Container: container}, nil
}

// OtlpHttpURL returns the OTLP HTTP URL
func (c *GrafanaContainer) OtlpHttpURL() (string, error) {
	url, err := baseURL(c, OtlpHttpPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustOtlpHttpURL returns the OTLP HTTP URL or panics if an error occurs
func (c *GrafanaContainer) MustOtlpHttpURL() string {
	url, err := c.OtlpHttpURL()
	if err != nil {
		panic(err)
	}

	return url
}

// OtlpGrpcURL returns the OTLP gRPC URL
func (c *GrafanaContainer) OtlpGrpcURL() (string, error) {
	url, err := baseURL(c, OtlpGrpcPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustOtlpGrpcURL returns the OTLP gRPC URL or panics if an error occurs
func (c *GrafanaContainer) MustOtlpGrpcURL() string {
	url, err := c.OtlpGrpcURL()
	if err != nil {
		panic(err)
	}

	return url
}

// PrometheusHttpURL returns the Prometheus HTTP URL
func (c *GrafanaContainer) PrometheusHttpURL() (string, error) {
	url, err := baseURL(c, PrometheusPort)
	if err != nil {
		return "", err
	}

	return url, nil
}

// MustPrometheusHttp returns the Prometheus HTTP URL or panics if an error occurs
func (c *GrafanaContainer) MustPrometheusHttp() string {
	url, err := c.PrometheusHttpURL()
	if err != nil {
		panic(err)
	}

	return url
}

func baseURL(c *GrafanaContainer, port nat.Port) (string, error) {
	host, err := c.Host(context.Background())
	if err != nil {
		return "", err
	}

	mappedPort, err := c.MappedPort(context.Background(), port)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, mappedPort.Port()), nil
}
