package couchdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/couchdb"
)

func ExampleRun() {
	// runCouchDBContainer {
	ctx := context.Background()

	couchdbContainer, err := couchdb.Run(ctx, "couchdb:3",
		couchdb.WithAdminCredentials("admin", "password"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(couchdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	connStr, err := couchdbContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	fmt.Println(connStr[:len("http://admin:password@")])

	// Output:
	// http://admin:password@
}
