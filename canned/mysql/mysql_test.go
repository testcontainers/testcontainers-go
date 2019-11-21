package canned

import (
	"context"
	"testing"
)

func TestWriteIntoAMySQLContainerViaDriver(t *testing.T) {
	ctx := context.Background()
	c, err := MySQLContainer(ctx, MySQLContainerRequest{
		RootPassword: "root",
		Database:     "hello",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer c.Container.Terminate(ctx)
	sqlC, err := c.GetDriver("root", "root", "hello")
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = sqlC.ExecContext(ctx, "CREATE TABLE example ( id integer, data varchar(32) )")
	if err != nil {
		t.Fatal(err.Error())
	}
}
