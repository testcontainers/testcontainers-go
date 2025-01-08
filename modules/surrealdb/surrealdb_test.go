package surrealdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surrealdb/surrealdb.go"

	"github.com/testcontainers/testcontainers-go"
)

func TestSurrealDBSelect(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "surrealdb/surrealdb:v1.1.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	url, err := ctr.URL(ctx)
	require.NoError(t, err)

	db, err := surrealdb.New(url)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Use("test", "test")
	require.NoError(t, err)

	_, err = db.Create("person.tobie", map[string]any{
		"title": "Founder & CEO",
		"name": map[string]string{
			"first": "Tobie",
			"last":  "Morgan Hitchcock",
		},
		"marketing": true,
	})
	require.NoError(t, err)

	result, err := db.Select("person.tobie")
	require.NoError(t, err)

	resultData := result.([]any)[0].(map[string]interface{})
	require.Equal(t, "Founder & CEO", resultData["title"])
	require.Equal(t, "Tobie", resultData["name"].(map[string]interface{})["first"])
	require.Equal(t, "Morgan Hitchcock", resultData["name"].(map[string]interface{})["last"])
	require.Equal(t, true, resultData["marketing"])
}

func TestSurrealDBWithAuth(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "surrealdb/surrealdb:v1.1.1", WithAuthentication())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// websocketURL {
	url, err := ctr.URL(ctx)
	// }
	require.NoError(t, err)

	db, err := surrealdb.New(url)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Signin(map[string]string{"user": "root", "pass": "root"})
	require.NoError(t, err)

	_, err = db.Use("test", "test")
	require.NoError(t, err)

	_, err = db.Create("person.tobie", map[string]any{
		"title": "Founder & CEO",
		"name": map[string]string{
			"first": "Tobie",
			"last":  "Morgan Hitchcock",
		},
		"marketing": true,
	})
	require.NoError(t, err)

	result, err := db.Select("person.tobie")
	require.NoError(t, err)

	resultData := result.([]any)[0].(map[string]interface{})
	require.Equal(t, "Founder & CEO", resultData["title"])
	require.Equal(t, "Tobie", resultData["name"].(map[string]interface{})["first"])
	require.Equal(t, "Morgan Hitchcock", resultData["name"].(map[string]interface{})["last"])
	require.Equal(t, true, resultData["marketing"])
}
