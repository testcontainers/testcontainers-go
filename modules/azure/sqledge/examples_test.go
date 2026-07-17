package sqledge_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/sqledge"
)

func ExampleRun() {
	// runSQLEdgeContainer {
	ctx := context.Background()

	sqlEdgeContainer, err := sqledge.Run(ctx,
		"mcr.microsoft.com/azure-sql-edge:1.0.7",
		sqledge.WithAcceptEULA(),
		sqledge.WithPassword("Strong!Passw0rd"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(sqlEdgeContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := sqlEdgeContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
