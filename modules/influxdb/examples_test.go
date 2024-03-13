package influxdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/influxdb"
)

func ExampleRunContainer() {
	// runInfluxContainer {
	ctx := context.Background()

	influxdbContainer, err := influxdb.RunContainer(
		ctx, testcontainers.WithImage("influxdb:1.8.10"),
		influxdb.WithDatabase("influx"),
		influxdb.WithUsername("root"),
		influxdb.WithPassword("password"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := influxdbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := influxdbContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
