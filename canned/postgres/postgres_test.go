package postgres

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
)

func TestWriteIntoAPostgreSQLContainerViaDriver(t *testing.T) {

	ctx := context.Background()

	c, err := NewContainer(ctx, ContainerRequest{
		Version:  "9.6.15-alpine",
		Database: "hello",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer c.Container.Terminate(ctx)

	connectURL, err := c.ConnectURL(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	connectURL += "?sslmode=disable"

	sqlC, err := sql.Open("postgres", connectURL)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer sqlC.Close()

	_, err = sqlC.ExecContext(ctx, "CREATE TABLE example ( id integer, data varchar(32) )")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func ExampleContainerRequest() {

	// Optional
	containerRequest := testcontainers.ContainerRequest{
		Image: "docker.io/library/postgres:11.5",
	}

	genericContainerRequest := testcontainers.GenericContainerRequest{
		Started:          true,
		ContainerRequest: containerRequest,
	}

	// Database, User, and Password are optional,
	// the driver will use default ones if not provided
	pgContainerRequest := ContainerRequest{
		Database:                "mycustomdatabase",
		User:                    "anyuser",
		Password:                "yoursecurepassword",
		GenericContainerRequest: genericContainerRequest,
	}

	pgContainerRequest.Validate()
}

func ExampleNewContainer() {
	ctx := context.Background()

	// Create your PostgreSQL database,
	// by setting Started this function will not return
	// until a test connection has been established
	c, _ := NewContainer(ctx, ContainerRequest{
		Database: "hello",
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})
	defer c.Container.Terminate(ctx)
}
