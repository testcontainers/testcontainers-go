package mariadb_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

func ExampleRun() {
	// runMariaDBContainer {
	ctx := context.Background()

	mariadbContainer, err := mariadb.Run(ctx,
		"mariadb:11.0.3",
		mariadb.WithConfigFile(filepath.Join("testdata", "my.cnf")),
		mariadb.WithScripts(filepath.Join("testdata", "schema.sql")),
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("root"),
		mariadb.WithPassword(""),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(mariadbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := mariadbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
