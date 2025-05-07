package dolt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	rootUser            = "root"
	defaultUser         = "test"
	defaultPassword     = "test"
	defaultDatabaseName = "test"
)

// DoltContainer represents the Dolt container type used in the module
type DoltContainer struct {
	testcontainers.Container
	username string
	password string
	database string
}

func WithDefaultCredentials() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		username := req.Env["DOLT_USER"]
		if strings.EqualFold(rootUser, username) {
			delete(req.Env, "DOLT_USER")
			delete(req.Env, "DOLT_PASSWORD")
		}

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Couchbase container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DoltContainer, error) {
	return Run(ctx, "dolthub/dolt-sql-server:1.32.4", opts...)
}

// Run creates an instance of the Dolt container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DoltContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"DOLT_USER":     defaultUser,
			"DOLT_PASSWORD": defaultPassword,
			"DOLT_DATABASE": defaultDatabaseName,
		},
		WaitingFor: wait.ForLog("Server ready. Accepting connections."),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	opts = append(opts, WithDefaultCredentials())

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	createUser := true
	username, ok := req.Env["DOLT_USER"]
	if !ok {
		username = rootUser
		createUser = false
	}
	password := req.Env["DOLT_PASSWORD"]

	database := req.Env["DOLT_DATABASE"]
	if database == "" {
		database = defaultDatabaseName
	}

	if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
		return nil, errors.New("empty password can be used only with the root user")
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var dc *DoltContainer
	if container != nil {
		dc = &DoltContainer{Container: container, username: username, password: password, database: database}
	}
	if err != nil {
		return dc, err
	}

	// dolthub/dolt-sql-server does not create user or database, so we do so here
	if err = dc.initialize(ctx, createUser); err != nil {
		return dc, fmt.Errorf("initialize: %w", err)
	}

	return dc, nil
}

func (c *DoltContainer) initialize(ctx context.Context, createUser bool) error {
	connectionString, err := c.initialConnectionString(ctx)
	if err != nil {
		return err
	}

	var db *sql.DB
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}
	defer func() {
		rerr := db.Close()
		if err == nil {
			err = rerr
		}
	}()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("error pinging db: %w", err)
	}

	// create database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", c.database))
	if err != nil {
		return fmt.Errorf("error creating database %s: %w", c.database, err)
	}

	if createUser {
		// create user
		_, err = db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s' IDENTIFIED BY '%s';", c.username, c.password))
		if err != nil {
			return fmt.Errorf("error creating user %s: %w", c.username, err)
		}

		q := fmt.Sprintf("GRANT ALL ON %s.* TO '%s';", c.database, c.username)
		// grant user privileges
		_, err = db.Exec(q)
		if err != nil {
			return fmt.Errorf("error creating user %s: %w", c.username, err)
		}
	}

	return nil
}

func (c *DoltContainer) initialConnectionString(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "3306/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("root:@tcp(%s:%s)/", host, containerPort.Port())
	return connectionString, nil
}

func (c *DoltContainer) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

func (c *DoltContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "3306/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = strings.Join(args, "&")
	}
	if extraArgs != "" {
		extraArgs = "?" + extraArgs
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", c.username, c.password, host, containerPort.Port(), c.database, extraArgs)
	return connectionString, nil
}

func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DOLT_USER"] = username
		return nil
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DOLT_PASSWORD"] = password
		return nil
	}
}

func WithDoltCredsPublicKey(key string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DOLT_CREDS_PUB_KEY"] = key
		return nil
	}
}

//nolint:revive,staticcheck //FIXME
func WithDoltCloneRemoteUrl(url string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DOLT_REMOTE_CLONE_URL"] = url
		return nil
	}
}

func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DOLT_DATABASE"] = database
		return nil
	}
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/dolt/servercfg.d/server.cnf",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)
		return nil
	}
}

func WithCredsFile(credsFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      credsFile,
			ContainerFilePath: "/root/.dolt/creds/" + filepath.Base(credsFile),
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)
		return nil
	}
}

func WithScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		var initScripts []testcontainers.ContainerFile
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0o755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)
		return nil
	}
}
