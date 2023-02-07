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
		Networks: []string{"foo1", "bar1"},
		NetworkAliases: map[string][]string{
			"foo1": {"bar"},
		},
		WaitingFor: wait.ForLog("foo"),
		SkipReaper: true,
	})(req)

	assert.Equal(t, "BAR", merged.Env["FOO"])
	assert.True(t, merged.SkipReaper)
	assert.Equal(t, []string{"foo1", "bar1"}, merged.Networks)
	assert.Equal(t, []string{"foo0", "foo1", "foo2", "foo3"}, merged.NetworkAliases["foo"])
	assert.Equal(t, []string{"bar"}, merged.NetworkAliases["foo1"])
	assert.Equal(t, wait.ForLog("foo"), merged.WaitingFor)
}

func TestWithCredentials(t *testing.T) {
	tests := []struct {
		cred     Credentials
		expected Credentials
	}{
		{
			cred: Credentials{
				AccessKeyID:     "",
				SecretAccessKey: "",
				Token:           "",
			},
			expected: Credentials{},
		},
		{
			cred: Credentials{
				AccessKeyID:     "foo",
				SecretAccessKey: "bar",
				Token:           "baz",
			},
			expected: Credentials{
				AccessKeyID:     "foo",
				SecretAccessKey: "bar",
				Token:           "baz",
			},
		},
	}

	for _, test := range tests {
		req := generateContainerRequest()

		WithCredentials(test.cred)(req)
		assert.Equal(t, test.expected.AccessKeyID, req.Env["AWS_ACCESS_KEY_ID"])
		assert.Equal(t, test.expected.SecretAccessKey, req.Env["AWS_SECRET_ACCESS_KEY"])
		assert.Equal(t, test.expected.Token, req.Env["AWS_SESSION_TOKEN"])
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
		assert.Equal(t, test.expectedRegion, req.region)
	}
}

func TestWithLegacyMode(t *testing.T) {
	tests := []struct {
		legacyMode         bool
		expectedLegacyMode bool
	}{
		{
			legacyMode:         false,
			expectedLegacyMode: false,
		},
		{
			legacyMode:         true,
			expectedLegacyMode: true,
		},
	}

	for _, test := range tests {
		req := generateContainerRequest()

		if test.legacyMode {
			WithLegacyMode(req)
		}

		assert.Equal(t, test.expectedLegacyMode, req.legacyMode)
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

func TestWithVersion(t *testing.T) {
	tests := []struct {
		version         string
		expectedVersion string
	}{
		{
			version:         "",
			expectedVersion: defaultVersion,
		},
		{
			version:         "0.10.5",
			expectedVersion: "0.10.5",
		},
	}

	for _, test := range tests {
		req := generateContainerRequest()

		WithVersion(test.version)(req)
		assert.Equal(t, test.expectedVersion, req.version)
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
