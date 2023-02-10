package localstack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func generateContainerRequest() *LocalStackContainerRequest {
	return &LocalStackContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:          map[string]string{},
			ExposedPorts: []string{},
		},
	}
}

func TestWithContainerRequest(t *testing.T) {
	req := testcontainers.ContainerRequest{
		Env:          map[string]string{},
		Image:        "foo",
		ExposedPorts: []string{},
		SkipReaper:   false,
		WaitingFor: wait.ForNop(
			func(ctx context.Context, target wait.StrategyTarget) error {
				return nil
			},
		),
		Networks: []string{"foo", "bar", "baaz"},
		NetworkAliases: map[string][]string{
			"foo": {"foo0", "foo1", "foo2", "foo3"},
		},
	}

	merged := OverrideContainerRequest(testcontainers.ContainerRequest{
		Env: map[string]string{
			"FOO": "BAR",
		},
		Image:    "bar",
		Networks: []string{"foo1", "bar1"},
		NetworkAliases: map[string][]string{
			"foo1": {"bar"},
		},
		WaitingFor: wait.ForLog("foo"),
		SkipReaper: true,
	})(req)

	assert.Equal(t, "BAR", merged.Env["FOO"])
	assert.True(t, merged.SkipReaper)
	assert.Equal(t, "bar", merged.Image)
	assert.Equal(t, []string{"foo1", "bar1"}, merged.Networks)
	assert.Equal(t, []string{"foo0", "foo1", "foo2", "foo3"}, merged.NetworkAliases["foo"])
	assert.Equal(t, []string{"bar"}, merged.NetworkAliases["foo1"])
	assert.Equal(t, wait.ForLog("foo"), merged.WaitingFor)
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

		WithServices(test.services...)(req)
		assert.Equal(t, test.expectedServices, req.Env["SERVICES"])
		assert.Equal(t, len(test.services), len(req.enabledServices))
		if len(test.expectedLegacyPorts) > 0 {
			assert.Equal(t, len(expectedNonLegacyPorts), len(req.ExposedPorts))
			assert.Equal(t, expectedNonLegacyPorts, req.ExposedPorts)
		} else {
			assert.Equal(t, []string{}, req.ExposedPorts)
		}
	}
}

func TestServicePort(t *testing.T) {
	tests := []struct {
		service  Service
		expected int
	}{
		{
			service:  S3,
			expected: defaultPort,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.service.servicePort())
	}

}
