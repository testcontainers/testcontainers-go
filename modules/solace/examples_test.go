package solace_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	sc "github.com/testcontainers/testcontainers-go/modules/solace"
)

func ExampleRun() {
	// runSolaceContainer {
	ctx := context.Background()
	ctr, err := sc.Run(ctx, "solace/solace-pubsub-standard:latest",
		sc.WithCredentials("admin", "admin"),
		sc.WithServices(sc.ServiceAMQP, sc.ServiceManagement),
		testcontainers.WithEnv(map[string]string{
			"username_admin_globalaccesslevel": "admin",
			"username_admin_password":          "admin",
		}),
		sc.WithShmSize(1<<30),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	fmt.Println(err)
	// }

	// Output:
	// <nil>
}
