package testcontainers_test

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_ContainerValidation(t *testing.T) {
	type ContainerValidationTestCase struct {
		Name             string
		ExpectedError    error
		ContainerRequest testcontainers.ContainerRequest
	}

	testTable := []ContainerValidationTestCase{
		{
			Name:          "cannot set both context and image",
			ExpectedError: errors.New("you cannot specify both an Image and Context in a ContainerRequest"),
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: ".",
				},
				Image: "redis:latest",
			},
		},
		{
			Name:          "can set image without context",
			ExpectedError: nil,
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
			},
		},
		{
			Name:          "can set context without image",
			ExpectedError: nil,
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: ".",
				},
			},
		},
		{
			Name:          "Can mount same source to multiple targets",
			ExpectedError: nil,
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.Binds = []string{"/data:/srv", "/data:/data"}
				},
			},
		},
		{
			Name:          "Cannot mount multiple sources to same target",
			ExpectedError: errors.New("duplicate mount target detected: /data"),
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.Binds = []string{"/data:/data", "/data:/data"}
				},
			},
		},
		{
			Name:          "Invalid bind mount",
			ExpectedError: errors.New("invalid bind mount: /data:/data:/data"),
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.Binds = []string{"/data:/data:/data"}
				},
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.ContainerRequest.Validate()
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

func Test_GetDockerfile(t *testing.T) {
	type TestCase struct {
		name                   string
		ExpectedDockerfileName string
		ContainerRequest       testcontainers.ContainerRequest
	}

	testTable := []TestCase{
		{
			name:                   "defaults to \"Dockerfile\" 1",
			ExpectedDockerfileName: "Dockerfile",
			ContainerRequest:       testcontainers.ContainerRequest{},
		},
		{
			name:                   "defaults to \"Dockerfile\" 2",
			ExpectedDockerfileName: "Dockerfile",
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{},
			},
		},
		{
			name:                   "will override name",
			ExpectedDockerfileName: "CustomDockerfile",
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
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
		ExpectedError      string
	}

	testCases := []TestCase{
		{
			Name: "test build from context archive",
			// fromDockerfileWithContextArchive {
			ContextArchive: func() (io.Reader, error) {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				files := []struct {
					Name     string
					Contents string
				}{
					{
						Name: "Dockerfile",
						Contents: `FROM docker.io/alpine
								CMD ["echo", "this is from the archive"]`,
					},
				}

				for _, f := range files {
					header := tar.Header{
						Name:     f.Name,
						Mode:     0o777,
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
			// }
			ExpectedEchoOutput: "this is from the archive",
		},
		{
			Name: "test build from context archive and be able to use files in it",
			ContextArchive: func() (io.Reader, error) {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				files := []struct {
					Name     string
					Contents string
				}{
					{
						Name:     "say_hi.sh",
						Contents: `echo hi this is from the say_hi.sh file!`,
					},
					{
						Name: "Dockerfile",
						Contents: `FROM docker.io/alpine
								WORKDIR /app
								COPY . .
								CMD ["sh", "./say_hi.sh"]`,
					},
				}

				for _, f := range files {
					header := tar.Header{
						Name:     f.Name,
						Mode:     0o0777,
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
		{
			Name:               "test building from a context on the filesystem",
			ContextPath:        "./testdata",
			Dockerfile:         "echo.Dockerfile",
			ExpectedEchoOutput: "this is from the echo test Dockerfile",
			ContextArchive: func() (io.Reader, error) {
				return nil, nil
			},
		},
		{
			Name:        "it should error if neither a context nor a context archive are specified",
			ContextPath: "",
			ContextArchive: func() (io.Reader, error) {
				return nil, nil
			},
			ExpectedError: "create container: you must specify either a build context or an image",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			a, err := testCase.ContextArchive()
			if err != nil {
				t.Fatal(err)
			}
			req := testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					ContextArchive: a,
					Context:        testCase.ContextPath,
					Dockerfile:     testCase.Dockerfile,
				},
				WaitingFor: wait.ForLog(testCase.ExpectedEchoOutput).WithStartupTimeout(1 * time.Minute),
			}

			c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})

			defer terminateContainerOnEnd(t, ctx, c)

			if testCase.ExpectedError != "" {
				require.EqualError(t, err, testCase.ExpectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func Test_GetLogsFromFailedContainer(t *testing.T) {
	ctx := context.Background()
	// directDockerHubReference {
	req := testcontainers.ContainerRequest{
		Image:      "docker.io/alpine",
		Cmd:        []string{"echo", "-n", "I was not expecting this"},
		WaitingFor: wait.ForLog("I was expecting this").WithStartupTimeout(5 * time.Second),
	}
	// }

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil && err.Error() != "failed to start container: container exited with code 0" {
		t.Fatal(err)
	} else if err == nil {
		terminateContainerOnEnd(t, ctx, c)
		t.Fatal("was expecting error starting container")
	}

	logs, logErr := c.Logs(ctx)
	if logErr != nil {
		t.Fatal(logErr)
	}

	b, err := io.ReadAll(logs)
	if err != nil {
		t.Fatal(err)
	}

	log := string(b)
	if strings.Contains(log, "I was not expecting this") == false {
		t.Fatalf("could not find expected log in %s", log)
	}
}

// dockerImageSubstitutor {
type dockerImageSubstitutor struct{}

func (s dockerImageSubstitutor) Description() string {
	return "DockerImageSubstitutor (prepends docker.io)"
}

func (s dockerImageSubstitutor) Substitute(image string) (string, error) {
	return "docker.io/" + image, nil
}

// }

// noopImageSubstitutor {
type NoopImageSubstitutor struct{}

// Description returns a description of what is expected from this Substitutor,
// which is used in logs.
func (s NoopImageSubstitutor) Description() string {
	return "NoopImageSubstitutor (noop)"
}

// Substitute returns the original image, without any change
func (s NoopImageSubstitutor) Substitute(image string) (string, error) {
	return image, nil
}

// }

type errorSubstitutor struct{}

var errSubstitution = errors.New("substitution error")

// Description returns a description of what is expected from this Substitutor,
// which is used in logs.
func (s errorSubstitutor) Description() string {
	return "errorSubstitutor"
}

// Substitute returns the original image, but returns an error
func (s errorSubstitutor) Substitute(image string) (string, error) {
	return image, errSubstitution
}

func TestImageSubstitutors(t *testing.T) {
	tests := []struct {
		name          string
		image         string // must be a valid image, as the test will try to create a container from it
		substitutors  []testcontainers.ImageSubstitutor
		expectedImage string
		expectedError error
	}{
		{
			name:          "No substitutors",
			image:         "alpine",
			expectedImage: "alpine",
		},
		{
			name:          "Noop substitutor",
			image:         "alpine",
			substitutors:  []testcontainers.ImageSubstitutor{NoopImageSubstitutor{}},
			expectedImage: "alpine",
		},
		{
			name:          "Prepend namespace",
			image:         "alpine",
			substitutors:  []testcontainers.ImageSubstitutor{dockerImageSubstitutor{}},
			expectedImage: "docker.io/alpine",
		},
		{
			name:          "Substitution with error",
			image:         "alpine",
			substitutors:  []testcontainers.ImageSubstitutor{errorSubstitutor{}},
			expectedImage: "alpine",
			expectedError: errSubstitution,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			req := testcontainers.ContainerRequest{
				Image:             test.image,
				ImageSubstitutors: test.substitutors,
			}

			container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if test.expectedError != nil {
				require.ErrorIs(t, err, test.expectedError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				terminateContainerOnEnd(t, ctx, container)
			}()

			// enforce the concrete type, as GenericContainer returns an interface,
			// which will be changed in future implementations of the library
			dockerContainer := container.(*testcontainers.DockerContainer)
			assert.Equal(t, test.expectedImage, dockerContainer.Image)
		})
	}
}

func TestShouldStartContainersInParallel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	t.Cleanup(cancel)

	for i := 0; i < 3; i++ {
		i := i
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			t.Parallel()

			req := testcontainers.ContainerRequest{
				Image:        nginxAlpineImage,
				ExposedPorts: []string{nginxDefaultPort},
				WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
			}
			container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if err != nil {
				t.Fatalf("could not start container: %v", err)
			}
			// mappedPort {
			port, err := container.MappedPort(ctx, nginxDefaultPort)
			// }
			if err != nil {
				t.Fatalf("could not get mapped port: %v", err)
			}

			terminateContainerOnEnd(t, ctx, container)

			t.Logf("Parallel container [iteration_%d] listening on %d\n", i, port.Int())
		})
	}
}

func ExampleGenericContainer_withSubstitutors() {
	ctx := context.Background()

	// applyImageSubstitutors {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:             "alpine:latest",
			ImageSubstitutors: []testcontainers.ImageSubstitutor{dockerImageSubstitutor{}},
		},
		Started: true,
	})
	// }
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}

	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			log.Fatalf("could not terminate container: %v", err)
		}
	}()

	// enforce the concrete type, as GenericContainer returns an interface,
	// which will be changed in future implementations of the library
	dockerContainer := container.(*testcontainers.DockerContainer)

	fmt.Println(dockerContainer.Image)

	// Output: docker.io/alpine:latest
}
