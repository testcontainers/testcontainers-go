package cockroachdb_test

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func ExampleRunContainer() {
	// runCockroachDBContainer {
	ctx := context.Background()

	cockroachdbContainer, err := cockroachdb.RunContainer(ctx)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := cockroachdbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := cockroachdbContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}
	fmt.Println(state.Running)

	addr, err := cockroachdbContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}
	u, err := url.Parse(addr)
	if err != nil {
		log.Fatalf("failed to parse connection string: %s", err)
	}
	u.Host = fmt.Sprintf("%s:%s", u.Hostname(), "xxx")
	fmt.Println(u.String())

	// Output:
	// true
	// postgres://root@localhost:xxx/defaultdb?sslmode=disable
}
