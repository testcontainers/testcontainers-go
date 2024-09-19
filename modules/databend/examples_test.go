package databend_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/datafuselabs/databend-go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/databend"
)

func ExampleRun() {
	ctx := context.Background()

	databendContainer, err := databend.Run(ctx,
		"datafuselabs/databend:v1.2.615",
		databend.WithUsername("test1"),
		databend.WithPassword("pass1"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(databendContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := databendContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	ctx := context.Background()

	databendContainer, err := databend.Run(ctx,
		"datafuselabs/databend:v1.2.615",
		databend.WithUsername("root"),
		databend.WithPassword("password"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(databendContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connectionString, err := databendContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	db, err := sql.Open("databend", connectionString)
	if err != nil {
		log.Printf("failed to connect to Databend: %s", err)
		return
	}
	defer db.Close()

	var i int
	row := db.QueryRow("select 1")
	err = row.Scan(&i)
	if err != nil {
		log.Printf("failed to scan result: %s", err)
		return
	}

	fmt.Println(i)

	// Output:
	// 1
}
