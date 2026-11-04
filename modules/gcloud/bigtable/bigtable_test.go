package bigtable_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcbigtable "github.com/testcontainers/testcontainers-go/modules/gcloud/bigtable"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/testcbt"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	bigTableContainer, err := tcbigtable.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:emulators",
		tcbigtable.WithProjectID("bigtable-project"),
	)
	testcontainers.CleanupContainer(t, bigTableContainer)
	require.NoError(t, err)

	projectID := bigTableContainer.ProjectID()

	const (
		instanceID = "test-instance"
		tableName  = "test-table"
	)

	// WithHostPortAccess forwards localhost ports. This fixed relay retains
	// the actual generated host as well as its mapped port, even for remote Docker.
	relay, err := testcbt.NewRelay(bigTableContainer.URI())
	require.NoError(t, err)
	t.Cleanup(relay.Close)
	t.Logf("cbt relay upstream: %s", bigTableContainer.URI())

	helper, err := testcontainers.Run(ctx, testcbt.Image,
		testcontainers.WithEntrypoint("/bin/sleep"),
		testcontainers.WithCmd("180"),
		testcontainers.WithHostPortAccess(relay.Port()),
	)
	testcontainers.CleanupContainer(t, helper)
	require.NoError(t, err)
	endpoint := testcbt.Endpoint(relay.Port())

	_, err = testcbt.Run(ctx, helper, endpoint, projectID, instanceID, "createtable", tableName)
	require.NoError(t, err)

	_, err = testcbt.Run(ctx, helper, endpoint, projectID, instanceID, "createfamily", tableName, "name")
	require.NoError(t, err)

	// Omitting @timestamp selects cbt's client-current bigtable.Now().
	_, err = testcbt.Run(ctx, helper, endpoint, projectID, instanceID, "set", tableName, "1", "name:firstName=Gopher")
	require.NoError(t, err)

	// The trailing colon preserves FamilyFilter("name") without adding a qualifier filter.
	output, err := testcbt.Run(ctx, helper, endpoint, projectID, instanceID, "lookup", tableName, "1", "columns=name:")
	require.NoError(t, err)

	row, column, value, err := testcbt.LookupCell(output)
	require.NoError(t, err)
	require.Equal(t, "1", row)
	require.Equal(t, "name:firstName", column)
	require.Equal(t, "Gopher", value)
}
