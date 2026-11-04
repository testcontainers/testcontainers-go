package surrealdb_test

import (
	"context"
	"fmt"

	client "github.com/surrealdb/surrealdb.go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/surrealdb"
)

func ExampleRun_connect() {
	// connectWithWebSocketClient {
	ctx := context.Background()
	ctr, err := surrealdb.Run(ctx, "surrealdb/surrealdb:v1.1.1", surrealdb.WithAuthentication())
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	endpoint, err := ctr.URL(ctx)
	if err != nil {
		panic(err)
	}
	db, err := client.New(endpoint)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if _, err := db.Signin(map[string]string{"user": "root", "pass": "root"}); err != nil {
		panic(err)
	}
	if _, err := db.Use("test", "test"); err != nil {
		panic(err)
	}
	if _, err := db.Create("person.tobie", map[string]any{
		"title":     "Founder & CEO",
		"name":      map[string]string{"first": "Tobie", "last": "Morgan Hitchcock"},
		"marketing": true,
	}); err != nil {
		panic(err)
	}
	result, err := db.Select("person.tobie")
	if err != nil {
		panic(err)
	}
	person := result.([]any)[0].(map[string]any)
	fmt.Println(person["title"])
	fmt.Println(person["name"].(map[string]any)["first"])
	fmt.Println(person["name"].(map[string]any)["last"])
	fmt.Println(person["marketing"])
	// }

	// Output:
	// Founder & CEO
	// Tobie
	// Morgan Hitchcock
	// true
}
