package canned

import (
	"context"
	"testing"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestWriteIntoAMySQLContainerViaDriver(t *testing.T) {

	ctx := context.Background()

	c, err := NewMySQLContainer(ctx, MySQLContainerRequest{
		RootPassword: "root",
		Database:     "hello",
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

func ExampleMySQLContainerRequest() {

	// Optional
	containerRequest := testcontainers.ContainerRequest{
		Image: "docker.io/library/mysql:8.0",
	}

	genericContainerRequest := testcontainers.GenericContainerRequest{
		Started:          true,
		ContainerRequest: containerRequest,
	}

	// RootPassword, Database, User, and Password are optional,
	// the driver will use default ones if not provided
	mysqlContainerRequest := MySQLContainerRequest{
		RootPassword:            "rootpassword",
		User:                    "anyuser",
		Password:                "yoursecurepassword",
		Database:                "mycustomdatabase",
		GenericContainerRequest: genericContainerRequest,
	}

	mysqlContainerRequest.Validate()
}

func ExampleNewMySQLContainer() {
	ctx := context.Background()

	c, _ := NewMySQLContainer(ctx, MySQLContainerRequest{
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})
	defer c.Container.Terminate(ctx)
}

func ExampleMySQLContainer_GetDriver() {
	ctx := context.Background()

	c, _ := NewMySQLContainer(ctx, MySQLContainerRequest{
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})

	db, _ := c.GetDriver(ctx)

	db.Ping()
}
