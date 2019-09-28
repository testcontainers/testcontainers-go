package canned

import (
	"context"
	"testing"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestWriteIntoAPostgreSQLContainerViaDriver(t *testing.T) {

	ctx := context.Background()

	c, err := PostgreSQLContainer(ctx, PostgreSQLContainerRequest{
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
