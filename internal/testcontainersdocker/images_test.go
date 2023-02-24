package testcontainersdocker

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractImagesFromDockerfile(t *testing.T) {
	tests := []struct {
		name          string
		dockerfile    string
		expected      []string
		expectedError bool
	}{
		{
			name:          "Wrong file",
			dockerfile:    "",
			expected:      []string{},
			expectedError: true,
		},
		{
			name:       "Single Image",
			dockerfile: filepath.Join("testresources", "Dockerfile"),
			expected:   []string{"nginx:${tag}"},
		},
		{
			name:       "Multiple Images",
			dockerfile: filepath.Join("testresources", "Dockerfile.multistage"),
			expected:   []string{"nginx:a", "nginx:b", "nginx:c", "scratch"},
		},
	}

	for _, tt := range tests {
		images, err := ExtractImagesFromDockerfile(tt.dockerfile)
		if tt.expectedError {
			require.Error(t, err)
			assert.Empty(t, images)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tt.expected, images)
		}
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
			expected: "localhost:5000",
		},
		{
			name:     "Local Registry with Port + Repository + Image",
			image:    "localhost:5000/testcontainers/ryuk",
			expected: "localhost:5000",
		},
		{
			name:     "Local Registry with Port + Image + Tag",
			image:    "localhost:5000/ryuk:latest",
			expected: "localhost:5000",
		},
		{
			name:     "Local Registry with Port + Image",
			image:    "localhost:5000/nginx",
			expected: "localhost:5000",
		},
		{
			name:     "Local Registry with Protocol and Port + Repository + Image + Tag",
			image:    "http://localhost:5000/testcontainers/ryuk:latest",
			expected: "http://localhost:5000",
		},
		{
			name:     "Local Registry with Protocol and Port + Repository + Image",
			image:    "http://localhost:5000/testcontainers/ryuk",
			expected: "http://localhost:5000",
		},
		{
			name:     "Local Registry with Protocol and Port + Image + Tag",
			image:    "http://localhost:5000/ryuk:latest",
			expected: "http://localhost:5000",
		},
		{
			name:     "Local Registry with Protocol and Port + Image",
			image:    "http://localhost:5000/nginx",
			expected: "http://localhost:5000",
		},
		{
			name:     "IP Registry with Port + Repository + Image + Tag",
			image:    "127.0.0.1:5000/testcontainers/ryuk:latest",
			expected: "127.0.0.1:5000",
		},
		{
			name:     "IP Registry with Port + Repository + Image",
			image:    "127.0.0.1:5000/testcontainers/ryuk",
			expected: "127.0.0.1:5000",
		},
		{
			name:     "IP Registry with Port + Image + Tag",
			image:    "127.0.0.1:5000/ryuk:latest",
			expected: "127.0.0.1:5000",
		},
		{
			name:     "IP Registry with Port + Image",
			image:    "127.0.0.1:5000/nginx",
			expected: "127.0.0.1:5000",
		},
		{
			name:     "IP Registry with Protocol and Port + Repository + Image + Tag",
			image:    "http://127.0.0.1:5000/testcontainers/ryuk:latest",
			expected: "http://127.0.0.1:5000",
		},
		{
			name:     "IP Registry with Protocol and Port + Repository + Image",
			image:    "http://127.0.0.1:5000/testcontainers/ryuk",
			expected: "http://127.0.0.1:5000",
		},
		{
			name:     "IP Registry with Protocol and Port + Image + Tag",
			image:    "http://127.0.0.1:5000/ryuk:latest",
			expected: "http://127.0.0.1:5000",
		},
		{
			name:     "IP Registry with Protocol and Port + Image",
			image:    "http://127.0.0.1:5000/nginx",
			expected: "http://127.0.0.1:5000",
		},
		{
			name:     "DNS Registry + Repository + Image + Tag",
			image:    "docker.elastic.co/elasticsearch/elasticsearch:8.6.2",
			expected: "docker.elastic.co",
		},
		{
			name:     "DNS Registry + Repository + Image",
			image:    "docker.elastic.co/elasticsearch/elasticsearch",
			expected: "docker.elastic.co",
		},
		{
			name:     "DNS Registry + Image + Tag",
			image:    "docker.elastic.co/elasticsearch:latest",
			expected: "docker.elastic.co",
		},
		{
			name:     "DNS Registry + Image",
			image:    "docker.elastic.co/elasticsearch",
			expected: "docker.elastic.co",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ExtractRegistry(test.image, IndexDockerIO)
			assert.Equal(t, test.expected, actual, "expected %s, got %s", test.expected, actual)
		})
	}
}
