// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainerFileValidation(t *testing.T) {
	type ContainerFileValidationTestCase struct {
		Name          string
		ExpectedError string
		File          ContainerFile
	}

	f, err := os.Open(filepath.Join(".", "testdata", "hello.sh"))
	require.NoError(t, err)

	testTable := []ContainerFileValidationTestCase{
		{
			Name: "valid container file: has hostfilepath",
			File: ContainerFile{
				HostFilePath:      "/path/to/host",
				ContainerFilePath: "/path/to/container",
			},
		},
		{
			Name: "valid container file: has reader",
			File: ContainerFile{
				Reader:            f,
				ContainerFilePath: "/path/to/container",
			},
		},
		{
			Name:          "invalid container file",
			ExpectedError: "either HostFilePath or Reader must be specified",
			File: ContainerFile{
				HostFilePath:      "",
				Reader:            nil,
				ContainerFilePath: "/path/to/container",
			},
		},
		{
			Name:          "invalid container file",
			ExpectedError: "ContainerFilePath must be specified",
			File: ContainerFile{
				HostFilePath:      "/path/to/host",
				ContainerFilePath: "",
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.File.validate()
			if testCase.ExpectedError != "" {
				require.EqualError(t, err, testCase.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
