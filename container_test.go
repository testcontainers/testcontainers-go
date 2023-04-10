package testcontainers

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_ContainerValidation(t *testing.T) {

	type ContainerValidationTestCase struct {
		Name             string
		ExpectedError    error
		ContainerRequest ContainerRequest
	}

	testTable := []ContainerValidationTestCase{
		{
			Name:          "cannot set both context and image",
			ExpectedError: errors.New("you cannot specify both an Image and Context in a ContainerRequest"),
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{
					Context: ".",
				},
				Image: "redis:latest",
			},
		},
		{
			Name:          "can set image without context",
			ExpectedError: nil,
			ContainerRequest: ContainerRequest{
				Image: "redis:latest",
			},
		},
		{
			Name:          "can set context without image",
			ExpectedError: nil,
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{
					Context: ".",
				},
			},
		},
		{
			Name:          "Can mount same source to multiple targets",
			ExpectedError: nil,
			ContainerRequest: ContainerRequest{
				Image:  "redis:latest",
				Mounts: Mounts(BindMount("/data", "/srv"), BindMount("/data", "/data")),
			},
		},
		{
			Name:          "Cannot mount multiple sources to same target",
			ExpectedError: errors.New("duplicate mount target detected: /data"),
			ContainerRequest: ContainerRequest{
				Image:  "redis:latest",
				Mounts: Mounts(BindMount("/srv", "/data"), BindMount("/data", "/data")),
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
		{
			name:                   "defaults to \"Dockerfile\" 1",
			ExpectedDockerfileName: "Dockerfile",
			ContainerRequest:       ContainerRequest{},
		},
		{
			name:                   "defaults to \"Dockerfile\" 2",
			ExpectedDockerfileName: "Dockerfile",
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{},
			},
		},
		{
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
		{
			Name: "test build from context archive",
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
			Name:               "test buildling from a context on the filesystem",
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
			ExpectedError: errors.New("you must specify either a build context or an image: failed to create container"),
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
				terminateContainerOnEnd(t, ctx, c)
			}
		})
	}
}

func Test_GetLogsFromFailedContainer(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:      "docker.io/alpine",
		Cmd:        []string{"echo", "-n", "I was not expecting this"},
		WaitingFor: wait.ForLog("I was expecting this").WithStartupTimeout(5 * time.Second),
	}

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
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

func TestShouldStartContainersInParallel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	t.Cleanup(cancel)

	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			t.Parallel()
			createTestContainer(t, ctx)
		})
	}
}

func createTestContainer(t *testing.T, ctx context.Context) int {
	req := ContainerRequest{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForHTTP("/"),
	}
	container, err := GenericContainer(ctx, GenericContainerRequest{
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

	return port.Int()
}

func TestBindMount(t *testing.T) {
	t.Parallel()
	type args struct {
		hostPath    string
		mountTarget ContainerMountTarget
	}
	tests := []struct {
		name string
		args args
		want ContainerMount
	}{
		{
			name: "/var/run/docker.sock:/var/run/docker.sock",
			args: args{hostPath: "/var/run/docker.sock", mountTarget: "/var/run/docker.sock"},
			want: ContainerMount{Source: GenericBindMountSource{HostPath: "/var/run/docker.sock"}, Target: "/var/run/docker.sock"},
		},
		{
			name: "/var/lib/app/data:/data",
			args: args{hostPath: "/var/lib/app/data", mountTarget: "/data"},
			want: ContainerMount{Source: GenericBindMountSource{HostPath: "/var/lib/app/data"}, Target: "/data"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, BindMount(tt.args.hostPath, tt.args.mountTarget), "BindMount(%v, %v)", tt.args.hostPath, tt.args.mountTarget)
		})
	}
}

func TestVolumeMount(t *testing.T) {
	t.Parallel()
	type args struct {
		volumeName  string
		mountTarget ContainerMountTarget
	}
	tests := []struct {
		name string
		args args
		want ContainerMount
	}{
		{
			name: "sample-data:/data",
			args: args{volumeName: "sample-data", mountTarget: "/data"},
			want: ContainerMount{Source: GenericVolumeMountSource{Name: "sample-data"}, Target: "/data"},
		},
		{
			name: "web:/var/nginx/html",
			args: args{volumeName: "web", mountTarget: "/var/nginx/html"},
			want: ContainerMount{Source: GenericVolumeMountSource{Name: "web"}, Target: "/var/nginx/html"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, VolumeMount(tt.args.volumeName, tt.args.mountTarget), "VolumeMount(%v, %v)", tt.args.volumeName, tt.args.mountTarget)
		})
	}
}

func TestOverrideContainerRequest(t *testing.T) {
	req := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Env: map[string]string{
				"BAR": "BAR",
			},
			Image:        "foo",
			ExposedPorts: []string{"12345/tcp"},
			WaitingFor: wait.ForNop(
				func(ctx context.Context, target wait.StrategyTarget) error {
					return nil
				},
			),
			Networks: []string{"foo", "bar", "baaz"},
			NetworkAliases: map[string][]string{
				"foo": {"foo0", "foo1", "foo2", "foo3"},
			},
		},
	}

	toBeMergedRequest := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Env: map[string]string{
				"FOO": "FOO",
			},
			Image:        "bar",
			ExposedPorts: []string{"67890/tcp"},
			Networks:     []string{"foo1", "bar1"},
			NetworkAliases: map[string][]string{
				"foo1": {"bar"},
			},
			WaitingFor: wait.ForLog("foo"),
		},
	}

	// the toBeMergedRequest should be merged into the req
	CustomizeRequest(toBeMergedRequest)(&req)

	// toBeMergedRequest should not be changed
	assert.Equal(t, "", toBeMergedRequest.Env["BAR"])
	assert.Equal(t, 1, len(toBeMergedRequest.ExposedPorts))
	assert.Equal(t, "67890/tcp", toBeMergedRequest.ExposedPorts[0])

	// req should be merged with toBeMergedRequest
	assert.Equal(t, "FOO", req.Env["FOO"])
	assert.Equal(t, "BAR", req.Env["BAR"])
	assert.Equal(t, "bar", req.Image)
	assert.Equal(t, []string{"12345/tcp", "67890/tcp"}, req.ExposedPorts)
	assert.Equal(t, []string{"foo", "bar", "baaz", "foo1", "bar1"}, req.Networks)
	assert.Equal(t, []string{"foo0", "foo1", "foo2", "foo3"}, req.NetworkAliases["foo"])
	assert.Equal(t, []string{"bar"}, req.NetworkAliases["foo1"])
	assert.Equal(t, wait.ForLog("foo"), req.WaitingFor)
}
