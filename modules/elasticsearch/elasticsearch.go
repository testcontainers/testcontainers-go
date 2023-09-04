package elasticsearch

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

const (
	defaultHTTPPort     = "9200"
	defaultTCPPort      = "9300"
	defaultPassword     = "changeme"
	minimalImageVersion = "7.9.2"
)
const (
	DefaultBaseImage    = "docker.elastic.co/elasticsearch/elasticsearch"
	DefaultBaseImageOSS = "docker.elastic.co/elasticsearch/elasticsearch-oss"
)

// ElasticsearchContainer represents the Elasticsearch container type used in the module
type ElasticsearchContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Elasticsearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ElasticsearchContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: fmt.Sprintf("%s:%s", DefaultBaseImage, minimalImageVersion),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &ElasticsearchContainer{Container: container}, nil
}
