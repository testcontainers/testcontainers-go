package openldap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

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
func (c *OpenLDAPContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "1389/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := fmt.Sprintf("ldap://%s", net.JoinHostPort(host, containerPort.Port()))
	return connStr, nil
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
func WithAdminUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["LDAP_ADMIN_USERNAME"] = username

		return nil
	}
}

// WithAdminPassword sets the initial admin password of the user to be created when the container starts
// It is used in conjunction with WithAdminUsername to set a username and its password.
// It will set the admin password for OpenLDAP.
func WithAdminPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["LDAP_ADMIN_PASSWORD"] = password

		return nil
	}
}

// WithRoot sets the root of the OpenLDAP instance
func WithRoot(root string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["LDAP_ROOT"] = root

		return nil
	}
}

// WithInitialLdif sets the initial ldif file to be loaded into the OpenLDAP container
func WithInitialLdif(ldif string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      ldif,
			ContainerFilePath: "/initial_ldif.ldif",
			FileMode:          0o644,
		})

		req.LifecycleHooks = append(req.LifecycleHooks, testcontainers.ContainerLifecycleHooks{
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
		})

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the OpenLDAP container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error) {
	return Run(ctx, "bitnami/openldap:2.6.6", opts...)
}

// Run creates an instance of the OpenLDAP container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": defaultUser,
			"LDAP_ADMIN_PASSWORD": defaultPassword,
			"LDAP_ROOT":           defaultRoot,
		},
		ExposedPorts: []string{"1389/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("** Starting slapd **"),
			wait.ForListeningPort("1389/tcp"),
		),
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			{
				PostReadies: []testcontainers.ContainerHook{},
			},
		},
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
	var c *OpenLDAPContainer
	if container != nil {
		c = &OpenLDAPContainer{
			Container:     container,
			adminUsername: req.Env["LDAP_ADMIN_USERNAME"],
			adminPassword: req.Env["LDAP_ADMIN_PASSWORD"],
			rootDn:        req.Env["LDAP_ROOT"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
