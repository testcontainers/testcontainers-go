// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDockerIgnore(t *testing.T) {
	testCases := []struct {
		filePath         string
		exists           bool
		expectedErr      error
		expectedExcluded []string
	}{
		{
			filePath:         "./testdata/dockerignore",
			expectedErr:      nil,
			exists:           true,
			expectedExcluded: []string{"vendor", "foo", "bar"},
		},
		{
			filePath:         "./testdata",
			expectedErr:      nil,
			exists:           true,
			expectedExcluded: []string{"Dockerfile", "echo.Dockerfile"},
		},
		{
			filePath:         "./testdata/data",
			expectedErr:      nil,
			expectedExcluded: nil, // it's nil because the parseDockerIgnore function uses the zero value of a slice
		},
	}

	for _, testCase := range testCases {
		exists, excluded, err := parseDockerIgnore(testCase.filePath)
		assert.Equal(t, testCase.exists, exists)
		require.ErrorIs(t, testCase.expectedErr, err)
		assert.Equal(t, testCase.expectedExcluded, excluded)
	}
}
