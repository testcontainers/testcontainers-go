package openldap

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"LDAP_ADMIN_USERNAME": defaultUser,
			"LDAP_ADMIN_PASSWORD": defaultPassword,
			"LDAP_ROOT":           defaultRoot,
		},
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the OpenLDAP container.
type Option func(opts *options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(_ *testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithAdminUsername sets the initial admin username to be created when the container starts
// It is used in conjunction with WithAdminPassword to set a username and its password.
// It will create the specified user with admin power.
func WithAdminUsername(username string) Option {
	return func(opts *options) error {
		opts.env["LDAP_ADMIN_USERNAME"] = username

		return nil
	}
}

// WithAdminPassword sets the initial admin password of the user to be created when the container starts
// It is used in conjunction with WithAdminUsername to set a username and its password.
// It will set the admin password for OpenLDAP.
func WithAdminPassword(password string) Option {
	return func(opts *options) error {
		opts.env["LDAP_ADMIN_PASSWORD"] = password

		return nil
	}
}

// WithRoot sets the root of the OpenLDAP instance
func WithRoot(root string) Option {
	return func(opts *options) error {
		opts.env["LDAP_ROOT"] = root

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
