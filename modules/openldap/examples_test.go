package openldap_test

import (
	"context"
	"fmt"

	"github.com/go-ldap/ldap/v3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openldap"
)

func ExampleRunContainer() {
	// runOpenLDAPContainer {
	ctx := context.Background()

	openldapContainer, err := openldap.RunContainer(ctx, testcontainers.WithImage("bitnami/openldap:2.6.6"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := openldapContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := openldapContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connect() {
	// connectToOpenLdap {
	ctx := context.Background()

	openldapContainer, err := openldap.RunContainer(ctx, testcontainers.WithImage("bitnami/openldap:2.6.6"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := openldapContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	connectionString, err := openldapContainer.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	client, err := ldap.DialURL(connectionString)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=admin,dc=example,dc=org", "adminpassword")
	if err != nil {
		panic(err)
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=org",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalPerson)(uid=user01))",
		[]string{"dn"},
		nil,
	)

	sr, err := client.Search(searchRequest)
	if err != nil {
		panic(err)
	}

	if len(sr.Entries) != 1 {
		panic("User does not exist or too many entries returned")
	}

	fmt.Println(sr.Entries[0].DN)

	// Output:
	// cn=user01,ou=users,dc=example,dc=org
}
