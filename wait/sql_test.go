package wait_test

import (
	"context"
	"testing"

	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_ForSql(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8.0.20",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      "db",
			},
			WaitingFor: wait.ForSQL("3306/tcp", "mysql", func(p nat.Port) string {
				return "root:root@tcp(localhost:" + p.Port() + ")/db"
			}),
		},
		Started: true,
	}
	c, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Terminate(ctx)
}
