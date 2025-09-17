package atlaslocal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

const (
	passwordContainerPath      = "/run/secrets/mongo-root-password"
	usernameContainerPath      = "/run/secrets/mongo-root-username"
	envMongotLogFile           = "MONGOT_LOG_FILE"
	envRunnerLogFile           = "RUNNER_LOG_FILE"
	envMongoDBInitDatabase     = "MONGODB_INITDB_DATABASE"
	envMongoDBInitUsername     = "MONGODB_INITDB_ROOT_USERNAME"
	envMongoDBInitPassword     = "MONGODB_INITDB_ROOT_PASSWORD"
	envMongoDBInitUsernameFile = "MONGODB_INITDB_ROOT_USERNAME_FILE"
	envMongoDBInitPasswordFile = "MONGODB_INITDB_ROOT_PASSWORD_FILE"
	envDoNotTrack              = "DO_NOT_TRACK"
)

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

type options struct {
	username          string
	password          string
	localUsernamefile string
	localPasswordFile string
	noTelemetry       bool
	database          string
	mongotLogPath     string
	runnerLogPath     string

	files []testcontainers.ContainerFile
}

func (opts options) env() map[string]string {
	env := map[string]string{}

	if opts.username != "" {
		env[envMongoDBInitUsername] = opts.username
	}

	if opts.password != "" {
		env[envMongoDBInitPassword] = opts.password
	}

	if opts.localUsernamefile != "" {
		env[envMongoDBInitUsernameFile] = usernameContainerPath
	}

	if opts.localPasswordFile != "" {
		env[envMongoDBInitPasswordFile] = passwordContainerPath
	}

	if opts.noTelemetry {
		env[envDoNotTrack] = "1"
	}

	if opts.database != "" {
		env[envMongoDBInitDatabase] = opts.database
	}

	if opts.mongotLogPath != "" {
		env[envMongotLogFile] = opts.mongotLogPath
	}

	if opts.runnerLogPath != "" {
		env[envRunnerLogFile] = opts.runnerLogPath
	}

	return env
}

func (opts options) validate() error {
	username := opts.username
	password := opts.password

	// If username or password is specified, both must be provided.
	if username != "" && password == "" || username == "" && password != "" {
		return errors.New("if you specify username or password, you must provide both of them")
	}

	usernameFile := opts.localUsernamefile
	passwordFile := opts.localPasswordFile

	// If username file or password file is specified, both must be provided.
	if usernameFile != "" && passwordFile == "" || usernameFile == "" && passwordFile != "" {
		return errors.New("if you specify username file or password file, you must provide both of them")
	}

	// Setting credentials both inline and using files will result in a panic
	// from the container, so we short circuit here.
	if (username != "" || password != "") && (usernameFile != "" || passwordFile != "") {
		return errors.New("you cannot specify both inline credentials and files for credentials")
	}

	return nil
}

// parseUsername will return either the username provided by WithUsername or
// from the local file specified by WithUsernameFile. If both are provided, this
// function will return an error. If neither is provided, an empty string is
// returned.
func (opts options) parseUsername() (string, error) {
	if opts.username == "" && opts.localUsernamefile == "" {
		return "", nil
	}

	if opts.username != "" && opts.localUsernamefile != "" {
		return "", errors.New("cannot specify both inline credentials and files for credentials")
	}

	if opts.username != "" {
		return opts.username, nil
	}

	r, err := os.ReadFile(opts.localUsernamefile)
	return strings.TrimSpace(string(r)), err
}

// parsePassword will return either the password provided by WithPassword or
// from the local file specified by WithPasswordFile. If both are provided, this
// function will return an error. If neither is provided, an empty string is
// returned.
func (opts options) parsePassword() (string, error) {
	if opts.password == "" && opts.localPasswordFile == "" {
		return "", nil
	}

	if opts.password != "" && opts.localPasswordFile != "" {
		return "", errors.New("cannot specify both inline credentials and files for credentials")
	}

	if opts.password != "" {
		return opts.password, nil
	}

	r, err := os.ReadFile(opts.localPasswordFile)
	return strings.TrimSpace(string(r)), err
}

// Option is an option for the Redpanda container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithUsername sets the MongoDB root username by setting the
// MONGODB_INITDB_ROOT_USERNAME environment variable.
func WithUsername(username string) Option {
	return func(opts *options) error {
		if username != "" {
			opts.username = username
		}

		return nil
	}
}

// WithPassword sets the MongoDB root password by setting the the
// MONGODB_INITDB_ROOT_PASSWORD environment variable.
func WithPassword(password string) Option {
	return func(opts *options) error {
		if password != "" {
			opts.password = password
		}

		return nil
	}
}

// WithUsernameFile mounts a local file as the MongoDB root username secret at
// /run/secrets/mongo-root-username and sets MONGODB_INITDB_ROOT_USERNAME_FILE.
// The path must be absolute and exist; no-op if empty.
func WithUsernameFile(usernameFile string) Option {
	return func(opts *options) error {
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

		opts.localUsernamefile = usernameFile

		opts.files = append(opts.files, testcontainers.ContainerFile{
			HostFilePath:      usernameFile,
			ContainerFilePath: usernameContainerPath,
			FileMode:          0o444,
		})

		return nil
	}
}

// WithPasswordFile mounts a local file as the MongoDB root password secret at
// /run/secrets/mongo-root-password and sets MONGODB_INITDB_ROOT_PASSWORD_FILE.
// Path must be absolute and an existing file; no-op if empty.
func WithPasswordFile(passwordFile string) Option {
	return func(opts *options) error {
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

		opts.localPasswordFile = passwordFile

		opts.files = append(opts.files, testcontainers.ContainerFile{
			HostFilePath:      passwordFile,
			ContainerFilePath: passwordContainerPath,
			FileMode:          0o444,
		})

		return nil
	}
}

// WithNoTelemetry opts out of telemetry for the MongoDB Atlas Local
// container by setting the DO_NOT_TRACK environment variable to 1.
func WithNoTelemetry() Option {
	return func(opts *options) error {
		opts.noTelemetry = true

		return nil
	}
}

// WithInitDatabase sets MONGODB_INITDB_DATABASE environment variable so the
// init scripts and the default connection string target the specified database
// instead of the default "test" database.
func WithInitDatabase(database string) Option {
	return func(opts *options) error {
		opts.database = database

		return nil
	}
}

// WithInitScripts mounts a directory containing .sh/.js init scripts into
// /docker-entrypoint-initdb.d so they run in alphabetical order on startup. If
// called multiple times, this function removes any prior init-scripts bind and
// uses only the latest on specified.
func WithInitScripts(scriptsDir string) Option {
	return func(opts *options) error {
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

		filtered := opts.files[:0]
		for _, file := range opts.files {
			if !strings.HasPrefix(file.ContainerFilePath, dstDir) {
				filtered = append(filtered, file)
			}
		}

		opts.files = filtered

		entries, err := os.ReadDir(abs)
		if err != nil {
			return fmt.Errorf("read init scripts dir: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasSuffix(name, ".sh") && !strings.HasSuffix(name, ".js") {
				continue
			}

			f := testcontainers.ContainerFile{
				HostFilePath:      filepath.Join(abs, name),
				ContainerFilePath: filepath.Join(dstDir, name),
				FileMode:          0o644,
			}

			if strings.HasSuffix(name, ".sh") {
				f.FileMode = 0o755 // Make shell scripts executable.
			}

			opts.files = append(opts.files, f)
		}

		return nil
	}
}

// WithMongotLogStdout writes to /dev/stdout inside the container. See
// (*Container).ReadMongotLogs to read the logs locally.
func WithMongotLogToStdout() Option {
	return func(opts *options) error {
		opts.mongotLogPath = "/dev/stdout"

		return nil
	}
}

// WithMongotLogToStderr writes to /dev/stderr inside the container. See
// (*Container).ReadMongotLogs to read the logs locally.
func WithMongotLogToStderr() Option {
	return func(opts *options) error {
		opts.mongotLogPath = "/dev/stderr"

		return nil
	}
}

// WithMongotLogFile writes the mongot logs to /tmp/mongot.log inside the
// container. See (*Container).ReadMongotLogs to read the logs locally.
func WithMongotLogFile() Option {
	return func(opts *options) error {
		opts.mongotLogPath = "/tmp/mongot.log"

		return nil
	}
}

// WithRunnerLogToStdout writes to /dev/stdout inside the container. See
// (*Container).ReadRunnerLogs to read the logs locally.
func WithRunnerLogToStdout() Option {
	return func(opts *options) error {
		opts.runnerLogPath = "/dev/stdout"

		return nil
	}
}

// WithRunnerLogToStderr writes to /dev/stderr inside the container. See
// (*Container).ReadRunnerLogs to read the logs locally.
func WithRunnerLogToStderr() Option {
	return func(opts *options) error {
		opts.runnerLogPath = "/dev/stderr"

		return nil
	}
}

// WithRunnerLogFile writes the runner logs to /tmp/runner.log inside the
// container. See (*Container).ReadRunnerLogs to read the logs locally.
func WithRunnerLogFile() Option {
	return func(opts *options) error {
		opts.runnerLogPath = "/tmp/runner.log"

		return nil
	}
}
