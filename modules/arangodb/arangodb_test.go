package arangodb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcarangodb "github.com/testcontainers/testcontainers-go/modules/arangodb"
)

func TestArangoDB(t *testing.T) {
	ctx := context.Background()

	const password = "t3stc0ntain3rs!"

	ctr, err := tcarangodb.Run(ctx, "arangodb:3.11.5", tcarangodb.WithRootPassword(password))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	httpAddress, err := ctr.HTTPEndpoint(ctx)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, httpAddress+"/_api/version", nil)
	require.NoError(t, err)
	username, credentialPassword := ctr.Credentials()
	req.SetBasicAuth(username, credentialPassword)
	resp, err := (&http.Client{}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var versionInfo struct {
		Server string `json:"server"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&versionInfo))
	require.Equal(t, "arango", versionInfo.Server)
}
