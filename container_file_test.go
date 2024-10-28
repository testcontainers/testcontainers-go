// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainerFileValidation(t *testing.T) {
	type ContainerFileValidationTestCase struct {
		Name          string
		ExpectedError error
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
			ExpectedError: errors.New("either HostFilePath or Reader must be specified"),
			File: ContainerFile{
				HostFilePath:      "",
				Reader:            nil,
				ContainerFilePath: "/path/to/container",
			},
		},
		{
			Name:          "invalid container file",
			ExpectedError: errors.New("ContainerFilePath must be specified"),
			File: ContainerFile{
				HostFilePath:      "/path/to/host",
				ContainerFilePath: "",
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.File.validate()
			switch {
			case err == nil && testCase.ExpectedError == nil:
				return
			case err == nil && testCase.ExpectedError != nil:
				t.Errorf("did not receive expected error: %s", testCase.ExpectedError.Error())
			case err != nil && testCase.ExpectedError == nil:
				t.Errorf("received unexpected error: %s", err.Error())
			case err.Error() != testCase.ExpectedError.Error():
				t.Errorf("errors mismatch: %s != %s", err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
}
