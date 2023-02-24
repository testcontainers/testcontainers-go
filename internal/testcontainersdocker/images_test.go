package testcontainersdocker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRegistry(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "Repository + Image + Tag",
			image:    "testcontainers/ryuk:latest",
			expected: "",
		},
		{
			name:     "Repository + Image",
			image:    "testcontainers/ryuk",
			expected: "",
		},
		{
			name:     "Image + Tag",
			image:    "nginx:latest",
			expected: "",
		},
		{
			name:     "Image",
			image:    "nginx",
			expected: "",
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
			actual := ExtractRegistry(test.image)
			assert.Equal(t, test.expected, actual, "expected %s, got %s", test.expected, actual)
		})
	}
}
