package atlaslocal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	passwordContainerPath = "/run/secrets/mongo-root-password"
	usernameContainerPath = "/run/secrets/mongo-root-username"
)

// Container represents the MongoDBAtlasLocal container type used in the module
type Container struct {
	testcontainers.Container
	username      string
	password      string
	mongotLogPath string
	runnerLogPath string
}

// Run creates an instance of the MongoDBAtlasLocal container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("27017/tcp"),
			wait.ForHealthCheck(),
		),
		Env: map[string]string{},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("failed to create customized request: %w", err)
		}
	}

	if err := validateRequest(&genericContainerReq); err != nil {
		return nil, fmt.Errorf("incompatible configuration: %w", err)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)

	var c *Container
	if container != nil {
		// No need to check the error here, as we already validated the request.
		username, _ := parseUsername(genericContainerReq.Env)
		password, _ := parsePassword(genericContainerReq.Env)

		c = &Container{
			Container:     container,
			username:      username,
			password:      password,
			mongotLogPath: genericContainerReq.Env["MONGOT_LOG_FILE"],
			runnerLogPath: genericContainerReq.Env["RUNNER_LOG_FILE"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("create container: %w", err)
	}

	return c, nil
}

func validateRequest(req *testcontainers.GenericContainerRequest) error {
	username := getRootUsername(req.Env)
	password := getRootPassword(req.Env)

	// If username or password is specified, both must be provided.
	if username != "" && password == "" || username == "" && password != "" {
		return errors.New("if you specify username or password, you must provide both of them")
	}

	usernameFile := getRootUsernameFile(req.Env)
	passwordFile := getRootPasswordFile(req.Env)

	// If username file or password file is specified, both must be provided.
	if usernameFile != "" && passwordFile == "" || usernameFile == "" && passwordFile != "" {
		return errors.New("if you specify username file or password file, you must provide both of them")
	}

	// Setting credentials both inline and using files will result in an panic
	// from the container, so we short circuit here.
	if (username != "" || password != "") && (usernameFile != "" || passwordFile != "") {
		return errors.New("you cannot specify both inline credentials and files for credentials")
	}

	return nil
}

// ConnectionString returns the connection string for the MongoDB Atlas Local
// container. If you provide a username and a password, the connection string
// will also include them.
func (ctr *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := ctr.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := ctr.MappedPort(ctx, "27017")
	if err != nil {
		return "", err
	}

	uri := &url.URL{
		Scheme:   "mongodb",
		Host:     net.JoinHostPort(host, mappedPort.Port()),
		Path:     "/",
		RawQuery: "directConnection=true",
	}

	if ctr.username != "" && ctr.password != "" {
		uri.User = url.UserPassword(ctr.username, ctr.password)
	}

	return uri.String(), nil
}

// ReadMongotLogs returns a reader for mongot logs in the container. Reads from
// stdout/stderr or the configured log file.
//
// This method return the os.ErrNotExist sentinel error if it is called with
// no log file configured.
func (ctr *Container) ReadMongotLogs(ctx context.Context) (io.ReadCloser, error) {
	path := ctr.mongotLogPath
	if path == "" {
		return nil, os.ErrNotExist
	}

	switch ctr.mongotLogPath {
	case "/dev/stdout", "/dev/stderr":
		return ctr.Logs(ctx)
	default:
		return ctr.CopyFileFromContainer(ctx, ctr.mongotLogPath)
	}
}

func (ctr *Container) ReadRunnerLogs(ctx context.Context) (io.ReadCloser, error) {
	path := ctr.runnerLogPath
	if path == "" {
		return nil, os.ErrNotExist
	}

	switch ctr.runnerLogPath {
	case "/dev/stdout", "/dev/stderr":
		return ctr.Logs(ctx)
	default:
		return ctr.CopyFileFromContainer(ctx, ctr.runnerLogPath)
	}
}

// WithUsername sets the MongoDB root username by setting the
// MONGODB_INITDB_ROOT_USERNAME environment variable.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if username != "" {
			req.Env["MONGODB_INITDB_ROOT_USERNAME"] = username
		}

		return nil
	}
}

// WithPassword sets the MongoDB root password by setting the the
// MONGODB_INITDB_ROOT_PASSWORD environment variable.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if password != "" {
			req.Env["MONGODB_INITDB_ROOT_PASSWORD"] = password
		}

		return nil
	}
}

// WithUsernameFile mounts a local file as the MongoDB root username secret at
// /run/secrets/mongo-root-username and sets MONGODB_INITDB_ROOT_USERNAME_FILE.
// The path must be absolute and exist; no-op if empty.
func WithUsernameFile(usernameFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if usernameFile == "" {
			return nil
		}

		// Must be an absolute path.
		if !filepath.IsAbs(usernameFile) {
			return fmt.Errorf("username file mount path must be absolute, got: %s", usernameFile)
		}

		// Must exist and be a file.
		info, err := os.Stat(usernameFile)
		if err != nil {
			return fmt.Errorf("username file does not exist or is not accessible: %w", err)
		}

		if info.IsDir() {
			return fmt.Errorf("username file must be a file, got a directory: %s", usernameFile)
		}

		req.Env["MONGODB_INITDB_ROOT_USERNAME_FILE"] = usernameContainerPath

		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      usernameFile,
			ContainerFilePath: usernameContainerPath,
			FileMode:          0444,
		})

		if b, err := os.ReadFile(usernameFile); err == nil {
			req.Env["TC_ATLAS_LOCAL_MONGODB_INITDB_ROOT_USERNAME"] = strings.TrimSpace(string(b))
		} else {
			return fmt.Errorf("read username file: %w", err)
		}

		return nil
	}
}

// WithPasswordFile mounts a local file as the MongoDB root password secret at
// /run/secrets/mongo-root-password and sets MONGODB_INITDB_ROOT_PASSWORD_FILE.
// Path must be absolute and an existing file; no-op if empty.
func WithPasswordFile(passwordFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if passwordFile == "" {
			return nil
		}

		// Must be an absolute path.
		if !filepath.IsAbs(passwordFile) {
			return fmt.Errorf("password file mount path must be absolute, got: %s", passwordFile)
		}

		// Must exist and be a file.
		info, err := os.Stat(passwordFile)
		if err != nil {
			return fmt.Errorf("password file does not exist or is not accessible: %w", err)
		}

		if info.IsDir() {
			return fmt.Errorf("password file must be a file, got a directory: %s", passwordFile)
		}

		req.Env["MONGODB_INITDB_ROOT_PASSWORD_FILE"] = passwordContainerPath

		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      passwordFile,
			ContainerFilePath: passwordContainerPath,
			FileMode:          0444,
		})

		if b, err := os.ReadFile(passwordFile); err == nil {
			req.Env["TC_ATLAS_LOCAL_MONGODB_INITDB_ROOT_PASSWORD"] = strings.TrimSpace(string(b))
		} else {
			return fmt.Errorf("read password file: %w", err)
		}

		return nil
	}
}

// WithNoTelemetry opts out of telemetry for the MongoDB Atlas Local
// container by setting the DO_NOT_TRACK environment variable to 1.
func WithNoTelemetry() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DO_NOT_TRACK"] = "1"

		return nil
	}
}

// WithInitDatabase sets MONGODB_INITDB_DATABASE environment variable so the
// init scripts and the default connection string target the specified database
// instead of the default "test" database.
func WithInitDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGODB_INITDB_DATABASE"] = database

		return nil
	}
}

// WithInitScripts mounts a directory containing .sh/.js init scripts into
// /docker-entrypoint-initdb.d so they run in alphabetical order on startup. If
// called multiple times, this funcion removes any prior init-scripts bind and
// uses only the latest on specified.
func WithInitScripts(scriptsDir string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if scriptsDir == "" {
			return nil
		}

		abs, err := filepath.Abs(scriptsDir)
		if err != nil {
			return fmt.Errorf("get absolute path of init scripts dir: %w", err)
		}

		st, err := os.Stat(abs)
		if err != nil {
			return fmt.Errorf("stat init scripts dir: %w", err)
		}

		if !st.IsDir() {
			return fmt.Errorf("init scripts path is not a directory: %s", abs)
		}

		const dstDir = "/docker-entrypoint-initdb.d/"

		filtered := req.Files[:0]
		for _, file := range req.Files {
			if !strings.HasPrefix(file.ContainerFilePath, dstDir) {
				filtered = append(filtered, file)
			}
		}

		req.Files = filtered

		entries, err := os.ReadDir(abs)
		if err != nil {
			return fmt.Errorf("read init scripts dir: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !(strings.HasSuffix(name, ".sh") || strings.HasSuffix(name, ".js")) {
				continue
			}

			f := testcontainers.ContainerFile{
				HostFilePath:      filepath.Join(abs, name),
				ContainerFilePath: filepath.Join(dstDir, name),
				FileMode:          0644,
			}

			if strings.HasSuffix(name, ".sh") {
				f.FileMode = 0755 // Make shell scripts executable.
			}

			req.Files = append(req.Files, f)
		}

		return nil
	}
}

// WithMongotLogStdout writes to /dev/stdout inside the container. See
// (*Container).ReadMongotLogs to read the logs locally.
func WithMongotLogToStdout() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGOT_LOG_FILE"] = "/dev/stdout"

		return nil
	}
}

// WithMongotLogToStderr writes to /dev/stderr inside the container. See
// (*Container).ReadMongotLogs to read the logs locally.
func WithMongotLogToStderr() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGOT_LOG_FILE"] = "/dev/stderr"

		return nil
	}
}

// WithMongotLogFile writes the mongot logs to /tmp/mongot.log inside the
// container. See (*Container).ReadMongotLogs to read the logs locally.
func WithMongotLogFile() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGOT_LOG_FILE"] = "/tmp/mongot.log"

		return nil
	}
}

// WithRunnerLogToStdout writes to /dev/stdout inside the container. See
// (*Container).ReadRunnerLogs to read the logs locally.
func WithRunnerLogToStdout() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["RUNNER_LOG_FILE"] = "/dev/stdout"

		return nil
	}
}

// WithRunnerLogToStderr writes to /dev/stderr inside the container. See
// (*Container).ReadRunnerLogs to read the logs locally.
func WithRunnerLogToStderr() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["RUNNER_LOG_FILE"] = "/dev/stderr"

		return nil
	}
}

// WithRunnerLogFile writes the runner logs to /tmp/runner.log inside the
// container. See (*Container).ReadRunnerLogs to read the logs locally.
func WithRunnerLogFile() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["RUNNER_LOG_FILE"] = "/tmp/runner.log"

		return nil
	}
}

func getRootPassword(env map[string]string) string {
	return env["MONGODB_INITDB_ROOT_PASSWORD"]
}

func getRootUsername(env map[string]string) string {
	return env["MONGODB_INITDB_ROOT_USERNAME"]
}

func getRootUsernameFile(env map[string]string) string {
	return env["MONGODB_INITDB_ROOT_USERNAME_FILE"]
}

func getRootPasswordFile(env map[string]string) string {
	return env["MONGODB_INITDB_ROOT_PASSWORD_FILE"]
}

// parseUsername will try to parse the username from the environment by either
// reading the MONGODB_INITDB_ROOT_USERNAME environment variable or from the
// the file specified in the MONGODB_INITDB_ROOT_USERNAME_FILE environment
// variable.
func parseUsername(env map[string]string) (string, error) {
	if username := env["TC_ATLAS_LOCAL_MONGODB_INITDB_ROOT_USERNAME"]; username != "" {
		return username, nil
	}

	if username := getRootUsername(env); username != "" {
		return username, nil
	}

	if usernameFile := getRootUsernameFile(env); usernameFile != "" {
		r, err := os.ReadFile(usernameFile)
		return strings.TrimSpace(string(r)), err
	}

	return "", nil
}

// parsePassword will try to parse the password from the environment by either
// reading the MONGODB_INITDB_ROOT_PASSWORD environment variable or from
// the file specified in the MONGODB_INITDB_ROOT_PASSWORD_FILE environment
// variable.
func parsePassword(env map[string]string) (string, error) {
	if password := env["TC_ATLAS_LOCAL_MONGODB_INITDB_ROOT_PASSWORD"]; password != "" {
		return password, nil
	}

	if password := getRootPassword(env); password != "" {
		return password, nil
	}

	if passwordFile := getRootPasswordFile(env); passwordFile != "" {
		r, err := os.ReadFile(passwordFile)
		return strings.TrimSpace(string(r)), err
	}

	return "", nil
}
