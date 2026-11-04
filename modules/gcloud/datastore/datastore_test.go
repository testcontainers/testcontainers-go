package datastore_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcdatastore "github.com/testcontainers/testcontainers-go/modules/gcloud/datastore"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/testrest"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	datastoreContainer, err := tcdatastore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:emulators",
		tcdatastore.WithProjectID("datastore-project"),
	)
	testcontainers.CleanupContainer(t, datastoreContainer)
	require.NoError(t, err)

	projectID := datastoreContainer.ProjectID()

	requestCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 30 * time.Second}
	defer client.CloseIdleConnections()
	endpoint := "http://" + datastoreContainer.URI() + "/v1/projects/" + url.PathEscape(projectID)

	// NameKey("Task", "sample", nil), with the same upsert as Client.Put.
	key := map[string]any{
		"partitionId": map[string]string{"projectId": projectID},
		"path":        []any{map[string]string{"kind": "Task", "name": "sample"}},
	}
	entity := map[string]any{
		"key": key,
		"properties": map[string]any{
			"Description": map[string]string{"stringValue": "my description"},
		},
	}
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+":commit", map[string]any{
		"mode":      "NON_TRANSACTIONAL",
		"mutations": []any{map[string]any{"upsert": entity}},
	}, nil))

	type lookupResult struct {
		Found []struct {
			Entity struct {
				Properties map[string]struct {
					StringValue string `json:"stringValue"`
				} `json:"properties"`
			} `json:"entity"`
		} `json:"found"`
		Missing  []json.RawMessage `json:"missing"`
		Deferred []json.RawMessage `json:"deferred"`
	}
	keys := any([]any{key})
	var saved lookupResult
	for {
		require.NoError(t, requestCtx.Err())
		saved = lookupResult{}
		require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+":lookup", map[string]any{"keys": keys}, &saved))
		require.Empty(t, saved.Missing)
		if len(saved.Deferred) == 0 {
			break
		}
		keys = saved.Deferred
	}
	require.Len(t, saved.Found, 1)
	require.Equal(t, "my description", saved.Found[0].Entity.Properties["Description"].StringValue)
}
