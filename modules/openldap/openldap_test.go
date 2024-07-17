package openldap_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-ldap/ldap/v3"

	"github.com/testcontainers/testcontainers-go/modules/openldap"
)

func TestOpenLDAP(t *testing.T) {
	ctx := context.Background()

	container, err := openldap.Run(ctx, "bitnami/openldap:2.6.6")
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

	container, err := openldap.Run(ctx,
		"bitnami/openldap:2.6.6",
		openldap.WithAdminUsername("openldap"),
		openldap.WithAdminPassword("openldap"),
	)
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

	container, err := openldap.Run(ctx, "bitnami/openldap:2.6.6", openldap.WithRoot("dc=mydomain,dc=com"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// connectionString {
	connectionString, err := container.ConnectionString(ctx)
	// }
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

func TestOpenLDAPLoadLdif(t *testing.T) {
	ctx := context.Background()

	container, err := openldap.Run(ctx, "bitnami/openldap:2.6.6")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// loadLdif {
	ldif := `
dn: uid=test.user,ou=users,dc=example,dc=org
changetype: add
objectclass: iNetOrgPerson
cn: Test User
sn: Test
mail: test.user@example.org
userPassword: Password1
`

	err = container.LoadLdif(ctx, []byte(ldif))
	// }
	if err != nil {
		t.Fatal(err)
	}

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
	err = client.Bind("cn=admin,dc=example,dc=org", "adminpassword")
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Search(&ldap.SearchRequest{
		BaseDN:     "uid=test.user,ou=users,dc=example,dc=org",
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=*)",
		Attributes: []string{"dn"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Entries) != 1 {
		t.Fatal("Invalid number of entries returned", result.Entries)
	}
	if result.Entries[0].DN != "uid=test.user,ou=users,dc=example,dc=org" {
		t.Fatal("Invalid entry returned", result.Entries[0].DN)
	}
}

func TestOpenLDAPWithInitialLdif(t *testing.T) {
	ctx := context.Background()

	ldif := `dn: uid=test.user,ou=users,dc=example,dc=org
changetype: add
objectclass: iNetOrgPerson
cn: Test User
sn: Test
mail: test.user@example.org
userPassword: Password1
`

	f, err := os.CreateTemp(t.TempDir(), "test.ldif")
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.WriteString(ldif)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	container, err := openldap.Run(ctx, "bitnami/openldap:2.6.6", openldap.WithInitialLdif(f.Name()))
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
	err = client.Bind("cn=admin,dc=example,dc=org", "adminpassword")
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Search(&ldap.SearchRequest{
		BaseDN:     "uid=test.user,ou=users,dc=example,dc=org",
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=*)",
		Attributes: []string{"dn"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Entries) != 1 {
		t.Fatal("Invalid number of entries returned", result.Entries)
	}
	if result.Entries[0].DN != "uid=test.user,ou=users,dc=example,dc=org" {
		t.Fatal("Invalid entry returned", result.Entries[0].DN)
	}
}
