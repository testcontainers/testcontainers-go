// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDockerIgnore(t *testing.T) {
	testCases := []struct {
		filePath          string
		expectedErr       error
		expectedExcluded  []string
		expectedHasIgnore bool
	}{
		{
			filePath:          "./testdata/dockerignore",
			expectedErr:       nil,
			expectedExcluded:  []string{"vendor", "foo", "bar"},
			expectedHasIgnore: true,
		},
		{
			filePath:          "./testdata",
			expectedErr:       nil,
			expectedExcluded:  []string{"Dockerfile", "echo.Dockerfile"},
			expectedHasIgnore: true,
		},
	}

	for _, testCase := range testCases {
		excluded, hasIgnore, err := parseDockerIgnore(testCase.filePath)
		assert.Equal(t, testCase.expectedErr, err)
		assert.Equal(t, testCase.expectedExcluded, excluded)
		assert.Equal(t, testCase.expectedHasIgnore, hasIgnore)
	}
}
