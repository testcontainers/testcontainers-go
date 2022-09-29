# CockroachDB

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Task represents a unit of work to complete. We're going to be using this in
// our example as a way to organize data that is being manipulated in
// the database.
type task struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	DateDue     *time.Time `json:"date_due,string"`
	DateCreated time.Time  `json:"date_created,string"`
	DateUpdated time.Time  `json:"date_updated"`
}

type cockroachDBContainer struct {
	testcontainers.Container
	URI string
}

func setupCockroachDB(ctx context.Context) (*cockroachDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "cockroachdb/cockroach:latest-v21.1",
		ExposedPorts: []string{"26257/tcp", "8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080"),
		Cmd:          []string{"start-single-node", "--insecure"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "26257")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("postgres://root@%s:%s", hostIP, mappedPort.Port())

	return &cockroachDBContainer{Container: container, URI: uri}, nil
}

func initCockroachDB(ctx context.Context, db sql.DB) error {
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

func truncateCockroachDB(ctx context.Context, db sql.DB) error {
	const query = `TRUNCATE projectmanagement.task`
	_, err := db.ExecContext(ctx, query)
	return err
}

func TestIntegrationDBInsertSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	cdbContainer, err := setupCockroachDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cdbContainer.Terminate(ctx)

	db, err := sql.Open("pgx", cdbContainer.URI+"/projectmanagement")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = initCockroachDB(ctx, *db)
	if err != nil {
		t.Fatal(err)
	}
	defer truncateCockroachDB(ctx, *db)

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
	if err != nil {
		t.Fatal(err)
	}

	// Select data
	savedTsk := task{ID: tsk.ID}
	const findQuery = `select description, date_due, date_created, date_updated
		from task
		where id = $1`
	row := db.QueryRowContext(ctx, findQuery, tsk.ID)
	err = row.Scan(&savedTsk.Description, &savedTsk.DateDue, &savedTsk.DateCreated, &savedTsk.DateUpdated)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(tsk, savedTsk) {
		t.Fatalf("Saved task is not the same:\n%s", cmp.Diff(tsk, savedTsk))
	}
}
```
