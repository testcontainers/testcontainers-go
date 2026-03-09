package forgejo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultHTTPPort = "3000/tcp"
	defaultSSHPort  = "22/tcp"
	defaultUser     = "forgejo-admin"
	defaultPassword = "forgejo-admin"
	defaultEmail    = "admin@forgejo.local"
)

// Container represents the Forgejo container type used in the module
type Container struct {
	testcontainers.Container
	adminUsername string
	adminPassword string
}

// AdminUsername returns the admin username for the Forgejo instance.
func (c *Container) AdminUsername() string {
	return c.adminUsername
}

// AdminPassword returns the admin password for the Forgejo instance.
func (c *Container) AdminPassword() string {
	return c.adminPassword
}

// extractAdminCredentials parses FORGEJO_ADMIN_* env vars from the container
// environment, falling back to the default values for any that are not set.
func extractAdminCredentials(env []string) (username, password, email string) {
	username, password, email = defaultUser, defaultPassword, defaultEmail
	for _, e := range env {
		if v, ok := strings.CutPrefix(e, "FORGEJO_ADMIN_USERNAME="); ok {
			username = v
		}
		if v, ok := strings.CutPrefix(e, "FORGEJO_ADMIN_PASSWORD="); ok {
			password = v
		}
		if v, ok := strings.CutPrefix(e, "FORGEJO_ADMIN_EMAIL="); ok {
			email = v
		}
	}
	return
}

// Run creates an instance of the Forgejo container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Closure variables populated by the PostReadies hook so we can avoid
	// a second container.Inspect call after Run returns.
	var adminUser, adminPass string

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 4+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultHTTPPort, defaultSSHPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/api/healthz").WithPort(defaultHTTPPort),
		),
		// Use SQLite for simplicity in tests (no external DB needed).
		// INSTALL_LOCK skips the install wizard so the instance is ready to use.
		testcontainers.WithEnv(map[string]string{
			"FORGEJO__database__DB_TYPE":      "sqlite3",
			"FORGEJO__security__INSTALL_LOCK": "true",
			"FORGEJO_ADMIN_USERNAME":          defaultUser,
			"FORGEJO_ADMIN_PASSWORD":          defaultPassword,
			"FORGEJO_ADMIN_EMAIL":             defaultEmail,
		}),
	)

	moduleOpts = append(moduleOpts, opts...)

	// Add lifecycle hook to create admin user after container is ready.
	// The hook reads credentials from container env vars so that user-provided
	// options (which override the defaults above) are respected.
	// The command runs as the "git" user because Forgejo refuses to run CLI
	// commands as root.
	adminHook := testcontainers.ContainerLifecycleHooks{
		PostReadies: []testcontainers.ContainerHook{
			func(ctx context.Context, container testcontainers.Container) error {
				inspect, err := container.Inspect(ctx)
				if err != nil {
					return fmt.Errorf("inspect forgejo: %w", err)
				}

				username, password, email := extractAdminCredentials(inspect.Config.Env)

				// Store credentials in closure for Run to use later.
				adminUser = username
				adminPass = password

				code, output, err := container.Exec(ctx, []string{
					"forgejo", "admin", "user", "create",
					"--username", username,
					"--password", password,
					"--email", email,
					"--admin",
					"--must-change-password=false",
				}, exec.WithUser("git"))
				if err != nil {
					return fmt.Errorf("create admin user: %w", err)
				}
				if code != 0 {
					data, _ := io.ReadAll(output)
					return fmt.Errorf("create admin user: exit code %d: %s", code, string(data))
				}
				return nil
			},
		},
	}

	moduleOpts = append(moduleOpts, testcontainers.WithAdditionalLifecycleHooks(adminHook))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run forgejo: %w", err)
	}

	// Credentials were populated by the PostReadies hook above.
	c.adminUsername = adminUser
	c.adminPassword = adminPass

	return c, nil
}

// ConnectionString returns the HTTP URL for the Forgejo instance
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultHTTPPort, "http")
}

// SSHConnectionString returns the SSH endpoint for Git operations
func (c *Container) SSHConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultSSHPort, "")
}

// WithAdminCredentials sets the admin username, password, and email for the Forgejo instance.
// These credentials are used to create an admin user after the container is ready.
func WithAdminCredentials(username, password, email string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if username == "" || password == "" || email == "" {
			return errors.New("WithAdminCredentials: username, password, and email must not be empty")
		}
		if req.Env == nil {
			req.Env = make(map[string]string)
		}
		req.Env["FORGEJO_ADMIN_USERNAME"] = username
		req.Env["FORGEJO_ADMIN_PASSWORD"] = password
		req.Env["FORGEJO_ADMIN_EMAIL"] = email
		return nil
	}
}

// WithConfig sets a Forgejo configuration value using the FORGEJO__section__key
// environment variable format.
// See https://forgejo.org/docs/latest/admin/config-cheat-sheet/ for available options.
func WithConfig(section, key, value string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if section == "" || key == "" {
			return fmt.Errorf("WithConfig: section and key must not be empty (got section=%q, key=%q)", section, key)
		}
		if strings.Contains(section, "__") || strings.Contains(key, "__") {
			return fmt.Errorf("WithConfig: section and key must not contain \"__\" (got section=%q, key=%q)", section, key)
		}
		if req.Env == nil {
			req.Env = make(map[string]string)
		}
		envKey := fmt.Sprintf("FORGEJO__%s__%s", section, key)
		req.Env[envKey] = value
		return nil
	}
}
