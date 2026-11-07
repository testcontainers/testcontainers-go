package fakesnow_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/snowflakedb/gosnowflake"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/fakesnow"
)

func ExampleRun() {
	ctx := context.Background()

	fakesnowContainer, err := fakesnow.Run(ctx, "ghcr.io/tekumara/fakesnow:0.11.21")
	defer func() {
		if err := testcontainers.TerminateContainer(fakesnowContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := fakesnowContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleContainer_ConnectionString() {
	ctx := context.Background()

	fakesnowContainer, err := fakesnow.Run(ctx, "ghcr.io/tekumara/fakesnow:0.11.21")
	defer func() {
		if err := testcontainers.TerminateContainer(fakesnowContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	dsn, err := fakesnowContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		log.Printf("failed to open database: %s", err)
		return
	}
	defer db.Close()

	// fakesnow starts empty, so create what the DSN points at.
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS db1"); err != nil {
		log.Printf("failed to create database: %s", err)
		return
	}
	if _, err := db.Exec("CREATE SCHEMA IF NOT EXISTS db1.schema1"); err != nil {
		log.Printf("failed to create schema: %s", err)
		return
	}

	var planet string
	if err := db.QueryRow(`SELECT PARSE_JSON('{"planet":"saturn"}'):planet::varchar`).Scan(&planet); err != nil {
		log.Printf("failed to query: %s", err)
		return
	}

	fmt.Println(planet)

	// Output:
	// saturn
}
