package cockroachdb

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Task represents a unit of work to complete. We're going to be using this in
// our example as a way to organize data that is being manipulated in
// the database.
type task struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	DateDue     *time.Time `json:"date_due"`
	DateCreated time.Time  `json:"date_created"`
	DateUpdated time.Time  `json:"date_updated"`
}

func initCockroachDB(ctx context.Context, db *sql.DB) error {
	// Actual SQL for initializing the database should probably live elsewhere
	const query = `CREATE DATABASE projectmanagement;
		CREATE TABLE projectmanagement.task(
			id uuid primary key not null,
			description varchar(255) not null,
			date_due timestamp with time zone,
			date_created timestamp with time zone not null,
			date_updated timestamp with time zone not null);`
	_, err := db.ExecContext(ctx, query)
	return err
}

func truncateCockroachDB(ctx context.Context, db *sql.DB) error {
	const query = `TRUNCATE projectmanagement.task`
	_, err := db.ExecContext(ctx, query)
	return err
}

func TestIntegrationDBInsertSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	cdbContainer, err := startContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := cdbContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	db, err := sql.Open("pgx", cdbContainer.URI+"/projectmanagement")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, initCockroachDB(ctx, db))
	defer func(t *testing.T, ctx context.Context, db *sql.DB) {
		require.NoError(t, truncateCockroachDB(ctx, db))
	}(t, ctx, db)

	now := time.Now()

	// Insert data
	tsk := task{ID: uuid.NewString(), Description: "Update resumé", DateCreated: now, DateUpdated: now}
	const insertQuery = `insert into "task" (id, description, date_due, date_created, date_updated)
		values ($1, $2, $3, $4, $5)`
	_, err = db.ExecContext(
		ctx,
		insertQuery,
		tsk.ID,
		tsk.Description,
		tsk.DateDue,
		tsk.DateCreated,
		tsk.DateUpdated)
	require.NoError(t, err)

	// Select data
	savedTsk := task{ID: tsk.ID}
	const findQuery = `select description, date_due, date_created, date_updated
		from task
		where id = $1`
	row := db.QueryRowContext(ctx, findQuery, tsk.ID)
	err = row.Scan(&savedTsk.Description, &savedTsk.DateDue, &savedTsk.DateCreated, &savedTsk.DateUpdated)
	require.NoError(t, err)
	assert.Equal(t, tsk.ID, savedTsk.ID)
	assert.Equal(t, tsk.Description, savedTsk.Description)
	assert.Equal(t, tsk.DateDue, savedTsk.DateDue)
}
