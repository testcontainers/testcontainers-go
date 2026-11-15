package openldap_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openldap"
)

func ExampleRun() {
	// runOpenLDAPContainer {
	ctx := context.Background()

	openldapContainer, err := openldap.Run(ctx, "bitnamilegacy/openldap:2.6.6")
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
