package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SpannerContainer represents the GCloud container type used in the module for Spanner
type SpannerContainer struct {
	testcontainers.Container
	Settings     options
	GRPCEndpoint string
}

func (c *SpannerContainer) uri(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "9010")
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}

// RunSpannerContainer creates an instance of the GCloud container type for Spanner
func RunSpannerContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*SpannerContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/cloud-spanner-emulator/emulator:1.4.0",
			ExposedPorts: []string{"9010/tcp"},
			WaitingFor:   wait.ForLog("Cloud Spanner emulator running"),
		},
		Started: true,
	}

	settings := applyOptions(req, opts)

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	spannerContainer := SpannerContainer{
		Container: container,
		Settings:  settings,
	}

	uri, err := containerURI(ctx, &spannerContainer)
	if err != nil {
		return nil, err
	}

	spannerContainer.GRPCEndpoint = uri

	return &spannerContainer, nil
}
