package mariadb_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

func ExampleRunContainer() {
	// runMariaDBContainer {
	ctx := context.Background()

	mariadbContainer, err := mariadb.RunContainer(ctx,
		testcontainers.WithImage("mariadb:11.0.3"),
		mariadb.WithConfigFile(filepath.Join("testdata", "my.cnf")),
		mariadb.WithScripts(filepath.Join("testdata", "schema.sql")),
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("root"),
		mariadb.WithPassword(""),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := mariadbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := mariadbContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
