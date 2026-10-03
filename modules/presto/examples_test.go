package presto_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/presto"
)

func ExampleRun() {
	// runPrestoContainer {
	ctx := context.Background()

	prestoContainer, err := presto.Run(ctx, "prestodb/presto:0.286")
	defer func() {
		if err := testcontainers.TerminateContainer(prestoContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	connStr, err := prestoContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	fmt.Println(connStr[:7]) // print "http://"

	// Output:
	// http://
}
