package influxdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/influxdb"
)

func ExampleRun() {
	// runInfluxContainer {
	ctx := context.Background()

	influxdbContainer, err := influxdb.Run(ctx,
		"influxdb:1.8.10",
		influxdb.WithDatabase("influx"),
		influxdb.WithUsername("root"),
		influxdb.WithPassword("password"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(influxdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := influxdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
