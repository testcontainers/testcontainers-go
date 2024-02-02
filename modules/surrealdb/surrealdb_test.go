package surrealdb

import (
	"context"
	"testing"

	"github.com/surrealdb/surrealdb.go"
	"github.com/testcontainers/testcontainers-go"
)

func TestSurrealDBSelect(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("surrealdb/surrealdb:v1.1.1"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// websocketURL {
	url, err := container.URL(ctx)
	// }
	if err != nil {
		panic(err)
	}

	db, err := surrealdb.New(url)
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
		"title": "Founder & CEO",
		"name": map[string]string{
			"first": "Tobie",
			"last":  "Morgan Hitchcock",
		},
		"marketing": true,
	}); err != nil {
		panic(err)
	}

	result, err := db.Select("person.tobie")
	if err != nil {
		panic(err)
	}

	resultData := result.([]any)[0].(map[string]interface{})
	if resultData["title"] != "Founder & CEO" {
		panic("title is not Founder & CEO")
	}
	if resultData["name"].(map[string]interface{})["first"] != "Tobie" {
		panic("name.first is not Tobie")
	}
	if resultData["name"].(map[string]interface{})["last"] != "Morgan Hitchcock" {
		panic("name.last is not Morgan Hitchcock")
	}
	if resultData["marketing"] != true {
		panic("marketing is not true")
	}
}

func TestSurrealDBNoAuth(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("surrealdb/surrealdb:v1.1.1"), WithAuthentication(false))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	url, err := container.URL(ctx)
	if err != nil {
		panic(err)
	}

	db, err := surrealdb.New(url)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if _, err := db.Use("test", "test"); err != nil {
		panic(err)
	}

	if _, err := db.Create("person.tobie", map[string]any{
		"title": "Founder & CEO",
		"name": map[string]string{
			"first": "Tobie",
			"last":  "Morgan Hitchcock",
		},
		"marketing": true,
	}); err != nil {
		panic(err)
	}

	result, err := db.Select("person.tobie")
	if err != nil {
		panic(err)
	}

	resultData := result.([]any)[0].(map[string]interface{})
	if resultData["title"] != "Founder & CEO" {
		panic("title is not Founder & CEO")
	}
	if resultData["name"].(map[string]interface{})["first"] != "Tobie" {
		panic("name.first is not Tobie")
	}
	if resultData["name"].(map[string]interface{})["last"] != "Morgan Hitchcock" {
		panic("name.last is not Morgan Hitchcock")
	}
	if resultData["marketing"] != true {
		panic("marketing is not true")
	}
}
