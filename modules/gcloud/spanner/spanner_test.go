package spanner_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/testrest"
	tcspanner "github.com/testcontainers/testcontainers-go/modules/gcloud/spanner"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	spannerContainer, err := tcspanner.Run(
		ctx,
		"gcr.io/cloud-spanner-emulator/emulator:1.5.40",
		tcspanner.WithProjectID("spanner-project"),
		// The image runs a REST gateway on 9020 in addition to gRPC on 9010.
		testcontainers.WithExposedPorts("9020/tcp"),
		testcontainers.WithAdditionalWaitStrategy(wait.ForListeningPort("9020/tcp")),
	)
	testcontainers.CleanupContainer(t, spannerContainer)
	require.NoError(t, err)

	projectID := spannerContainer.ProjectID()

	const (
		instanceID   = "test-instance"
		databaseName = "test-db"
	)

	requestCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	// Keep coverage of the public gRPC URI alongside the REST operations below.
	host, err := spannerContainer.Host(requestCtx)
	require.NoError(t, err)
	port, err := spannerContainer.MappedPort(requestCtx, "9010/tcp")
	require.NoError(t, err)
	require.Equal(t, net.JoinHostPort(host, port.Port()), spannerContainer.URI())
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(requestCtx, "tcp", spannerContainer.URI())
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	client := &http.Client{Timeout: 30 * time.Second}
	defer client.CloseIdleConnections()
	endpoint, err := spannerContainer.PortEndpoint(requestCtx, "9020/tcp", "http")
	require.NoError(t, err)
	endpoint += "/v1/"

	type operation struct {
		Name  string `json:"name"`
		Done  bool   `json:"done"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Response json.RawMessage `json:"response"`
	}
	// Both admin calls return long-running operations. Poll their returned
	// resource names and check errors just as Operation.Wait does.
	waitOperation := func(op operation) {
		t.Helper()
		require.NotEmpty(t, op.Name)
		operationName := op.Name
		for !op.Done {
			select {
			case <-requestCtx.Done():
				t.Fatalf("wait for %s: %v", operationName, requestCtx.Err())
			case <-time.After(100 * time.Millisecond):
			}
			op = operation{}
			require.NoError(t, testrest.Do(requestCtx, client, http.MethodGet, endpoint+operationName, nil, &op))
		}
		require.Nil(t, op.Error, "operation %s failed: %+v", operationName, op.Error)
		require.NotEmpty(t, op.Response)
	}
	var instanceOp operation
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+"projects/"+projectID+"/instances", map[string]any{
		"instanceId": instanceID,
		"instance":   map[string]string{"displayName": instanceID},
	}, &instanceOp))
	waitOperation(instanceOp)

	instanceName := fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID)
	var databaseOp operation
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+instanceName+"/databases", map[string]any{
		"createStatement": fmt.Sprintf("CREATE DATABASE `%s`", databaseName),
		"extraStatements": []string{"CREATE TABLE Languages (Language STRING(MAX), Mascot STRING(MAX)) PRIMARY KEY (Language)"},
	}, &databaseOp))
	waitOperation(databaseOp)

	db := instanceName + "/databases/" + databaseName
	var session struct {
		Name string `json:"name"`
	}
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+db+"/sessions", map[string]any{}, &session))
	require.NotEmpty(t, session.Name)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		require.NoError(t, testrest.Do(cleanupCtx, client, http.MethodDelete, endpoint+session.Name, nil, nil))
	}()

	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+session.Name+":commit", map[string]any{
		"singleUseTransaction": map[string]any{"readWrite": map[string]any{}},
		"mutations": []any{map[string]any{"insert": map[string]any{
			"table":   "Languages",
			"columns": []string{"language", "mascot"},
			"values":  [][]string{{"Go", "Gopher"}},
		}}},
	}, nil))

	var result struct {
		Metadata struct {
			RowType struct {
				Fields []struct {
					Name string `json:"name"`
				} `json:"fields"`
			} `json:"rowType"`
		} `json:"metadata"`
		Rows [][]string `json:"rows"`
	}
	require.NoError(t, testrest.Do(requestCtx, client, http.MethodPost, endpoint+session.Name+":read", map[string]any{
		"table":       "Languages",
		"columns":     []string{"mascot"},
		"keySet":      map[string]any{"keys": [][]string{{"Go"}}},
		"transaction": map[string]any{"singleUse": map[string]any{"readOnly": map[string]bool{"strong": true}}},
	}, &result))
	require.Len(t, result.Metadata.RowType.Fields, 1)
	require.Equal(t, "Mascot", result.Metadata.RowType.Fields[0].Name)
	require.Len(t, result.Rows, 1)
	require.Len(t, result.Rows[0], 1)
	require.Equal(t, "Gopher", result.Rows[0][0])
}
