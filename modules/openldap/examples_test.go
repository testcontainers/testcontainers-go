package openldap_test

import (
	"context"
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openldap"
)

func ExampleRun() {
	// runOpenLDAPContainer {
	ctx := context.Background()

	openldapContainer, err := openldap.Run(ctx, "bitnami/openldap:2.6.6")
	defer func() {
		if err := testcontainers.TerminateContainer(openldapContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := openldapContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	// connectToOpenLdap {
	ctx := context.Background()

	openldapContainer, err := openldap.Run(ctx, "bitnami/openldap:2.6.6")
	defer func() {
		if err := testcontainers.TerminateContainer(openldapContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connectionString, err := openldapContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	client, err := ldap.DialURL(connectionString)
	if err != nil {
		log.Printf("failed to connect to LDAP server: %s", err)
		return
	}
	defer client.Close()

	// First bind with a read only user
	err = client.Bind("cn=admin,dc=example,dc=org", "adminpassword")
	if err != nil {
		log.Printf("failed to bind to LDAP server: %s", err)
		return
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
		log.Printf("failed to search LDAP server: %s", err)
		return
	}

	if len(sr.Entries) != 1 {
		log.Print("User does not exist or too many entries returned")
		return
	}

	fmt.Println(sr.Entries[0].DN)

	// Output:
	// cn=user01,ou=users,dc=example,dc=org
}
