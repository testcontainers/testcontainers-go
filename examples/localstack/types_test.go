package localstack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func generateContainerRequest() *LocalStackContainerRequest {
	return &LocalStackContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:          map[string]string{},
			ExposedPorts: []string{},
		},
	}
}

func TestWithRegion(t *testing.T) {
	tests := []struct {
		region         string
		expectedRegion string
	}{
		{
			region:         "",
			expectedRegion: defaultRegion,
		},
		{
			region:         "us-east-1",
			expectedRegion: defaultRegion,
		},
		{
			region:         "eu-west-1",
			expectedRegion: "eu-west-1",
		},
	}

	for _, test := range tests {
		req := generateContainerRequest()

		WithRegion(test.region)(req)
		assert.Equal(t, test.expectedRegion, req.Env["DEFAULT_REGION"])
	}
}

func TestWithServices(t *testing.T) {
	expectedNonLegacyPorts := []string{"4566/tcp"}

	tests := []struct {
		services            []Service
		expectedServices    string
		expectedLegacyPorts []string
	}{
		{
			services:            []Service{},
			expectedServices:    "",
			expectedLegacyPorts: []string{},
		},
		{
			services:            []Service{S3},
			expectedServices:    "s3",
			expectedLegacyPorts: []string{"4572/tcp"},
		},
		{
			services:            []Service{S3, DynamoDB, DynamoDBStreams},
			expectedServices:    "s3,dynamodb,dynamodbstreams",
			expectedLegacyPorts: []string{"4572/tcp", "4569/tcp", "4570/tcp"},
		},
	}

	for _, test := range tests {
		req := generateContainerRequest()
		req.legacyMode = false

		WithServices(test.services...)(req)
		assert.Equal(t, test.expectedServices, req.Env["SERVICES"])
		assert.Equal(t, len(test.services), len(req.enabledServices))
		if len(test.expectedLegacyPorts) > 0 {
			assert.Equal(t, len(expectedNonLegacyPorts), len(req.ExposedPorts))
			assert.Equal(t, expectedNonLegacyPorts, req.ExposedPorts)
		} else {
			assert.Equal(t, []string{}, req.ExposedPorts)
		}

		// legacy mode
		req = generateContainerRequest()
		req.legacyMode = true

		WithServices(test.services...)(req)
		assert.Equal(t, test.expectedServices, req.Env["SERVICES"])
		assert.Equal(t, len(test.services), len(req.enabledServices))
		assert.Equal(t, len(test.expectedLegacyPorts), len(req.ExposedPorts))
		for _, p := range test.expectedLegacyPorts {
			assert.Contains(t, req.ExposedPorts, p)
		}
	}
}

func TestServicePort(t *testing.T) {
	tests := []struct {
		service  Service
		legacy   bool
		expected int
	}{
		{
			service:  S3,
			legacy:   false,
			expected: defaultPort,
		},
		{
			service:  S3,
			legacy:   true,
			expected: 4572,
		},
	}

	for _, test := range tests {
		test.service.legacyMode = test.legacy
		assert.Equal(t, test.expected, test.service.servicePort())
	}

}
