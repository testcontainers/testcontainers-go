package testcontainers

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_ContainerValidation(t *testing.T) {

	type ContainerValidationTestCase struct {
		Name             string
		ExpectedError    error
		ContainerRequest ContainerRequest
	}

	testTable := []ContainerValidationTestCase{
		ContainerValidationTestCase{
			Name:          "cannot set both context and image",
			ExpectedError: errors.New("you cannot specify both an Image and Context in a ContainerRequest"),
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{
					Context: ".",
				},
				Image: "redis:latest",
			},
		},
		ContainerValidationTestCase{
			Name:          "can set image without context",
			ExpectedError: nil,
			ContainerRequest: ContainerRequest{
				Image: "redis:latest",
			},
		},
		ContainerValidationTestCase{
			Name:          "can set context without image",
			ExpectedError: nil,
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{
					Context: ".",
				},
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.ContainerRequest.Validate()
			if err == nil && testCase.ExpectedError == nil {
				return
			} else if err == nil && testCase.ExpectedError != nil {
				t.Errorf("did not receive expected error: %s", testCase.ExpectedError.Error())
			} else if err != nil && testCase.ExpectedError == nil {
				t.Errorf("received unexpected error: %s", err.Error())
			} else if err.Error() != testCase.ExpectedError.Error() {
				t.Errorf("errors mismatch: %s != %s", err.Error(), testCase.ExpectedError.Error())
			}
		})
	}

}

func Test_GetDockerfile(t *testing.T) {
	type TestCase struct {
		name                   string
		ExpectedDockerfileName string
		ContainerRequest       ContainerRequest
	}

	testTable := []TestCase{
		TestCase{
			name:                   "defaults to \"Dockerfile\" 1",
			ExpectedDockerfileName: "Dockerfile",
			ContainerRequest:       ContainerRequest{},
		},
		TestCase{
			name:                   "defaults to \"Dockerfile\" 2",
			ExpectedDockerfileName: "Dockerfile",
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{},
			},
		},
		TestCase{
			name:                   "will override name",
			ExpectedDockerfileName: "CustomDockerfile",
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{
					Dockerfile: "CustomDockerfile",
				},
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			n := testCase.ContainerRequest.GetDockerfile()
			if n != testCase.ExpectedDockerfileName {
				t.Fatalf("expected Dockerfile name: %s, received: %s", testCase.ExpectedDockerfileName, n)
			}
		})
	}
}

func Test_BuildImageWithContexts(t *testing.T) {
	type TestCase struct {
		Name               string
		ContextPath        string
		ContextArchive     func() (io.Reader, error)
		ExpectedEchoOutput string
		Dockerfile         string
		ExpectedError      error
	}

	testCases := []TestCase{
		TestCase{
			Name: "test build from context archive",
			ContextArchive: func() (io.Reader, error) {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				files := []struct {
					Name     string
					Contents string
				}{
					{
						Name:     "Dockerfile",
						Contents: `FROM alpine
						CMD ["echo", "this is from the archive"]`,
					},
				}

				for _, f := range files {
					header := tar.Header{
						Name:     f.Name,
						Mode:     0777,
						Size:     int64(len(f.Contents)),
						Typeflag: tar.TypeReg,
						Format:   tar.FormatGNU,
					}

					if err := tarWriter.WriteHeader(&header); err != nil {
						return nil, err
					}

					if _, err := tarWriter.Write([]byte(f.Contents)); err != nil {
						return nil, err
					}

					if err := tarWriter.Close(); err != nil {
						return nil, err
					}
				}

				reader := bytes.NewReader(buf.Bytes())

				return reader, nil
			},
			ExpectedEchoOutput: "this is from the archive",
		},
		TestCase{
			Name: "test build from context archive and be able to use files in it",
			ContextArchive: func() (io.Reader, error) {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				files := []struct {
					Name     string
					Contents string
				}{
					{
						Name: "say_hi.sh",
						Contents: `echo hi this is from the say_hi.sh file!`,
					},
					{
						Name:     "Dockerfile",
						Contents: `FROM alpine
						WORKDIR /app
						COPY . .
						CMD ["sh", "./say_hi.sh"]`,
					},
				}

				for _, f := range files {
					header := tar.Header{
						Name:     f.Name,
						Mode:     0777,
						Size:     int64(len(f.Contents)),
						Typeflag: tar.TypeReg,
						Format:   tar.FormatGNU,
					}

					if err := tarWriter.WriteHeader(&header); err != nil {
						return nil, err
					}

					if _, err := tarWriter.Write([]byte(f.Contents)); err != nil {
						return nil, err
					}
				}

				if err := tarWriter.Close(); err != nil {
					return nil, err
				}

				reader := bytes.NewReader(buf.Bytes())

				return reader, nil
			},
			ExpectedEchoOutput: "hi this is from the say_hi.sh file!",
		},
		TestCase{
			Name:               "test buildling from a context on the filesystem",
			ContextPath:        "./testresources",
			Dockerfile:         "echo.Dockerfile",
			ExpectedEchoOutput: "this is from the echo test Dockerfile",
			ContextArchive: func() (io.Reader, error) {
				return nil, nil
			},
		},
		TestCase{
			Name:        "it should error if neither a context nor a context archive are specified",
			ContextPath: "",
			ContextArchive: func() (io.Reader, error) {
				return nil, nil
			},
			ExpectedError: errors.New("failed to create container: you must specify either a build context or an image"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			a, err := testCase.ContextArchive()
			if err != nil {
				t.Fatal(err)
			}
			req := ContainerRequest{
				FromDockerfile: FromDockerfile{
					ContextArchive: a,
					Context:        testCase.ContextPath,
					Dockerfile:     testCase.Dockerfile,
				},
				WaitingFor: wait.ForLog(testCase.ExpectedEchoOutput).WithStartupTimeout(1 * time.Minute),
			}

			c, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if testCase.ExpectedError != nil && err != nil {
				if testCase.ExpectedError.Error() != err.Error() {
					t.Fatalf("unexpected error: %s, was expecting %s", err.Error(), testCase.ExpectedError.Error())
				}
			} else if err != nil {
				t.Fatal(err)
			} else {
				c.Terminate(ctx)
			}

		})

	}
}
