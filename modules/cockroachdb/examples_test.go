package cockroachdb_test

import (
	"context"
	"fmt"
	"net/url"

	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func ExampleRunContainer() {
	// runCockroachDBContainer {
	ctx := context.Background()

	cockroachdbContainer, err := cockroachdb.RunContainer(ctx, cockroachdb.WithImageTag("latest-v23.1"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := cockroachdbContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := cockroachdbContainer.State(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(state.Running)

	addr, err := cockroachdbContainer.Address(ctx)
	if err != nil {
		panic(err)
	}
	u, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	u.Host = fmt.Sprintf("%s:%s", u.Hostname(), "xxx")
	fmt.Println(u.String())

	// Output:
	// true
	// postgres://root@localhost:xxx/defaultdb?sslmode=disable
}
