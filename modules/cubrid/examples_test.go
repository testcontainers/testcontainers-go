package cubrid_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cubrid"
)

func ExampleRun() {
	// runCUBRIDContainer {
	ctx := context.Background()

	cubridContainer, err := cubrid.Run(ctx,
		"cubrid/cubrid:11.4",
		cubrid.WithDatabase("testdb"),
		cubrid.WithUsername("testuser"),
		cubrid.WithPassword("testpass"),
		cubrid.WithLocale("ko_KR.utf8"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(cubridContainer, testcontainers.StopTimeout(0)); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := cubridContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
