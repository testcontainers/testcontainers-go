package dolt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// Deprecated: use Run instead
// RunContainer creates an instance of the Couchbase container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DoltContainer, error) {
	return Run(ctx, "dolthub/dolt-sql-server:1.32.4", opts...)
}

// Run creates an instance of the Dolt container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DoltContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("3306/tcp", "33060/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("Server ready. Accepting connections.")),
	}

	opts = append(opts, WithDefaultCredentials())

	moduleOpts = append(moduleOpts, opts...)

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("dolt option: %w", err)
			}
		}
	}

	// module options take precedence over default options
	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultOptions.env), testcontainers.WithFiles(defaultOptions.files...))

	createUser := true
	username, ok := defaultOptions.env["DOLT_USER"]
	if !ok {
		username = rootUser
		createUser = false
	}
	password := defaultOptions.env["DOLT_PASSWORD"]

	database := defaultOptions.env["DOLT_DATABASE"]
	if database == "" {
		database = defaultDatabaseName
	}

	if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
		return nil, errors.New("empty password can be used only with the root user")
	}

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var dc *DoltContainer
	if container != nil {
		dc = &DoltContainer{Container: container, username: username, password: password, database: database}
	}
	if err != nil {
		return dc, fmt.Errorf("run: %w", err)
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
	endpoint, err := c.PortEndpoint(ctx, "3306/tcp", "")
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("root:@tcp(%s)/", endpoint)
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
	endpoint, err := c.PortEndpoint(ctx, "3306/tcp", "")
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

	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s%s", c.username, c.password, endpoint, c.database, extraArgs)
	return connectionString, nil
}
