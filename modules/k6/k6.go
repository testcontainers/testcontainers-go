package k6

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/mount"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// cacheTarget is the path to the cache volume in the container.
const cacheTarget = "/cache"

// K6Container represents the K6 container type used in the module
type K6Container struct {
	testcontainers.Container
}

type DownloadableFile struct {
	Uri         url.URL
	DownloadDir string
	User        string
	Password    string
}

func (d *DownloadableFile) getDownloadPath() string {
	baseName := path.Base(d.Uri.Path)
	return path.Join(d.DownloadDir, baseName)
}

func downloadFileFromDescription(d DownloadableFile) error {
	client := http.Client{Timeout: time.Second * 60}
	req, err := http.NewRequest(http.MethodGet, d.Uri.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/javascript")
	// Set up HTTPS request with basic authorization.
	if d.User != "" && d.Password != "" {
		req.SetBasicAuth(d.User, d.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	downloadedFile, err := os.Create(d.getDownloadPath())
	if err != nil {
		return err
	}
	defer downloadedFile.Close()

	_, err = io.Copy(downloadedFile, resp.Body)
	return err
}

// WithTestScript mounts the given script into the ./test directory in the container
// and passes it to k6 as the test to run.
// The path to the script must be an absolute path
func WithTestScript(scriptPath string) testcontainers.CustomizeRequestOption {
	scriptBaseName := filepath.Base(scriptPath)
	f, err := os.Open(scriptPath)
	if err != nil {
		return func(req *testcontainers.GenericContainerRequest) error {
			return fmt.Errorf("cannot create reader for test file: %w", err)
		}
	}

	return WithTestScriptReader(f, scriptBaseName)
}

// WithTestScriptReader copies files into the Container using the Reader API
// The script base name is not a path, neither absolute or relative and should
// be just the file name of the script
func WithTestScriptReader(reader io.Reader, scriptBaseName string) testcontainers.CustomizeRequestOption {
	opt := func(req *testcontainers.GenericContainerRequest) error {
		target := "/home/k6x/" + scriptBaseName
		req.Files = append(
			req.Files,
			testcontainers.ContainerFile{
				Reader:            reader,
				ContainerFilePath: target,
				FileMode:          0o644,
			},
		)

		// add script to the k6 run command
		req.Cmd = append(req.Cmd, target)

		return nil
	}
	return opt
}

// WithRemoteTestScript takes a RemoteTestFileDescription and copies to container
func WithRemoteTestScript(d DownloadableFile) testcontainers.CustomizeRequestOption {
	err := downloadFileFromDescription(d)
	if err != nil {
		return func(req *testcontainers.GenericContainerRequest) error {
			return fmt.Errorf("not able to download required test script: %w", err)
		}
	}

	return WithTestScript(d.getDownloadPath())
}

// WithCmdOptions pass the given options to the k6 run command
func WithCmdOptions(options ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = append(req.Cmd, options...)

		return nil
	}
}

// SetEnvVar adds a '--env' command-line flag to the k6 command in the container for setting an environment variable for the test script.
func SetEnvVar(variable string, value string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = append(req.Cmd, "--env", fmt.Sprintf("%s=%s", variable, value))

		return nil
	}
}

// WithCache sets a volume as a cache for building the k6 binary
// If a volume name is provided in the TC_K6_BUILD_CACHE, this volume is used and it will
// persist across test sessions.
// If no value is provided, a volume is created and automatically deleted when the test session ends.
func WithCache() testcontainers.CustomizeRequestOption {
	var volOptions *mount.VolumeOptions

	cacheVol := os.Getenv("TC_K6_BUILD_CACHE")
	// if no volume is provided, create one and ensure add labels for garbage collection
	if cacheVol == "" {
		cacheVol = fmt.Sprintf("k6-cache-%s", testcontainers.SessionID())
		volOptions = &mount.VolumeOptions{
			Labels: testcontainers.GenericLabels(),
		}
	}

	return func(req *testcontainers.GenericContainerRequest) error {
		mount := testcontainers.ContainerMount{
			Source: testcontainers.DockerVolumeMountSource{
				Name:          cacheVol,
				VolumeOptions: volOptions,
			},
			Target: cacheTarget,
		}
		req.Mounts = append(req.Mounts, mount)

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the K6 container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*K6Container, error) {
	return Run(ctx, "szkiba/k6x:v0.3.1", opts...)
}

// Run creates an instance of the K6 container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*K6Container, error) {
	req := testcontainers.ContainerRequest{
		Image:      img,
		Cmd:        []string{"run"},
		WaitingFor: wait.ForExit(),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *K6Container
	if container != nil {
		c = &K6Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// CacheMount returns the name of volume used as a cache or an empty string
// if no cache was found.
func (k *K6Container) CacheMount(ctx context.Context) (string, error) {
	inspect, err := k.Inspect(ctx)
	if err != nil {
		return "", fmt.Errorf("inspect: %w", err)
	}

	for _, m := range inspect.Mounts {
		if m.Type == mount.TypeVolume && m.Destination == cacheTarget {
			return m.Name, nil
		}
	}

	return "", nil
}
