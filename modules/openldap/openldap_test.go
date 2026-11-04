package openldap_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openldap"
)

func TestOpenLDAP(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx, "bitnamilegacy/openldap:2.6.6")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestOpenLDAPWithAdminUsernameAndPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx,
		"bitnamilegacy/openldap:2.6.6",
		openldap.WithAdminUsername("openldap"),
		openldap.WithAdminPassword("openldap"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(connectionString, "ldap://"), "expected LDAP endpoint, got: %s", connectionString)
	// Bind and Search against this generated public endpoint are covered in
	// examples/openldap_test.go; this test retains module setup and options.
	inspect, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	require.Contains(t, inspect.Config.Env, "LDAP_ADMIN_USERNAME=openldap")
	require.Contains(t, inspect.Config.Env, "LDAP_ADMIN_PASSWORD=openldap")
}

func TestOpenLDAPWithDifferentRoot(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx, "bitnamilegacy/openldap:2.6.6", openldap.WithRoot("dc=mydomain,dc=com"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(connectionString, "ldap://"), "expected LDAP endpoint, got: %s", connectionString)
	// Bind and Search against this generated public endpoint are covered in
	// examples/openldap_test.go; this test retains module setup and options.
	inspect, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	require.Contains(t, inspect.Config.Env, "LDAP_ROOT=dc=mydomain,dc=com")
}

func TestOpenLDAPLoadLdif(t *testing.T) {
	ctx := context.Background()

	ctr, err := openldap.Run(ctx, "bitnamilegacy/openldap:2.6.6")
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

	require.True(t, strings.HasPrefix(connectionString, "ldap://"), "expected LDAP endpoint, got: %s", connectionString)
	// Bind and Search against this generated public endpoint are covered in
	// examples/openldap_test.go; this test retains module setup and options.
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

	ctr, err := openldap.Run(ctx, "bitnamilegacy/openldap:2.6.6", openldap.WithInitialLdif(f.Name()))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(connectionString, "ldap://"), "expected LDAP endpoint, got: %s", connectionString)
	// Bind and Search against this generated public endpoint are covered in
	// examples/openldap_test.go; this test retains module setup and options.
}
