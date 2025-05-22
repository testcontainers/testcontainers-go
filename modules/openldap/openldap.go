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
func (c *OpenLDAPContainer) ConnectionString(ctx context.Context, _ ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "1389/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := "ldap://" + net.JoinHostPort(host, containerPort.Port())
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

// Deprecated: use Run instead
// RunContainer creates an instance of the OpenLDAP container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error) {
	return Run(ctx, "bitnami/openldap:2.6.6", opts...)
}

// Run creates an instance of the OpenLDAP container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error) {
	modulesOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithEnv(map[string]string{
			"LDAP_ADMIN_USERNAME": defaultUser,
			"LDAP_ADMIN_PASSWORD": defaultPassword,
			"LDAP_ROOT":           defaultRoot,
		}),
		testcontainers.WithExposedPorts("1389/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("** Starting slapd **"),
			wait.ForListeningPort("1389/tcp"),
		)),
	}

	modulesOpts = append(modulesOpts, opts...)

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, fmt.Errorf("openldap option: %w", err)
			}
		}
	}

	modulesOpts = append(modulesOpts, testcontainers.WithEnv(settings.env))

	ctr, err := testcontainers.Run(ctx, img, modulesOpts...)
	var c *OpenLDAPContainer
	if ctr != nil {
		c = &OpenLDAPContainer{
			Container:     ctr,
			adminUsername: settings.env["LDAP_ADMIN_USERNAME"],
			adminPassword: settings.env["LDAP_ADMIN_PASSWORD"],
			rootDn:        settings.env["LDAP_ROOT"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}
