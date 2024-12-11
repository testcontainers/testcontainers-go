package openldap_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openldap"
)

func TestOpenLDAP(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx, "bitnami/openldap:2.6.6")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestOpenLDAPWithAdminUsernameAndPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx,
		"bitnami/openldap:2.6.6",
		openldap.WithAdminUsername("openldap"),
		openldap.WithAdminPassword("openldap"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := ldap.DialURL(connectionString)
	require.NoError(t, err)
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=openldap,dc=example,dc=org", "openldap")
	require.NoError(t, err)
}

func TestOpenLDAPWithDifferentRoot(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx, "bitnami/openldap:2.6.6", openldap.WithRoot("dc=mydomain,dc=com"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	client, err := ldap.DialURL(connectionString)
	require.NoError(t, err)
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=admin,dc=mydomain,dc=com", "adminpassword")
	require.NoError(t, err)
}

func TestOpenLDAPLoadLdif(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx, "bitnami/openldap:2.6.6")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

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

	err = ctr.LoadLdif(ctx, []byte(ldif))
	// }
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := ldap.DialURL(connectionString)
	require.NoError(t, err)
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=admin,dc=example,dc=org", "adminpassword")
	require.NoError(t, err)

	result, err := client.Search(&ldap.SearchRequest{
		BaseDN:     "uid=test.user,ou=users,dc=example,dc=org",
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=*)",
		Attributes: []string{"dn"},
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	require.Equal(t, "uid=test.user,ou=users,dc=example,dc=org", result.Entries[0].DN)
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
	require.NoError(t, err)

	_, err = f.WriteString(ldif)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	ctr, err := openldap.Run(ctx, "bitnami/openldap:2.6.6", openldap.WithInitialLdif(f.Name()))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := ldap.DialURL(connectionString)
	require.NoError(t, err)
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=admin,dc=example,dc=org", "adminpassword")
	require.NoError(t, err)

	result, err := client.Search(&ldap.SearchRequest{
		BaseDN:     "uid=test.user,ou=users,dc=example,dc=org",
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=*)",
		Attributes: []string{"dn"},
	})
	require.NoError(t, err)

	require.Len(t, result.Entries, 1)
	require.Equal(t, "uid=test.user,ou=users,dc=example,dc=org", result.Entries[0].DN)
}
