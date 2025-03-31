package testcontainers_test

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_ContainerValidation(t *testing.T) {
	type ContainerValidationTestCase struct {
		Name             string
		ExpectedError    string
		ContainerRequest testcontainers.ContainerRequest
	}

	testTable := []ContainerValidationTestCase{
		{
			Name:          "cannot set both context and image",
			ExpectedError: "you cannot specify both an Image and Context in a ContainerRequest",
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: ".",
				},
				Image: "redis:latest",
			},
		},
		{
			Name: "can set image without context",
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
			},
		},
		{
			Name: "can set context without image",
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: ".",
				},
			},
		},
		{
			Name: "Can mount same source to multiple targets",
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.Binds = []string{"/data:/srv", "/data:/data"}
				},
			},
		},
		{
			Name:          "Cannot mount multiple sources to same target",
			ExpectedError: "duplicate mount target detected: /data",
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.Binds = []string{"/data:/data", "/data:/data"}
				},
			},
		},
		{
			Name:          "Invalid bind mount",
			ExpectedError: "invalid bind mount: /data:/data:a:b",
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.Binds = []string{"/data:/data:a:b"}
				},
			},
		},
		{
			Name: "bind-options/provided",
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "redis:latest",
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.Binds = []string{
						"/a:/a:nocopy",
						"/b:/b:ro",
						"/c:/c:rw",
						"/d:/d:z",
						"/e:/e:Z",
						"/f:/f:shared",
						"/g:/g:rshared",
						"/h:/h:slave",
						"/i:/i:rslave",
						"/j:/j:private",
						"/k:/k:rprivate",
						"/l:/l:ro,z,shared",
					}
				},
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.ContainerRequest.Validate()
			if testCase.ExpectedError != "" {
				require.EqualError(t, err, testCase.ExpectedError)
			} else {
				require.NoError(t, err)
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
			require.Equalf(t, n, testCase.ExpectedDockerfileName, "expected Dockerfile name: %s, received: %s", testCase.ExpectedDockerfileName, n)
		})
	}
}

func Test_BuildImageWithContexts(t *testing.T) {
	type TestCase struct {
		Name               string
		ContextPath        string
		ContextArchive     func() (io.ReadSeeker, error)
		ExpectedEchoOutput string
		Dockerfile         string
		ExpectedError      string
	}

	testCases := []TestCase{
		{
			Name: "test build from context archive",
			// fromDockerfileWithContextArchive {
			ContextArchive: func() (io.ReadSeeker, error) {
				var buf bytes.Buffer
				tarWriter := tar.NewWriter(&buf)
				files := []struct {
					Name     string
					Contents string
				}{
					{
						Name: "Dockerfile",
						Contents: `FROM alpine
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
			ContextArchive: func() (io.ReadSeeker, error) {
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
						Contents: `FROM alpine
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
			ContextArchive: func() (io.ReadSeeker, error) {
				return nil, nil
			},
		},
		{
			Name:        "it should error if neither a context nor a context archive are specified",
			ContextPath: "",
			ContextArchive: func() (io.ReadSeeker, error) {
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
			require.NoError(t, err)

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
			testcontainers.CleanupContainer(t, c)

			if testCase.ExpectedError != "" {
				require.EqualError(t, err, testCase.ExpectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCustomLabelsImage(t *testing.T) {
	const (
		myLabelName  = "org.my.label"
		myLabelValue = "my-label-value"
	)

	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:  "alpine:latest",
			Labels: map[string]string{myLabelName: myLabelValue},
		},
	}

	ctr, err := testcontainers.GenericContainer(ctx, req)

	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ctr.Terminate(ctx)) })

	ctrJSON, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	assert.Equal(t, myLabelValue, ctrJSON.Config.Labels[myLabelName])
}

func TestCustomLabelsBuildOptionsModifier(t *testing.T) {
	const (
		myLabelName        = "org.my.label"
		myLabelValue       = "my-label-value"
		myBuildOptionLabel = "org.my.bo.label"
		myBuildOptionValue = "my-bo-label-value"
	)

	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    "./testdata",
				Dockerfile: "Dockerfile",
				BuildOptionsModifier: func(opts *types.ImageBuildOptions) {
					opts.Labels = map[string]string{
						myBuildOptionLabel: myBuildOptionValue,
					}
				},
			},
			Labels: map[string]string{myLabelName: myLabelValue},
		},
	}

	ctr, err := testcontainers.GenericContainer(ctx, req)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	ctrJSON, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, myLabelValue, ctrJSON.Config.Labels[myLabelName])
	require.Equal(t, myBuildOptionValue, ctrJSON.Config.Labels[myBuildOptionLabel])
}

func Test_GetLogsFromFailedContainer(t *testing.T) {
	ctx := context.Background()
	// directDockerHubReference {
	req := testcontainers.ContainerRequest{
		Image:      "alpine",
		Cmd:        []string{"echo", "-n", "I was not expecting this"},
		WaitingFor: wait.ForLog("I was expecting this").WithStartupTimeout(5 * time.Second),
	}
	// }

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	testcontainers.CleanupContainer(t, c)
	require.ErrorContains(t, err, "container exited with code 0")

	logs, logErr := c.Logs(ctx)
	require.NoError(t, logErr)

	b, err := io.ReadAll(logs)
	require.NoError(t, err)

	log := string(b)
	require.Contains(t, log, "I was not expecting this")
}

// dockerImageSubstitutor {
type dockerImageSubstitutor struct{}

func (s dockerImageSubstitutor) Description() string {
	return "DockerImageSubstitutor (prepends registry.hub.docker.com)"
}

func (s dockerImageSubstitutor) Substitute(image string) (string, error) {
	return "registry.hub.docker.com/library/" + image, nil
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
			expectedImage: "registry.hub.docker.com/library/alpine",
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

			ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			testcontainers.CleanupContainer(t, ctr)
			if test.expectedError != nil {
				require.ErrorIs(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)

			// enforce the concrete type, as GenericContainer returns an interface,
			// which will be changed in future implementations of the library
			dockerContainer := ctr.(*testcontainers.DockerContainer)
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
			ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			// mappedPort {
			port, err := ctr.MappedPort(ctx, nginxDefaultPort)
			// }
			require.NoError(t, err)

			t.Logf("Parallel container [iteration_%d] listening on %d\n", i, port.Int())
		})
	}
}

func ExampleGenericContainer_withSubstitutors() {
	ctx := context.Background()

	// applyImageSubstitutors {
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:             "alpine:latest",
			ImageSubstitutors: []testcontainers.ImageSubstitutor{dockerImageSubstitutor{}},
		},
		Started: true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	// }
	if err != nil {
		log.Printf("could not start container: %v", err)
		return
	}

	// enforce the concrete type, as GenericContainer returns an interface,
	// which will be changed in future implementations of the library
	dockerContainer := ctr.(*testcontainers.DockerContainer)

	fmt.Println(dockerContainer.Image)

	// Output: registry.hub.docker.com/library/alpine:latest
}
