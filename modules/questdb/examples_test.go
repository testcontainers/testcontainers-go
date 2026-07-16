package questdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/questdb"
)

func ExampleRun() {
	// runQuestDBContainer {
	ctx := context.Background()

	questdbContainer, err := questdb.Run(ctx, "questdb/questdb:7.4.2")
	defer func() {
		if err := testcontainers.TerminateContainer(questdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := questdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleContainer_HTTPEndpoint() {
	ctx := context.Background()

	questdbContainer, err := questdb.Run(ctx, "questdb/questdb:7.4.2")
	defer func() {
		if err := testcontainers.TerminateContainer(questdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// httpEndpoint {
	httpEndpoint, err := questdbContainer.HTTPEndpoint(ctx)
	// }
	if err != nil {
		log.Printf("failed to get HTTP endpoint: %s", err)
		return
	}

	fmt.Println(httpEndpoint != "")

	// Output:
	// true
}

func ExampleContainer_PGEndpoint() {
	ctx := context.Background()

	questdbContainer, err := questdb.Run(ctx, "questdb/questdb:7.4.2")
	defer func() {
		if err := testcontainers.TerminateContainer(questdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// pgEndpoint {
	pgEndpoint, err := questdbContainer.PGEndpoint(ctx)
	// }
	if err != nil {
		log.Printf("failed to get PG endpoint: %s", err)
		return
	}

	fmt.Println(pgEndpoint != "")

	// Output:
	// true
}

func ExampleContainer_InfluxDBEndpoint() {
	ctx := context.Background()

	questdbContainer, err := questdb.Run(ctx, "questdb/questdb:7.4.2")
	defer func() {
		if err := testcontainers.TerminateContainer(questdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// influxDBEndpoint {
	ilpEndpoint, err := questdbContainer.InfluxDBEndpoint(ctx)
	// }
	if err != nil {
		log.Printf("failed to get InfluxDB line protocol endpoint: %s", err)
		return
	}

	fmt.Println(ilpEndpoint != "")

	// Output:
	// true
}
