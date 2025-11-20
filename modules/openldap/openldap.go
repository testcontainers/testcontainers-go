package openldap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultUser     = "admin"
	defaultPassword = "adminpassword"
	defaultRoot     = "dc=example,dc=org"
	defaultAdminDn  = "cn=admin,dc=example,dc=org"
)

// OpenLDAPContainer represents the OpenLDAP container type used in the module
type OpenLDAPContainer struct {
	testcontainers.Container
	adminUsername string
	adminPassword string
	rootDn        string
}

// ConnectionString returns the connection string for the OpenLDAP container
func (c *OpenLDAPContainer) ConnectionString(ctx context.Context, _ ...string) (string, error) {
	return c.PortEndpoint(ctx, "1389/tcp", "ldap")
}

// LoadLdif loads an ldif file into the OpenLDAP container
func (c *OpenLDAPContainer) LoadLdif(ctx context.Context, ldif []byte) error {
	err := c.CopyToContainer(ctx, ldif, "/tmp/ldif.ldif", 0o644)
	if err != nil {
		return err
	}
	code, output, err := c.Exec(ctx, []string{"ldapadd", "-H", "ldap://localhost:1389", "-x", "-D", fmt.Sprintf("cn=%s,%s", c.adminUsername, c.rootDn), "-w", c.adminPassword, "-f", "/tmp/ldif.ldif"})
	if err != nil {
		return err
	}
	if code != 0 {
		data, _ := io.ReadAll(output)
		return errors.New(string(data))
	}
	return nil
}

// WithAdminUsername sets the initial admin username to be created when the container starts
// It is used in conjunction with WithAdminPassword to set a username and its password.
// It will create the specified user with admin power.
func WithAdminUsername(username string) testcontainers.ContainerCustomizer {
	return testcontainers.WithEnv(map[string]string{"LDAP_ADMIN_USERNAME": username})
}

// WithAdminPassword sets the initial admin password of the user to be created when the container starts
// It is used in conjunction with WithAdminUsername to set a username and its password.
// It will set the admin password for OpenLDAP.
func WithAdminPassword(password string) testcontainers.ContainerCustomizer {
	return testcontainers.WithEnv(map[string]string{"LDAP_ADMIN_PASSWORD": password})
}

// WithRoot sets the root of the OpenLDAP instance
func WithRoot(root string) testcontainers.ContainerCustomizer {
	return testcontainers.WithEnv(map[string]string{"LDAP_ROOT": root})
}

// WithInitialLdif sets the initial ldif file to be loaded into the OpenLDAP container
func WithInitialLdif(ldif string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      ldif,
			ContainerFilePath: "/initial_ldif.ldif",
			FileMode:          0o644,
		}

		hook := testcontainers.ContainerLifecycleHooks{
			PostReadies: []testcontainers.ContainerHook{
				func(ctx context.Context, container testcontainers.Container) error {
					username := req.Env["LDAP_ADMIN_USERNAME"]
					rootDn := req.Env["LDAP_ROOT"]
					password := req.Env["LDAP_ADMIN_PASSWORD"]
					code, output, err := container.Exec(ctx, []string{"ldapadd", "-H", "ldap://localhost:1389", "-x", "-D", fmt.Sprintf("cn=%s,%s", username, rootDn), "-w", password, "-f", "/initial_ldif.ldif"})
					if err != nil {
						return err
					}
					if code != 0 {
						data, _ := io.ReadAll(output)
						return errors.New(string(data))
					}
					return nil
				},
			},
		}

		if err := testcontainers.WithFiles(cf)(req); err != nil {
			return err
		}

		return testcontainers.WithAdditionalLifecycleHooks(hook)(req)
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the OpenLDAP container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error) {
	return Run(ctx, "bitnamilegacy/openldap:2.6.6", opts...)
}

// Run creates an instance of the OpenLDAP container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithEnv(map[string]string{
			"LDAP_ADMIN_USERNAME": defaultUser,
			"LDAP_ADMIN_PASSWORD": defaultPassword,
			"LDAP_ROOT":           defaultRoot,
		}),
		testcontainers.WithExposedPorts("1389/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("** Starting slapd **"),
			wait.ForListeningPort("1389/tcp"),
		),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *OpenLDAPContainer
	if ctr != nil {
		c = &OpenLDAPContainer{
			Container:     ctr,
			adminUsername: defaultUser,
			adminPassword: defaultPassword,
			rootDn:        defaultRoot,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run openldap: %w", err)
	}

	// Retrieve the actual env vars set on the container
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect openldap: %w", err)
	}

	var foundUser, foundPass, foundRoot bool
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "LDAP_ADMIN_USERNAME="); ok {
			c.adminUsername, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, "LDAP_ADMIN_PASSWORD="); ok {
			c.adminPassword, foundPass = v, true
		}
		if v, ok := strings.CutPrefix(env, "LDAP_ROOT="); ok {
			c.rootDn, foundRoot = v, true
		}

		if foundUser && foundPass && foundRoot {
			break
		}
	}

	return c, nil
}
