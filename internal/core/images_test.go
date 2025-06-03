package core

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	localhost5000     = "localhost:5000"
	httpLocalhost5000 = "http://" + localhost5000
	loopback5000      = "127.0.0.1:5000"
	httpLoopback5000  = "http://" + loopback5000
	dockerElasticCo   = "docker.elastic.co"
)

func TestExtractImagesFromDockerfile(t *testing.T) {
	baseImage := "scratch"
	registryHost := "localhost"
	registryPort := "5000"
	nginxImage := "nginx:latest"

	tests := []struct {
		name          string
		dockerfile    string
		buildArgs     map[string]*string
		expected      []string
		expectedError bool
	}{
		{
			name:          "Wrong file",
			dockerfile:    "",
			buildArgs:     nil,
			expected:      []string{},
			expectedError: true,
		},
		{
			name:       "Single Image",
			dockerfile: filepath.Join("testdata", "Dockerfile"),
			buildArgs:  nil,
			expected:   []string{"nginx:${tag}"},
		},
		{
			name:       "Multiple Images",
			dockerfile: filepath.Join("testdata", "Dockerfile.multistage"),
			buildArgs:  nil,
			expected:   []string{"nginx:a", "nginx:b", "nginx:c", "scratch"},
		},
		{
			name:       "Multiple Images with one build arg",
			dockerfile: filepath.Join("testdata", "Dockerfile.multistage.singleBuildArgs"),
			buildArgs:  map[string]*string{"BASE_IMAGE": &baseImage},
			expected:   []string{"nginx:a", "nginx:b", "nginx:c", "scratch"},
		},
		{
			name:       "Multiple Images with multiple build args",
			dockerfile: filepath.Join("testdata", "Dockerfile.multistage.multiBuildArgs"),
			buildArgs:  map[string]*string{"BASE_IMAGE": &baseImage, "REGISTRY_HOST": &registryHost, "REGISTRY_PORT": &registryPort, "NGINX_IMAGE": &nginxImage},
			expected:   []string{"nginx:latest", "localhost:5000/nginx:latest", "scratch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images, err := ExtractImagesFromDockerfile(tt.dockerfile, tt.buildArgs)
			if tt.expectedError {
				require.Error(t, err)
				require.Empty(t, images)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, images)
			}
		})
	}
}

func TestExtractRegistry(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "Empty",
			image:    "",
			expected: "",
		},
		{
			name:     "Numbers",
			image:    "1234567890",
			expected: IndexDockerIO,
		},
		{
			name:     "Malformed Image",
			image:    "--malformed--",
			expected: IndexDockerIO,
		},
		{
			name:     "Repository + Image + Tag",
			image:    "testcontainers/ryuk:latest",
			expected: IndexDockerIO,
		},
		{
			name:     "Repository + Image",
			image:    "testcontainers/ryuk",
			expected: IndexDockerIO,
		},
		{
			name:     "Image + Tag",
			image:    "nginx:latest",
			expected: IndexDockerIO,
		},
		{
			name:     "Image",
			image:    "nginx",
			expected: IndexDockerIO,
		},
		{
			name:     "Local Registry with Port + Repository + Image + Tag",
			image:    "localhost:5000/testcontainers/ryuk:latest",
			expected: localhost5000,
		},
		{
			name:     "Local Registry with Port + Repository + Image",
			image:    "localhost:5000/testcontainers/ryuk",
			expected: localhost5000,
		},
		{
			name:     "Local Registry with Port + Image + Tag",
			image:    "localhost:5000/ryuk:latest",
			expected: localhost5000,
		},
		{
			name:     "Local Registry with Port + Image",
			image:    "localhost:5000/nginx",
			expected: localhost5000,
		},
		{
			name:     "Local Registry with Protocol and Port + Repository + Image + Tag",
			image:    "http://localhost:5000/testcontainers/ryuk:latest",
			expected: httpLocalhost5000,
		},
		{
			name:     "Local Registry with Protocol and Port + Repository + Image",
			image:    "http://localhost:5000/testcontainers/ryuk",
			expected: httpLocalhost5000,
		},
		{
			name:     "Local Registry with Protocol and Port + Image + Tag",
			image:    "http://localhost:5000/ryuk:latest",
			expected: httpLocalhost5000,
		},
		{
			name:     "Local Registry with Protocol and Port + Image",
			image:    "http://localhost:5000/nginx",
			expected: httpLocalhost5000,
		},
		{
			name:     "IP Registry with Port + Repository + Image + Tag",
			image:    "127.0.0.1:5000/testcontainers/ryuk:latest",
			expected: loopback5000,
		},
		{
			name:     "IP Registry with Port + Repository + Image",
			image:    "127.0.0.1:5000/testcontainers/ryuk",
			expected: loopback5000,
		},
		{
			name:     "IP Registry with Port + Image + Tag",
			image:    "127.0.0.1:5000/ryuk:latest",
			expected: loopback5000,
		},
		{
			name:     "IP Registry with Port + Image",
			image:    "127.0.0.1:5000/nginx",
			expected: loopback5000,
		},
		{
			name:     "IP Registry with Protocol and Port + Repository + Image + Tag",
			image:    "http://127.0.0.1:5000/testcontainers/ryuk:latest",
			expected: httpLoopback5000,
		},
		{
			name:     "IP Registry with Protocol and Port + Repository + Image",
			image:    "http://127.0.0.1:5000/testcontainers/ryuk",
			expected: httpLoopback5000,
		},
		{
			name:     "IP Registry with Protocol and Port + Image + Tag",
			image:    "http://127.0.0.1:5000/ryuk:latest",
			expected: httpLoopback5000,
		},
		{
			name:     "IP Registry with Protocol and Port + Image",
			image:    "http://127.0.0.1:5000/nginx",
			expected: httpLoopback5000,
		},
		{
			name:     "DNS Registry + Repository + Image + Tag",
			image:    "docker.elastic.co/elasticsearch/elasticsearch:8.6.2",
			expected: dockerElasticCo,
		},
		{
			name:     "DNS Registry + Repository + Image",
			image:    "docker.elastic.co/elasticsearch/elasticsearch",
			expected: dockerElasticCo,
		},
		{
			name:     "DNS Registry + Image + Tag",
			image:    "docker.elastic.co/elasticsearch:latest",
			expected: dockerElasticCo,
		},
		{
			name:     "DNS Registry + Image",
			image:    "docker.elastic.co/elasticsearch",
			expected: dockerElasticCo,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ExtractRegistry(test.image, IndexDockerIO)
			assert.Equal(t, test.expected, actual, "expected %s, got %s", test.expected, actual)
		})
	}
}
