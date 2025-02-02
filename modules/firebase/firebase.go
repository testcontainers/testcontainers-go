package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"slices"

	"github.com/docker/docker/api/types/mount"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents the Firebase container type used in the module
type Container struct {
	testcontainers.Container
}

const rootFilePath = "/srv/firebase"

// WithRoot sets the directory which is copied to the destination container as firebase root
func WithRoot(rootPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      rootPath,
			ContainerFilePath: rootFilePath,
			FileMode:          0o775,
		})

		return nil
	}
}

// WithData names the data directory (by default under firebase root), can be used as a way of setting up fixtures.
// Usage of absolute path will imply that the user knows how to mount external directory into the container.
func WithData(dataPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DATA_DIRECTORY"] = dataPath
		return nil
	}
}

const cacheFilePath = "/root/.cache/firebase"

// WithCache enables firebase binary cache based on session (meaningful only when multiple tests are used)
func WithCache() testcontainers.CustomizeRequestOption {
	volumeName := fmt.Sprintf("firestore-cache-%s", testcontainers.SessionID())
	volumeOptions := &mount.VolumeOptions{
		Labels: testcontainers.GenericLabels(),
	}

	return func(req *testcontainers.GenericContainerRequest) error {
		m := testcontainers.ContainerMount{
			Source: testcontainers.DockerVolumeMountSource{
				Name:          volumeName,
				VolumeOptions: volumeOptions,
			},
			Target: cacheFilePath,
		}
		req.Mounts = append(req.Mounts, m)
		return nil
	}
}

func gatherPorts(config partialFirebaseConfig) ([]string, error) {
	var ports []string

	v := reflect.ValueOf(config.Emulators)
	for i := 0; i < v.NumField(); i++ {
		emulator := v.Field(i)
		if emulator.Kind() != reflect.Struct {
			continue
		}
		name := v.Type().Field(i).Name

		enabledF := emulator.FieldByName("Enabled")
		if enabledF != (reflect.Value{}) && !enabledF.Bool() {
			continue
		}

		hostF := emulator.FieldByName("Host")
		portF := emulator.FieldByName("Port")
		websocketPortF := emulator.FieldByName("WebsocketPort")

		if hostF != (reflect.Value{}) && !hostF.IsZero() {
			host := hostF.String()
			if host != "0.0.0.0" {
				return nil, fmt.Errorf("config specified %s emulator host on non public ip: %s", name, host)
			}
		}
		if portF != (reflect.Value{}) && !portF.IsZero() {
			port := fmt.Sprintf("%d/tcp", portF.Int())
			ports = append(ports, port)
		}
		if websocketPortF != (reflect.Value{}) && !websocketPortF.IsZero() {
			port := fmt.Sprintf("%d/tcp", websocketPortF.Int())
			ports = append(ports, port)
		}
	}

	return ports, nil
}

// Run creates an instance of the Firebase container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      img,
			Env:        map[string]string{},
			WaitingFor: wait.ForLog("All emulators ready! It is now safe to connect your app."),
		},
		Started: true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	// Check if user supplied root:
	rootPathIdx := slices.IndexFunc(req.ContainerRequest.Files, func(file testcontainers.ContainerFile) bool {
		return file.ContainerFilePath == rootFilePath
	})
	if rootPathIdx == -1 {
		return nil, fmt.Errorf("firebase root not provided (WithRoot is required)")
	}
	// Parse expected emulators from the root:
	userRoot := req.ContainerRequest.Files[rootPathIdx].HostFilePath
	cfg, err := os.Open(path.Join(userRoot, "firebase.json"))
	if err != nil {
		return nil, fmt.Errorf("open firebase.json: %w", err)
	}
	defer cfg.Close()
	bytes, err := io.ReadAll(cfg)
	if err != nil {
		return nil, fmt.Errorf("read firebase.json: %w", err)
	}
	var parsed partialFirebaseConfig
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		return nil, fmt.Errorf("parse firebase.json: %w", err)
	}
	expectedExposedPorts, err := gatherPorts(parsed)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	req.ExposedPorts = expectedExposedPorts

	container, err := testcontainers.GenericContainer(ctx, req)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
