package canned

import (
	"context"
	"testing"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestWriteIntoAPostgreSQLContainerViaDriver(t *testing.T) {

	ctx := context.Background()

	c, err := NewPostgreSQLContainer(ctx, PostgreSQLContainerRequest{
		Database: "hello",
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer c.Container.Terminate(ctx)

	sqlC, err := c.GetDriver(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = sqlC.ExecContext(ctx, "CREATE TABLE example ( id integer, data varchar(32) )")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func ExamplePostgreSQLContainerRequest() {

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
	postgreSQLContainerRequest := PostgreSQLContainerRequest{
		Database:                "mycustomdatabase",
		User:                    "anyuser",
		Password:                "yoursecurepassword",
		GenericContainerRequest: genericContainerRequest,
	}

	postgreSQLContainerRequest.Validate()
}

func ExampleNewPostgreSQLContainer() {
	ctx := context.Background()

	// Create your PostgreSQL database,
	// by setting Started this function will not return
	// until a test connection has been established
	c, _ := NewPostgreSQLContainer(ctx, PostgreSQLContainerRequest{
		Database: "hello",
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})
	defer c.Container.Terminate(ctx)
}

func ExamplePostgreSQLContainer_GetDriver() {
	ctx := context.Background()

	c, _ := NewPostgreSQLContainer(ctx, PostgreSQLContainerRequest{
		Database: "hello",
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})
	defer c.Container.Terminate(ctx)

	// Now you can simply interact with your DB
	db, _ := c.GetDriver(ctx)

	db.Ping()
}
