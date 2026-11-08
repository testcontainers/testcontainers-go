package firestore_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcfirestore "github.com/testcontainers/testcontainers-go/modules/gcloud/firestore"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/testrest"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	firestoreContainer, err := tcfirestore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:emulators",
		tcfirestore.WithProjectID("firestore-project"),
	)
	testcontainers.CleanupContainer(t, firestoreContainer)
	require.NoError(t, err)

	projectID := firestoreContainer.ProjectID()

	requestCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 30 * time.Second}
	defer client.CloseIdleConnections()
	endpoint := "http://" + firestoreContainer.URI() + "/v1/projects/" + url.PathEscape(projectID) + "/databases/(default)/documents/users"

	// Firestore's SDK uses Go field names unless a firestore tag is present;
	// the original Person's json tags do not rename these stored fields.
	document := map[string]any{"fields": map[string]any{
		"Firstname": map[string]string{"stringValue": "Ada"},
		"Lastname":  map[string]string{"stringValue": "Lovelace"},
	}}
	// createDocument, like DocumentRef.Create, fails if the document exists.
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+"?documentId=alovelace", document, nil))

	var saved struct {
		Fields map[string]struct {
			StringValue string `json:"stringValue"`
		} `json:"fields"`
	}
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodGet, endpoint+"/alovelace", nil, &saved))
	require.Equal(t, "Ada", saved.Fields["Firstname"].StringValue)
	require.Equal(t, "Lovelace", saved.Fields["Lastname"].StringValue)
}

func TestRunWithDatastore(t *testing.T) {
	ctx := context.Background()

	firestoreContainer, err := tcfirestore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:emulators",
		tcfirestore.WithProjectID("firestore-project"),
		tcfirestore.WithDatastoreMode(),
	)
	testcontainers.CleanupContainer(t, firestoreContainer)
	require.NoError(t, err)

	projectID := firestoreContainer.ProjectID()

	requestCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 30 * time.Second}
	defer client.CloseIdleConnections()
	endpoint := "http://" + firestoreContainer.URI() + "/v1/projects/" + url.PathEscape(projectID)

	// This separate emulator is in Datastore mode, using the Datastore API
	// and the same named key/upsert as datastore.Client.Put.
	key := map[string]any{
		"partitionId": map[string]string{"projectId": projectID},
		"path":        []any{map[string]string{"kind": "users", "name": "alovelace"}},
	}
	entity := map[string]any{
		"key": key,
		"properties": map[string]any{
			"Firstname": map[string]string{"stringValue": "Ada"},
			"Lastname":  map[string]string{"stringValue": "Lovelace"},
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
	require.Equal(t, "Ada", saved.Found[0].Entity.Properties["Firstname"].StringValue)
	require.Equal(t, "Lovelace", saved.Found[0].Entity.Properties["Lastname"].StringValue)
}
