package openldap

import (
	"context"
	"testing"

	"github.com/go-ldap/ldap/v3"

	"github.com/testcontainers/testcontainers-go"
)

func TestOpenLDAP(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("bitnami/openldap:2.6.6"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
}

func TestOpenLDAPWithAdminUsernameAndPassword(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("bitnami/openldap:2.6.6"), WithAdminUsername("openldap"), WithAdminPassword("openldap"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	client, err := ldap.DialURL(connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=openldap,dc=example,dc=org", "openldap")
	if err != nil {
		t.Fatal(err)
	}
}

func TestOpenLDAPWithDifferentRoot(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("bitnami/openldap:2.6.6"), WithRoot("dc=mydomain,dc=com"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	client, err := ldap.DialURL(connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=admin,dc=mydomain,dc=com", "adminpassword")
	if err != nil {
		t.Fatal(err)
	}
}
